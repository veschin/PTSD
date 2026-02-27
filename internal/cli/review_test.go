package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupReviewProject creates a minimal .ptsd project suitable for review CLI tests.
// Returns the temp dir and a cleanup function that restores the original working directory.
func setupReviewProject(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	// ptsd.yaml with min_score=7
	configContent := "review:\n  min_score: 7\n  auto_redo: false\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// features.yaml
	featuresContent := "features:\n  - id: my-feat\n    status: planned\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresContent), 0644); err != nil {
		t.Fatal(err)
	}

	// state.yaml with my-feat
	stateContent := "features:\n  my-feat:\n    stage: impl\n    hashes: {}\n    scores: {}\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	// review-status.yaml
	rsContent := "features:\n  my-feat:\n    stage: impl\n    tests: written\n    review: pending\n    issues: 0\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "review-status.yaml"), []byte(rsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// tasks.yaml (needed for auto-redo path)
	tasksContent := "tasks: []\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "tasks.yaml"), []byte(tasksContent), 0644); err != nil {
		t.Fatal(err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	return dir, func() {
		if err := os.Chdir(orig); err != nil {
			t.Logf("warning: could not restore working directory: %v", err)
		}
	}
}

// TestRunReview_NoArgs verifies that RunReview with no args exits with code 2.
func TestRunReview_NoArgs(t *testing.T) {
	_, cleanup := setupReviewProject(t)
	defer cleanup()

	code := RunReview([]string{}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for no args, got %d", code)
	}
}

