@feature:bdd-mgmt
Feature: BDD Management
  Manage Gherkin .feature files linked to features.

  Scenario: Add BDD file for feature with seed
    Given feature "user-auth" exists and has seed data
    When I run "ptsd bdd add user-auth"
    Then .ptsd/bdd/user-auth.feature is created with @feature:user-auth tag

  Scenario: Add BDD file without seed refuses
    Given feature "user-auth" exists but has no seed
    When I run "ptsd bdd add user-auth"
    Then exit code is 1
    And output contains "err:pipeline user-auth has no seed"

  Scenario: Check BDD coverage
    Given features "user-auth" (has BDD) and "catalog" (no BDD) are in-progress
    When I run "ptsd bdd check --agent"
    Then output contains "err:pipeline catalog has no bdd"

  Scenario: Show BDD scenarios compact
    Given user-auth.feature has 3 scenarios
    When I run "ptsd bdd show user-auth --agent"
    Then each scenario is one line: "Title: Given X / When Y / Then Z"

  Scenario: BDD file references nonexistent feature
    Given .ptsd/bdd/ghost.feature has @feature:ghost but ghost is not in registry
    When I run "ptsd bdd check --agent"
    Then output contains "err:validation unknown feature tag ghost"
