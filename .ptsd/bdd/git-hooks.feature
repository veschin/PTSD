@feature:git-hooks
Feature: Git Hook Enforcement
  Pre-commit hook validates commit message format and staged file classification.

  Scenario: Valid commit with matching scope
    Given staged files are only in .ptsd/bdd/
    And commit message is "[BDD] add: login scenarios"
    When pre-commit hook runs
    Then hook passes

  Scenario: Scope mismatch - impl files with BDD scope
    Given staged files include src/auth.ts and .ptsd/bdd/auth.feature
    And commit message is "[BDD] add: auth scenarios"
    When pre-commit hook runs
    Then hook fails
    And output contains "err:git staged files require [IMPL] scope"

  Scenario: Missing scope
    Given staged files include .ptsd/docs/PRD.md
    And commit message is "update PRD"
    When pre-commit hook runs
    Then hook fails
    And output contains "err:git missing [SCOPE] in commit message"

  Scenario: Invalid scope
    Given commit message is "[UNKNOWN] add: something"
    When pre-commit hook runs
    Then hook fails
    And output contains "err:git unknown scope UNKNOWN"

  Scenario: IMPL scope triggers full validation
    Given staged files include implementation code
    And commit message is "[IMPL] feat: user auth"
    When pre-commit hook runs
    Then ptsd validate is executed
    And any validation error blocks the commit

  Scenario: TASK and STATUS scopes skip pipeline validation
    Given staged files are only .ptsd/tasks.yaml
    And commit message is "[TASK] add: new auth task"
    When pre-commit hook runs
    Then hook passes without pipeline validation

  Scenario: File classification by path
    Given ptsd.yaml has testing.patterns.files ["tests/**/*.test.ts"]
    Then files in .ptsd/docs/ are classified as PRD
    And files in .ptsd/seeds/ are classified as SEED
    And files in .ptsd/bdd/ are classified as BDD
    And files matching tests/**/*.test.ts are classified as TEST
    And all other files are classified as IMPL
