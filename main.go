package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
)

// /rest/api/3/search?jql=assignee%3DcurrentUser%28%29+and+status+%21%3D+Closed+and+status+%21%3D+Resolved+ORDER+BY+created+DESC&startAt=0&maxResults=100
// /rest/api/3/search?jql=assignee%3DcurrentUser%28%29+and+status+%21%3D+Closed+and+status+%21%3D+Resolved+ORDER+BY+created+DESC&startAt=0&maxResults=100", nil)

var enc = json.NewEncoder(os.Stdout)

func init() {
	enc.SetIndent("", "  ")
}

func dumpJson(a any) {
	enc.Encode(a)
}

func Spinner() {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " Suffix..."
	s.FinalMSG = "Finished!\n"
	s.Color("fgHiCyan")
	s.Start()
	time.Sleep(4 * time.Second)
	fmt.Println("AAAA")
	s.Stop()
}

type Client struct {
	apiEndpoint string
	username    string
	password    string
	httpClient  *http.Client
	fieldMetas  map[string]*FieldMeta
}

type Config struct {
	ApiEndpoint string
	Username    string
	Password    string
}

func NewClient(cfg *Config) (client *Client) {
	client = &Client{
		apiEndpoint: cfg.ApiEndpoint,
		username:    cfg.Username,
		password:    cfg.Password,
		httpClient:  &http.Client{},
	}

	return client
}

func (client *Client) Do(method string, p string, data []byte, out any) (err error) {
	url := client.apiEndpoint + p

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return errors.Wrap(err, "http.NewRequest")
	}

	if client.username != "" {
		req.SetBasicAuth(client.username, client.password)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "httpClient.Do")
	}

	if resp.StatusCode > 299 {
		return errors.Errorf("HTTP Request failed: %v", resp.Status)
	}

	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&out); err != nil {
		return errors.Wrap(err, "json.Decode")
	}

	return nil
}

func (client *Client) Init() (err error) {
	if err := client.loadAllFields(); err != nil {
		return errors.Wrap(err, "loadFields")
	}
	return nil
}

func (client *Client) ListRaw(p, name string) (arr []any, err error) {
	var maxResults = 100
	var startAt = 0

	for {
		var v map[string]any
		p := fmt.Sprintf("%v?startAt=%v&maxResults=%v", p, startAt, maxResults)
		if err := client.Do("GET", p, nil, &v); err != nil {
			return nil, errors.Wrap(err, "client.Do")
		}

		items := v[name].([]any)
		total := int(v["total"].(float64))
		arr = append(arr, items...)
		startAt = len(arr)
		if total <= startAt {
			break
		}
	}

	return arr, nil
}

func (client *Client) GetRaw(p string) (obj map[string]any, err error) {
	if err := client.Do("GET", p, nil, &obj); err != nil {
		return nil, errors.Wrap(err, "client.Do")
	}

	return obj, nil
}

func (client *Client) GetIssueTypes(project string) (arr []any, err error) {
	p := fmt.Sprintf("/rest/api/3/issue/createmeta/%v/issuetypes", project)
	return client.ListRaw(p, "issueTypes")
}

func (client *Client) GetFields(project, typeId string) (arr []any, err error) {
	p := fmt.Sprintf("/rest/api/3/issue/createmeta/%v/issuetypes/%v", project, typeId)
	return client.ListRaw(p, "fields")
}

func (client *Client) loadAllFields() (err error) {
	var fieldMetas = make(map[string]*FieldMeta)

	var arr []any
	if err := client.Do("GET", "/rest/api/3/field", nil, &arr); err != nil {
		return errors.Wrap(err, "client.Do")
	}

	for _, v := range arr {
		fm := v.(map[string]any)
		item := new(FieldMeta)
		item.Key = fm["key"].(string)
		item.Name = fm["name"].(string)
		item.Custom = fm["custom"].(bool)

		if v, ok := fm["schema"]; ok {
			fschema := v.(map[string]any)
			item.Type = fschema["type"].(string)
			if item.Type == "array" {
				item.Item = fschema["items"].(string)
			}
		}

		fieldMetas[item.Key] = item
	}

	client.fieldMetas = fieldMetas

	log.Printf("The number of FieldMeta: %v", len(client.fieldMetas))

	return nil
}

type User struct {
	DisplayName  string
	EmailAddress string
}

type IssueLink struct {
	Kind    string
	Key     string
	Summary string
	Status  string
}

type SubTask struct {
	Key     string
	Summary string
	Status  string
}

type Parent struct {
	Key     string
	Summary string
	Status  string
}

