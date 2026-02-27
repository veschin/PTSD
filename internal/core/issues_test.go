package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddAndListIssues(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatal(err)
	}

	issue := Issue{
		ID:       "venv-wrong-python",
		Category: "env",
		Summary:  "venv uses system python instead of project python",
		Fix:      "rm -rf .venv && python3.12 -m venv .venv",
	}

	if err := AddIssue(dir, issue); err != nil {
		t.Fatalf("AddIssue failed: %v", err)
	}

	issues, err := ListIssues(dir, "")
	if err != nil {
		t.Fatalf("ListIssues failed: %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].ID != "venv-wrong-python" {
		t.Errorf("expected ID venv-wrong-python, got %s", issues[0].ID)
	}
	if issues[0].Category != "env" {
		t.Errorf("expected category env, got %s", issues[0].Category)
	}
	if issues[0].Summary != "venv uses system python instead of project python" {
		t.Errorf("unexpected summary: %s", issues[0].Summary)
	}
	if issues[0].Fix != "rm -rf .venv && python3.12 -m venv .venv" {
		t.Errorf("unexpected fix: %s", issues[0].Fix)
	}
}

func TestAddIssueDuplicateRejected(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatal(err)
	}

	issue := Issue{
		ID:       "missing-api-key",
		Category: "access",
		Summary:  "OPENAI_API_KEY not set in .env",
		Fix:      "cp .env.example .env && fill keys",
	}

	if err := AddIssue(dir, issue); err != nil {
		t.Fatalf("first AddIssue failed: %v", err)
	}

	err := AddIssue(dir, issue)
	if err == nil {
		t.Fatal("expected error for duplicate ID")
	}
	if err.Error() != "err:validation issue missing-api-key already exists" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRemoveIssue(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatal(err)
	}

	issues := []Issue{
		{ID: "issue-a", Category: "env", Summary: "problem a", Fix: "fix a"},
		{ID: "issue-b", Category: "io", Summary: "problem b", Fix: "fix b"},
	}
	for _, issue := range issues {
		if err := AddIssue(dir, issue); err != nil {
			t.Fatalf("AddIssue failed: %v", err)
		}
	}

	if err := RemoveIssue(dir, "issue-a"); err != nil {
		t.Fatalf("RemoveIssue failed: %v", err)
	}

	remaining, err := ListIssues(dir, "")
	if err != nil {
		t.Fatalf("ListIssues failed: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 issue after remove, got %d", len(remaining))
	}
	if remaining[0].ID != "issue-b" {
		t.Errorf("expected issue-b to remain, got %s", remaining[0].ID)
	}
}

func TestRemoveNonExistentIssue(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatal(err)
	}

	err := RemoveIssue(dir, "ghost-id")
	if err == nil {
		t.Fatal("expected error for non-existent issue")
	}
	if err.Error() != "err:validation issue ghost-id not found" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddIssueInvalidCategory(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatal(err)
	}

	issue := Issue{
		ID:       "bad-category",
		Category: "typo",
		Summary:  "some problem",
		Fix:      "fix it",
	}

	err := AddIssue(dir, issue)
	if err == nil {
		t.Fatal("expected error for invalid category")
	}
	if err.Error() != `err:validation invalid category "typo": must be env|access|io|config|test|llm` {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddIssueMissingSummary(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatal(err)
	}

	issue := Issue{
		ID:       "no-summary",
		Category: "env",
		Summary:  "",
		Fix:      "fix it",
	}

	err := AddIssue(dir, issue)
	if err == nil {
		t.Fatal("expected error for missing summary")
	}
	if err.Error() != "err:validation summary required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddIssueMissingFix(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatal(err)
	}

	issue := Issue{
		ID:       "no-fix",
		Category: "env",
		Summary:  "some problem",
		Fix:      "",
	}

	err := AddIssue(dir, issue)
	if err == nil {
		t.Fatal("expected error for missing fix")
	}
	if err.Error() != "err:validation fix required" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoadIssuesNonExistentFile(t *testing.T) {
	dir := t.TempDir()

	issues, err := LoadIssues(dir)
	if err != nil {
		t.Fatalf("expected no error for non-existent file, got: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected empty list, got %d issues", len(issues))
	}
}

func TestListIssuesFilteredByCategory(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatal(err)
	}

	issueList := []Issue{
		{ID: "env-issue", Category: "env", Summary: "env problem", Fix: "env fix"},
		{ID: "access-issue", Category: "access", Summary: "access problem", Fix: "access fix"},
		{ID: "env-issue-2", Category: "env", Summary: "another env problem", Fix: "another env fix"},
	}
	for _, issue := range issueList {
		if err := AddIssue(dir, issue); err != nil {
			t.Fatalf("AddIssue failed: %v", err)
		}
	}

	filtered, err := ListIssues(dir, "env")
	if err != nil {
		t.Fatalf("ListIssues failed: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 env issues, got %d", len(filtered))
	}
	for _, issue := range filtered {
		if issue.Category != "env" {
			t.Errorf("expected env category, got %s", issue.Category)
		}
	}
}

func TestAddMultipleIssuesAndListAll(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".ptsd"), 0755); err != nil {
		t.Fatal(err)
	}

	issueList := []Issue{
		{ID: "issue-1", Category: "env", Summary: "env problem", Fix: "env fix"},
		{ID: "issue-2", Category: "config", Summary: "config problem", Fix: "config fix"},
		{ID: "issue-3", Category: "llm", Summary: "llm problem", Fix: "llm fix"},
	}
	for _, issue := range issueList {
		if err := AddIssue(dir, issue); err != nil {
			t.Fatalf("AddIssue failed: %v", err)
		}
	}

	all, err := ListIssues(dir, "")
	if err != nil {
		t.Fatalf("ListIssues failed: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 issues, got %d", len(all))
	}
}
