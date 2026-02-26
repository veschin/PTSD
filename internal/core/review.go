package core

import (
	"fmt"
	"time"
)

var validReviewStages = map[string]bool{
	"prd":  true,
	"seed": true,
	"bdd":  true,
	"test": true,
	"impl": true,
}

func RecordReview(projectDir string, featureID string, stage string, score int) error {
	if score < 0 || score > 10 {
		return fmt.Errorf("err:user score must be 0-10, got %d", score)
	}

	if !validReviewStages[stage] {
		return fmt.Errorf("err:user invalid stage %q: must be prd|seed|bdd|test|impl", stage)
	}

	state, err := LoadState(projectDir)
	if err != nil {
		return err
	}

	fs, ok := state.Features[featureID]
	if !ok {
		fs = FeatureState{
			Hashes: make(map[string]string),
			Scores: make(map[string]ScoreEntry),
		}
	}
	if fs.Scores == nil {
		fs.Scores = make(map[string]ScoreEntry)
	}

	fs.Scores[stage] = ScoreEntry{
		Value:     score,
		Timestamp: time.Now(),
	}
	state.Features[featureID] = fs

	if err := writeState(projectDir, state); err != nil {
		return err
	}

	// Auto-redo check
	cfg, err := LoadConfig(projectDir)
	if err != nil {
		return nil
	}

	if cfg.Review.AutoRedo && score < cfg.Review.MinScore {
		title := fmt.Sprintf("redo %s for %s", stage, featureID)
		tasks, _ := loadTasks(projectDir)

		maxNum := 0
		for _, t := range tasks {
			if len(t.ID) > 2 && t.ID[:2] == "T-" {
				n := 0
				fmt.Sscanf(t.ID[2:], "%d", &n)
				if n > maxNum {
					maxNum = n
				}
			}
		}
		redoTask := Task{
			ID:       fmt.Sprintf("T-%d", maxNum+1),
			Feature:  featureID,
			Title:    title,
			Status:   "TODO",
			Priority: "A",
		}
		tasks = append(tasks, redoTask)

		if err := saveTasks(projectDir, tasks); err != nil {
			return fmt.Errorf("err:io failed to save redo task: %w", err)
		}
	}

	return nil
}

func CheckReviewGate(projectDir string, featureID string, stage string) (bool, error) {
	cfg, err := LoadConfig(projectDir)
	if err != nil {
		return false, err
	}

	state, err := LoadState(projectDir)
	if err != nil {
		return false, err
	}

	fs, ok := state.Features[featureID]
	if !ok {
		return false, nil
	}

	score, ok := fs.Scores[stage]
	if !ok {
		return false, nil
	}

	return score.Value >= cfg.Review.MinScore, nil
}
