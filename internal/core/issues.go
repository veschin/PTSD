package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Issue represents a recurring problem entry in the common issues registry.
type Issue struct {
	ID       string
	Category string
	Summary  string
	Fix      string
}

var validIssueCategories = map[string]bool{
	"env":    true,
	"access": true,
	"io":     true,
	"config": true,
	"test":   true,
	"llm":    true,
}

// LoadIssues reads issues from .ptsd/issues.yaml. Returns empty list if file does not exist.
func LoadIssues(projectDir string) ([]Issue, error) {
	issuesPath := filepath.Join(projectDir, ".ptsd", "issues.yaml")
	data, err := os.ReadFile(issuesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("err:io %w", err)
	}

	return parseIssues(string(data)), nil
}

// ListIssues returns all issues, optionally filtered by category.
func ListIssues(projectDir string, categoryFilter string) ([]Issue, error) {
	issues, err := LoadIssues(projectDir)
	if err != nil {
		return nil, err
	}

	if categoryFilter == "" {
		return issues, nil
	}

	var filtered []Issue
	for _, issue := range issues {
		if issue.Category == categoryFilter {
			filtered = append(filtered, issue)
		}
	}
	return filtered, nil
}

// AddIssue validates and appends a new issue to .ptsd/issues.yaml.
func AddIssue(projectDir string, issue Issue) error {
	if !validIssueCategories[issue.Category] {
		return fmt.Errorf("err:user invalid category %q: must be env|access|io|config|test|llm", issue.Category)
	}
	if strings.TrimSpace(issue.Summary) == "" {
		return fmt.Errorf("err:user summary required")
	}
	if strings.TrimSpace(issue.Fix) == "" {
		return fmt.Errorf("err:user fix required")
	}

	issues, err := LoadIssues(projectDir)
	if err != nil {
		return err
	}

	for _, existing := range issues {
		if existing.ID == issue.ID {
			return fmt.Errorf("err:validation issue %s already exists", issue.ID)
		}
	}

	issues = append(issues, issue)
	return saveIssues(projectDir, issues)
}

// RemoveIssue deletes an issue by ID from .ptsd/issues.yaml.
func RemoveIssue(projectDir string, id string) error {
	issues, err := LoadIssues(projectDir)
	if err != nil {
		return err
	}

	found := false
	var remaining []Issue
	for _, issue := range issues {
		if issue.ID == id {
			found = true
			continue
		}
		remaining = append(remaining, issue)
	}
	if !found {
		return fmt.Errorf("err:validation issue %s not found", id)
	}

	return saveIssues(projectDir, remaining)
}

func parseIssues(content string) []Issue {
	var issues []Issue
	lines := strings.Split(content, "\n")
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "- id: ") {
			issue := Issue{ID: strings.TrimPrefix(trimmed, "- id: ")}
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if strings.HasPrefix(next, "- id: ") || next == "" || (!strings.HasPrefix(lines[j], "    ") && !strings.HasPrefix(lines[j], "  ")) {
					break
				}
				switch {
				case strings.HasPrefix(next, "category: "):
					issue.Category = strings.TrimPrefix(next, "category: ")
				case strings.HasPrefix(next, "summary: "):
					val := strings.TrimPrefix(next, "summary: ")
					issue.Summary = stripQuotes(val)
				case strings.HasPrefix(next, "fix: "):
					val := strings.TrimPrefix(next, "fix: ")
					issue.Fix = stripQuotes(val)
				}
			}
			issues = append(issues, issue)
		}
	}
	return issues
}

func saveIssues(projectDir string, issues []Issue) error {
	issuesPath := filepath.Join(projectDir, ".ptsd", "issues.yaml")

	if err := os.MkdirAll(filepath.Dir(issuesPath), 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	var b strings.Builder
	b.WriteString("issues:\n")
	for _, issue := range issues {
		b.WriteString("  - id: " + issue.ID + "\n")
		b.WriteString("    category: " + issue.Category + "\n")
		b.WriteString("    summary: \"" + issue.Summary + "\"\n")
		b.WriteString("    fix: \"" + issue.Fix + "\"\n")
	}

	return os.WriteFile(issuesPath, []byte(b.String()), 0644)
}
