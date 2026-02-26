package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Task struct {
	ID       string
	Feature  string
	Title    string
	Status   string
	Priority string
}

func AddTask(projectDir string, featureID string, title string, priority string) (Task, error) {
	if featureID == "" {
		return Task{}, fmt.Errorf("err:user --feature required")
	}

	features, err := loadFeatures(projectDir)
	if err != nil {
		return Task{}, err
	}

	found := false
	for _, f := range features {
		if f.ID == featureID {
			found = true
			break
		}
	}
	if !found {
		return Task{}, fmt.Errorf("err:validation feature %s not found", featureID)
	}

	tasks, err := loadTasks(projectDir)
	if err != nil {
		return Task{}, err
	}

	maxNum := 0
	for _, t := range tasks {
		if strings.HasPrefix(t.ID, "T-") {
			n, err := strconv.Atoi(t.ID[2:])
			if err == nil && n > maxNum {
				maxNum = n
			}
		}
	}

	task := Task{
		ID:       fmt.Sprintf("T-%d", maxNum+1),
		Feature:  featureID,
		Title:    title,
		Status:   "TODO",
		Priority: priority,
	}

	tasks = append(tasks, task)
	if err := saveTasks(projectDir, tasks); err != nil {
		return Task{}, err
	}

	return task, nil
}

func ListTasks(projectDir string, featureFilter string, statusFilter string) ([]Task, error) {
	tasks, err := loadTasks(projectDir)
	if err != nil {
		return nil, err
	}

	var filtered []Task
	for _, t := range tasks {
		if featureFilter != "" && t.Feature != featureFilter {
			continue
		}
		if statusFilter != "" && t.Status != statusFilter {
			continue
		}
		filtered = append(filtered, t)
	}

	return filtered, nil
}

func UpdateTask(projectDir string, id string, status string) error {
	tasks, err := loadTasks(projectDir)
	if err != nil {
		return err
	}

	found := false
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Status = status
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("err:validation task %s not found", id)
	}

	return saveTasks(projectDir, tasks)
}

func TaskNext(projectDir string, limit int) ([]Task, error) {
	tasks, err := loadTasks(projectDir)
	if err != nil {
		return nil, err
	}

	var todo []Task
	for _, t := range tasks {
		if t.Status == "TODO" {
			todo = append(todo, t)
		}
	}

	sort.Slice(todo, func(i, j int) bool {
		return todo[i].Priority < todo[j].Priority
	})

	if limit > 0 && len(todo) > limit {
		todo = todo[:limit]
	}

	return todo, nil
}

func loadTasks(projectDir string) ([]Task, error) {
	tasksPath := filepath.Join(projectDir, ".ptsd", "tasks.yaml")
	data, err := os.ReadFile(tasksPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("err:io %w", err)
	}

	var tasks []Task
	lines := strings.Split(string(data), "\n")
	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "- id: ") {
			t := Task{ID: strings.TrimPrefix(trimmed, "- id: ")}
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if strings.HasPrefix(next, "- id: ") || next == "" || (!strings.HasPrefix(lines[j], "    ") && !strings.HasPrefix(lines[j], "  ")) {
					break
				}
				if strings.HasPrefix(next, "feature: ") {
					t.Feature = strings.TrimPrefix(next, "feature: ")
				}
				if strings.HasPrefix(next, "title: ") {
					t.Title = strings.TrimPrefix(next, "title: ")
				}
				if strings.HasPrefix(next, "status: ") {
					t.Status = strings.TrimPrefix(next, "status: ")
				}
				if strings.HasPrefix(next, "priority: ") {
					t.Priority = strings.TrimPrefix(next, "priority: ")
				}
			}
			tasks = append(tasks, t)
		}
	}

	return tasks, nil
}

func saveTasks(projectDir string, tasks []Task) error {
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

	return os.WriteFile(tasksPath, []byte(b.String()), 0644)
}
