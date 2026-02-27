package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupValidateCleanProject creates a complete, valid .ptsd project with:
// - ptsd.yaml
// - features.yaml with one "planned" feature (planned is skipped by validate)
// - state.yaml (empty)
// - docs/PRD.md with feature anchor
// Returns the project root directory.
func setupValidateCleanProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")

	for _, d := range []string{
		ptsdDir,
		filepath.Join(ptsdDir, "bdd"),
		filepath.Join(ptsdDir, "seeds"),
		filepath.Join(ptsdDir, "docs"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644); err != nil {
		t.Fatalf("write ptsd.yaml: %v", err)
	}

	// Use "planned" status — Validate skips planned/deferred features,
	// so this yields a clean project with exit 0.
	featuresContent := "features:\n  - id: beta\n    title: Beta Feature\n    status: planned\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresContent), 0644); err != nil {
		t.Fatalf("write features.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	prdContent := "# PRD\n\n<!-- feature:beta -->\n\n### Beta\n\nDescription.\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "docs", "PRD.md"), []byte(prdContent), 0644); err != nil {
		t.Fatalf("write PRD.md: %v", err)
	}

	return dir
}

// setupValidateViolationProject creates a project with a pipeline violation:
// an active feature with BDD but no seed (triggers "has bdd but no seed" error).
func setupValidateViolationProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")

	for _, d := range []string{
		ptsdDir,
		filepath.Join(ptsdDir, "bdd"),
		filepath.Join(ptsdDir, "seeds"),
		filepath.Join(ptsdDir, "docs"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644); err != nil {
		t.Fatalf("write ptsd.yaml: %v", err)
	}

	// Active feature (not planned/deferred) — pipeline checks apply
	featuresContent := "features:\n  - id: gamma\n    title: Gamma Feature\n    status: in-progress\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresContent), 0644); err != nil {
		t.Fatalf("write features.yaml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	// PRD anchor present
	prdContent := "# PRD\n\n<!-- feature:gamma -->\n\n### Gamma\n\nDescription.\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "docs", "PRD.md"), []byte(prdContent), 0644); err != nil {
		t.Fatalf("write PRD.md: %v", err)
	}

	// BDD file exists — but NO seed directory → "has bdd but no seed" violation
	bddContent := "@feature:gamma\nFeature: Gamma Feature\n  Scenario: something\n    Given a condition\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "bdd", "gamma.feature"), []byte(bddContent), 0644); err != nil {
		t.Fatalf("write bdd: %v", err)
	}

	return dir
}

func TestRunValidate_CleanProject_Exit0(t *testing.T) {
	dir := setupValidateCleanProject(t)
	chdirTo(t, dir)

	code := RunValidate([]string{}, true)
	if code != 0 {
		t.Errorf("expected exit 0 for clean project, got %d", code)
	}
}

func TestRunValidate_CleanProject_HumanMode(t *testing.T) {
	dir := setupValidateCleanProject(t)
	chdirTo(t, dir)

	code := RunValidate([]string{}, false)
	if code != 0 {
		t.Errorf("expected exit 0 for clean project in human mode, got %d", code)
	}
}

func TestRunValidate_Violations_Exit1(t *testing.T) {
	dir := setupValidateViolationProject(t)
	chdirTo(t, dir)

	code := RunValidate([]string{}, true)
	if code != 1 {
		t.Errorf("expected exit 1 for project with violations, got %d", code)
	}
}

func TestRunValidate_Violations_ReportsErrors(t *testing.T) {
	dir := setupValidateViolationProject(t)
	chdirTo(t, dir)

	// Capture stdout to inspect error output
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	code := RunValidate([]string{}, true)

	w.Close()
	os.Stdout = origStdout

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 1 {
		t.Errorf("expected exit 1, got %d", code)
	}

	// Agent mode prefixes errors with "err:<category>"
	if len(output) == 0 {
		t.Error("expected non-empty error output")
	}
}

