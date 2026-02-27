package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidCommitWithMatchingScope(t *testing.T) {
	dir := t.TempDir()
	setupPtsdConfig(t, dir, []string{"tests/**/*.test.ts"})

	stagedFiles := []string{".ptsd/bdd/login.feature"}
	err := ValidateCommit(dir, "[BDD] add: login scenarios", stagedFiles)
	if err != nil {
		t.Fatalf("expected hook to pass, got: %v", err)
	}
}

func TestScopeMismatch(t *testing.T) {
	dir := t.TempDir()
	setupPtsdConfig(t, dir, []string{"tests/**/*.test.ts"})

	stagedFiles := []string{"src/auth.ts", ".ptsd/bdd/auth.feature"}
	err := ValidateCommit(dir, "[BDD] add: auth scenarios", stagedFiles)
	if err == nil {
		t.Fatal("expected hook to fail for scope mismatch")
	}
	if !containsErr(err, "err:git") {
		t.Fatalf("expected err:git error, got: %v", err)
	}
}

func TestMissingScope(t *testing.T) {
	dir := t.TempDir()
	setupPtsdConfig(t, dir, []string{"tests/**/*.test.ts"})

	stagedFiles := []string{".ptsd/docs/PRD.md"}
	err := ValidateCommit(dir, "update PRD", stagedFiles)
	if err == nil {
		t.Fatal("expected hook to fail for missing scope")
	}
	if !containsErr(err, "err:git missing [SCOPE] in commit message") {
		t.Fatalf("expected 'err:git missing [SCOPE] in commit message', got: %v", err)
	}
}

func TestInvalidScope(t *testing.T) {
	dir := t.TempDir()
	setupPtsdConfig(t, dir, []string{"tests/**/*.test.ts"})

	stagedFiles := []string{"src/main.go"}
	err := ValidateCommit(dir, "[UNKNOWN] add: something", stagedFiles)
	if err == nil {
		t.Fatal("expected hook to fail for invalid scope")
	}
	if !containsErr(err, "err:git unknown scope UNKNOWN") {
		t.Fatalf("expected 'err:git unknown scope UNKNOWN', got: %v", err)
	}
}

func TestIMPLScopeTriggersFullValidation(t *testing.T) {
	dir := t.TempDir()
	setupPtsdConfig(t, dir, []string{"tests/**/*.test.ts"})

	// Set up features.yaml with an active feature
	ptsdDir := filepath.Join(dir, ".ptsd")
	featuresContent := "- id: auth\n  title: Authentication\n  status: active\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresContent), 0644); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	// Create BDD without seed — pipeline violation
	bddDir := filepath.Join(ptsdDir, "bdd")
	if err := os.MkdirAll(bddDir, 0755); err != nil {
		t.Fatalf("failed to create bdd dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bddDir, "auth.feature"), []byte("Feature: Auth\n"), 0644); err != nil {
		t.Fatalf("failed to write bdd file: %v", err)
	}

	// Create PRD with anchor so CheckPRDAnchors does not fail on missing file
	docsDir := filepath.Join(ptsdDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "PRD.md"), []byte("<!-- feature:auth -->\n"), 0644); err != nil {
		t.Fatalf("failed to write PRD.md: %v", err)
	}

	// IMPL scope must trigger full Validate which catches "bdd but no seed"
	stagedFiles := []string{"src/auth.go"}
	err := ValidateCommit(dir, "[IMPL] feat: user auth", stagedFiles)
	if err == nil {
		t.Fatal("expected IMPL scope to trigger full validation and catch pipeline violation")
	}
	if !containsErr(err, "err:pipeline") {
		t.Fatalf("expected err:pipeline, got: %v", err)
	}
	if !strings.Contains(err.Error(), "has bdd but no seed") {
		t.Fatalf("expected 'has bdd but no seed' violation, got: %v", err)
	}
}

