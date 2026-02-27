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

func TestGateCheck_ReviewStatusBlocked(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")

	result := GateCheck(dir, ".ptsd/review-status.yaml")
	if result.Allowed {
		t.Error("expected .ptsd/review-status.yaml to be blocked (BYPASS-2 fix)")
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

func TestGateCheck_ClaudeHooksAlwaysAllowed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")

	paths := []string{
		".claude/hooks/ptsd-context.sh",
		".claude/hooks/ptsd-gate.sh",
		".claude/hooks/custom.sh",
	}
	for _, p := range paths {
		result := GateCheck(dir, p)
		if !result.Allowed {
			t.Errorf("expected .claude/hooks/ path %q to be allowed, got blocked: %s", p, result.Reason)
		}
	}
}

func TestGateCheck_SkillsAlwaysAllowed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")

	result := GateCheck(dir, ".ptsd/skills/custom-skill.md")
	if !result.Allowed {
		t.Errorf("expected .ptsd/skills/ path to be allowed, got blocked: %s", result.Reason)
	}
}

func TestGateCheck_UnknownNonCodeFileAllowed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")

	paths := []string{"README.md", "docs/guide.txt", ".gitignore", "Makefile"}
	for _, p := range paths {
		result := GateCheck(dir, p)
		if !result.Allowed {
			t.Errorf("expected non-code file %q to be allowed, got blocked: %s", p, result.Reason)
		}
	}
}

func TestGateCheck_FeatureIDSubstringCollision(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress", "authorization:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	// BDD exists for "auth" but NOT for "authorization"
	os.WriteFile(filepath.Join(ptsd, "bdd", "auth.feature"), []byte("@feature:auth\nFeature: Auth\n"), 0644)

	// File "authorization_test.go" should match feature "authorization", not "auth"
	result := GateCheck(dir, "internal/core/authorization_test.go")
	// If it matched "auth" (which has BDD), it would be allowed.
	// It should match "authorization" (which has no BDD) and be blocked.
	if result.Feature == "auth" {
		t.Errorf("expected feature=authorization, got feature=auth (substring collision)")
	}
}

func TestGateCheck_DeepNestedSeedFile(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	// Create PRD anchor for auth
	os.MkdirAll(filepath.Join(ptsd, "docs"), 0755)
	os.WriteFile(filepath.Join(ptsd, "docs", "PRD.md"), []byte("<!-- feature:auth -->\n## Auth\n"), 0644)

	result := GateCheck(dir, ".ptsd/seeds/auth/data/nested/fixture.json")
	if result.Feature != "auth" {
		t.Errorf("expected feature=auth for deeply nested seed, got %q", result.Feature)
	}
	if !result.Allowed {
		t.Errorf("expected deeply nested seed file to be allowed with PRD anchor, got blocked: %s", result.Reason)
	}
}

func TestGateCheck_ImplBlockedWithoutTests(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	// Create empty state
	os.WriteFile(filepath.Join(ptsd, "state.yaml"), []byte("features: {}\n"), 0644)

	result := GateCheck(dir, "internal/core/auth.go")
	if result.Allowed {
		t.Error("expected impl write to be blocked when no tests exist for feature")
	}
	if result.Feature != "auth" {
		t.Errorf("expected feature=auth, got %q", result.Feature)
	}
}

func TestGateCheck_EmptyFeaturesAllowsAll(t *testing.T) {
	dir := setupProjectWithFeatures(t)

	// Test file that can't be mapped to any feature
	result := GateCheck(dir, "internal/core/unknown_test.go")
	if !result.Allowed {
		t.Errorf("expected test file with no feature match to be allowed, got blocked: %s", result.Reason)
	}
}

func TestGateCheck_IssuesYAMLAlwaysAllowed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")

	result := GateCheck(dir, ".ptsd/issues.yaml")
	if !result.Allowed {
		t.Errorf("expected .ptsd/issues.yaml to be always allowed, got blocked: %s", result.Reason)
	}
}

func TestGateCheck_SettingsJSONAlwaysAllowed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")

	result := GateCheck(dir, ".claude/settings.json")
	if !result.Allowed {
		t.Errorf("expected .claude/settings.json to be always allowed, got blocked: %s", result.Reason)
	}
}
