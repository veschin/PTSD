package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RecordReview(projectDir string, featureID string, stage string, score int) error {
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
		return nil // config not required for review recording
	}

	if cfg.Review.AutoRedo && score < cfg.Review.MinScore {
		title := fmt.Sprintf("redo %s for %s", stage, featureID)
		tasks, _ := loadTasks(projectDir)
		maxNum := 0
		for _, t := range tasks {
			if strings.HasPrefix(t.ID, "T-") {
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

		tasksPath := filepath.Join(projectDir, ".ptsd", "tasks.yaml")
		var b strings.Builder
		b.WriteString("tasks:\n")
		for _, t := range tasks {
			b.WriteString("  - id: " + t.ID + "\n")
			b.WriteString("    feature: " + t.Feature + "\n")
			b.WriteString("    title: " + t.Title + "\n")
			b.WriteString("    status: " + t.Status + "\n")
			b.WriteString("    priority: " + t.Priority + "\n")
		}
		os.WriteFile(tasksPath, []byte(b.String()), 0644)
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
