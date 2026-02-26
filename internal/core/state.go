package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type FeatureState struct {
	Stage  string
	Hashes map[string]string
	Scores map[string]ScoreEntry
	Tests  interface{}
}

type ScoreEntry struct {
	Value     int
	Timestamp time.Time
}

type State struct {
	Features map[string]FeatureState
}

type RegressionWarning struct {
	Feature  string
	File     string
	FileType string
	Severity string // "error" or "warn"
	Category string
	Message  string
}

func LoadState(projectDir string) (*State, error) {
	statePath := filepath.Join(projectDir, ".ptsd", "state.yaml")
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &State{Features: make(map[string]FeatureState)}, nil
		}
		return nil, fmt.Errorf("err:io %w", err)
	}

	return parseState(string(data))
}

func parseState(content string) (*State, error) {
	state := &State{Features: make(map[string]FeatureState)}
	lines := strings.Split(content, "\n")

	var currentFeature string
	var currentSection string
	var currentScoreStage string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || trimmed == "features:" || trimmed == "features: {}" {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))

		if indent == 2 && strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, " ") {
			currentFeature = strings.TrimSuffix(trimmed, ":")
			if _, ok := state.Features[currentFeature]; !ok {
				state.Features[currentFeature] = FeatureState{
					Hashes: make(map[string]string),
					Scores: make(map[string]ScoreEntry),
				}
			}
			currentSection = ""
			currentScoreStage = ""
			continue
		}

		if currentFeature == "" {
			continue
		}

		if indent == 4 {
			if strings.HasPrefix(trimmed, "stage: ") {
				fs := state.Features[currentFeature]
				fs.Stage = strings.TrimPrefix(trimmed, "stage: ")
				state.Features[currentFeature] = fs
				continue
			}
			if trimmed == "hashes:" || trimmed == "hashes: {}" {
				currentSection = "hashes"
				continue
			}
			if trimmed == "scores:" || trimmed == "scores: {}" {
				currentSection = "scores"
				continue
			}
			if trimmed == "tests:" || strings.HasPrefix(trimmed, "tests:") {
				currentSection = "tests"
				continue
			}
			if strings.Contains(trimmed, ": ") {
				continue
			}
		}

		if indent == 6 && currentSection == "hashes" && strings.Contains(trimmed, ": ") {
			parts := strings.SplitN(trimmed, ": ", 2)
			if len(parts) == 2 {
				fs := state.Features[currentFeature]
				fs.Hashes[parts[0]] = parts[1]
				state.Features[currentFeature] = fs
			}
			continue
		}

		if indent == 6 && currentSection == "scores" && strings.HasSuffix(trimmed, ":") {
			currentScoreStage = strings.TrimSuffix(trimmed, ":")
			continue
		}

		if indent == 8 && currentSection == "scores" && currentScoreStage != "" {
			if strings.HasPrefix(trimmed, "score: ") {
				val, _ := strconv.Atoi(strings.TrimPrefix(trimmed, "score: "))
				fs := state.Features[currentFeature]
				entry := fs.Scores[currentScoreStage]
				entry.Value = val
				fs.Scores[currentScoreStage] = entry
				state.Features[currentFeature] = fs
			}
			if strings.HasPrefix(trimmed, "at: ") {
				ts := strings.Trim(strings.TrimPrefix(trimmed, "at: "), "\"")
				t, err := time.Parse(time.RFC3339Nano, ts)
				if err != nil {
					t, _ = time.Parse(time.RFC3339, ts)
				}
				fs := state.Features[currentFeature]
				entry := fs.Scores[currentScoreStage]
				entry.Timestamp = t
				fs.Scores[currentScoreStage] = entry
				state.Features[currentFeature] = fs
			}
			continue
		}

		if indent == 6 && currentSection == "tests" && strings.HasPrefix(trimmed, "- ") {
			val := strings.TrimPrefix(trimmed, "- ")
			fs := state.Features[currentFeature]
			var testsList []string
			if fs.Tests != nil {
				if existing, ok := fs.Tests.([]string); ok {
					testsList = existing
				}
			}
			testsList = append(testsList, val)
			fs.Tests = testsList
			state.Features[currentFeature] = fs
			continue
		}
	}

	return state, nil
}

func SyncState(projectDir string) error {
	features, err := loadFeatures(projectDir)
	if err != nil {
		return err
	}

	state, err := LoadState(projectDir)
	if err != nil {
		return err
	}

	for _, f := range features {
		fs, ok := state.Features[f.ID]
		if !ok {
			fs = FeatureState{
				Hashes: make(map[string]string),
				Scores: make(map[string]ScoreEntry),
			}
		}

		seedPath := filepath.Join(projectDir, ".ptsd", "seeds", f.ID, "seed.yaml")
		if h, err := computeFileHash(seedPath); err == nil {
			fs.Hashes["seed"] = h
		}

		bddPath := filepath.Join(projectDir, ".ptsd", "bdd", f.ID+".feature")
		if h, err := computeFileHash(bddPath); err == nil {
			fs.Hashes["bdd"] = h
		}

		testPath := filepath.Join(projectDir, "internal", "core", f.ID+"_test.go")
		if h, err := computeFileHash(testPath); err == nil {
			fs.Hashes["test"] = h
		}

		prdPath := filepath.Join(projectDir, ".ptsd", "docs", "PRD.md")
		if h, err := computeFileHash(prdPath); err == nil {
			fs.Hashes["prd"] = h
		}

		state.Features[f.ID] = fs
	}

	return writeState(projectDir, state)
}

