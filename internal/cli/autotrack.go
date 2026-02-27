package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/core"
)

func RunAutoTrack(args []string, agentMode bool) int {
	filePath := ""
	for i, arg := range args {
		if arg == "--file" && i+1 < len(args) {
			filePath = args[i+1]
		}
	}
	if filePath == "" {
		fmt.Fprintln(os.Stderr, "err:user usage: ptsd auto-track --file <path>")
		return 2
	}

	dir, err := os.Getwd()
	if err != nil {
		return coreError(agentMode, err)
	}

	result, err := core.AutoTrack(dir, filePath)
	if err != nil {
		return coreError(agentMode, err)
	}

	if result == nil {
		if agentMode {
			fmt.Println("ok no-op")
		}
		return 0
	}

	if result.Updated {
		if agentMode {
			fmt.Printf("tracked: %s stage=%s tests=%s\n", result.Feature, result.Stage, result.Tests)
		} else {
			fmt.Printf("Updated %s: stage=%s tests=%s\n", result.Feature, result.Stage, result.Tests)
		}
	} else {
		if agentMode {
			fmt.Println("ok no-op")
		}
	}

	return 0
}
