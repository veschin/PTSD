@feature:state-tracking
Feature: State & Regression Detection
  Track file hashes per feature, detect regressions when artifacts change.

  Scenario: Store hashes on state update
    Given feature "user-auth" has seed, BDD, and test files
    When state is synced
    Then state.yaml contains SHA256 hashes for each file

  Scenario: BDD change detected for implemented feature
    Given "user-auth" is at stage "implemented"
    And user-auth.feature hash changed
    When state-reading command runs
    Then feature stage is downgraded
    And output contains regression warning

  Scenario: PRD change creates redo tasks
    Given "user-auth" is at stage "bdd"
    And PRD section hash for user-auth changed
    When state-reading command runs
    Then feature stage is downgraded to "prd"
    And a redo task is auto-created

  Scenario: Seed change warning
    Given "user-auth" is at stage "test"
    And seed file hash changed
    When state-reading command runs
    Then output contains warning about stale downstream artifacts

  Scenario: No regression on expected change
    Given "user-auth" is at stage "bdd"
    And user-auth.feature hash changed
    When state-reading command runs
    Then hash is updated silently (change expected at this stage)

  Scenario: Store review scores
    Given LLM reviews "user-auth" PRD with score 8
    When review is recorded
    Then state.yaml contains prd score 8 with timestamp
