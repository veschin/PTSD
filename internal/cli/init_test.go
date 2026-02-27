package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureOutput redirects os.Stdout during fn execution and returns the captured output.
func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// setupGitRepo creates a minimal git repository (.git directory) in dir.
func setupGitRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatalf("setupGitRepo: failed to create .git: %v", err)
	}
	// git requires HEAD to exist to be usable in many contexts
	hooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("setupGitRepo: failed to create hooks dir: %v", err)
	}
}

// chdirTemp changes the working directory to dir for the duration of the test.
// It restores the original directory after the test completes.
func chdirTemp(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("chdirTemp: failed to get cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdirTemp: failed to chdir to %s: %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Logf("chdirTemp: failed to restore cwd: %v", err)
		}
	})
}

// --- RunInit tests ---

// TestRunInitSuccessNoName covers BDD: "Init new project" without a name argument.
// RunInit with no args should succeed, default the project name to the dir basename,
// and return exit code 0.
func TestRunInitSuccessNoName(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)
	chdirTemp(t, dir)

	var output string
	code := -1
	output = captureOutput(func() {
		code = RunInit([]string{}, true)
	})

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %q", code, output)
	}

	// Agent mode: must print init:ok
	if !strings.Contains(output, "init:ok") {
		t.Errorf("expected 'init:ok' in agent output, got: %q", output)
	}

	// .ptsd/ must have been created
	ptsdDir := filepath.Join(dir, ".ptsd")
	if _, err := os.Stat(ptsdDir); err != nil {
		t.Errorf(".ptsd/ not created: %v", err)
	}
}

// TestRunInitSuccessWithName covers BDD: "Init new project" with a name argument.
// RunInit with name arg should embed the name in ptsd.yaml and return exit code 0.
func TestRunInitSuccessWithName(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)
	chdirTemp(t, dir)

	var output string
	code := -1
	output = captureOutput(func() {
		code = RunInit([]string{"MyApp"}, true)
	})

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %q", code, output)
	}

	// Agent mode output must say init:ok and include the dir path.
	if !strings.Contains(output, "init:ok") {
		t.Errorf("expected 'init:ok' in agent output, got: %q", output)
	}
	if !strings.Contains(output, dir) {
		t.Errorf("expected dir path %q in agent output, got: %q", dir, output)
	}

	// ptsd.yaml must contain the project name.
	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "ptsd.yaml"))
	if err != nil {
		t.Fatalf("ptsd.yaml not found: %v", err)
	}
	if !strings.Contains(string(data), "MyApp") {
		t.Errorf("ptsd.yaml does not contain 'MyApp':\n%s", data)
	}
}

// TestRunInitAgentModeOutputFormat verifies the exact agent mode output format.
func TestRunInitAgentModeOutputFormat(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)
	chdirTemp(t, dir)

	var output string
	output = captureOutput(func() {
		RunInit([]string{"MyApp"}, true)
	})

	// Agent output format: "init:ok dir:<path>"
	expected := "init:ok dir:" + dir
	if !strings.Contains(output, expected) {
		t.Errorf("agent output format mismatch.\nexpected to contain: %q\ngot: %q", expected, output)
	}
}

// TestRunInitHumanModeOutputFormat verifies the human mode output format.
func TestRunInitHumanModeOutputFormat(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)
	chdirTemp(t, dir)

	var output string
	output = captureOutput(func() {
		RunInit([]string{}, false)
	})

	if !strings.Contains(output, "Initialized ptsd project in") {
		t.Errorf("expected human-mode message in output, got: %q", output)
	}
	if !strings.Contains(output, dir) {
		t.Errorf("expected dir path in human output, got: %q", output)
	}
}

// TestRunInitRefusesWithoutGit covers BDD: "Init refuses without git".
// A directory without .git must return exit code 3 (config error).
func TestRunInitRefusesWithoutGit(t *testing.T) {
	dir := t.TempDir()
	// No .git directory.
	chdirTemp(t, dir)

	var output string
	code := -1
	output = captureStderr(t, func() {
		code = RunInit([]string{}, true)
	})

	if code != 3 {
		t.Errorf("expected exit code 3 (config), got %d. output: %q", code, output)
	}
	if !strings.Contains(output, "err:config") {
		t.Errorf("expected 'err:config' in output, got: %q", output)
	}
	if !strings.Contains(output, "git repository required") {
		t.Errorf("expected 'git repository required' in output, got: %q", output)
	}
}