func TestRunValidate_NoPRDAnchor_Exit1(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	for _, d := range []string{
		ptsdDir,
		filepath.Join(ptsdDir, "bdd"),
		filepath.Join(ptsdDir, "seeds"),
		filepath.Join(ptsdDir, "docs"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Active feature but PRD has NO anchor for it
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(
		"features:\n  - id: delta\n    title: Delta\n    status: in-progress\n",
	), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// PRD with NO feature anchor
	if err := os.WriteFile(filepath.Join(ptsdDir, "docs", "PRD.md"), []byte("# PRD\n\nNo anchors here.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	chdirTo(t, dir)

	code := RunValidate([]string{}, true)
	if code != 1 {
		t.Errorf("expected exit 1 when PRD anchor is missing, got %d", code)
	}
}

// setupValidateBDDNoTestsProject creates a project with an active feature that
// has a BDD file and seed but NO test mapping in state.yaml.
// This should trigger the "has bdd but no tests" pipeline violation.
func setupValidateBDDNoTestsProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")

	for _, d := range []string{
		ptsdDir,
		filepath.Join(ptsdDir, "bdd"),
		filepath.Join(ptsdDir, "seeds", "epsilon"),
		filepath.Join(ptsdDir, "docs"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644); err != nil {
		t.Fatalf("write ptsd.yaml: %v", err)
	}

	featuresContent := "features:\n  - id: epsilon\n    title: Epsilon Feature\n    status: in-progress\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresContent), 0644); err != nil {
		t.Fatalf("write features.yaml: %v", err)
	}

	// state.yaml has no test mapping for epsilon
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	prdContent := "# PRD\n\n<!-- feature:epsilon -->\n\n### Epsilon\n\nDescription.\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "docs", "PRD.md"), []byte(prdContent), 0644); err != nil {
		t.Fatalf("write PRD.md: %v", err)
	}

	// BDD file exists
	bddContent := "@feature:epsilon\nFeature: Epsilon Feature\n  Scenario: basic\n    Given a condition\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "bdd", "epsilon.feature"), []byte(bddContent), 0644); err != nil {
		t.Fatalf("write bdd: %v", err)
	}

	// seed.yaml exists — so "has bdd but no seed" is NOT triggered, only "has bdd but no tests"
	if err := os.WriteFile(filepath.Join(ptsdDir, "seeds", "epsilon", "seed.yaml"), []byte("data: example\n"), 0644); err != nil {
		t.Fatalf("write seed.yaml: %v", err)
	}

	return dir
}

// TestRunValidate_BDDWithNoTests_Exit1 covers BDD scenario:
// "Feature with BDD but no tests" → exit 1 with err:pipeline message.
func TestRunValidate_BDDWithNoTests_Exit1(t *testing.T) {
	dir := setupValidateBDDNoTestsProject(t)
	chdirTo(t, dir)

	code := RunValidate([]string{}, true)
	if code != 1 {
		t.Errorf("expected exit 1 when feature has BDD but no tests, got %d", code)
	}
}

// TestRunValidate_BDDWithNoTests_OutputContainsError verifies the exact error
// message text emitted for the "has bdd but no tests" violation.
func TestRunValidate_BDDWithNoTests_OutputContainsError(t *testing.T) {
	dir := setupValidateBDDNoTestsProject(t)
	chdirTo(t, dir)

	output := captureStdout(t, func() {
		RunValidate([]string{}, true)
	})

	if !strings.Contains(output, "err:pipeline") {
		t.Errorf("expected err:pipeline prefix in output, got: %q", output)
	}
	if !strings.Contains(output, "has bdd but no tests") {
		t.Errorf("expected 'has bdd but no tests' in output, got: %q", output)
	}
}

// TestRunValidate_MultipleErrors_AllReported covers BDD scenario:
// "Multiple errors reported" — 3 active features with distinct pipeline violations,
// all 3 errors must appear in the output.
func TestRunValidate_MultipleErrors_AllReported(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")

	for _, d := range []string{
		ptsdDir,
		filepath.Join(ptsdDir, "bdd"),
		filepath.Join(ptsdDir, "seeds"),
		filepath.Join(ptsdDir, "docs"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Three active features. Each will produce a different pipeline violation:
	// f1 — no PRD anchor
	// f2 — BDD with no seed
	// f3 — BDD with no tests (seed exists)
	featuresContent := "features:\n" +
		"  - id: f1\n    title: F1\n    status: in-progress\n" +
		"  - id: f2\n    title: F2\n    status: in-progress\n" +
		"  - id: f3\n    title: F3\n    status: in-progress\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// PRD anchor only for f2 and f3 (f1 is missing → "has no prd anchor")
	prdContent := "# PRD\n\n<!-- feature:f2 -->\n\n### F2\n\n<!-- feature:f3 -->\n\n### F3\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "docs", "PRD.md"), []byte(prdContent), 0644); err != nil {
		t.Fatal(err)
	}

	// f2: BDD exists, seed does NOT → "has bdd but no seed"
	bddF2 := "@feature:f2\nFeature: F2\n  Scenario: x\n    Given y\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "bdd", "f2.feature"), []byte(bddF2), 0644); err != nil {
		t.Fatal(err)
	}

	// f3: BDD exists, seed exists, no test mapping → "has bdd but no tests"
	if err := os.MkdirAll(filepath.Join(ptsdDir, "seeds", "f3"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "seeds", "f3", "seed.yaml"), []byte("data: x\n"), 0644); err != nil {
		t.Fatal(err)
	}
	bddF3 := "@feature:f3\nFeature: F3\n  Scenario: x\n    Given y\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "bdd", "f3.feature"), []byte(bddF3), 0644); err != nil {
		t.Fatal(err)
	}

	chdirTo(t, dir)

	output := captureStdout(t, func() {
		code := RunValidate([]string{}, true)
		if code != 1 {
			t.Errorf("expected exit 1 with multiple violations, got %d", code)
		}
	})

	// All three errors must be reported.
	for _, want := range []string{
		"f1", // no prd anchor
		"f2", // has bdd but no seed
		"f3", // has bdd but no tests
	} {
		if !strings.Contains(output, want) {
			t.Errorf("expected feature %q to appear in error output, got: %q", want, output)
		}
	}
	if !strings.Contains(output, "err:pipeline") {
		t.Errorf("expected err:pipeline prefix in output, got: %q", output)
	}
}

