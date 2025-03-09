package main

import (
	"log"
	"os"

	"github.com/ymtdzzz/issue-scouter/pkg/client"
	"github.com/ymtdzzz/issue-scouter/pkg/config"
)

func main() {
	configFile := os.Getenv("INPUT_CONFIG_FILE")
	if configFile == "" {
		log.Fatal("No config file specified")
		os.Exit(1)
	}

	co, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
		os.Exit(1)
	}

	c := client.NewClient(co)
	issues, err := c.FetchIssues()
	if err != nil {
		log.Fatalf("Failed to fetch issues: %v", err)
		os.Exit(1)
	}

	err = saveToFiles(co, issues)
	if err != nil {
		log.Fatalf("Failed to save Markdown file: %v", err)
		os.Exit(1)
	}
}
