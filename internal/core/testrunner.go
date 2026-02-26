package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
	// Parse feature tag from BDD file
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

	// Load state
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

	// Store mapping as test entry
	mapping := bddFile + "::" + testFile
	var testsList []string
	if fs.Tests != nil {
		if existing, ok := fs.Tests.([]string); ok {
			testsList = existing
		}
	}
	testsList = append(testsList, mapping)
	fs.Tests = testsList
	state.Features[featureID] = fs

	// Write state with test mappings
	return writeStateWithTests(projectDir, state)
}

func CheckTestCoverage(projectDir string) ([]CoverageEntry, error) {
	bddDir := filepath.Join(projectDir, ".ptsd", "bdd")
	entries, _ := os.ReadDir(bddDir)

	state, err := LoadState(projectDir)
	if err != nil {
		return nil, err
	}

	// Read test mappings from state.yaml raw content
	statePath := filepath.Join(projectDir, ".ptsd", "state.yaml")
	stateData, _ := os.ReadFile(statePath)
	stateContent := string(stateData)

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

		// Count test mappings for this feature
		testCount := 0
		_ = state // use raw content instead
		inFeature := false
		inTests := false
		for _, line := range strings.Split(stateContent, "\n") {
			trimmed := strings.TrimSpace(line)
			indent := len(line) - len(strings.TrimLeft(line, " "))

			if indent == 2 && strings.HasSuffix(trimmed, ":") {
				name := strings.TrimSuffix(trimmed, ":")
				inFeature = name == featureID
				inTests = false
				continue
			}
			if inFeature && indent == 4 && (trimmed == "tests:" || strings.HasPrefix(trimmed, "tests:")) {
				inTests = true
				continue
			}
			if inFeature && inTests && indent == 6 && strings.HasPrefix(trimmed, "- ") {
				testCount++
				continue
			}
			if inFeature && indent <= 4 && trimmed != "" && !strings.HasPrefix(trimmed, "- ") {
				inTests = false
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

	// Execute test runner
	cmd := exec.Command("sh", "-c", cfg.Testing.Runner)
	cmd.Dir = projectDir
	output, _ := cmd.CombinedOutput()

	// Parse TAP format
	results := parseTAPOutput(string(output))

	// Update state with results
	updateStateWithResults(projectDir, featureFilter, results)

	return results, nil
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

func updateStateWithResults(projectDir string, featureFilter string, results TestResults) {
	statePath := filepath.Join(projectDir, ".ptsd", "state.yaml")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return
	}

	content := string(data)
	resultLine := fmt.Sprintf("    passed: %d\n    failed: %d\n", results.Passed, results.Failed)

	if featureFilter != "" {
		// Find the feature section and append results
		lines := strings.Split(content, "\n")
		var newLines []string
		inFeature := false
		added := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			indent := len(line) - len(strings.TrimLeft(line, " "))

			if indent == 2 && strings.HasSuffix(trimmed, ":") {
				if inFeature && !added {
					newLines = append(newLines, resultLine)
					added = true
				}
				name := strings.TrimSuffix(trimmed, ":")
				inFeature = name == featureFilter
			}
			newLines = append(newLines, line)
		}
		if inFeature && !added {
			newLines = append(newLines, resultLine)
		}
		content = strings.Join(newLines, "\n")
	} else {
		// Append results at the end
		content += resultLine
	}

	os.WriteFile(statePath, []byte(content), 0644)
}

func writeStateWithTests(projectDir string, state *State) error {
	statePath := filepath.Join(projectDir, ".ptsd", "state.yaml")

	var b strings.Builder
	b.WriteString("features:\n")

	for featureID, fs := range state.Features {
		b.WriteString("  " + featureID + ":\n")
		if fs.Stage != "" {
			b.WriteString("    stage: " + fs.Stage + "\n")
		}

		if len(fs.Hashes) > 0 {
			b.WriteString("    hashes:\n")
			for k, v := range fs.Hashes {
				b.WriteString("      " + k + ": " + v + "\n")
			}
		}

		if len(fs.Scores) > 0 {
			b.WriteString("    scores:\n")
			for stage, entry := range fs.Scores {
				b.WriteString("      " + stage + ":\n")
				b.WriteString(fmt.Sprintf("        score: %d\n", entry.Value))
				b.WriteString("        at: \"" + entry.Timestamp.Format(time.RFC3339Nano) + "\"\n")
			}
		}

		if fs.Tests != nil {
			if tests, ok := fs.Tests.([]string); ok && len(tests) > 0 {
				b.WriteString("    tests:\n")
				for _, t := range tests {
					b.WriteString("      - " + t + "\n")
				}
			}
		}
	}

	return os.WriteFile(statePath, []byte(b.String()), 0644)
}
