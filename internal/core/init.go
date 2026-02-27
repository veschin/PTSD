package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InitProject scaffolds .ptsd/ directory structure in the given directory.
// It requires a git repository (presence of .git/) and refuses if .ptsd/ already exists.
// name is the project name written into ptsd.yaml; if empty, defaults to basename of dir.
func InitProject(dir string, name string) error {
	// Require git repository.
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf("err:config git repository required")
	}

	// Refuse re-init.
	ptsdDir := filepath.Join(dir, ".ptsd")
	if _, err := os.Stat(ptsdDir); err == nil {
		return fmt.Errorf("err:validation .ptsd already exists")
	}

	if name == "" {
		name = filepath.Base(dir)
	}

	// Create directory structure.
	dirs := []string{
		ptsdDir,
		filepath.Join(ptsdDir, "docs"),
		filepath.Join(ptsdDir, "seeds"),
		filepath.Join(ptsdDir, "bdd"),
		filepath.Join(ptsdDir, "skills"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("err:io %w", err)
		}
	}

	// Detect test runner from project layout.
	runner := detectTestRunner(dir)

	// Write ptsd.yaml.
	ptsdYAML, err := renderTemplate("templates/ptsd.yaml.tmpl", struct{ Name, Runner string }{name, runner})
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}
	if err := writeFile(filepath.Join(ptsdDir, "ptsd.yaml"), ptsdYAML); err != nil {
		return err
	}

	// Write empty registry files.
	emptyFiles := map[string]string{
		"features.yaml":      "features: []\n",
		"state.yaml":         "features: {}\n",
		"tasks.yaml":         "tasks: []\n",
		"review-status.yaml": "features: {}\n",
	}
	for filename, content := range emptyFiles {
		if err := writeFile(filepath.Join(ptsdDir, filename), content); err != nil {
			return err
		}
	}

	// Write PRD template.
	prdContent, err := renderTemplate("templates/prd.md.tmpl", struct{ Name string }{name})
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}
	if err := writeFile(filepath.Join(ptsdDir, "docs", "PRD.md"), prdContent); err != nil {
		return err
	}

	// Write skills.
	if err := GenerateAllSkills(dir); err != nil {
		return err
	}

	// Write CLAUDE.md at project root.
	claudeMD, err := readTemplate("templates/claude.md")
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}
	if err := writeFile(filepath.Join(dir, "CLAUDE.md"), claudeMD); err != nil {
		return err
	}

	// Install git hooks.
	if err := GeneratePreCommitHook(dir); err != nil {
		return err
	}
	if err := GenerateCommitMsgHook(dir); err != nil {
		return err
	}

	// Generate Claude Code hooks.
	if err := generateClaudeHooks(dir); err != nil {
		return err
	}

	return nil
}

func generateClaudeHooks(dir string) error {
	bin := ptsdBinaryPath()

	// Create .claude/hooks/ directory
	hooksDir := filepath.Join(dir, ".claude", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	binData := struct{ Bin string }{bin}

	// Generate hook scripts from templates
	hookFiles := []struct {
		tmpl string
		dest string
	}{
		{"templates/hooks/context.sh", "ptsd-context.sh"},
		{"templates/hooks/gate.sh", "ptsd-gate.sh"},
		{"templates/hooks/track.sh", "ptsd-track.sh"},
	}

	for _, hf := range hookFiles {
		content, err := renderTemplate(hf.tmpl, binData)
		if err != nil {
			return fmt.Errorf("err:io %w", err)
		}
		if err := os.WriteFile(filepath.Join(hooksDir, hf.dest), []byte(content), 0755); err != nil {
			return fmt.Errorf("err:io %w", err)
		}
	}

	// Generate .claude/settings.json from template
	settingsData := struct{ ContextHook, GateHook, TrackHook string }{
		ContextHook: filepath.Join(hooksDir, "ptsd-context.sh"),
		GateHook:    filepath.Join(hooksDir, "ptsd-gate.sh"),
		TrackHook:   filepath.Join(hooksDir, "ptsd-track.sh"),
	}

	settingsJSON, err := renderTemplate("templates/settings.json.tmpl", settingsData)
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	if err := os.WriteFile(settingsPath, []byte(settingsJSON), 0644); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	return nil
}

func writeFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("err:io %w", err)
	}
	return nil
}

// detectTestRunner inspects the project directory for known package managers/frameworks.
func detectTestRunner(dir string) string {
	// Check package.json for vitest or jest.
	pkgJSON := filepath.Join(dir, "package.json")
	if data, err := os.ReadFile(pkgJSON); err == nil {
		content := string(data)
		if strings.Contains(content, "vitest") {
			return "npx vitest run"
		}
		if strings.Contains(content, "jest") {
			return "npx jest"
		}
	}

	// Check for Go module.
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return "go test ./..."
	}

	// Check for pytest.
	if _, err := os.Stat(filepath.Join(dir, "pytest.ini")); err == nil {
		return "pytest"
	}
	if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
		return "pytest"
	}

	return ""
}

