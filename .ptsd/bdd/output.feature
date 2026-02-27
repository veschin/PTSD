@feature:output
Feature: Dual Render Mode
  All output via Bubbletea. --agent flag switches to compact mode.

  Scenario: Agent flag produces compact output
    Given an initialized project
    When any command is run with --agent
    Then output has zero decoration, no colors, no borders

  Scenario: Default mode launches TUI
    Given an initialized project
    When "ptsd status" is run without --agent
    Then interactive TUI renders with colors and navigation

  Scenario: Agent error format
    Given a validation error occurs
    When running with --agent
    Then output is "err:<category> <message>" on a single line

  Scenario: PTSD_AGENT env var
    Given PTSD_AGENT=1 is set
    When any command is run without --agent flag
    Then agent mode is used automatically

  Scenario: Agent output includes file coordinates
    Given feature "user-auth" has PRD at lines 30-40
    When I run "ptsd task next --agent"
    Then output includes "PRD:l30-40"

  # CLI Entry Point Routing
  Scenario: Unknown subcommand shows error
    Given an initialized project
    When I run "ptsd unknown-cmd"
    Then exit code is 2
    And output contains "err:user unknown command"

  Scenario: --agent flag passed to handlers
    Given an initialized project
    When I run "ptsd config show --agent"
    Then agent mode output format is used

  Scenario: All subcommands route correctly
    Given an initialized project
    When I run "ptsd status"
    Then it executes RunStatus handler

  Scenario: Missing subcommand shows usage
    Given an initialized project
    When I run "ptsd"
    Then exit code is 2
    And output contains "usage:"
