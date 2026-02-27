package main

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/cli"
)

func main() {
	agentMode := false
	var filteredArgs []string

	for _, arg := range os.Args[1:] {
		if arg == "--agent" || arg == "-agent" {
			agentMode = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if len(filteredArgs) == 0 {
		fmt.Fprintln(os.Stderr, "usage: ptsd <command> [options]\nRun 'ptsd help' for available commands.")
		os.Exit(2)
	}

	cmd := filteredArgs[0]
	if cmd == "--help" || cmd == "-h" {
		cmd = "help"
	}
	subargs := filteredArgs[1:]

	var exitCode int
	switch cmd {
	case "init":
		exitCode = cli.RunInit(subargs, agentMode)
	case "adopt":
		exitCode = cli.RunAdopt(subargs, agentMode)
	case "feature":
		exitCode = cli.RunFeature(subargs, agentMode)
	case "config":
		exitCode = cli.RunConfig(subargs, agentMode)
	case "task":
		exitCode = cli.RunTask(subargs, agentMode)
	case "prd":
		exitCode = cli.RunPrd(subargs, agentMode)
	case "seed":
		exitCode = cli.RunSeed(subargs, agentMode)
	case "bdd":
		exitCode = cli.RunBdd(subargs, agentMode)
	case "test":
		exitCode = cli.RunTest(subargs, agentMode)
	case "status":
		exitCode = cli.RunStatus(subargs, agentMode)
	case "validate":
		exitCode = cli.RunValidate(subargs, agentMode)
	case "hooks":
		exitCode = cli.RunHooks(subargs, agentMode)
	case "review":
		exitCode = cli.RunReview(subargs, agentMode)
	case "skills":
		exitCode = cli.RunSkills(subargs, agentMode)
	case "issues":
		exitCode = cli.RunIssues(subargs, agentMode)
	case "context":
		exitCode = cli.RunContext(subargs, agentMode)
	case "gate-check":
		exitCode = cli.RunGateCheck(subargs, agentMode)
	case "auto-track":
		exitCode = cli.RunAutoTrack(subargs, agentMode)
	case "help":
		exitCode = cli.RunHelp(subargs, agentMode)
	default:
		fmt.Fprintf(os.Stderr, "err:user unknown command: %s\n", cmd)
		exitCode = 2
	}

	os.Exit(exitCode)
}
