package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/veschin/ptsd/internal/core"
)

// RunInit handles `ptsd init [name]`.
func RunInit(args []string, agentMode bool) int {
	cwd, err := os.Getwd()
	if err != nil {
		return renderError(agentMode, "io", err.Error())
	}

	name := ""
	for i, arg := range args {
		if arg == "--name" && i+1 < len(args) {
			name = args[i+1]
			break
		}
		if !strings.HasPrefix(arg, "-") && name == "" {
			name = arg
		}
	}

	result, err := core.InitProject(cwd, name)
	if err != nil {
		return coreError(agentMode, err)
	}

	if result.Reinit {
		if agentMode {
			fmt.Printf("reinit:ok hooks:5 skills:12\n")
		} else {
			fmt.Printf("Re-initialized ptsd project in %s\n", cwd)
		}
	} else {
		if agentMode {
			fmt.Printf("init:ok dir:%s\n", cwd)
		} else {
			fmt.Printf("Initialized ptsd project in %s\n", cwd)
		}
	}
	return 0
}

// RunAdopt handles `ptsd adopt [--dry-run]`.
func RunAdopt(args []string, agentMode bool) int {
	dryRun := false
	for _, a := range args {
		if a == "--dry-run" {
			dryRun = true
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return renderError(agentMode, "io", err.Error())
	}

	if dryRun {
		result, err := core.AdoptDryRun(cwd)
		if err != nil {
			return coreError(agentMode, err)
		}
		if agentMode {
			fmt.Printf("dry-run:ok bdd:%d tests:%d features:%s\n",
				len(result.BDDFiles), len(result.TestFiles), result.FeaturesFile)
		} else {
			fmt.Printf("Dry run — would create: %s\n", result.FeaturesFile)
			fmt.Printf("BDD features found: %d\n", len(result.BDDFiles))
			fmt.Printf("Test files found: %d\n", len(result.TestFiles))
		}
		return 0
	}

	if err := core.AdoptProject(cwd); err != nil {
		return coreError(agentMode, err)
	}

	if agentMode {
		fmt.Printf("adopt:ok dir:%s\n", cwd)
	} else {
		fmt.Printf("Adopted project in %s\n", cwd)
	}
	return 0
}
