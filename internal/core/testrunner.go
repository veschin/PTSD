package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TestResults struct {
	Total    int
	Passed   int
	Failed   int
	Failures []string
}

type CoverageEntry struct {
	Feature string
	Status  string
}

func MapTest(projectDir string, bddFile string, testFile string) error {
	bddPath := filepath.Join(projectDir, bddFile)
	data, err := os.ReadFile(bddPath)
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	featureID := ""
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "@feature:") {
			featureID = strings.TrimPrefix(line, "@feature:")
			break
		}
	}
	if featureID == "" {
		return fmt.Errorf("err:validation no @feature tag in %s", bddFile)
	}

	// Verify test file exists
	testPath := filepath.Join(projectDir, testFile)
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		return fmt.Errorf("err:validation test file %s not found", testFile)
	}

	state, err := LoadState(projectDir)
	if err != nil {
		return err
	}

	fs, ok := state.Features[featureID]
	if !ok {
		fs = FeatureState{
			Hashes: make(map[string]string),
			Scores: make(map[string]ScoreEntry),
		}
	}

	mapping := bddFile + "::" + testFile

	// Check for duplicate
	var testsList []string
	if fs.Tests != nil {
		if existing, ok := fs.Tests.([]string); ok {
			for _, t := range existing {
				if t == mapping {
					return nil // already mapped
				}
			}
			testsList = existing
		}
	}
	testsList = append(testsList, mapping)
	fs.Tests = testsList
	state.Features[featureID] = fs

	return writeState(projectDir, state)
}

func CheckTestCoverage(projectDir string) ([]CoverageEntry, error) {
	bddDir := filepath.Join(projectDir, ".ptsd", "bdd")
	entries, err := os.ReadDir(bddDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("err:io %w", err)
	}

	state, err := LoadState(projectDir)
	if err != nil {
		return nil, err
	}

	var coverage []CoverageEntry

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".feature") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(bddDir, e.Name()))
		if err != nil {
			continue
		}

		featureID := ""
		scenarioCount := 0
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "@feature:") {
				featureID = strings.TrimPrefix(line, "@feature:")
			}
			if strings.HasPrefix(line, "Scenario:") {
				scenarioCount++
			}
		}
		if featureID == "" {
			continue
		}

		// Count test mappings from state
		testCount := 0
		if fs, ok := state.Features[featureID]; ok {
			if tests, ok := fs.Tests.([]string); ok {
				testCount = len(tests)
			}
		}

		status := "no-tests"
		if testCount > 0 && testCount >= scenarioCount {
			status = "covered"
		} else if testCount > 0 {
			status = "partial"
		}

		coverage = append(coverage, CoverageEntry{Feature: featureID, Status: status})
	}

	return coverage, nil
}

func RunTests(projectDir string, featureFilter string) (TestResults, error) {
	cfg, err := LoadConfig(projectDir)
	if err != nil {
		return TestResults{}, err
	}

	if cfg.Testing.Runner == "" {
		return TestResults{}, fmt.Errorf("err:config no test runner configured")
	}

	runner := cfg.Testing.Runner

	// When a feature filter is specified, extract that feature's test files
	// from state and append them to the runner command.
	if featureFilter != "" {
		testFiles, err := featureTestFiles(projectDir, featureFilter)
		if err != nil {
			return TestResults{}, err
		}
		if len(testFiles) == 0 {
			return TestResults{}, fmt.Errorf("err:test no test files mapped for feature %s", featureFilter)
		}
		runner = runner + " " + strings.Join(testFiles, " ")
	}

	cmd := exec.Command("sh", "-c", runner)
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()

	// Parse results based on adapter selection
	var results TestResults
	if cfg.Testing.ResultParser.Format != "" {
		// Generic configurable adapter
		results = parseTAPOutput(string(output))
	} else if containsKnownRunner(cfg.Testing.Runner) {
		// Known runner preset
		results = parseTAPOutput(string(output))
	} else {
		// Exit-code adapter: pass if exit 0
		if err == nil {
			results = TestResults{Total: 1, Passed: 1}
		} else {
			results = TestResults{Total: 1, Failed: 1}
			if len(output) > 0 {
				results.Failures = []string{strings.TrimSpace(string(output))}
			}
		}
	}

	// Override with TAP parse if we got TAP-like output
	if strings.Contains(string(output), "ok ") || strings.Contains(string(output), "not ok ") {
		results = parseTAPOutput(string(output))
	}

	// Update state with results
	updateStateWithResults(projectDir, featureFilter, results)

	return results, nil
}

func containsKnownRunner(runner string) bool {
	known := []string{"vitest", "jest", "pytest", "go test"}
	for _, k := range known {
		if strings.Contains(runner, k) {
			return true
		}
	}
	return false
}

func parseTAPOutput(output string) TestResults {
	var results TestResults
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ok ") {
			results.Total++
			results.Passed++
		} else if strings.HasPrefix(line, "not ok ") {
			results.Total++
			results.Failed++
		} else if strings.HasPrefix(line, "# Failed at ") {
			failure := strings.TrimPrefix(line, "# Failed at ")
			results.Failures = append(results.Failures, failure)
		}
	}
	return results
}

// featureTestFiles extracts test file paths mapped to a feature from state.
// Mappings use the format "bddFile::testFile"; this returns only the test file parts.
func featureTestFiles(projectDir string, featureID string) ([]string, error) {
	state, err := LoadState(projectDir)
	if err != nil {
		return nil, err
	}
	fs, ok := state.Features[featureID]
	if !ok {
		return nil, nil
	}
	mappings, ok := fs.Tests.([]string)
	if !ok || len(mappings) == 0 {
		return nil, nil
	}
	var files []string
	for _, m := range mappings {
		parts := strings.SplitN(m, "::", 2)
		if len(parts) == 2 {
			files = append(files, parts[1])
		} else {
			// Plain test file path (no bdd:: prefix)
			files = append(files, m)
		}
	}
	return files, nil
}

func updateStateWithResults(projectDir string, featureFilter string, results TestResults) {
	state, err := LoadState(projectDir)
	if err != nil {
		return
	}

	resultStr := fmt.Sprintf("passed:%d failed:%d", results.Passed, results.Failed)

	if featureFilter != "" {
		fs, ok := state.Features[featureFilter]
		if !ok {
			fs = FeatureState{
				Hashes: make(map[string]string),
				Scores: make(map[string]ScoreEntry),
			}
		}
		// Store test results as a special hash entry for now
		if fs.Hashes == nil {
			fs.Hashes = make(map[string]string)
		}
		fs.Hashes["test_results"] = resultStr
		if results.Failed == 0 && results.Total > 0 {
			fs.Hashes["test_status"] = "passing"
		} else {
			fs.Hashes["test_status"] = "failing"
		}
		state.Features[featureFilter] = fs
	}

	writeState(projectDir, state)
}
