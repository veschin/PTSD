# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What Is This

PTSD (PRD-Test-Seed Dashboard) — CLI tool enforcing structured AI development.
Go, single binary, zero third-party dependencies (stdlib only, no `go.sum`).

**This project dogfoods itself.** All PTSD practices apply here: features, pipeline, seeds, BDD, reviews.

## Build & Test Commands

```bash
go build ./cmd/ptsd/...           # build binary
go test ./...                     # run all tests
go test ./internal/core/...       # run core package tests
go test -run TestValidate ./internal/core/  # run a single test
go test -v -run TestName ./path/  # verbose single test
```

No Makefile, no linter config. Standard Go tooling only.

## Architecture

```
cmd/ptsd/main.go → internal/cli/* → internal/core/* → internal/yaml/*
                                   → internal/render/*
```

- `core/` — domain logic, zero TUI imports. All YAML parsing is inline here (line-by-line `strings.Split`/`HasPrefix`/`TrimPrefix`, no third-party parser)
- `render/` — output formatting. Only `AgentRenderer` exists; HumanRenderer (Bubbletea TUI) is not yet implemented
- `cli/` — glue: args → core → render. Signature `func RunX(args []string, agentMode bool) int`. Most commands get their own file, but `pipeline.go` groups `prd`/`seed`/`bdd`/`test` and `init.go` groups `init`/`adopt`
- `yaml/` — declared as leaf package but currently empty; all parsing lives in `core/`
- `core/templates.go` — uses `//go:embed templates/*` to ship skills, hook scripts, `settings.json` template inside the binary

### CLI Command Registration

`main.go` is a flat `switch cmd` dispatcher. No cobra, no flag package. Arg parsing is manual loop-over-args. To add a new command:
1. New `case "cmd-name":` in `main.go`
2. New `cli/cmd-name.go` with `func RunCmdName(args []string, agentMode bool) int`

Go module: `github.com/veschin/ptsd`. Internal imports use `github.com/veschin/ptsd/internal/...`

### Error Protocol

Errors flow as `fmt.Errorf("err:category message")`. The `coreError()` helper in `cli/helpers.go` parses this prefix to route to exit codes:
- `err:validation` / `err:pipeline` → exit 1
- `err:user` → exit 2
- `err:config` → exit 3
- `err:io` → exit 4
- `err:test` → exit 5

### Hook Auto-Wiring

`ptsd init` generates `.claude/hooks/*.sh` scripts and `.claude/settings.json` that wires them as Claude Code hooks:
- `SessionStart` + `UserPromptSubmit` → `ptsd-context.sh` → `ptsd context --agent` (injects pipeline state)
- `PreToolUse` (Edit|Write) → `ptsd-gate.sh` → `ptsd hooks pre-tool-use` → `GateCheck()` (blocks pipeline-violating writes)
- `PostToolUse` (Edit|Write) → `ptsd-track.sh` → `ptsd hooks post-tool-use` → `AutoTrack()` (auto-advances feature stage)

Hooks read Claude Code's JSON from stdin, extract `file_path` via string search (no JSON decoder), return exit 2 to block or 0 to allow.

`ptsd init` also generates git hooks: `pre-commit` runs `ptsd validate`, `commit-msg` runs `ptsd hooks validate-commit`.

Re-running `ptsd init` is safe (idempotent) — regenerates hooks/skills/CLAUDE.md section without touching data files.

### Key Domain Types

- `core/registry.go` — `Feature` struct, CRUD on `.ptsd/features.yaml`
- `core/state.go` — `State`/`FeatureState` with hashes, scores, test mappings; `CheckRegressions()` compares SHA256 hashes (PRD changes downgrade stage; seed/BDD/test changes warn only)
- `core/pipeline.go` — `Validate()` orchestrates all checks; `ClassifyFile()` maps paths to scopes
- `core/context.go` — `BuildContext()` combines features + review-status + tasks → emits `next`/`blocked`/`done`/`task` lines
- `core/review.go` — `ReviewStatusEntry`, `RecordReview()`, `CheckReviewGate()`

### Feature ID as Canonical Link