func CheckRegressions(projectDir string) ([]RegressionWarning, error) {
	state, err := LoadState(projectDir)
	if err != nil {
		return nil, err
	}

	var warnings []RegressionWarning
	stageOrder := map[string]int{"prd": 0, "seed": 1, "bdd": 2, "test": 3, "implemented": 4}

	for featureID, fs := range state.Features {
		currentStageIdx, ok := stageOrder[fs.Stage]
		if !ok {
			continue
		}

		type hashCheck struct {
			key      string
			path     string
			fileType string
			stageIdx int
		}

		ptsdDir := filepath.Join(projectDir, ".ptsd")
		checks := []hashCheck{
			{"prd", filepath.Join(ptsdDir, "docs", "PRD.md"), "prd", 0},
			{"seed", filepath.Join(ptsdDir, "seeds", featureID, "seed.yaml"), "seed", 1},
			{"bdd", filepath.Join(ptsdDir, "bdd", featureID+".feature"), "bdd", 2},
			{"test", filepath.Join(projectDir, "internal", "core", featureID+"_test.go"), "test", 3},
		}

		for _, c := range checks {
			oldHash, hasOld := fs.Hashes[c.key]
			if !hasOld {
				continue
			}

			newHash, err := computeFileHash(c.path)
			if err != nil {
				continue
			}

			if newHash == oldHash {
				continue
			}

			if c.stageIdx < currentStageIdx {
				if c.fileType == "prd" {
					// PRD change = ERROR: downgrade stage, create redo task
					warnings = append(warnings, RegressionWarning{
						Feature:  featureID,
						File:     c.path,
						FileType: c.fileType,
						Severity: "error",
						Category: "regression",
						Message:  fmt.Sprintf("%s changed at stage %s, stage downgraded", c.fileType, fs.Stage),
					})
					fs.Stage = c.fileType
					fs.Hashes[c.key] = newHash
					state.Features[featureID] = fs
					currentStageIdx = c.stageIdx
				} else if c.fileType == "test" && fs.Stage == "implemented" {
					// Test change at implemented = WARN, re-run tests, no downgrade
					warnings = append(warnings, RegressionWarning{
						Feature:  featureID,
						File:     c.path,
						FileType: c.fileType,
						Severity: "warn",
						Category: "regression",
						Message:  fmt.Sprintf("test changed at stage %s, re-run tests", fs.Stage),
					})
					fs.Hashes[c.key] = newHash
					state.Features[featureID] = fs
				} else {
					// Seed/BDD change = WARN, downstream may be stale, no downgrade
					warnings = append(warnings, RegressionWarning{
						Feature:  featureID,
						File:     c.path,
						FileType: c.fileType,
						Severity: "warn",
						Category: "regression",
						Message:  fmt.Sprintf("%s changed at stage %s, downstream may be stale", c.fileType, fs.Stage),
					})
					fs.Hashes[c.key] = newHash
					state.Features[featureID] = fs
				}
			} else {
				fs.Hashes[c.key] = newHash
				state.Features[featureID] = fs
			}
		}
	}

	if err := writeState(projectDir, state); err != nil {
		return warnings, fmt.Errorf("err:io failed to persist state: %w", err)
	}

	return warnings, nil
}

func writeState(projectDir string, state *State) error {
	statePath := filepath.Join(projectDir, ".ptsd", "state.yaml")

	var b strings.Builder
	b.WriteString("features:\n")

	// Sort feature keys for deterministic output
	keys := make([]string, 0, len(state.Features))
	for k := range state.Features {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, featureID := range keys {
		fs := state.Features[featureID]
		b.WriteString("  " + featureID + ":\n")
		b.WriteString("    stage: " + fs.Stage + "\n")

		b.WriteString("    hashes:\n")
		hkeys := make([]string, 0, len(fs.Hashes))
		for k := range fs.Hashes {
			hkeys = append(hkeys, k)
		}
		sort.Strings(hkeys)
		for _, k := range hkeys {
			b.WriteString("      " + k + ": " + fs.Hashes[k] + "\n")
		}

		b.WriteString("    scores:\n")
		skeys := make([]string, 0, len(fs.Scores))
		for k := range fs.Scores {
			skeys = append(skeys, k)
		}
		sort.Strings(skeys)
		for _, stage := range skeys {
			entry := fs.Scores[stage]
			b.WriteString("      " + stage + ":\n")
			b.WriteString("        score: " + strconv.Itoa(entry.Value) + "\n")
			b.WriteString("        at: \"" + entry.Timestamp.Format(time.RFC3339Nano) + "\"\n")
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

func computeFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
