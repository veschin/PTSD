package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var validScopes = map[string]bool{
	"PRD":    true,
	"SEED":   true,
	"BDD":    true,
	"TEST":   true,
	"IMPL":   true,
	"TASK":   true,
	"STATUS": true,
}

var validCommitTypes = map[string]bool{
	"feat":     true,
	"add":      true,
	"fix":      true,
	"refactor": true,
	"remove":   true,
	"update":   true,
}

func ptsdBinaryPath() string {
	if exe, err := os.Executable(); err == nil {
		return exe
	}
	return "ptsd"
}

func GeneratePreCommitHook(projectDir string) error {
	hookDir := filepath.Join(projectDir, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	bin := ptsdBinaryPath()
	hookContent := "#!/bin/sh\n" + bin + " validate\n"
	hookPath := filepath.Join(hookDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	return nil
}

func GenerateCommitMsgHook(projectDir string) error {
	hookDir := filepath.Join(projectDir, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	bin := ptsdBinaryPath()
	hookContent := "#!/bin/sh\n" + bin + " hooks validate-commit --msg-file \"$1\"\n"
	hookPath := filepath.Join(hookDir, "commit-msg")
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	return nil
}

func ValidateCommitFromFile(projectDir, msgFile string) error {
	data, err := os.ReadFile(msgFile)
	if err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	msg := strings.TrimSpace(string(data))
	if msg == "" {
		return fmt.Errorf("err:git empty commit message")
	}

	// Parse and validate format only (no staged file classification — git handles that in pre-commit)
	scope, commitType, _, err := ParseCommitMessage(msg)
	if err != nil {
		return err
	}

	if !validScopes[scope] {
		return fmt.Errorf("err:git unknown scope %s", scope)
	}

	if commitType != "" && !validCommitTypes[commitType] {
		return fmt.Errorf("err:git invalid commit type %q: must be feat|add|fix|refactor|remove|update", commitType)
	}

	return nil
}

func ValidateCommit(projectDir string, message string, stagedFiles []string) error {
	scope, commitType, _, err := ParseCommitMessage(message)
	if err != nil {
		return err
	}

	if !validScopes[scope] {
		return fmt.Errorf("err:git unknown scope %s", scope)
	}

	if commitType != "" && !validCommitTypes[commitType] {
		return fmt.Errorf("err:git invalid commit type %q: must be feat|add|fix|refactor|remove|update", commitType)
	}

	if scope == "TASK" || scope == "STATUS" {
		return nil
	}

	// Classify all staged files and check bidirectional scope matching
	for _, file := range stagedFiles {
		class, _ := ClassifyFile(projectDir, file)
		if class != scope {
			return fmt.Errorf("err:git file %s classified as %s but scope is [%s]", file, class, scope)
		}
	}

	// IMPL is the final pipeline stage — trigger full validation
	if scope == "IMPL" {
		validationErrors, err := Validate(projectDir)
		if err != nil {
			return fmt.Errorf("err:pipeline %w", err)
		}
		if len(validationErrors) > 0 {
			msgs := make([]string, len(validationErrors))
			for i, ve := range validationErrors {
				msgs[i] = ve.Feature + ": " + ve.Message
			}
			return fmt.Errorf("err:pipeline validation failed: %s", strings.Join(msgs, "; "))
		}
	}

	return nil
}

func ClassifyFile(projectDir string, path string) (string, error) {
	// .ptsd/ internal files
	if strings.HasPrefix(path, ".ptsd/") {
		switch {
		case strings.HasPrefix(path, ".ptsd/docs/"):
			return "PRD", nil
		case strings.HasPrefix(path, ".ptsd/seeds/"):
			return "SEED", nil
		case strings.HasPrefix(path, ".ptsd/bdd/"):
			return "BDD", nil
		case path == ".ptsd/tasks.yaml":
			return "TASK", nil
		case path == ".ptsd/state.yaml" || path == ".ptsd/review-status.yaml":
			return "STATUS", nil
		case path == ".ptsd/features.yaml" || path == ".ptsd/ptsd.yaml" || path == ".ptsd/issues.yaml":
			return "STATUS", nil
		case strings.HasPrefix(path, ".ptsd/skills/"):
			return "STATUS", nil
		}
	}

	// Try config-based test patterns
	cfg, err := LoadConfig(projectDir)
	if err == nil && len(cfg.Testing.Patterns.Files) > 0 {
		for _, pattern := range cfg.Testing.Patterns.Files {
			if matchesTestPattern(path, pattern) {
				return "TEST", nil
			}
		}
	}

	// Fallback: common test file patterns
	if strings.HasSuffix(path, "_test.go") {
		return "TEST", nil
	}
	if strings.HasSuffix(path, ".test.ts") || strings.HasSuffix(path, ".test.js") {
		return "TEST", nil
	}

	return "IMPL", nil
}

func matchesTestPattern(path, pattern string) bool {
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffixRaw := strings.TrimPrefix(parts[1], "/")
			suffix := strings.TrimPrefix(suffixRaw, "*")
			suffix = strings.TrimPrefix(suffix, "/")

			if prefix == "" {
				return strings.HasSuffix(path, suffix)
			}
			return strings.HasPrefix(path, prefix+"/") && strings.HasSuffix(path, suffix)
		}
	}

	matched, _ := filepath.Match(pattern, path)
	return matched
}

func ParseCommitMessage(msg string) (scope string, commitType string, text string, err error) {
	if !strings.HasPrefix(msg, "[") {
		return "", "", "", fmt.Errorf("err:git missing [SCOPE] in commit message")
	}

	endIdx := strings.Index(msg, "]")
	if endIdx == -1 {
		return "", "", "", fmt.Errorf("err:git missing [SCOPE] in commit message")
	}

	scope = msg[1:endIdx]

	rest := strings.TrimSpace(msg[endIdx+1:])
	colonIdx := strings.Index(rest, ":")
	if colonIdx == -1 {
		return scope, "", rest, nil
	}

	commitType = strings.TrimSpace(rest[:colonIdx])
	text = strings.TrimSpace(rest[colonIdx+1:])

	return scope, commitType, text, nil
}
