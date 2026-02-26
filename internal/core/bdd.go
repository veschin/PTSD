package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FeatureFileData struct {
	Tag       string
	Title     string
	Scenarios []ScenarioData
}

type ScenarioData struct {
	Name  string
	Title string
	Steps []string
}

func AddBDD(projectDir string, featureID string) error {
	seedPath := filepath.Join(projectDir, ".ptsd", "seeds", featureID, "seed.yaml")
	if _, err := os.Stat(seedPath); os.IsNotExist(err) {
		return fmt.Errorf("err:pipeline %s has no seed", featureID)
	}

	bddDir := filepath.Join(projectDir, ".ptsd", "bdd")
	if err := os.MkdirAll(bddDir, 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	bddPath := filepath.Join(bddDir, featureID+".feature")
	content := "@feature:" + featureID + "\nFeature: " + featureID + "\n"
	if err := os.WriteFile(bddPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	return nil
}

func CheckBDD(projectDir string) ([]string, error) {
	features, err := loadFeatures(projectDir)
	if err != nil {
		return nil, err
	}

	featureSet := make(map[string]bool)
	for _, f := range features {
		featureSet[f.ID] = true
	}

	bddDir := filepath.Join(projectDir, ".ptsd", "bdd")
	entries, _ := os.ReadDir(bddDir)
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".feature") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(bddDir, e.Name()))
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "@feature:") {
				tag := strings.TrimPrefix(line, "@feature:")
				if !featureSet[tag] {
					return nil, fmt.Errorf("err:validation unknown feature tag %s", tag)
				}
			}
		}
	}

	var missing []string
	for _, f := range features {
		if f.Status == "planned" || f.Status == "deferred" {
			continue
		}
		bddPath := filepath.Join(bddDir, f.ID+".feature")
		if _, err := os.Stat(bddPath); os.IsNotExist(err) {
			missing = append(missing, f.ID)
		}
	}

	if len(missing) > 0 {
		return missing, fmt.Errorf("err:pipeline %s has no bdd", missing[0])
	}

	return nil, nil
}

func ShowBDD(projectDir string, featureID string) ([]string, error) {
	bddPath := filepath.Join(projectDir, ".ptsd", "bdd", featureID+".feature")
	data, err := os.ReadFile(bddPath)
	if err != nil {
		return nil, fmt.Errorf("err:validation feature %s not found", featureID)
	}

	ff, err := parseFeatureContent(string(data))
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, s := range ff.Scenarios {
		var parts []string
		for _, step := range s.Steps {
			parts = append(parts, step)
		}
		line := s.Name + ": " + strings.Join(parts, " / ")
		lines = append(lines, line)
	}

	return lines, nil
}

func ParseFeatureFile(path string) (FeatureFileData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FeatureFileData{}, fmt.Errorf("err:validation file not found: %s", path)
	}

	return parseFeatureContent(string(data))
}

func parseFeatureContent(content string) (FeatureFileData, error) {
	ff := FeatureFileData{}
	lines := strings.Split(content, "\n")

	var currentScenario *ScenarioData

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "@feature:") {
			ff.Tag = strings.TrimPrefix(trimmed, "@feature:")
			continue
		}

		if strings.HasPrefix(trimmed, "Feature:") {
			ff.Title = strings.TrimSpace(strings.TrimPrefix(trimmed, "Feature:"))
			continue
		}

		if strings.HasPrefix(trimmed, "Scenario:") {
			if currentScenario != nil {
				ff.Scenarios = append(ff.Scenarios, *currentScenario)
			}
			name := strings.TrimSpace(strings.TrimPrefix(trimmed, "Scenario:"))
			currentScenario = &ScenarioData{Name: name, Title: name}
			continue
		}

		if currentScenario != nil {
			if strings.HasPrefix(trimmed, "Given ") ||
				strings.HasPrefix(trimmed, "When ") ||
				strings.HasPrefix(trimmed, "Then ") ||
				strings.HasPrefix(trimmed, "And ") {
				currentScenario.Steps = append(currentScenario.Steps, trimmed)
			}
		}
	}

	if currentScenario != nil {
		ff.Scenarios = append(ff.Scenarios, *currentScenario)
	}

	return ff, nil
}
