@feature:common-issues
Feature: Common Issues Registry
  Maintain a compressed knowledge base of recurring problems in .ptsd/issues.yaml.

  Scenario: Add a new issue
    Given no issues exist
    When I run "ptsd issues add --category env 'venv uses wrong python' --fix 'rm -rf .venv && python3.12 -m venv .venv'"
    Then issues.yaml contains issue with category "env" and summary "venv uses wrong python"

  Scenario: List all issues
    Given issues.yaml contains two issues
    When I run "ptsd issues list --agent"
    Then output shows both issues with id, category, summary, and fix

  Scenario: List issues filtered by category
    Given issues with categories "env" and "access" exist
    When I run "ptsd issues list --category env --agent"
    Then output shows only env issues

  Scenario: Remove a resolved issue
    Given issue "venv-wrong-python" exists
    When I run "ptsd issues remove venv-wrong-python"
    Then issues.yaml does not contain "venv-wrong-python"

  Scenario: Prevent duplicate issue IDs
    Given issue "missing-api-key" already exists
    When I run "ptsd issues add --category access 'OPENAI_API_KEY not set' --fix 'cp .env.example .env'"
    Then exit code is 1
    And output contains "err:validation issue missing-api-key already exists"

  Scenario: Reject invalid category
    When I run "ptsd issues add --category typo 'some problem' --fix 'fix it'"
    Then exit code is 2
    And output contains "err:validation invalid category"

  Scenario: Reject missing summary
    When I run "ptsd issues add --category env '' --fix 'fix it'"
    Then exit code is 2
    And output contains "err:validation summary required"

  Scenario: Reject missing fix
    When I run "ptsd issues add --category env 'some problem' --fix ''"
    Then exit code is 2
    And output contains "err:validation fix required"

  Scenario: Load from non-existent file returns empty list
    Given no issues.yaml exists
    When I call ListIssues
    Then result is empty list with no error

  Scenario: Remove non-existent issue
    Given no issues exist
    When I run "ptsd issues remove ghost-id"
    Then exit code is 1
    And output contains "err:validation issue ghost-id not found"
