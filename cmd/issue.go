/*
Copyright © 2024 Jeehoon <wlgns823@gmail.com>
*/
package cmd

import (
	"github.com/jeehoon/jira/internal/cmd/issue"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Control JIRA issue",
	Run: func(cmd *cobra.Command, args []string) {
		if err := issue.Run(args); err != nil {
			cobra.CheckErr(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(issueCmd)
	defer viper.BindPFlags(issueCmd.Flags())
	defer viper.BindPFlags(issueCmd.PersistentFlags())
}
