package client

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/issue-scouter/pkg/config"
	"golang.org/x/oauth2"
)

func createMockIssue(number int, title, repo string, updatedAt time.Time) *github.Issue {
	return &github.Issue{
		Number:    github.Ptr(number),
		Title:     github.Ptr(title),
		URL:       github.Ptr("https://github.com/" + repo),
		UpdatedAt: &github.Timestamp{Time: updatedAt},
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantToken bool
	}{
		{
			name:      "should create client without token",
			token:     "",
			wantToken: false,
		},
		{
			name:      "should create client with token",
			token:     "test-token",
			wantToken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.token != "" {
				os.Setenv("GITHUB_TOKEN", tt.token)
				defer os.Unsetenv("GITHUB_TOKEN")
			} else {
				os.Unsetenv("GITHUB_TOKEN")
			}

			cfg := &config.Config{}
			client := NewClient(cfg)

			assert.NotNil(t, client)
			assert.NotNil(t, client.ghc)
			assert.Equal(t, cfg, client.config)

			transport := client.ghc.Client().Transport
			_, hasToken := transport.(*oauth2.Transport)
			if tt.wantToken {
				assert.True(t, hasToken, "Transport should be oauth2.Transport when token is provided")
			} else {
				assert.False(t, hasToken, "Transport should not be oauth2.Transport when token is not provided")
			}
		})
	}
}

func TestFetchIssues(t *testing.T) {
	baseTime := time.Now()
	mockIssues := []*github.Issue{
		createMockIssue(1, "Issue 1", "owner1/repo1", baseTime),
		createMockIssue(2, "Issue 2", "owner2/repo2", baseTime.Add(-1*time.Hour)),
	}

	tests := []struct {
		name          string
		config        *config.Config
		mockResponses []mock.MockBackendOption
		wantErr       bool
		wantCount     map[string]int
	}{
		{
			name: "should fetch issues from empty repos",
			config: &config.Config{
				Repos: map[string][]string{},
			},
			mockResponses: nil,
			wantErr:       false,
			wantCount:     map[string]int{},
		},
		{
			name: "should fetch issues successfully",
			config: &config.Config{
				Repos: map[string][]string{
					"test": {"https://github.com/owner1/repo1", "https://github.com/owner2/repo2"},
				},
				Labels: []string{"help-wanted"},
			},
			mockResponses: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.GetSearchIssues,
					&github.IssuesSearchResult{
						Total:  github.Ptr(2),
						Issues: mockIssues,
					},
				),
			},
			wantErr: false,
			wantCount: map[string]int{
				"test": 2,
			},
		},
		{
			name: "should handle API error",
			config: &config.Config{
				Repos: map[string][]string{
					"test": {"https://github.com/owner/repo"},
				},
				Labels: []string{"help-wanted"},
			},
			mockResponses: []mock.MockBackendOption{
				mock.WithRequestMatchHandler(
					mock.GetSearchIssues,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						mock.WriteError(
							w,
							http.StatusInternalServerError,
							"github API error",
						)
					}),
				),
			},
			wantErr: false,
			wantCount: map[string]int{
				"test": 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedHTTPClient := mock.NewMockedHTTPClient(tt.mockResponses...)
			ghClient := github.NewClient(mockedHTTPClient)

			client := &client{
				ghc:    ghClient,
				config: tt.config,
			}

			issues, err := client.FetchIssues()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for key, count := range tt.wantCount {
					assert.Len(t, issues[key], count)
				}
			}
		})
	}
}

func Test_fetchIssuesByRepos(t *testing.T) {
	baseTime := time.Now()
	mockIssues := []*github.Issue{
		createMockIssue(1, "Issue 1", "owner1/repo1", baseTime),
		createMockIssue(2, "Issue 2", "owner2/repo2", baseTime.Add(-1*time.Hour)),
	}

	tests := []struct {
		name          string
		ownerRepos    []string
		labels        []string
		mockResponses []mock.MockBackendOption
		wantErr       bool
		wantCount     int
	}{
		{
			name:       "should handle empty repos",
			ownerRepos: []string{},
			labels:     []string{"help-wanted"},
			mockResponses: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.GetSearchIssues,
					&github.IssuesSearchResult{},
				),
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:       "should handle multiple repos",
			ownerRepos: []string{"https://github.com/owner1/repo1", "https://github.com/owner2/repo2"},
			labels:     []string{"help-wanted"},
			mockResponses: []mock.MockBackendOption{
				mock.WithRequestMatch(
					mock.GetSearchIssues,
					&github.IssuesSearchResult{
						Total:  github.Ptr(2),
						Issues: mockIssues,
					},
				),
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:       "should handle API error",
			ownerRepos: []string{"https://github.com/owner/repo"},
			labels:     []string{"help-wanted"},
			mockResponses: []mock.MockBackendOption{
				mock.WithRequestMatchHandler(
					mock.GetSearchIssues,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						mock.WriteError(
							w,
							http.StatusInternalServerError,
							"github API error",
						)
					}),
				),
			},
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedHTTPClient := mock.NewMockedHTTPClient(tt.mockResponses...)
			ghClient := github.NewClient(mockedHTTPClient)

			client := &client{
				ghc: ghClient,
				config: &config.Config{
					Labels: tt.labels,
				},
			}

			issues, err := client.fetchIssuesByRepos(tt.ownerRepos)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, issues, tt.wantCount)

				if tt.wantCount > 0 {
					// Verify sorting
					for i := 0; i < len(issues)-1; i++ {
						_, repo1, _ := config.ParseRepoURL(issues[i].GetURL())
						_, repo2, _ := config.ParseRepoURL(issues[i+1].GetURL())
						if repo1 == repo2 {
							assert.True(t, issues[i].GetUpdatedAt().After(issues[i+1].GetUpdatedAt().Time))
						}
					}
				}
			}
		})
	}
}
