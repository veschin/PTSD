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
// Avoids encoding/json to handle potentially huge content fields efficiently.
func extractFilePathFromStdin() string {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}

	input := string(data)

	// Look for "file_path" key in JSON
	key := `"file_path"`
	idx := strings.Index(input, key)
	if idx == -1 {
		return ""
	}

	// Find the colon after the key
	rest := input[idx+len(key):]
	rest = strings.TrimSpace(rest)
	if len(rest) == 0 || rest[0] != ':' {
		return ""
	}
	rest = strings.TrimSpace(rest[1:])

	if len(rest) == 0 || rest[0] != '"' {
		return ""
	}

	// Extract string value
	rest = rest[1:]
	end := strings.Index(rest, `"`)
	if end == -1 {
		return ""
	}

	return rest[:end]
}
