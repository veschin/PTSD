package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupGitDir creates a minimal git repository in dir (just the .git directory).
func setupGitDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
}

// TestInitCreatesAllExpectedDirsAndFiles covers BDD: "Init new project".
func TestInitCreatesAllExpectedDirsAndFiles(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	err := InitProject(dir, "MyApp")
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// Subdirectories.
	for _, sub := range []string{"docs", "seeds", "bdd", "skills"} {
		path := filepath.Join(dir, ".ptsd", sub)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf(".ptsd/%s does not exist: %v", sub, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf(".ptsd/%s is not a directory", sub)
		}
	}

	// Required files.
	requiredFiles := []string{
		".ptsd/ptsd.yaml",
		".ptsd/features.yaml",
		".ptsd/state.yaml",
		".ptsd/tasks.yaml",
		".ptsd/review-status.yaml",
		".ptsd/docs/PRD.md",
		"CLAUDE.md",
		".git/hooks/pre-commit",
	}
	for _, rel := range requiredFiles {
		path := filepath.Join(dir, rel)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist: %v", rel, err)
		}
	}
}

// TestInitPtsdYAMLContainsProjectName verifies ptsd.yaml stores the project name.
func TestInitPtsdYAMLContainsProjectName(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	if err := InitProject(dir, "MyApp"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "ptsd.yaml"))
	if err != nil {
		t.Fatalf("failed to read ptsd.yaml: %v", err)
	}

	if !strings.Contains(string(data), "MyApp") {
		t.Errorf("ptsd.yaml does not contain project name 'MyApp':\n%s", data)
	}
}

// TestInitRefusesIfPtsdAlreadyExists covers BDD: "Init refuses if .ptsd already exists".
func TestInitRefusesIfPtsdAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	// First init.
	if err := InitProject(dir, "MyApp"); err != nil {
		t.Fatalf("first InitProject failed: %v", err)
	}

	// Second init must fail.
	err := InitProject(dir, "MyApp")
	if err == nil {
		t.Fatal("expected error on re-init, got nil")
	}
	if !strings.Contains(err.Error(), "err:validation") {
		t.Errorf("expected err:validation error, got: %q", err.Error())
	}
	if !strings.Contains(err.Error(), ".ptsd already exists") {
		t.Errorf("expected '.ptsd already exists' in error, got: %q", err.Error())
	}
}

// TestInitRefusesWithoutGit covers BDD: "Init refuses without git".
func TestInitRefusesWithoutGit(t *testing.T) {
	dir := t.TempDir()
	// No .git directory.

	err := InitProject(dir, "MyApp")
	if err == nil {
		t.Fatal("expected error when .git is missing, got nil")
	}
	if !strings.Contains(err.Error(), "err:config") {
		t.Errorf("expected err:config error, got: %q", err.Error())
	}
	if !strings.Contains(err.Error(), "git repository required") {
		t.Errorf("expected 'git repository required' in error, got: %q", err.Error())
	}
}

// TestInitCreatedFilesHaveValidContent verifies the created files have parseable YAML content.
func TestInitCreatedFilesHaveValidContent(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	if err := InitProject(dir, "MyApp"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	// features.yaml must be loadable.
	features, err := loadFeatures(dir)
	if err != nil {
		t.Errorf("features.yaml not parseable: %v", err)
	}
	if len(features) != 0 {
		t.Errorf("expected 0 features, got %d", len(features))
	}

	// ptsd.yaml must be loadable via LoadConfig.
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Errorf("ptsd.yaml not parseable: %v", err)
	}
	if cfg.Project.Name != "MyApp" {
		t.Errorf("config project name = %q, want 'MyApp'", cfg.Project.Name)
	}
}

// TestInitGeneratesSkills covers BDD: "Init generates skills".
func TestInitGeneratesSkills(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	if err := InitProject(dir, "MyApp"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	expectedSkills := []string{
		"write-prd.md",
		"write-seed.md",
		"write-bdd.md",
		"write-tests.md",
		"write-impl.md",
		"review-prd.md",
		"review-seed.md",
		"review-bdd.md",
		"review-tests.md",
		"review-impl.md",
		"create-tasks.md",
		"workflow.md",
	}

	for _, skill := range expectedSkills {
		path := filepath.Join(dir, ".ptsd", "skills", skill)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected skill file %s to exist: %v", skill, err)
		}
	}
}

