@feature:review
Feature: Quality Scoring
  LLM reviews each pipeline stage, ptsd stores scores.

  Scenario: Record review score
    Given feature "user-auth" exists
    When I run "ptsd review user-auth --stage prd --score 8"
    Then state.yaml contains prd score 8 for user-auth with timestamp

  Scenario: Score below threshold blocks progression
    Given review.min_score is 7
    And "user-auth" prd score is 5
    When feature tries to advance past prd stage
    Then advancement is refused
    And output contains "err:pipeline score 5 below threshold 7"

  Scenario: Score above threshold allows progression
    Given review.min_score is 7
    And "user-auth" prd score is 8
    When feature tries to advance past prd stage
    Then advancement is allowed

  Scenario: Auto-redo task on low score
    Given review.auto_redo is true
    And review score is below threshold
    When review is recorded
    Then a redo task is auto-created for the feature and stage
