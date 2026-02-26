@feature:skills
Feature: Skill Generation
  Generate pipeline skills into .ptsd/skills/ on init.

  Scenario: All skills generated on init
    When I run "ptsd init --name MyApp"
    Then 13 skill files exist in .ptsd/skills/
    And each has name, description, trigger fields in frontmatter

  Scenario: Skills follow universal format
    Given .ptsd/skills/write-bdd.md exists
    Then it starts with YAML frontmatter containing name and description
    And body contains structured instructions

  Scenario: Workflow skill references all other skills
    Given .ptsd/skills/workflow.md exists
    Then it lists the pipeline order
    And references which skill to use at each stage
