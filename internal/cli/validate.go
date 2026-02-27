package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/core"
)

// RunValidate executes `ptsd validate`. Returns an exit code.
// Exit 0 = clean, 1 = validation errors present.
func RunValidate(args []string, agentMode bool) int {
	cwd, err := os.Getwd()
	if err != nil {
		return renderError(agentMode, "io", err.Error())
	}

	errs, err := core.Validate(cwd)
	if err != nil {
		return coreError(agentMode, err)
	}

	if len(errs) == 0 {
		if !agentMode {
			fmt.Println("ok")
		}
		return 0
	}

	for _, ve := range errs {
		if agentMode {
			feature := ve.Feature
			if feature == "" {
				feature = "-"
			}
			fmt.Printf("err:%s %s: %s\n", ve.Category, feature, ve.Message)
		} else {
			feature := ve.Feature
			if feature == "" {
				feature = "(global)"
			}
			fmt.Printf("[%s] %s: %s\n", ve.Category, feature, ve.Message)
		}
	}

	return 1
}
