@feature:prd-check
Feature: PRD Validation
  Validate PRD anchors against feature registry.

  Scenario: All features have anchors
    Given all registered features have <!-- feature:id --> anchors in PRD
    When I run "ptsd prd check --agent"
    Then exit code is 0

  Scenario: Missing anchor
    Given feature "user-auth" has no anchor in PRD
    When I run "ptsd prd check --agent"
    Then exit code is 1
    And output contains "err:pipeline user-auth has no prd anchor"

  Scenario: Orphaned anchor
    Given PRD contains <!-- feature:ghost --> but ghost is not in registry
    When I run "ptsd prd check --agent"
    Then output contains "err:pipeline orphaned anchor ghost"

  Scenario: Extract PRD section for feature
    Given PRD has <!-- feature:user-auth --> followed by content until next anchor
    When prd section is extracted for "user-auth"
    Then the content between anchors is returned with line numbers
