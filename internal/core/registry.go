package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Feature struct {
	ID       string
	Title    string
	Status   string
	Pipeline string
}

type FeatureDetail struct {
	ID            string
	Status        string
	Pipeline      string
	PRDAnchor     string
	SeedStatus    string
	ScenarioCount int
	TestCount     int
	TestPassed    int
}

var validFeatureID = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

var validStatuses = map[string]bool{
	"planned":     true,
	"in-progress": true,
	"implemented": true,
	"deferred":    true,
}

func AddFeature(projectDir string, id string, title string) error {
	if !validFeatureID.MatchString(id) {
		return fmt.Errorf("err:validation invalid feature ID %q: must be ASCII slug (a-z0-9 with hyphens)", id)
	}

	features, err := loadFeatures(projectDir)
	if err != nil {
		return err
	}

	for _, f := range features {
		if f.ID == id {
			return fmt.Errorf("err:validation feature %s already exists", id)
		}
	}

	features = append(features, Feature{ID: id, Title: title, Status: "planned"})
	return saveFeatures(projectDir, features)
}

func ListFeatures(projectDir string, statusFilter string) ([]Feature, error) {
	features, err := loadFeatures(projectDir)
	if err != nil {
		return nil, err
	}

	if statusFilter == "" {
		return features, nil
	}

	var filtered []Feature
	for _, f := range features {
		if f.Status == statusFilter {
			filtered = append(filtered, f)
		}
	}
	return filtered, nil
}

func ShowFeature(projectDir string, id string) (FeatureDetail, error) {
	features, err := loadFeatures(projectDir)
	if err != nil {
		return FeatureDetail{}, err
	}

	var found *Feature
	for i := range features {
		if features[i].ID == id {
			found = &features[i]
			break
		}
	}
	if found == nil {
		return FeatureDetail{}, fmt.Errorf("err:validation feature %s not found", id)
	}

	detail := FeatureDetail{
		ID:       found.ID,
		Status:   found.Status,
		Pipeline: found.Pipeline,
	}
	if detail.Pipeline == "" {
		detail.Pipeline = resolveDefaultPipeline(projectDir)
	}

	seedDir := filepath.Join(projectDir, ".ptsd", "seeds", id)
	if info, err := os.Stat(seedDir); err == nil && info.IsDir() {
		detail.SeedStatus = "ok"
	} else {
		detail.SeedStatus = "missing"
	}

	bddFile := filepath.Join(projectDir, ".ptsd", "bdd", id+".feature")
	if data, err := os.ReadFile(bddFile); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "Scenario:") {
				detail.ScenarioCount++
			}
		}
	}

	detail.TestCount, detail.TestPassed = readTestStats(projectDir, id)

	prdPath := filepath.Join(projectDir, ".ptsd", "docs", "PRD.md")
	if data, err := os.ReadFile(prdPath); err == nil {
		anchor := "<!-- feature:" + id + " -->"
		for i, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) == anchor {
				detail.PRDAnchor = fmt.Sprintf("l%d", i+1)
				break
			}
		}
	}

	return detail, nil
}

func UpdateFeatureStatus(projectDir string, id string, newStatus string) error {
	if !validStatuses[newStatus] {
		return fmt.Errorf("err:validation invalid status %q: must be planned|in-progress|implemented|deferred", newStatus)
	}

	features, err := loadFeatures(projectDir)
	if err != nil {
		return err
	}

	idx := -1
	for i := range features {
		if features[i].ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("err:validation feature %s not found", id)
	}

	if newStatus == "implemented" {
		statePath := filepath.Join(projectDir, ".ptsd", "state.yaml")
		data, err := os.ReadFile(statePath)
		if err != nil {
			return fmt.Errorf("err:pipeline tests not passing for %s", id)
		}
		testStatus := parseTestStatus(string(data), id)
		if testStatus != "passing" {
			return fmt.Errorf("err:pipeline tests not passing for %s", id)
		}
	}

	features[idx].Status = newStatus
	return saveFeatures(projectDir, features)
}

func RemoveFeature(projectDir string, id string) error {
	features, err := loadFeatures(projectDir)
	if err != nil {
		return err
	}

	found := false
	var filtered []Feature
	for _, f := range features {
		if f.ID == id {
			found = true
		} else {
			filtered = append(filtered, f)
		}
	}

	if !found {
		return fmt.Errorf("err:validation feature %s not found", id)
	}

	return saveFeatures(projectDir, filtered)
}

