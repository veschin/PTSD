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

	// Check PRD anchors (CheckPRDAnchors includes all features; filter planned/deferred here)
	plannedOrDeferred := make(map[string]bool)
	for _, f := range features {
		if f.Status == "planned" || f.Status == "deferred" {
			plannedOrDeferred[f.ID] = true
		}
	}
	prdErrors, _ := CheckPRDAnchors(projectDir)
	for _, e := range prdErrors {
		if e.Type == "missing-anchor" && !plannedOrDeferred[e.FeatureID] {
			errors = append(errors, ValidationError{
				Feature:  e.FeatureID,
				Category: "pipeline",
				Message:  "has no prd anchor",
			})
		}
	}

	// Load state and review-status for per-feature stage check
	state, _ := LoadState(projectDir)
	reviewStatus, _ := loadReviewStatus(projectDir)

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

		// Only require tests if feature is past bdd stage
		currentStage := ""
		if rs, ok := reviewStatus[f.ID]; ok {
			currentStage = rs.Stage
		}
		if hasBDD && currentStage != "prd" && currentStage != "seed" && currentStage != "bdd" {
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

	// Check review gates per feature
	for _, f := range features {
		if f.Status == "planned" || f.Status == "deferred" {
			continue
		}
		if state == nil {
			continue
		}
		fs, ok := state.Features[f.ID]
		if !ok || fs.Stage == "" {
			continue
		}
		passed, err := CheckReviewGate(projectDir, f.ID, fs.Stage)
		if err != nil {
			errors = append(errors, ValidationError{
				Feature:  f.ID,
				Category: "pipeline",
				Message:  "review gate check failed: " + err.Error(),
			})
			continue
		}
		if !passed {
			errors = append(errors, ValidationError{
				Feature:  f.ID,
				Category: "pipeline",
				Message:  "review gate not passed for stage " + fs.Stage,
			})
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

	// Fallback: walk project for test files specific to this feature
	found := false
	filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && strings.Contains(path, ".ptsd") {
			return filepath.SkipDir
		}
		base := filepath.Base(path)
		if strings.Contains(base, featureID) &&
			(strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".test.ts") || strings.HasSuffix(path, ".test.js")) {
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
