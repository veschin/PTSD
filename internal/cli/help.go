package cli

import "fmt"

func RunHelp(args []string, agentMode bool) int {
	fmt.Println(`ptsd — PRD → Seed → BDD → Tests → Implementation

Project setup:
  init [--name <name>]     Initialize .ptsd/, .claude/, git hooks
  adopt                    Bootstrap ptsd onto existing project

Features:
  feature add <id> <title> Register a new feature
  feature list             All features and their status
  feature status <id> <s>  Set status (planned/in-progress/done)
  feature show <id>        Show feature details
  feature remove <id>      Remove a feature

Pipeline:
  seed add <feature>       Initialize seed data
  bdd add <feature>        Initialize BDD scenarios
  prd check                Validate PRD anchors
  test map <f> <file>      Map test file to feature
  test run <feature>       Run feature's tests
  review <f> <stage> <n>   Record review (score 0-10)
  validate                 Check all pipeline gates

Context & tracking:
  context                  Show pipeline state (next/blocked/done)
  status                   Project overview
  task next                Next task to work on
  task add <f> <title>     Add a task
  task done <id>           Mark task done

Other:
  config show              Show config
  skills                   List pipeline skills
  issues                   Common issues registry
  help                     This message

Flags:
  --agent                  Machine-readable output (all commands)`)
	return 0
}
