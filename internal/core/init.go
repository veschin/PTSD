package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InitResult reports what InitProject did.
type InitResult struct {
	Reinit bool
}

// InitProject scaffolds .ptsd/ directory structure in the given directory.
// If .ptsd/ already exists, it performs a re-init (regenerates hooks, skills, CLAUDE.md section)
// without touching project data files.
// name is the project name written into ptsd.yaml; if empty, defaults to basename of dir.
func InitProject(dir string, name string) (*InitResult, error) {
	// Require git repository.
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return nil, fmt.Errorf("err:config git repository required")
	}

	// Auto-detect re-init.
	ptsdDir := filepath.Join(dir, ".ptsd")
	if _, err := os.Stat(ptsdDir); err == nil {
		if err := ReInitProject(dir); err != nil {
			return nil, err
		}
		return &InitResult{Reinit: true}, nil
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
			return nil, fmt.Errorf("err:io %w", err)
		}
	}

	// Detect test runner from project layout.
	runner := detectTestRunner(dir)

	// Write ptsd.yaml.
	ptsdYAML, err := renderTemplate("templates/ptsd.yaml.tmpl", struct{ Name, Runner string }{name, runner})
	if err != nil {
		return nil, fmt.Errorf("err:io %w", err)
	}
	if err := writeFile(filepath.Join(ptsdDir, "ptsd.yaml"), ptsdYAML); err != nil {
		return nil, err
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
			return nil, err
		}
	}

	// Write PRD template.
	prdContent, err := renderTemplate("templates/prd.md.tmpl", struct{ Name string }{name})
	if err != nil {
		return nil, fmt.Errorf("err:io %w", err)
	}
	if err := writeFile(filepath.Join(ptsdDir, "docs", "PRD.md"), prdContent); err != nil {
		return nil, err
	}

	// Write skills.
	if err := GenerateAllSkills(dir); err != nil {
		return nil, err
	}

	// Generate Claude Code skill discovery files.
	if err := generateClaudeSkills(dir); err != nil {
		return nil, err
	}

	// Write CLAUDE.md at project root (with markers for future re-init).
	if err := updateClaudeMDSection(dir); err != nil {
		return nil, err
	}

	// Write .gitignore if it doesn't exist.
	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignore := "# Build artifacts\n*.exe\n*.dll\n*.so\n*.dylib\n\n# Binary output (match project name)\n/" + name + "\n"
		if err := writeFile(gitignorePath, gitignore); err != nil {
			return nil, err
		}
	}

	// Install git hooks.
	if err := GeneratePreCommitHook(dir); err != nil {
		return nil, err
	}
	if err := GenerateCommitMsgHook(dir); err != nil {
		return nil, err
	}

	// Generate Claude Code hooks.
	if err := generateClaudeHooks(dir); err != nil {
		return nil, err
	}

	return &InitResult{Reinit: false}, nil
}

const ptsdMarker = "<!-- ---ptsd--- -->"

// ReInitProject regenerates hooks, skills, and CLAUDE.md section without touching project data.
func ReInitProject(dir string) error {
	if err := GenerateAllSkills(dir); err != nil {
		return err
	}
	if err := generateClaudeSkills(dir); err != nil {
		return err
	}
	if err := GeneratePreCommitHook(dir); err != nil {
		return err
	}
	if err := GenerateCommitMsgHook(dir); err != nil {
		return err
	}
	if err := generateClaudeHooks(dir); err != nil {
		return err
	}
	if err := updateClaudeMDSection(dir); err != nil {
		return err
	}
	return nil
}

// updateClaudeMDSection writes or updates the ptsd-owned section in CLAUDE.md using markers.
func updateClaudeMDSection(dir string) error {
	claudeMD, err := readTemplate("templates/claude.md")
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	section := ptsdMarker + "\n" + claudeMD + "\n" + ptsdMarker

	path := filepath.Join(dir, "CLAUDE.md")
	existing, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist — create with markers.
		return writeFile(path, section+"\n")
	}

	content := string(existing)
	first := strings.Index(content, ptsdMarker)
	if first == -1 {
		// No markers — append.
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + section + "\n"
		return writeFile(path, content)
	}

	second := strings.Index(content[first+len(ptsdMarker):], ptsdMarker)
	if second == -1 {
		// Only one marker (malformed) — replace from first marker to end, append closing.
		content = content[:first] + section + "\n"
		return writeFile(path, content)
	}

	// Both markers found — replace everything from first marker to end of second marker.
	afterSecond := first + len(ptsdMarker) + second + len(ptsdMarker)
	content = content[:first] + section + content[afterSecond:]
	return writeFile(path, content)
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

