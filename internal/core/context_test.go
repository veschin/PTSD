package core

import (
	"os"
	"path/filepath"
	"strings"
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

func TestBuildContext_ImplPendingReviewShowsReviewAction(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.MkdirAll(filepath.Join(ptsd, "docs"), 0755)
	os.WriteFile(filepath.Join(ptsd, "docs", "PRD.md"), []byte("<!-- feature:auth -->\n## Auth\n"), 0644)

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: impl\n    tests: written\n    review: pending\n    issues: 0\n",
	), 0644)

	os.WriteFile(filepath.Join(ptsd, "tasks.yaml"), []byte("tasks: []\n"), 0644)

	result, err := BuildContext(dir)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	found := false
	for _, line := range result.Lines {
		if line.Feature == "auth" && line.Type == ContextNext && line.Action == "review-impl" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected next action=review-impl for auth at impl+pending, got: %+v", result.Lines)
	}
}

func TestBuildContext_SkipsDeferred(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:deferred")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte("features: {}\n"), 0644)
	os.WriteFile(filepath.Join(ptsd, "tasks.yaml"), []byte("tasks: []\n"), 0644)

	result, err := BuildContext(dir)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	for _, line := range result.Lines {
		if line.Feature == "auth" {
			t.Errorf("deferred feature should be skipped, got: %+v", line)
		}
	}
}

func TestBuildContext_NoReviewStatusDefaultsToPrd(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	// Empty review-status — no entry for auth
	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte("features: {}\n"), 0644)
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
		t.Errorf("expected default prd stage with write-seed action, got: %+v", result.Lines)
	}
}

func TestBuildContext_BlockedByMissingPrerequisite(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.MkdirAll(filepath.Join(ptsd, "docs"), 0755)
	os.WriteFile(filepath.Join(ptsd, "docs", "PRD.md"), []byte("<!-- feature:auth -->\n## Auth\n"), 0644)

	// Feature at bdd stage but no seed
	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
		"features:\n  auth:\n    stage: bdd\n    tests: absent\n    review: pending\n    issues: 0\n",
	), 0644)
	os.WriteFile(filepath.Join(ptsd, "tasks.yaml"), []byte("tasks: []\n"), 0644)

	result, err := BuildContext(dir)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	found := false
	for _, line := range result.Lines {
		if line.Feature == "auth" && line.Type == ContextBlocked && strings.Contains(line.Reason, "missing seed") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected auth blocked by missing seed, got: %+v", result.Lines)
	}
}

func TestBuildContext_AllFeaturesStageProgression(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.MkdirAll(filepath.Join(ptsd, "docs"), 0755)
	os.WriteFile(filepath.Join(ptsd, "docs", "PRD.md"), []byte("<!-- feature:auth -->\n"), 0644)
	os.WriteFile(filepath.Join(ptsd, "tasks.yaml"), []byte("tasks: []\n"), 0644)

	// Create prerequisite artifacts so higher stages aren't blocked
	seedDir := filepath.Join(ptsd, "seeds", "auth")
	os.MkdirAll(seedDir, 0755)
	os.WriteFile(filepath.Join(seedDir, "seed.yaml"), []byte("feature: auth\n"), 0644)
	os.WriteFile(filepath.Join(ptsd, "bdd", "auth.feature"), []byte("@feature:auth\nFeature: Auth\n"), 0644)

	tests := []struct {
		stage  string
		action string
	}{
		{"prd", "write-seed"},
		{"seed", "write-bdd"},
		{"bdd", "write-tests"},
		{"tests", "write-impl"},
	}

	for _, tc := range tests {
		os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte(
			"features:\n  auth:\n    stage: "+tc.stage+"\n    tests: absent\n    review: passed\n    issues: 0\n",
		), 0644)

		result, err := BuildContext(dir)
		if err != nil {
			t.Fatalf("BuildContext at %s: %v", tc.stage, err)
		}

		found := false
		for _, line := range result.Lines {
			if line.Feature == "auth" && line.Type == ContextNext && line.Action == tc.action {
				found = true
			}
		}
		if !found {
			t.Errorf("at stage=%s expected action=%s, got: %+v", tc.stage, tc.action, result.Lines)
		}
	}
}

func TestBuildContext_DoneTasksExcluded(t *testing.T) {
	dir := setupProjectWithFeatures(t, "auth:in-progress")
	ptsd := filepath.Join(dir, ".ptsd")

	os.WriteFile(filepath.Join(ptsd, "review-status.yaml"), []byte("features: {}\n"), 0644)
	os.WriteFile(filepath.Join(ptsd, "tasks.yaml"), []byte(
		"tasks:\n  - id: T-1\n    feature: auth\n    title: Done task\n    status: DONE\n    priority: A\n",
	), 0644)

	result, err := BuildContext(dir)
	if err != nil {
		t.Fatalf("BuildContext: %v", err)
	}

	for _, line := range result.Lines {
		if line.Type == ContextTask && line.TaskID == "T-1" {
			t.Error("DONE task should not appear in context")
		}
	}
}
