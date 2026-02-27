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

// skillTemplates maps skill filename to frontmatter + body content.
var skillTemplates = map[string]struct {
	name        string
	description string
	trigger     string
	body        string
}{
	"write-prd.md": {
		name:        "write-prd",
		description: "Guide for writing a PRD section: structure, anchors, edge cases, non-goals",
		trigger:     "When creating or updating a PRD section for a feature",
		body: `## Instructions

1. Start with a one-line summary of the feature purpose.
2. Define the problem being solved and who it affects.
3. List acceptance criteria as testable statements.
4. Define non-goals explicitly — what is out of scope.
5. Cover edge cases: empty input, missing files, invalid state.
6. Add a feature anchor comment: <!-- feature:<id> -->
7. Keep language precise. No ambiguity.
`,
	},
	"write-seed.md": {
		name:        "write-seed",
		description: "Guide for creating golden seed data: realistic data, happy and edge cases, manifest",
		trigger:     "When creating .ptsd/seeds/<feature>/ directory and seed.yaml",
		body: `## Instructions

1. Create seed.yaml with feature field and files list.
2. Include at least one happy-path data file.
3. Include edge-case data: empty collections, boundary values, invalid inputs.
4. Use realistic data — not "test" or "foo".
5. Every file referenced in seed.yaml must exist on disk.
6. Formats: JSON, YAML, or CSV depending on what the feature consumes.
`,
	},
	"write-bdd.md": {
		name:        "write-bdd",
		description: "Guide for writing Gherkin BDD scenarios from PRD and seed data",
		trigger:     "When creating .ptsd/bdd/<feature>.feature files",
		body: `## Instructions

1. Write one scenario per acceptance criterion in the PRD.
2. Cover happy path, edge cases, and error paths.
3. Use seed data values in Given steps.
4. Each scenario must be independently runnable.
5. Use standard Gherkin: Given/When/Then. No And/But stacking.
6. Tag the feature: @feature:<id> at top of file.
`,
	},
	"write-tests.md": {
		name:        "write-tests",
		description: "Guide for writing tests from BDD: 1:1 mapping, no mocks, real assertions",
		trigger:     "When creating *_test.go files for a feature",
		body: `## Instructions

1. One test function per BDD scenario, named after the scenario.
2. Use real files in temp directories — no mocks for internal code.
3. Assert exact output values, not just "no error".
4. Test error paths: verify error message prefix (err:<category>).
5. Use t.TempDir() for isolation.
6. No test helpers that obscure what is being tested.
`,
	},
	"write-impl.md": {
		name:        "write-impl",
		description: "Guide for implementing from tests: no extra code, no over-engineering",
		trigger:     "When writing implementation to make failing tests pass",
		body: `## Instructions

1. Make each failing test pass, one at a time.
2. Write only the code required — no speculative features.
3. Follow the package structure: core/ for logic, cli/ for glue, render/ for TUI.
4. Error format: fmt.Errorf("err:<category> <message>").
5. No mocks in implementation. Use real I/O.
6. Run go test ./... after each change.
`,
	},
	"create-tasks.md": {
		name:        "create-tasks",
		description: "Guide for creating tasks: feature link, priority, clear checklist",
		trigger:     "When adding entries to .ptsd/tasks.yaml",
		body: `## Instructions

1. Every task must link to a feature via the feature field.
2. Use IDs in format T-<n>, incrementing from last existing ID.
3. Priority: A (urgent), B (normal), C (low).
4. Title must be a clear action: "Implement X", "Fix Y", "Add Z".
5. Add a checklist of subtasks if the task has multiple steps.
6. Status: TODO → IN-PROGRESS → DONE.
`,
	},
	"review-prd.md": {
		name:        "review-prd",
		description: "Guide for reviewing a PRD section: completeness, edge cases, non-goals, clarity",
		trigger:     "When reviewing a PRD section before advancing to seed stage",
		body: `## Review Checklist

Score 0-10 based on how many items pass.

- [ ] One-line summary present and accurate
- [ ] Problem statement is clear
- [ ] Acceptance criteria are testable
- [ ] Non-goals explicitly listed
- [ ] Edge cases covered
- [ ] No ambiguous language
- [ ] Feature anchor comment present

Output: score and list of specific issues found.
`,
	},
	"review-seed.md": {
		name:        "review-seed",
		description: "Guide for reviewing seed data: coverage, realism, happy and edge cases",
		trigger:     "When reviewing .ptsd/seeds/<feature>/ before advancing to bdd stage",
		body: `## Review Checklist

Score 0-10 based on how many items pass.

- [ ] seed.yaml has feature field
- [ ] At least one happy-path data file
- [ ] Edge case data present (empty, boundary, invalid)
- [ ] All files in manifest exist on disk
- [ ] Data is realistic (not placeholder values)
- [ ] File formats match what the feature consumes

Output: score and list of specific issues found.
`,
	},
	"review-bdd.md": {
		name:        "review-bdd",
		description: "Guide for reviewing BDD scenarios: match PRD, all paths, no gaps",
		trigger:     "When reviewing .ptsd/bdd/<feature>.feature before advancing to tests stage",
		body: `## Review Checklist

Score 0-10 based on how many items pass.

- [ ] One scenario per PRD acceptance criterion
- [ ] Happy path covered
- [ ] Error paths covered
- [ ] Edge cases from seed data used
- [ ] Each scenario independently runnable
- [ ] Gherkin syntax correct
- [ ] Feature tag present

Output: score and list of specific issues found.
`,
	},
	"review-tests.md": {
		name:        "review-tests",
		description: "Guide for reviewing tests: 1:1 BDD mapping, no mocks, real assertions",
		trigger:     "When reviewing *_test.go before advancing to impl stage",
		body: `## Review Checklist

Score 0-10 based on how many items pass.

- [ ] One test per BDD scenario
- [ ] Test names match scenarios
- [ ] Real files used, no mocks
- [ ] Error messages checked (err:<category> prefix)
- [ ] Assertions on actual values not just err==nil
- [ ] t.TempDir() used for isolation
- [ ] Tests pass independently

Output: score and list of specific issues found.
`,
	},
	"review-impl.md": {
		name:        "review-impl",
		description: "Guide for reviewing implementation: all tests pass, code matches design, no cheating",
		trigger:     "When reviewing implementation after all tests pass",
		body: `## Review Checklist

Score 0-10 based on how many items pass.

- [ ] All tests pass (go test ./...)
- [ ] No skipped or disabled tests
- [ ] No mock or stub patterns in implementation
- [ ] Error format: err:<category>
- [ ] No code outside the task scope
- [ ] Package boundaries respected (core/render/cli/yaml)
- [ ] No premature abstractions

Output: score and list of specific issues found.
`,
	},
	"adopt.md": {
		name:        "adopt",
		description: "Guide for bootstrapping an existing project into PTSD pipeline",
		trigger:     "When running ptsd adopt on an existing codebase",
		body: `## Instructions

1. Run ptsd adopt --name <name> in the project root.
2. PTSD creates .ptsd/ with config, features.yaml, and empty state.
3. Register existing features with realistic status values.
4. For each feature, assess current stage: which pipeline steps are complete.
5. Create seed data from existing tests or documentation.
6. Write BDD scenarios to capture existing behavior.
7. Do not rewrite working code — document and track it.
`,
	},
	"workflow.md": {
		name:        "workflow",
		description: "Full pipeline, mandatory order, which skill to use at each stage",
		trigger:     "At session start or when unsure what to do next",
		body: `## Pipeline Order (mandatory, no skipping)

PRD → Seed → BDD → Tests → Implementation

### At each stage

| Stage | Create skill | Review skill |
|-------|-------------|--------------|
| PRD | write-prd.md | review-prd.md |
| Seed | write-seed.md | review-seed.md |
| BDD | write-bdd.md | review-bdd.md |
| Tests | write-tests.md | review-tests.md |
| Impl | write-impl.md | review-impl.md |

### Session protocol

1. Read .ptsd/review-status.yaml — find where each feature is.
2. Pick the next feature/stage with review: pending or tests: absent.
3. Apply the appropriate skill from the table above.
4. Record progress immediately in review-status.yaml after each action.
5. Run ptsd validate --agent before committing.
6. Commit with [SCOPE] type: message format.

### Gate rules

- No BDD without seed initialized
- No tests without BDD written
- No impl without passing test review
- No stage advance without review score >= min_score (default 7)
`,
	},
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

	content := buildSkillContent(stage+"-"+featureID, "Skill for "+stage+" stage of feature "+featureID, "When working on "+stage+" stage of "+featureID, "## Instructions\n\nFollow the PTSD pipeline for the "+stage+" stage.\n")

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

	for filename, tmpl := range skillTemplates {
		content := buildSkillContent(tmpl.name, tmpl.description, tmpl.trigger, tmpl.body)
		skillPath := filepath.Join(skillsDir, filename)
		if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("err:io %w", err)
		}
	}

	return nil
}

// buildSkillContent formats a skill file with YAML frontmatter and body.
func buildSkillContent(name, description, trigger, body string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString("name: " + name + "\n")
	sb.WriteString("description: " + description + "\n")
	sb.WriteString("trigger: " + trigger + "\n")
	sb.WriteString("---\n\n")
	sb.WriteString(body)
	return sb.String()
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
