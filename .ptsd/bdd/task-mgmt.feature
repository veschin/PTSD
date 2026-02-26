@feature:task-mgmt
Feature: Task Management
  Create, list, update tasks linked to features.

  Scenario: Add task
    Given feature "user-auth" exists
    When I run "ptsd task add --feature user-auth 'Implement login endpoint' --priority A"
    Then tasks.yaml contains task T-1 linked to "user-auth" with priority A and status TODO

  Scenario: Add task without feature
    When I run "ptsd task add 'Do something'"
    Then exit code is 2
    And output contains "err:user --feature required"

  Scenario: Add task with nonexistent feature
    When I run "ptsd task add --feature ghost 'Do something'"
    Then exit code is 1
    And output contains "err:validation feature ghost not found"

  Scenario: Task next returns highest priority unblocked task
    Given tasks T-1 [TODO A], T-2 [TODO B], T-3 [WIP A] exist
    When I run "ptsd task next --agent"
    Then output is "T-1 [TODO] [A] [PRD:l30-40 BDD:l30-100]: Implement login endpoint"

  Scenario: Task next with limit
    Given 5 TODO tasks exist
    When I run "ptsd task next --agent --limit 3"
    Then output shows exactly 3 tasks

  Scenario: Task next when all done
    Given all tasks are DONE
    When I run "ptsd task next --agent"
    Then output is empty
    And exit code is 0

  Scenario: Update task status
    Given task T-1 exists with status TODO
    When I run "ptsd task update T-1 --status WIP"
    Then task T-1 status is WIP

  Scenario: List tasks filtered
    Given tasks linked to "user-auth" and "catalog"
    When I run "ptsd task list --feature user-auth --agent"
    Then output shows only user-auth tasks

  Scenario: Auto-increment task IDs
    Given tasks T-1 and T-2 exist
    When I run "ptsd task add --feature catalog 'New task'"
    Then new task has id T-3
