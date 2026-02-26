package core

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreHashesOnSync(t *testing.T) {
	dir := t.TempDir()
	setupFeatureFiles(t, dir, "user-auth", "seed", "bdd", "test")

	err := SyncState(dir)
	if err != nil {
		t.Fatalf("SyncState failed: %v", err)
	}

	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	feature, ok := state.Features["user-auth"]
	if !ok {
		t.Fatal("expected user-auth feature in state")
	}

	seedPath := filepath.Join(dir, ".ptsd", "seeds", "user-auth", "seed.yaml")
	bddPath := filepath.Join(dir, ".ptsd", "bdd", "user-auth.feature")
	testPath := filepath.Join(dir, "internal", "core", "user-auth_test.go")

	expectedSeedHash := fileHash(t, seedPath)
	expectedBDDHash := fileHash(t, bddPath)
	expectedTestHash := fileHash(t, testPath)

	if feature.Hashes["seed"] != expectedSeedHash {
		t.Errorf("seed hash mismatch: got %s, want %s", feature.Hashes["seed"], expectedSeedHash)
	}
	if feature.Hashes["bdd"] != expectedBDDHash {
		t.Errorf("bdd hash mismatch: got %s, want %s", feature.Hashes["bdd"], expectedBDDHash)
	}
	if feature.Hashes["test"] != expectedTestHash {
		t.Errorf("test hash mismatch: got %s, want %s", feature.Hashes["test"], expectedTestHash)
	}
}

func TestBDDChangeForImplementedFeature(t *testing.T) {
	dir := t.TempDir()
	setupFeatureFiles(t, dir, "user-auth", "seed", "bdd", "test")
	setState(t, dir, "user-auth", "implemented", nil, nil)

	// Change BDD file
	bddPath := filepath.Join(dir, ".ptsd", "bdd", "user-auth.feature")
	appendFile(t, bddPath, "\n# new scenario")

	regressions, err := CheckRegressions(dir)
	if err != nil {
		t.Fatalf("CheckRegressions failed: %v", err)
	}

	if len(regressions) == 0 {
		t.Fatal("expected regression detected")
	}

	found := false
	for _, r := range regressions {
		if r.Feature == "user-auth" && r.FileType == "bdd" {
			found = true
			if r.Message == "" {
				t.Error("regression message should not be empty")
			}
		}
	}
	if !found {
		t.Error("expected bdd regression for user-auth")
	}

	state, _ := LoadState(dir)
	if state.Features["user-auth"].Stage != "bdd" {
		t.Errorf("stage should be downgraded to bdd, got %s", state.Features["user-auth"].Stage)
	}
}

func TestPRDChangeCreatesRedoTasks(t *testing.T) {
	dir := t.TempDir()
	setupFeatureFiles(t, dir, "user-auth", "seed", "bdd", "test")
	setState(t, dir, "user-auth", "bdd", nil, nil)

	// Change PRD section
	prdPath := filepath.Join(dir, ".ptsd", "docs", "PRD.md")
	appendFile(t, prdPath, "\n## user-auth\nNew requirement\n")

	regressions, err := CheckRegressions(dir)
	if err != nil {
		t.Fatalf("CheckRegressions failed: %v", err)
	}

	state, _ := LoadState(dir)
	if state.Features["user-auth"].Stage != "prd" {
		t.Errorf("stage should be downgraded to prd, got %s", state.Features["user-auth"].Stage)
	}

	found := false
	for _, r := range regressions {
		if r.Feature == "user-auth" && r.FileType == "prd" {
			found = true
		}
	}
	if !found {
		t.Error("expected prd regression for user-auth")
	}
}

func TestSeedChangeWarning(t *testing.T) {
	dir := t.TempDir()
	setupFeatureFiles(t, dir, "user-auth", "seed", "bdd", "test")
	setState(t, dir, "user-auth", "test", nil, nil)

	// Change seed file
	seedPath := filepath.Join(dir, ".ptsd", "seeds", "user-auth", "seed.yaml")
	appendFile(t, seedPath, "\nupdated: true\n")

	regressions, err := CheckRegressions(dir)
	if err != nil {
		t.Fatalf("CheckRegressions failed: %v", err)
	}

	found := false
	for _, r := range regressions {
		if r.Feature == "user-auth" && r.FileType == "seed" {
			found = true
		}
	}
	if !found {
		t.Error("expected seed regression warning for user-auth")
	}
}

