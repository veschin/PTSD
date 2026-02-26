@feature:seed-mgmt
Feature: Seed Management
  Manage golden seed data per feature.

  Scenario: Init seed for feature
    Given feature "user-auth" exists
    When I run "ptsd seed init user-auth"
    Then .ptsd/seeds/user-auth/seed.yaml is created with feature field

  Scenario: Init seed for nonexistent feature
    When I run "ptsd seed init ghost"
    Then exit code is 1
    And output contains "err:validation feature ghost not found"

  Scenario: Add file to seed
    Given seed initialized for "user-auth"
    And file user.json exists
    When I run "ptsd seed add user-auth user.json --type data"
    Then user.json is copied to .ptsd/seeds/user-auth/
    And seed.yaml files list includes user.json with type data

  Scenario: Check seeds
    Given features "user-auth" (has seed) and "catalog" (no seed) exist as in-progress
    When I run "ptsd seed check --agent"
    Then output contains "err:pipeline catalog has no seed"
    And exit code is 1

  Scenario: Planned features skip seed check
    Given feature "future" has status planned and no seed
    When I run "ptsd seed check --agent"
    Then "future" is not mentioned in output
