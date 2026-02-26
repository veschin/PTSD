@feature:test-integration
Feature: Test Integration
  Map BDD scenarios to test files, discover tests, run via configured runner.

  Scenario: Map BDD to test file
    Given .ptsd/bdd/user-auth.feature and tests/auth.test.ts exist
    When I run "ptsd test map .ptsd/bdd/user-auth.feature tests/auth.test.ts"
    Then mapping is stored in state

  Scenario: Check test coverage
    Given "user-auth" has 3 BDD scenarios and 2 mapped tests
    When I run "ptsd test check --agent"
    Then output shows "user-auth: 2/3 partial"

  Scenario: Run tests for feature
    Given tests mapped for "user-auth"
    When I run "ptsd test run --feature user-auth --agent"
    Then configured test runner executes
    And output is "pass:N fail:N" with failing file:line paths

  Scenario: Run all tests
    When I run "ptsd test run --agent"
    Then all mapped tests are executed

  Scenario: Test results update state
    Given test run completes with 5 pass 1 fail
    Then state.yaml records test results for the feature with timestamp

  Scenario: No test runner configured
    Given ptsd.yaml has no testing.runner
    When I run "ptsd test run --agent"
    Then exit code is 3
    And output contains "err:config no test runner configured"
