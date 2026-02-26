package core

import (
	"fmt"
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

func ValidateCommit(projectDir string, message string, stagedFiles []string) error {
	scope, _, _, err := ParseCommitMessage(message)
	if err != nil {
		return err
	}

	if !validScopes[scope] {
		return fmt.Errorf("err:git unknown scope %s", scope)
	}

	if scope == "TASK" || scope == "STATUS" {
		return nil
	}

	hasImplFile := false
	for _, file := range stagedFiles {
		class, _ := ClassifyFile(projectDir, file)
		if class == "IMPL" {
			hasImplFile = true
			break
		}
	}

	if hasImplFile && scope != "IMPL" {
		return fmt.Errorf("err:git staged files require [IMPL] scope")
	}

	return nil
}

func ClassifyFile(projectDir string, path string) (string, error) {
	switch {
	case strings.HasPrefix(path, ".ptsd/docs/"):
		return "PRD", nil
	case strings.HasPrefix(path, ".ptsd/seeds/"):
		return "SEED", nil
	case strings.HasPrefix(path, ".ptsd/bdd/"):
		return "BDD", nil
	}

	// Try to load config for test patterns
	cfg, err := LoadConfig(projectDir)
	if err == nil && len(cfg.Testing.Patterns.Files) > 0 {
		for _, pattern := range cfg.Testing.Patterns.Files {
			if matchesTestPattern(path, pattern) {
				return "TEST", nil
			}
		}
	}

	// Fallback: common test file patterns when config has no patterns or fails to load
	if strings.HasPrefix(path, "tests/") && strings.HasSuffix(path, ".test.ts") {
		return "TEST", nil
	}
	if strings.HasSuffix(path, "_test.go") {
		return "TEST", nil
	}
	if strings.HasSuffix(path, ".test.ts") {
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
			// Remove leading "*" from suffix and any following "/"
			suffix := strings.TrimPrefix(suffixRaw, "*")
			suffix = strings.TrimPrefix(suffix, "/")

			// Path must start with prefix/ and end with suffix
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
