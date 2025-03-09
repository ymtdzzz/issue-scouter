package main

import (
	"context"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"

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
	}

	c := client.NewClient(co)
	ctx := context.Background()

	for _, k := range slices.Sorted(maps.Keys(co.Repos)) {
		fmt.Printf("\n=== Category: %s ===\n", k)
		for _, repo := range co.Repos[k] {
			owner, repoName, err := config.ParseRepoURL(repo)
			if err != nil {
				log.Printf("Failed to parse repository URL %s: %v\n", repo, err)
				continue
			}

			fmt.Printf("\nRepository: %s/%s\n", owner, repoName)
			labels, _, err := c.GetClient().Issues.ListLabels(ctx, owner, repoName, nil)
			if err != nil {
				log.Printf("Failed to fetch labels for %s/%s: %v\n", owner, repoName, err)
				continue
			}

			fmt.Println("Labels:")
			for _, label := range labels {
				fmt.Printf("  - %s\n", label.GetName())
			}
		}
	}
}
