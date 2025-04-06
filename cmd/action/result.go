package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ymtdzzz/issue-scouter/pkg/client"
	"github.com/ymtdzzz/issue-scouter/pkg/config"
)

type IssueMetadata struct {
	Title     string           `json:"title"`
	Body      string           `json:"body"`
	Labels    []LabelMetadata  `json:"labels"`
	Assignee  AssigneeMetadata `json:"assignee,omitempty"`
	Comments  int              `json:"comments"`
	UpdatedAt string           `json:"updated_at"`
	URL       string           `json:"url"`
}

type LabelMetadata struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

type AssigneeMetadata struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

func generateMarkdown(c *config.Config, issues client.Issues) markdownFiles {
	var (
		sb, sbi strings.Builder
		files   markdownFiles
	)

	sbi.WriteString("# Issue List\n\n")
	sbi.WriteString(fmt.Sprintf("Last Updated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sbi.WriteString(fmt.Sprintf("\n%s\n\n", c.Description))
	sbi.WriteString("## Index\n\n")

	basePath := c.Destination

	for _, k := range slices.Sorted(maps.Keys(issues)) {
		issuePath := fmt.Sprintf("%s/issues/%s.md", basePath, k)

		sb.Reset()
		sb.WriteString(fmt.Sprintf("# %s\n\n", k))

		sb.WriteString("| Repository | Title | UpdatedAt | Labels | Assignee | Comments |\n")
		sb.WriteString("| --- | --- | --- | --- | --- | --- |\n")

		// Add an entry to index
		sbi.WriteString(fmt.Sprintf("- [%s - %d issues available](./issues/%s.md)\n", k, len(issues[k]), k))

		for _, issue := range issues[k] {
			labels := make([]string, len(issue.Labels))
			for i, label := range issue.Labels {
				labels[i] = label.GetName()
			}
			var assignee string
			if issue.Assignee.GetLogin() != "" {
				assignee = "@" + issue.Assignee.GetLogin()
			}
			owner, repoName, _ := config.ParseRepoURL(issue.GetURL())
			sb.WriteString(fmt.Sprintf(
				"| [%s](https://github.com/%s/%s) | [%s](%s) | %s | %s | %s | %d |\n",
				repoName,
				owner,
				repoName,
				issue.GetTitle(),
				issue.GetURL(),
				issue.GetUpdatedAt().Time.Format("2006-01-02"),
				strings.Join(labels, ", "),
				assignee,
				issue.GetComments(),
			))
		}
		sb.WriteString("\n")

		if c.IncludeMetadata {
			for _, issue := range issues[k] {
				metadata := IssueMetadata{
					Title: issue.GetTitle(),
					Body:  issue.GetBody(),
					Labels: func() []LabelMetadata {
						labels := make([]LabelMetadata, len(issue.Labels))
						for i, l := range issue.Labels {
							labels[i] = LabelMetadata{
								Name:        l.GetName(),
								Color:       l.GetColor(),
								Description: l.GetDescription(),
							}
						}
						return labels
					}(),
					Comments:  issue.GetComments(),
					UpdatedAt: issue.GetUpdatedAt().Time.Format(time.RFC3339),
					URL:       issue.GetURL(),
				}
				if issue.Assignee != nil {
					metadata.Assignee = AssigneeMetadata{
						Login: issue.Assignee.GetLogin(),
						Name:  issue.Assignee.GetName(),
						Email: issue.Assignee.GetEmail(),
					}
				}
				jsonData, err := json.MarshalIndent(metadata, "", "  ")
				if err != nil {
					log.Printf("Failed to marshal metadata for issue %s: %v", issue.GetTitle(), err)
					continue
				}
				sb.WriteString("\n<!--\n")
				sb.Write(jsonData)
				sb.WriteString("\n-->\n")
			}
		}

		files = append(files, markdownFile{
			pathRelative: issuePath,
			content:      sb.String(),
		})
	}
	files = append(files, markdownFile{
		pathRelative: fmt.Sprintf("%s/README.md", basePath),
		content:      sbi.String(),
	})

	return files
}

func saveToFiles(config *config.Config, issues client.Issues) error {
	return generateMarkdown(config, issues).saveToFiles(config)
}

type markdownFile struct {
	pathRelative string
	content      string
}

type markdownFiles []markdownFile

func (fs markdownFiles) saveToFiles(config *config.Config) error {
	// First, delete issues directory to remove old files if exists
	issuesDir := fmt.Sprintf("%s/issues", config.Destination)
	if _, err := os.Stat(issuesDir); err == nil {
		log.Printf("Removing old issues directory: %s", issuesDir)
		if err := os.RemoveAll(issuesDir); err != nil {
			return fmt.Errorf("failed to remove issues directory: %v", err)
		}
	}

	for _, f := range fs {
		dir := filepath.Dir(f.pathRelative)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0750)
			if err != nil {
				return fmt.Errorf("failed to create directory %s: %v", dir, err)
			}
		}

		if err := os.WriteFile(f.pathRelative, []byte(f.content), 0640); err != nil {
			return fmt.Errorf("failed to save a file: %v", err)
		}
	}

	return nil
}
