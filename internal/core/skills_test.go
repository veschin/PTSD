package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSkillCreatesFileInCorrectLocation(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := GenerateSkill(dir, "bdd", "my-feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	skillPath := filepath.Join(dir, ".ptsd", "skills", "bdd-my-feature.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Fatal("skill file not created at expected path")
	}
}

func TestGenerateSkillCreatesFileWithCorrectContentStructure(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := GenerateSkill(dir, "prd", "auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "skills", "prd-auth.md"))
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		t.Error("skill file must start with YAML frontmatter (---)")
	}
	if !strings.Contains(content, "name:") {
		t.Error("frontmatter missing name field")
	}
	if !strings.Contains(content, "description:") {
		t.Error("frontmatter missing description field")
	}
	if !strings.Contains(content, "trigger:") {
		t.Error("frontmatter missing trigger field")
	}
}

func TestGenerateSkillInvalidStageFails(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := GenerateSkill(dir, "deploy", "my-feature")
	if err == nil {
		t.Fatal("expected error for invalid stage")
	}
	if !strings.HasPrefix(err.Error(), "err:user") {
		t.Errorf("expected err:user prefix, got: %v", err)
	}
}

func TestGenerateSkillCreatesSkillsDir(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	// skills/ dir does not exist yet
	skillsDir := filepath.Join(dir, ".ptsd", "skills")
	if _, err := os.Stat(skillsDir); !os.IsNotExist(err) {
		t.Fatal("skills dir should not exist before test")
	}

	err := GenerateSkill(dir, "impl", "core")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		t.Fatal("skills dir was not created")
	}
}

func TestListSkillsReturnsAllSkills(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	skillsDir := filepath.Join(ptsdDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write three skill files
	files := []string{"bdd-auth.md", "prd-config.md", "workflow.md"}
	for _, f := range files {
		content := "---\nname: test\ndescription: test\ntrigger: test\n---\n\nbody\n"
		if err := os.WriteFile(filepath.Join(skillsDir, f), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	skills, err := ListSkills(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(skills) != 3 {
		t.Errorf("expected 3 skills, got %d", len(skills))
	}
}

func TestListSkillsEmptyDirReturnsNil(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	skillsDir := filepath.Join(ptsdDir, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	skills, err := ListSkills(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}

func TestListSkillsMissingDirReturnsNil(t *testing.T) {
	dir := t.TempDir()

	skills, err := ListSkills(dir)
	if err != nil {
		t.Fatalf("unexpected error for missing skills dir: %v", err)
	}
	if skills != nil {
		t.Errorf("expected nil, got %v", skills)
	}
}

func TestListSkillsPopulatesFields(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, ".ptsd", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := "---\nname: test\ndescription: test\ntrigger: test\n---\n\nbody\n"
	if err := os.WriteFile(filepath.Join(skillsDir, "bdd-auth.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	skills, err := ListSkills(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	s := skills[0]
	if s.ID != "bdd-auth" {
		t.Errorf("expected ID=bdd-auth, got %q", s.ID)
	}
	if s.Stage != "bdd" {
		t.Errorf("expected Stage=bdd, got %q", s.Stage)
	}
	if s.Feature != "auth" {
		t.Errorf("expected Feature=auth, got %q", s.Feature)
	}
	if s.Path == "" {
		t.Error("Path must not be empty")
	}
}

func TestGenerateAllSkillsCreates13Files(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	err := GenerateAllSkills(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	skills, err := ListSkills(dir)
	if err != nil {
		t.Fatal(err)
	}

	// PRD says 13 skill files exist after init
	if len(skills) != 13 {
		t.Errorf("expected 13 skills, got %d", len(skills))
	}
}

func TestGenerateAllSkillsFrontmatterFields(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := GenerateAllSkills(dir); err != nil {
		t.Fatal(err)
	}

	skillsDir := filepath.Join(dir, ".ptsd", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(skillsDir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)
		if !strings.HasPrefix(content, "---\n") {
			t.Errorf("%s: missing YAML frontmatter start", entry.Name())
		}
		if !strings.Contains(content, "name:") {
			t.Errorf("%s: frontmatter missing name", entry.Name())
		}
		if !strings.Contains(content, "description:") {
			t.Errorf("%s: frontmatter missing description", entry.Name())
		}
		if !strings.Contains(content, "trigger:") {
			t.Errorf("%s: frontmatter missing trigger", entry.Name())
		}
	}
}

func TestWorkflowSkillExists(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.Mkdir(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := GenerateAllSkills(dir); err != nil {
		t.Fatal(err)
	}

	workflowPath := filepath.Join(dir, ".ptsd", "skills", "workflow.md")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("workflow.md not found: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "PRD") {
		t.Error("workflow.md must reference PRD stage")
	}
	if !strings.Contains(content, "Seed") || !strings.Contains(content, "seed") {
		t.Error("workflow.md must reference Seed stage")
	}
	if !strings.Contains(content, "BDD") || !strings.Contains(content, "bdd") {
		t.Error("workflow.md must reference BDD stage")
	}
}
