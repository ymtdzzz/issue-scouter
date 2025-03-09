package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/assert"
	"github.com/ymtdzzz/issue-scouter/pkg/client"
	"github.com/ymtdzzz/issue-scouter/pkg/config"
)

func TestGenerateMarkdown(t *testing.T) {
	fixedTime := time.Date(2025, 3, 9, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		config *config.Config
		issues client.Issues
		want   markdownFiles
	}{
		{
			name: "generates markdown files correctly",
			config: &config.Config{
				Destination: "output",
				Description: "Test description",
			},
			issues: client.Issues{
				"team-a": []*github.Issue{
					{
						Title:     github.Ptr("Issue 1"),
						HTMLURL:   github.Ptr("https://github.com/owner/repo/issues/1"),
						UpdatedAt: &github.Timestamp{Time: fixedTime},
						URL:       github.Ptr("https://github.com/owner/repo/issues/1"),
						Labels: []*github.Label{
							{Name: github.Ptr("bug")},
						},
						Assignee: &github.User{Login: github.Ptr("user1"), Name: github.Ptr("user1")},
						Comments: github.Ptr(2),
					},
				},
			},
			want: markdownFiles{
				{
					pathRelative: "output/issues/team-a.md",
					content: "# team-a\n\n" +
						"| Repository | Title | UpdatedAt | Labels | Assignee | Comments |\n" +
						"| --- | --- | --- | --- | --- | --- |\n" +
						"| [repo](https://github.com/owner/repo) | [Issue 1](https://github.com/owner/repo/issues/1) | 2025-03-09 | bug | user1 | 2 |\n\n",
				},
				{
					pathRelative: "output/README.md",
					content: "# Issue List\n\n" +
						fmt.Sprintf("Last Updated: %s\n", time.Now().Format("2006-01-02 15:04:05")) +
						"\nTest description\n\n" +
						"## Index\n\n" +
						"- [team-a - 1 issues available](output/issues/team-a.md)\n",
				},
			},
		},
		{
			name: "handles empty issues",
			config: &config.Config{
				Destination: "output",
				Description: "Test description",
			},
			issues: client.Issues{},
			want: markdownFiles{
				{
					pathRelative: "output/README.md",
					content: "# Issue List\n\n" +
						fmt.Sprintf("Last Updated: %s\n", time.Now().Format("2006-01-02 15:04:05")) +
						"\nTest description\n\n" +
						"## Index\n\n",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateMarkdown(tt.config, tt.issues)
			assert.Equal(t, len(tt.want), len(got))

			// Compare content excluding the dynamic timestamp
			for i := range got {
				assert.Equal(t, tt.want[i].pathRelative, got[i].pathRelative)
				// Skip exact content comparison for README.md as it contains dynamic timestamp
				if !strings.HasSuffix(got[i].pathRelative, "README.md") {
					assert.Equal(t, tt.want[i].content, got[i].content)
				}
			}
		})
	}
}

func TestSaveToFiles(t *testing.T) {
	tests := []struct {
		name    string
		files   markdownFiles
		wantErr bool
	}{
		{
			name: "saves files successfully",
			files: markdownFiles{
				{
					pathRelative: "test/output/file1.md",
					content:      "test content 1",
				},
				{
					pathRelative: "test/output/subdir/file2.md",
					content:      "test content 2",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Update paths to use temp directory
			for i := range tt.files {
				tt.files[i].pathRelative = filepath.Join(tmpDir, tt.files[i].pathRelative)
			}

			err := tt.files.saveToFiles()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Verify files were created with correct content
			for _, f := range tt.files {
				content, err := os.ReadFile(f.pathRelative)
				assert.NoError(t, err)
				assert.Equal(t, f.content, string(content))
			}
		})
	}
}
