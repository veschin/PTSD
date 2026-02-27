package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/core"
)

// RunSkills handles the `ptsd skills` command.
// Subcommands:
//   ptsd skills generate <stage> <feature>
//   ptsd skills generate-all
//   ptsd skills list
func RunSkills(args []string, agentMode bool) int {
	cwd, err := os.Getwd()
	if err != nil {
		return renderError(agentMode, "io", "cannot determine working directory: "+err.Error())
	}

	if len(args) == 0 {
		return renderError(agentMode, "user", "usage: ptsd skills generate <stage> <feature> | ptsd skills generate-all | ptsd skills list")
	}

	switch args[0] {
	case "generate":
		return runSkillsGenerate(args[1:], cwd, agentMode)
	case "generate-all":
		return runSkillsGenerateAll(cwd, agentMode)
	case "list":
		return runSkillsList(cwd, agentMode)
	default:
		return renderError(agentMode, "user", "unknown subcommand: "+args[0])
	}
}

func runSkillsGenerate(args []string, cwd string, agentMode bool) int {
	if len(args) < 2 {
		return renderError(agentMode, "user", "usage: ptsd skills generate <stage> <feature>")
	}

	stage := args[0]
	feature := args[1]

	if err := core.GenerateSkill(cwd, stage, feature); err != nil {
		return coreError(agentMode, err)
	}

	if agentMode {
		fmt.Printf("generated skill: %s-%s\n", stage, feature)
	} else {
		fmt.Printf("skill generated: stage=%s feature=%s\n", stage, feature)
	}

	return 0
}

func runSkillsGenerateAll(cwd string, agentMode bool) int {
	if err := core.GenerateAllSkills(cwd); err != nil {
		return coreError(agentMode, err)
	}

	if agentMode {
		fmt.Println("generated all skills")
	} else {
		fmt.Println("all standard skills generated")
	}

	return 0
}

func runSkillsList(cwd string, agentMode bool) int {
	skills, err := core.ListSkills(cwd)
	if err != nil {
		return coreError(agentMode, err)
	}

	if len(skills) == 0 {
		if agentMode {
			// no output for empty list in agent mode
		} else {
			fmt.Println("no skills found")
		}
		return 0
	}

	for _, s := range skills {
		if agentMode {
			fmt.Printf("%s %s %s %s\n", s.ID, s.Stage, s.Feature, s.Path)
		} else {
			fmt.Printf("%-30s stage=%-6s feature=%-20s path=%s\n", s.ID, s.Stage, s.Feature, s.Path)
		}
	}

	return 0
}
