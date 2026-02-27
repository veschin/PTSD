package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func getPtsdBinary(t *testing.T) string {
	dir := t.TempDir()
	bin := filepath.Join(dir, "ptsd")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = filepath.Join("..", "..", "cmd", "ptsd")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build ptsd: %s %s", err, out)
	}
	return bin
}

func setupOutputProject(t *testing.T) string {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: 1\n"), 0644)
	os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte("features: []\n"), 0644)
	os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644)

	return dir
}

// Scenario: Unknown subcommand shows error
// Given an initialized project
// When I run "ptsd unknown-cmd"
// Then exit code is 2
// And output contains "err:user unknown command"
func TestMain_UnknownSubcommand(t *testing.T) {
	bin := getPtsdBinary(t)
	dir := setupOutputProject(t)
	cmd := exec.Command(bin, "unknown-cmd")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("expected non-zero exit, got 0. output: %s", out)
	}

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 2 {
		t.Errorf("expected exit 2, got %d. output: %s", exitCode, out)
	}

	if !strings.Contains(string(out), "err:user") {
		t.Errorf("expected err:user in output, got: %s", out)
	}
}

// Scenario: Missing subcommand shows usage
// Given an initialized project
// When I run "ptsd"
// Then exit code is 2
// And output contains "usage:"
func TestMain_MissingSubcommand(t *testing.T) {
	bin := getPtsdBinary(t)
	dir := setupOutputProject(t)
	cmd := exec.Command(bin)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("expected non-zero exit, got 0. output: %s", out)
	}

	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 2 {
		t.Errorf("expected exit 2, got %d. output: %s", exitCode, out)
	}

	if !strings.Contains(string(out), "usage:") {
		t.Errorf("expected usage in output, got: %s", out)
	}
}

// Scenario: help command shows all commands
// Given ptsd binary
// When I run "ptsd help"
// Then exit code is 0
// And output contains command names
func TestMain_Help(t *testing.T) {
	bin := getPtsdBinary(t)
	cmd := exec.Command(bin, "help")
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("expected exit 0, got error: %s %s", err, out)
	}

	output := string(out)
	for _, name := range []string{"init", "feature", "validate", "review", "context", "help"} {
		if !strings.Contains(output, name) {
			t.Errorf("help output missing command %q, got: %s", name, output)
		}
	}
}

// Scenario: --help flag shows help
// Given ptsd binary
// When I run "ptsd --help"
// Then exit code is 0
// And output matches "ptsd help"
func TestMain_HelpFlag(t *testing.T) {
	bin := getPtsdBinary(t)
	cmd := exec.Command(bin, "--help")
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("expected exit 0, got error: %s %s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "init") {
		t.Errorf("--help output missing commands, got: %s", output)
	}
}

// Scenario: --agent flag passed to handlers
// Given an initialized project
// When I run "ptsd config show --agent"
// Then agent mode output format is used
func TestMain_AgentFlag(t *testing.T) {
	bin := getPtsdBinary(t)
	dir := setupOutputProject(t)
	cmd := exec.Command(bin, "config", "show", "--agent")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("expected success, got error: %s", out)
	}

	if strings.Contains(string(out), "project:") {
		t.Errorf("expected agent mode (key=value), got formatted: %s", out)
	}
}

// Scenario: All subcommands route correctly
// Given an initialized project
// When I run "ptsd status"
// Then it executes RunStatus handler
func TestMain_StatusRouted(t *testing.T) {
	bin := getPtsdBinary(t)
	dir := setupOutputProject(t)
	cmd := exec.Command(bin, "status")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("expected success, got error: %s", out)
	}

	if !strings.Contains(string(out), "FEAT") && !strings.Contains(string(out), "Features") {
		t.Errorf("expected status output, got: %s", out)
	}
}
