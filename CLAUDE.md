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

## Workflow

1. `ptsd task next --agent` — get next task
2. Read the linked PRD section, BDD scenarios, seed data
3. Do the work
4. `ptsd validate --agent` — check before commit
5. Commit with proper `[SCOPE] type: message`

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
