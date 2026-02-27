package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAutoTrack_BDDAdvancesStage(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: seed\n    tests: absent\n    review: pending\n    issues: 0\n",
	), 0644)

	result, err := AutoTrack(dir, ".ptsd/bdd/auth.feature")
	if err != nil {
		t.Fatalf("AutoTrack: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.Updated {
		t.Error("expected updated=true")
	}
	if result.Stage != "bdd" {
		t.Errorf("expected stage=bdd, got %s", result.Stage)
	}
	if result.Feature != "auth" {
		t.Errorf("expected feature=auth, got %s", result.Feature)
	}
}

func TestAutoTrack_NeverRegresses(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	// Feature is at impl stage
	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: impl\n    tests: written\n    review: pending\n    issues: 0\n",
	), 0644)

	// Writing a BDD file should NOT regress stage from impl to bdd
	result, err := AutoTrack(dir, ".ptsd/bdd/auth.feature")
	if err != nil {
		t.Fatalf("AutoTrack: %v", err)
	}
	if result != nil && result.Updated && result.Stage == "bdd" {
		t.Error("auto-track should never regress stage from impl to bdd")
	}
}

func TestAutoTrack_TestSetsWritten(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: bdd\n    tests: absent\n    review: pending\n    issues: 0\n",
	), 0644)

	result, err := AutoTrack(dir, "internal/core/auth_test.go")
	if err != nil {
		t.Fatalf("AutoTrack: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Tests != "written" {
		t.Errorf("expected tests=written, got %s", result.Tests)
	}
}

func TestAutoTrack_SeedAdvancesStage(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: prd\n    tests: absent\n    review: pending\n    issues: 0\n",
	), 0644)

	result, err := AutoTrack(dir, ".ptsd/seeds/auth/seed.yaml")
	if err != nil {
		t.Fatalf("AutoTrack: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Stage != "seed" {
		t.Errorf("expected stage=seed, got %s", result.Stage)
	}
}

func TestAutoTrack_UnknownFileReturnsNil(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte("features: {}\n"), 0644)

	result, err := AutoTrack(dir, "README.md")
	if err != nil {
		t.Fatalf("AutoTrack: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for untracked file, got: %+v", result)
	}
}

func TestAutoTrack_Idempotent(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: bdd\n    tests: absent\n    review: pending\n    issues: 0\n",
	), 0644)

	// First call
	result1, _ := AutoTrack(dir, ".ptsd/bdd/auth.feature")
	if result1 != nil && result1.Updated {
		t.Error("BDD at bdd stage should not advance")
	}

	// Second call — same result
	result2, _ := AutoTrack(dir, ".ptsd/bdd/auth.feature")
	if result2 != nil && result2.Updated {
		t.Error("second call should also be no-op")
	}
}
