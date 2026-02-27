package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ReviewStatusEntry represents a feature's review status in review-status.yaml.
type ReviewStatusEntry struct {
	Stage      string
	Tests      string
	Review     string
	Issues     int
	IssuesList []string
}

func loadReviewStatus(projectDir string) (map[string]ReviewStatusEntry, error) {
	rsPath := filepath.Join(projectDir, ".ptsd", "review-status.yaml")
	data, err := os.ReadFile(rsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]ReviewStatusEntry), nil
		}
		return nil, fmt.Errorf("err:io %w", err)
	}
	return parseReviewStatus(string(data)), nil
}

func parseReviewStatus(content string) map[string]ReviewStatusEntry {
	entries := make(map[string]ReviewStatusEntry)
	lines := strings.Split(content, "\n")

	var currentFeature string
	var inIssuesList bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "features:" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))

		if indent == 2 && strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, " ") {
			currentFeature = strings.TrimSuffix(trimmed, ":")
			entries[currentFeature] = ReviewStatusEntry{
				Stage:  "prd",
				Tests:  "absent",
				Review: "pending",
			}
			inIssuesList = false
			continue
		}

		if currentFeature == "" {
			continue
		}

		if indent == 4 {
			inIssuesList = false
			e := entries[currentFeature]
			if strings.HasPrefix(trimmed, "stage: ") {
				e.Stage = strings.TrimPrefix(trimmed, "stage: ")
			} else if strings.HasPrefix(trimmed, "tests: ") {
				e.Tests = strings.TrimPrefix(trimmed, "tests: ")
			} else if strings.HasPrefix(trimmed, "review: ") {
				e.Review = strings.TrimPrefix(trimmed, "review: ")
			} else if strings.HasPrefix(trimmed, "issues: ") {
				n, _ := strconv.Atoi(strings.TrimPrefix(trimmed, "issues: "))
				e.Issues = n
			} else if trimmed == "issues_list:" {
				inIssuesList = true
			}
			entries[currentFeature] = e
			continue
		}

		if indent == 6 && inIssuesList && strings.HasPrefix(trimmed, "- ") {
			e := entries[currentFeature]
			item := strings.TrimPrefix(trimmed, "- ")
			item = strings.Trim(item, "\"")
			e.IssuesList = append(e.IssuesList, item)
			entries[currentFeature] = e
		}
	}

	return entries
}

func saveReviewStatus(projectDir string, entries map[string]ReviewStatusEntry) error {
	rsPath := filepath.Join(projectDir, ".ptsd", "review-status.yaml")

	var b strings.Builder
	b.WriteString("features:\n")

	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, id := range keys {
		e := entries[id]
		b.WriteString("  " + id + ":\n")
		b.WriteString("    stage: " + e.Stage + "\n")
		b.WriteString("    tests: " + e.Tests + "\n")
		b.WriteString("    review: " + e.Review + "\n")
		b.WriteString("    issues: " + strconv.Itoa(e.Issues) + "\n")
		if e.Issues > 0 && len(e.IssuesList) > 0 {
			b.WriteString("    issues_list:\n")
			for _, issue := range e.IssuesList {
				b.WriteString("      - \"" + issue + "\"\n")
			}
		}
	}

	return os.WriteFile(rsPath, []byte(b.String()), 0644)
}

var validReviewStages = map[string]bool{
	"prd":   true,
	"seed":  true,
	"bdd":   true,
	"tests": true,
	"impl":  true,
}

func RecordReview(projectDir string, featureID string, stage string, score int) error {
	if score < 0 || score > 10 {
		return fmt.Errorf("err:user score must be 0-10, got %d", score)
	}

	if !validReviewStages[stage] {
		return fmt.Errorf("err:user invalid stage %q: must be prd|seed|bdd|tests|impl", stage)
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

	// Update review-status.yaml
	cfg, err := LoadConfig(projectDir)
	if err != nil {
		// No config means default min_score=7
		cfg = &Config{Review: ReviewConfig{MinScore: 7}}
	}

	rs, err := loadReviewStatus(projectDir)
	if err != nil {
		return fmt.Errorf("err:io failed to load review-status: %w", err)
	}

	entry, ok := rs[featureID]
	if !ok {
		entry = ReviewStatusEntry{
			Stage:  stage,
			Tests:  "absent",
			Review: "pending",
		}
	}

	if score >= cfg.Review.MinScore {
		entry.Review = "passed"
		entry.Issues = 0
		entry.IssuesList = nil
	} else {
		entry.Review = "failed"
		entry.Issues = 1
		entry.IssuesList = []string{fmt.Sprintf("score %d below min %d at %s stage", score, cfg.Review.MinScore, stage)}
	}

	rs[featureID] = entry
	if err := saveReviewStatus(projectDir, rs); err != nil {
		return fmt.Errorf("err:io failed to save review-status: %w", err)
	}

	// Auto-redo check

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
		// No config means default min_score=7
		cfg = &Config{Review: ReviewConfig{MinScore: 7}}
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
