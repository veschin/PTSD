@feature:git-hooks
Feature: Git Hook Enforcement
  Pre-commit hook validates commit message format and staged file classification.

  Scenario: Valid commit with matching scope
    Given staged files are only in .ptsd/bdd/
    And commit message is "[BDD] add: login scenarios"
    When pre-commit hook runs
    Then hook passes

  Scenario: Scope mismatch - impl files with BDD scope
    Given staged files include src/auth.ts and .ptsd/bdd/auth.feature
    And commit message is "[BDD] add: auth scenarios"
    When pre-commit hook runs
    Then hook fails
    And output contains "err:git staged files require [IMPL] scope"

  Scenario: Missing scope
    Given staged files include .ptsd/docs/PRD.md
    And commit message is "update PRD"
    When pre-commit hook runs
    Then hook fails
    And output contains "err:git missing [SCOPE] in commit message"

  Scenario: Invalid scope
    Given commit message is "[UNKNOWN] add: something"
    When pre-commit hook runs
    Then hook fails
    And output contains "err:git unknown scope UNKNOWN"

  Scenario: IMPL scope triggers full validation
    Given staged files include implementation code
    And commit message is "[IMPL] feat: user auth"
    When pre-commit hook runs
    Then ptsd validate is executed
    And any validation error blocks the commit

  Scenario: TASK and STATUS scopes skip pipeline validation
    Given staged files are only .ptsd/tasks.yaml
    And commit message is "[TASK] add: new auth task"
    When pre-commit hook runs
    Then hook passes without pipeline validation

  Scenario: File classification by path
    Given ptsd.yaml has testing.patterns.files ["tests/**/*.test.ts"]
    Then files in .ptsd/docs/ are classified as PRD
    And files in .ptsd/seeds/ are classified as SEED
    And files in .ptsd/bdd/ are classified as BDD
    And files matching tests/**/*.test.ts are classified as TEST
    And all other files are classified as IMPL

  # --- Claude Code Hook Integration ---

  # GateCheck (PreToolUse)
  Scenario: Gate blocks BDD write without seed
    Given feature "auth" exists with status "in-progress"
    And no seed exists for "auth"
    When gate-check runs for ".ptsd/bdd/auth.feature"
    Then write is blocked
    And reason contains "no seed for auth"

  Scenario: Gate allows BDD write with seed
    Given feature "auth" exists with status "in-progress"
    And seed exists for "auth"
    When gate-check runs for ".ptsd/bdd/auth.feature"
    Then write is allowed

  Scenario: Gate blocks seed write without PRD anchor
    Given feature "auth" exists with status "in-progress"
    And PRD has no anchor for "auth"
    When gate-check runs for ".ptsd/seeds/auth/seed.yaml"
    Then write is blocked
    And reason contains "no PRD anchor for auth"

  Scenario: Gate blocks test write without BDD
    Given feature "auth" exists with status "in-progress"
    And no BDD exists for "auth"
    When gate-check runs for "internal/core/auth_test.go"
    Then write is blocked
    And reason contains "no BDD scenarios for auth"

  Scenario: Gate blocks impl write without tests
    Given feature "auth" exists with status "in-progress"
    And no tests exist for "auth"
    When gate-check runs for "internal/core/auth.go"
    Then write is blocked
    And reason contains "no tests for auth"

  Scenario: Gate allows always-allowed files regardless of state
    Given any project state
    When gate-check runs for ".ptsd/docs/PRD.md"
    Then write is allowed
    When gate-check runs for ".ptsd/tasks.yaml"
    Then write is allowed
    When gate-check runs for "CLAUDE.md"
    Then write is allowed
    When gate-check runs for ".ptsd/issues.yaml"
    Then write is allowed

  Scenario: Gate allows .claude/hooks/ files
    Given any project state
    When gate-check runs for ".claude/hooks/ptsd-context.sh"
    Then write is allowed

  Scenario: Gate allows .ptsd/skills/ files
    Given any project state
    When gate-check runs for ".ptsd/skills/custom-skill.md"
    Then write is allowed

  Scenario: Gate allows unknown non-code files
    Given any project state
    When gate-check runs for "README.md"
    Then write is allowed

  Scenario: Gate handles absolute paths
    Given feature "auth" exists with always-allowed file
    When gate-check runs for "/abs/path/to/project/.ptsd/docs/PRD.md"
    Then write is allowed

  Scenario: Gate handles feature ID substring collision
    Given features "auth" and "authorization" exist
    And BDD exists for "auth" but not for "authorization"
    When gate-check runs for "internal/core/authorization_test.go"
    Then write is blocked for feature "authorization" not "auth"

  Scenario: Gate handles deeply nested seed file
    Given feature "auth" exists with PRD anchor
    When gate-check runs for ".ptsd/seeds/auth/data/nested/fixture.json"
    Then feature is identified as "auth"
    And write is allowed

  # extractFilePathFromStdin edge cases
  Scenario: Stdin JSON with no file_path key
    Given stdin contains '{"tool": "Bash", "command": "ls"}'
    When pre-tool-use hook runs
    Then hook allows (exit 0)

  Scenario: Stdin is empty
    Given stdin is empty
    When pre-tool-use hook runs
    Then hook allows (exit 0)

  Scenario: Stdin JSON with file_path in content value before key
    Given stdin contains content field mentioning "file_path" before actual file_path key
    When pre-tool-use hook runs
    Then hook extracts the correct file_path from the JSON key, not from content

  Scenario: Stdin JSON file_path with escaped quotes
    Given stdin contains '"file_path": "path/with\\"quote"'
    When pre-tool-use hook runs
    Then hook extracts "path/with" (truncated at escaped quote)

  # AutoTrack (PostToolUse)
  Scenario: Track advances stage from prd to seed
    Given feature "auth" at stage "prd"
    When file ".ptsd/seeds/auth/seed.yaml" is written
    Then stage advances to "seed"

  Scenario: Track advances stage from seed to bdd
    Given feature "auth" at stage "seed"
    When file ".ptsd/bdd/auth.feature" is written
    Then stage advances to "bdd"

  Scenario: Track sets tests to written
    Given feature "auth" at stage "bdd" with tests "absent"
    When file "internal/core/auth_test.go" is written
    Then tests field becomes "written"
    And stage advances to "tests"

  Scenario: Track never regresses stage
    Given feature "auth" at stage "impl"
    When file ".ptsd/bdd/auth.feature" is written
    Then stage remains "impl"

  Scenario: Track creates entry for unknown feature
    Given review-status has no entry for "auth"
    And feature "auth" is registered
    When file ".ptsd/bdd/auth.feature" is written
    Then new entry is created with stage "bdd"

  Scenario: Track ignores PRD file writes
    Given feature "auth" at stage "prd"
    When file ".ptsd/docs/PRD.md" is written
    Then no tracking occurs (result is nil)

  Scenario: Track ignores non-feature files
    Given any project state
    When file "README.md" is written
    Then no tracking occurs (result is nil)

  Scenario: Track is idempotent for same stage
    Given feature "auth" at stage "bdd"
    When file ".ptsd/bdd/auth.feature" is written twice
    Then no update occurs on second write

  # Context (SessionStart)
  Scenario: Context shows next action per stage
    Given feature "auth" at stage "prd" with review "passed"
    Then context shows "next: auth action=write-seed"
    Given feature "auth" at stage "seed" with review "passed"
    Then context shows "next: auth action=write-bdd"
    Given feature "auth" at stage "bdd" with review "passed"
    Then context shows "next: auth action=write-tests"
    Given feature "auth" at stage "tests" with review "passed"
    Then context shows "next: auth action=write-impl"

  Scenario: Context shows blocked for failed review
    Given feature "auth" at stage "seed" with review "failed"
    Then context shows "blocked: auth reason=review failed at seed stage"

  Scenario: Context shows done for impl with passed review
    Given feature "auth" at stage "impl" with review "passed"
    Then context shows "done: auth"

  Scenario: Context shows review-impl for impl with pending review
    Given feature "auth" at stage "impl" with review "pending"
    Then context shows "next: auth action=review-impl"

  Scenario: Context skips planned and deferred features
    Given feature "auth" with status "planned"
    Then context has no lines for "auth"

  Scenario: Context includes TODO tasks
    Given task "T-1" with status "TODO" for feature "auth"
    Then context shows "task: T-1 status=TODO"

  Scenario: Context shows blocked when prerequisite missing
    Given feature "auth" at stage "bdd" with no seed
    Then context shows "blocked: auth reason=missing seed"

  # Init Claude Code hooks
  Scenario: Init generates Claude Code hook scripts
    Given an empty directory with git initialized
    When I run "ptsd init --name MyApp"
    Then .claude/hooks/ptsd-context.sh exists and is executable
    And .claude/hooks/ptsd-gate.sh exists and is executable
    And .claude/hooks/ptsd-track.sh exists and is executable

  Scenario: Init generates Claude Code settings.json
    Given an empty directory with git initialized
    When I run "ptsd init --name MyApp"
    Then .claude/settings.json exists
    And settings.json has SessionStart hook pointing to ptsd-context.sh
    And settings.json has PreToolUse hook with matcher "Edit|Write" pointing to ptsd-gate.sh
    And settings.json has PostToolUse hook with matcher "Edit|Write" pointing to ptsd-track.sh

  Scenario: Init uses correct binary path in hook scripts
    Given an empty directory with git initialized
    When I run "ptsd init --name MyApp"
    Then hook scripts reference the ptsd binary path
