package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/veschin/ptsd/internal/core"
)

// setupHooksProject creates a minimal .ptsd project with an initialised git
// repository so that GeneratePreCommitHook can write .git/hooks/pre-commit.
func setupHooksProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Init a bare-minimum git repo (just the .git/hooks directory is enough).
	gitHooksDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(gitHooksDir, 0755); err != nil {
		t.Fatalf("mkdir .git/hooks: %v", err)
	}

	// .ptsd structure required by core helpers called during hooks install
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatalf("mkdir .ptsd: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644); err != nil {
		t.Fatalf("write ptsd.yaml: %v", err)
	}

	return dir
}

func TestRunHooks_NoSubcommand_Exit2(t *testing.T) {
	dir := setupHooksProject(t)
	chdirTo(t, dir)

	code := RunHooks([]string{}, true)
	if code != 2 {
		t.Errorf("expected exit 2 when no subcommand given, got %d", code)
	}
}

func TestRunHooks_NoSubcommand_HumanMode_Exit2(t *testing.T) {
	dir := setupHooksProject(t)
	chdirTo(t, dir)

	code := RunHooks([]string{}, false)
	if code != 2 {
		t.Errorf("expected exit 2 in human mode when no subcommand given, got %d", code)
	}
}

func TestRunHooks_UnknownSubcommand_Exit2(t *testing.T) {
	dir := setupHooksProject(t)
	chdirTo(t, dir)

	code := RunHooks([]string{"unknown"}, true)
	if code != 2 {
		t.Errorf("expected exit 2 for unknown subcommand, got %d", code)
	}
}

func TestRunHooks_UnknownSubcommand_HumanMode_Exit2(t *testing.T) {
	dir := setupHooksProject(t)
	chdirTo(t, dir)

	code := RunHooks([]string{"bogus"}, false)
	if code != 2 {
		t.Errorf("expected exit 2 for unknown subcommand in human mode, got %d", code)
	}
}

func TestRunHooks_Install_Exit0(t *testing.T) {
	dir := setupHooksProject(t)
	chdirTo(t, dir)

	code := RunHooks([]string{"install"}, true)
	if code != 0 {
		t.Errorf("expected exit 0 for hooks install, got %d", code)
	}
}

func TestRunHooks_Install_CreatesHookFile(t *testing.T) {
	dir := setupHooksProject(t)
	chdirTo(t, dir)

	code := RunHooks([]string{"install"}, true)
	if code != 0 {
		t.Fatalf("expected exit 0 for hooks install, got %d", code)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("expected pre-commit hook to exist at %s: %v", hookPath, err)
	}

	content := string(data)
	if !strings.Contains(content, "validate") {
		t.Errorf("expected hook to contain 'validate', got: %q", content)
	}
}

func TestRunHooks_Install_HookIsExecutable(t *testing.T) {
	dir := setupHooksProject(t)
	chdirTo(t, dir)

	code := RunHooks([]string{"install"}, true)
	if code != 0 {
		t.Fatalf("expected exit 0 for hooks install, got %d", code)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("stat pre-commit hook: %v", err)
	}

	if info.Mode()&0111 == 0 {
		t.Errorf("expected pre-commit hook to be executable, mode: %v", info.Mode())
	}
}

func TestRunHooks_Install_HumanMode_Exit0(t *testing.T) {
	dir := setupHooksProject(t)
	chdirTo(t, dir)

	code := RunHooks([]string{"install"}, false)
	if code != 0 {
		t.Errorf("expected exit 0 for hooks install in human mode, got %d", code)
	}
}

func TestRunHooks_NoSubcommand_PrintsUsage(t *testing.T) {
	dir := setupHooksProject(t)
	chdirTo(t, dir)

	// Capture stderr to verify usage message
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	RunHooks([]string{}, true)

	w.Close()
	os.Stderr = origStderr

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	errOutput := string(buf[:n])

	if !strings.Contains(errOutput, "hooks") {
		t.Errorf("expected usage message mentioning 'hooks' on stderr, got: %q", errOutput)
	}
}

func TestRunHooks_UnknownSubcommand_PrintsError(t *testing.T) {
	dir := setupHooksProject(t)
	chdirTo(t, dir)

	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	RunHooks([]string{"nope"}, true)

	w.Close()
	os.Stderr = origStderr

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	errOutput := string(buf[:n])

	if !strings.Contains(errOutput, "nope") && !strings.Contains(errOutput, "unknown") {
		t.Errorf("expected error mentioning unknown subcommand on stderr, got: %q", errOutput)
	}
}

