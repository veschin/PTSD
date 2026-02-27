package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AdoptResult contains results from a dry-run or actual adopt operation.
type AdoptResult struct {
	BDDFiles     []string // feature IDs discovered from .feature files
	TestFiles    []string // test file paths discovered
	FeaturesFile string   // path to features.yaml that would be created
}

// AdoptProject bootstraps .ptsd/ structure for an existing project.
// It scans for BDD .feature files and test files, extracts feature IDs,
// and creates the .ptsd/ directory structure. Fails if .ptsd/ already exists.
func AdoptProject(dir string) error {
	ptsdDir := filepath.Join(dir, ".ptsd")
	if _, err := os.Stat(ptsdDir); err == nil {
		return fmt.Errorf("err:validation already initialized")
	}

	result, err := scanProject(dir)
	if err != nil {
		return err
	}

	return applyAdopt(dir, result)
}

// AdoptDryRun scans the project and returns what would be done without making changes.
func AdoptDryRun(dir string) (*AdoptResult, error) {
	ptsdDir := filepath.Join(dir, ".ptsd")
	if _, err := os.Stat(ptsdDir); err == nil {
		return nil, fmt.Errorf("err:validation already initialized")
	}

	return scanProject(dir)
}

// scanProject discovers BDD files and test files in the project directory.
func scanProject(dir string) (*AdoptResult, error) {
	result := &AdoptResult{
		FeaturesFile: filepath.Join(dir, ".ptsd", "features.yaml"),
	}

	// Discover BDD .feature files with @feature: tags
	bddFiles, err := discoverBDDFiles(dir)
	if err != nil {
		return nil, err
	}
	result.BDDFiles = bddFiles

	// Discover test files using default pattern
	testFiles, err := discoverTestFiles(dir)
	if err != nil {
		return nil, err
	}
	result.TestFiles = testFiles

	return result, nil
}

// discoverBDDFiles finds .feature files and extracts feature IDs from @feature: tags.
func discoverBDDFiles(dir string) ([]string, error) {
	var featureIDs []string
	seen := make(map[string]bool)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		// Skip .ptsd directory (shouldn't exist yet, but be safe)
		if info.IsDir() && filepath.Base(path) == ".ptsd" {
			return filepath.SkipDir
		}
		if !strings.HasSuffix(path, ".feature") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "@feature:") {
				id := strings.TrimPrefix(line, "@feature:")
				id = strings.TrimSpace(id)
				if id != "" && !seen[id] {
					seen[id] = true
					featureIDs = append(featureIDs, id)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("err:io %w", err)
	}

	return featureIDs, nil
}

// discoverTestFiles finds test files matching the default Go test pattern.
func discoverTestFiles(dir string) ([]string, error) {
	var testFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && filepath.Base(path) == ".ptsd" {
			return filepath.SkipDir
		}
		if strings.HasSuffix(path, "_test.go") {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				rel = path
			}
			testFiles = append(testFiles, rel)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("err:io %w", err)
	}

	return testFiles, nil
}

// applyAdopt creates the .ptsd/ directory structure and imports discovered artifacts.
func applyAdopt(dir string, result *AdoptResult) error {
	ptsdDir := filepath.Join(dir, ".ptsd")

	// Create directory structure
	dirs := []string{
		ptsdDir,
		filepath.Join(ptsdDir, "seeds"),
		filepath.Join(ptsdDir, "bdd"),
		filepath.Join(ptsdDir, "docs"),
		filepath.Join(ptsdDir, "skills"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("err:io %w", err)
		}
	}

	// Create ptsd.yaml with defaults
	ptsdYAML := "project:\n  name: \"\"\ntesting:\n  patterns:\n    files: [\"**/*_test.go\"]\nreview:\n  min_score: 7\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(ptsdYAML), 0644); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	// Create features.yaml from discovered BDD feature IDs
	var b strings.Builder
	b.WriteString("features:\n")
	for _, id := range result.BDDFiles {
		b.WriteString("  - id: " + id + "\n")
		b.WriteString("    title: " + id + "\n")
		b.WriteString("    status: planned\n")
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	// Move discovered .feature files to .ptsd/bdd/
	bddDir := filepath.Join(ptsdDir, "bdd")
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && filepath.Base(path) == ".ptsd" {
			return filepath.SkipDir
		}
		if !strings.HasSuffix(path, ".feature") {
			return nil
		}

		basename := filepath.Base(path)
		dst := filepath.Join(bddDir, basename)

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("err:io %w", err)
		}
		return os.Remove(path)
	})
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	return nil
}
