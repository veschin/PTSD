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

	// Create BDD file on disk for hash computation
	os.WriteFile(filepath.Join(ptsd, "bdd", "auth.feature"), []byte("Feature: auth\n"), 0644)

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

	// Verify state.yaml was synced
	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	fs := state.Features["auth"]
	if fs.Stage != "bdd" {
		t.Errorf("state.yaml stage: got %q, want bdd", fs.Stage)
	}
	if fs.Hashes["bdd"] == "" {
		t.Error("state.yaml should have bdd hash")
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

	// Create test file on disk for hash computation
	testDir := filepath.Join(dir, "internal", "core")
	os.MkdirAll(testDir, 0755)
	os.WriteFile(filepath.Join(testDir, "auth_test.go"), []byte("package core\n"), 0644)

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

	// Verify state.yaml was synced
	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	fs := state.Features["auth"]
	if fs.Stage != "tests" {
		t.Errorf("state.yaml stage: got %q, want tests", fs.Stage)
	}
	if fs.Hashes["test"] == "" {
		t.Error("state.yaml should have test hash")
	}
}

func TestAutoTrack_SeedAdvancesStage(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: prd\n    tests: absent\n    review: pending\n    issues: 0\n",
	), 0644)

	// Create seed file on disk for hash computation
	seedDir := filepath.Join(ptsd, "seeds", "auth")
	os.MkdirAll(seedDir, 0755)
	os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte("data: x\n"), 0644)

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

	// Verify state.yaml was synced
	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	fs := state.Features["auth"]
	if fs.Stage != "seed" {
		t.Errorf("state.yaml stage: got %q, want seed", fs.Stage)
	}
	if fs.Hashes["seed"] == "" {
		t.Error("state.yaml should have seed hash")
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

func TestAutoTrack_CreatesEntryForNewFeature(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	// Empty review-status — no entry for auth
	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte("features: {}\n"), 0644)

	result, err := AutoTrack(dir, ".ptsd/bdd/auth.feature")
	if err != nil {
		t.Fatalf("AutoTrack: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for new feature entry")
	}
	if result.Stage != "bdd" {
		t.Errorf("expected stage=bdd for new entry, got %s", result.Stage)
	}
	if !result.Updated {
		t.Error("expected updated=true for new entry")
	}
}

func TestAutoTrack_PRDFileIgnored(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: prd\n    tests: absent\n    review: pending\n    issues: 0\n",
	), 0644)

	result, err := AutoTrack(dir, ".ptsd/docs/PRD.md")
	if err != nil {
		t.Fatalf("AutoTrack: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for PRD file, got: %+v", result)
	}
}

func TestAutoTrack_ImplAdvancesStage(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: tests\n    tests: written\n    review: pending\n    issues: 0\n",
	), 0644)

	result, err := AutoTrack(dir, "internal/core/auth.go")
	if err != nil {
		t.Fatalf("AutoTrack: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Stage != "impl" {
		t.Errorf("expected stage=impl, got %s", result.Stage)
	}

	// Verify state.yaml stage (impl has no hash)
	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState: %v", err)
	}
	fs := state.Features["auth"]
	if fs.Stage != "impl" {
		t.Errorf("state.yaml stage: got %q, want impl", fs.Stage)
	}
}

func TestAutoTrack_AbsolutePath(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: seed\n    tests: absent\n    review: pending\n    issues: 0\n",
	), 0644)

	absPath := filepath.Join(dir, ".ptsd", "bdd", "auth.feature")
	result, err := AutoTrack(dir, absPath)
	if err != nil {
		t.Fatalf("AutoTrack: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result for absolute path")
	}
	if result.Stage != "bdd" {
		t.Errorf("expected stage=bdd, got %s", result.Stage)
	}
}

func TestAutoTrack_ManagementFilesIgnored(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte("features: {}\n"), 0644)

	ignored := []string{
		".ptsd/tasks.yaml",
		".ptsd/state.yaml",
		".ptsd/ptsd.yaml",
		".ptsd/features.yaml",
		".ptsd/issues.yaml",
		"CLAUDE.md",
		".claude/settings.json",
	}

	for _, f := range ignored {
		result, err := AutoTrack(dir, f)
		if err != nil {
			t.Fatalf("AutoTrack(%s): %v", f, err)
		}
		if result != nil {
			t.Errorf("expected nil result for management file %q, got: %+v", f, result)
		}
	}
}
