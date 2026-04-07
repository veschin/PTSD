package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/v2/internal/core"
)

// RunMap handles `ptsd map`.
func RunMap(args []string, agentMode bool) int {
	cwd, err := os.Getwd()
	if err != nil {
		return renderError(agentMode, "io", err.Error())
	}

	// Check .ptsd/ exists
	ptsdDir := cwd + "/.ptsd"
	if _, err := os.Stat(ptsdDir); os.IsNotExist(err) {
		return renderError(agentMode, "validation", "not a ptsd project (run ptsd init first)")
	}

	result, err := core.MapProject(cwd)
	if err != nil {
		return coreError(agentMode, err)
	}

	if agentMode {
		fmt.Printf("map:ok lang:%s sources:%d tests:%d dirs:%d\n",
			result.Language, result.SourceCount, result.TestCount, len(result.Dirs))
	} else {
		fmt.Printf("Generated codebase map: .ptsd/docs/CODEBASE.md\n")
		fmt.Printf("Language: %s | Sources: %d | Tests: %d\n",
			result.Language, result.SourceCount, result.TestCount)
	}
	return 0
}
