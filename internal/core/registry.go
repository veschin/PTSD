package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Feature struct {
	ID     string
	Title  string
	Status string
}

type FeatureDetail struct {
	ID            string
	Status        string
	PRDAnchor     string
	SeedStatus    string
	ScenarioCount int
	TestCount     int
}

func AddFeature(projectDir string, id string, title string) error {
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
		ID:     found.ID,
		Status: found.Status,
	}

	// Seed status
	seedDir := filepath.Join(projectDir, ".ptsd", "seeds", id)
	if info, err := os.Stat(seedDir); err == nil && info.IsDir() {
		detail.SeedStatus = "ok"
	} else {
		detail.SeedStatus = "missing"
	}

	// BDD scenario count
	bddFile := filepath.Join(projectDir, ".ptsd", "bdd", id+".feature")
	if data, err := os.ReadFile(bddFile); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "Scenario:") {
				detail.ScenarioCount++
			}
		}
	}

	// Test count from state.yaml
	detail.TestCount = readTestCount(projectDir, id)

	// PRD anchor
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

	var filtered []Feature
	for _, f := range features {
		if f.ID != id {
			filtered = append(filtered, f)
		}
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
			// Read subsequent indented fields
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if strings.HasPrefix(next, "- id: ") || next == "" {
					break
				}
				if strings.HasPrefix(next, "title: ") {
					f.Title = strings.TrimPrefix(next, "title: ")
					// Remove surrounding quotes
					f.Title = strings.Trim(f.Title, "\"")
				}
				if strings.HasPrefix(next, "status: ") {
					f.Status = strings.TrimPrefix(next, "status: ")
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
		if strings.Contains(title, " ") {
			title = "\"" + title + "\""
		}
		b.WriteString("    title: " + title + "\n")
		b.WriteString("    status: " + f.Status + "\n")
	}

	return os.WriteFile(featPath, []byte(b.String()), 0644)
}

func readTestCount(projectDir string, featureID string) int {
	statePath := filepath.Join(projectDir, ".ptsd", "state.yaml")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(data), "\n")
	inFeature := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == featureID+":" {
			inFeature = true
			continue
		}
		if inFeature && strings.HasPrefix(trimmed, "tests: ") {
			val := strings.TrimPrefix(trimmed, "tests: ")
			n := 0
			for _, c := range val {
				if c >= '0' && c <= '9' {
					n = n*10 + int(c-'0')
				} else {
					break
				}
			}
			return n
		}
		if inFeature && !strings.HasPrefix(line, "    ") && trimmed != "" {
			inFeature = false
		}
	}
	return 0
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
