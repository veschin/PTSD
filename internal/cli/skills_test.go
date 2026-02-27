package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupSkillsProject creates a minimal .ptsd project for skills CLI tests.
// Returns the temp dir and a cleanup function that restores the original working directory.
func setupSkillsProject(t *testing.T) (string, func()) {
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

	// features.yaml
	featuresContent := "features:\n  - id: my-feat\n    status: planned\n"
	if err := os.WriteFile(filepath.Join(ptsdDir, "features.yaml"), []byte(featuresContent), 0644); err != nil {
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

// TestRunSkills_NoArgs verifies that RunSkills with no args exits with code 2.
func TestRunSkills_NoArgs(t *testing.T) {
	_, cleanup := setupSkillsProject(t)
	defer cleanup()

	code := RunSkills([]string{}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for no args, got %d", code)
	}
}

// TestRunSkills_UnknownSubcommand verifies exit 2 for unknown subcommand.
func TestRunSkills_UnknownSubcommand(t *testing.T) {
	_, cleanup := setupSkillsProject(t)
	defer cleanup()

	code := RunSkills([]string{"frobnicate"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for unknown subcommand, got %d", code)
	}
}

// TestRunSkills_List_Empty verifies that "skills list" with no skills exits 0.
func TestRunSkills_List_Empty(t *testing.T) {
	_, cleanup := setupSkillsProject(t)
	defer cleanup()

	code := RunSkills([]string{"list"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 for empty list, got %d", code)
	}
}

// TestRunSkills_List_WithSkills verifies "skills list" returns 0 and lists the skill.
func TestRunSkills_List_WithSkills(t *testing.T) {
	dir, cleanup := setupSkillsProject(t)
	defer cleanup()

	// Pre-create a skill file so list has something to return
	skillsDir := filepath.Join(dir, ".ptsd", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}
	skillContent := "---\nname: bdd-my-feat\ndescription: test skill\ntrigger: test\n---\n\n## body\n"
	if err := os.WriteFile(filepath.Join(skillsDir, "bdd-my-feat.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	code := RunSkills([]string{"list"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 for list with skills, got %d", code)
	}
}

// TestRunSkills_Generate creates a skill file for a specific stage+feature.
func TestRunSkills_Generate(t *testing.T) {
	dir, cleanup := setupSkillsProject(t)
	defer cleanup()

	code := RunSkills([]string{"generate", "impl", "my-feat"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	skillPath := filepath.Join(dir, ".ptsd", "skills", "impl-my-feat.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Errorf("expected skill file at %s, not found", skillPath)
	}
}

// TestRunSkills_Generate_CreatesFileWithFrontmatter verifies generated file has YAML frontmatter.
func TestRunSkills_Generate_CreatesFileWithFrontmatter(t *testing.T) {
	dir, cleanup := setupSkillsProject(t)
	defer cleanup()

	code := RunSkills([]string{"generate", "bdd", "my-feat"}, true)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "skills", "bdd-my-feat.md"))
	if err != nil {
		t.Fatalf("cannot read generated skill: %v", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		t.Error("generated skill must start with YAML frontmatter (---)")
	}
	if !strings.Contains(content, "name:") {
		t.Error("generated skill missing name field in frontmatter")
	}
	if !strings.Contains(content, "description:") {
		t.Error("generated skill missing description field in frontmatter")
	}
	if !strings.Contains(content, "trigger:") {
		t.Error("generated skill missing trigger field in frontmatter")
	}
}

// TestRunSkills_Generate_InvalidStage verifies that an invalid stage exits non-zero.
func TestRunSkills_Generate_InvalidStage(t *testing.T) {
	_, cleanup := setupSkillsProject(t)
	defer cleanup()

	code := RunSkills([]string{"generate", "deploy", "my-feat"}, true)
	if code == 0 {
		t.Error("expected non-zero exit code for invalid stage")
	}
}

// TestRunSkills_Generate_MissingArgs verifies exit 2 when generate has fewer than 2 args.
func TestRunSkills_Generate_MissingArgs(t *testing.T) {
	_, cleanup := setupSkillsProject(t)
	defer cleanup()

	code := RunSkills([]string{"generate", "impl"}, true)
	if code != 2 {
		t.Errorf("expected exit code 2 for generate with missing args, got %d", code)
	}
}

// TestRunSkills_GenerateAll creates all standard skill files.
func TestRunSkills_GenerateAll(t *testing.T) {
	dir, cleanup := setupSkillsProject(t)
	defer cleanup()

	code := RunSkills([]string{"generate-all"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	skillsDir := filepath.Join(dir, ".ptsd", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("cannot read skills dir: %v", err)
	}

	// generate-all should produce 13 standard skill files per BDD
	if len(entries) != 13 {
		t.Errorf("expected 13 skill files after generate-all, got %d", len(entries))
	}
}

// TestRunSkills_GenerateAll_ThenList verifies list shows all generated skills.
func TestRunSkills_GenerateAll_ThenList(t *testing.T) {
	_, cleanup := setupSkillsProject(t)
	defer cleanup()

	if code := RunSkills([]string{"generate-all"}, true); code != 0 {
		t.Fatalf("generate-all failed with code %d", code)
	}

	code := RunSkills([]string{"list"}, true)
	if code != 0 {
		t.Errorf("expected exit code 0 after generate-all, got %d", code)
	}
}

// TestRunSkills_WorkflowSkillContent verifies BDD "Workflow skill references all other skills":
// workflow.md must list the pipeline order and reference which skill to use at each stage.
func TestRunSkills_WorkflowSkillContent(t *testing.T) {
	dir, cleanup := setupSkillsProject(t)
	defer cleanup()

	if code := RunSkills([]string{"generate-all"}, true); code != 0 {
		t.Fatalf("generate-all failed with code %d", code)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".ptsd", "skills", "workflow.md"))
	if err != nil {
		t.Fatalf("workflow.md not found: %v", err)
	}
	content := string(data)

	// Must contain pipeline order
	for _, stage := range []string{"PRD", "Seed", "BDD", "Tests", "Impl"} {
		if !strings.Contains(content, stage) {
			t.Errorf("workflow.md missing pipeline stage %q", stage)
		}
	}

	// Must reference skills used at each stage
	for _, skill := range []string{"write-prd", "write-seed", "write-bdd", "write-tests", "write-impl"} {
		if !strings.Contains(content, skill) {
			t.Errorf("workflow.md does not reference skill %q", skill)
		}
	}
}

// TestRunSkills_List_HumanMode verifies human mode output for list command.
func TestRunSkills_List_HumanMode(t *testing.T) {
	dir, cleanup := setupSkillsProject(t)
	defer cleanup()

	// Pre-create a skill file
	skillsDir := filepath.Join(dir, ".ptsd", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}
	skillContent := "---\nname: bdd-my-feat\ndescription: test skill\ntrigger: test\n---\n\n## body\n"
	if err := os.WriteFile(filepath.Join(skillsDir, "bdd-my-feat.md"), []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunSkills([]string{"list"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "bdd-my-feat") {
		t.Errorf("human mode list missing skill id, got: %q", output)
	}
}

// TestRunSkills_List_Empty_HumanMode verifies human mode output when no skills exist.
func TestRunSkills_List_Empty_HumanMode(t *testing.T) {
	_, cleanup := setupSkillsProject(t)
	defer cleanup()

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunSkills([]string{"list"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0 for empty list, got %d", code)
	}
	if !strings.Contains(output, "no skills") {
		t.Errorf("human mode empty list missing 'no skills' message, got: %q", output)
	}
}

// TestRunSkills_Generate_HumanMode verifies human mode output for generate command.
func TestRunSkills_Generate_HumanMode(t *testing.T) {
	_, cleanup := setupSkillsProject(t)
	defer cleanup()

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunSkills([]string{"generate", "impl", "my-feat"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "impl") || !strings.Contains(output, "my-feat") {
		t.Errorf("human mode generate output missing stage/feature, got: %q", output)
	}
}

// TestRunSkills_GenerateAll_HumanMode verifies human mode output for generate-all command.
func TestRunSkills_GenerateAll_HumanMode(t *testing.T) {
	_, cleanup := setupSkillsProject(t)
	defer cleanup()

	var output string
	var code int
	output = captureStdout(t, func() {
		code = RunSkills([]string{"generate-all"}, false)
	})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(output, "all standard skills generated") {
		t.Errorf("human mode generate-all output unexpected, got: %q", output)
	}
}

// TestRunSkills_Generate_AllValidStages verifies that each pipeline stage is accepted.
func TestRunSkills_Generate_AllValidStages(t *testing.T) {
	stages := []string{"prd", "seed", "bdd", "tests", "impl"}
	for _, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			dir, cleanup := setupSkillsProject(t)
			defer cleanup()

			code := RunSkills([]string{"generate", stage, "feat-x"}, true)
			if code != 0 {
				t.Errorf("stage %s: expected exit code 0, got %d", stage, code)
			}

			skillPath := filepath.Join(dir, ".ptsd", "skills", stage+"-feat-x.md")
			if _, err := os.Stat(skillPath); os.IsNotExist(err) {
				t.Errorf("stage %s: skill file not created at %s", stage, skillPath)
			}
		})
	}
}
