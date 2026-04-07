package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MapResult contains the codebase analysis.
type MapResult struct {
	Language    string
	Runner      string
	TestCount   int
	SourceCount int
	Dirs        []DirSummary
}

// DirSummary describes a directory in the project.
type DirSummary struct {
	Path    string
	Sources int
	Tests   int
}

// MapProject scans the project and generates .ptsd/docs/CODEBASE.md.
func MapProject(dir string) (*MapResult, error) {
	result := &MapResult{}

	// Detect language and runner
	result.Runner = detectTestRunner(dir)
	if fileExists(filepath.Join(dir, "go.mod")) {
		result.Language = "Go"
	} else if fileExists(filepath.Join(dir, "package.json")) {
		result.Language = "JavaScript/TypeScript"
	} else if fileExists(filepath.Join(dir, "pyproject.toml")) || fileExists(filepath.Join(dir, "setup.py")) {
		result.Language = "Python"
	} else if fileExists(filepath.Join(dir, "Cargo.toml")) {
		result.Language = "Rust"
	} else if fileExists(filepath.Join(dir, "Gemfile")) {
		result.Language = "Ruby"
	} else if fileExists(filepath.Join(dir, "pom.xml")) || fileExists(filepath.Join(dir, "build.gradle")) {
		result.Language = "Java"
	}

	// Scan directories
	dirStats := make(map[string]*DirSummary)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".ptsd" || base == ".git" || base == "vendor" || base == "node_modules" || base == ".claude" || base == ".opencode" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, _ := filepath.Rel(dir, path)
		d := filepath.Dir(rel)
		if d == "" {
			d = "."
		}

		if _, ok := dirStats[d]; !ok {
			dirStats[d] = &DirSummary{Path: d}
		}

		if IsTestFile(path) {
			dirStats[d].Tests++
			result.TestCount++
		} else {
			dirStats[d].Sources++
			result.SourceCount++
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("err:io %w", err)
	}

	// Sort directories by path
	for _, ds := range dirStats {
		result.Dirs = append(result.Dirs, *ds)
	}
	sort.Slice(result.Dirs, func(i, j int) bool {
		return result.Dirs[i].Path < result.Dirs[j].Path
	})

	// Generate CODEBASE.md
	var md strings.Builder
	md.WriteString("# Codebase Map\n\n")
	md.WriteString(fmt.Sprintf("**Language:** %s\n", result.Language))
	if result.Runner != "" {
		md.WriteString(fmt.Sprintf("**Test runner:** `%s`\n", result.Runner))
	}
	md.WriteString(fmt.Sprintf("**Source files:** %d | **Test files:** %d\n\n", result.SourceCount, result.TestCount))

	md.WriteString("## Directory Structure\n\n")
	md.WriteString("| Directory | Sources | Tests |\n")
	md.WriteString("|-----------|---------|-------|\n")
	for _, ds := range result.Dirs {
		md.WriteString(fmt.Sprintf("| `%s` | %d | %d |\n", ds.Path, ds.Sources, ds.Tests))
	}
	md.WriteString("\n")

	// Write to .ptsd/docs/CODEBASE.md
	docsDir := filepath.Join(dir, ".ptsd", "docs")
	os.MkdirAll(docsDir, 0755)
	codebakePath := filepath.Join(docsDir, "CODEBASE.md")
	if err := os.WriteFile(codebakePath, []byte(md.String()), 0644); err != nil {
		return nil, fmt.Errorf("err:io %w", err)
	}

	return result, nil
}
