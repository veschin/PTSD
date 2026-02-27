package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- RunConfig tests --------------------------------------------------------

func TestRunConfig_NoSubcommand_Exit2(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunConfig([]string{}, true)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunConfig_UnknownSubcommand_Exit2(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunConfig([]string{"unknown"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunConfig_Show_Exit0(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunConfig([]string{"show"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunConfig_Show_AgentModeOutput(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	out := captureStdout(t, func() {
		RunConfig([]string{"show"}, true)
	})

	// Agent mode prints key=value pairs
	if !strings.Contains(out, "project.name=") {
		t.Errorf("expected 'project.name=' in agent output, got: %q", out)
	}
	if !strings.Contains(out, "testing.runner=") {
		t.Errorf("expected 'testing.runner=' in agent output, got: %q", out)
	}
	if !strings.Contains(out, "review.min_score=") {
		t.Errorf("expected 'review.min_score=' in agent output, got: %q", out)
	}
	if !strings.Contains(out, "hooks.pre_commit=") {
		t.Errorf("expected 'hooks.pre_commit=' in agent output, got: %q", out)
	}
}

func TestRunConfig_Show_HumanModeOutput(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	out := captureStdout(t, func() {
		RunConfig([]string{"show"}, false)
	})

	// Human mode prints YAML-style
	if !strings.Contains(out, "project:") {
		t.Errorf("expected 'project:' section in human output, got: %q", out)
	}
	if !strings.Contains(out, "testing:") {
		t.Errorf("expected 'testing:' section in human output, got: %q", out)
	}
	if !strings.Contains(out, "review:") {
		t.Errorf("expected 'review:' section in human output, got: %q", out)
	}
	if !strings.Contains(out, "hooks:") {
		t.Errorf("expected 'hooks:' section in human output, got: %q", out)
	}
}

func TestRunConfig_Show_ReflectsProjectName(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	out := captureStdout(t, func() {
		RunConfig([]string{"show"}, true)
	})

	// setupPTSDDir writes project.name = "TestProject"
	if !strings.Contains(out, "TestProject") {
		t.Errorf("expected project name 'TestProject' in output, got: %q", out)
	}
}

func TestRunConfig_Show_ReflectsCustomConfig(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	customYAML := `project:
  name: "CustomApp"

testing:
  runner: "make test"
  patterns:
    files: ["**/*_test.go"]

review:
  min_score: 9
  auto_redo: true

hooks:
  pre_commit: false
  scopes: [PRD, BDD]
  types: [feat, fix]
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(customYAML), 0644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)

	out := captureStdout(t, func() {
		RunConfig([]string{"show"}, true)
	})

	if !strings.Contains(out, "CustomApp") {
		t.Errorf("expected 'CustomApp' in output, got: %q", out)
	}
	if !strings.Contains(out, "review.min_score=9") {
		t.Errorf("expected 'review.min_score=9' in output, got: %q", out)
	}
	if !strings.Contains(out, "review.auto_redo=true") {
		t.Errorf("expected 'review.auto_redo=true' in output, got: %q", out)
	}
	if !strings.Contains(out, "hooks.pre_commit=false") {
		t.Errorf("expected 'hooks.pre_commit=false' in output, got: %q", out)
	}
}

func TestRunConfig_Show_MissingConfig_Exit3(t *testing.T) {
	dir := t.TempDir()
	// No .ptsd/ directory — config load must fail with exit code 3.
	chdir(t, dir)

	out := captureStderr(t, func() {
		RunConfig([]string{"show"}, true)
	})
	exitCode := RunConfig([]string{"show"}, true)

	// config error — exit code 3
	if exitCode != 3 {
		t.Errorf("expected exit code 3 when config is missing, got %d", exitCode)
	}
	if !strings.Contains(out, "err:") {
		t.Errorf("expected error output, got: %q", out)
	}
}

func TestRunConfig_Show_DefaultValues(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Minimal config — defaults must be applied
	minimalYAML := `project:
  name: "Minimal"
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(minimalYAML), 0644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)

	out := captureStdout(t, func() {
		RunConfig([]string{"show"}, true)
	})

	// Default min_score = 7
	if !strings.Contains(out, "review.min_score=7") {
		t.Errorf("expected default 'review.min_score=7' in output, got: %q", out)
	}
	// Default pre_commit = true
	if !strings.Contains(out, "hooks.pre_commit=true") {
		t.Errorf("expected default 'hooks.pre_commit=true' in output, got: %q", out)
	}
	// Default patterns contain test pattern
	if !strings.Contains(out, "_test.go") {
		t.Errorf("expected default test pattern in output, got: %q", out)
	}
}

func TestRunConfig_NoSubcommand_OutputContainsError(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	out := captureStderr(t, func() {
		RunConfig([]string{}, true)
	})

	if !strings.Contains(out, "err:") {
		t.Errorf("expected error output for no subcommand, got: %q", out)
	}
}

func TestRunConfig_UnknownSubcommand_OutputContainsError(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	out := captureStderr(t, func() {
		RunConfig([]string{"get"}, true)
	})

	if !strings.Contains(out, "err:") {
		t.Errorf("expected error output for unknown subcommand, got: %q", out)
	}
	if !strings.Contains(out, "get") {
		t.Errorf("expected unknown subcommand name in error, got: %q", out)
	}
}

func TestRunConfig_Show_AllAgentFields(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	out := captureStdout(t, func() {
		RunConfig([]string{"show"}, true)
	})

	requiredFields := []string{
		"project.name=",
		"testing.runner=",
		"testing.patterns.files=",
		"testing.result_parser.format=",
		"review.min_score=",
		"review.auto_redo=",
		"hooks.pre_commit=",
		"hooks.scopes=",
		"hooks.types=",
	}
	for _, field := range requiredFields {
		if !strings.Contains(out, field) {
			t.Errorf("expected field %q in agent output, got: %q", field, out)
		}
	}
}

// TestRunConfig_WalkUpToFindPTSD verifies that RunConfig finds .ptsd/ when it
// is located two directories above the current working directory.
func TestRunConfig_WalkUpToFindPTSD(t *testing.T) {
	root := t.TempDir()
	setupPTSDDir(t, root)

	// Create a subdirectory two levels deep; .ptsd/ lives only in root.
	subDir := filepath.Join(root, "level1", "level2")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	// CWD is two directories below .ptsd/; config must still be found.
	chdir(t, subDir)

	out := captureStdout(t, func() {
		RunConfig([]string{"show"}, true)
	})
	code := RunConfig([]string{"show"}, true)

	if code != 0 {
		t.Errorf("expected exit code 0 when walking up to find .ptsd, got %d", code)
	}
	if !strings.Contains(out, "TestProject") {
		t.Errorf("expected project name 'TestProject' after walk-up, got: %q", out)
	}
}

// TestRunConfig_InvalidYAML_Exit3 verifies that a malformed ptsd.yaml returns
// exit code 3 (config error) with an err:config prefix in output.
func TestRunConfig_InvalidYAML_Exit3(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Unclosed bracket triggers err:config in parseConfig.
	invalidYAML := `project:
  name: "Test"
hooks:
  scopes: [PRD, BDD
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(invalidYAML), 0644); err != nil {
		t.Fatal(err)
	}
	chdir(t, dir)

	out := captureStderr(t, func() {
		RunConfig([]string{"show"}, true)
	})
	exitCode := RunConfig([]string{"show"}, true)

	if exitCode != 3 {
		t.Errorf("expected exit code 3 for invalid YAML, got %d", exitCode)
	}
	if !strings.Contains(out, "err:config") {
		t.Errorf("expected 'err:config' in output for invalid YAML, got: %q", out)
	}
}
