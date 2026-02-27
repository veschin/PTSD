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

// initProject is a test helper that calls InitProject and fails on error.
func initProject(t *testing.T, dir, name string) *InitResult {
	t.Helper()
	result, err := InitProject(dir, name)
	if err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}
	return result
}

// TestInitCreatesAllExpectedDirsAndFiles covers BDD: "Init new project".
func TestInitCreatesAllExpectedDirsAndFiles(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	initProject(t, dir, "MyApp")

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
		".gitignore",
		".git/hooks/pre-commit",
		".claude/skills/write-prd/SKILL.md",
		".claude/skills/workflow/SKILL.md",
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

	initProject(t, dir, "MyApp")

	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "ptsd.yaml"))
	if err != nil {
		t.Fatalf("failed to read ptsd.yaml: %v", err)
	}

	if !strings.Contains(string(data), "MyApp") {
		t.Errorf("ptsd.yaml does not contain project name 'MyApp':\n%s", data)
	}
}

// TestInitReInitSucceeds covers BDD: "Re-init existing project regenerates hooks and skills".
func TestInitReInitSucceeds(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	// First init.
	result := initProject(t, dir, "MyApp")
	if result.Reinit {
		t.Error("first init should not be reinit")
	}

	// Second init should succeed as re-init.
	result2, err := InitProject(dir, "MyApp")
	if err != nil {
		t.Fatalf("re-init failed: %v", err)
	}
	if !result2.Reinit {
		t.Error("second init should report Reinit=true")
	}
}

// TestInitRefusesWithoutGit covers BDD: "Init refuses without git".
func TestInitRefusesWithoutGit(t *testing.T) {
	dir := t.TempDir()
	// No .git directory.

	_, err := InitProject(dir, "MyApp")
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

	initProject(t, dir, "MyApp")

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

	initProject(t, dir, "MyApp")

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

	initProject(t, dir, "MyApp")

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

	initProject(t, dir, "MyApp")

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

	initProject(t, dir, "")

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

	initProject(t, dir, "MyApp")

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

	initProject(t, dir, "MyApp")

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

	initProject(t, dir, "MyApp")

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

	initProject(t, dir, "MyApp")

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

// TestInitFullInitCreatesClaudeMDWithMarkers verifies full init wraps CLAUDE.md content in markers.
func TestInitFullInitCreatesClaudeMDWithMarkers(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	initProject(t, dir, "MyApp")

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("CLAUDE.md not found: %v", err)
	}

	content := string(data)
	marker := "<!-- ---ptsd--- -->"
	count := strings.Count(content, marker)
	if count != 2 {
		t.Errorf("expected 2 markers in CLAUDE.md, got %d", count)
	}
	if !strings.Contains(content, "# Claude Agent Instructions") {
		t.Error("CLAUDE.md missing template content")
	}
}

// TestReInitRegeneratesHooks verifies git + claude hooks are refreshed on re-init.
func TestReInitRegeneratesHooks(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	initProject(t, dir, "MyApp")

	// Corrupt a hook file.
	hookPath := filepath.Join(dir, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(hookPath, []byte("corrupted"), 0644); err != nil {
		t.Fatalf("failed to corrupt hook: %v", err)
	}

	// Re-init.
	result := initProject(t, dir, "MyApp")
	if !result.Reinit {
		t.Error("expected Reinit=true")
	}

	// Hook should be regenerated.
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("pre-commit hook not found after re-init: %v", err)
	}
	if !strings.Contains(string(data), "validate") {
		t.Errorf("pre-commit hook not regenerated, content: %s", data)
	}

	// Claude hooks should exist.
	for _, script := range []string{"ptsd-context.sh", "ptsd-gate.sh", "ptsd-track.sh"} {
		path := filepath.Join(dir, ".claude", "hooks", script)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected %s to exist after re-init: %v", script, err)
		}
	}

	// Settings should exist.
	if _, err := os.Stat(filepath.Join(dir, ".claude", "settings.json")); err != nil {
		t.Error("expected .claude/settings.json after re-init")
	}

	// Claude skills should be regenerated.
	claudeSkillPath := filepath.Join(dir, ".claude", "skills", "write-prd", "SKILL.md")
	if _, err := os.Stat(claudeSkillPath); err != nil {
		t.Errorf("expected .claude/skills/write-prd/SKILL.md after re-init: %v", err)
	}
}

// TestReInitRegeneratesSkills verifies skills are overwritten on re-init.
func TestReInitRegeneratesSkills(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	initProject(t, dir, "MyApp")

	// Corrupt a skill file.
	skillPath := filepath.Join(dir, ".ptsd", "skills", "workflow.md")
	if err := os.WriteFile(skillPath, []byte("corrupted"), 0644); err != nil {
		t.Fatalf("failed to corrupt skill: %v", err)
	}

	// Re-init.
	initProject(t, dir, "MyApp")

	// Skill should be regenerated.
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("skill not found after re-init: %v", err)
	}
	if string(data) == "corrupted" {
		t.Error("skill was not regenerated on re-init")
	}
}

