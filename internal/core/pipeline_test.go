package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanProjectPasses(t *testing.T) {
	dir := setupCleanProject(t)
	errors, err := Validate(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errors) > 0 {
		t.Errorf("expected no errors, got: %v", errors)
	}
}

func TestFeatureWithoutPRDAnchor(t *testing.T) {
	dir := setupProjectWithFeature(t, "user-auth", func(base string) {
		// Create feature without PRD anchor
		writeFeaturesYAML(t, base, `- id: user-auth
  title: "User Auth"
  status: active
`)
		// Create other required files but NO PRD anchor
		createBDD(t, base, "user-auth")
		createSeed(t, base, "user-auth")
	})

	errors, err := Validate(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertHasError(t, errors, "user-auth", "pipeline", "has no prd anchor")
}

func TestFeatureWithBDDButNoSeed(t *testing.T) {
	dir := setupProjectWithFeature(t, "user-auth", func(base string) {
		writeFeaturesYAML(t, base, `- id: user-auth
  title: "User Auth"
  status: active
`)
		createPRDAnchor(t, base, "user-auth")
		createBDD(t, base, "user-auth")
		// No seed created
	})

	errors, err := Validate(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertHasError(t, errors, "user-auth", "pipeline", "has bdd but no seed")
}

func TestFeatureWithBDDButNoTests(t *testing.T) {
	dir := setupProjectWithFeature(t, "user-auth", func(base string) {
		writeFeaturesYAML(t, base, `- id: user-auth
  title: "User Auth"
  status: active
`)
		createPRDAnchor(t, base, "user-auth")
		createBDD(t, base, "user-auth")
		createSeed(t, base, "user-auth")
		// No test files
	})

	errors, err := Validate(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertHasError(t, errors, "user-auth", "pipeline", "has bdd but no tests")
}

func TestPlannedAndDeferredFeaturesAreSkipped(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, ".ptsd")
	createDirs(t, base)

	writeFeaturesYAML(t, base, `- id: future
  title: "Future"
  status: planned
- id: old
  title: "Old"
  status: deferred
`)
	// Intentionally incomplete - no seeds, no BDD, no PRD anchors
	// But these should be skipped due to status

	errors, err := Validate(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, e := range errors {
		if e.Feature == "future" || e.Feature == "old" {
			t.Errorf("planned/deferred feature should be skipped: %v", e)
		}
	}
}

func TestMockPatternsDetectedInTestFiles(t *testing.T) {
	dir := setupProjectWithFeature(t, "user-auth", func(base string) {
		writeFeaturesYAML(t, base, `- id: user-auth
  title: "User Auth"
  status: active
`)
		createPRDAnchor(t, base, "user-auth")
		createBDD(t, base, "user-auth")
		createSeed(t, base, "user-auth")

		// Create test file with mock pattern
		testDir := filepath.Join(base, "..", "internal", "auth")
		os.MkdirAll(testDir, 0755)
		os.WriteFile(filepath.Join(testDir, "auth_test.go"), []byte(`package auth
import "testing"
func TestAuth(t *testing.T) {
	vi.mock("something") // mock pattern
}
`), 0644)
	})

	errors, err := Validate(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertHasError(t, errors, "", "pipeline", "mock detected")
}

func TestMultipleErrorsReported(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, ".ptsd")
	createDirs(t, base)

	writeFeaturesYAML(t, base, `- id: feature-a
  title: "Feature A"
  status: active
- id: feature-b
  title: "Feature B"
  status: active
- id: feature-c
  title: "Feature C"
  status: active
`)
	// All three missing PRD anchors - should get 3 errors

	errors, err := Validate(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errors) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(errors), errors)
	}
}

// Helpers

func setupCleanProject(t *testing.T) string {
	dir := t.TempDir()
	base := filepath.Join(dir, ".ptsd")
	createDirs(t, base)

	writeFeaturesYAML(t, base, `- id: user-auth
  title: "User Auth"
  status: active
`)
	createPRDAnchor(t, base, "user-auth")
	createBDD(t, base, "user-auth")
	createSeed(t, base, "user-auth")

	// Create test file (no mocks)
	testDir := filepath.Join(dir, "internal", "auth")
	os.MkdirAll(testDir, 0755)
	os.WriteFile(filepath.Join(testDir, "auth_test.go"), []byte(`package auth
import "testing"
func TestAuth(t *testing.T) {}
`), 0644)

	return dir
}

func setupProjectWithFeature(t *testing.T, featureID string, configure func(base string)) string {
	dir := t.TempDir()
	base := filepath.Join(dir, ".ptsd")
	createDirs(t, base)
	configure(base)
	return dir
}

func createDirs(t *testing.T, base string) {
	dirs := []string{
		base,
		filepath.Join(base, "bdd"),
		filepath.Join(base, "seeds"),
		filepath.Join(base, "docs"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	// Create minimal config
	os.WriteFile(filepath.Join(base, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644)
	// Create minimal PRD
	os.WriteFile(filepath.Join(base, "docs", "PRD.md"), []byte("# PRD\n"), 0644)
}

func writeFeaturesYAML(t *testing.T, base, content string) {
	if err := os.WriteFile(filepath.Join(base, "features.yaml"), []byte("features:\n"+content), 0644); err != nil {
		t.Fatalf("write features.yaml: %v", err)
	}
}

func createPRDAnchor(t *testing.T, base, featureID string) {
	prdPath := filepath.Join(base, "docs", "PRD.md")
	content := "# PRD\n\n<!-- feature:" + featureID + " -->\n\n### " + featureID + "\n\nDescription here.\n"
	if err := os.WriteFile(prdPath, []byte(content), 0644); err != nil {
		t.Fatalf("write PRD: %v", err)
	}
}

func createBDD(t *testing.T, base, featureID string) {
	bddPath := filepath.Join(base, "bdd", featureID+".feature")
	content := "@feature:" + featureID + "\nFeature: " + featureID + "\n"
	if err := os.WriteFile(bddPath, []byte(content), 0644); err != nil {
		t.Fatalf("write BDD: %v", err)
	}
}

func createSeed(t *testing.T, base, featureID string) {
	seedDir := filepath.Join(base, "seeds", featureID)
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		t.Fatalf("mkdir seed: %v", err)
	}
	os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte("id: "+featureID+"\n"), 0644)
}

func assertHasError(t *testing.T, errors []ValidationError, feature, category, contains string) {
	for _, e := range errors {
		if (feature == "" || e.Feature == feature) && e.Category == category {
			return
		}
	}
	t.Errorf("expected error with feature=%q category=%q, got: %v", feature, category, errors)
}
