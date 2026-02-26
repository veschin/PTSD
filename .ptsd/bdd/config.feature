@feature:config
Feature: Configuration System
  Load and validate ptsd.yaml. Walk up from CWD to find .ptsd/.

  Scenario: Load config from current directory
    Given .ptsd/ptsd.yaml exists with project name "MyApp"
    When config is loaded
    Then config.project.name equals "MyApp"

  Scenario: Walk up to find .ptsd
    Given .ptsd/ptsd.yaml exists two directories above CWD
    When config is loaded from CWD
    Then config is found and loaded correctly

  Scenario: Missing config
    Given no .ptsd/ exists in any parent directory
    When config is loaded
    Then exit code is 3
    And output contains "err:config ptsd.yaml not found"

  Scenario: Defaults fill missing sections
    Given .ptsd/ptsd.yaml exists with only project.name
    When config is loaded
    Then testing.patterns.files has default value
    And review.min_score defaults to 7
    And hooks.pre_commit defaults to true

  Scenario: Invalid YAML
    Given .ptsd/ptsd.yaml contains invalid YAML
    When config is loaded
    Then exit code is 3
    And output contains "err:config"
