package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGateCheck_AlwaysAllowed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")

	allowed := []string{
		".ptsd/docs/PRD.md",
		".ptsd/tasks.yaml",
		".ptsd/review-status.yaml",
		".ptsd/state.yaml",
		"CLAUDE.md",
	}

	for _, f := range allowed {
		result := GateCheck(dir, f)
		if !result.Allowed {
			t.Errorf("expected %s to be always allowed, got blocked: %s", f, result.Reason)
		}
	}
}

func TestGateCheck_BDDRequiresSeed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")

	// No seed exists
	result := GateCheck(dir, ".ptsd/bdd/auth.feature")
	if result.Allowed {
		t.Error("expected BDD write to be blocked when no seed exists")
	}
	if result.Feature != "auth" {
		t.Errorf("expected feature=auth, got %s", result.Feature)
	}
}

func TestGateCheck_BDDAllowedWithSeed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	// Create seed
	seedDir := filepath.Join(ptsd, "seeds", "auth")
	os.MkdirAll(seedDir, 0755)
	os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte("feature: auth\n"), 0644)

	result := GateCheck(dir, ".ptsd/bdd/auth.feature")
	if !result.Allowed {
		t.Errorf("expected BDD write to be allowed with seed, got blocked: %s", result.Reason)
	}
}

func TestGateCheck_SeedRequiresPRDAnchor(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	// PRD with no anchor for auth
	os.MkdirAll(filepath.Join(ptsd, "docs"), 0755)
	os.WriteFile(filepath.Join(ptsd, "docs", "PRD.md"), []byte("# PRD\n"), 0644)

	result := GateCheck(dir, ".ptsd/seeds/auth/seed.yaml")
	if result.Allowed {
		t.Error("expected seed write to be blocked when no PRD anchor exists")
	}
}

func TestGateCheck_SeedAllowedWithPRDAnchor(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.MkdirAll(filepath.Join(ptsd, "docs"), 0755)
	os.WriteFile(filepath.Join(ptsd, "docs", "PRD.md"), []byte("<!-- feature:auth -->\n## Auth\n"), 0644)

	result := GateCheck(dir, ".ptsd/seeds/auth/seed.yaml")
	if !result.Allowed {
		t.Errorf("expected seed write to be allowed with PRD anchor, got blocked: %s", result.Reason)
	}
}

func TestGateCheck_TestRequiresBDD(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")

	// No BDD file exists
	result := GateCheck(dir, "internal/core/auth_test.go")
	if result.Allowed {
		t.Error("expected test write to be blocked when no BDD exists")
	}
}

func TestGateCheck_TestAllowedWithBDD(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	// Create BDD
	os.WriteFile(filepath.Join(ptsd, "bdd", "auth.feature"), []byte("@feature:auth\nFeature: Auth\n"), 0644)

	result := GateCheck(dir, "internal/core/auth_test.go")
	if !result.Allowed {
		t.Errorf("expected test write to be allowed with BDD, got blocked: %s", result.Reason)
	}
}

func TestGateCheck_AbsolutePath(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")

	absPath := filepath.Join(dir, ".ptsd", "docs", "PRD.md")
	result := GateCheck(dir, absPath)
	if !result.Allowed {
		t.Errorf("expected absolute path to PRD.md to be allowed, got blocked: %s", result.Reason)
	}
}
