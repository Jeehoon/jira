package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pkg/errors"
)

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

type FieldMeta struct {
	Key              string
	Name             string
	UntranslatedName string
	Type             string
	Item             string
	Custom           bool
}

func (client *Client) GetFieldMeta(key string) *FieldMeta {
	return client.fieldMetas[key]
}