// TestRunValidate_ErrorOutput_HasErrPipelinePrefix verifies that agent-mode
// validation errors are prefixed with "err:pipeline" as required by the output
// contract.
func TestRunValidate_ErrorOutput_HasErrPipelinePrefix(t *testing.T) {
	dir := setupValidateViolationProject(t)
	chdirTo(t, dir)

	output := captureStdout(t, func() {
		RunValidate([]string{}, true)
	})

	if !strings.Contains(output, "err:pipeline") {
		t.Errorf("expected err:pipeline prefix in agent-mode error output, got: %q", output)
	}
}

func TestRunValidate_PlannedFeatureSkipped(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	for _, d := range []string{
		ptsdDir,
		filepath.Join(ptsdDir, "bdd"),
		filepath.Join(ptsdDir, "seeds"),
		filepath.Join(ptsdDir, "docs"),
	} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// "planned" and "deferred" both skipped
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(
		"features:\n  - id: future\n    title: Future\n    status: planned\n  - id: old\n    title: Old\n    status: deferred\n",
	), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte("features: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// PRD with no anchors — would fail if features were checked
	if err := os.WriteFile(filepath.Join(ptsdDir, "docs", "PRD.md"), []byte("# PRD\n"), 0644); err != nil {
		t.Fatal(err)
	}

	chdirTo(t, dir)

	code := RunValidate([]string{}, true)
	if code != 0 {
		t.Errorf("expected exit 0 when only planned/deferred features exist, got %d", code)
	}
}
