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
	if err := writeFile(filepath.Join(ptsdDir, "ptsd.yaml"), buildPtsdYAML(name, runner)); err != nil {
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
	if err := writeFile(filepath.Join(ptsdDir, "docs", "PRD.md"), buildPRDTemplate(name)); err != nil {
		return err
	}

	// Write skills.
	skills := buildSkills()
	for filename, content := range skills {
		if err := writeFile(filepath.Join(ptsdDir, "skills", filename), content); err != nil {
			return err
		}
	}

	// Write CLAUDE.md at project root.
	if err := writeFile(filepath.Join(dir, "CLAUDE.md"), buildClaudeMD()); err != nil {
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

	// ptsd-context.sh
	contextScript := "#!/bin/sh\n" + bin + " context --agent 2>/dev/null\nexit 0\n"
	if err := os.WriteFile(filepath.Join(hooksDir, "ptsd-context.sh"), []byte(contextScript), 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	// ptsd-gate.sh
	gateScript := "#!/bin/sh\n" + bin + " hooks pre-tool-use --agent\n"
	if err := os.WriteFile(filepath.Join(hooksDir, "ptsd-gate.sh"), []byte(gateScript), 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	// ptsd-track.sh
	trackScript := "#!/bin/sh\n" + bin + " hooks post-tool-use --agent 2>/dev/null\nexit 0\n"
	if err := os.WriteFile(filepath.Join(hooksDir, "ptsd-track.sh"), []byte(trackScript), 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	// .claude/settings.json
	contextHook := hooksDir + "/ptsd-context.sh"
	gateHook := hooksDir + "/ptsd-gate.sh"
	trackHook := hooksDir + "/ptsd-track.sh"

	settingsJSON := `{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "` + contextHook + `"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "` + contextHook + `"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "` + gateHook + `"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "` + trackHook + `"
          }
        ]
      }
    ]
  }
}
`
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

func buildPtsdYAML(name, runner string) string {
	var b strings.Builder
	b.WriteString("project:\n")
	b.WriteString("  name: \"" + name + "\"\n")
	b.WriteString("\n")
	b.WriteString("testing:\n")
	if runner != "" {
		b.WriteString("  runner: \"" + runner + "\"\n")
	}
	b.WriteString("  patterns:\n")
	b.WriteString("    files: [\"**/*_test.go\"]\n")
	b.WriteString("\n")
	b.WriteString("review:\n")
	b.WriteString("  min_score: 7\n")
	b.WriteString("  auto_redo: true\n")
	b.WriteString("\n")
	b.WriteString("hooks:\n")
	b.WriteString("  pre_commit: true\n")
	b.WriteString("  scopes: [PRD, SEED, BDD, TEST, IMPL, TASK, STATUS]\n")
	b.WriteString("  types: [feat, add, fix, refactor, remove, update]\n")
	return b.String()
}

func buildPRDTemplate(name string) string {
	return "# " + name + " — Product Requirements Document\n\n" +
		"## Overview\n\n" +
		"[Describe the project purpose and goals here.]\n\n" +
		"## Features\n\n" +
		"[Add feature sections below. Use <!-- feature:id --> anchors.]\n"
}

func buildClaudeMD() string {
	return `# Claude Agent Instructions

## Authority Hierarchy (ENFORCED BY HOOKS)

PTSD (iron law) > User (context provider) > Assistant (executor)

- PTSD decides what CAN and CANNOT be done. Pipeline, gates, validation — non-negotiable.
  Hooks enforce this automatically — writes that violate pipeline are BLOCKED.
- User provides context and requirements. User also follows ptsd rules.
- Assistant executes within ptsd constraints. Writes code, docs, tests on behalf of user.

## Session Start Protocol

EVERY session, BEFORE any work:
1. Run: ptsd context --agent — see full pipeline state
2. Run: ptsd task next --agent — get next task
3. Follow output exactly.

## Commands (always use --agent flag)

- ptsd context --agent              — full pipeline state (auto-injected by hooks)
- ptsd status --agent               — project overview
- ptsd task next --agent            — next task to work on
- ptsd task update <id> --status WIP — mark task in progress
- ptsd validate --agent             — check pipeline before commit
- ptsd feature list --agent         — list all features
- ptsd seed init <id> --agent       — initialize seed directory
- ptsd gate-check --file <path> --agent — check if file write is allowed

## Pipeline (strict order, no skipping)

PRD → Seed → BDD → Tests → Implementation

Each stage requires review score ≥ 7 before advancing.
Hooks enforce gates automatically — blocked writes show the reason.

## Rules

- NO mocks for internal code. Real tests, real files, temp directories.
- NO garbage files. Every file must link to a feature.
- NO hiding errors. Explain WHY something failed.
- NO over-engineering. Minimum code for the current task.
- ALWAYS run: ptsd validate --agent before committing.
- COMMIT FORMAT: [SCOPE] type: message
  Scopes: PRD, SEED, BDD, TEST, IMPL, TASK, STATUS
  Types: feat, add, fix, refactor, remove, update

## Forbidden

- Mocking internal code
- Skipping pipeline steps
- Hiding errors or pretending something works
- Generating files not linked to a feature
- Using --force, --skip-validation, --no-verify
`
}

func buildSkills() map[string]string {
	return map[string]string{
		"write-prd.md":    buildSkillWritePRD(),
		"write-seed.md":   buildSkillWriteSeed(),
		"write-bdd.md":    buildSkillWriteBDD(),
		"write-tests.md":  buildSkillWriteTests(),
		"write-impl.md":   buildSkillWriteImpl(),
		"review-prd.md":   buildSkillReviewPRD(),
		"review-seed.md":  buildSkillReviewSeed(),
		"review-bdd.md":   buildSkillReviewBDD(),
		"review-tests.md": buildSkillReviewTests(),
		"review-impl.md":  buildSkillReviewImpl(),
		"create-tasks.md": buildSkillCreateTasks(),
		"workflow.md":     buildSkillWorkflow(),
	}
}

func buildSkillWritePRD() string {
	return `---
name: write-prd
description: Guide for writing PRD sections for a feature
trigger: When adding a feature section to .ptsd/docs/PRD.md
---

1. Add <!-- feature:id --> anchor before the section.
2. Describe purpose, behavior, edge cases, non-goals.
3. Include acceptance criteria (bullet list).
4. Do not over-specify implementation details.
`
}

func buildSkillWriteSeed() string {
	return `---
name: write-seed
description: Guide for creating golden seed data for a feature
trigger: When running ptsd seed init <feature-id>
---

1. Create realistic data — not placeholder "foo/bar" values.
2. Include happy path and edge cases.
3. Write a seed.yaml manifest listing all seed files.
4. Seed data must be parseable by the project's own code.
`
}

func buildSkillWriteBDD() string {
	return `---
name: write-bdd
description: Guide for writing Gherkin BDD scenarios from PRD and seed data
trigger: When creating .feature files for a feature
---

1. One Scenario per acceptance criterion.
2. Reference seed data in Given steps (use exact field values).
3. Cover happy path, error cases, edge cases.
4. Use concrete values, not abstract descriptions.
`
}

func buildSkillWriteTests() string {
	return `---
name: write-tests
description: Guide for writing tests from BDD scenarios
trigger: When implementing tests for a feature
---

1. One test function per BDD Scenario.
2. No mocks for internal code — use real files, real temp dirs.
3. Assert exact outcomes, not just "no error".
4. Test function name must reference the Scenario title.
`
}

func buildSkillWriteImpl() string {
	return `---
name: write-impl
description: Guide for implementing code to make tests pass
trigger: When writing implementation for a feature
---

1. Make failing tests pass — nothing more.
2. No extra abstractions, no premature optimization.
3. Handle every error path tested in BDD.
4. Run tests after every change: go test ./...
`
}

func buildSkillReviewPRD() string {
	return `---
name: review-prd
description: Review a PRD section for quality (score 0-10)
trigger: When reviewing PRD artifacts for a feature
---

Check:
- Feature anchor present?
- Acceptance criteria listed?
- Edge cases covered?
- Non-goals stated?
- No implementation details leaked?

Output: score (0-10), issues list (explicit), verdict (pass/redo).
`
}

func buildSkillReviewSeed() string {
	return `---
name: review-seed
description: Review seed data for quality (score 0-10)
trigger: When reviewing seed artifacts for a feature
---

Check:
- Data realistic (not placeholder)?
- Happy path covered?
- Edge cases covered?
- Manifest lists all files?
- Data parseable?

Output: score (0-10), issues list, verdict (pass/redo).
`
}

func buildSkillReviewBDD() string {
	return `---
name: review-bdd
description: Review BDD scenarios for quality (score 0-10)
trigger: When reviewing .feature files for a feature
---

Check:
- All acceptance criteria covered?
- Scenarios reference seed data?
- Error cases included?
- No abstract/vague steps?

Output: score (0-10), issues list, verdict (pass/redo).
`
}

func buildSkillReviewTests() string {
	return `---
name: review-tests
description: Review tests for quality (score 0-10)
trigger: When reviewing test files for a feature
---

Check:
- 1:1 mapping to BDD scenarios?
- No mocks for internal code?
- Assertions are concrete?
- All error paths tested?

Output: score (0-10), issues list, verdict (pass/redo).
`
}

func buildSkillReviewImpl() string {
	return `---
name: review-impl
description: Review implementation for quality (score 0-10)
trigger: When reviewing implementation code for a feature
---

Check:
- All tests pass?
- No extra unrelated code?
- Every error path handled?
- No hardcoded values?

Output: score (0-10), issues list, verdict (pass/redo).
`
}

func buildSkillCreateTasks() string {
	return `---
name: create-tasks
description: Guide for creating well-structured tasks
trigger: When adding tasks with ptsd task add
---

1. Link every task to a feature (--feature flag).
2. Set priority: A=blocking, B=important, C=future.
3. Title must describe the concrete output, not the activity.
4. One task = one deliverable.
`
}

func buildSkillWorkflow() string {
	return `---
name: workflow
description: Full PTSD pipeline workflow reference
trigger: At session start or when unsure what to do next
---

## Mandatory Session Start

1. ptsd context --agent   — see full pipeline state (auto-injected by hooks)
2. ptsd task next --agent — get next task
3. Read linked PRD section, BDD scenarios, seed data
4. Do the work (hooks auto-track progress)
5. ptsd validate --agent  — before every commit
6. Commit: [SCOPE] type: message

## Pipeline Order (per feature)

PRD → Seed → BDD → Tests → Implementation

- ptsd prd check         — validate PRD anchors
- ptsd seed init <id>    — create seed directory
- ptsd bdd add <id>      — create .feature file (requires seed)
- ptsd test map <f> <t>  — map BDD to test file (requires BDD)
- ptsd review <id> --stage <stage>  — record review score

## Gate Rules (enforced by hooks)

- No BDD without seed (PreToolUse blocks write)
- No test mapping without BDD
- No impl without tests
- No implemented status without all tests passing
- Progress auto-tracked via PostToolUse hook
`
}