func loadFeatures(projectDir string) ([]Feature, error) {
	featPath := filepath.Join(projectDir, ".ptsd", "features.yaml")
	data, err := os.ReadFile(featPath)
	if err != nil {
		return nil, fmt.Errorf("err:io %w", err)
	}

	var features []Feature
	lines := strings.Split(string(data), "\n")
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "- id: ") {
			f := Feature{ID: strings.TrimPrefix(trimmed, "- id: ")}
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if strings.HasPrefix(next, "- id: ") || next == "" {
					break
				}
				if strings.HasPrefix(next, "title: ") {
					f.Title = strings.TrimPrefix(next, "title: ")
					f.Title = strings.Trim(f.Title, "\"")
				}
				if strings.HasPrefix(next, "status: ") {
					f.Status = strings.TrimPrefix(next, "status: ")
				}
				if strings.HasPrefix(next, "pipeline: ") {
					f.Pipeline = strings.TrimPrefix(next, "pipeline: ")
				}
			}
			features = append(features, f)
		}
	}

	return features, nil
}

func saveFeatures(projectDir string, features []Feature) error {
	featPath := filepath.Join(projectDir, ".ptsd", "features.yaml")

	var b strings.Builder
	b.WriteString("features:\n")
	for _, f := range features {
		b.WriteString("  - id: " + f.ID + "\n")
		title := f.Title
		if strings.ContainsAny(title, " :\"'#") {
			title = "\"" + strings.ReplaceAll(title, "\"", "\\\"") + "\""
		}
		b.WriteString("    title: " + title + "\n")
		b.WriteString("    status: " + f.Status + "\n")
		if f.Pipeline != "" {
			b.WriteString("    pipeline: " + f.Pipeline + "\n")
		}
	}

	return os.WriteFile(featPath, []byte(b.String()), 0644)
}

// readTestStats returns (total, passed) test counts for a feature.
// Reads from test_results hash ("passed:N failed:M") in state.yaml,
// falling back to counting test file mappings.
func readTestStats(projectDir string, featureID string) (total int, passed int) {
	state, err := LoadState(projectDir)
	if err != nil {
		return 0, 0
	}
	fs, ok := state.Features[featureID]
	if !ok {
		return 0, 0
	}
	// Primary: parse test_results hash "passed:N failed:M"
	if tr, ok := fs.Hashes["test_results"]; ok {
		var p, f int
		fmt.Sscanf(tr, "passed:%d failed:%d", &p, &f)
		if p+f > 0 {
			return p + f, p
		}
	}
	// Fallback: count test file mappings
	if tests, ok := fs.Tests.([]string); ok {
		return len(tests), 0
	}
	return 0, 0
}

// AddFeatureWithPipeline creates a feature with a specific pipeline profile.
func AddFeatureWithPipeline(projectDir string, id string, title string, pipeline string) error {
	if pipeline != "" && !ValidPipelines[pipeline] {
		return fmt.Errorf("err:validation invalid pipeline %q: must be full|standard|lite", pipeline)
	}

	if !validFeatureID.MatchString(id) {
		return fmt.Errorf("err:validation invalid feature ID %q: must be ASCII slug (a-z0-9 with hyphens)", id)
	}

	features, err := loadFeatures(projectDir)
	if err != nil {
		return err
	}

	for _, f := range features {
		if f.ID == id {
			return fmt.Errorf("err:validation feature %s already exists", id)
		}
	}

	features = append(features, Feature{ID: id, Title: title, Status: "planned", Pipeline: pipeline})
	return saveFeatures(projectDir, features)
}

// UpdateFeaturePipeline changes a feature's pipeline profile.
func UpdateFeaturePipeline(projectDir string, id string, pipeline string) error {
	if !ValidPipelines[pipeline] {
		return fmt.Errorf("err:validation invalid pipeline %q: must be full|standard|lite", pipeline)
	}

	features, err := loadFeatures(projectDir)
	if err != nil {
		return err
	}

	idx := -1
	for i := range features {
		if features[i].ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("err:validation feature %s not found", id)
	}

	features[idx].Pipeline = pipeline
	return saveFeatures(projectDir, features)
}

func parseTestStatus(content string, featureID string) string {
	lines := strings.Split(content, "\n")
	inFeature := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == featureID+":" {
			inFeature = true
			continue
		}
		if inFeature && strings.HasPrefix(trimmed, "test_status: ") {
			return strings.TrimPrefix(trimmed, "test_status: ")
		}
		if inFeature && !strings.HasPrefix(line, "    ") && trimmed != "" {
			inFeature = false
		}
	}
	return ""
}
