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

  Scenario: Init refuses if .ptsd already exists
    Given a directory with existing .ptsd/
    When I run "ptsd init"
    Then exit code is 1
    And output contains "err:validation .ptsd already exists"

  Scenario: Init generates skills
    Given an empty directory with git initialized
    When I run "ptsd init --name MyApp"
    Then .ptsd/skills/ contains write-prd.md, write-seed.md, write-bdd.md, write-tests.md, write-impl.md
    And .ptsd/skills/ contains review-prd.md, review-seed.md, review-bdd.md, review-tests.md, review-impl.md
    And .ptsd/skills/ contains create-tasks.md, workflow.md
