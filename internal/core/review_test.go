package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRecordReviewScore(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:planned")
	ptsdDir := filepath.Join(dir, ".ptsd")

	// Create initial state.yaml
	statePath := filepath.Join(ptsdDir, "state.yaml")
	stateContent := "features:\n  user-auth:\n    stage: prd\n    hashes: {}\n    scores: {}\n"
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	before := time.Now()
	err := RecordReview(dir, "user-auth", "prd", 8)
	if err != nil {
		t.Fatalf("RecordReview failed: %v", err)
	}

	// Read state.yaml and verify score and timestamp
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("cannot read state.yaml: %v", err)
	}
	content := string(data)

	// Verify score exists
	if !strings.Contains(content, "score: 8") {
		t.Errorf("state.yaml should contain 'score: 8', got:\n%s", content)
	}
	// Verify stage key exists
	if !strings.Contains(content, "prd:") {
		t.Errorf("state.yaml should contain 'prd:' key, got:\n%s", content)
	}
	// Verify timestamp format (contains T for ISO time)
	if !strings.Contains(content, "T") {
		t.Errorf("state.yaml should contain timestamp with 'T', got:\n%s", content)
	}
	// Verify timestamp is after 'before'
	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}
	feature, ok := state.Features["user-auth"]
	if !ok {
		t.Fatal("expected user-auth feature in state")
	}
	score, ok := feature.Scores["prd"]
	if !ok {
		t.Fatal("expected prd score in state")
	}
	if score.Value != 8 {
		t.Errorf("expected score 8, got %d", score.Value)
	}
	if score.Timestamp.Before(before) {
		t.Error("timestamp should be after RecordReview call")
	}
}

func TestScoreBelowThresholdBlocksProgression(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:planned")
	ptsdDir := filepath.Join(dir, ".ptsd")

	// Create ptsd.yaml with review.min_score = 7
	configPath := filepath.Join(ptsdDir, "ptsd.yaml")
	configContent := "review:\n  min_score: 7\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create state.yaml with user-auth prd score = 5
	statePath := filepath.Join(ptsdDir, "state.yaml")
	stateContent := "features:\n  user-auth:\n    stage: prd\n    hashes: {}\n    scores:\n      prd:\n        score: 5\n        at: \"2026-02-26T10:00:00Z\"\n"
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	passed, err := CheckReviewGate(dir, "user-auth", "prd")
	if err != nil {
		t.Fatalf("CheckReviewGate failed: %v", err)
	}
	if passed {
		t.Error("expected gate to block (score 5 < min 7), got passed=true")
	}
}

func TestScoreAboveThresholdAllowsProgression(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:planned")
	ptsdDir := filepath.Join(dir, ".ptsd")

	// Create ptsd.yaml with review.min_score = 7
	configPath := filepath.Join(ptsdDir, "ptsd.yaml")
	configContent := "review:\n  min_score: 7\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create state.yaml with user-auth prd score = 8
	statePath := filepath.Join(ptsdDir, "state.yaml")
	stateContent := "features:\n  user-auth:\n    stage: prd\n    hashes: {}\n    scores:\n      prd:\n        score: 8\n        at: \"2026-02-26T10:00:00Z\"\n"
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	passed, err := CheckReviewGate(dir, "user-auth", "prd")
	if err != nil {
		t.Fatalf("CheckReviewGate failed: %v", err)
	}
	if !passed {
		t.Error("expected gate to allow (score 8 >= min 7), got passed=false")
	}
}

