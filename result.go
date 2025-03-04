package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
)

type issues map[string][]*github.Issue

func (r issues) generateMarkdown(config *Config) markdownFiles {
	var (
		sb, sbi strings.Builder
		files   markdownFiles
	)

	sbi.WriteString("# Issue List\n\n")
	sbi.WriteString(fmt.Sprintf("Last Updated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sbi.WriteString(fmt.Sprintf("\n%s\n\n", config.Description))
	sbi.WriteString("## Index\n\n")

	basePath := config.Destination

	for _, k := range slices.Sorted(maps.Keys(r)) {
		issuePath := fmt.Sprintf("%s/issues/%s.md", basePath, k)

		sb.Reset()
		sb.WriteString(fmt.Sprintf("# %s\n\n", k))

		sb.WriteString("| Repository | Title | UpdatedAt | Labels | Assignee | Comments |\n")
		sb.WriteString("| --- | --- | --- | --- | --- | --- |\n")

		// Add an entry to index
		sbi.WriteString(fmt.Sprintf("- [%s - %d issues available](%s)\n", k, len(r[k]), issuePath))

		for _, issue := range r[k] {
			labels := make([]string, len(issue.Labels))
			for i, label := range issue.Labels {
				labels[i] = label.GetName()
			}
			assignee := issue.Assignee.GetName()
			owner, repoName, _ := parseRepoURL(issue.GetURL())
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

func (r issues) saveToFiles(config *Config) error {
	return r.generateMarkdown(config).saveToFiles()
}

type markdownFile struct {
	pathRelative string
	content      string
}

type markdownFiles []markdownFile

func (fs markdownFiles) saveToFiles() error {
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
