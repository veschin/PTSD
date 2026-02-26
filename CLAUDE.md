# PTSD Development Guide

## What Is This

PTSD (PRD-Test-Seed Dashboard) — CLI tool enforcing structured AI development.
Go + Bubbletea. Single binary. Zero runtime deps.

**This project dogfoods itself.** All PTSD practices apply here: features, pipeline, seeds, BDD, reviews.

## Architecture

```
cmd/ptsd/main.go → internal/cli/* → internal/core/* → internal/yaml/*
                                   → internal/render/*
```

- `core/` — domain logic, zero TUI imports
- `render/` — Bubbletea output, zero domain logic
- `cli/` — glue: args → core → render
- `yaml/` — hand-rolled parser, leaf package

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

## Rules

1. **No mocks.** Tests prove real behavior. Mock only external integrations.
2. **No garbage files.** Think before creating. Every file has a purpose and a feature link.
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

LLM MUST record progress immediately as it happens. Not at the end, not in batch, not "later".

### When to update review-status.yaml

Update IMMEDIATELY when any of these happen:
- Test written → `tests: written`
- Test reviewed → `review: passed` or `failed` + `issues`/`issues_list`
- Implementation written → `stage: impl`, `review: pending`
- Implementation reviewed → `review: passed` or `failed`
- Issue fixed → remove from `issues_list`, decrement `issues`
- Stage advanced → update `stage`

Do NOT batch updates. Do NOT defer to end of session. Every state change = immediate file update.

### review-status.yaml format

File always contains ALL registered features. Fields:
- `stage`: `prd` | `seed` | `bdd` | `tests` | `impl`
- `tests`: `absent` | `written`
- `review`: `pending` | `passed` | `failed`
- `issues`: int (0 = clean)
- `issues_list`: string[] (only when issues > 0)

On review: set verdict, count issues, list them. On fix: remove from list, decrement. At 0: set passed, drop list.

### Bootstrapping phase

**ptsd CLI does not exist yet.** Track manually:
- `.ptsd/state.yaml` — feature stages, hashes, scores
- `.ptsd/review-status.yaml` — review verdicts per feature
- `.ptsd/tasks.yaml` — task status updates

Once ptsd CLI is operational — ALL tracking goes exclusively through `ptsd` commands. No direct file edits for state.

## Workflow

1. `ptsd task next --agent` — get next task (bootstrapping: check tasks.yaml manually)
2. Read the linked PRD section, BDD scenarios, seed data
3. Do the work
4. **Record progress immediately** in state.yaml / review-status.yaml / tasks.yaml
5. `ptsd validate --agent` — check before commit (when ptsd exists)
6. Commit with proper `[SCOPE] type: message`

## Commit Scope Validation

ptsd classifies staged files by path → pipeline stage. On commit:
1. Files must match declared `[SCOPE]`
2. Missing scope = blocked
3. Mismatched scope = blocked

## Output Modes

- **Human mode** (default): interactive TUI — tables, colors, navigation
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

## Testing

```bash
go test ./...
```

Tests use real files, real CLI, temp directories. No mocks for internal code.

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
