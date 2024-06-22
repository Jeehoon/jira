package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
)

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
	time.Sleep(2 * time.Second)
	fmt.Println("AAAA")
	s.Stop()
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
