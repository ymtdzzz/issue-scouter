package client

import (
	"context"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/google/go-github/v69/github"
	"github.com/ymtdzzz/issue-scouter/pkg/config"
	"golang.org/x/oauth2"
)

const (
	API_URL_BASE = "api.github.com/repos"
	URL_BASE     = "github.com"
)

type client struct {
	ghc    *github.Client
	config *config.Config
}

type Issues map[string][]*github.Issue

func NewClient(config *config.Config) *client {
	var ghc *github.Client
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Println("GITHUB_TOKEN is not set, initialize Github client without credentials")
		ghc = github.NewClient(nil)

		return &client{ghc, config}
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	ghc = github.NewClient(tc)

	log.Println("Github client is initialized with given credentials")

	return &client{ghc, config}
}

func (c *client) GetClient() *github.Client {
	return c.ghc
}

func (c *client) FetchIssues() (Issues, error) {
	issues := Issues{}
	chunkSize := 100

	for _, k := range slices.Sorted(maps.Keys(c.config.Repos)) {
		var gis []*github.Issue

		ownerRepos := make([]string, 0, len(c.config.Repos[k]))
		for _, r := range c.config.Repos[k] {
			owner, repo, err := config.ParseRepoURL(r)
			if err != nil {
				log.Printf("Failed to parse repository URL: %v", err)
				continue
			}
			ownerRepos = append(ownerRepos, owner+"/"+repo)
		}

		// Process repositories in chunks
		for i := 0; i < len(ownerRepos); i += chunkSize {
			end := i + chunkSize
			if end > len(ownerRepos) {
				end = len(ownerRepos)
			}
			chunk := ownerRepos[i:end]

			is, err := c.fetchIssuesByRepos(chunk)
			if err != nil {
				log.Printf("Failed to fetch issues for chunk in %s: %v", k, err)
				continue
			}
			gis = append(gis, is...)
		}

		issues[k] = gis
	}
	return issues, nil
}

func (c *client) fetchIssuesByRepos(ownerRepo []string) ([]*github.Issue, error) {
	ctx := context.Background()

	repos := make([]string, len(ownerRepo))
	for i, r := range ownerRepo {
		repos[i] = "repo:" + r
	}
	reposForQuery := strings.Join(repos, " ")

	q := fmt.Sprintf("%s is:open is:issue label:%s", reposForQuery, c.config.LabelsForQuery())
	log.Printf("Query: %s", q)
	opts := &github.SearchOptions{
		TextMatch: true,
		ListOptions: github.ListOptions{
			PerPage: c.config.PerPage,
		},
	}
	results, _, err := c.ghc.Search.Issues(ctx, q, opts)
	if err != nil {
		return nil, err
	}
	issues := results.Issues

	// Replace URL
	for i := range issues {
		replacedURL := strings.Replace(issues[i].GetURL(), API_URL_BASE, URL_BASE, 1)
		issues[i].URL = &replacedURL
	}

	// Sort by Repository and UpdatedAt
	sort.Slice(issues, func(i, j int) bool {
		a, b := issues[i], issues[j]
		_, aRepo, _ := config.ParseRepoURL(a.GetURL())
		_, bRepo, _ := config.ParseRepoURL(b.GetURL())
		switch {
		case aRepo < bRepo:
			return true
		case aRepo > bRepo:
			return false
		default:
			return a.GetUpdatedAt().Time.After(b.GetUpdatedAt().Time)
		}
	})

	return issues, err
}
