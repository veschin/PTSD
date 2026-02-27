package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Skill represents a pipeline skill file.
type Skill struct {
	ID      string
	Stage   string
	Feature string
	Path    string
}

// validSkillStages lists all pipeline stages a skill can be associated with.
var validSkillStages = map[string]bool{
	"prd":   true,
	"seed":  true,
	"bdd":   true,
	"tests": true,
	"impl":  true,
}

// standardSkillFiles lists all standard pipeline skill templates to be embedded.
var standardSkillFiles = []string{
	"write-prd.md", "write-seed.md", "write-bdd.md", "write-tests.md",
	"write-impl.md", "create-tasks.md", "review-prd.md", "review-seed.md",
	"review-bdd.md", "review-tests.md", "review-impl.md", "adopt.md", "workflow.md",
}

// GenerateSkill generates a single skill file for the given stage and feature.
// Filename format: <stage>-<feature>.md
// projectDir is the root directory containing .ptsd/.
func GenerateSkill(projectDir, stage, featureID string) error {
	if !validSkillStages[stage] {
		return fmt.Errorf("err:user invalid stage %q: must be prd|seed|bdd|tests|impl", stage)
	}

	skillsDir := filepath.Join(projectDir, ".ptsd", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	filename := stage + "-" + featureID + ".md"
	skillPath := filepath.Join(skillsDir, filename)

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("name: " + stage + "-" + featureID + "\n")
	sb.WriteString("description: Use when working on " + stage + " stage of " + featureID + "\n")
	sb.WriteString("---\n\n")
	sb.WriteString("## Instructions\n\nFollow the PTSD pipeline for the " + stage + " stage.\n")
	content := sb.String()

	if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	return nil
}

// GenerateAllSkills generates all standard pipeline skills into .ptsd/skills/.
// This is called by ptsd init.
func GenerateAllSkills(projectDir string) error {
	skillsDir := filepath.Join(projectDir, ".ptsd", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("err:io %w", err)
	}

	for _, filename := range standardSkillFiles {
		content, err := readTemplate("templates/skills/" + filename)
		if err != nil {
			return fmt.Errorf("err:io %w", err)
		}
		skillPath := filepath.Join(skillsDir, filename)
		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("err:io %w", err)
		}
	}

	return nil
}

// generateClaudeSkills generates .claude/skills/<name>/SKILL.md for each standard skill.
// This enables Claude Code auto-discovery of skills.
func generateClaudeSkills(dir string) error {
	for _, filename := range standardSkillFiles {
		name := strings.TrimSuffix(filename, ".md")
		skillDir := filepath.Join(dir, ".claude", "skills", name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("err:io %w", err)
		}
		content, err := readTemplate("templates/skills/" + filename)
		if err != nil {
			return fmt.Errorf("err:io %w", err)
		}
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644); err != nil {
			return fmt.Errorf("err:io %w", err)
		}
	}
	return nil
}

// ListSkills returns all skill files found in .ptsd/skills/.
func ListSkills(projectDir string) ([]Skill, error) {
	skillsDir := filepath.Join(projectDir, ".ptsd", "skills")

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("err:io %w", err)
	}

	var skills []Skill
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		skill := parseSkillFilename(entry.Name(), filepath.Join(skillsDir, entry.Name()))
		skills = append(skills, skill)
	}

	return skills, nil
}

// parseSkillFilename derives Skill fields from the filename.
// Expected formats: <stage>-<feature>.md or <name>.md (for standalone skills like workflow.md).
func parseSkillFilename(filename, fullPath string) Skill {
	name := strings.TrimSuffix(filename, ".md")

	// Check if it matches a known stage prefix: prd-, seed-, bdd-, tests-, impl-
	for stage := range validSkillStages {
		prefix := stage + "-"
		if strings.HasPrefix(name, prefix) {
			featureID := strings.TrimPrefix(name, prefix)
			return Skill{
				ID:      name,
				Stage:   stage,
				Feature: featureID,
				Path:    fullPath,
			}
		}
	}

	// Standalone skill (e.g. workflow, write-prd, review-bdd)
	return Skill{
		ID:      name,
		Stage:   "",
		Feature: "",
		Path:    fullPath,
	}
}
