package core

import (
	"path/filepath"
	"strings"
)

type AutoTrackResult struct {
	Feature  string
	Stage    string
	Tests    string
	Updated  bool
	Previous string
}

var stageOrder = map[string]int{
	"":      -1,
	"prd":   0,
	"seed":  1,
	"bdd":   2,
	"tests": 3,
	"impl":  4,
}

func AutoTrack(projectDir, filePath string) (*AutoTrackResult, error) {
	rel := filePath
	if filepath.IsAbs(filePath) {
		r, err := filepath.Rel(projectDir, filePath)
		if err == nil {
			rel = r
		}
	}

	featureID, newStage, newTests := classifyForTracking(projectDir, rel)
	if featureID == "" {
		return nil, nil
	}

	rs, err := loadReviewStatus(projectDir)
	if err != nil {
		return nil, err
	}

	entry, ok := rs[featureID]
	if !ok {
		entry = ReviewStatusEntry{
			Stage:  "prd",
			Tests:  "absent",
			Review: "pending",
		}
	}

	result := &AutoTrackResult{
		Feature:  featureID,
		Stage:    entry.Stage,
		Tests:    entry.Tests,
		Previous: entry.Stage,
	}

	updated := false

	// Update tests field
	if newTests == "written" && entry.Tests != "written" {
		entry.Tests = "written"
		updated = true
		result.Tests = "written"
	}

	// Only advance stage, never regress
	if newStage != "" && stageOrder[newStage] > stageOrder[entry.Stage] {
		entry.Stage = newStage
		updated = true
		result.Stage = newStage
	}

	if updated {
		result.Updated = true
		rs[featureID] = entry
		if err := saveReviewStatus(projectDir, rs); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func classifyForTracking(projectDir, rel string) (featureID, stage, tests string) {
	// BDD file
	if strings.HasPrefix(rel, ".ptsd/bdd/") && strings.HasSuffix(rel, ".feature") {
		featureID = strings.TrimSuffix(filepath.Base(rel), ".feature")
		return featureID, "bdd", ""
	}

	// Seed file
	if strings.HasPrefix(rel, ".ptsd/seeds/") {
		parts := strings.Split(rel, "/")
		if len(parts) >= 3 {
			featureID = parts[2]
			return featureID, "seed", ""
		}
	}

	// Test file
	if strings.HasSuffix(rel, "_test.go") || strings.HasSuffix(rel, ".test.ts") || strings.HasSuffix(rel, ".test.js") {
		featureID = inferFeatureFromTestFile(projectDir, rel)
		if featureID != "" {
			return featureID, "tests", "written"
		}
		return "", "", ""
	}

	// Impl file
	if isImplFile(rel) {
		featureID = inferFeatureFromImplFile(projectDir, rel)
		if featureID != "" {
			return featureID, "impl", ""
		}
	}

	return "", "", ""
}
