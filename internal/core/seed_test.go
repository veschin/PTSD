package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitSeed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")

	err := InitSeed(dir, "user-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seedPath := filepath.Join(dir, ".ptsd", "seeds", "user-auth", "seed.yaml")
	if _, err := os.Stat(seedPath); os.IsNotExist(err) {
		t.Fatal("seed.yaml not created")
	}

	data, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "feature") {
		t.Error("seed.yaml missing feature field")
	}
}

func TestInitSeedNonexistentFeature(t *testing.T) {
	dir := setupProjectWithFeatures(t)

	err := InitSeed(dir, "ghost")
	if err == nil {
		t.Fatal("expected error for nonexistent feature")
	}
	if !strings.HasPrefix(err.Error(), "err:validation") {
		t.Errorf("expected err:validation, got: %v", err)
	}
}

func TestAddSeedFile(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")
	if err := InitSeed(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}

	srcFile := filepath.Join(dir, "user.json")
	if err := os.WriteFile(srcFile, []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	err := AddSeedFile(dir, "user-auth", srcFile, "data", "test user data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dstPath := filepath.Join(dir, ".ptsd", "seeds", "user-auth", "user.json")
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Fatal("user.json not copied to seed dir")
	}

	seedData, err := os.ReadFile(filepath.Join(dir, ".ptsd", "seeds", "user-auth", "seed.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	seedStr := string(seedData)
	if seedStr == "" || !strings.Contains(seedStr, "user.json") || !strings.Contains(seedStr, "data") {
		t.Error("seed.yaml files list missing user.json with type data")
	}
}

func TestCheckSeeds(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress", "catalog:in-progress")
	if err := InitSeed(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}

	missing, err := CheckSeeds(dir)
	if err == nil {
		t.Fatal("expected error for missing seeds")
	}
	if len(missing) != 1 || missing[0] != "catalog" {
		t.Errorf("expected [catalog] missing, got: %v", missing)
	}
	if !strings.HasPrefix(err.Error(), "err:pipeline") {
		t.Errorf("expected err:pipeline, got: %v", err)
	}
}

func TestCheckSeedsPlannedSkipped(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress", "future:planned")
	if err := InitSeed(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}

	missing, _ := CheckSeeds(dir)
	for _, f := range missing {
		if f == "future" {
			t.Error("planned feature should be skipped in seed check")
		}
	}
}

func TestCheckSeedsAllPresent(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")
	if err := InitSeed(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}

	missing, err := CheckSeeds(dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing seeds, got: %v", missing)
	}
}

func TestCheckSeedsManifestFilesMissing(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")
	if err := InitSeed(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}

	// Write manifest referencing files that do not exist on disk
	seedPath := filepath.Join(dir, ".ptsd", "seeds", "user-auth", "seed.yaml")
	manifest := "feature: user-auth\nfiles:\n  - path: missing.json\n    type: data\n  - path: gone.csv\n    type: fixture\n"
	if err := os.WriteFile(seedPath, []byte(manifest), 0644); err != nil {
		t.Fatal(err)
	}

	missing, err := CheckSeeds(dir)
	if err == nil {
		t.Fatal("expected error for manifest with missing files")
	}
	if !strings.Contains(err.Error(), "missing file") {
		t.Errorf("expected 'missing file' in error, got: %v", err)
	}
	if len(missing) != 1 || missing[0] != "user-auth" {
		t.Errorf("expected [user-auth] in missing, got: %v", missing)
	}
}

func TestCheckSeedsManifestFilesAllPresent(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")
	if err := InitSeed(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}

	// Add a real file via AddSeedFile so both manifest and file exist
	srcFile := filepath.Join(dir, "data.json")
	if err := os.WriteFile(srcFile, []byte(`{"ok":true}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := AddSeedFile(dir, "user-auth", srcFile, "data", "test"); err != nil {
		t.Fatal(err)
	}

	missing, err := CheckSeeds(dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(missing) != 0 {
		t.Errorf("expected no missing seeds, got: %v", missing)
	}
}