// TestRunHooks_Install_IOError_Exit4 verifies exit code 4 when the .git/hooks
// directory cannot be created because .git is a plain file rather than a directory.
// This exercises the io error path (exit code 4) in runHooksInstall.
func TestRunHooks_Install_IOError_Exit4(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatalf("mkdir .ptsd: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte("version: \"1\"\n"), 0644); err != nil {
		t.Fatalf("write ptsd.yaml: %v", err)
	}

	// Create .git as a regular file so os.MkdirAll(".git/hooks") fails with an io error.
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("not a directory\n"), 0644); err != nil {
		t.Fatalf("write fake .git file: %v", err)
	}

	chdirTo(t, dir)

	code := RunHooks([]string{"install"}, true)
	if code != 4 {
		t.Errorf("expected exit 4 (io error) when .git is not a directory, got %d", code)
	}
}

// --- extractFilePath tests (stdin JSON parsing edge cases) ---

func TestExtractFilePath_ValidJSON(t *testing.T) {
	input := `{"tool": "Edit", "file_path": "/home/user/project/src/main.go", "content": "..."}`
	result := extractFilePathFromReader(strings.NewReader(input))
	if result != "/home/user/project/src/main.go" {
		t.Errorf("expected /home/user/project/src/main.go, got %q", result)
	}
}

func TestExtractFilePath_EmptyInput(t *testing.T) {
	result := extractFilePathFromReader(strings.NewReader(""))
	if result != "" {
		t.Errorf("expected empty string for empty input, got %q", result)
	}
}

func TestExtractFilePath_NoFilePathKey(t *testing.T) {
	input := `{"tool": "Bash", "command": "ls -la"}`
	result := extractFilePathFromReader(strings.NewReader(input))
	if result != "" {
		t.Errorf("expected empty string when no file_path key, got %q", result)
	}
}

func TestExtractFilePath_FilePathInContentBeforeKey(t *testing.T) {
	// The content field contains "file_path" as text before the actual key
	input := `{"tool": "Write", "content": "check the file_path value", "file_path": "/real/path.go"}`
	result := extractFilePathFromReader(strings.NewReader(input))
	// Should extract the real key, not the one inside content
	if result != "/real/path.go" {
		t.Errorf("expected /real/path.go, got %q", result)
	}
}

func TestExtractFilePath_MalformedJSON(t *testing.T) {
	input := `not json at all`
	result := extractFilePathFromReader(strings.NewReader(input))
	if result != "" {
		t.Errorf("expected empty string for malformed input, got %q", result)
	}
}

func TestExtractFilePath_FilePathWithSpaces(t *testing.T) {
	input := `{"file_path": "/home/user/my project/file.go"}`
	result := extractFilePathFromReader(strings.NewReader(input))
	if result != "/home/user/my project/file.go" {
		t.Errorf("expected path with spaces, got %q", result)
	}
}

func TestExtractFilePath_FilePathWithEscapedQuote(t *testing.T) {
	// Escaped quote in path — properly handles escape sequences
	input := `{"file_path": "path/with\"quote"}`
	result := extractFilePathFromReader(strings.NewReader(input))
	if result != `path/with"quote` {
		t.Errorf("expected 'path/with\"quote', got %q", result)
	}
}

func TestExtractFilePath_NullValue(t *testing.T) {
	input := `{"file_path": null}`
	result := extractFilePathFromReader(strings.NewReader(input))
	if result != "" {
		t.Errorf("expected empty string for null value, got %q", result)
	}
}

func TestExtractFilePath_WhitespaceAroundColon(t *testing.T) {
	input := `{"file_path"  :  "/path/to/file.go"}`
	result := extractFilePathFromReader(strings.NewReader(input))
	if result != "/path/to/file.go" {
		t.Errorf("expected /path/to/file.go, got %q", result)
	}
}

// --- core hook behavior tests (BDD scenarios) ---
//
// These test core.ClassifyFile and core.ValidateCommit directly.
// The CLI RunHooks command only exposes `install`; the pre-commit hook
// behavior runs via `ptsd validate` and is implemented in the core layer.

// setupHooksCoreProject creates a minimal project dir with a ptsd.yaml that
// declares test patterns, for use with core.ClassifyFile / ValidateCommit.
func setupHooksCoreProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatalf("mkdir .ptsd: %v", err)
	}
	configContent := "testing:\n  patterns:\n    files:\n      - \"tests/**/*.test.ts\"\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("write ptsd.yaml: %v", err)
	}
	return dir
}

// TestHooks_ClassifyFile_PRD verifies that .ptsd/docs/ paths are classified as PRD.
// BDD scenario: "File classification by path".
func TestHooks_ClassifyFile_PRD(t *testing.T) {
	dir := setupHooksCoreProject(t)
	class, err := core.ClassifyFile(dir, ".ptsd/docs/PRD.md")
	if err != nil {
		t.Fatalf("ClassifyFile error: %v", err)
	}
	if class != "PRD" {
		t.Errorf("expected PRD, got %s", class)
	}
}

