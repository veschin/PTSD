@feature:adopt
Feature: Existing Project Bootstrap
  Scan existing project, discover artifacts, create .ptsd/ structure.

  Scenario: Adopt discovers BDD files
    Given project has bdd/*.feature files with @feature tags
    When I run "ptsd adopt"
    Then features are extracted from tags and added to registry
    And .feature files are moved to .ptsd/bdd/

  Scenario: Adopt discovers test files
    Given ptsd.yaml has testing.patterns configured
    When I run "ptsd adopt"
    Then test files are discovered and mapped

  Scenario: Dry run shows what would happen
    When I run "ptsd adopt --dry-run"
    Then discovered artifacts are listed
    And no files are created or moved

  Scenario: Adopt refuses if .ptsd already initialized
    Given .ptsd/ already exists with features.yaml
    When I run "ptsd adopt"
    Then exit code is 1
    And output contains "err:validation already initialized"