func TestRecordReviewUpdatesReviewStatus(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:planned")
	ptsdDir := filepath.Join(dir, ".ptsd")

	// Create ptsd.yaml with review.min_score = 7
	configPath := filepath.Join(ptsdDir, "ptsd.yaml")
	configContent := "review:\n  min_score: 7\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create initial state.yaml
	statePath := filepath.Join(ptsdDir, "state.yaml")
	stateContent := "features:\n  user-auth:\n    stage: prd\n    hashes: {}\n    scores: {}\n"
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create initial review-status.yaml
	rsPath := filepath.Join(ptsdDir, "review-status.yaml")
	rsContent := "features:\n  user-auth:\n    stage: prd\n    tests: absent\n    review: pending\n    issues: 0\n"
	if err := os.WriteFile(rsPath, []byte(rsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Record a passing review (score 8 >= min 7)
	err := RecordReview(dir, "user-auth", "prd", 8)
	if err != nil {
		t.Fatalf("RecordReview failed: %v", err)
	}

	// Verify review-status.yaml was updated to passed
	data, err := os.ReadFile(rsPath)
	if err != nil {
		t.Fatalf("cannot read review-status.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "review: passed") {
		t.Errorf("review-status.yaml should contain 'review: passed', got:\n%s", content)
	}
	if !strings.Contains(content, "issues: 0") {
		t.Errorf("review-status.yaml should contain 'issues: 0', got:\n%s", content)
	}

	// Now record a failing review (score 3 < min 7)
	err = RecordReview(dir, "user-auth", "prd", 3)
	if err != nil {
		t.Fatalf("RecordReview failed: %v", err)
	}

	data, err = os.ReadFile(rsPath)
	if err != nil {
		t.Fatalf("cannot read review-status.yaml: %v", err)
	}
	content = string(data)

	if !strings.Contains(content, "review: failed") {
		t.Errorf("review-status.yaml should contain 'review: failed', got:\n%s", content)
	}
	if !strings.Contains(content, "issues: 1") {
		t.Errorf("review-status.yaml should contain 'issues: 1', got:\n%s", content)
	}
	if !strings.Contains(content, "issues_list:") {
		t.Errorf("review-status.yaml should contain 'issues_list:', got:\n%s", content)
	}
}

func TestRecordReviewCreatesReviewStatusIfMissing(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:planned")
	ptsdDir := filepath.Join(dir, ".ptsd")

	// Create state.yaml but NO review-status.yaml
	statePath := filepath.Join(ptsdDir, "state.yaml")
	stateContent := "features:\n  user-auth:\n    stage: impl\n    hashes: {}\n    scores: {}\n"
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := RecordReview(dir, "user-auth", "impl", 9)
	if err != nil {
		t.Fatalf("RecordReview failed: %v", err)
	}

	// Verify review-status.yaml was created
	rsPath := filepath.Join(ptsdDir, "review-status.yaml")
	data, err := os.ReadFile(rsPath)
	if err != nil {
		t.Fatalf("review-status.yaml should have been created: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "user-auth:") {
		t.Errorf("review-status.yaml should contain 'user-auth:', got:\n%s", content)
	}
	if !strings.Contains(content, "review: passed") {
		t.Errorf("review-status.yaml should contain 'review: passed', got:\n%s", content)
	}
}

func TestCheckReviewGateNoConfig(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:planned")
	ptsdDir := filepath.Join(dir, ".ptsd")

	// No ptsd.yaml — CheckReviewGate must fall back to default min_score=7.

	// Score 8 >= 7 default: should pass.
	statePath := filepath.Join(ptsdDir, "state.yaml")
	stateContent := "features:\n  user-auth:\n    stage: prd\n    hashes: {}\n    scores:\n      prd:\n        score: 8\n        at: \"2026-02-26T10:00:00Z\"\n"
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	passed, err := CheckReviewGate(dir, "user-auth", "prd")
	if err != nil {
		t.Fatalf("CheckReviewGate should not error without config, got: %v", err)
	}
	if !passed {
		t.Error("expected gate to pass (score 8 >= default min 7), got passed=false")
	}

	// Score 5 < 7 default: should fail.
	stateContent = "features:\n  user-auth:\n    stage: prd\n    hashes: {}\n    scores:\n      prd:\n        score: 5\n        at: \"2026-02-26T10:00:00Z\"\n"
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	passed, err = CheckReviewGate(dir, "user-auth", "prd")
	if err != nil {
		t.Fatalf("CheckReviewGate should not error without config, got: %v", err)
	}
	if passed {
		t.Error("expected gate to block (score 5 < default min 7), got passed=true")
	}
}

func TestRecordReviewAdvancesStage(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:planned")
	ptsdDir := filepath.Join(dir, ".ptsd")

	// state.yaml with empty stage
	statePath := filepath.Join(ptsdDir, "state.yaml")
	os.WriteFile(statePath, []byte("features:\n  user-auth:\n    stage: \n    hashes: {}\n    scores: {}\n"), 0644)

	err := RecordReview(dir, "user-auth", "bdd", 9)
	if err != nil {
		t.Fatalf("RecordReview: %v", err)
	}

	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	fs := state.Features["user-auth"]
	if fs.Stage != "bdd" {
		t.Errorf("expected stage=bdd after reviewing bdd, got %q", fs.Stage)
	}
}

func TestRecordReviewNeverRegressesStage(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:planned")
	ptsdDir := filepath.Join(dir, ".ptsd")

	// state.yaml at impl stage
	statePath := filepath.Join(ptsdDir, "state.yaml")
	os.WriteFile(statePath, []byte("features:\n  user-auth:\n    stage: impl\n    hashes: {}\n    scores: {}\n"), 0644)

	err := RecordReview(dir, "user-auth", "prd", 8)
	if err != nil {
		t.Fatalf("RecordReview: %v", err)
	}

	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	fs := state.Features["user-auth"]
	if fs.Stage != "impl" {
		t.Errorf("stage should remain impl, got %q", fs.Stage)
	}
}

func TestAutoRedoTaskOnLowScore(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:planned")
	ptsdDir := filepath.Join(dir, ".ptsd")

	// Create ptsd.yaml with review.min_score = 7 and auto_redo = true
	configPath := filepath.Join(ptsdDir, "ptsd.yaml")
	configContent := "review:\n  min_score: 7\n  auto_redo: true\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create state.yaml with user-auth
	statePath := filepath.Join(ptsdDir, "state.yaml")
	stateContent := "features:\n  user-auth:\n    stage: prd\n    hashes: {}\n    scores: {}\n"
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create tasks.yaml
	tasksPath := filepath.Join(ptsdDir, "tasks.yaml")
	tasksContent := "tasks: []\n"
	if err := os.WriteFile(tasksPath, []byte(tasksContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Record review with low score (5 < threshold 7)
	err := RecordReview(dir, "user-auth", "prd", 5)
	if err != nil {
		t.Fatalf("RecordReview failed: %v", err)
	}

	// Read tasks.yaml and verify redo task was created
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		t.Fatalf("cannot read tasks.yaml: %v", err)
	}
	content := string(data)

	// Verify task exists for user-auth
	if !strings.Contains(content, "user-auth") {
		t.Errorf("tasks.yaml should contain task for user-auth, got:\n%s", content)
	}
	// Verify it's a redo task
	if !strings.Contains(content, "redo") && !strings.Contains(content, "retry") && !strings.Contains(content, "rework") {
		t.Errorf("tasks.yaml should contain redo/retry/rework task, got:\n%s", content)
	}
}
