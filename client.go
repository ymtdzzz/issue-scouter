package main

import (
	"context"
	"log"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

const (
	API_URL_BASE = "api.github.com/repos"
	URL_BASE     = "github.com"
)

type client struct {
	ghc    *github.Client
	config *Config
}

func newClient(config *Config) *client {
	var ghc *github.Client
	token := os.Getenv("ISSUE_SCOUTER_PAT")
	if token == "" {
		log.Println("ISSUE_SCOUTER_PAT is not set, initialize Github client without credentials")
		ghc = github.NewClient(nil)

		return &client{ghc, config}
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	ghc = github.NewClient(tc)

	log.Println("Github client is initialized with given credentials")

	return &client{ghc, config}
}

func (c *client) fetchIssues() (issues, error) {
	issues := issues{}
	for _, k := range slices.Sorted(maps.Keys(c.config.Repos)) {
		var gis []*github.Issue
		for _, repo := range c.config.Repos[k] {
			owner, repo, err := parseRepoURL(repo)
			if err != nil {
				log.Fatalf("Failed to parse repository URL: %v", err)
				return nil, err
			}
			for _, label := range c.config.Labels {
				is, err := c.fetchIssuesByRepo(owner, repo, label)
				if err != nil {
					log.Fatalf("Failed to fetch issues: %v", err)
					return nil, err
				}
				gis = append(gis, is...)
			}
		}
		issues[k] = gis
	}
	return issues, nil
}

func (c *client) fetchIssuesByRepo(owner, repo, label string) ([]*github.Issue, error) {
	ctx := context.Background()

	opts := &github.IssueListByRepoOptions{
		Labels: []string{label},
		State:  "open",
		ListOptions: github.ListOptions{
			PerPage: c.config.PerPage,
		},
	}

	issues, _, err := c.ghc.Issues.ListByRepo(ctx, owner, repo, opts)

	// Replace URL
	for i := range issues {
		replacedURL := strings.Replace(issues[i].GetURL(), API_URL_BASE, URL_BASE, 1)
		issues[i].URL = &replacedURL
	}

	return issues, err
}
