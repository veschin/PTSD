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

func TestConfigAdapterSelectionTAPOverride(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Config with result_parser.format = tap overrides default adapter selection
	configYAML := `project:
  name: TAPApp
testing:
  runner: ./custom-runner.sh
  result_parser:
    format: tap
    root: results
    status_field: status
    passed_value: pass
    failed_value: fail
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify result_parser fields are parsed correctly
	if cfg.Testing.ResultParser.Format != "tap" {
		t.Errorf("expected result_parser.format 'tap', got %q", cfg.Testing.ResultParser.Format)
	}
	if cfg.Testing.ResultParser.Root != "results" {
		t.Errorf("expected result_parser.root 'results', got %q", cfg.Testing.ResultParser.Root)
	}
	if cfg.Testing.ResultParser.StatusField != "status" {
		t.Errorf("expected result_parser.status_field 'status', got %q", cfg.Testing.ResultParser.StatusField)
	}
	if cfg.Testing.ResultParser.PassedValue != "pass" {
		t.Errorf("expected result_parser.passed_value 'pass', got %q", cfg.Testing.ResultParser.PassedValue)
	}
	if cfg.Testing.ResultParser.FailedValue != "fail" {
		t.Errorf("expected result_parser.failed_value 'fail', got %q", cfg.Testing.ResultParser.FailedValue)
	}

	// When format is set, it signals TAP adapter override.
	// A custom runner without format would use exit-code adapter,
	// but format: tap forces TAP parsing.
	if cfg.Testing.Runner != "./custom-runner.sh" {
		t.Errorf("expected runner './custom-runner.sh', got %q", cfg.Testing.Runner)
	}

	// Verify the adapter selection logic: format != "" means TAP override
	if cfg.Testing.ResultParser.Format == "" {
		t.Error("expected non-empty format to signal TAP adapter override")
	}
}

func TestConfigAdapterSelectionNoTAPOverride(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Config without result_parser — custom runner uses exit-code adapter
	configYAML := `project:
  name: ExitCodeApp
testing:
  runner: ./custom-runner.sh
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Without result_parser.format, a custom runner should NOT trigger TAP override
	if cfg.Testing.ResultParser.Format != "" {
		t.Errorf("expected empty result_parser.format for exit-code adapter, got %q", cfg.Testing.ResultParser.Format)
	}
	if cfg.Testing.Runner != "./custom-runner.sh" {
		t.Errorf("expected runner './custom-runner.sh', got %q", cfg.Testing.Runner)
	}
}

func TestConfigFullIntegrationAllSections(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Complete config covering ALL sections and fields
	configYAML := `project:
  name: FullApp
testing:
  runner: go test ./...
  patterns:
    files: ["**/*_test.go", "**/*.spec.ts"]
  result_parser:
    format: tap
    root: testResults
    status_field: outcome
    passed_value: ok
    failed_value: error
review:
  min_score: 8
  auto_redo: true
hooks:
  pre_commit: false
  scopes: [PRD, SEED, BDD, TEST, IMPL]
  types: [feat, fix, refactor]
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// project section
	if cfg.Project.Name != "FullApp" {
		t.Errorf("project.name: expected 'FullApp', got %q", cfg.Project.Name)
	}

	// testing.runner
	if cfg.Testing.Runner != "go test ./..." {
		t.Errorf("testing.runner: expected 'go test ./...', got %q", cfg.Testing.Runner)
	}

	// testing.patterns.files
	if len(cfg.Testing.Patterns.Files) != 2 {
		t.Fatalf("testing.patterns.files: expected 2 entries, got %d", len(cfg.Testing.Patterns.Files))
	}
	if cfg.Testing.Patterns.Files[0] != "**/*_test.go" {
		t.Errorf("testing.patterns.files[0]: expected '**/*_test.go', got %q", cfg.Testing.Patterns.Files[0])
	}
	if cfg.Testing.Patterns.Files[1] != "**/*.spec.ts" {
		t.Errorf("testing.patterns.files[1]: expected '**/*.spec.ts', got %q", cfg.Testing.Patterns.Files[1])
	}

	// testing.result_parser
	if cfg.Testing.ResultParser.Format != "tap" {
		t.Errorf("testing.result_parser.format: expected 'tap', got %q", cfg.Testing.ResultParser.Format)
	}
	if cfg.Testing.ResultParser.Root != "testResults" {
		t.Errorf("testing.result_parser.root: expected 'testResults', got %q", cfg.Testing.ResultParser.Root)
	}
	if cfg.Testing.ResultParser.StatusField != "outcome" {
		t.Errorf("testing.result_parser.status_field: expected 'outcome', got %q", cfg.Testing.ResultParser.StatusField)
	}
	if cfg.Testing.ResultParser.PassedValue != "ok" {
		t.Errorf("testing.result_parser.passed_value: expected 'ok', got %q", cfg.Testing.ResultParser.PassedValue)
	}
	if cfg.Testing.ResultParser.FailedValue != "error" {
		t.Errorf("testing.result_parser.failed_value: expected 'error', got %q", cfg.Testing.ResultParser.FailedValue)
	}

	// review section
	if cfg.Review.MinScore != 8 {
		t.Errorf("review.min_score: expected 8, got %d", cfg.Review.MinScore)
	}
	if cfg.Review.AutoRedo != true {
		t.Error("review.auto_redo: expected true")
	}

	// hooks section
	if cfg.Hooks.PreCommit != false {
		t.Error("hooks.pre_commit: expected false")
	}
	if len(cfg.Hooks.Scopes) != 5 {
		t.Fatalf("hooks.scopes: expected 5 entries, got %d", len(cfg.Hooks.Scopes))
	}
	expectedScopes := []string{"PRD", "SEED", "BDD", "TEST", "IMPL"}
	for i, s := range expectedScopes {
		if cfg.Hooks.Scopes[i] != s {
			t.Errorf("hooks.scopes[%d]: expected %q, got %q", i, s, cfg.Hooks.Scopes[i])
		}
	}
	if len(cfg.Hooks.Types) != 3 {
		t.Fatalf("hooks.types: expected 3 entries, got %d", len(cfg.Hooks.Types))
	}
	expectedTypes := []string{"feat", "fix", "refactor"}
	for i, ty := range expectedTypes {
		if cfg.Hooks.Types[i] != ty {
			t.Errorf("hooks.types[%d]: expected %q, got %q", i, ty, cfg.Hooks.Types[i])
		}
	}
}
