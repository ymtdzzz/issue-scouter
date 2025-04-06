package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelsForQuery(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   string
	}{
		{
			name: "single label",
			config: &Config{
				Labels: []string{"good first issue"},
			},
			want: "\"good first issue\"",
		},
		{
			name: "multiple labels",
			config: &Config{
				Labels: []string{"good first issue", "help wanted"},
			},
			want: "\"good first issue\",\"help wanted\"",
		},
		{
			name: "empty labels",
			config: &Config{
				Labels: []string{},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.LabelsForQuery()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		validate func(*testing.T, *Config)
	}{
		{
			name: "valid config with defaults",
			content: `
repositories:
  owner1:
    - repo1
    - repo2`,
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, []string{"good first issue"}, c.Labels)
				assert.Equal(t, 100, c.PerPage)
				assert.Equal(t, ".", c.Destination)
				assert.Contains(t, c.Description, "issue-scouter")
				assert.Contains(t, c.Repos, "owner1")
				assert.Equal(t, []string{"repo1", "repo2"}, c.Repos["owner1"])
				assert.False(t, c.IncludeMetadata, "IncludeMetadata should be false by default")
			},
		},
		{
			name: "with include_metadata true",
			content: `
repositories:
  owner1:
    - repo1
include_metadata: true`,
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				assert.True(t, c.IncludeMetadata, "IncludeMetadata should be true when explicitly set")
			},
		},
		{
			name: "with include_metadata false",
			content: `
repositories:
  owner1:
    - repo1
include_metadata: false`,
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				assert.False(t, c.IncludeMetadata, "IncludeMetadata should be false when explicitly set")
			},
		},
		{
			name: "custom values",
			content: `
repositories:
  owner1:
    - repo1
labels:
  - help wanted
per_page: 50
destination: "./output"
description: "Custom description"`,
			wantErr: false,
			validate: func(t *testing.T, c *Config) {
				assert.Equal(t, []string{"help wanted"}, c.Labels)
				assert.Equal(t, 50, c.PerPage)
				assert.Equal(t, "./output", c.Destination)
				assert.Equal(t, "Custom description", c.Description)
			},
		},
		{
			name:     "invalid yaml",
			content:  "invalid: [yaml: content",
			wantErr:  true,
			validate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config.yml")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			assert.NoError(t, err)

			config, err := LoadConfig(tmpFile)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			tt.validate(t, config)
		})
	}
}

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "valid github url",
			url:       "https://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "invalid url format",
			url:       "https://github.com/owner",
			wantOwner: "",
			wantRepo:  "",
			wantErr:   true,
		},
		{
			name:      "not github url",
			url:       "https://gitlab.com/owner/repo",
			wantOwner: "",
			wantRepo:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepoURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOwner, owner)
			assert.Equal(t, tt.wantRepo, repo)
		})
	}
}