// TestRunInitReInitSucceeds covers BDD: "Re-init existing project regenerates hooks and skills".
// Re-init of an already-initialized directory must succeed and print reinit:ok.
func TestRunInitReInitSucceeds(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)
	chdirTemp(t, dir)

	// First init — must succeed.
	captureOutput(func() {
		if code := RunInit([]string{}, true); code != 0 {
			t.Fatalf("first RunInit failed, expected 0")
		}
	})

	// Second init — must succeed as re-init.
	output := captureOutput(func() {
		code := RunInit([]string{}, true)
		if code != 0 {
			t.Errorf("expected exit code 0 for re-init, got %d", code)
		}
	})

	if !strings.Contains(output, "reinit:ok") {
		t.Errorf("expected 'reinit:ok' in output, got: %q", output)
	}
}

// TestRunInitCreatesExpectedFiles verifies the full set of scaffolded artifacts
// and checks that key files contain expected content.
func TestRunInitCreatesExpectedFiles(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)
	chdirTemp(t, dir)

	captureOutput(func() {
		if code := RunInit([]string{"MyApp"}, true); code != 0 {
			t.Fatalf("RunInit failed")
		}
	})

	required := []string{
		".ptsd/ptsd.yaml",
		".ptsd/features.yaml",
		".ptsd/state.yaml",
		".ptsd/tasks.yaml",
		".ptsd/review-status.yaml",
		".ptsd/docs/PRD.md",
		"CLAUDE.md",
		".git/hooks/pre-commit",
	}
	for _, rel := range required {
		path := filepath.Join(dir, rel)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", rel, err)
		}
	}

	// PRD.md must contain template heading.
	prdData, err := os.ReadFile(filepath.Join(dir, ".ptsd", "docs", "PRD.md"))
	if err != nil {
		t.Fatalf("PRD.md not readable: %v", err)
	}
	if !strings.Contains(string(prdData), "# ") {
		t.Errorf("PRD.md expected to start with '# ' heading, got:\n%s", prdData)
	}

	// state.yaml must have a proper header.
	stateData, err := os.ReadFile(filepath.Join(dir, ".ptsd", "state.yaml"))
	if err != nil {
		t.Fatalf("state.yaml not readable: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(stateData)), "features:") {
		t.Errorf("state.yaml expected to start with 'features:', got:\n%s", stateData)
	}
}

// TestRunInitCreatesSkills covers BDD: "Init generates skills".
func TestRunInitCreatesSkills(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)
	chdirTemp(t, dir)

	captureOutput(func() {
		if code := RunInit([]string{"MyApp"}, true); code != 0 {
			t.Fatalf("RunInit failed")
		}
	})

	skills := []string{
		"write-prd.md", "write-seed.md", "write-bdd.md",
		"write-tests.md", "write-impl.md",
		"review-prd.md", "review-seed.md", "review-bdd.md",
		"review-tests.md", "review-impl.md",
		"create-tasks.md", "workflow.md",
	}
	for _, skill := range skills {
		path := filepath.Join(dir, ".ptsd", "skills", skill)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected skill file .ptsd/skills/%s to exist: %v", skill, err)
		}
	}
}

