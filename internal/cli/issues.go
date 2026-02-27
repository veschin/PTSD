package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/core"
)

// RunIssues handles the `ptsd issues` command.
// Subcommands:
//   ptsd issues add <id> <category> <summary> <fix>
//   ptsd issues list [--category <cat>]
//   ptsd issues remove <id>
func RunIssues(args []string, agentMode bool) int {
	cwd, err := os.Getwd()
	if err != nil {
		return renderError(agentMode, "io", "cannot determine working directory: "+err.Error())
	}

	if len(args) == 0 {
		return renderError(agentMode, "user", "usage: ptsd issues add <id> <category> <summary> <fix> | ptsd issues list [--category <cat>] | ptsd issues remove <id>")
	}

	switch args[0] {
	case "add":
		return runIssuesAdd(args[1:], cwd, agentMode)
	case "list":
		return runIssuesList(args[1:], cwd, agentMode)
	case "remove":
		return runIssuesRemove(args[1:], cwd, agentMode)
	default:
		return renderError(agentMode, "user", "unknown subcommand: "+args[0])
	}
}

func runIssuesAdd(args []string, cwd string, agentMode bool) int {
	if len(args) < 4 {
		return renderError(agentMode, "user", "usage: ptsd issues add <id> <category> <summary> <fix>")
	}

	issue := core.Issue{
		ID:       args[0],
		Category: args[1],
		Summary:  args[2],
		Fix:      args[3],
	}

	if err := core.AddIssue(cwd, issue); err != nil {
		return coreError(agentMode, err)
	}

	if agentMode {
		fmt.Printf("added issue: %s\n", issue.ID)
	} else {
		fmt.Printf("issue added: id=%s category=%s\n", issue.ID, issue.Category)
	}

	return 0
}

func runIssuesList(args []string, cwd string, agentMode bool) int {
	categoryFilter := ""
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--category" {
			categoryFilter = args[i+1]
			break
		}
	}

	issues, err := core.ListIssues(cwd, categoryFilter)
	if err != nil {
		return coreError(agentMode, err)
	}

	if len(issues) == 0 {
		if !agentMode {
			fmt.Println("no issues found")
		}
		return 0
	}

	for _, issue := range issues {
		if agentMode {
			fmt.Printf("%s [%s] %s\n", issue.ID, issue.Category, issue.Summary)
		} else {
			fmt.Printf("%-20s [%-8s] %s\n", issue.ID, issue.Category, issue.Summary)
		}
	}

	return 0
}

func runIssuesRemove(args []string, cwd string, agentMode bool) int {
	if len(args) < 1 {
		return renderError(agentMode, "user", "usage: ptsd issues remove <id>")
	}

	id := args[0]

	if err := core.RemoveIssue(cwd, id); err != nil {
		return coreError(agentMode, err)
	}

	if agentMode {
		fmt.Printf("removed issue: %s\n", id)
	} else {
		fmt.Printf("issue removed: id=%s\n", id)
	}

	return 0
}
