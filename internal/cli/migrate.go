package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/core"
)

// RunMigrate handles `ptsd migrate`.
func RunMigrate(args []string, agentMode bool) int {
	cwd, err := os.Getwd()
	if err != nil {
		return renderError(agentMode, "io", err.Error())
	}

	result, err := core.MigrateProject(cwd)
	if err != nil {
		return coreError(agentMode, err)
	}

	if agentMode {
		fmt.Printf("migrate:ok features:%d config:%v hooks:%v\n",
			result.FeaturesUpdated, result.ConfigUpdated, result.HooksRegenerated)
	} else {
		fmt.Println("Migration complete:")
		if result.FeaturesUpdated > 0 {
			fmt.Printf("  Features updated with pipeline field: %d\n", result.FeaturesUpdated)
		}
		if result.ConfigUpdated {
			fmt.Println("  Config updated with pipeline section")
		}
		if result.HooksRegenerated {
			fmt.Println("  Hooks, skills, CLAUDE.md regenerated")
		}
	}

	return 0
}
