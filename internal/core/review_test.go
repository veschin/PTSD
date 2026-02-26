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
