package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStderr redirects os.Stderr around fn and returns whatever was written.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stderr
	os.Stderr = w
	fn()
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	r.Close()
	return buf.String()
}

// setupPipelineProject creates a minimal .ptsd/ project structure for CLI pipeline tests.
// It returns the temp directory path.
func setupPipelineProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(filepath.Join(ptsdDir, "seeds"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ptsdDir, "bdd"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ptsdDir, "docs"), 0755); err != nil {
		t.Fatal(err)
	}

	featuresYAML := "features:\n  - id: my-feat\n    status: in-progress\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresYAML), 0644); err != nil {
		t.Fatal(err)
	}

	stateYAML := "features: {}\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	configYAML := "project:\n  name: TestProject\ntesting:\n  runner: echo ok\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	return dir
}

// --- RunPrd ---

func TestRunPrdNoSubcommand(t *testing.T) {
	code := RunPrd([]string{}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for no subcommand, got %d", code)
	}
}

func TestRunPrdUnknownSubcommand(t *testing.T) {
	code := RunPrd([]string{"unknown"}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for unknown subcommand, got %d", code)
	}
}

func TestRunPrdCheckAllAnchorsPresent(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	prd := "# PRD\n<!-- feature:my-feat -->\nSection content\n"
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "docs", "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunPrd([]string{"check"}, true)
	if code != 0 {
		t.Errorf("expected exit 0 when all anchors present, got %d", code)
	}
}

func TestRunPrdCheckMissingAnchor(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	prd := "# PRD\nNo anchors here\n"
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "docs", "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunPrd([]string{"check"}, true)
	if code != 1 {
		t.Errorf("expected exit 1 when anchor missing, got %d", code)
	}
}

func TestRunPrdCheckNoPRDFile(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)
	// No PRD.md written — io error expected (err:io → exit 4)

	code := RunPrd([]string{"check"}, true)
	if code != 4 {
		t.Errorf("expected exit 4 (io error) when PRD.md missing, got %d", code)
	}
}

// --- RunSeed ---

func TestRunSeedNoSubcommand(t *testing.T) {
	code := RunSeed([]string{}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for no subcommand, got %d", code)
	}
}

func TestRunSeedAddMissingArgs(t *testing.T) {
	// "add" requires at least 2 more args (feature + file)
	code := RunSeed([]string{"add"}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for missing args, got %d", code)
	}
}

func TestRunSeedAddMissingFileArg(t *testing.T) {
	// "add feature" — still missing file argument
	code := RunSeed([]string{"add", "my-feat"}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for missing file arg, got %d", code)
	}
}

func TestRunSeedAddSuccess(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	// Initialize seed directory for the feature first
	seedDir := filepath.Join(dir, ".ptsd", "seeds", "my-feat")
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		t.Fatal(err)
	}
	seedManifest := "feature: my-feat\nfiles:\n"
	if err := os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte(seedManifest), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a source data file to add
	srcFile := filepath.Join(dir, "seed.json")
	if err := os.WriteFile(srcFile, []byte(`{"key":"value"}`), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunSeed([]string{"add", "my-feat", srcFile}, true)
	if code != 0 {
		t.Errorf("expected exit 0 on successful seed add, got %d", code)
	}

	// Verify file was copied into seed dir
	copied := filepath.Join(seedDir, "seed.json")
	if _, err := os.Stat(copied); os.IsNotExist(err) {
		t.Error("expected seed.json to be copied into seed dir")
	}
}

func TestRunSeedAddSeedNotInitialized(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	srcFile := filepath.Join(dir, "data.json")
	if err := os.WriteFile(srcFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// No seed.yaml initialized for "my-feat" — err:validation → exit 1
	var code int
	out := captureStdout(t, func() {
		code = RunSeed([]string{"add", "my-feat", srcFile}, true)
	})
	if code != 1 {
		t.Errorf("expected exit 1 (validation) when seed not initialized, got %d", code)
	}
	if !strings.Contains(out, "err:validation") {
		t.Errorf("expected output to contain err:validation, got: %q", out)
	}
}

func TestRunSeedUnknownSubcommand(t *testing.T) {
	code := RunSeed([]string{"unknown"}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for unknown subcommand, got %d", code)
	}
}

// --- RunBdd ---

func TestRunBddNoSubcommand(t *testing.T) {
	code := RunBdd([]string{}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for no subcommand, got %d", code)
	}
}

func TestRunBddUnknownSubcommand(t *testing.T) {
	code := RunBdd([]string{"unknown"}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for unknown subcommand, got %d", code)
	}
}

func TestRunBddAddMissingFeatureArg(t *testing.T) {
	// "add" with no feature specified
	code := RunBdd([]string{"add"}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for missing feature arg, got %d", code)
	}
}

func TestRunBddAddSuccess(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	// Seed must exist for bdd add to succeed
	seedDir := filepath.Join(dir, ".ptsd", "seeds", "my-feat")
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte("feature: my-feat\nfiles:\n"), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunBdd([]string{"add", "my-feat"}, true)
	if code != 0 {
		t.Errorf("expected exit 0 on successful bdd add, got %d", code)
	}

	bddPath := filepath.Join(dir, ".ptsd", "bdd", "my-feat.feature")
	if _, err := os.Stat(bddPath); os.IsNotExist(err) {
		t.Error("expected my-feat.feature to be created")
	}
}

func TestRunBddAddWithoutSeed(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)
	// No seed initialized — err:pipeline → exit 1

	var code int
	out := captureStdout(t, func() {
		code = RunBdd([]string{"add", "my-feat"}, true)
	})
	if code != 1 {
		t.Errorf("expected exit 1 (pipeline) when seed missing, got %d", code)
	}
	if !strings.Contains(out, "err:pipeline") {
		t.Errorf("expected output to contain err:pipeline, got: %q", out)
	}
}

