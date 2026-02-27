package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/veschin/ptsd/internal/core"
)

// RunHooks executes `ptsd hooks <subcommand>`. Returns an exit code.
//
// Supported subcommands:
//
//	install         — write .git/hooks/pre-commit + commit-msg
//	validate-commit — validate commit message format from file
//	pre-tool-use    — gate-check for Claude Code PreToolUse hook
//	post-tool-use   — auto-track for Claude Code PostToolUse hook
func RunHooks(args []string, agentMode bool) int {
	if len(args) == 0 {
		if agentMode {
			fmt.Fprintf(os.Stderr, "err:user hooks requires a subcommand: install|validate-commit|pre-tool-use|post-tool-use\n")
		} else {
			fmt.Fprintln(os.Stderr, "usage: ptsd hooks <install|validate-commit|pre-tool-use|post-tool-use>")
		}
		return 2
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "install":
		return runHooksInstall(agentMode)
	case "validate-commit":
		return runValidateCommit(subargs, agentMode)
	case "pre-tool-use":
		return runPreToolUse(agentMode)
	case "post-tool-use":
		return runPostToolUse(agentMode)
	default:
		if agentMode {
			fmt.Fprintf(os.Stderr, "err:user unknown hooks subcommand: %s\n", subcmd)
		} else {
			fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", subcmd)
		}
		return 2
	}
}

func runHooksInstall(agentMode bool) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err:io %s\n", err)
		return 4
	}

	if err := core.GeneratePreCommitHook(cwd); err != nil {
		return coreError(agentMode, err)
	}

	if err := core.GenerateCommitMsgHook(cwd); err != nil {
		return coreError(agentMode, err)
	}

	if agentMode {
		fmt.Println("ok hooks installed")
	} else {
		fmt.Println("Git hooks installed at .git/hooks/")
	}

	return 0
}

func runValidateCommit(args []string, agentMode bool) int {
	msgFile := ""
	for i, arg := range args {
		if arg == "--msg-file" && i+1 < len(args) {
			msgFile = args[i+1]
		}
	}
	if msgFile == "" {
		fmt.Fprintln(os.Stderr, "err:user usage: ptsd hooks validate-commit --msg-file <path>")
		return 2
	}

	cwd, err := os.Getwd()
	if err != nil {
		return coreError(agentMode, err)
	}

	if err := core.ValidateCommitFromFile(cwd, msgFile); err != nil {
		return coreError(agentMode, err)
	}

	if agentMode {
		fmt.Println("ok")
	}
	return 0
}

// runPreToolUse reads Claude Code hook JSON from stdin, extracts file_path, runs gate-check.
// Exit 0 = allow, exit 2 = block.
func runPreToolUse(agentMode bool) int {
	filePath := extractFilePathFromStdin()
	if filePath == "" {
		return 0 // No file_path → not a file write → allow
	}

	cwd, err := os.Getwd()
	if err != nil {
		return 0
	}

	result := core.GateCheck(cwd, filePath)
	if result.Allowed {
		return 0
	}

	fmt.Fprintln(os.Stderr, result.Reason)
	return 2
}

// runPostToolUse reads Claude Code hook JSON from stdin, extracts file_path, runs auto-track.
func runPostToolUse(agentMode bool) int {
	filePath := extractFilePathFromStdin()
	if filePath == "" {
		return 0
	}

	cwd, err := os.Getwd()
	if err != nil {
		return 0
	}

	result, err := core.AutoTrack(cwd, filePath)
	if err != nil {
		return 0 // Don't block on tracking errors
	}

	if result != nil && result.Updated {
		fmt.Printf("tracked: %s stage=%s tests=%s\n", result.Feature, result.Stage, result.Tests)
	}

	return 0
}

// extractFilePathFromStdin scans stdin JSON for "file_path":"..." using simple string search.
func extractFilePathFromStdin() string {
	return extractFilePathFromReader(os.Stdin)
}

// extractFilePathFromReader scans JSON from a reader for "file_path":"..." using string search.
// Avoids encoding/json to handle potentially huge content fields efficiently.
// Looks for "file_path" as a JSON key (preceded by { or ,) to avoid matching inside content values.
func extractFilePathFromReader(r io.Reader) string {
	data, err := io.ReadAll(r)
	if err != nil {
		return ""
	}

	input := string(data)
	key := `"file_path"`

	// Search for all occurrences and pick the one that looks like a JSON key
	offset := 0
	for {
		idx := strings.Index(input[offset:], key)
		if idx == -1 {
			return ""
		}
		pos := offset + idx

		// Check if this looks like a JSON key: preceded by { or , (ignoring whitespace)
		isKey := false
		for i := pos - 1; i >= 0; i-- {
			c := input[i]
			if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				continue
			}
			if c == '{' || c == ',' {
				isKey = true
			}
			break
		}

		if isKey {
			return extractJSONStringValue(input[pos+len(key):])
		}

		offset = pos + len(key)
	}
}

// extractJSONStringValue extracts a string value from `: "value"` pattern.
func extractJSONStringValue(rest string) string {
	rest = strings.TrimSpace(rest)
	if len(rest) == 0 || rest[0] != ':' {
		return ""
	}
	rest = strings.TrimSpace(rest[1:])

	if len(rest) == 0 || rest[0] != '"' {
		return ""
	}

	// Extract string value, handling escaped quotes
	rest = rest[1:]
	var result strings.Builder
	for i := 0; i < len(rest); i++ {
		if rest[i] == '\\' && i+1 < len(rest) {
			result.WriteByte(rest[i+1])
			i++
			continue
		}
		if rest[i] == '"' {
			return result.String()
		}
		result.WriteByte(rest[i])
	}
	return ""
}
