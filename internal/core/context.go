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
	Type     ContextLineType
	Feature  string
	Stage    string
	Action   string
	Reason   string
	Pipeline string
	// Task fields (only when Type == ContextTask)
	TaskID     string
	TaskStatus string
	TaskTitle  string
}

type ContextResult struct {
	Lines []ContextLine
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

	// Check if there are any active features
	hasActive := false
	for _, f := range features {
		if f.Status != "planned" && f.Status != "deferred" {
			hasActive = true
			break
		}
	}
	if !hasActive {
		result.Lines = append(result.Lines, ContextLine{
			Type:   ContextNext,
			Action: "add-feature",
			Reason: "no active features -- add with: ptsd feature add <id> \"title\"",
		})
		// Still emit tasks below
	}

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

		pipeline := f.Pipeline
		if pipeline == "" {
			pipeline = resolveDefaultPipeline(projectDir)
		}

		// Determine stage from artifacts if review-status is empty
		if stage == "" {
			stage = ComputeStageFromArtifacts(projectDir, f.ID)
			if stage == "" {
				stage = "prd"
			}
		}

		// Cross-check with state.yaml: if state has a more advanced stage, use it
		// (review-status.yaml may be stale if reviews were recorded before this fix)
		if stateStage := loadStateStage(projectDir, f.ID); stateStage != "" && stageOrder[stateStage] > stageOrder[stage] {
			stage = stateStage
		}

		// Check for blockers
		if review == "failed" {
			result.Lines = append(result.Lines, ContextLine{
				Type:     ContextBlocked,
				Feature:  f.ID,
				Stage:    stage,
				Reason:   fmt.Sprintf("review failed at %s stage", stage),
				Pipeline: pipeline,
			})
			continue
		}

		// Check missing prerequisites
		if blocked, reason := checkPrerequisite(projectDir, f.ID, stage, pipeline); blocked {
			result.Lines = append(result.Lines, ContextLine{
				Type:     ContextBlocked,
				Feature:  f.ID,
				Stage:    stage,
				Reason:   reason,
				Pipeline: pipeline,
			})
			continue
		}

		if stage == "impl" && review == "passed" {
			continue // skip done features -- saves tokens
		}

		if stage == "impl" && review == "pending" {
			result.Lines = append(result.Lines, ContextLine{
				Type:     ContextNext,
				Feature:  f.ID,
				Stage:    stage,
				Action:   "review-impl",
				Pipeline: pipeline,
			})
			continue
		}

		var action string
		if review == "passed" {
			// Stage reviewed and passed -- move to next stage
			action = NextAction(pipeline, stage)
			if action == "" {
				action = "review-" + stage
			}
		} else {
			// Stage not yet reviewed -- work on current stage
			action = "write-" + stage
		}

		result.Lines = append(result.Lines, ContextLine{
			Type:     ContextNext,
			Feature:  f.ID,
			Stage:    stage,
			Action:   action,
			Pipeline: pipeline,
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

// loadStateStage reads a feature's explicit stage from state.yaml.
func loadStateStage(projectDir, featureID string) string {
	state, err := LoadState(projectDir)
	if err != nil {
		return ""
	}
	if fs, ok := state.Features[featureID]; ok {
		return fs.Stage
	}
	return ""
}

func checkPrerequisite(projectDir, featureID, stage, pipeline string) (blocked bool, reason string) {
	switch stage {
	case "bdd":
		if StageRequired(pipeline, "seed") {
			seedPath := filepath.Join(projectDir, ".ptsd", "seeds", featureID, "seed.yaml")
			if !fileExists(seedPath) {
				return true, "missing seed"
			}
		}
	case "tests":
		if StageRequired(pipeline, "bdd") {
			bddPath := filepath.Join(projectDir, ".ptsd", "bdd", featureID+".feature")
			if !fileExists(bddPath) {
				return true, "missing bdd"
			}
		}
	}
	return false, ""
}
