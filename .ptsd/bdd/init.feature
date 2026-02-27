@feature:init
Feature: Project Initialization
  ptsd init scaffolds .ptsd/ structure and generates agent instructions.

  Scenario: Init new project
    Given an empty directory with git initialized
    When I run "ptsd init --name MyApp"
    Then .ptsd/ directory is created with subdirectories: docs, seeds, bdd, skills
    And .ptsd/ptsd.yaml exists with project name "MyApp"
    And .ptsd/features.yaml exists and is empty
    And .ptsd/state.yaml exists and is empty
    And .ptsd/tasks.yaml exists and is empty
    And .ptsd/docs/PRD.md exists with template content
    And CLAUDE.md is generated at project root with workflow instructions
    And .git/hooks/pre-commit is installed

  Scenario: Init detects test runner
    Given a directory with package.json containing vitest
    When I run "ptsd init"
    Then .ptsd/ptsd.yaml testing.runner equals "npx vitest run"

  Scenario: Init refuses without git
    Given a directory without .git
    When I run "ptsd init"
    Then exit code is 3
    And output contains "err:config git repository required"

  Scenario: Re-init existing project regenerates hooks and skills
    Given a directory with existing .ptsd/ from previous init
    When I run "ptsd init"
    Then exit code is 0
    And .git/hooks/pre-commit is regenerated
    And .git/hooks/commit-msg is regenerated
    And .claude/hooks/ scripts are regenerated
    And .claude/settings.json is regenerated
    And .ptsd/skills/ files are regenerated

  Scenario: Re-init preserves .ptsd/ data files
    Given a directory with existing .ptsd/ containing custom data
    When I run "ptsd init"
    Then .ptsd/features.yaml is unchanged
    And .ptsd/state.yaml is unchanged
    And .ptsd/tasks.yaml is unchanged
    And .ptsd/review-status.yaml is unchanged
    And .ptsd/docs/PRD.md is unchanged
    And .ptsd/ptsd.yaml is unchanged

  Scenario: Re-init updates CLAUDE.md ptsd section only
    Given a directory with CLAUDE.md containing ptsd markers and user content
    When I run "ptsd init"
    Then content between <!-- ---ptsd--- --> markers is replaced with latest template
    And user content outside markers is preserved

  Scenario: Re-init creates CLAUDE.md section if markers absent
    Given a directory with CLAUDE.md that has no ptsd markers
    When I run "ptsd init"
    Then <!-- ---ptsd--- --> markers with template content are appended to CLAUDE.md
    And existing file content is preserved above markers

  Scenario: Re-init on project without CLAUDE.md creates file with markers
    Given a directory with .ptsd/ but no CLAUDE.md
    When I run "ptsd init"
    Then CLAUDE.md is created with <!-- ---ptsd--- --> markers wrapping template content

  Scenario: Init generates skills
    Given an empty directory with git initialized
    When I run "ptsd init --name MyApp"
    Then .ptsd/skills/ contains write-prd.md, write-seed.md, write-bdd.md, write-tests.md, write-impl.md
    And .ptsd/skills/ contains review-prd.md, review-seed.md, review-bdd.md, review-tests.md, review-impl.md
    And .ptsd/skills/ contains create-tasks.md, workflow.md
