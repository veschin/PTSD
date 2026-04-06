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
	Tool   string
}

// InitProject scaffolds .ptsd/ directory structure in the given directory.
// If .ptsd/ already exists, it performs a re-init (regenerates hooks, skills, CLAUDE.md section)
// without touching project data files.
// name is the project name written into ptsd.yaml; if empty, defaults to basename of dir.
// tool selects the AI tool adapter ("claude", "opencode", "generic"); if empty, auto-detected.
func InitProject(dir string, name string, tool string) (*InitResult, error) {
	// Require git repository.
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return nil, fmt.Errorf("err:config git repository required")
	}

	if tool == "" {
		tool = detectTool(dir)
	}

	// Auto-detect re-init.
	ptsdDir := filepath.Join(dir, ".ptsd")
	if _, err := os.Stat(ptsdDir); err == nil {
		if err := ReInitProjectWithTool(dir, tool); err != nil {
			return nil, err
		}
		return &InitResult{Reinit: true, Tool: tool}, nil
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
	ptsdYAML, err := renderTemplate("templates/ptsd.yaml.tmpl", struct{ Name, Runner, Tool string }{name, runner, tool})
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

	// Write .gitignore if it doesn't exist.
	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignore := "# Build artifacts\n*.exe\n*.dll\n*.so\n*.dylib\n\n# Binary output (match project name)\n/" + name + "\n"
		if err := writeFile(gitignorePath, gitignore); err != nil {
			return nil, err
		}
	}

	// Install git hooks (universal -- not tool-specific).
	if err := GeneratePreCommitHook(dir); err != nil {
		return nil, err
	}
	if err := GenerateCommitMsgHook(dir); err != nil {
		return nil, err
	}

	// Generate tool-specific adapter files.
	switch tool {
	case "claude":
		if err := generateClaudeSkills(dir); err != nil {
			return nil, err
		}
		if err := updateClaudeMDSection(dir); err != nil {
			return nil, err
		}
		if err := generateClaudeHooks(dir); err != nil {
			return nil, err
		}
	case "opencode":
		if err := generateOpenCodeCommands(dir); err != nil {
			return nil, err
		}
		if err := generateOpenCodePlugin(dir); err != nil {
			return nil, err
		}
		if err := updateAgentsMDSection(dir); err != nil {
			return nil, err
		}
	default:
		if err := updateAgentsMDSection(dir); err != nil {
			return nil, err
		}
	}

	return &InitResult{Reinit: false, Tool: tool}, nil
}

// detectTool auto-detects the AI tool based on existing config directories.
// Falls back to "claude" as the default.
func detectTool(dir string) string {
	if _, err := os.Stat(filepath.Join(dir, ".claude")); err == nil {
		return "claude"
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode")); err == nil {
		return "opencode"
	}
	return "claude"
}

const ptsdMarker = "<!-- ---ptsd--- -->"

// ReInitProject regenerates hooks, skills, and AI tool config without touching project data.
func ReInitProject(dir string) error {
	return ReInitProjectWithTool(dir, "")
}

// ReInitProjectWithTool regenerates with an explicit tool override.
// If tool is empty, auto-detects from filesystem/config.
func ReInitProjectWithTool(dir string, tool string) error {
	if err := GenerateAllSkills(dir); err != nil {
		return err
	}
	if err := GeneratePreCommitHook(dir); err != nil {
		return err
	}
	if err := GenerateCommitMsgHook(dir); err != nil {
		return err
	}

	if tool == "" {
		tool = detectTool(dir)
		cfg, err := LoadConfig(dir)
		if err == nil && cfg.Project.Tool != "" {
			tool = cfg.Project.Tool
		}
	}

	switch tool {
	case "claude":
		if err := generateClaudeSkills(dir); err != nil {
			return err
		}
		if err := generateClaudeHooks(dir); err != nil {
			return err
		}
		if err := updateClaudeMDSection(dir); err != nil {
			return err
		}
	case "opencode":
		if err := generateOpenCodeCommands(dir); err != nil {
			return err
		}
		if err := generateOpenCodePlugin(dir); err != nil {
			return err
		}
		if err := updateAgentsMDSection(dir); err != nil {
			return err
		}
	default:
		if err := updateAgentsMDSection(dir); err != nil {
			return err
		}
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
		// File doesn't exist -- create with markers.
		return writeFile(path, section+"\n")
	}

	content := string(existing)
	first := strings.Index(content, ptsdMarker)
	if first == -1 {
		// No markers -- append.
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + section + "\n"
		return writeFile(path, content)
	}

	second := strings.Index(content[first+len(ptsdMarker):], ptsdMarker)
	if second == -1 {
		// Only one marker (malformed) -- replace from first marker to end, append closing.
		content = content[:first] + section + "\n"
		return writeFile(path, content)
	}

	// Both markers found -- replace everything from first marker to end of second marker.
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

// generateOpenCodePlugin creates .opencode/plugins/ptsd.ts for gate and track hooks.
func generateOpenCodePlugin(dir string) error {
	pluginDir := filepath.Join(dir, ".opencode", "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	bin := ptsdBinaryPath()
	content, err := renderTemplate("templates/opencode/plugin.ts", struct{ Bin string }{bin})
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	return os.WriteFile(filepath.Join(pluginDir, "ptsd.ts"), []byte(content), 0644)
}

// generateOpenCodeCommands creates .opencode/commands/<name>.md for each standard skill.
func generateOpenCodeCommands(dir string) error {
	cmdDir := filepath.Join(dir, ".opencode", "commands")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}
	for _, filename := range standardSkillFiles {
		content, err := readTemplate("templates/skills/" + filename)
		if err != nil {
			return fmt.Errorf("err:io %w", err)
		}
		if err := os.WriteFile(filepath.Join(cmdDir, filename), []byte(content), 0644); err != nil {
			return fmt.Errorf("err:io %w", err)
		}
	}
	return nil
}

// updateAgentsMDSection writes or updates the ptsd-owned section in AGENTS.md using markers.
// Same logic as updateClaudeMDSection but targets AGENTS.md.
func updateAgentsMDSection(dir string) error {
	claudeMD, err := readTemplate("templates/claude.md")
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	section := ptsdMarker + "\n" + claudeMD + "\n" + ptsdMarker

	path := filepath.Join(dir, "AGENTS.md")
	existing, err := os.ReadFile(path)
	if err != nil {
		return writeFile(path, section+"\n")
	}

	content := string(existing)
	first := strings.Index(content, ptsdMarker)
	if first == -1 {
		if len(content) > 0 && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + section + "\n"
		return writeFile(path, content)
	}

	second := strings.Index(content[first+len(ptsdMarker):], ptsdMarker)
	if second == -1 {
		content = content[:first] + section + "\n"
		return writeFile(path, content)
	}

	afterSecond := first + len(ptsdMarker) + second + len(ptsdMarker)
	content = content[:first] + section + content[afterSecond:]
	return writeFile(path, content)
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

	// Check for Rust.
	if _, err := os.Stat(filepath.Join(dir, "Cargo.toml")); err == nil {
		return "cargo test"
	}

	// Check for Ruby.
	if _, err := os.Stat(filepath.Join(dir, "Gemfile")); err == nil {
		return "bundle exec rspec"
	}

	// Check for Java Maven.
	if _, err := os.Stat(filepath.Join(dir, "pom.xml")); err == nil {
		return "mvn test"
	}

	// Check for Java Gradle.
	if _, err := os.Stat(filepath.Join(dir, "build.gradle")); err == nil {
		return "gradle test"
	}

	return ""
}