// TestReInitPreservesDataFiles verifies project data files are not touched by re-init.
func TestReInitPreservesDataFiles(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	initProject(t, dir, "MyApp")

	// Write custom content to data files.
	customFeatures := "features:\n  - id: my-feature\n    name: My Feature\n    status: active\n"
	customTasks := "tasks:\n  - id: T-1\n    title: Custom task\n"
	customState := "features:\n  my-feature:\n    hash: abc123\n"
	customReview := "features:\n  my-feature:\n    stage: impl\n"

	dataFiles := map[string]string{
		"features.yaml":      customFeatures,
		"tasks.yaml":         customTasks,
		"state.yaml":         customState,
		"review-status.yaml": customReview,
	}

	for name, content := range dataFiles {
		path := filepath.Join(dir, ".ptsd", name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	// Also write custom PRD.
	customPRD := "# My Custom PRD\nThis is important."
	if err := os.WriteFile(filepath.Join(dir, ".ptsd", "docs", "PRD.md"), []byte(customPRD), 0644); err != nil {
		t.Fatalf("failed to write PRD: %v", err)
	}

	// Re-init.
	initProject(t, dir, "MyApp")

	// Verify all data files are unchanged.
	for name, expected := range dataFiles {
		path := filepath.Join(dir, ".ptsd", name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", name, err)
		}
		if string(data) != expected {
			t.Errorf("%s was modified by re-init.\nexpected: %q\ngot: %q", name, expected, string(data))
		}
	}

	// PRD should be unchanged.
	prdData, err := os.ReadFile(filepath.Join(dir, ".ptsd", "docs", "PRD.md"))
	if err != nil {
		t.Fatalf("failed to read PRD: %v", err)
	}
	if string(prdData) != customPRD {
		t.Errorf("PRD was modified by re-init.\nexpected: %q\ngot: %q", customPRD, string(prdData))
	}
}

// TestReInitUpdatesClaudeMDSection verifies content between markers is replaced.
func TestReInitUpdatesClaudeMDSection(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	initProject(t, dir, "MyApp")

	// Read the generated CLAUDE.md.
	path := filepath.Join(dir, "CLAUDE.md")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("CLAUDE.md not found: %v", err)
	}

	// Corrupt the ptsd section but keep markers.
	marker := "<!-- ---ptsd--- -->"
	corrupted := marker + "\ncorrupted content\n" + marker + "\n"
	if err := os.WriteFile(path, []byte(corrupted), 0644); err != nil {
		t.Fatalf("failed to corrupt CLAUDE.md: %v", err)
	}

	// Re-init.
	initProject(t, dir, "MyApp")

	// Section should be restored.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("CLAUDE.md not found after re-init: %v", err)
	}
	if string(data) != string(original) {
		t.Errorf("CLAUDE.md not properly restored.\nexpected: %q\ngot: %q", original, data)
	}
}

// TestReInitAppendsClaudeMDSection verifies markers are appended if absent.
func TestReInitAppendsClaudeMDSection(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	initProject(t, dir, "MyApp")

	// Replace CLAUDE.md with content that has no markers.
	path := filepath.Join(dir, "CLAUDE.md")
	userContent := "# My Project\n\nCustom instructions here.\n"
	if err := os.WriteFile(path, []byte(userContent), 0644); err != nil {
		t.Fatalf("failed to write CLAUDE.md: %v", err)
	}

	// Re-init.
	initProject(t, dir, "MyApp")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("CLAUDE.md not found after re-init: %v", err)
	}

	content := string(data)
	marker := "<!-- ---ptsd--- -->"

	// User content should be preserved at the top.
	if !strings.HasPrefix(content, userContent) {
		t.Errorf("user content not preserved at top of CLAUDE.md.\ngot: %q", content)
	}

	// Markers should be present.
	if strings.Count(content, marker) != 2 {
		t.Errorf("expected 2 markers, got %d", strings.Count(content, marker))
	}

	// Template content should be present.
	if !strings.Contains(content, "# Claude Agent Instructions") {
		t.Error("CLAUDE.md missing template content after re-init")
	}
}

// TestReInitCreatesClaudeMDWithMarkers verifies CLAUDE.md is created if missing.
func TestReInitCreatesClaudeMDWithMarkers(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	initProject(t, dir, "MyApp")

	// Delete CLAUDE.md.
	path := filepath.Join(dir, "CLAUDE.md")
	if err := os.Remove(path); err != nil {
		t.Fatalf("failed to remove CLAUDE.md: %v", err)
	}

	// Re-init.
	initProject(t, dir, "MyApp")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("CLAUDE.md not found after re-init: %v", err)
	}

	content := string(data)
	marker := "<!-- ---ptsd--- -->"

	if strings.Count(content, marker) != 2 {
		t.Errorf("expected 2 markers, got %d", strings.Count(content, marker))
	}
	if !strings.Contains(content, "# Claude Agent Instructions") {
		t.Error("CLAUDE.md missing template content")
	}
}

// TestReInitPreservesUserContentInClaudeMD verifies user content above/below markers is kept.
func TestReInitPreservesUserContentInClaudeMD(t *testing.T) {
	dir := t.TempDir()
	setupGitDir(t, dir)

	initProject(t, dir, "MyApp")

	// Add user content above and below markers.
	path := filepath.Join(dir, "CLAUDE.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("CLAUDE.md not found: %v", err)
	}

	above := "# My Custom Header\n\nUser notes go here.\n\n"
	below := "\n# User Footer\n\nMore custom stuff.\n"
	newContent := above + string(data) + below
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		t.Fatalf("failed to write CLAUDE.md: %v", err)
	}

	// Re-init.
	initProject(t, dir, "MyApp")

	result, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("CLAUDE.md not found after re-init: %v", err)
	}

	content := string(result)

	// User content above should be preserved.
	if !strings.HasPrefix(content, above) {
		t.Errorf("user content above markers not preserved.\ngot prefix: %q", content[:min(len(content), len(above)+20)])
	}

	// User content below should be preserved.
	if !strings.HasSuffix(content, below) {
		t.Errorf("user content below markers not preserved.\ngot suffix: %q", content[max(0, len(content)-len(below)-20):])
	}

	// Template content should be present.
	if !strings.Contains(content, "# Claude Agent Instructions") {
		t.Error("CLAUDE.md missing template content after re-init")
	}
}
