package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/veschin/ptsd/internal/core"
)

func RunConfig(args []string, agentMode bool) int {
	if len(args) == 0 {
		return usageError(agentMode, "config", "subcommand required: show")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return renderError(agentMode, "io", err.Error())
	}

	sub := args[0]

	switch sub {
	case "show":
		cfg, err := core.LoadConfig(cwd)
		if err != nil {
			return coreError(agentMode, err)
		}
		printConfig(agentMode, cfg)
		return 0

	default:
		return usageError(agentMode, "config", fmt.Sprintf("unknown subcommand %q: use show", sub))
	}
}

func printConfig(agentMode bool, cfg *core.Config) {
	if agentMode {
		fmt.Printf("project.name=%s\n", cfg.Project.Name)
		fmt.Printf("testing.runner=%s\n", cfg.Testing.Runner)
		fmt.Printf("testing.patterns.files=%s\n", strings.Join(cfg.Testing.Patterns.Files, ","))
		fmt.Printf("testing.result_parser.format=%s\n", cfg.Testing.ResultParser.Format)
		fmt.Printf("review.min_score=%d\n", cfg.Review.MinScore)
		fmt.Printf("review.auto_redo=%v\n", cfg.Review.AutoRedo)
		fmt.Printf("hooks.pre_commit=%v\n", cfg.Hooks.PreCommit)
		fmt.Printf("hooks.scopes=%s\n", strings.Join(cfg.Hooks.Scopes, ","))
		fmt.Printf("hooks.types=%s\n", strings.Join(cfg.Hooks.Types, ","))
	} else {
		fmt.Printf("project:\n")
		fmt.Printf("  name: %s\n", cfg.Project.Name)
		fmt.Printf("testing:\n")
		fmt.Printf("  runner: %s\n", cfg.Testing.Runner)
		fmt.Printf("  patterns.files: %s\n", strings.Join(cfg.Testing.Patterns.Files, ", "))
		fmt.Printf("  result_parser.format: %s\n", cfg.Testing.ResultParser.Format)
		fmt.Printf("review:\n")
		fmt.Printf("  min_score: %d\n", cfg.Review.MinScore)
		fmt.Printf("  auto_redo: %v\n", cfg.Review.AutoRedo)
		fmt.Printf("hooks:\n")
		fmt.Printf("  pre_commit: %v\n", cfg.Hooks.PreCommit)
		fmt.Printf("  scopes: %s\n", strings.Join(cfg.Hooks.Scopes, ", "))
		fmt.Printf("  types: %s\n", strings.Join(cfg.Hooks.Types, ", "))
	}
}
