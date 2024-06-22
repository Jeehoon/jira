package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

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
	Duedate                  string
}

func (issue *Issue) parseStandardField(fkey string, fvalue any) {
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

	case "duedate":
		issue.Duedate = fvalue.(string)
	}
}

func (issue *Issue) parseCustomField(fkey string, fmeta *FieldMeta, fvalue any) {
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
		var items []any
		arr := fvalue.([]any)
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
			}
		}

		if len(items) > 0 {
			issue.CustomFields[fkey] = items
		}
	}
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

		if v, ok := fvalue.(map[string]any); ok && len(v) == 0 {
			continue
		}

		if strings.HasPrefix(fkey, "customfield_") {
			fmeta := client.GetFieldMeta(fkey)
			if fmeta == nil {
				continue
			}
			issue.parseCustomField(fkey, fmeta, fvalue)
		} else {
			issue.parseStandardField(fkey, fvalue)
		}
	}

	return issue, nil
}
