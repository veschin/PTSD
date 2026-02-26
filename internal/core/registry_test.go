package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddFeature(t *testing.T) {
	dir := t.TempDir()
	setupFeaturesYAML(t, dir)

	err := AddFeature(dir, "user-auth", "User Authentication")
	if err != nil {
		t.Fatalf("AddFeature failed: %v", err)
	}

	features, err := ListFeatures(dir, "")
	if err != nil {
		t.Fatalf("ListFeatures failed: %v", err)
	}
	if len(features) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(features))
	}
	if features[0].ID != "user-auth" {
		t.Errorf("expected ID 'user-auth', got %q", features[0].ID)
	}
	if features[0].Title != "User Authentication" {
		t.Errorf("expected Title 'User Authentication', got %q", features[0].Title)
	}
	if features[0].Status != "planned" {
		t.Errorf("expected status 'planned', got %q", features[0].Status)
	}
}

func TestAddDuplicateFeature(t *testing.T) {
	dir := t.TempDir()
	setupFeaturesYAML(t, dir)

	err := AddFeature(dir, "user-auth", "User Authentication")
	if err != nil {
		t.Fatalf("first AddFeature failed: %v", err)
	}

	err = AddFeature(dir, "user-auth", "Another Title")
	if err == nil {
		t.Fatal("expected error for duplicate feature")
	}
	if !strings.HasPrefix(err.Error(), "err:validation") {
		t.Errorf("expected error starting with 'err:validation', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "user-auth") {
		t.Errorf("expected error to contain 'user-auth', got %q", err.Error())
	}
}

func TestListFeatures(t *testing.T) {
	dir := t.TempDir()
	setupFeaturesYAML(t, dir)

	addFeatures(t, dir, "user-auth", "catalog", "payments")

	features, err := ListFeatures(dir, "")
	if err != nil {
		t.Fatalf("ListFeatures failed: %v", err)
	}
	if len(features) != 3 {
		t.Fatalf("expected 3 features, got %d", len(features))
	}

	ids := make(map[string]bool)
	for _, f := range features {
		ids[f.ID] = true
	}
	for _, id := range []string{"user-auth", "catalog", "payments"} {
		if !ids[id] {
			t.Errorf("missing feature %q", id)
		}
	}
}

func TestListFeaturesFilteredByStatus(t *testing.T) {
	dir := t.TempDir()
	setupFeaturesYAML(t, dir)

	addFeatures(t, dir, "user-auth", "catalog")

	if err := UpdateFeatureStatus(dir, "user-auth", "in-progress"); err != nil {
		t.Fatalf("UpdateFeatureStatus failed: %v", err)
	}

	features, err := ListFeatures(dir, "planned")
	if err != nil {
		t.Fatalf("ListFeatures failed: %v", err)
	}
	if len(features) != 1 {
		t.Fatalf("expected 1 planned feature, got %d", len(features))
	}
	if features[0].ID != "catalog" {
		t.Errorf("expected 'catalog', got %q", features[0].ID)
	}
}

func TestShowFeatureDetails(t *testing.T) {
	dir := t.TempDir()
	setupFeaturesYAML(t, dir)
	setupFeatureArtifacts(t, dir, "user-auth")

	detail, err := ShowFeature(dir, "user-auth")
	if err != nil {
		t.Fatalf("ShowFeature failed: %v", err)
	}
	if detail.ID != "user-auth" {
		t.Errorf("expected ID 'user-auth', got %q", detail.ID)
	}
	if detail.Status == "" {
		t.Error("expected non-empty status")
	}
	if detail.PRDAnchor == "" {
		t.Error("expected non-empty PRD anchor")
	}
	if detail.SeedStatus == "" {
		t.Error("expected non-empty seed status")
	}
	if detail.ScenarioCount != 3 {
		t.Errorf("expected 3 scenarios, got %d", detail.ScenarioCount)
	}
	if detail.TestCount != 2 {
		t.Errorf("expected 2 tests, got %d", detail.TestCount)
	}
}

func TestUpdateFeatureStatus(t *testing.T) {
	dir := t.TempDir()
	setupFeaturesYAML(t, dir)
	addFeatures(t, dir, "user-auth")

	err := UpdateFeatureStatus(dir, "user-auth", "in-progress")
	if err != nil {
		t.Fatalf("UpdateFeatureStatus failed: %v", err)
	}

	features, _ := ListFeatures(dir, "")
	if len(features) != 1 || features[0].Status != "in-progress" {
		t.Errorf("expected status 'in-progress', got %q", features[0].Status)
	}
}

func TestUpdateFeatureStatusImplementedRequiresPassingTests(t *testing.T) {
	dir := t.TempDir()
	setupFeaturesYAML(t, dir)
	addFeatures(t, dir, "user-auth")
	setTestsFailing(t, dir, "user-auth")

	err := UpdateFeatureStatus(dir, "user-auth", "implemented")
	if err == nil {
		t.Fatal("expected error when tests are failing")
	}
	if !strings.HasPrefix(err.Error(), "err:pipeline") {
		t.Errorf("expected error starting with 'err:pipeline', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "tests not passing") {
		t.Errorf("expected error to contain 'tests not passing', got %q", err.Error())
	}
}

func TestRemoveFeature(t *testing.T) {
	dir := t.TempDir()
	setupFeaturesYAML(t, dir)
	addFeatures(t, dir, "user-auth")

	err := RemoveFeature(dir, "user-auth")
	if err != nil {
		t.Fatalf("RemoveFeature failed: %v", err)
	}

	features, _ := ListFeatures(dir, "")
	if len(features) != 0 {
		t.Errorf("expected 0 features after removal, got %d", len(features))
	}
}

// Helpers

func setupFeaturesYAML(t *testing.T, dir string) {
	t.Helper()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	yamlPath := filepath.Join(ptsdDir, "features.yaml")
	content := "features: []\n"
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func addFeatures(t *testing.T, dir string, ids ...string) {
	t.Helper()
	for _, id := range ids {
		if err := AddFeature(dir, id, "Title for "+id); err != nil {
			t.Fatalf("AddFeature(%q) failed: %v", id, err)
		}
	}
}

func setupFeatureArtifacts(t *testing.T, dir string, id string) {
	t.Helper()
	if err := AddFeature(dir, id, "User Auth"); err != nil {
		t.Fatal(err)
	}
	// Create BDD scenarios
	bddDir := filepath.Join(dir, ".ptsd", "bdd")
	if err := os.MkdirAll(bddDir, 0755); err != nil {
		t.Fatal(err)
	}
	scenarios := "Feature: " + id + "\nScenario: one\nScenario: two\nScenario: three\n"
	if err := os.WriteFile(filepath.Join(bddDir, id+".feature"), []byte(scenarios), 0644); err != nil {
		t.Fatal(err)
	}
	// Create PRD with anchor
	docsDir := filepath.Join(dir, ".ptsd", "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}
	prd := "# PRD\n<!-- feature:" + id + " -->\nSection for " + id + "\n"
	if err := os.WriteFile(filepath.Join(docsDir, "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}
	// Create seed marker
	seedDir := filepath.Join(dir, ".ptsd", "seeds", id)
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Update state with test count and failing status
	statePath := filepath.Join(dir, ".ptsd", "state.yaml")
	stateContent := "features:\n  " + id + ":\n    tests: 2\n    test_status: passing\n"
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}
}

func setTestsFailing(t *testing.T, dir string, id string) {
	t.Helper()
	statePath := filepath.Join(dir, ".ptsd", "state.yaml")
	stateContent := "features:\n  " + id + ":\n    tests: 2\n    test_status: failing\n"
	if err := os.WriteFile(statePath, []byte(stateContent), 0644); err != nil {
		t.Fatal(err)
	}
}