Everything resolves by feature ID: test files map via `matchFeatureID()` (exact match first, then longest substring match), BDD files are `<id>.feature`, seeds are `seeds/<id>/`, PRD anchors are `<!-- feature:<id> -->`.

`planned` and `deferred` features are excluded from all pipeline checks.

## Testing Patterns

Tests use real files, real CLI, temp directories. No mocks for internal code.

**Central test helper** (`internal/core/test_helper_test.go`):
```go
func setupProjectWithFeatures(t *testing.T, features ...string) string
// Creates temp dir with full .ptsd/ structure. Features are "id:status" strings.
```

**Integration tests** (`cmd/ptsd/main_test.go`): build real binary via `exec.Command("go", "build", "-o", bin, ".")`, run in temp project dir, assert on stdout/stderr/exit codes.

**Assertion helper**: `assertHasError(t, errors, feature, category, contains)` — searches error slice for matching feature+category+substring.

## Project Structure

```
.ptsd/                    # all ptsd artifacts (git-tracked)
  ptsd.yaml               # config
  features.yaml           # feature registry (source of truth)
  state.yaml              # hashes, scores, test results
  review-status.yaml      # per-feature review verdicts and issues
  tasks.yaml              # tasks
  issues.yaml             # common issues registry
  docs/PRD.md             # product requirements
  seeds/<id>/             # golden seed data per feature
  bdd/<id>.feature        # Gherkin scenarios per feature
  skills/                 # pipeline skills for every stage
```

## Pipeline (strict order per feature)

```
PRD → Seed → BDD → Tests → Implementation
```

Every artifact links to a feature. No orphan files. No skipping steps.

### Pipeline Gates (enforced)

- No BDD without seed (`ptsd bdd add` refuses)
- No test mapping without BDD (`ptsd test map` refuses)
- No impl tasks without tests
- No `implemented` status unless all tests pass
- `ptsd validate` blocks commit on any violation

### Review Gate

Each stage requires review with score 0-10. Score < `review.min_score` (default 7) = redo.
Review stored in `.ptsd/state.yaml`. Review status in `.ptsd/review-status.yaml`.
When `review.auto_redo: true` and score < min, a redo task is automatically appended to `tasks.yaml`.

## Rules

1. **No mocks.** Tests prove real behavior. Mock only external integrations.
2. **No garbage files.** Every file has a purpose and a feature link.
3. **No hiding errors.** If something fails, explain why. Never suppress.
4. **No over-engineering.** Minimum code for current task. No premature abstractions.
5. **Commit format:** `[SCOPE] type: message`. Scopes: PRD, SEED, BDD, TEST, IMPL, TASK, STATUS. Types: feat, add, fix, refactor, remove, update.
6. **Token economy.** Minimal output, minimal code, minimal comments. Only what matters.
7. **All English.** Code, comments, docs, commits — English only.
8. **Feature is the atom.** Nothing exists without feature link. No orphan files, tests, or code.
9. **No bypass flags.** No `--force`, `--skip-validation`, `--no-verify`. Ever.
10. **Transparency.** If something fails, explain WHY. Never pretend it works.

## Mandatory Session Start Protocol

On EVERY session start, BEFORE any work:
1. Read `.ptsd/review-status.yaml` — where each feature is
2. Read `.ptsd/tasks.yaml` — pending tasks
3. Read `.ptsd/issues.yaml` — known recurring issues (avoid repeating mistakes)
4. Decide what to work on based on current state

## Mandatory Progress Tracking

Record progress IMMEDIATELY as it happens. Not at end, not in batch. Every state change = immediate file update to `review-status.yaml`, `state.yaml`, or `tasks.yaml`.

### review-status.yaml format

File contains ALL registered features. Fields:
- `stage`: `prd` | `seed` | `bdd` | `tests` | `impl`
- `tests`: `absent` | `written`
- `review`: `pending` | `passed` | `failed`
- `issues`: int (0 = clean)
- `issues_list`: string[] (only when issues > 0)

Update triggers: test written, review done, stage advanced, issue fixed.
On fix: remove from `issues_list`, decrement. At 0: set `passed`, drop `issues_list`.

### Gate Check: AI-Blocked Files