func TestNoRegressionOnExpectedChange(t *testing.T) {
	dir := t.TempDir()
	setupFeatureFiles(t, dir, "user-auth", "seed", "bdd", "test")
	setState(t, dir, "user-auth", "bdd", nil, nil)

	// Change BDD file (expected at bdd stage)
	bddPath := filepath.Join(dir, ".ptsd", "bdd", "user-auth.feature")
	appendFile(t, bddPath, "\n# new scenario")

	regressions, err := CheckRegressions(dir)
	if err != nil {
		t.Fatalf("CheckRegressions failed: %v", err)
	}

	for _, r := range regressions {
		if r.Feature == "user-auth" && r.FileType == "bdd" {
			t.Error("unexpected regression for bdd change at bdd stage")
		}
	}
}

func TestStoreReviewScores(t *testing.T) {
	dir := t.TempDir()
	setupFeatureFiles(t, dir, "user-auth", "seed", "bdd", "test")

	before := time.Now()
	err := RecordReview(dir, "user-auth", "prd", 8)
	if err != nil {
		t.Fatalf("RecordReview failed: %v", err)
	}

	state, err := LoadState(dir)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	feature, ok := state.Features["user-auth"]
	if !ok {
		t.Fatal("expected user-auth feature")
	}

	score, ok := feature.Scores["prd"]
	if !ok {
		t.Fatal("expected prd score")
	}
	if score.Value != 8 {
		t.Errorf("expected score 8, got %d", score.Value)
	}
	if score.Timestamp.Before(before) {
		t.Error("timestamp should be after record call")
	}
}

// Helpers

func setupFeatureFiles(t *testing.T, dir, feature string, artifacts ...string) {
	t.Helper()

	ptsdDir := filepath.Join(dir, ".ptsd")
	os.MkdirAll(ptsdDir, 0755)

	// features.yaml
	featuresYAML := "features:\n  - id: " + feature + "\n    title: " + feature + "\n"
	os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresYAML), 0644)

	// state.yaml
	os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644)

	// PRD
	prdDir := filepath.Join(ptsdDir, "docs")
	os.MkdirAll(prdDir, 0755)
	os.WriteFile(filepath.Join(prdDir, "PRD.md"), []byte("# PRD\n## "+feature+"\nContent\n"), 0644)

	// BDD
	bddDir := filepath.Join(ptsdDir, "bdd")
	os.MkdirAll(bddDir, 0755)
	os.WriteFile(filepath.Join(bddDir, feature+".feature"), []byte("Feature: "+feature+"\n"), 0644)

	// Seeds
	seedDir := filepath.Join(ptsdDir, "seeds", feature)
	os.MkdirAll(seedDir, 0755)
	os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte("data: value\n"), 0644)

	// Test file
	testDir := filepath.Join(dir, "internal", "core")
	os.MkdirAll(testDir, 0755)
	os.WriteFile(filepath.Join(testDir, feature+"_test.go"), []byte("package core\n"), 0644)
}

func setState(t *testing.T, dir, feature, stage string, hashes, scores map[string]string) {
	t.Helper()

	ptsdDir := filepath.Join(dir, ".ptsd")
	statePath := filepath.Join(ptsdDir, "state.yaml")

	if hashes == nil {
		hashes = map[string]string{
			"seed": fileHash(t, filepath.Join(ptsdDir, "seeds", feature, "seed.yaml")),
			"bdd":  fileHash(t, filepath.Join(ptsdDir, "bdd", feature+".feature")),
			"test": fileHash(t, filepath.Join(dir, "internal", "core", feature+"_test.go")),
			"prd":  fileHash(t, filepath.Join(ptsdDir, "docs", "PRD.md")),
		}
	}

	content := "features:\n  " + feature + ":\n    stage: " + stage + "\n    hashes:\n"
	for k, v := range hashes {
		content += "      " + k + ": " + v + "\n"
	}
	content += "    scores: {}\n"

	os.WriteFile(statePath, []byte(content), 0644)
}

func fileHash(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func appendFile(t *testing.T, path, content string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("cannot open %s: %v", path, err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("cannot write %s: %v", path, err)
	}
}
