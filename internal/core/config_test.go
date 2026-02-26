package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigFromCurrentDir(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	configYAML := `project:
  name: MyApp
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Project.Name != "MyApp" {
		t.Errorf("expected project name MyApp, got %s", cfg.Project.Name)
	}
}

func TestWalkUpToFindPtsd(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	configYAML := `project:
  name: RootApp
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	subdir := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(subdir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Project.Name != "RootApp" {
		t.Errorf("expected project name RootApp, got %s", cfg.Project.Name)
	}
}

func TestMissingConfig(t *testing.T) {
	dir := t.TempDir()

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	if !strings.HasPrefix(err.Error(), "err:config") {
		t.Errorf("expected error to start with 'err:config', got: %v", err)
	}
}

func TestDefaultsFillMissingSections(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	configYAML := `project:
  name: MinimalApp
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Testing.Patterns.Files) == 0 {
		t.Error("expected default testing.patterns.files")
	}
	if cfg.Review.MinScore != 7 {
		t.Errorf("expected review.min_score default 7, got %d", cfg.Review.MinScore)
	}
	if cfg.Hooks.PreCommit != true {
		t.Error("expected hooks.pre_commit default true")
	}
}

func TestInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	invalidYAML := `project:
  name: [broken
    unclosed
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(invalidYAML), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !strings.HasPrefix(err.Error(), "err:config") {
		t.Errorf("expected error to start with 'err:config', got: %v", err)
	}
}
