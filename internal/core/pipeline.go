package core

import (
	"os"
	"path/filepath"
	"strings"
)

type ValidationError struct {
	Feature  string
	Category string
	Message  string
}

func Validate(projectDir string) ([]ValidationError, error) {
	features, err := loadFeatures(projectDir)
	if err != nil {
		return nil, err
	}

	var errors []ValidationError

	// Check PRD anchors
	prdErrors, _ := CheckPRDAnchors(projectDir)
	for _, e := range prdErrors {
		if e.Type == "missing-anchor" {
			errors = append(errors, ValidationError{
				Feature:  e.FeatureID,
				Category: "pipeline",
				Message:  "has no prd anchor",
			})
		}
	}

	// Load state for per-feature test check
	state, _ := LoadState(projectDir)

	// Check pipeline consistency per feature
	for _, f := range features {
		if f.Status == "planned" || f.Status == "deferred" {
			continue
		}

		bddPath := filepath.Join(projectDir, ".ptsd", "bdd", f.ID+".feature")
		hasBDD := fileExists(bddPath)

		seedPath := filepath.Join(projectDir, ".ptsd", "seeds", f.ID, "seed.yaml")
		hasSeed := fileExists(seedPath)

		if hasBDD && !hasSeed {
			errors = append(errors, ValidationError{
				Feature:  f.ID,
				Category: "pipeline",
				Message:  "has bdd but no seed",
			})
		}

		if hasBDD {
			hasTests := hasTestsForFeature(projectDir, f.ID, state)
			if !hasTests {
				errors = append(errors, ValidationError{
					Feature:  f.ID,
					Category: "pipeline",
					Message:  "has bdd but no tests",
				})
			}
		}
	}

	// Check regressions
	regressions, _ := CheckRegressions(projectDir)
	for _, r := range regressions {
		errors = append(errors, ValidationError{
			Feature:  r.Feature,
			Category: "pipeline",
			Message:  r.Message,
		})
	}

	// Check for mock patterns in test files
	mockErrors := scanForMocks(projectDir)
	errors = append(errors, mockErrors...)

	return errors, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasTestsForFeature(projectDir string, featureID string, state *State) bool {
	// Check state for test mappings
	if state != nil {
		if fs, ok := state.Features[featureID]; ok {
			if tests, ok := fs.Tests.([]string); ok && len(tests) > 0 {
				return true
			}
		}
	}

	// Fallback: walk project for test files (global check)
	found := false
	filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && strings.Contains(path, ".ptsd") {
			return filepath.SkipDir
		}
		if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".test.ts") || strings.HasSuffix(path, ".test.js") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func scanForMocks(projectDir string) []ValidationError {
	var errors []ValidationError
	mockPatterns := []string{
		"vi.mock", "jest.mock", "unittest.mock",
		"gomock", "testify/mock", "mock.Mock",
	}

	filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && strings.Contains(path, ".ptsd") {
			return filepath.SkipDir
		}
		if !strings.HasSuffix(path, "_test.go") && !strings.HasSuffix(path, ".test.ts") && !strings.HasSuffix(path, ".test.js") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)
		for _, pattern := range mockPatterns {
			if strings.Contains(content, pattern) {
				relPath, _ := filepath.Rel(projectDir, path)
				errors = append(errors, ValidationError{
					Feature:  "",
					Category: "pipeline",
					Message:  "mock detected in " + relPath,
				})
				break
			}
		}
		return nil
	})

	return errors
}
