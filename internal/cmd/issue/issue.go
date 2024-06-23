package issue

import (
	"github.com/jeehoon/jira/pkg/jira"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func Run(args []string) (err error) {
	endpoint := viper.GetString("endpoint")
	username := viper.GetString("username")
	password := viper.GetString("password")

	if len(args) == 0 {
		return errors.Errorf("Invalid Input: Issue Key required")
	}
	issueKey := args[0]

	cfg := &jira.Config{
		ApiEndpoint: endpoint,
		Username:    username,
		Password:    password,
	}

	client := jira.NewClient(cfg)
	if err := client.Init(); err != nil {
		return errors.Wrap(err, "client.Init")
	}

	_, err = client.GetIssue(issueKey)
	if err != nil {
		return errors.Wrap(err, "client.GetIssue")
	}

	return nil
}
