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

	var problems []string
	for _, f := range features {
		if f.Status == "planned" || f.Status == "deferred" {
			continue
		}
		seedDir := filepath.Join(projectDir, ".ptsd", "seeds", f.ID)
		seedPath := filepath.Join(seedDir, "seed.yaml")
		if _, err := os.Stat(seedPath); os.IsNotExist(err) {
			problems = append(problems, f.ID+" has no seed")
			continue
		}

		// Verify files listed in manifest exist on disk
		data, err := os.ReadFile(seedPath)
		if err != nil {
			return nil, fmt.Errorf("err:io %w", err)
		}
		for _, file := range parseSeedManifestFiles(string(data)) {
			filePath := filepath.Join(seedDir, file)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				problems = append(problems, f.ID+" seed manifest references missing file: "+file)
			}
		}
	}

	if len(problems) > 0 {
		// Extract feature IDs from problems for backward-compatible return value
		seen := map[string]bool{}
		var missing []string
		for _, f := range features {
			for _, p := range problems {
				if strings.HasPrefix(p, f.ID+" ") && !seen[f.ID] {
					seen[f.ID] = true
					missing = append(missing, f.ID)
				}
			}
		}
		return missing, fmt.Errorf("err:pipeline %s", strings.Join(problems, "; "))
	}

	return nil, nil
}

// parseSeedManifestFiles extracts file paths from seed.yaml manifest lines like "  - path: foo.json".
func parseSeedManifestFiles(content string) []string {
	var files []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- path: ") {
			file := strings.TrimPrefix(trimmed, "- path: ")
			file = strings.TrimSpace(file)
			if file != "" {
				files = append(files, file)
			}
		}
	}
	return files
}
