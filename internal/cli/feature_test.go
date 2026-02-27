package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupPTSDDir creates a minimal .ptsd/ directory structure in dir with
// ptsd.yaml, features.yaml, and state.yaml. It also runs git init so that
// any code that walks up looking for .git does not escape the temp tree.
func setupPTSDDir(t *testing.T, dir string) {
	t.Helper()

	ptsdDir := filepath.Join(dir, ".ptsd")
	for _, sub := range []string{"", "bdd", "seeds", "docs"} {
		if err := os.MkdirAll(filepath.Join(ptsdDir, sub), 0755); err != nil {
			t.Fatalf("MkdirAll %s: %v", sub, err)
		}
	}

	ptsdYAML := `project:
  name: "TestProject"

testing:
  runner: "go test ./..."
  patterns:
    files: ["**/*_test.go"]

review:
  min_score: 7
  auto_redo: false

hooks:
  pre_commit: true
  scopes: [PRD, SEED, BDD, TEST, IMPL]
  types: [feat, fix, add, refactor]
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(ptsdYAML), 0644); err != nil {
		t.Fatalf("write ptsd.yaml: %v", err)
	}

	featuresYAML := "features:\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresYAML), 0644); err != nil {
		t.Fatalf("write features.yaml: %v", err)
	}

	stateYAML := "features:\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	// git init so that directory traversal stops here.
	cmd := exec.Command("git", "init", dir)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		t.Logf("git init warning (non-fatal): %v", err)
	}
}

// captureStdout redirects os.Stdout around fn and returns whatever was printed.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	r.Close()
	return buf.String()
}

// chdir changes the working directory for the duration of the test and
// restores it via t.Cleanup.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%s): %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restore Chdir: %v", err)
		}
	})
}

// ---- RunFeature tests -------------------------------------------------------

func TestRunFeature_NoSubcommand_Exit2(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunFeature([]string{}, true)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunFeature_UnknownSubcommand_Exit2(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunFeature([]string{"unknown"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunFeature_Add_Success(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	out := captureStdout(t, func() {
		RunFeature([]string{"add", "my-feat", "My Feature"}, true)
	})

	exitCode := RunFeature([]string{"list"}, true)
	if exitCode != 0 {
		t.Fatalf("expected exit 0 after add, got %d", exitCode)
	}

	if !strings.Contains(out, "my-feat") {
		t.Errorf("expected output to contain 'my-feat', got: %q", out)
	}
}

func TestRunFeature_Add_ExitCode0(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunFeature([]string{"add", "my-feat", "My Feature"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunFeature_Add_AgentModeOutput(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	out := captureStdout(t, func() {
		RunFeature([]string{"add", "my-feat", "My Feature"}, true)
	})

	if !strings.Contains(out, "feature.add") {
		t.Errorf("agent mode: expected 'feature.add' in output, got: %q", out)
	}
	if !strings.Contains(out, "my-feat") {
		t.Errorf("agent mode: expected 'my-feat' in output, got: %q", out)
	}
}

func TestRunFeature_Add_HumanModeOutput(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	out := captureStdout(t, func() {
		RunFeature([]string{"add", "my-feat", "My Feature"}, false)
	})

	if !strings.Contains(out, "my-feat") {
		t.Errorf("human mode: expected 'my-feat' in output, got: %q", out)
	}
}

func TestRunFeature_Add_TooFewArgs_Exit2(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	// Only id, no title
	code := RunFeature([]string{"add", "my-feat"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunFeature_Add_DuplicateId_Error(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	out := captureStderr(t, func() {
		RunFeature([]string{"add", "my-feat", "Duplicate"}, true)
	})
	exitCode := RunFeature([]string{"add", "my-feat", "Duplicate"}, true)

	// duplicate feature is a validation error — exit code 1
	if exitCode != 1 {
		t.Errorf("expected exit code 1 for duplicate feature, got %d", exitCode)
	}
	if !strings.Contains(out, "err:") {
		t.Errorf("expected error output for duplicate, got: %q", out)
	}
}

func TestRunFeature_List_Empty(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunFeature([]string{"list"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 for empty list, got %d", code)
	}
}

func TestRunFeature_List_ShowsFeatures(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "feat-alpha", "Alpha"}, true)
	RunFeature([]string{"add", "feat-beta", "Beta"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"list"}, true)
	})

	if !strings.Contains(out, "feat-alpha") {
		t.Errorf("expected 'feat-alpha' in list output, got: %q", out)
	}
	if !strings.Contains(out, "feat-beta") {
		t.Errorf("expected 'feat-beta' in list output, got: %q", out)
	}
}

func TestRunFeature_List_AgentFormat(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "feat-alpha", "Alpha Feature"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"list"}, true)
	})

	// Agent format: "<id> [<status>] <title>"
	if !strings.Contains(out, "feat-alpha") {
		t.Errorf("expected feature id in output, got: %q", out)
	}
	if !strings.Contains(out, "[planned]") {
		t.Errorf("expected [planned] status in agent output, got: %q", out)
	}
}

func TestRunFeature_List_HumanMode(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "feat-alpha", "Alpha Feature"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"list"}, false)
	})

	if !strings.Contains(out, "feat-alpha") {
		t.Errorf("human mode list: expected 'feat-alpha' in output, got: %q", out)
	}
	if !strings.Contains(out, "planned") {
		t.Errorf("human mode list: expected status 'planned' in output, got: %q", out)
	}
}

func TestRunFeature_List_FilterByStatus(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "feat-alpha", "Alpha"}, true)
	RunFeature([]string{"add", "feat-beta", "Beta"}, true)
	// Update feat-alpha to in-progress
	RunFeature([]string{"status", "feat-alpha", "in-progress"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"list", "planned"}, true)
	})

	if strings.Contains(out, "feat-alpha") {
		t.Errorf("expected 'feat-alpha' (in-progress) to be filtered out, got: %q", out)
	}
	if !strings.Contains(out, "feat-beta") {
		t.Errorf("expected 'feat-beta' (planned) in filtered output, got: %q", out)
	}
}

func TestRunFeature_List_FilterMatchingNothing(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "feat-alpha", "Alpha"}, true)
	RunFeature([]string{"add", "feat-beta", "Beta"}, true)
	// Both features are 'planned'; filter by 'implemented' should return nothing.

	out := captureStdout(t, func() {
		RunFeature([]string{"list", "implemented"}, true)
	})

	if strings.Contains(out, "feat-alpha") || strings.Contains(out, "feat-beta") {
		t.Errorf("expected empty list for 'implemented' filter, got: %q", out)
	}
	code := RunFeature([]string{"list", "implemented"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 for filter with no matches, got %d", code)
	}
}

func TestRunFeature_Remove_Success(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	code := RunFeature([]string{"remove", "my-feat"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	out := captureStdout(t, func() {
		RunFeature([]string{"list"}, true)
	})
	if strings.Contains(out, "my-feat") {
		t.Errorf("expected 'my-feat' to be removed, still in list: %q", out)
	}
}

func TestRunFeature_Remove_AgentModeOutput(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"remove", "my-feat"}, true)
	})

	if !strings.Contains(out, "feature.remove") {
		t.Errorf("expected 'feature.remove' in agent output, got: %q", out)
	}
	if !strings.Contains(out, "my-feat") {
		t.Errorf("expected 'my-feat' in agent output, got: %q", out)
	}
}

func TestRunFeature_Remove_HumanMode(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"remove", "my-feat"}, false)
	})

	if !strings.Contains(out, "my-feat") {
		t.Errorf("human mode remove: expected 'my-feat' in output, got: %q", out)
	}
}

func TestRunFeature_Remove_NotFound_Error(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	// not found is a validation error — exit code 1
	code := RunFeature([]string{"remove", "nonexistent"}, true)
	if code != 1 {
		t.Errorf("expected exit code 1 for nonexistent feature, got %d", code)
	}
}

func TestRunFeature_Remove_TooFewArgs_Exit2(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunFeature([]string{"remove"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunFeature_Status_UpdateToInProgress(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	code := RunFeature([]string{"status", "my-feat", "in-progress"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	out := captureStdout(t, func() {
		RunFeature([]string{"list"}, true)
	})
	if !strings.Contains(out, "in-progress") {
		t.Errorf("expected 'in-progress' status in list, got: %q", out)
	}
}

func TestRunFeature_Status_AgentModeOutput(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"status", "my-feat", "in-progress"}, true)
	})

	if !strings.Contains(out, "feature.status") {
		t.Errorf("expected 'feature.status' in agent output, got: %q", out)
	}
	if !strings.Contains(out, "my-feat") {
		t.Errorf("expected 'my-feat' in agent output, got: %q", out)
	}
	if !strings.Contains(out, "in-progress") {
		t.Errorf("expected 'in-progress' in agent output, got: %q", out)
	}
}

func TestRunFeature_Status_HumanMode(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"status", "my-feat", "in-progress"}, false)
	})

	if !strings.Contains(out, "my-feat") {
		t.Errorf("human mode status: expected 'my-feat' in output, got: %q", out)
	}
	if !strings.Contains(out, "in-progress") {
		t.Errorf("human mode status: expected 'in-progress' in output, got: %q", out)
	}
}

func TestRunFeature_Status_InvalidStatus_Error(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	// invalid status is a validation error — exit code 1
	code := RunFeature([]string{"status", "my-feat", "invalid-status"}, true)
	if code != 1 {
		t.Errorf("expected exit code 1 for invalid status, got %d", code)
	}
}

func TestRunFeature_Status_TooFewArgs_Exit2(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunFeature([]string{"status", "my-feat"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunFeature_Status_ImplementedRequiresPassingTests(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	// Write failing test state
	stateContent := "features:\n  my-feat:\n    tests: 2\n    test_status: failing\n"
	statePath := filepath.Join(dir, ".ptsd", "state.yaml")
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	out := captureStderr(t, func() {
		RunFeature([]string{"status", "my-feat", "implemented"}, true)
	})
	// pipeline error — exit code 1
	exitCode := RunFeature([]string{"status", "my-feat", "implemented"}, true)

	if exitCode != 1 {
		t.Errorf("expected exit code 1 when tests are failing, got %d", exitCode)
	}
	if !strings.Contains(out, "err:pipeline") {
		t.Errorf("expected 'err:pipeline' in output, got: %q", out)
	}
}

func TestRunFeature_Show_BasicDetails(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	code := RunFeature([]string{"show", "my-feat"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunFeature_Show_OutputContainsID(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"show", "my-feat"}, true)
	})

	if !strings.Contains(out, "my-feat") {
		t.Errorf("expected 'my-feat' in show output, got: %q", out)
	}
}

func TestRunFeature_Show_OutputContainsTestCount(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"show", "my-feat"}, true)
	})

	// Agent mode renders test count as TEST:<n>/<total>
	if !strings.Contains(out, "TEST:") {
		t.Errorf("expected 'TEST:' field in show output, got: %q", out)
	}
}

func TestRunFeature_Show_HumanMode(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	out := captureStdout(t, func() {
		RunFeature([]string{"show", "my-feat"}, false)
	})

	if !strings.Contains(out, "my-feat") {
		t.Errorf("human mode show: expected 'my-feat' in output, got: %q", out)
	}
}

func TestRunFeature_Show_WithArtifacts(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	RunFeature([]string{"add", "my-feat", "My Feature"}, true)

	// Add BDD scenarios
	bddContent := "Feature: my-feat\nScenario: first scenario\nScenario: second scenario\n"
	bddPath := filepath.Join(dir, ".ptsd", "bdd", "my-feat.feature")
	if err := os.WriteFile(bddPath, []byte(bddContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Add seed directory
	seedDir := filepath.Join(dir, ".ptsd", "seeds", "my-feat")
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Add PRD anchor
	prdContent := "# PRD\n<!-- feature:my-feat -->\nSection content\n"
	prdPath := filepath.Join(dir, ".ptsd", "docs", "PRD.md")
	if err := os.WriteFile(prdPath, []byte(prdContent), 0644); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		RunFeature([]string{"show", "my-feat"}, true)
	})

	if !strings.Contains(out, "my-feat") {
		t.Errorf("expected 'my-feat' in show output, got: %q", out)
	}
	if !strings.Contains(out, "planned") {
		t.Errorf("expected status 'planned' in show output, got: %q", out)
	}
	// BDD count should be 2
	if !strings.Contains(out, "2scn") {
		t.Errorf("expected '2scn' in show output, got: %q", out)
	}
	// Seed present
	if !strings.Contains(out, "ok") {
		t.Errorf("expected seed status 'ok' in show output, got: %q", out)
	}
	// Test count field must be present
	if !strings.Contains(out, "TEST:") {
		t.Errorf("expected 'TEST:' field in show output with artifacts, got: %q", out)
	}
}

func TestRunFeature_Show_NotFound_Error(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	// not found is a validation error — exit code 1
	code := RunFeature([]string{"show", "nonexistent"}, true)
	if code != 1 {
		t.Errorf("expected exit code 1 for nonexistent feature, got %d", code)
	}
}

func TestRunFeature_Show_TooFewArgs_Exit2(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunFeature([]string{"show"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunFeature_Add_MultiWordTitle(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	code := RunFeature([]string{"add", "my-feat", "My", "Multi", "Word", "Title"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	out := captureStdout(t, func() {
		RunFeature([]string{"list"}, true)
	})
	if !strings.Contains(out, "My Multi Word Title") {
		t.Errorf("expected multi-word title in list, got: %q", out)
	}
}

func TestRunFeature_AddAndList_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	setupPTSDDir(t, dir)
	chdir(t, dir)

	ids := []string{"feat-one", "feat-two", "feat-three"}
	for _, id := range ids {
		code := RunFeature([]string{"add", id, fmt.Sprintf("Title for %s", id)}, true)
		if code != 0 {
			t.Fatalf("add %s: expected exit 0, got %d", id, code)
		}
	}

	out := captureStdout(t, func() {
		RunFeature([]string{"list"}, true)
	})

	for _, id := range ids {
		if !strings.Contains(out, id) {
			t.Errorf("expected %q in list output, got: %q", id, out)
		}
	}
}
