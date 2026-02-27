package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupOutputProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte("features: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.Chdir(orig)
	})

	return dir
}

// TestAgentMode_ConfigShow verifies that --agent flag produces compact output.
func TestAgentMode_ConfigShow(t *testing.T) {
	setupOutputProject(t)

	out := captureStdout(t, func() {
		code := RunConfig([]string{"show"}, true)
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if strings.Contains(out, "\n\n") {
		t.Errorf("expected compact output, got multiline: %q", out)
	}
	if strings.Contains(out, "project:") {
		t.Errorf("expected key=value format, got: %q", out)
	}
}

// TestHumanMode_ConfigShow verifies that default mode produces formatted output.
func TestHumanMode_ConfigShow(t *testing.T) {
	setupOutputProject(t)

	out := captureStdout(t, func() {
		code := RunConfig([]string{"show"}, false)
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	})

	if !strings.Contains(out, "project:") {
		t.Errorf("expected formatted output, got: %q", out)
	}
}