// TestHooks_ClassifyFile_SEED verifies that .ptsd/seeds/ paths are classified as SEED.
func TestHooks_ClassifyFile_SEED(t *testing.T) {
	dir := setupHooksCoreProject(t)
	class, err := core.ClassifyFile(dir, ".ptsd/seeds/auth/seed.yaml")
	if err != nil {
		t.Fatalf("ClassifyFile error: %v", err)
	}
	if class != "SEED" {
		t.Errorf("expected SEED, got %s", class)
	}
}

// TestHooks_ClassifyFile_BDD verifies that .ptsd/bdd/ paths are classified as BDD.
func TestHooks_ClassifyFile_BDD(t *testing.T) {
	dir := setupHooksCoreProject(t)
	class, err := core.ClassifyFile(dir, ".ptsd/bdd/login.feature")
	if err != nil {
		t.Fatalf("ClassifyFile error: %v", err)
	}
	if class != "BDD" {
		t.Errorf("expected BDD, got %s", class)
	}
}

// TestHooks_ClassifyFile_TEST verifies that files matching the configured
// test pattern are classified as TEST.
func TestHooks_ClassifyFile_TEST(t *testing.T) {
	dir := setupHooksCoreProject(t)
	class, err := core.ClassifyFile(dir, "tests/auth/login.test.ts")
	if err != nil {
		t.Fatalf("ClassifyFile error: %v", err)
	}
	if class != "TEST" {
		t.Errorf("expected TEST for tests/**/*.test.ts pattern, got %s", class)
	}
}

// TestHooks_ClassifyFile_IMPL verifies that ordinary source files are classified as IMPL.
func TestHooks_ClassifyFile_IMPL(t *testing.T) {
	dir := setupHooksCoreProject(t)
	class, err := core.ClassifyFile(dir, "internal/core/auth.go")
	if err != nil {
		t.Fatalf("ClassifyFile error: %v", err)
	}
	if class != "IMPL" {
		t.Errorf("expected IMPL for source file, got %s", class)
	}
}

// TestHooks_ValidCommitScope covers BDD scenario "Valid commit with matching scope":
// staged files only in .ptsd/bdd/ with [BDD] commit message → hook passes.
func TestHooks_ValidCommitScope(t *testing.T) {
	dir := setupHooksCoreProject(t)
	stagedFiles := []string{".ptsd/bdd/login.feature"}
	if err := core.ValidateCommit(dir, "[BDD] add: login scenarios", stagedFiles); err != nil {
		t.Errorf("expected valid commit to pass, got: %v", err)
	}
}

// TestHooks_ScopeMismatch covers BDD scenario "Scope mismatch - impl files with BDD scope":
// staging both an IMPL file and a BDD file with [BDD] scope → hook fails with err:git.
func TestHooks_ScopeMismatch(t *testing.T) {
	dir := setupHooksCoreProject(t)
	// src/auth.ts classified as IMPL; .ptsd/bdd/auth.feature classified as BDD.
	// [BDD] scope mismatches the IMPL file → error.
	stagedFiles := []string{"src/auth.ts", ".ptsd/bdd/auth.feature"}
	err := core.ValidateCommit(dir, "[BDD] add: auth scenarios", stagedFiles)
	if err == nil {
		t.Error("expected scope mismatch to fail, got nil error")
	}
	if !strings.Contains(err.Error(), "err:git") {
		t.Errorf("expected err:git error, got: %v", err)
	}
}

// TestHooks_MissingScope covers BDD scenario "Missing scope":
// commit message without [SCOPE] → hook fails with "missing [SCOPE]" error.
func TestHooks_MissingScope(t *testing.T) {
	dir := setupHooksCoreProject(t)
	stagedFiles := []string{".ptsd/docs/PRD.md"}
	err := core.ValidateCommit(dir, "update PRD", stagedFiles)
	if err == nil {
		t.Error("expected missing scope to fail, got nil error")
	}
	if !strings.Contains(err.Error(), "err:git missing [SCOPE]") {
		t.Errorf("expected 'err:git missing [SCOPE]' error, got: %v", err)
	}
}

// TestHooks_TaskScopeSkipsPipelineValidation covers BDD scenario
// "TASK and STATUS scopes skip pipeline validation":
// [TASK] scope with .ptsd/tasks.yaml staged → passes without running full validate.
func TestHooks_TaskScopeSkipsPipelineValidation(t *testing.T) {
	dir := setupHooksCoreProject(t)
	stagedFiles := []string{".ptsd/tasks.yaml"}
	if err := core.ValidateCommit(dir, "[TASK] add: new auth task", stagedFiles); err != nil {
		t.Errorf("expected [TASK] scope to pass without pipeline validation, got: %v", err)
	}
}