func TestIMPLScopePassesWhenPipelineClean(t *testing.T) {
	dir := t.TempDir()
	setupPtsdConfig(t, dir, []string{"tests/**/*.test.ts"})

	// Set up features.yaml with an active feature
	ptsdDir := filepath.Join(dir, ".ptsd")
	featuresContent := "- id: auth\n  title: Authentication\n  status: active\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresContent), 0644); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	// Create seed + BDD + a test file — pipeline is clean
	seedDir := filepath.Join(ptsdDir, "seeds", "auth")
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		t.Fatalf("failed to create seed dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte("data: test\n"), 0644); err != nil {
		t.Fatalf("failed to write seed file: %v", err)
	}

	bddDir := filepath.Join(ptsdDir, "bdd")
	if err := os.MkdirAll(bddDir, 0755); err != nil {
		t.Fatalf("failed to create bdd dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bddDir, "auth.feature"), []byte("Feature: Auth\n"), 0644); err != nil {
		t.Fatalf("failed to write bdd file: %v", err)
	}

	// Create a test file so "has bdd but no tests" is not triggered
	if err := os.MkdirAll(filepath.Join(dir, "internal"), 0755); err != nil {
		t.Fatalf("failed to create internal dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal", "auth_test.go"), []byte("package internal\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create PRD with anchor for the feature
	docsDir := filepath.Join(ptsdDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "PRD.md"), []byte("<!-- feature:auth -->\n"), 0644); err != nil {
		t.Fatalf("failed to write PRD.md: %v", err)
	}

	stagedFiles := []string{"src/auth.go"}
	err := ValidateCommit(dir, "[IMPL] feat: user auth", stagedFiles)
	if err != nil {
		t.Fatalf("expected IMPL scope to pass with clean pipeline, got: %v", err)
	}
}

func TestTaskAndStatusScopesSkipPipelineValidation(t *testing.T) {
	dir := t.TempDir()
	setupPtsdConfig(t, dir, []string{"tests/**/*.test.ts"})

	stagedFiles := []string{".ptsd/tasks.yaml"}
	err := ValidateCommit(dir, "[TASK] add: new auth task", stagedFiles)
	if err != nil {
		t.Fatalf("expected TASK scope to pass without pipeline validation, got: %v", err)
	}

	// Test STATUS scope
	stagedFiles = []string{".ptsd/state.yaml"}
	err = ValidateCommit(dir, "[STATUS] update: task completed", stagedFiles)
	if err != nil {
		t.Fatalf("expected STATUS scope to pass without pipeline validation, got: %v", err)
	}
}

func TestFileClassificationByPath(t *testing.T) {
	dir := t.TempDir()
	setupPtsdConfig(t, dir, []string{"tests/**/*.test.ts"})

	tests := []struct {
		path     string
		expected string
	}{
		{".ptsd/docs/PRD.md", "PRD"},
		{".ptsd/seeds/auth/login.yaml", "SEED"},
		{".ptsd/bdd/auth.feature", "BDD"},
		{"tests/auth.test.ts", "TEST"},
		{"src/auth.go", "IMPL"},
		{"internal/core/hooks.go", "IMPL"},
	}

	for _, tt := range tests {
		class, err := ClassifyFile(dir, tt.path)
		if err != nil {
			t.Fatalf("ClassifyFile(%s) error: %v", tt.path, err)
		}
		if class != tt.expected {
			t.Errorf("ClassifyFile(%s) = %s, want %s", tt.path, class, tt.expected)
		}
	}
}

func TestParseCommitMessage(t *testing.T) {
	tests := []struct {
		msg          string
		scope        string
		commitType   string
		text         string
		shouldFail   bool
	}{
		{"[BDD] add: login scenarios", "BDD", "add", "login scenarios", false},
		{"[IMPL] feat: user auth", "IMPL", "feat", "user auth", false},
		{"[PRD] update: requirements", "PRD", "update", "requirements", false},
		{"[TEST] fix: flaky test", "TEST", "fix", "flaky test", false},
		{"[SEED] remove: old data", "SEED", "remove", "old data", false},
		{"[TASK] add: new task", "TASK", "add", "new task", false},
		{"[STATUS] update: done", "STATUS", "update", "done", false},
		{"update PRD", "", "", "", true},
		{"[UNKNOWN] add: x", "UNKNOWN", "add", "x", false},
	}

	for _, tt := range tests {
		scope, commitType, text, err := ParseCommitMessage(tt.msg)
		if tt.shouldFail {
			if err == nil {
				t.Errorf("ParseCommitMessage(%q) expected error, got none", tt.msg)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseCommitMessage(%q) unexpected error: %v", tt.msg, err)
			continue
		}
		if scope != tt.scope {
			t.Errorf("ParseCommitMessage(%q) scope = %q, want %q", tt.msg, scope, tt.scope)
		}
		if commitType != tt.commitType {
			t.Errorf("ParseCommitMessage(%q) type = %q, want %q", tt.msg, commitType, tt.commitType)
		}
		if text != tt.text {
			t.Errorf("ParseCommitMessage(%q) text = %q, want %q", tt.msg, text, tt.text)
		}
	}
}

func setupPtsdConfig(t *testing.T, dir string, testPatterns []string) {
	t.Helper()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatalf("failed to create .ptsd dir: %v", err)
	}

	// Create minimal ptsd.yaml with test patterns
	configContent := "testing:\n  patterns:\n    files:\n"
	for _, p := range testPatterns {
		configContent += "      - \"" + p + "\"\n"
	}

	configPath := filepath.Join(ptsdDir, "ptsd.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write ptsd.yaml: %v", err)
	}
}

func containsErr(err error, substr string) bool {
	return err != nil && strings.Contains(err.Error(), substr)
}
