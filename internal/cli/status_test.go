package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupStatusProject creates a minimal .ptsd project in dir and returns dir.
// It includes ptsd.yaml, features.yaml (with one feature), state.yaml,
// docs/PRD.md with an anchor, bdd/, and seeds/ directories.
// No BDD files or seed data are written so that the feature is at the prd stage.
func setupStatusProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")

	dirs := []string{
		ptsdDir,
		filepath.Join(ptsdDir, "bdd"),
		filepath.Join(ptsdDir, "seeds"),
		filepath.Join(ptsdDir, "docs"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644); err != nil {
		t.Fatalf("write ptsd.yaml: %v", err)
	}

	featuresContent := "features:\n  - id: alpha\n    title: Alpha Feature\n    status: planned\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresContent), 0644); err != nil {
		t.Fatalf("write features.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	prdContent := "# PRD\n\n<!-- feature:alpha -->\n\n### Alpha\n\nDescription.\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "docs", "PRD.md"), []byte(prdContent), 0644); err != nil {
		t.Fatalf("write PRD.md: %v", err)
	}

	// Write a minimal tasks.yaml so ListTasks does not fail
	if err := os.WriteFile(filepath.Join(ptsdDir, "tasks.yaml"), []byte("tasks: []\n"), 0644); err != nil {
		t.Fatalf("write tasks.yaml: %v", err)
	}

	return dir
}

// chdirTo changes the working directory to dir and registers a cleanup that
// restores the original working directory.
func chdirTo(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restore chdir: %v", err)
		}
	})
}

func TestRunStatus_AgentMode(t *testing.T) {
	dir := setupStatusProject(t)
	chdirTo(t, dir)

	code := RunStatus([]string{}, true)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

func TestRunStatus_HumanMode(t *testing.T) {
	dir := setupStatusProject(t)
	chdirTo(t, dir)

	code := RunStatus([]string{}, false)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

func TestRunStatus_EmptyProject(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte("features:\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	chdirTo(t, dir)

	code := RunStatus([]string{}, true)
	if code != 0 {
		t.Errorf("expected exit 0 for empty project, got %d", code)
	}
}

func TestRunStatus_WithBDDAndTests(t *testing.T) {
	dir := setupStatusProject(t)
	ptsdDir := filepath.Join(dir, ".ptsd")

	// Add BDD file so the feature shows BDD coverage
	bddContent := "@feature:alpha\nFeature: Alpha Feature\n  Scenario: basic\n    Given something\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "bdd", "alpha.feature"), []byte(bddContent), 0644); err != nil {
		t.Fatalf("write bdd: %v", err)
	}

	// Write state with bdd hash to simulate feature at bdd stage with hashes recorded
	stateContent := "features:\n  alpha:\n    stage: bdd\n    hashes:\n      bdd: abc123\n    scores: {}\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateContent), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	chdirTo(t, dir)

	code := RunStatus([]string{}, true)
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

func TestRunStatus_OutputFormat_AgentMode(t *testing.T) {
	dir := setupStatusProject(t)
	chdirTo(t, dir)

	// Capture stdout by redirecting
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	code := RunStatus([]string{}, true)

	w.Close()
	os.Stdout = origStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}

	// Agent mode output must contain the compact format tokens
	for _, token := range []string{"FEAT:", "BDD:", "TESTS:", "WIP:", "TODO:", "DONE:"} {
		if !strings.Contains(output, token) {
			t.Errorf("agent mode output missing %q, got: %q", token, output)
		}
	}
}

// TestRunStatus_FailingTests verifies that a feature recorded with no test hash
// causes the FAIL counter in agent-mode output to reflect the missing coverage.
// BDD scenario: "Status with failing tests".
func TestRunStatus_FailingTests(t *testing.T) {
	dir := setupStatusProject(t)
	ptsdDir := filepath.Join(dir, ".ptsd")

	// Write state.yaml: one feature at impl stage with a bdd hash but NO test hash.
	// The status builder counts features without a "test" hash as TestFail.
	stateContent := "features:\n  alpha:\n    stage: impl\n    hashes:\n      bdd: abc123\n    scores: {}\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateContent), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	chdirTo(t, dir)

	output := captureStdout(t, func() {
		RunStatus([]string{}, true)
	})

	// The alpha feature has no test hash → TestFail should be 1.
	// Agent output format: [TESTS:<total> FAIL:<fail>]
	// With one feature and no test hash: total=0, fail=1 → [TESTS:0 FAIL:1]
	if !strings.Contains(output, "FAIL:1") {
		t.Errorf("expected FAIL:1 in agent output when feature has no test coverage, got: %q", output)
	}
}

// TestRunStatus_OutputFormat_ExactValues verifies the exact bracketed format
// with correct numeric values for a known project state.
// Agent output: [FEAT:<n> FAIL:<f>] [BDD:<n> FAIL:<f>] [TESTS:<n> FAIL:<f>] [T:<n> WIP:<w> TODO:<t> DONE:<d>]
func TestRunStatus_OutputFormat_ExactValues(t *testing.T) {
	dir := setupStatusProject(t)
	ptsdDir := filepath.Join(dir, ".ptsd")

	// State: one feature with both bdd and test hashes recorded.
	stateContent := "features:\n  alpha:\n    stage: impl\n    hashes:\n      bdd: abc123\n      test: def456\n    scores: {}\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateContent), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}
	// Add two tasks to tasks.yaml: one WIP, one TODO.
	tasksContent := "tasks:\n  - id: T-1\n    status: WIP\n    priority: A\n    title: Do something\n  - id: T-2\n    status: TODO\n    priority: B\n    title: Do another thing\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "tasks.yaml"), []byte(tasksContent), 0644); err != nil {
		t.Fatalf("write tasks.yaml: %v", err)
	}

	chdirTo(t, dir)

	output := captureStdout(t, func() {
		code := RunStatus([]string{}, true)
		if code != 0 {
			t.Errorf("expected exit 0, got %d", code)
		}
	})

	// 1 feature total, 0 without stage (alpha has stage "impl") → FEAT:1 FAIL:0
	// 1 feature with bdd hash → BDD:1 FAIL:0
	// 1 feature with test hash → TESTS:1 FAIL:0
	// 2 tasks total, 1 WIP, 1 TODO, 0 DONE → T:2 WIP:1 TODO:1 DONE:0
	expected := fmt.Sprintf("[FEAT:%d FAIL:%d] [BDD:%d FAIL:%d] [TESTS:%d FAIL:%d] [T:%d WIP:%d TODO:%d DONE:%d]",
		1, 0,
		1, 0,
		1, 0,
		2, 1, 1, 0)
	if !strings.Contains(output, expected) {
		t.Errorf("agent output does not match expected format\nwant (substring): %q\ngot:              %q", expected, output)
	}
}