// TestRunReview_RecordHighScore verifies that recording a passing score (8 >= 7) exits 0
// and updates review-status.yaml to passed.
func TestRunReview_RecordHighScore(t *testing.T) {
	dir, cleanup := setupReviewProject(t)
	defer cleanup()

	code := RunReview([]string{"my-feat", "impl", "8"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Verify review-status.yaml updated to passed
	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "review-status.yaml"))
	if err != nil {
		t.Fatalf("cannot read review-status.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "review: passed") {
		t.Errorf("expected review: passed in review-status.yaml, got:\n%s", content)
	}
}

// TestRunReview_RecordLowScore verifies that recording a failing score (5 < 7) exits 0
// but sets verdict=fail in review-status.yaml.
func TestRunReview_RecordLowScore(t *testing.T) {
	dir, cleanup := setupReviewProject(t)
	defer cleanup()

	code := RunReview([]string{"my-feat", "impl", "5"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Verify review-status.yaml updated to failed
	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "review-status.yaml"))
	if err != nil {
		t.Fatalf("cannot read review-status.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "review: failed") {
		t.Errorf("expected review: failed in review-status.yaml, got:\n%s", content)
	}
	if !strings.Contains(content, "issues: 1") {
		t.Errorf("expected issues: 1 in review-status.yaml, got:\n%s", content)
	}
}

// TestRunReview_GatePass verifies that review gate returns exit 0 when score >= min_score.
func TestRunReview_GatePass(t *testing.T) {
	dir, cleanup := setupReviewProject(t)
	defer cleanup()

	// First record a passing score
	stateContent := "features:\n  my-feat:\n    stage: impl\n    hashes: {}\n    scores:\n      impl:\n        score: 8\n        at: \"2026-02-27T10:00:00Z\"\n"
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "state.yaml"), []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunReview([]string{"gate", "my-feat", "impl"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 for gate pass, got %d", code)
	}
}

// TestRunReview_GateFail verifies that review gate returns exit 1 when no score or score < min_score.
func TestRunReview_GateFail(t *testing.T) {
	dir, cleanup := setupReviewProject(t)
	defer cleanup()

	// State with low score for my-feat impl
	stateContent := "features:\n  my-feat:\n    stage: impl\n    hashes: {}\n    scores:\n      impl:\n        score: 4\n        at: \"2026-02-27T10:00:00Z\"\n"
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "state.yaml"), []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunReview([]string{"gate", "my-feat", "impl"}, true)
	if code != 1 {
		t.Errorf("expected exit code 1 for gate fail, got %d", code)
	}
}

// TestRunReview_GateMissingFeature verifies exit 1 when feature has no score at all.
func TestRunReview_GateMissingFeature(t *testing.T) {
	_, cleanup := setupReviewProject(t)
	defer cleanup()

	// state.yaml has my-feat with no scores; gate for unknown-feat should fail
	code := RunReview([]string{"gate", "unknown-feat", "impl"}, true)
	if code != 1 {
		t.Errorf("expected exit code 1 for missing feature gate, got %d", code)
	}
}

// TestRunReview_InvalidScore verifies that a non-integer score returns exit 2.
func TestRunReview_InvalidScore(t *testing.T) {
	_, cleanup := setupReviewProject(t)
	defer cleanup()

	code := RunReview([]string{"my-feat", "impl", "abc"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for non-integer score, got %d", code)
	}
}

// TestRunReview_InvalidStage verifies that an invalid stage returns a non-zero exit.
func TestRunReview_InvalidStage(t *testing.T) {
	_, cleanup := setupReviewProject(t)
	defer cleanup()

	code := RunReview([]string{"my-feat", "deploy", "8"}, true)
	if code == 0 {
		t.Error("expected non-zero exit code for invalid stage")
	}
}

// TestRunReview_GateMissingArgs verifies that gate with fewer than 2 args exits 2.
func TestRunReview_GateMissingArgs(t *testing.T) {
	_, cleanup := setupReviewProject(t)
	defer cleanup()

	code := RunReview([]string{"gate", "my-feat"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for gate with missing stage arg, got %d", code)
	}
}

// TestRunReview_RecordMissingArgs verifies that record with fewer than 3 args exits 2.
func TestRunReview_RecordMissingArgs(t *testing.T) {
	_, cleanup := setupReviewProject(t)
	defer cleanup()

	code := RunReview([]string{"my-feat", "impl"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for record with missing score arg, got %d", code)
	}
}

// TestRunReview_RecordLowScore_OutputContainsVerdictFail verifies that recording a failing score
// prints output containing "verdict:fail" (agent mode) as required by BDD "Score below threshold" scenario.
func TestRunReview_RecordLowScore_OutputContainsVerdictFail(t *testing.T) {
	_, cleanup := setupReviewProject(t)
	defer cleanup()

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunReview([]string{"my-feat", "impl", "5"}, true)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "verdict:fail") {
		t.Errorf("expected output to contain verdict:fail, got: %q", output)
	}
}

// TestRunReview_RecordHighScore_HumanMode verifies human mode output for a passing review.
func TestRunReview_RecordHighScore_HumanMode(t *testing.T) {
	_, cleanup := setupReviewProject(t)
	defer cleanup()

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunReview([]string{"my-feat", "impl", "8"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "my-feat") {
		t.Errorf("human mode output missing feature name, got: %q", output)
	}
	if !strings.Contains(output, "pass") {
		t.Errorf("human mode output missing verdict pass, got: %q", output)
	}
}

// TestRunReview_RecordLowScore_HumanMode verifies human mode output for a failing review.
func TestRunReview_RecordLowScore_HumanMode(t *testing.T) {
	_, cleanup := setupReviewProject(t)
	defer cleanup()

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunReview([]string{"my-feat", "impl", "5"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "fail") {
		t.Errorf("human mode output missing verdict fail, got: %q", output)
	}
}

// TestRunReview_GatePass_HumanMode verifies human mode output for a passing gate.
func TestRunReview_GatePass_HumanMode(t *testing.T) {
	dir, cleanup := setupReviewProject(t)
	defer cleanup()

	stateContent := "features:\n  my-feat:\n    stage: impl\n    hashes: {}\n    scores:\n      impl:\n        score: 8\n        at: \"2026-02-27T10:00:00Z\"\n"
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "state.yaml"), []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunReview([]string{"gate", "my-feat", "impl"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "pass") {
		t.Errorf("human mode gate output missing pass verdict, got: %q", output)
	}
}

// TestRunReview_GateFail_HumanMode verifies human mode output for a failing gate.
func TestRunReview_GateFail_HumanMode(t *testing.T) {
	dir, cleanup := setupReviewProject(t)
	defer cleanup()

	stateContent := "features:\n  my-feat:\n    stage: impl\n    hashes: {}\n    scores:\n      impl:\n        score: 4\n        at: \"2026-02-27T10:00:00Z\"\n"
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "state.yaml"), []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunReview([]string{"gate", "my-feat", "impl"}, false)
	})
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(output, "fail") {
		t.Errorf("human mode gate output missing fail verdict, got: %q", output)
	}
}

// TestRunReview_AutoRedoCreatesTask verifies that auto_redo=true + low score creates a redo task.
func TestRunReview_AutoRedoCreatesTask(t *testing.T) {
	dir, cleanup := setupReviewProject(t)
	defer cleanup()

	// Enable auto_redo
	configContent := "review:\n  min_score: 7\n  auto_redo: true\n"
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "ptsd.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunReview([]string{"my-feat", "impl", "3"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "tasks.yaml"))
	if err != nil {
		t.Fatalf("cannot read tasks.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "my-feat") {
		t.Errorf("expected redo task for my-feat in tasks.yaml, got:\n%s", content)
	}
}
