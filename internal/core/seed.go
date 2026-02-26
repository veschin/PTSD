package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var validSeedTypes = map[string]bool{
	"data":    true,
	"fixture": true,
	"schema":  true,
}

func InitSeed(projectDir string, featureID string) error {
	features, err := loadFeatures(projectDir)
	if err != nil {
		return err
	}

	found := false
	for _, f := range features {
		if f.ID == featureID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("err:validation feature %s not found", featureID)
	}

	seedDir := filepath.Join(projectDir, ".ptsd", "seeds", featureID)
	seedPath := filepath.Join(seedDir, "seed.yaml")

	if _, err := os.Stat(seedPath); err == nil {
		return fmt.Errorf("err:validation seed already initialized for %s", featureID)
	}

	if err := os.MkdirAll(seedDir, 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	seedYAML := "feature: " + featureID + "\nfiles:\n"
	if err := os.WriteFile(seedPath, []byte(seedYAML), 0644); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	return nil
}

func AddSeedFile(projectDir string, featureID string, filePath string, fileType string, description string) error {
	if !validSeedTypes[fileType] {
		return fmt.Errorf("err:validation invalid seed type %q: must be data|fixture|schema", fileType)
	}

	seedDir := filepath.Join(projectDir, ".ptsd", "seeds", featureID)
	seedYAMLPath := filepath.Join(seedDir, "seed.yaml")

	if _, err := os.Stat(seedYAMLPath); os.IsNotExist(err) {
		return fmt.Errorf("err:validation seed not initialized for %s", featureID)
	}

	basename := filepath.Base(filePath)

	// Check for duplicate entry
	data, err := os.ReadFile(seedYAMLPath)
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}
	if strings.Contains(string(data), "path: "+basename) {
		return fmt.Errorf("err:validation file %s already in seed manifest for %s", basename, featureID)
	}

	dstPath := filepath.Join(seedDir, basename)

	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return fmt.Errorf("err:io %w", err)
	}
	if err := dst.Close(); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	content := string(data)
	entry := "  - path: " + basename + "\n    type: " + fileType + "\n"
	if description != "" {
		entry += "    description: \"" + description + "\"\n"
	}
	content += entry

	if err := os.WriteFile(seedYAMLPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	return nil
}

func CheckSeeds(projectDir string) ([]string, error) {
	features, err := loadFeatures(projectDir)
	if err != nil {
		return nil, err
	}

	var missing []string
	for _, f := range features {
		if f.Status == "planned" || f.Status == "deferred" {
			continue
		}
		seedPath := filepath.Join(projectDir, ".ptsd", "seeds", f.ID, "seed.yaml")
		if _, err := os.Stat(seedPath); os.IsNotExist(err) {
			missing = append(missing, f.ID)
		}
	}

	if len(missing) > 0 {
		msgs := make([]string, len(missing))
		for i, m := range missing {
			msgs[i] = m + " has no seed"
		}
		return missing, fmt.Errorf("err:pipeline %s", strings.Join(msgs, "; "))
	}

	return nil, nil
}
