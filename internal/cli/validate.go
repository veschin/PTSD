package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/v2/internal/core"
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
		// Print advisories (non-blocking hints)
		advisories := core.CollectAdvisories(cwd)
		for _, a := range advisories {
			if agentMode {
				if a.Feature != "" {
					fmt.Printf("advisory: %s %s\n", a.Feature, a.Message)
				} else {
					fmt.Printf("advisory: %s\n", a.Message)
				}
			}
		}
		return 0
	}

	for _, ve := range errs {
		if agentMode {
			feature := ve.Feature
			if feature == "" {
				feature = "-"
			}
			fmt.Fprintf(os.Stderr, "err:%s %s: %s\n", ve.Category, feature, ve.Message)
		} else {
			feature := ve.Feature
			if feature == "" {
				feature = "(global)"
			}
			fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", ve.Category, feature, ve.Message)
		}
	}

	return 1
}
