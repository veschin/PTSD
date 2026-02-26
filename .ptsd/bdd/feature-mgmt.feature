@feature:feature-mgmt
Feature: Feature Management
  Register, list, show, update features in .ptsd/features.yaml.

  Scenario: Add feature
    Given an initialized ptsd project
    When I run "ptsd feature add user-auth --title 'User Authentication'"
    Then features.yaml contains feature with id "user-auth"
    And feature status defaults to "planned"

  Scenario: Add duplicate feature
    Given feature "user-auth" already exists
    When I run "ptsd feature add user-auth"
    Then exit code is 1
    And output contains "err:validation feature user-auth already exists"

  Scenario: List features
    Given features "user-auth", "catalog", "payments" exist
    When I run "ptsd feature list --agent"
    Then output shows all three features with status

  Scenario: List features filtered by status
    Given "user-auth" is in-progress and "catalog" is planned
    When I run "ptsd feature list --status planned --agent"
    Then output shows only "catalog"

  Scenario: Show feature details
    Given feature "user-auth" exists with PRD anchor, seed, 3 BDD scenarios, 2 tests
    When I run "ptsd feature show user-auth --agent"
    Then output contains id, status, PRD line range, seed status, scenario count, test count, scores

  Scenario: Update feature status
    Given feature "user-auth" exists as "planned"
    When I run "ptsd feature status user-auth in-progress"
    Then feature status is "in-progress"

  Scenario: Set implemented requires passing tests
    Given feature "user-auth" has failing tests
    When I run "ptsd feature status user-auth implemented"
    Then exit code is 1
    And output contains "err:pipeline tests not passing"

  Scenario: Remove feature
    Given feature "user-auth" exists
    When I run "ptsd feature remove user-auth"
    Then features.yaml no longer contains "user-auth"
