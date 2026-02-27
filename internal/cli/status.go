package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/core"
	"github.com/veschin/ptsd/internal/render"
)

// RunStatus executes `ptsd status`. Returns an exit code.
func RunStatus(args []string, agentMode bool) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "err:io %s\n", err)
		return 4
	}

	result, err := core.ProjectStatus(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err:io %s\n", err)
		return 4
	}

	// Build StatusData from ProjectStatusResult.
	data := buildStatusData(cwd, result)

	if agentMode {
		r := &render.AgentRenderer{}

		// Print regression warnings before status line.
		for _, w := range result.Regressions {
			fmt.Fprintf(os.Stderr, "err:pipeline regression %s: %s\n", w.Feature, w.Message)
		}

		fmt.Println(r.RenderStatus(data))
	} else {
		// Human mode: simple table output (no Bubbletea dependency in cli layer).
		printStatusHuman(data, result.Regressions)
	}

	return 0
}

// buildStatusData converts a ProjectStatusResult into render.StatusData.
// It loads tasks independently to fill task counters.
func buildStatusData(projectDir string, result core.ProjectStatusResult) render.StatusData {
	data := render.StatusData{}

	// Feature counts.
	features := result.Features
	data.FeatTotal = len(features)
	for _, fs := range features {
		if fs.Stage == "" {
			data.FeatFail++
		}
	}

	// BDD counts: features that have a bdd hash recorded are considered to have BDD.
	bddFail := 0
	for _, fs := range features {
		if _, hasBDD := fs.Hashes["bdd"]; hasBDD {
			data.BDDTotal++
		} else {
			bddFail++
		}
	}
	data.BDDFail = bddFail

	// Test counts: features that have a test hash are considered to have tests.
	testFail := 0
	for _, fs := range features {
		if _, hasTest := fs.Hashes["test"]; hasTest {
			data.TestTotal++
		} else {
			testFail++
		}
	}
	data.TestFail = testFail

	// Task counts.
	tasks, _ := core.ListTasks(projectDir, "", "")
	data.TaskTotal = len(tasks)
	for _, t := range tasks {
		switch t.Status {
		case "WIP":
			data.TaskWIP++
		case "TODO":
			data.TaskTodo++
		case "DONE":
			data.TaskDone++
		}
	}

	return data
}

// printStatusHuman prints a simple human-readable status summary.
func printStatusHuman(data render.StatusData, regressions []core.RegressionWarning) {
	fmt.Printf("Features : %d total, %d without stage\n", data.FeatTotal, data.FeatFail)
	fmt.Printf("BDD      : %d covered, %d missing\n", data.BDDTotal, data.BDDFail)
	fmt.Printf("Tests    : %d covered, %d missing\n", data.TestTotal, data.TestFail)
	fmt.Printf("Tasks    : %d total  WIP:%d  TODO:%d  DONE:%d\n",
		data.TaskTotal, data.TaskWIP, data.TaskTodo, data.TaskDone)

	if len(regressions) > 0 {
		fmt.Println("\nRegressions:")
		for _, w := range regressions {
			fmt.Printf("  [%s] %s: %s\n", w.Severity, w.Feature, w.Message)
		}
	}
}
