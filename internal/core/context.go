package core

import (
	"fmt"
	"path/filepath"
)

type ContextLineType string

const (
	ContextNext    ContextLineType = "next"
	ContextBlocked ContextLineType = "blocked"
	ContextDone    ContextLineType = "done"
	ContextTask    ContextLineType = "task"
)

type ContextLine struct {
	Type    ContextLineType
	Feature string
	Stage   string
	Action  string
	Reason  string
	// Task fields (only when Type == ContextTask)
	TaskID     string
	TaskStatus string
	TaskTitle  string
}

type ContextResult struct {
	Lines []ContextLine
}

var stageActions = map[string]string{
	"prd":   "write-seed",
	"seed":  "write-bdd",
	"bdd":   "write-tests",
	"tests": "write-impl",
}

func BuildContext(projectDir string) (ContextResult, error) {
	features, err := loadFeatures(projectDir)
	if err != nil {
		return ContextResult{}, err
	}

	rs, err := loadReviewStatus(projectDir)
	if err != nil {
		return ContextResult{}, err
	}

	tasks, err := loadTasks(projectDir)
	if err != nil {
		return ContextResult{}, err
	}

	var result ContextResult

	for _, f := range features {
		if f.Status == "planned" || f.Status == "deferred" {
			continue
		}

		entry, ok := rs[f.ID]
		stage := "prd"
		review := "pending"
		if ok {
			stage = entry.Stage
			review = entry.Review
		}

		// Determine stage from artifacts if review-status is empty
		if stage == "" {
			stage = ComputeStageFromArtifacts(projectDir, f.ID)
			if stage == "" {
				stage = "prd"
			}
		}

		// Check for blockers
		if review == "failed" {
			result.Lines = append(result.Lines, ContextLine{
				Type:    ContextBlocked,
				Feature: f.ID,
				Stage:   stage,
				Reason:  fmt.Sprintf("review failed at %s stage", stage),
			})
			continue
		}

		// Check missing prerequisites
		if blocked, reason := checkPrerequisite(projectDir, f.ID, stage); blocked {
			result.Lines = append(result.Lines, ContextLine{
				Type:    ContextBlocked,
				Feature: f.ID,
				Stage:   stage,
				Reason:  reason,
			})
			continue
		}

		if stage == "impl" && review == "passed" {
			result.Lines = append(result.Lines, ContextLine{
				Type:    ContextDone,
				Feature: f.ID,
				Stage:   stage,
			})
			continue
		}

		if stage == "impl" && review == "pending" {
			result.Lines = append(result.Lines, ContextLine{
				Type:    ContextNext,
				Feature: f.ID,
				Stage:   stage,
				Action:  "review-impl",
			})
			continue
		}

		action, ok := stageActions[stage]
		if !ok {
			action = "write-seed"
		}

		result.Lines = append(result.Lines, ContextLine{
			Type:    ContextNext,
			Feature: f.ID,
			Stage:   stage,
			Action:  action,
		})
	}

	// Emit TODO tasks
	for _, t := range tasks {
		if t.Status != "TODO" && t.Status != "WIP" {
			continue
		}
		result.Lines = append(result.Lines, ContextLine{
			Type:       ContextTask,
			Feature:    t.Feature,
			TaskID:     t.ID,
			TaskStatus: t.Status,
			TaskTitle:  t.Title,
		})
	}

	return result, nil
}

func checkPrerequisite(projectDir, featureID, stage string) (blocked bool, reason string) {
	switch stage {
	case "bdd":
		seedPath := filepath.Join(projectDir, ".ptsd", "seeds", featureID, "seed.yaml")
		if !fileExists(seedPath) {
			return true, "missing seed"
		}
	case "tests":
		bddPath := filepath.Join(projectDir, ".ptsd", "bdd", featureID+".feature")
		if !fileExists(bddPath) {
			return true, "missing bdd"
		}
	}
	return false, ""
}