type Issue struct {
	Key                      string
	ProjectKey               string
	Type                     string
	Status                   string
	Summary                  string
	Description              string
	Created                  string
	Updated                  string
	LastViewed               string
	StatusCategoryChangeDate string
	Parent                   *Parent
	SubTasks                 []*SubTask
	IssueLinks               []*IssueLink
	Assignee                 *User
	Reporter                 *User
	Creator                  *User
	Labels                   []string
	Priority                 string
	CustomFields             map[string]any
	IsWatching               bool
	WatchCount               int
	Comment                  string
}

type FieldMeta struct {
	Key    string
	Name   string
	Type   string
	Item   string
	Custom bool
}

func (client *Client) GetFieldMeta(key string) *FieldMeta {
	return client.fieldMetas[key]
}

func (client *Client) GetIssue(key string) (issue *Issue, err error) {
	p := fmt.Sprintf("/rest/api/3/issue/%v", key)
	issueObj, err := client.GetRaw(p)
	if err != nil {
		return nil, errors.Wrap(err, "client.GetRaw")
	}

	issue = new(Issue)
	issue.CustomFields = make(map[string]any)

	issue.Key = issueObj["key"].(string)

	for fkey, fvalue := range issueObj["fields"].(map[string]any) {

		if fvalue == nil {
			continue
		}

		if v, ok := fvalue.([]any); ok && len(v) == 0 {
			continue
		}

		if v, ok := fvalue.([]map[string]any); ok && len(v) == 0 {
			continue
		}

		switch fkey {
		case "summary":
			issue.Summary = fvalue.(string)
		case "created":
			issue.Created = fvalue.(string)
		case "statuscategorychangedate":
			issue.StatusCategoryChangeDate = fvalue.(string)
		case "updated":
			issue.Updated = fvalue.(string)
		case "lastViewed":
			issue.LastViewed = fvalue.(string)
		case "issuelinks":
			issueLinks := fvalue.([]any)
			for _, v := range issueLinks {
				item := new(IssueLink)

				issueLink := v.(map[string]any)
				linkType := issueLink["type"].(map[string]any)

				if v, ok := issueLink["inwardIssue"]; ok {
					item.Kind = linkType["inward"].(string)

					linkIssue := v.(map[string]any)
					fields := linkIssue["fields"].(map[string]any)
					item.Key = linkIssue["key"].(string)
					item.Summary = fields["summary"].(string)

					status := fields["status"].(map[string]any)
					item.Status = status["name"].(string)
				}

				if v, ok := issueLink["outwardIssue"]; ok {
					item.Kind = linkType["outward"].(string)

					linkIssue := v.(map[string]any)
					fields := linkIssue["fields"].(map[string]any)
					item.Key = linkIssue["key"].(string)
					item.Summary = fields["summary"].(string)

					status := fields["status"].(map[string]any)
					item.Status = status["name"].(string)
				}

				issue.IssueLinks = append(issue.IssueLinks, item)
			}
		case "subtasks":
			subtasks := fvalue.([]any)

			for _, v := range subtasks {
				item := new(SubTask)

				subtask := v.(map[string]any)
				fields := subtask["fields"].(map[string]any)
				item.Key = subtask["key"].(string)
				item.Summary = fields["summary"].(string)
				status := fields["status"].(map[string]any)
				item.Status = status["name"].(string)

				issue.SubTasks = append(issue.SubTasks, item)
			}
		case "assignee":
			item := new(User)
			user := fvalue.(map[string]any)
			item.DisplayName = user["displayName"].(string)
			if _, ok := user["emailAddress"]; ok {
				item.EmailAddress = user["emailAddress"].(string)
			}
			issue.Assignee = item
		case "reporter":
			item := new(User)
			user := fvalue.(map[string]any)
			item.DisplayName = user["displayName"].(string)
			if _, ok := user["emailAddress"]; ok {
				item.EmailAddress = user["emailAddress"].(string)
			}
			issue.Reporter = item
		case "issuetype":
			issueType := fvalue.(map[string]any)
			issue.Type = issueType["name"].(string)
		case "creator":
			item := new(User)
			user := fvalue.(map[string]any)
			item.DisplayName = user["displayName"].(string)
			if _, ok := user["emailAddress"]; ok {
				item.EmailAddress = user["emailAddress"].(string)
			}
			issue.Creator = item
		case "status":
			status := fvalue.(map[string]any)
			issue.Status = status["name"].(string)
		case "project":
			project := fvalue.(map[string]any)
			issue.ProjectKey = project["key"].(string)
		case "labels":
			for _, v := range fvalue.([]any) {
				issue.Labels = append(issue.Labels, v.(string))
			}
		case "parent":
			item := new(Parent)
			parent := fvalue.(map[string]any)
			fields := parent["fields"].(map[string]any)
			item.Key = parent["key"].(string)
			item.Summary = fields["summary"].(string)
			status := fields["status"].(map[string]any)
			item.Status = status["name"].(string)
			issue.Parent = item
		case "priority":
			obj := fvalue.(map[string]any)
			issue.Priority = obj["name"].(string)

		case "description":
			issue.Description = fmt.Sprintf("%+v", fvalue)

		case "comment":
			issue.Comment = fmt.Sprintf("%+v", fvalue)

		case "watches":
			obj := fvalue.(map[string]any)
			issue.IsWatching = obj["isWatching"].(bool)
			issue.WatchCount = int(obj["watchCount"].(float64))

			/*
				case "attachment":
				case "environment":
				case "components":
				case "security":
				case "aggregateprogress":
				case "workratio":
				case "aggregatetimeoriginalestimate":
				case "timespent":
				case "progress":
				case "resolution":
				case "issuerestriction":
				case "aggregatetimespent":
				case "timeestimate":
				case "versions":
				case "timeoriginalestimate":
				case "aggregatetimeestimate":
				case "worklog":
				case "resolutiondate":
				case "fixVersions":
				case "votes":
			*/

		default:
			if strings.HasPrefix(fkey, "customfield_") {
				// CUSTOM FIELD
				fmeta := client.GetFieldMeta(fkey)
				if fmeta == nil {
					continue
				}

				switch fmeta.Type {
				case "user":
					item := new(User)

					obj := fvalue.(map[string]any)
					item.DisplayName = obj["displayName"].(string)
					item.EmailAddress = obj["emailAddress"].(string)

					issue.CustomFields[fkey] = item
				case "datetime":
					issue.CustomFields[fkey] = fvalue.(string)
				case "date":
					issue.CustomFields[fkey] = fvalue.(string)
				case "number":
					issue.CustomFields[fkey] = fvalue.(float64)

				case "option":
					obj := fvalue.(map[string]any)
					issue.CustomFields[fkey] = obj["value"].(string)

				case "string":
					if _, ok := fvalue.(string); ok {
						issue.CustomFields[fkey] = fvalue.(string)
					} else {
						issue.CustomFields[fkey] = fmt.Sprintf("%+v", fvalue)
					}

				case "option-with-child":
					obj := fvalue.(map[string]any)
					value1 := obj["value"].(string)
					child := obj["child"].(map[string]any)
					value2 := child["value"].(string)
					issue.CustomFields[fkey] = []string{value1, value2}

				case "array":
					arr := fvalue.([]any)

					var items []any

					for _, v := range arr {
						switch fmeta.Item {
						case "string":
							value := v.(string)
							items = append(items, value)
						case "user":
							item := new(User)
							obj := v.(map[string]any)
							item.DisplayName = obj["displayName"].(string)
							if _, ok := obj["emailAddress"]; ok {
								item.EmailAddress = obj["emailAddress"].(string)
							}
							items = append(items, item)
						case "option":
							obj := v.(map[string]any)
							value := obj["value"].(string)
							items = append(items, value)
							//						case "json":
							// TODO Sprint
							//							dumpJson(fvalue)
						default:
							fmt.Printf("C %v(%v) %v %v\n", fkey, fmeta.Name, fmeta.Type, fmeta.Item)
						}
					}

					if len(items) > 0 {
						issue.CustomFields[fkey] = items
					}

				default:
					fmt.Printf("C %v(%v) %v\n", fkey, fmeta.Name, fmeta.Type)
					dumpJson(fvalue)
					fmt.Println()
					fmt.Println()
				}

			} else {
				// STANDARD FIELD
				fmt.Printf("S %v %T\n", fkey, fvalue)
				dumpJson(fvalue)
				fmt.Println()
				fmt.Println()
			}
		}
	}

	return issue, nil
}

func main() {
	client := NewClient(&Config{
		ApiEndpoint: os.Getenv("JIRA_ENDPOINT"),
		Username:    os.Getenv("JIRA_USERNAME"),
		Password:    os.Getenv("JIRA_PASSWORD"),
	})

	if err := client.Init(); err != nil {
		log.Fatal(err)
	}

	issue, err := client.GetIssue(os.Getenv("JIRA_ISSUE"))
	if err != nil {
		log.Fatal(err)
	}

	dumpJson(issue)
}