func TestRunBddListEmpty(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	// list with no feature arg: ShowBDD reads "<dir>/.ptsd/bdd/.feature" — not found.
	// coreError maps err:validation → exit 1.
	code := RunBdd([]string{"list"}, true)
	if code != 1 {
		t.Errorf("expected exit 1 (validation error) for list with empty feature ID, got %d", code)
	}
}

func TestRunBddListSpecificFeature(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	// Write a real .feature file with scenarios
	bddDir := filepath.Join(dir, ".ptsd", "bdd")
	content := `@feature:my-feat
Feature: My Feature

  Scenario: Happy path
    Given setup done
    When action taken
    Then result visible
`
	if err := os.WriteFile(filepath.Join(bddDir, "my-feat.feature"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunBdd([]string{"list", "my-feat"}, true)
	if code != 0 {
		t.Errorf("expected exit 0 for list of existing feature, got %d", code)
	}
}

func TestRunBddListNonexistentFeature(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	// "ghost" has no .feature file — ShowBDD returns err:validation → exit 1
	var code int
	out := captureStdout(t, func() {
		code = RunBdd([]string{"list", "ghost"}, true)
	})
	if code != 1 {
		t.Errorf("expected exit 1 (validation) for nonexistent feature, got %d", code)
	}
	if !strings.Contains(out, "err:validation") {
		t.Errorf("expected output to contain err:validation, got: %q", out)
	}
}

// --- RunTest ---

func TestRunTestNoSubcommand(t *testing.T) {
	code := RunTest([]string{}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for no subcommand, got %d", code)
	}
}

func TestRunTestUnknownSubcommand(t *testing.T) {
	code := RunTest([]string{"unknown"}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for unknown subcommand, got %d", code)
	}
}

func TestRunTestMapMissingArgs(t *testing.T) {
	// "map" requires bdd-file and test-file
	code := RunTest([]string{"map"}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for missing map args, got %d", code)
	}
}

func TestRunTestMapMissingTestFileArg(t *testing.T) {
	// "map bdd-file" — missing test-file arg
	code := RunTest([]string{"map", ".ptsd/bdd/my-feat.feature"}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for missing test-file arg, got %d", code)
	}
}

func TestRunTestMapSuccess(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	// Create BDD file with feature tag
	bddDir := filepath.Join(dir, ".ptsd", "bdd")
	bddContent := "@feature:my-feat\nFeature: My Feature\n  Scenario: X\n    Given A\n"
	bddFile := ".ptsd/bdd/my-feat.feature"
	if err := os.WriteFile(filepath.Join(bddDir, "my-feat.feature"), []byte(bddContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test file that the mapping references
	testFile := "my_test.go"
	if err := os.WriteFile(filepath.Join(dir, testFile), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunTest([]string{"map", bddFile, testFile}, true)
	if code != 0 {
		t.Errorf("expected exit 0 on successful map, got %d", code)
	}
}

func TestRunTestMapBDDNotFound(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	// BDD file does not exist — err:io → exit 4
	var code int
	out := captureStdout(t, func() {
		code = RunTest([]string{"map", ".ptsd/bdd/ghost.feature", "some_test.go"}, true)
	})
	if code != 4 {
		t.Errorf("expected exit 4 (io) when bdd file missing, got %d", code)
	}
	if !strings.Contains(out, "err:io") {
		t.Errorf("expected output to contain err:io, got: %q", out)
	}
}

func TestRunTestRunAllTests(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)
	// Config already sets runner: "echo ok" which exits 0.
	// The exit-code adapter treats exit 0 as 1 pass.

	code := RunTest([]string{"run"}, true)
	if code != 0 {
		t.Errorf("expected exit 0 when runner succeeds, got %d", code)
	}
}

func TestRunTestRunNoRunnerConfigured(t *testing.T) {
	dir := t.TempDir()
	chdirTo(t, dir)

	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Config with no testing.runner
	configYAML := "project:\n  name: TestProject\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}
	stateYAML := "features: {}\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunTest([]string{"run"}, true)
	// err:config → exit code 3
	if code != 3 {
		t.Errorf("expected exit 3 (config error) when no runner configured, got %d", code)
	}
}

func TestRunTestRunWithFeatureFilter(t *testing.T) {
	dir := t.TempDir()
	chdirTo(t, dir)

	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(filepath.Join(ptsdDir, "bdd"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0755); err != nil {
		t.Fatal(err)
	}

	// Write a passing test script
	runnerScript := "#!/bin/sh\necho 'ok 1 - pass'\nexit 0\n"
	runnerPath := filepath.Join(dir, "tests", "run.sh")
	if err := os.WriteFile(runnerPath, []byte(runnerScript), 0755); err != nil {
		t.Fatal(err)
	}

	configYAML := "project:\n  name: TestProject\ntesting:\n  runner: ./tests/run.sh\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// State with test mapping for feature "my-feat"
	testFile := "tests/my_test.go"
	if err := os.WriteFile(filepath.Join(dir, testFile), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	stateYAML := "features:\n  my-feat:\n    tests:\n      - .ptsd/bdd/my-feat.feature::" + testFile + "\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunTest([]string{"run", "my-feat"}, true)
	if code != 0 {
		t.Errorf("expected exit 0 when tests pass for feature filter, got %d", code)
	}
}

func TestRunTestRunWithFailingTests(t *testing.T) {
	dir := t.TempDir()
	chdirTo(t, dir)

	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0755); err != nil {
		t.Fatal(err)
	}

	// Script that outputs a TAP failure
	runnerScript := "#!/bin/sh\necho 'not ok 1 - fail'\nexit 1\n"
	runnerPath := filepath.Join(dir, "tests", "run.sh")
	if err := os.WriteFile(runnerPath, []byte(runnerScript), 0755); err != nil {
		t.Fatal(err)
	}

	configYAML := "project:\n  name: TestProject\ntesting:\n  runner: ./tests/run.sh\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	stateYAML := "features: {}\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunTest([]string{"run"}, true)
	// Failed tests → exit 5
	if code != 5 {
		t.Errorf("expected exit 5 when tests fail, got %d", code)
	}
}

// --- Issue 1: Orphaned anchor test ---

func TestRunPrdCheckOrphanedAnchor(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	// PRD has anchor for "ghost" which is not in features.yaml.
	// features.yaml only has "my-feat" but there is no anchor for it either.
	prd := "# PRD\n<!-- feature:ghost -->\nSection content\n"
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "docs", "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	var code int
	out := captureStderr(t, func() {
		code = RunPrd([]string{"check"}, true)
	})
	if code != 1 {
		t.Errorf("expected exit 1 when orphaned anchor present, got %d", code)
	}
	if !strings.Contains(out, "err:pipeline") {
		t.Errorf("expected stderr to contain err:pipeline, got: %q", out)
	}
	if !strings.Contains(out, "orphaned-anchor") {
		t.Errorf("expected stderr to contain orphaned-anchor, got: %q", out)
	}
	if !strings.Contains(out, "ghost") {
		t.Errorf("expected stderr to mention ghost feature ID, got: %q", out)
	}
}

// --- Issue 2: Human mode tests ---

func TestRunPrdHumanModeSuccess(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	prd := "# PRD\n<!-- feature:my-feat -->\nSection content\n"
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "docs", "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	// agentMode=false: success message is "PRD anchors OK"
	out := captureStdout(t, func() {
		code := RunPrd([]string{"check"}, false)
		if code != 0 {
			t.Errorf("expected exit 0 in human mode for valid PRD, got %d", code)
		}
	})
	if !strings.Contains(out, "PRD anchors OK") {
		t.Errorf("expected human mode output to contain 'PRD anchors OK', got: %q", out)
	}
}

func TestRunPrdHumanModeError(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	// PRD missing anchor for my-feat → errors written to stderr, exit 1
	prd := "# PRD\nNo anchors here\n"
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "docs", "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunPrd([]string{"check"}, false)
	if code != 1 {
		t.Errorf("expected exit 1 in human mode for missing anchor, got %d", code)
	}
}

func TestRunSeedHumanModeSuccess(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	seedDir := filepath.Join(dir, ".ptsd", "seeds", "my-feat")
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte("feature: my-feat\nfiles:\n"), 0644); err != nil {
		t.Fatal(err)
	}

	srcFile := filepath.Join(dir, "human_seed.json")
	if err := os.WriteFile(srcFile, []byte(`{"x":1}`), 0644); err != nil {
		t.Fatal(err)
	}

	// agentMode=false: success message is "Added seed file ... to feature ..."
	out := captureStdout(t, func() {
		code := RunSeed([]string{"add", "my-feat", srcFile}, false)
		if code != 0 {
			t.Errorf("expected exit 0 in human mode for seed add, got %d", code)
		}
	})
	if !strings.Contains(out, "Added seed file") {
		t.Errorf("expected human mode output to contain 'Added seed file', got: %q", out)
	}
}

func TestRunSeedHumanModeError(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	srcFile := filepath.Join(dir, "noop.json")
	if err := os.WriteFile(srcFile, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	// No seed initialized — err:validation → exit 1 in human mode too
	var code int
	out := captureStdout(t, func() {
		code = RunSeed([]string{"add", "my-feat", srcFile}, false)
	})
	if code != 1 {
		t.Errorf("expected exit 1 in human mode when seed not initialized, got %d", code)
	}
	if !strings.Contains(out, "err:validation") {
		t.Errorf("expected human mode error output to contain err:validation, got: %q", out)
	}
}

func TestRunBddHumanModeSuccess(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	seedDir := filepath.Join(dir, ".ptsd", "seeds", "my-feat")
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte("feature: my-feat\nfiles:\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// agentMode=false: success message is "BDD scaffold created for feature ..."
	out := captureStdout(t, func() {
		code := RunBdd([]string{"add", "my-feat"}, false)
		if code != 0 {
			t.Errorf("expected exit 0 in human mode for bdd add, got %d", code)
		}
	})
	if !strings.Contains(out, "BDD scaffold created") {
		t.Errorf("expected human mode output to contain 'BDD scaffold created', got: %q", out)
	}
}

func TestRunBddHumanModeError(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	// No seed initialized — err:pipeline → exit 1 in human mode too
	var code int
	out := captureStdout(t, func() {
		code = RunBdd([]string{"add", "my-feat"}, false)
	})
	if code != 1 {
		t.Errorf("expected exit 1 in human mode when seed missing, got %d", code)
	}
	if !strings.Contains(out, "err:pipeline") {
		t.Errorf("expected human mode error output to contain err:pipeline, got: %q", out)
	}
}

func TestRunTestHumanModeSuccess(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)
	// Config sets runner: "echo ok" → exit 0, treated as 1 pass.

	out := captureStdout(t, func() {
		code := RunTest([]string{"run"}, false)
		if code != 0 {
			t.Errorf("expected exit 0 in human mode when runner succeeds, got %d", code)
		}
	})
	// AgentRenderer is used for both modes; output should contain pass count
	if !strings.Contains(out, "pass:1") {
		t.Errorf("expected human mode output to contain pass:1, got: %q", out)
	}
}

func TestRunTestHumanModeError(t *testing.T) {
	dir := t.TempDir()
	chdirTo(t, dir)

	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	// No testing.runner configured — err:config → exit 3
	configYAML := "project:\n  name: TestProject\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}
	stateYAML := "features: {}\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	var code int
	out := captureStdout(t, func() {
		code = RunTest([]string{"run"}, false)
	})
	if code != 3 {
		t.Errorf("expected exit 3 (config) in human mode when no runner configured, got %d", code)
	}
	if !strings.Contains(out, "err:config") {
		t.Errorf("expected human mode error output to contain err:config, got: %q", out)
	}
}

// --- Issue 4: seed add with optional type and description args ---

func TestRunSeedAddWithTypeAndDescription(t *testing.T) {
	dir := setupPipelineProject(t)
	chdirTo(t, dir)

	seedDir := filepath.Join(dir, ".ptsd", "seeds", "my-feat")
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte("feature: my-feat\nfiles:\n"), 0644); err != nil {
		t.Fatal(err)
	}

	srcFile := filepath.Join(dir, "fixture.json")
	if err := os.WriteFile(srcFile, []byte(`{"key":"value"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Provide optional type "fixture" and description "test description"
	code := RunSeed([]string{"add", "my-feat", srcFile, "fixture", "test description"}, true)
	if code != 0 {
		t.Errorf("expected exit 0 for seed add with type and description, got %d", code)
	}

	// Verify the seed manifest includes type and description
	manifestData, err := os.ReadFile(filepath.Join(seedDir, "seed.yaml"))
	if err != nil {
		t.Fatalf("failed to read seed manifest: %v", err)
	}
	manifest := string(manifestData)
	if !strings.Contains(manifest, "type: fixture") {
		t.Errorf("expected manifest to contain 'type: fixture', got:\n%s", manifest)
	}
	if !strings.Contains(manifest, "test description") {
		t.Errorf("expected manifest to contain description 'test description', got:\n%s", manifest)
	}
}
