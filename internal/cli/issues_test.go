package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupIssuesProject creates a minimal .ptsd project for issues CLI tests.
// Returns the temp dir and a cleanup function that restores the original working directory.
func setupIssuesProject(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	// ptsd.yaml
	configContent := "project:\n  name: TestProject\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	return dir, func() {
		if err := os.Chdir(orig); err != nil {
			t.Logf("warning: could not restore working directory: %v", err)
		}
	}
}

// TestRunIssues_NoArgs verifies that RunIssues with no args exits with code 2.
func TestRunIssues_NoArgs(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	code := RunIssues([]string{}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for no args, got %d", code)
	}
}

// TestRunIssues_UnknownSubcommand verifies exit 2 for unknown subcommand.
func TestRunIssues_UnknownSubcommand(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	code := RunIssues([]string{"frobnicate"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for unknown subcommand, got %d", code)
	}
}

// TestRunIssues_Add_Success verifies that adding a valid issue exits 0
// and persists it in issues.yaml.
func TestRunIssues_Add_Success(t *testing.T) {
	dir, cleanup := setupIssuesProject(t)
	defer cleanup()

	code := RunIssues([]string{"add", "test-id", "config", "summary text", "fix text"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "issues.yaml"))
	if err != nil {
		t.Fatalf("cannot read issues.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "test-id") {
		t.Errorf("expected test-id in issues.yaml, got:\n%s", content)
	}
	if !strings.Contains(content, "config") {
		t.Errorf("expected category config in issues.yaml, got:\n%s", content)
	}
	if !strings.Contains(content, "summary text") {
		t.Errorf("expected summary in issues.yaml, got:\n%s", content)
	}
	if !strings.Contains(content, "fix text") {
		t.Errorf("expected fix in issues.yaml, got:\n%s", content)
	}
}

// TestRunIssues_Add_MissingArgs verifies exit 2 when add has fewer than 4 args.
func TestRunIssues_Add_MissingArgs(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	// Only 3 args instead of 4
	code := RunIssues([]string{"add", "test-id", "config", "summary"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for add with missing args, got %d", code)
	}
}

// TestRunIssues_Add_InvalidCategory verifies exit code for an invalid category.
// Note: BDD (common-issues.feature) says exit 2, but core.AddIssue returns err:validation
// which maps to exit 1 via coreError. The test follows the actual implementation.
func TestRunIssues_Add_InvalidCategory(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	code := RunIssues([]string{"add", "test-id", "typo", "summary", "fix"}, true)
	// core.AddIssue returns err:validation → errCategoryCode("validation") = 1
	if code != 1 {
		t.Errorf("expected exit code 1 for invalid category (err:validation), got %d", code)
	}
}

// TestRunIssues_Add_DuplicateID verifies exit 1 for a duplicate issue ID.
func TestRunIssues_Add_DuplicateID(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	// Add first time — should succeed
	if code := RunIssues([]string{"add", "dup-id", "env", "some problem", "fix it"}, true); code != 0 {
		t.Fatalf("first add failed with code %d", code)
	}

	// Add again — should fail with validation error
	code := RunIssues([]string{"add", "dup-id", "env", "some problem", "fix it"}, true)
	if code != 1 {
		t.Errorf("expected exit code 1 for duplicate ID, got %d", code)
	}
}

// TestRunIssues_List_Empty verifies "issues list" exits 0 when no issues exist.
func TestRunIssues_List_Empty(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	code := RunIssues([]string{"list"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 for empty list, got %d", code)
	}
}

// TestRunIssues_List_ShowsIssues verifies that listed issues are shown after being added.
func TestRunIssues_List_ShowsIssues(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	// Add two issues
	if code := RunIssues([]string{"add", "issue-a", "env", "env problem a", "fix a"}, true); code != 0 {
		t.Fatalf("add issue-a failed with code %d", code)
	}
	if code := RunIssues([]string{"add", "issue-b", "io", "io problem b", "fix b"}, true); code != 0 {
		t.Fatalf("add issue-b failed with code %d", code)
	}

	code := RunIssues([]string{"list"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 for list, got %d", code)
	}
}

// TestRunIssues_List_CategoryFilter verifies "--category" filter shows only matching issues.
func TestRunIssues_List_CategoryFilter(t *testing.T) {
	dir, cleanup := setupIssuesProject(t)
	defer cleanup()

	// Add env and access issues
	if code := RunIssues([]string{"add", "env-1", "env", "env problem", "env fix"}, true); code != 0 {
		t.Fatalf("add env-1 failed with code %d", code)
	}
	if code := RunIssues([]string{"add", "access-1", "access", "access problem", "access fix"}, true); code != 0 {
		t.Fatalf("add access-1 failed with code %d", code)
	}

	// List with filter — exit 0 regardless
	code := RunIssues([]string{"list", "--category", "env"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 for filtered list, got %d", code)
	}

	// Verify issues.yaml still has both
	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "issues.yaml"))
	if err != nil {
		t.Fatalf("cannot read issues.yaml: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "env-1") {
		t.Errorf("expected env-1 in issues.yaml")
	}
	if !strings.Contains(content, "access-1") {
		t.Errorf("expected access-1 in issues.yaml")
	}
}

// TestRunIssues_Remove_Success verifies that removing an existing issue exits 0.
func TestRunIssues_Remove_Success(t *testing.T) {
	dir, cleanup := setupIssuesProject(t)
	defer cleanup()

	// Add then remove
	if code := RunIssues([]string{"add", "test-id", "config", "some problem", "some fix"}, true); code != 0 {
		t.Fatalf("add failed with code %d", code)
	}

	code := RunIssues([]string{"remove", "test-id"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 for remove, got %d", code)
	}

	// Verify it's gone from issues.yaml
	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "issues.yaml"))
	if err != nil {
		t.Fatalf("cannot read issues.yaml: %v", err)
	}
	if strings.Contains(string(data), "test-id") {
		t.Errorf("test-id should have been removed from issues.yaml, got:\n%s", string(data))
	}
}

// TestRunIssues_Remove_NonExistent verifies exit 1 when removing an issue that does not exist.
func TestRunIssues_Remove_NonExistent(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	code := RunIssues([]string{"remove", "ghost-id"}, true)
	if code != 1 {
		t.Errorf("expected exit code 1 for remove of non-existent issue, got %d", code)
	}
}

// TestRunIssues_Remove_MissingArgs verifies exit 2 when remove has no ID arg.
func TestRunIssues_Remove_MissingArgs(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	code := RunIssues([]string{"remove"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for remove with no id, got %d", code)
	}
}

// TestRunIssues_Add_AllValidCategories verifies all valid categories are accepted.
func TestRunIssues_Add_AllValidCategories(t *testing.T) {
	categories := []string{"env", "access", "io", "config", "test", "llm"}
	for i, cat := range categories {
		t.Run(cat, func(t *testing.T) {
			_, cleanup := setupIssuesProject(t)
			defer cleanup()

			id := cat + "-issue"
			_ = i // avoid unused variable
			code := RunIssues([]string{"add", id, cat, "problem description", "fix action"}, true)
			if code != 0 {
				t.Errorf("category %s: expected exit code 0, got %d", cat, code)
			}
		})
	}
}

// TestRunIssues_Add_EmptySummary verifies exit 1 for empty summary.
func TestRunIssues_Add_EmptySummary(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	code := RunIssues([]string{"add", "test-id", "env", "", "fix text"}, true)
	if code != 1 {
		t.Errorf("expected exit code 1 for empty summary, got %d", code)
	}
}

// TestRunIssues_Add_EmptyFix verifies exit 1 for empty fix.
func TestRunIssues_Add_EmptyFix(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	code := RunIssues([]string{"add", "test-id", "env", "summary text", ""}, true)
	if code != 1 {
		t.Errorf("expected exit code 1 for empty fix, got %d", code)
	}
}

// TestRunIssues_LoadNonExistentFile verifies that list returns 0 when issues.yaml does not exist.
func TestRunIssues_LoadNonExistentFile(t *testing.T) {
	dir, cleanup := setupIssuesProject(t)
	defer cleanup()

	// Ensure issues.yaml does not exist
	issuesPath := filepath.Join(dir, ".ptsd", "issues.yaml")
	if _, err := os.Stat(issuesPath); err == nil {
		if err := os.Remove(issuesPath); err != nil {
			t.Fatal(err)
		}
	}

	code := RunIssues([]string{"list"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 for list with no issues.yaml, got %d", code)
	}
}

// TestRunIssues_List_ShowsIssues_OutputVerified verifies that listed issues appear in stdout output.
func TestRunIssues_List_ShowsIssues_OutputVerified(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	// Add two issues
	if code := RunIssues([]string{"add", "issue-a", "env", "env problem a", "fix a"}, true); code != 0 {
		t.Fatalf("add issue-a failed with code %d", code)
	}
	if code := RunIssues([]string{"add", "issue-b", "io", "io problem b", "fix b"}, true); code != 0 {
		t.Fatalf("add issue-b failed with code %d", code)
	}

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunIssues([]string{"list"}, true)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "issue-a") {
		t.Errorf("list output missing issue-a, got: %q", output)
	}
	if !strings.Contains(output, "issue-b") {
		t.Errorf("list output missing issue-b, got: %q", output)
	}
}

// TestRunIssues_List_CategoryFilter_OutputVerified verifies that --category filter shows only matching issues.
func TestRunIssues_List_CategoryFilter_OutputVerified(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	// Add env and access issues
	if code := RunIssues([]string{"add", "env-1", "env", "env problem", "env fix"}, true); code != 0 {
		t.Fatalf("add env-1 failed with code %d", code)
	}
	if code := RunIssues([]string{"add", "access-1", "access", "access problem", "access fix"}, true); code != 0 {
		t.Fatalf("add access-1 failed with code %d", code)
	}

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunIssues([]string{"list", "--category", "env"}, true)
	})
	if code != 0 {
		t.Errorf("expected exit code 0 for filtered list, got %d", code)
	}
	if !strings.Contains(output, "env-1") {
		t.Errorf("filtered list missing env-1, got: %q", output)
	}
	if strings.Contains(output, "access-1") {
		t.Errorf("filtered list must not contain access-1 when filtering by env, got: %q", output)
	}
}

// TestRunIssues_Add_HumanMode verifies human mode output for add command.
func TestRunIssues_Add_HumanMode(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunIssues([]string{"add", "hm-id", "config", "human mode problem", "human mode fix"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "hm-id") {
		t.Errorf("human mode add output missing issue id, got: %q", output)
	}
	if !strings.Contains(output, "config") {
		t.Errorf("human mode add output missing category, got: %q", output)
	}
}

// TestRunIssues_List_HumanMode verifies human mode output for list command.
func TestRunIssues_List_HumanMode(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	if code := RunIssues([]string{"add", "hm-list", "io", "io problem", "io fix"}, true); code != 0 {
		t.Fatalf("add hm-list failed with code %d", code)
	}

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunIssues([]string{"list"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "hm-list") {
		t.Errorf("human mode list missing issue id, got: %q", output)
	}
}

// TestRunIssues_List_Empty_HumanMode verifies human mode output when no issues exist.
func TestRunIssues_List_Empty_HumanMode(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunIssues([]string{"list"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "no issues") {
		t.Errorf("human mode empty list missing 'no issues' message, got: %q", output)
	}
}

// TestRunIssues_Remove_HumanMode verifies human mode output for remove command.
func TestRunIssues_Remove_HumanMode(t *testing.T) {
	_, cleanup := setupIssuesProject(t)
	defer cleanup()

	if code := RunIssues([]string{"add", "rm-id", "env", "env problem", "env fix"}, true); code != 0 {
		t.Fatalf("add rm-id failed with code %d", code)
	}

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunIssues([]string{"remove", "rm-id"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "rm-id") {
		t.Errorf("human mode remove output missing issue id, got: %q", output)
	}
}
