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

func (r issues) generateMarkdown() markdownFiles {
	var (
		sb, sbi strings.Builder
		files   markdownFiles
	)

	sbi.WriteString("# Issue List\n\n")
	sbi.WriteString(fmt.Sprintf("Last Updated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sbi.WriteString("\nThis file is generated by [issue-scouter](https://github.com/ymtdzzz/issue-scouter)\n\n")
	sbi.WriteString("## Index\n\n")

	for _, k := range slices.Sorted(maps.Keys(r)) {
		issuePath := fmt.Sprintf("./issues/%s.md", k)

		sb.Reset()
		sb.WriteString(fmt.Sprintf("# %s\n\n", k))

		sb.WriteString("| Repository | Title | UpdatedAt | Labels | Assignee |\n")
		sb.WriteString("| --- | --- | --- | --- | --- |\n")

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
				"| [%s](https://github.com/%s/%s) | [%s](%s) | %s | %s | %s |\n",
				repoName,
				owner,
				repoName,
				issue.GetTitle(),
				issue.GetURL(),
				issue.GetUpdatedAt(),
				strings.Join(labels, ", "),
				assignee,
			))
		}
		sb.WriteString("\n")
		files = append(files, markdownFile{
			pathRelative: issuePath,
			content:      sb.String(),
		})
	}
	files = append(files, markdownFile{
		pathRelative: "./README.md",
		content:      sbi.String(),
	})

	return files
}

func (r issues) saveToFiles() error {
	return r.generateMarkdown().saveToFiles()
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
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return fmt.Errorf("Failed to create directory %s: %v", dir, err)
			}
		}

		if err := os.WriteFile(f.pathRelative, []byte(f.content), 0644); err != nil {
			return fmt.Errorf("Failed to save a file: %v", err)
		}
	}

	return nil
}
