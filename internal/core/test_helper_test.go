package core

import (
	"os"
	"path/filepath"
	"testing"
)

// setupProjectWithFeatures creates a .ptsd directory with features.yaml.
// Features are specified as "id:status" strings (status defaults to "planned").
func setupProjectWithFeatures(t *testing.T, features ...string) string {
	t.Helper()
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	yamlContent := "features:\n"
	for _, f := range features {
		id, status := "unknown", "planned"
		for i, c := range f {
			if c == ':' {
				id = f[:i]
				status = f[i+1:]
				break
			}
		}
		if id == "unknown" && f != "" {
			id = f
		}
		yamlContent += "  - id: " + id + "\n    status: " + status + "\n"
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	seedsDir := filepath.Join(ptsdDir, "seeds")
	if err := os.Mkdir(seedsDir, 0755); err != nil {
		t.Fatal(err)
	}
	bddDir := filepath.Join(ptsdDir, "bdd")
	if err := os.Mkdir(bddDir, 0755); err != nil {
		t.Fatal(err)
	}

	return dir
}
