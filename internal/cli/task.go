package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/veschin/ptsd/internal/core"
	"github.com/veschin/ptsd/internal/render"
)

func RunTask(args []string, agentMode bool) int {
	r := newRenderer(agentMode)

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, r.RenderError("user", "subcommand required: add|list|next|update"))
		return 2
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, r.RenderError("io", err.Error()))
		return 4
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "add":
		return runTaskAdd(cwd, rest, agentMode)
	case "list":
		return runTaskList(cwd, rest, agentMode)
	case "next":
		return runTaskNext(cwd, rest, agentMode)
	case "update":
		return runTaskUpdate(cwd, rest, agentMode)
	default:
		fmt.Fprintln(os.Stderr, r.RenderError("user", fmt.Sprintf("unknown subcommand %q: use add|list|next|update", sub)))
		return 2
	}
}

// runTaskAdd handles: task add <feature> <title> [--priority A|B|C]
func runTaskAdd(cwd string, args []string, agentMode bool) int {
	r := newRenderer(agentMode)

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, r.RenderError("user", "usage: task add <feature> <title> [--priority A|B|C]"))
		return 2
	}

	feature := args[0]
	priority := "B"

	// Collect title tokens and parse --priority flag
	var titleParts []string
	for i := 1; i < len(args); i++ {
		if args[i] == "--priority" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, r.RenderError("user", "--priority requires a value: A|B|C"))
				return 2
			}
			priority = strings.ToUpper(args[i+1])
			i++
		} else {
			titleParts = append(titleParts, args[i])
		}
	}

	title := strings.Join(titleParts, " ")
	if title == "" {
		fmt.Fprintln(os.Stderr, r.RenderError("user", "title is required"))
		return 2
	}

	task, err := core.AddTask(cwd, feature, title, priority)
	if err != nil {
		return coreError(agentMode, err)
	}

	fmt.Printf("%s %s [%s] [%s]: %s\n", task.ID, task.Feature, task.Status, task.Priority, task.Title)
	return 0
}

// runTaskList handles: task list
func runTaskList(cwd string, args []string, agentMode bool) int {
	tasks, err := core.ListTasks(cwd, "", "")
	if err != nil {
		return coreError(agentMode, err)
	}

	for _, t := range tasks {
		fmt.Printf("%s %s [%s] [%s]: %s\n", t.ID, t.Feature, t.Status, t.Priority, t.Title)
	}
	return 0
}

// runTaskNext handles: task next [--limit N]
func runTaskNext(cwd string, args []string, agentMode bool) int {
	r := newRenderer(agentMode)
	limit := 1

	for i := 0; i < len(args); i++ {
		if args[i] == "--limit" {
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, r.RenderError("user", "--limit requires a numeric value"))
				return 2
			}
			n, err := strconv.Atoi(args[i+1])
			if err != nil || n < 1 {
				fmt.Fprintln(os.Stderr, r.RenderError("user", fmt.Sprintf("invalid --limit value %q: must be a positive integer", args[i+1])))
				return 2
			}
			limit = n
			i++
		}
	}

	tasks, err := core.TaskNext(cwd, limit)
	if err != nil {
		return coreError(agentMode, err)
	}

	views := make([]render.TaskView, len(tasks))
	for i, t := range tasks {
		views[i] = render.TaskView{
			ID:       t.ID,
			Status:   t.Status,
			Priority: t.Priority,
			Title:    t.Title,
		}
	}

	out := r.RenderTaskNext(views)
	if out != "" {
		fmt.Println(out)
	}
	return 0
}

// runTaskUpdate handles: task update <id> <status>
func runTaskUpdate(cwd string, args []string, agentMode bool) int {
	r := newRenderer(agentMode)

	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, r.RenderError("user", "usage: task update <id> <status>"))
		return 2
	}

	id := args[0]
	status := strings.ToUpper(args[1])

	if err := core.UpdateTask(cwd, id, status); err != nil {
		return coreError(agentMode, err)
	}

	fmt.Printf("%s updated to %s\n", id, status)
	return 0
}
