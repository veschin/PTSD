@feature:status
Feature: Project Status
  Show project overview and per-feature status.

  Scenario: Project status in agent mode
    Given 5 features, 3 with BDD, 2 with tests, 20 tasks
    When I run "ptsd status --agent"
    Then output is "[FEAT:5 FAIL:0] [BDD:3 FAIL:0] [TESTS:2 FAIL:0] [T:20 WIP:0 TODO:19 DONE:1]"

  Scenario: Feature status in agent mode
    Given feature "user-auth" at BDD stage with 3 scenarios and 2 tests
    When I run "ptsd status --feature user-auth --agent"
    Then output shows "user-auth [in-progress] PRD:l30-40 SEED:ok BDD:3scn TEST:2/3 SCORE:prd=8,seed=9,bdd=7"

  Scenario: Status with failing tests
    Given feature "catalog" has 1 failing test
    When I run "ptsd status --agent"
    Then FAIL count reflects the failure

  Scenario: Status in human mode
    Given an initialized project with features
    When I run "ptsd status"
    Then interactive TUI dashboard is displayed

  Scenario: Empty project status
    Given no features registered
    When I run "ptsd status --agent"
    Then output is "[FEAT:0 FAIL:0] [BDD:0 FAIL:0] [TESTS:0 FAIL:0] [T:0 WIP:0 TODO:0 DONE:0]"