`GateCheck()` blocks LLM writes to files not in the allowed list. Key restriction: **`.ptsd/review-status.yaml` cannot be edited directly** — use `ptsd review` commands instead. Allowed files include `.ptsd/docs/PRD.md`, `.ptsd/tasks.yaml`, `.ptsd/state.yaml`, `.ptsd/features.yaml`, `.ptsd/ptsd.yaml`, `.ptsd/issues.yaml`, `CLAUDE.md`, `.claude/settings.json`, `.ptsd/skills/**`, `.claude/hooks/**`.

### Skills Generation

`ptsd init` writes 13 skill files to both `.ptsd/skills/` (project reference) and `.claude/skills/<name>/SKILL.md` (Claude Code auto-discovery). Skills: `write-prd`, `write-seed`, `write-bdd`, `write-tests`, `write-impl`, `create-tasks`, `review-prd`, `review-seed`, `review-bdd`, `review-tests`, `review-impl`, `adopt`, `workflow`.

## CLI Commands

20 commands: `init`, `adopt`, `feature`, `config`, `task`, `prd`, `seed`, `bdd`, `test`, `status`, `validate`, `hooks`, `review`, `skills`, `issues`, `context`, `gate-check`, `auto-track`, `help`, `version`.

Key subcommands: `prd check|show`, `seed init|add`, `bdd add|list`, `test run|map`, `feature add|list|status`, `task add|list|next|done`.

Note: `ptsd test map` requires a `@feature:<id>` tag in the BDD file.

## Workflow

1. `ptsd task next --agent` — get next task
2. Read the linked PRD section, BDD scenarios, seed data
3. Do the work
4. **Record progress immediately** in state.yaml / review-status.yaml / tasks.yaml
5. `ptsd validate --agent` — check before commit
6. Commit with proper `[SCOPE] type: message`

## Commit Scope Validation

ptsd classifies staged files by path → pipeline stage. On commit:
1. Files must match declared `[SCOPE]`
2. Missing scope = blocked
3. Mismatched scope = blocked

## Output Modes

- **Human mode** (default): interactive TUI (not yet implemented — returns AgentRenderer)
- **Agent mode** (`--agent`): ultra-compact, zero decoration, exact file:line coordinates

LLM ALWAYS uses `--agent`. Error format: `err:<category> <message>` (single line, no stack traces).
Categories: pipeline, config, io, user, test.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Validation failure, pipeline violation |
| 2 | Bad arguments |
| 3 | Config error |
| 4 | I/O error |
| 5 | Test runner failure |

## Features (17 total)

### Core
- `config` — Configuration System
- `feature-mgmt` — Feature Management
- `task-mgmt` — Task Management
- `validate` — Pipeline Validation
- `git-hooks` — Git Hook Enforcement
- `status` — Project Status
- `output` — Dual Render Mode
- `init` — Project Initialization

### Pipeline
- `prd-check` — PRD Validation
- `seed-mgmt` — Seed Management
- `bdd-mgmt` — BDD Management
- `test-integration` — Test Integration
- `state-tracking` — State & Regression Detection

### Advanced
- `review` — Quality Scoring
- `skills` — Skill Generation
- `adopt` — Existing Project Bootstrap
- `common-issues` — Common Issues Registry

## Common Issues Registry

`.ptsd/issues.yaml` — recurring problems. LLM reads at session start.
- `id`: slug, unique
- `category`: env | access | io | config | test | llm
- `summary`: one line, max 80 chars
- `fix`: one line, concrete action

Add on second occurrence. Remove when root cause fixed. No duplicates, no essays.

## Key Paths

| What | Where |
|------|-------|
| PRD | `.ptsd/docs/PRD.md` |
| Features | `.ptsd/features.yaml` |
| Review Status | `.ptsd/review-status.yaml` |
| State | `.ptsd/state.yaml` |
| Tasks | `.ptsd/tasks.yaml` |
| Issues | `.ptsd/issues.yaml` |
| Config | `.ptsd/ptsd.yaml` |
| BDD | `.ptsd/bdd/<id>.feature` |
| Seeds | `.ptsd/seeds/<id>/` |
| Skills | `.ptsd/skills/` |
| Source | `cmd/`, `internal/` |
| Tests | `*_test.go` alongside source |