// TestInitDetectsVitestRunner covers BDD: "Init detects test runner".
func TestInitDetectsVitestRunner(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	// Create package.json referencing vitest.
	pkgJSON := `{"devDependencies": {"vitest": "^1.0.0"}}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	if err := InitProject(dir, "MyApp"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Testing.Runner != "npx vitest run" {
		t.Errorf("expected runner 'npx vitest run', got %q", cfg.Testing.Runner)
	}
}

// TestInitInstallsPreCommitHook verifies the git hook is executable and has correct content.
func TestInitInstallsPreCommitHook(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	if err := InitProject(dir, "MyApp"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("pre-commit hook not found: %v", err)
	}

	if !strings.Contains(string(data), "validate") {
		t.Errorf("pre-commit hook does not contain 'validate':\n%s", data)
	}

	// Check it is executable.
	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("stat hook: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("pre-commit hook is not executable, mode=%v", info.Mode())
	}
}

// TestInitDefaultNameFromDir verifies that when name is empty, basename of dir is used.
func TestInitDefaultNameFromDir(t *testing.T) {
	// Create a directory with a known name.
	parent := t.TempDir()
	dir := filepath.Join(parent, "myproject")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	setupGitDir(t, dir)

	if err := InitProject(dir, ""); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Project.Name != "myproject" {
		t.Errorf("expected project name 'myproject', got %q", cfg.Project.Name)
	}
}

// TestInitGeneratesClaudeHookScripts verifies .claude/hooks/ scripts are created.
func TestInitGeneratesClaudeHookScripts(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	if err := InitProject(dir, "MyApp"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	scripts := []string{
		".claude/hooks/ptsd-context.sh",
		".claude/hooks/ptsd-gate.sh",
		".claude/hooks/ptsd-track.sh",
	}

	for _, script := range scripts {
		path := filepath.Join(dir, script)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected %s to exist: %v", script, err)
			continue
		}
		if info.Mode()&0111 == 0 {
			t.Errorf("expected %s to be executable, mode=%v", script, info.Mode())
		}
	}
}

// TestInitGeneratesClaudeSettings verifies .claude/settings.json is created with hook wiring.
func TestInitGeneratesClaudeSettings(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	if err := InitProject(dir, "MyApp"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("expected .claude/settings.json to exist: %v", err)
	}

	content := string(data)

	// Check hook events are wired
	if !strings.Contains(content, "SessionStart") {
		t.Error("settings.json missing SessionStart hook")
	}
	if !strings.Contains(content, "UserPromptSubmit") {
		t.Error("settings.json missing UserPromptSubmit hook")
	}
	if !strings.Contains(content, "PreToolUse") {
		t.Error("settings.json missing PreToolUse hook")
	}
	if !strings.Contains(content, "PostToolUse") {
		t.Error("settings.json missing PostToolUse hook")
	}

	// Check matchers
	if !strings.Contains(content, `"matcher": "Edit|Write"`) {
		t.Error("settings.json missing Edit|Write matcher for PreToolUse")
	}

	// Check hook scripts are referenced
	if !strings.Contains(content, "ptsd-context.sh") {
		t.Error("settings.json missing ptsd-context.sh reference")
	}
	if !strings.Contains(content, "ptsd-gate.sh") {
		t.Error("settings.json missing ptsd-gate.sh reference")
	}
	if !strings.Contains(content, "ptsd-track.sh") {
		t.Error("settings.json missing ptsd-track.sh reference")
	}
}

// TestInitHookScriptsContainPtsdBinary verifies hook scripts call ptsd with correct subcommands.
func TestInitHookScriptsContainPtsdBinary(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	if err := InitProject(dir, "MyApp"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	cases := []struct {
		script  string
		contain string
	}{
		{"ptsd-context.sh", "context --agent"},
		{"ptsd-gate.sh", "hooks pre-tool-use"},
		{"ptsd-track.sh", "hooks post-tool-use"},
	}

	for _, tc := range cases {
		path := filepath.Join(dir, ".claude", "hooks", tc.script)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", tc.script, err)
		}
		if !strings.Contains(string(data), tc.contain) {
			t.Errorf("%s should contain %q, got: %s", tc.script, tc.contain, data)
		}
	}
}

// TestInitGeneratesCommitMsgHook verifies .git/hooks/commit-msg is created.
func TestInitGeneratesCommitMsgHook(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	if err := InitProject(dir, "MyApp"); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}

	hookPath := filepath.Join(dir, ".git", "hooks", "commit-msg")
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("commit-msg hook not found: %v", err)
	}

	if !strings.Contains(string(data), "validate-commit") {
		t.Errorf("commit-msg hook should contain 'validate-commit', got: %s", data)
	}

	info, err := os.Stat(hookPath)
	if err != nil {
		t.Fatalf("stat hook: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("commit-msg hook is not executable, mode=%v", info.Mode())
	}
}
