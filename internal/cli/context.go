package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/core"
)

func RunContext(args []string, agentMode bool) int {
	dir, err := os.Getwd()
	if err != nil {
		return coreError(agentMode, err)
	}

	result, err := core.BuildContext(dir)
	if err != nil {
		return coreError(agentMode, err)
	}

	for _, line := range result.Lines {
		switch line.Type {
		case core.ContextNext:
			fmt.Printf("next: %s stage=%s action=%s\n", line.Feature, line.Stage, line.Action)
		case core.ContextBlocked:
			fmt.Printf("blocked: %s stage=%s reason=%q\n", line.Feature, line.Stage, line.Reason)
		case core.ContextDone:
			fmt.Printf("done: %s stage=%s\n", line.Feature, line.Stage)
		case core.ContextTask:
			fmt.Printf("task: %s status=%s feature=%s title=%q\n", line.TaskID, line.TaskStatus, line.Feature, line.TaskTitle)
		}
	}

	return 0
}
