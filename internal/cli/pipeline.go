package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/veschin/ptsd/internal/core"
	"github.com/veschin/ptsd/internal/render"
)

// RunPrd handles: ptsd prd check
func RunPrd(args []string, agentMode bool) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "err:user usage: ptsd prd check")
		return 2
	}
	switch args[0] {
	case "check":
		dir, err := os.Getwd()
		if err != nil {
			return coreError(agentMode, err)
		}
		errs, err := core.CheckPRDAnchors(dir)
		if err != nil {
			return coreError(agentMode, err)
		}
		if len(errs) == 0 {
			if agentMode {
				fmt.Println("ok")
			} else {
				fmt.Println("PRD anchors OK")
			}
			return 0
		}
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "err:pipeline %s %s\n", e.Type, e.FeatureID)
		}
		return 1
	default:
		fmt.Fprintf(os.Stderr, "err:user unknown prd subcommand: %s\n", args[0])
		return 2
	}
}

// RunSeed handles: ptsd seed add <feature> <file> [type] [description]
func RunSeed(args []string, agentMode bool) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "err:user usage: ptsd seed add <feature> <file> [type] [description]")
		return 2
	}
	switch args[0] {
	case "add":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "err:user usage: ptsd seed add <feature> <file> [type] [description]")
			return 2
		}
		featureID := args[1]
		filePath := args[2]
		fileType := "data"
		description := ""
		if len(args) >= 4 {
			fileType = args[3]
		}
		if len(args) >= 5 {
			description = strings.Join(args[4:], " ")
		}
		dir, err := os.Getwd()
		if err != nil {
			return coreError(agentMode, err)
		}
		err = core.AddSeedFile(dir, featureID, filePath, fileType, description)
		if err != nil {
			return coreError(agentMode, err)
		}
		if agentMode {
			fmt.Printf("seed added: %s -> %s\n", filePath, featureID)
		} else {
			fmt.Printf("Added seed file %s to feature %s\n", filePath, featureID)
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "err:user unknown seed subcommand: %s\n", args[0])
		return 2
	}
}

// RunBdd handles: ptsd bdd add <feature> | ptsd bdd list [feature]
func RunBdd(args []string, agentMode bool) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "err:user usage: ptsd bdd <add|list> ...")
		return 2
	}
	switch args[0] {
	case "add":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "err:user usage: ptsd bdd add <feature>")
			return 2
		}
		featureID := args[1]
		dir, err := os.Getwd()
		if err != nil {
			return coreError(agentMode, err)
		}
		err = core.AddBDD(dir, featureID)
		if err != nil {
			return coreError(agentMode, err)
		}
		if agentMode {
			fmt.Printf("bdd added: %s\n", featureID)
		} else {
			fmt.Printf("BDD scaffold created for feature %s\n", featureID)
		}
		return 0
	case "list":
		featureID := ""
		if len(args) >= 2 {
			featureID = args[1]
		}
		dir, err := os.Getwd()
		if err != nil {
			return coreError(agentMode, err)
		}
		lines, err := core.ShowBDD(dir, featureID)
		if err != nil {
			return coreError(agentMode, err)
		}
		for _, l := range lines {
			fmt.Println(l)
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "err:user unknown bdd subcommand: %s\n", args[0])
		return 2
	}
}

// RunTest handles: ptsd test run [feature] | ptsd test map <bdd-file> <test-file>
func RunTest(args []string, agentMode bool) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "err:user usage: ptsd test <run|map> ...")
		return 2
	}
	switch args[0] {
	case "run":
		featureFilter := ""
		if len(args) >= 2 {
			featureFilter = args[1]
		}
		dir, err := os.Getwd()
		if err != nil {
			return coreError(agentMode, err)
		}
		results, err := core.RunTests(dir, featureFilter)
		if err != nil {
			return coreError(agentMode, err)
		}
		view := render.TestResultsView{
			Total:    results.Total,
			Passed:   results.Passed,
			Failed:   results.Failed,
			Failures: results.Failures,
		}
		r := &render.AgentRenderer{}
		fmt.Println(r.RenderTestResults(view))
		if results.Failed > 0 {
			return 5
		}
		return 0
	case "map":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "err:user usage: ptsd test map <bdd-file> <test-file>")
			return 2
		}
		bddFile := args[1]
		testFile := args[2]
		dir, err := os.Getwd()
		if err != nil {
			return coreError(agentMode, err)
		}
		err = core.MapTest(dir, bddFile, testFile)
		if err != nil {
			return coreError(agentMode, err)
		}
		if agentMode {
			fmt.Printf("mapped: %s -> %s\n", bddFile, testFile)
		} else {
			fmt.Printf("Mapped %s to %s\n", bddFile, testFile)
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "err:user unknown test subcommand: %s\n", args[0])
		return 2
	}
}
