package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAdoptCreatesStructure verifies adopt creates the .ptsd/ directory structure.
func TestAdoptCreatesStructure(t *testing.T) {
	dir := t.TempDir()

	err := AdoptProject(dir)
	if err != nil {
		t.Fatalf("AdoptProject failed: %v", err)
	}

	for _, sub := range []string{"", "seeds", "bdd", "docs", "skills"} {
		path := filepath.Join(dir, ".ptsd", sub)
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Errorf("expected directory %s to exist", path)
		}
	}

	// ptsd.yaml must exist
	ptsdYAML := filepath.Join(dir, ".ptsd", "ptsd.yaml")
	if _, err := os.Stat(ptsdYAML); err != nil {
		t.Error("ptsd.yaml not created")
	}

	// features.yaml must exist
	featYAML := filepath.Join(dir, ".ptsd", "features.yaml")
	if _, err := os.Stat(featYAML); err != nil {
		t.Error("features.yaml not created")
	}
}

// TestAdoptRefusesIfAlreadyInitialized verifies adopt fails when .ptsd/ exists.
func TestAdoptRefusesIfAlreadyInitialized(t *testing.T) {
	dir := setupProjectWithFeatures(t, "existing:planned")

	err := AdoptProject(dir)
	if err == nil {
		t.Fatal("expected error for already-initialized project")
	}
	if !strings.Contains(err.Error(), "err:validation") {
		t.Errorf("expected err:validation, got: %v", err)
	}
	if !strings.Contains(err.Error(), "already initialized") {
		t.Errorf("expected 'already initialized' in error, got: %v", err)
	}
}

// TestAdoptDiscoversBDDFiles verifies adopt finds .feature files and extracts feature IDs.
func TestAdoptDiscoversBDDFiles(t *testing.T) {
	dir := t.TempDir()

	// Create .feature files with @feature: tags
	bddDir := filepath.Join(dir, "bdd")
	if err := os.MkdirAll(bddDir, 0755); err != nil {
		t.Fatal(err)
	}

	features := []struct {
		id      string
		content string
	}{
		{"auth", "@feature:auth\nFeature: Authentication\n  Scenario: Login\n"},
		{"billing", "@feature:billing\nFeature: Billing\n  Scenario: Pay\n"},
	}

	for _, f := range features {
		path := filepath.Join(bddDir, f.id+".feature")
		if err := os.WriteFile(path, []byte(f.content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	err := AdoptProject(dir)
	if err != nil {
		t.Fatalf("AdoptProject failed: %v", err)
	}

	// features.yaml must contain discovered IDs
	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "features.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "auth") {
		t.Error("features.yaml missing 'auth'")
	}
	if !strings.Contains(content, "billing") {
		t.Error("features.yaml missing 'billing'")
	}

	// .feature files must be moved to .ptsd/bdd/
	for _, f := range features {
		dst := filepath.Join(dir, ".ptsd", "bdd", f.id+".feature")
		if _, err := os.Stat(dst); err != nil {
			t.Errorf("expected .feature file at %s", dst)
		}
		// original must be gone
		orig := filepath.Join(bddDir, f.id+".feature")
		if _, err := os.Stat(orig); err == nil {
			t.Errorf("original .feature file not removed: %s", orig)
		}
	}
}

// TestAdoptScansTestFiles verifies adopt discovers existing test files.
func TestAdoptScansTestFiles(t *testing.T) {
	dir := t.TempDir()

	// Create some test files
	srcDir := filepath.Join(dir, "internal", "core")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(srcDir, "auth_test.go")
	if err := os.WriteFile(testFile, []byte("package core\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := AdoptDryRun(dir)
	if err != nil {
		t.Fatalf("AdoptDryRun failed: %v", err)
	}

	found := false
	for _, f := range result.TestFiles {
		if strings.Contains(f, "auth_test.go") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected auth_test.go in TestFiles, got: %v", result.TestFiles)
	}
}

// TestAdoptDryRunMakesNoChanges verifies dry-run does not create any files.
func TestAdoptDryRunMakesNoChanges(t *testing.T) {
	dir := t.TempDir()

	// Create a .feature file
	featureFile := filepath.Join(dir, "login.feature")
	if err := os.WriteFile(featureFile, []byte("@feature:login\nFeature: Login\n"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := AdoptDryRun(dir)
	if err != nil {
		t.Fatalf("AdoptDryRun failed: %v", err)
	}

	// .ptsd/ must NOT exist
	ptsdDir := filepath.Join(dir, ".ptsd")
	if _, err := os.Stat(ptsdDir); err == nil {
		t.Error(".ptsd/ must not be created during dry run")
	}

	// Original .feature file must still be there
	if _, err := os.Stat(featureFile); err != nil {
		t.Error("original .feature file must not be moved during dry run")
	}

	// Result must list discovered features
	if len(result.BDDFiles) == 0 {
		t.Error("expected BDDFiles to contain 'login'")
	}
	if result.BDDFiles[0] != "login" {
		t.Errorf("expected 'login', got %q", result.BDDFiles[0])
	}
}

// TestAdoptDryRunRefusesIfAlreadyInitialized verifies dry-run also checks for existing .ptsd/.
func TestAdoptDryRunRefusesIfAlreadyInitialized(t *testing.T) {
	dir := setupProjectWithFeatures(t, "existing:planned")

	_, err := AdoptDryRun(dir)
	if err == nil {
		t.Fatal("expected error for already-initialized project")
	}
	if !strings.Contains(err.Error(), "err:validation") {
		t.Errorf("expected err:validation, got: %v", err)
	}
}

// TestAdoptEmptyProject verifies adopt works on a project with no artifacts.
func TestAdoptEmptyProject(t *testing.T) {
	dir := t.TempDir()

	err := AdoptProject(dir)
	if err != nil {
		t.Fatalf("AdoptProject on empty project failed: %v", err)
	}

	// features.yaml must be created with empty features
	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "features.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "features:") {
		t.Error("features.yaml missing 'features:' header")
	}
}
