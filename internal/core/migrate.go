package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MigrateResult reports what MigrateProject did.
type MigrateResult struct {
	FeaturesUpdated  int
	ConfigUpdated    bool
	HooksRegenerated bool
}

// MigrateProject migrates an existing .ptsd/ project to the current version.
// - Adds pipeline field to features missing it
// - Adds pipeline section to ptsd.yaml if missing
// - Regenerates hooks, skills, CLAUDE.md
func MigrateProject(dir string) (*MigrateResult, error) {
	ptsdDir := filepath.Join(dir, ".ptsd")
	if _, err := os.Stat(ptsdDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("err:validation no .ptsd/ directory -- run ptsd init first")
	}

	result := &MigrateResult{}

	// 1. Migrate features.yaml -- add pipeline field where missing
	featuresUpdated, err := migrateFeatures(dir)
	if err != nil {
		return nil, err
	}
	result.FeaturesUpdated = featuresUpdated

	// 2. Migrate ptsd.yaml -- add pipeline section if missing
	configUpdated, err := migrateConfig(dir)
	if err != nil {
		return nil, err
	}
	result.ConfigUpdated = configUpdated

	// 3. Regenerate hooks, skills, CLAUDE.md
	if err := ReInitProject(dir); err != nil {
		return nil, err
	}
	result.HooksRegenerated = true

	return result, nil
}

// migrateFeatures adds pipeline: standard to features that don't have one.
func migrateFeatures(dir string) (int, error) {
	features, err := loadFeatures(dir)
	if err != nil {
		return 0, err
	}

	updated := 0
	for i := range features {
		if features[i].Pipeline == "" {
			features[i].Pipeline = "standard"
			updated++
		}
	}

	if updated > 0 {
		if err := saveFeatures(dir, features); err != nil {
			return 0, err
		}
	}

	return updated, nil
}

// migrateConfig adds pipeline section and tool field to ptsd.yaml if missing.
func migrateConfig(dir string) (bool, error) {
	cfgPath := filepath.Join(dir, ".ptsd", "ptsd.yaml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return false, fmt.Errorf("err:io %w", err)
	}

	content := string(data)
	changed := false

	// Add pipeline section if missing
	if !strings.Contains(content, "pipeline:") {
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "pipeline:\n  default: standard\n"
		changed = true
	}

	// Add tool field under project section if missing
	if !strings.Contains(content, "tool:") {
		tool := "claude"
		if _, err := os.Stat(filepath.Join(dir, ".opencode")); err == nil {
			tool = "opencode"
		} else if _, err := os.Stat(filepath.Join(dir, ".claude")); err != nil {
			tool = "generic"
		}
		content = strings.Replace(content, "project:\n", "project:\n  tool: "+tool+"\n", 1)
		changed = true
	}

	if changed {
		if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
			return false, fmt.Errorf("err:io %w", err)
		}
	}

	return changed, nil
}
