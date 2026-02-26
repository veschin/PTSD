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
  features.yaml            # feature registry
  state.yaml              # hashes, scores, test results
  tasks.yaml              # tasks
  docs/PRD.md             # product requirements
  seeds/<id>/             # golden seed data per feature
  bdd/<id>.feature        # Gherkin scenarios per feature
  skills/                 # pipeline skills
```

## Pipeline (strict order per feature)

```
PRD → Seed → BDD → Tests → Implementation
```

Every artifact links to a feature. No orphan files. No skipping steps.

## Rules

1. **No mocks.** Tests prove real behavior. Mock only external integrations.
2. **No garbage files.** Think before creating. Every file has a purpose and a feature link.
3. **No hiding errors.** If something fails, explain why. Never suppress.
4. **No over-engineering.** Minimum code for current task. No premature abstractions.
5. **Commit format:** `[SCOPE] type: message`. Scopes: PRD, SEED, BDD, TEST, IMPL, TASK, STATUS.
6. **Token economy.** Minimal output, minimal code, minimal comments. Only what matters.
7. **All English.** Code, comments, docs, commits — English only.

## Mandatory Progress Tracking

LLM MUST record progress immediately as it happens. Not at the end, not in batch, not "later".

**ptsd CLI does not exist yet (bootstrapping phase).** Track manually:
- `.ptsd/state.yaml` — feature stages, hashes, scores
- `.ptsd/review-status.yaml` — review verdicts per feature
- `.ptsd/tasks.yaml` — task status updates

Once ptsd CLI is operational — ALL tracking goes exclusively through `ptsd` commands. No direct file edits for state.

Rationale: LLMs lose context between sessions. Unrecorded progress = lost progress.

### How to update review-status.yaml

File always contains ALL registered features. On review:

1. Set `review` to `passed` or `failed`
2. Set `issues` to count of problems found
3. If `issues` > 0, add `issues_list` array with one short line per problem
4. When issue is fixed — remove it from `issues_list`, decrement `issues`
5. When `issues` reaches 0 and all clean — set `review: passed`, remove `issues_list`

Fields:
- `stage`: `prd` | `seed` | `bdd` | `tests` | `impl`
- `tests`: `absent` | `written`
- `review`: `pending` | `passed` | `failed`
- `issues`: int (0 = clean)
- `issues_list`: string[] (only when issues > 0)

## Workflow

1. `ptsd task next --agent` — get next task (when ptsd exists; until then check tasks.yaml manually)
2. Read the linked PRD section, BDD scenarios, seed data
3. Do the work
4. **Record progress immediately** in state.yaml / review-status.yaml / tasks.yaml
5. `ptsd validate --agent` — check before commit (when ptsd exists)
6. Commit with proper `[SCOPE] type: message`

## Testing

```bash
go test ./...
```

Tests use real files, real CLI, temp directories. No mocks for internal code.

## Key Paths

| What | Where |
|------|-------|
| PRD | `.ptsd/docs/PRD.md` |
| Features | `.ptsd/features.yaml` |
| BDD | `.ptsd/bdd/<id>.feature` |
| Seeds | `.ptsd/seeds/<id>/` |
| Tasks | `.ptsd/tasks.yaml` |
| State | `.ptsd/state.yaml` |
| Config | `.ptsd/ptsd.yaml` |
| Skills | `.ptsd/skills/` |
| Source | `cmd/`, `internal/` |
| Tests | `*_test.go` alongside source |
