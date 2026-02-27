package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/core"
)

func RunGateCheck(args []string, agentMode bool) int {
	filePath := ""
	for i, arg := range args {
		if arg == "--file" && i+1 < len(args) {
			filePath = args[i+1]
		}
	}
	if filePath == "" {
		fmt.Fprintln(os.Stderr, "err:user usage: ptsd gate-check --file <path>")
		return 2
	}

	dir, err := os.Getwd()
	if err != nil {
		return coreError(agentMode, err)
	}

	result := core.GateCheck(dir, filePath)
	if result.Allowed {
		if agentMode {
			fmt.Println("ok")
		} else {
			fmt.Println("Gate check passed")
		}
		return 0
	}

	fmt.Fprintln(os.Stderr, result.Reason)
	return 2
}
