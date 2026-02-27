package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/veschin/ptsd/internal/core"
)

// RunReview handles the `ptsd review` command.
// Subcommands:
//
//	ptsd review <feature> <stage> <score>
//	ptsd review gate <feature> <stage>
func RunReview(args []string, agentMode bool) int {
	cwd, err := os.Getwd()
	if err != nil {
		return renderError(agentMode, "io", err.Error())
	}

	if len(args) == 0 {
		return renderError(agentMode, "user", "usage: ptsd review <feature> <stage> <score> | ptsd review gate <feature> <stage>")
	}

	if args[0] == "gate" {
		return runReviewGate(args[1:], cwd, agentMode)
	}

	return runReviewRecord(args, cwd, agentMode)
}

func runReviewRecord(args []string, cwd string, agentMode bool) int {
	if len(args) < 3 {
		return renderError(agentMode, "user", "usage: ptsd review <feature> <stage> <score>")
	}

	feature := args[0]
	stage := args[1]
	scoreStr := args[2]

	score, err := strconv.Atoi(scoreStr)
	if err != nil {
		return renderError(agentMode, "user", "score must be an integer, got: "+scoreStr)
	}

	if err := core.RecordReview(cwd, feature, stage, score); err != nil {
		return coreError(agentMode, err)
	}

	cfg, _ := core.LoadConfig(cwd)
	minScore := cfg.Review.MinScore
	if minScore == 0 {
		minScore = 7
	}

	verdict := "pass"
	if score < minScore {
		verdict = "fail"
	}

	if agentMode {
		fmt.Printf("score:%d verdict:%s\n", score, verdict)
	} else {
		fmt.Printf("review recorded: feature=%s stage=%s score=%d verdict=%s\n", feature, stage, score, verdict)
	}

	return 0
}

func runReviewGate(args []string, cwd string, agentMode bool) int {
	if len(args) < 2 {
		return renderError(agentMode, "user", "usage: ptsd review gate <feature> <stage>")
	}

	feature := args[0]
	stage := args[1]

	passed, err := core.CheckReviewGate(cwd, feature, stage)
	if err != nil {
		return coreError(agentMode, err)
	}

	verdict := "fail"
	if passed {
		verdict = "pass"
	}

	if agentMode {
		fmt.Printf("gate:%s feature:%s stage:%s\n", verdict, feature, stage)
	} else {
		fmt.Printf("review gate %s: feature=%s stage=%s\n", verdict, feature, stage)
	}

	if !passed {
		return 1
	}

	return 0
}