// TestRunInitDetectsVitestRunner covers BDD: "Init detects test runner" (vitest).
func TestRunInitDetectsVitestRunner(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)

	pkgJSON := `{"devDependencies": {"vitest": "^1.0.0"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	chdirTemp(t, dir)

	captureOutput(func() {
		if code := RunInit([]string{}, true); code != 0 {
			t.Fatalf("RunInit failed")
		}
	})

	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "ptsd.yaml"))
	if err != nil {
		t.Fatalf("ptsd.yaml not found: %v", err)
	}
	if !strings.Contains(string(data), "npx vitest run") {
		t.Errorf("ptsd.yaml should contain 'npx vitest run', got:\n%s", data)
	}
}

// --- RunAdopt tests ---

// TestRunAdoptDryRun covers BDD: "Dry run shows what would happen".
// --dry-run must not create any files and must print discovered artifacts.
func TestRunAdoptDryRun(t *testing.T) {
	dir := t.TempDir()

	// Create a .feature file with a tag.
	featureContent := "@feature:login\nFeature: Login\n  Scenario: User logs in\n"
	if err := os.WriteFile(filepath.Join(dir, "login.feature"), []byte(featureContent), 0644); err != nil {
		t.Fatalf("failed to write .feature file: %v", err)
	}

	chdirTemp(t, dir)

	var output string
	code := -1
	output = captureOutput(func() {
		code = RunAdopt([]string{"--dry-run"}, true)
	})

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %q", code, output)
	}

	// .ptsd/ must NOT be created.
	if _, err := os.Stat(filepath.Join(dir, ".ptsd")); err == nil {
		t.Error("dry-run must not create .ptsd/")
	}

	// Original .feature file must still be in place.
	if _, err := os.Stat(filepath.Join(dir, "login.feature")); err != nil {
		t.Error("dry-run must not move/remove .feature files")
	}

	// Agent output must contain discovered counts.
	if !strings.Contains(output, "dry-run:ok") {
		t.Errorf("expected 'dry-run:ok' in agent output, got: %q", output)
	}
	// bdd:1 because we have one feature tag
	if !strings.Contains(output, "bdd:1") {
		t.Errorf("expected 'bdd:1' in agent output, got: %q", output)
	}
}

// TestRunAdoptDryRunHumanMode verifies human-mode output for dry run with actual artifacts.
// The dir contains a .feature file and a _test.go file so that discovery is exercised.
func TestRunAdoptDryRunHumanMode(t *testing.T) {
	dir := t.TempDir()

	// Plant a .feature file with a @feature tag.
	featureContent := "@feature:payments\nFeature: Payments\n  Scenario: User pays\n"
	if err := os.WriteFile(filepath.Join(dir, "payments.feature"), []byte(featureContent), 0644); err != nil {
		t.Fatalf("failed to write .feature file: %v", err)
	}

	// Plant a _test.go file.
	testDir := filepath.Join(dir, "pkg")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "payments_test.go"), []byte("package pkg\n"), 0644); err != nil {
		t.Fatalf("failed to write _test.go file: %v", err)
	}

	chdirTemp(t, dir)

	var output string
	output = captureOutput(func() {
		RunAdopt([]string{"--dry-run"}, false)
	})

	// Human mode must mention "Dry run".
	if !strings.Contains(output, "Dry run") {
		t.Errorf("expected 'Dry run' in human output, got: %q", output)
	}
	// BDD count must reflect the discovered feature file.
	if !strings.Contains(output, "BDD features found: 1") {
		t.Errorf("expected 'BDD features found: 1' in human output, got: %q", output)
	}
	// Test file count must reflect the discovered _test.go.
	if !strings.Contains(output, "Test files found: 1") {
		t.Errorf("expected 'Test files found: 1' in human output, got: %q", output)
	}
}

// TestRunAdoptSuccess covers BDD: "Adopt discovers BDD files" + "Adopt discovers test files".
// RunAdopt without --dry-run must create .ptsd/ and move .feature files.
func TestRunAdoptSuccess(t *testing.T) {
	dir := t.TempDir()

	// Create a .feature file.
	featureContent := "@feature:auth\nFeature: Auth\n  Scenario: Login\n"
	if err := os.WriteFile(filepath.Join(dir, "auth.feature"), []byte(featureContent), 0644); err != nil {
		t.Fatalf("failed to write .feature file: %v", err)
	}

	chdirTemp(t, dir)

	var output string
	code := -1
	output = captureOutput(func() {
		code = RunAdopt([]string{}, true)
	})

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %q", code, output)
	}

	// .ptsd/ must be created.
	if _, err := os.Stat(filepath.Join(dir, ".ptsd")); err != nil {
		t.Errorf(".ptsd/ not created: %v", err)
	}

	// Agent mode output.
	if !strings.Contains(output, "adopt:ok") {
		t.Errorf("expected 'adopt:ok' in agent output, got: %q", output)
	}

	// .feature file must have been moved to .ptsd/bdd/.
	movedPath := filepath.Join(dir, ".ptsd", "bdd", "auth.feature")
	if _, err := os.Stat(movedPath); err != nil {
		t.Errorf("expected .feature file moved to .ptsd/bdd/: %v", err)
	}
	origPath := filepath.Join(dir, "auth.feature")
	if _, err := os.Stat(origPath); err == nil {
		t.Error("original .feature file must be removed after adopt")
	}

	// features.yaml must reference the discovered feature ID.
	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "features.yaml"))
	if err != nil {
		t.Fatalf("features.yaml not found: %v", err)
	}
	if !strings.Contains(string(data), "auth") {
		t.Errorf("features.yaml missing 'auth', got:\n%s", data)
	}
}

// TestRunAdoptSuccessAgentOutputFormat verifies adopt agent output contains dir.
func TestRunAdoptSuccessAgentOutputFormat(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	var output string
	output = captureOutput(func() {
		RunAdopt([]string{}, true)
	})

	expected := "adopt:ok dir:" + dir
	if !strings.Contains(output, expected) {
		t.Errorf("agent output format mismatch.\nexpected to contain: %q\ngot: %q", expected, output)
	}
}

// TestRunAdoptSuccessHumanMode verifies human-mode adopt output.
func TestRunAdoptSuccessHumanMode(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	var output string
	output = captureOutput(func() {
		RunAdopt([]string{}, false)
	})

	if !strings.Contains(output, "Adopted project in") {
		t.Errorf("expected 'Adopted project in' in human output, got: %q", output)
	}
	if !strings.Contains(output, dir) {
		t.Errorf("expected dir path in human output, got: %q", output)
	}
}

// TestRunAdoptRefusesIfAlreadyInitialized covers BDD: "Adopt refuses if .ptsd already initialized".
// Returns exit code 1 (validation error).
func TestRunAdoptRefusesIfAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()

	// Pre-create .ptsd/ to simulate already-initialized state.
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatalf("failed to pre-create .ptsd: %v", err)
	}
	// Write a minimal features.yaml so it's a valid .ptsd dir.
	if err := os.WriteFile(
		filepath.Join(dir, ".ptsd", "features.yaml"),
		[]byte("features: []\n"),
		0644,
	); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	chdirTemp(t, dir)

	var output string
	code := -1
	output = captureStderr(t, func() {
		code = RunAdopt([]string{}, true)
	})

	if code != 1 {
		t.Errorf("expected exit code 1 (validation), got %d. output: %q", code, output)
	}
	if !strings.Contains(output, "err:validation") {
		t.Errorf("expected 'err:validation' in output, got: %q", output)
	}
	if !strings.Contains(output, "already initialized") {
		t.Errorf("expected 'already initialized' in output, got: %q", output)
	}
}

// TestRunAdoptDryRunRefusesIfAlreadyInitialized verifies dry-run also checks .ptsd existence.
func TestRunAdoptDryRunRefusesIfAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatalf("failed to pre-create .ptsd: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, ".ptsd", "features.yaml"),
		[]byte("features: []\n"),
		0644,
	); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	chdirTemp(t, dir)

	var output string
	code := -1
	output = captureStderr(t, func() {
		code = RunAdopt([]string{"--dry-run"}, true)
	})

	if code != 1 {
		t.Errorf("expected exit code 1 (validation) for dry-run on initialized project, got %d. output: %q", code, output)
	}
	if !strings.Contains(output, "err:validation") {
		t.Errorf("expected 'err:validation' in output, got: %q", output)
	}
}

// TestRunAdoptEmptyProject verifies adopt works on a project with no artifacts.
func TestRunAdoptEmptyProject(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	code := -1
	captureOutput(func() {
		code = RunAdopt([]string{}, true)
	})

	if code != 0 {
		t.Errorf("expected exit code 0 for empty project adopt, got %d", code)
	}

	// features.yaml must be created.
	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "features.yaml"))
	if err != nil {
		t.Fatalf("features.yaml not created: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(data)), "features:") {
		t.Errorf("features.yaml missing 'features:' header, got:\n%s", data)
	}
}

// TestRunAdoptDryRunReportsTestFiles verifies dry-run counts test files correctly.
func TestRunAdoptDryRunReportsTestFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a test file.
	srcDir := filepath.Join(dir, "internal")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "auth_test.go"), []byte("package internal\n"), 0644); err != nil {
		t.Fatal(err)
	}

	chdirTemp(t, dir)

	var output string
	output = captureOutput(func() {
		RunAdopt([]string{"--dry-run"}, true)
	})

	// Agent mode: "dry-run:ok bdd:0 tests:1 features:..."
	if !strings.Contains(output, "tests:1") {
		t.Errorf("expected 'tests:1' in agent dry-run output, got: %q", output)
	}
}

// TestRunAdoptIgnoresUnknownArgs verifies that unknown args are silently ignored
// (only --dry-run is special).
func TestRunAdoptIgnoresUnknownArgs(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	code := -1
	captureOutput(func() {
		code = RunAdopt([]string{"--unknown-flag"}, true)
	})

	// Should still succeed (non-dry-run path), not fail on unknown flag.
	if code != 0 {
		t.Errorf("expected exit code 0 when unknown arg given, got %d", code)
	}
}

// TestRunAdoptWithTestFilesExitsZeroAndMentionsAdoption covers BDD: "Adopt discovers test files".
// RunAdopt (full, not dry-run) on a directory with a _test.go file must return exit 0
// and emit "adopt:ok" in agent output.
func TestRunAdoptWithTestFilesExitsZeroAndMentionsAdoption(t *testing.T) {
	dir := t.TempDir()

	// Create a _test.go file so discovery has something to find.
	srcDir := filepath.Join(dir, "internal")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "user_test.go"), []byte("package internal\n"), 0644); err != nil {
		t.Fatalf("failed to write _test.go: %v", err)
	}

	chdirTemp(t, dir)

	var output string
	code := -1
	output = captureOutput(func() {
		code = RunAdopt([]string{}, true)
	})

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. output: %q", code, output)
	}
	if !strings.Contains(output, "adopt:ok") {
		t.Errorf("expected 'adopt:ok' in agent output, got: %q", output)
	}

	// .ptsd/ must have been created.
	if _, err := os.Stat(filepath.Join(dir, ".ptsd")); err != nil {
		t.Errorf(".ptsd/ not created: %v", err)
	}
}

// TestRunInitHumanModeErrorNoGit verifies that RunInit in human mode returns exit code 3
// when no git repository is present and prints a human-readable error.
func TestRunInitHumanModeErrorNoGit(t *testing.T) {
	dir := t.TempDir()
	// Deliberately no .git directory.
	chdirTemp(t, dir)

	var output string
	code := -1
	output = captureStderr(t, func() {
		code = RunInit([]string{}, false)
	})

	if code != 3 {
		t.Errorf("expected exit code 3 (config), got %d. output: %q", code, output)
	}
	// Human mode must still communicate the error — exact message may vary,
	// but must mention "git repository".
	if !strings.Contains(output, "git repository") {
		t.Errorf("expected 'git repository' in human-mode error output, got: %q", output)
	}
}

// TestRunInitHumanModeReInit verifies that RunInit in human mode succeeds on re-init
// and prints a human-readable re-init message.
func TestRunInitHumanModeReInit(t *testing.T) {
	dir := t.TempDir()
	setupGitRepo(t, dir)
	chdirTemp(t, dir)

	// First init — must succeed.
	captureOutput(func() {
		if code := RunInit([]string{}, false); code != 0 {
			t.Fatalf("first RunInit failed")
		}
	})

	// Second init in human mode — must succeed as re-init.
	output := captureOutput(func() {
		code := RunInit([]string{}, false)
		if code != 0 {
			t.Errorf("expected exit code 0 for re-init, got %d", code)
		}
	})

	if !strings.Contains(output, "Re-initialized") {
		t.Errorf("expected 'Re-initialized' in output, got: %q", output)
	}
}

// TestRunAdoptHumanModeErrorAlreadyInitialized verifies that RunAdopt in human mode
// returns exit code 1 when .ptsd already exists and prints a human-readable error.
func TestRunAdoptHumanModeErrorAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()

	// Pre-create .ptsd/ to simulate already-initialized state.
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatalf("failed to pre-create .ptsd: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(dir, ".ptsd", "features.yaml"),
		[]byte("features: []\n"),
		0644,
	); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	chdirTemp(t, dir)

	var output string
	code := -1
	output = captureStderr(t, func() {
		code = RunAdopt([]string{}, false)
	})

	if code != 1 {
		t.Errorf("expected exit code 1 (validation), got %d. output: %q", code, output)
	}
	// Human mode must communicate that adoption failed.
	if !strings.Contains(output, "already initialized") {
		t.Errorf("expected 'already initialized' in human-mode error output, got: %q", output)
	}
}

// TestRunInitDetectsTestRunner covers BDD: "Init detects test runner" for go.mod, jest, and pytest.
func TestRunInitDetectsTestRunner(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(dir string)
		expectedRunner string
	}{
		{
			name: "go.mod",
			setupFunc: func(dir string) {
				if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\ngo 1.21\n"), 0644); err != nil {
					t.Fatalf("failed to write go.mod: %v", err)
				}
			},
			expectedRunner: "go test ./...",
		},
		{
			name: "jest",
			setupFunc: func(dir string) {
				pkgJSON := `{"devDependencies": {"jest": "^29.0.0"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
					t.Fatalf("failed to write package.json: %v", err)
				}
			},
			expectedRunner: "npx jest",
		},
		{
			name: "pytest.ini",
			setupFunc: func(dir string) {
				if err := os.WriteFile(filepath.Join(dir, "pytest.ini"), []byte("[pytest]\n"), 0644); err != nil {
					t.Fatalf("failed to write pytest.ini: %v", err)
				}
			},
			expectedRunner: "pytest",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			setupGitRepo(t, dir)
			tc.setupFunc(dir)
			chdirTemp(t, dir)

			captureOutput(func() {
				if code := RunInit([]string{}, true); code != 0 {
					t.Fatalf("RunInit failed")
				}
			})

			data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "ptsd.yaml"))
			if err != nil {
				t.Fatalf("ptsd.yaml not found: %v", err)
			}
			if !strings.Contains(string(data), tc.expectedRunner) {
				t.Errorf("ptsd.yaml should contain runner %q, got:\n%s", tc.expectedRunner, data)
			}
		})
	}
}
