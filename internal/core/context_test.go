package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildContext_NextAction(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	// Create PRD anchor
	os.MkdirAll(filepath.Join(ptsd, "docs"), 0755)
	os.WriteFile(filepath.Join(ptsd, "docs", "PRD.md"), []byte("<!-- feature:auth -->\n## Auth\n"), 0644)

	// review-status at prd stage
	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: prd\n    tests: absent\n    review: passed\n    issues: 0\n",
	), 0644)

	os.WriteFile(filepath.Join(ptsd, "tasks.yaml"), []byte("tasks: []\n"), 0644)

	result, err := BuildContext(dir)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	found := false
	for _, line := range result.Lines {
		if line.Feature == "auth" && line.Type == ContextNext && line.Action == "write-seed" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected next action=write-seed for auth at prd stage, got: %+v", result.Lines)
	}
}

func TestBuildContext_Blocked(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.MkdirAll(filepath.Join(ptsd, "docs"), 0755)
	os.WriteFile(filepath.Join(ptsd, "docs", "PRD.md"), []byte("<!-- feature:auth -->\n## Auth\n"), 0644)

	// review-status with failed review
	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: seed\n    tests: absent\n    review: failed\n    issues: 1\n    issues_list:\n      - \"low quality\"\n",
	), 0644)

	os.WriteFile(filepath.Join(ptsd, "tasks.yaml"), []byte("tasks: []\n"), 0644)

	result, err := BuildContext(dir)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	found := false
	for _, line := range result.Lines {
		if line.Feature == "auth" && line.Type == ContextBlocked {
			found = true
		}
	}
	if !found {
		t.Errorf("expected auth to be blocked with failed review, got: %+v", result.Lines)
	}
}

func TestBuildContext_Done(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.MkdirAll(filepath.Join(ptsd, "docs"), 0755)
	os.WriteFile(filepath.Join(ptsd, "docs", "PRD.md"), []byte("<!-- feature:auth -->\n## Auth\n"), 0644)

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: impl\n    tests: written\n    review: passed\n    issues: 0\n",
	), 0644)

	os.WriteFile(filepath.Join(ptsd, "tasks.yaml"), []byte("tasks: []\n"), 0644)

	result, err := BuildContext(dir)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	found := false
	for _, line := range result.Lines {
		if line.Feature == "auth" && line.Type == ContextDone {
			found = true
		}
	}
	if !found {
		t.Errorf("expected auth to be done, got: %+v", result.Lines)
	}
}

func TestBuildContext_SkipsPlanned(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:planned")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte("features: {}\n"), 0644)
	os.WriteFile(filepath.Join(ptsd, "tasks.yaml"), []byte("tasks: []\n"), 0644)

	result, err := BuildContext(dir)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	for _, line := range result.Lines {
		if line.Feature == "auth" {
			t.Errorf("planned feature should be skipped, got: %+v", line)
		}
	}
}

func TestBuildContext_IncludesTasks(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte("features: {}\n"), 0644)
	os.WriteFile(filepath.Join(ptsd, "tasks.yaml"), []byte(
		"tasks:\n  - id: T-1\n    feature: auth\n    title: Write tests\n    status: TODO\n    priority: A\n",
	), 0644)

	result, err := BuildContext(dir)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	found := false
	for _, line := range result.Lines {
		if line.Type == ContextTask && line.TaskID == "T-1" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected TODO task in context, got: %+v", result.Lines)
	}
}
