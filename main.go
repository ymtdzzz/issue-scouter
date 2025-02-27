package main

import (
	"log"
	"os"
)

func main() {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		log.Fatal("No config file specified")
		os.Exit(1)
	}

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
		os.Exit(1)
	}

	client := newClient(config)
	issues, err := client.fetchIssues()
	if err != nil {
		log.Fatalf("Failed to fetch issues: %v", err)
		os.Exit(1)
	}

	err = issues.saveToFiles()
	if err != nil {
		log.Fatalf("Failed to save Markdown file: %v", err)
		os.Exit(1)
	}
}
