/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/jeehoon/jira/internal/debug"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "jira",
	Short: "JIRA Command Line Tool",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	defer viper.BindPFlags(rootCmd.Flags())
	defer viper.BindPFlags(rootCmd.PersistentFlags())

	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.jira.yaml)")
	rootCmd.PersistentFlags().String("endpoint", "", "JIRA endpoint url")
	rootCmd.PersistentFlags().String("username", "", "JIRA username. email format")
	rootCmd.PersistentFlags().String("password", "", "JIRA password")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "display debugging logs")
}

func initConfig() {
	// Load Config
	cfgFile := viper.GetString("config")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".jira")
	}

	viper.SetEnvPrefix("JIRA")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		debug.Printf("Using config file: %v", viper.ConfigFileUsed())
	}

	// Set debug mode
	debug.Enabled = viper.GetBool("debug")

}
