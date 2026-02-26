package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddTask(t *testing.T) {
	dir := t.TempDir()
	setupTaskFeatures(t, dir, "user-auth")

	task, err := AddTask(dir, "user-auth", "Implement login endpoint", "A")
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if task.ID != "T-1" {
		t.Errorf("expected ID T-1, got %s", task.ID)
	}
	if task.Feature != "user-auth" {
		t.Errorf("expected feature user-auth, got %s", task.Feature)
	}
	if task.Priority != "A" {
		t.Errorf("expected priority A, got %s", task.Priority)
	}
	if task.Status != "TODO" {
		t.Errorf("expected status TODO, got %s", task.Status)
	}
}

func TestAddTaskWithoutFeature(t *testing.T) {
	dir := t.TempDir()
	setupTaskFeatures(t, dir)

	_, err := AddTask(dir, "", "Do something", "B")
	if err == nil {
		t.Fatal("expected error for missing feature")
	}
	if err.Error() != "err:user --feature required" {
		t.Errorf("expected err:user, got %v", err)
	}
}

func TestAddTaskWithNonexistentFeature(t *testing.T) {
	dir := t.TempDir()
	setupTaskFeatures(t, dir)

	_, err := AddTask(dir, "ghost", "Do something", "B")
	if err == nil {
		t.Fatal("expected error for nonexistent feature")
	}
	if err.Error() != "err:validation feature ghost not found" {
		t.Errorf("expected err:validation, got %v", err)
	}
}

func TestTaskNext(t *testing.T) {
	dir := t.TempDir()
	setupTaskFeatures(t, dir, "user-auth")
	setupTasks(t, dir,
		Task{ID: "T-1", Feature: "user-auth", Title: "Task A", Status: "TODO", Priority: "A"},
		Task{ID: "T-2", Feature: "user-auth", Title: "Task B", Status: "TODO", Priority: "B"},
		Task{ID: "T-3", Feature: "user-auth", Title: "Task C", Status: "WIP", Priority: "A"},
	)

	tasks, err := TaskNext(dir, 1)
	if err != nil {
		t.Fatalf("TaskNext failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != "T-1" {
		t.Errorf("expected T-1 (highest priority TODO), got %s", tasks[0].ID)
	}
}

func TestTaskNextWithLimit(t *testing.T) {
	dir := t.TempDir()
	setupTaskFeatures(t, dir, "user-auth")
	setupTasks(t, dir,
		Task{ID: "T-1", Feature: "user-auth", Title: "T1", Status: "TODO", Priority: "A"},
		Task{ID: "T-2", Feature: "user-auth", Title: "T2", Status: "TODO", Priority: "B"},
		Task{ID: "T-3", Feature: "user-auth", Title: "T3", Status: "TODO", Priority: "C"},
		Task{ID: "T-4", Feature: "user-auth", Title: "T4", Status: "TODO", Priority: "C"},
		Task{ID: "T-5", Feature: "user-auth", Title: "T5", Status: "TODO", Priority: "C"},
	)

	tasks, err := TaskNext(dir, 3)
	if err != nil {
		t.Fatalf("TaskNext failed: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("expected 3 tasks, got %d", len(tasks))
	}
}

func TestTaskNextWhenAllDone(t *testing.T) {
	dir := t.TempDir()
	setupTaskFeatures(t, dir, "user-auth")
	setupTasks(t, dir,
		Task{ID: "T-1", Feature: "user-auth", Title: "Done", Status: "DONE", Priority: "A"},
	)

	tasks, err := TaskNext(dir, 1)
	if err != nil {
		t.Fatalf("TaskNext failed: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestUpdateTaskStatus(t *testing.T) {
	dir := t.TempDir()
	setupTaskFeatures(t, dir, "user-auth")
	setupTasks(t, dir,
		Task{ID: "T-1", Feature: "user-auth", Title: "Task", Status: "TODO", Priority: "A"},
	)

	err := UpdateTask(dir, "T-1", "WIP")
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	tasks, _ := ListTasks(dir, "", "")
	for _, task := range tasks {
		if task.ID == "T-1" && task.Status != "WIP" {
			t.Errorf("expected status WIP, got %s", task.Status)
		}
	}
}

func TestListTasksFiltered(t *testing.T) {
	dir := t.TempDir()
	setupTaskFeatures(t, dir, "user-auth", "catalog")
	setupTasks(t, dir,
		Task{ID: "T-1", Feature: "user-auth", Title: "Auth task", Status: "TODO", Priority: "A"},
		Task{ID: "T-2", Feature: "catalog", Title: "Catalog task", Status: "TODO", Priority: "B"},
	)

	tasks, err := ListTasks(dir, "user-auth", "")
	if err != nil {
		t.Fatalf("ListTasks failed: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Feature != "user-auth" {
		t.Errorf("expected user-auth task, got %s", tasks[0].Feature)
	}
}

func TestAutoIncrementTaskIDs(t *testing.T) {
	dir := t.TempDir()
	setupTaskFeatures(t, dir, "catalog")
	setupTasks(t, dir,
		Task{ID: "T-1", Feature: "catalog", Title: "First", Status: "TODO", Priority: "A"},
		Task{ID: "T-2", Feature: "catalog", Title: "Second", Status: "TODO", Priority: "B"},
	)

	task, err := AddTask(dir, "catalog", "New task", "C")
	if err != nil {
		t.Fatalf("AddTask failed: %v", err)
	}
	if task.ID != "T-3" {
		t.Errorf("expected ID T-3, got %s", task.ID)
	}
}

func setupTaskFeatures(t *testing.T, dir string, names ...string) {
	t.Helper()
	featuresPath := filepath.Join(dir, ".ptsd", "features.yaml")
	content := "features:\n"
	for _, name := range names {
		content += "  - id: " + name + "\n    title: " + name + "\n    status: planned\n"
	}
	if err := os.MkdirAll(filepath.Dir(featuresPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(featuresPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func setupTasks(t *testing.T, dir string, tasks ...Task) {
	t.Helper()
	tasksPath := filepath.Join(dir, ".ptsd", "tasks.yaml")
	content := "tasks:\n"
	for _, task := range tasks {
		content += "  - id: " + task.ID + "\n"
		content += "    feature: " + task.Feature + "\n"
		content += "    title: " + task.Title + "\n"
		content += "    status: " + task.Status + "\n"
		content += "    priority: " + task.Priority + "\n"
	}
	if err := os.MkdirAll(filepath.Dir(tasksPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tasksPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
