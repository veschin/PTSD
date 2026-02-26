@feature:validate
Feature: Pipeline Validation
  Validate structural integrity of all cross-references and pipeline ordering.

  Scenario: Clean project passes validation
    Given all features have PRD anchors, seeds, BDD, tests, passing results
    When I run "ptsd validate --agent"
    Then exit code is 0

  Scenario: Feature without PRD anchor
    Given feature "user-auth" exists but has no PRD anchor
    When I run "ptsd validate --agent"
    Then exit code is 1
    And output contains "err:pipeline user-auth has no prd anchor"

  Scenario: Feature with BDD but no seed
    Given feature "user-auth" has BDD scenarios but no seed
    When I run "ptsd validate --agent"
    Then exit code is 1
    And output contains "err:pipeline user-auth has bdd but no seed"

  Scenario: Feature with BDD but no tests
    Given feature "user-auth" has BDD scenarios but no mapped tests
    When I run "ptsd validate --agent"
    Then exit code is 1
    And output contains "err:pipeline user-auth has bdd but no tests"

  Scenario: Planned and deferred features are skipped
    Given feature "future" has status "planned"
    And feature "old" has status "deferred"
    When I run "ptsd validate --agent"
    Then neither "future" nor "old" appear in validation output

  Scenario: Mock patterns detected in test files
    Given test file contains "vi.mock" or "jest.mock" or "unittest.mock"
    When I run "ptsd validate --agent"
    Then exit code is 1
    And output contains "err:pipeline mock detected"

  Scenario: Multiple errors reported
    Given 3 features have pipeline violations
    When I run "ptsd validate --agent"
    Then all 3 errors are reported
    And exit code is 1
