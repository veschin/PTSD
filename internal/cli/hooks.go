package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/core"
)

// RunHooks executes `ptsd hooks <subcommand>`. Returns an exit code.
//
// Supported subcommands:
//
//	install   — write .git/hooks/pre-commit
func RunHooks(args []string, agentMode bool) int {
	if len(args) == 0 {
		if agentMode {
			fmt.Fprintf(os.Stderr, "err:user hooks requires a subcommand: install\n")
		} else {
			fmt.Fprintln(os.Stderr, "usage: ptsd hooks install")
		}
		return 2
	}

	subcmd := args[0]

	switch subcmd {
	case "install":
		return runHooksInstall(agentMode)
	default:
		if agentMode {
			fmt.Fprintf(os.Stderr, "err:user unknown hooks subcommand: %s\n", subcmd)
		} else {
			fmt.Fprintf(os.Stderr, "unknown subcommand %q: expected install\n", subcmd)
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

	if agentMode {
		fmt.Println("ok hooks installed")
	} else {
		fmt.Println("pre-commit hook installed at .git/hooks/pre-commit")
	}

	return 0
}
