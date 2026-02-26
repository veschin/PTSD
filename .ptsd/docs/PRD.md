# PTSD — Product Requirements Document

## Overview

PTSD (PRD-Test-Seed Dashboard) is a CLI tool that enforces structured AI-driven development. Written in Go with Bubbletea TUI. Single binary, language-agnostic, zero runtime dependencies.

PTSD solves: LLMs skip pipeline steps, hide errors, write mocks instead of real tests, generate garbage files, corrupt existing code, and lie about quality.

PTSD enforces: strict per-feature pipeline (PRD → Seed → BDD → Tests → Implementation), quality scoring at every step, git hooks that block invalid commits, and skills that guide LLMs through each stage.

### Two Render Modes

One Bubbletea framework, two outputs:
- **Human mode** (default): interactive TUI dashboard — tables, colors, navigation, live updates.
- **Agent mode** (`--agent`): ultra-compact, zero decoration, exact file:line coordinates. Every token counts.

LLM is instructed to ALWAYS use `--agent`. Human TUI is never invoked by LLM.

---

## Design Principles

- **Feature is the atom.** Nothing exists without a feature link. No orphan files, no unlinked tests, no stray code.
- **Pipeline per feature.** Strict order, no skipping. Each feature moves independently: PRD → Seed → BDD → Tests → Impl.
- **Prevention over detection.** Design so mistakes CAN'T happen. Block, don't advise. Error, don't warn.
- **Strict by default.** No `--skip-validation`. No soft warnings. Pipeline violation = commit blocked.
- **Golden seed grounds reality.** Concrete example data before scenarios. Catches logical errors in data.
- **Quality over speed.** 0-10 review scoring. Below threshold = redo. Better to redo than ship garbage.
- **Token economy.** Every byte of output matters. Skills, status, errors — all minimal. Think more, output less.
- **Git-driven.** All artifacts committed. Hooks enforce pipeline. Commit prefixes allow partial work.
- **Transparency.** LLM must explain every problem. Hiding errors is worse than the error itself.
- **No mocks.** Real tests prove real behavior. Mocks only for external integrations. Internal mocks are banned.
- **Maximum automation.** Every enforceable rule IS enforced. Minimize LLM's room to deviate.

---

## Per-Feature Pipeline

```
PRD → Seed → BDD → Tests → Implementation
```

Each feature moves through stages independently. Each step requires review (quality score ≥ threshold) before proceeding to the next. Different features can be at different stages.

### Stage Requirements

| Stage | Entry Condition | Artifact |
|-------|----------------|----------|
| PRD | Feature registered | `<!-- feature:id -->` anchor in PRD |
| Seed | PRD anchor exists | `.ptsd/seeds/<id>/seed.yaml` + data files |
| BDD | Seed exists | `.ptsd/bdd/<id>.feature` with Gherkin scenarios |
| Tests | BDD scenarios exist | Test files in project's test location |
| Impl | Tests exist (may fail) | Implementation code; done when all tests pass |

### Pipeline Gates

- `ptsd bdd add <id>` — refused if feature has no seed.
- `ptsd test map <id>` — refused if feature has no BDD scenarios.
- Creating tasks with `--stage impl` — refused if feature has no tests.
- Advancing to `implemented` — refused unless all mapped tests pass.
- `ptsd validate` — errors on any pipeline violation. Blocks commit via git hook.

### Review Gate

After completing each pipeline stage, LLM runs the corresponding review skill. The review produces a score (0-10). If score < configured threshold (`review.min_score` in ptsd.yaml, default 7), the step must be redone. Score is stored in `.ptsd/state.yaml`.

### Mandatory Progress Tracking

LLM MUST record progress immediately as it happens — not at the end, not in batch, not "later". Every completed step, every review, every stage transition is recorded via `ptsd` commands the moment it occurs. This is non-negotiable.

When `ptsd` is not yet available (bootstrapping phase), LLM records progress in `.ptsd/state.yaml` and `.ptsd/review-status.yaml` manually. Once `ptsd` is operational, ALL progress tracking goes exclusively through `ptsd` commands — no direct file edits for state.

Review status format defined in [Quality Scoring](#quality-scoring) section under feature `review`.

Rationale: LLMs lose context between sessions. If progress isn't recorded immediately, it's lost. The next session starts blind. `ptsd` is the single source of truth for what was done, what passed review, and what's next.

---

## Core Concepts

### Feature

Unique parent key. Everything connects to a feature: PRD sections, seeds, BDD scenarios, tests, tasks.

- **ID**: ASCII slug (e.g., `user-auth`, `catalog`)
- **Title**: Human-readable name
- **Status**: `planned` | `in-progress` | `implemented` | `deferred`

Features with status `planned` or `deferred` are excluded from pipeline checks.

### Golden Seed

Concrete example data per feature. Lives in `.ptsd/seeds/<feature-id>/`.

**Purpose:**
- Ground BDD scenarios in reality — scenarios reference real data shapes.
- Catch logical errors — user sees actual data, validates it makes sense.
- Provide fixtures — tests CAN import seed data files.

**Manifest format** (`.ptsd/seeds/<id>/seed.yaml`):
```yaml
feature: user-auth
files:
  - path: user.json
    type: data
    description: "Valid user with all fields filled"
  - path: invalid-users.json
    type: data
    description: "Users with missing/invalid fields for error cases"
```

**Seed types:** `data` (example records), `fixture` (input/output pairs), `schema` (structural definitions).

### Quality Scoring

<!-- feature:review -->

LLM reviews each pipeline step using ptsd-provided skills. Each review outputs:
- **Score**: 0-10 (0 = doesn't match PRD at all, 10 = perfect)
- **Issues**: explicit list of problems found (can't just say "looks good")
- **Verdict**: pass / redo

Stored in `.ptsd/state.yaml` per feature per stage. Score history preserved.

#### Review Status File

`.ptsd/review-status.yaml` — per-feature review tracking. Always contains ALL registered features. Fields are deterministic, no free text. Managed by `ptsd review` command.

```yaml
features:
  config:
    stage: tests
    tests: written
    review: passed
    issues: 0
  output:
    stage: tests
    tests: written
    review: failed
    issues: 1
    issues_list:
      - "TestRenderTestResults format doesn't match PRD agent spec"
```

Fields:
- `stage` — current pipeline stage: `prd` | `seed` | `bdd` | `tests` | `impl`
- `tests` — test file status: `absent` | `written`
- `review` — review verdict: `pending` | `passed` | `failed`
- `issues` — count of open issues (0 = clean)
- `issues_list` — string array, one short line per issue (only present when issues > 0; removed when all resolved)

**Automation:** `ptsd review <feature-id> --stage <stage>` updates review-status.yaml automatically. Sets verdict, records issues, manages issues_list.

---

## Project Structure

```
project-root/
  .ptsd/                          # git-tracked, all ptsd artifacts
    ptsd.yaml                     # project config
    features.yaml                 # feature registry (source of truth)
    state.yaml                    # hashes, scores, test results
    review-status.yaml            # per-feature review verdicts and issues
    tasks.yaml                    # task list
    issues.yaml                   # common issues registry (recurring problems)
    docs/
      PRD.md                      # product requirements document
    seeds/
      <feature-id>/
        seed.yaml                 # manifest
        *.json|*.yaml             # data files
    bdd/
      <feature-id>.feature        # Gherkin scenarios
    skills/                       # skills for every pipeline stage
      write-prd.md
      write-seed.md
      write-bdd.md
      write-tests.md
      write-impl.md
      create-tasks.md
      review-prd.md
      review-seed.md
      review-bdd.md
      review-tests.md
      review-impl.md
      workflow.md
  CLAUDE.md / AGENTS.md           # generated by ptsd init
  src/                            # project code (normal location)
  tests/                          # project tests (normal location)
```

Tests and implementation live where the project normally puts them (not inside `.ptsd/`).

---

## Commands

### Project Lifecycle

**`ptsd init [--name N]`**

Scaffold `.ptsd/` directory with all subdirectories, config, empty registry, skills, and prompts. Generate `CLAUDE.md` / `AGENTS.md` with workflow instructions for AI assistants. Install git pre-commit hook.

**`ptsd adopt [--dry-run]`**

Scan existing project for artifacts (BDD files, tests, PRD). Extract feature IDs, create `.ptsd/` structure, import discovered artifacts, initialize state with current hashes.

### Feature Management

**`ptsd feature add <id> [--title T] [--status S]`** — Register feature. Status defaults to `planned`.

**`ptsd feature list [--status S]`** — List features with optional filter.

**`ptsd feature show <id>`** — Full context: PRD section, seed status, BDD scenario count, test count, tasks, review scores.

**`ptsd feature status <id> <new-status>`** — Update status. `implemented` requires all tests passing.

### Task Management

**`ptsd task add --feature <id> "title" [--status TODO|WIP] [--priority A|B|C]`** — Create task linked to feature. Priority: A=blocking, B=important, C=future.

**`ptsd task next [--limit N]`** — Next unblocked tasks. Agent mode output:
```
T-1 [TODO] [A] [PRD:l30-40 BDD:l30-100 TEST:l0-200]: Implement user auth
T-2 [TODO] [B] [PRD:l50-60 BDD:l110-150]: Write catalog tests
```

**`ptsd task list [--status S] [--feature F]`** — List tasks with filters.

**`ptsd task update <id> --status TODO|WIP|DONE`** — Update task status.

### Pipeline Steps

**`ptsd prd check`** — Validate PRD anchors vs feature registry. Missing anchors = error.

**`ptsd seed init <feature-id>`** — Create seed directory and manifest.

**`ptsd seed add <feature-id> <file> [--type T]`** — Add file to seed manifest.

**`ptsd seed check`** — Validate all active features have seeds.

**`ptsd bdd add <feature-id>`** — Create `.feature` file from template. Refuses without seed.

**`ptsd bdd check`** — Validate all feature tags exist in registry. Report uncovered features.

**`ptsd bdd show <feature-id>`** — Compact scenarios, one line each: `Title: Given X / When Y / Then Z`.

**`ptsd test map <bdd-file> <test-file>`** — Create explicit BDD-to-test mapping.

**`ptsd test check`** — Show mapping coverage. Status per BDD file: covered, partial, no-tests.

**`ptsd test run [--feature F]`** — Execute tests. Agent mode output: `pass:7 fail:2 fail:tests/foo.test.ts:42,tests/bar.test.ts:18`.

### Status & Validation

**`ptsd status`** — Project overview. Agent mode:
```
[FEAT:5 FAIL:0] [BDD:5 FAIL:0] [TESTS:5 FAIL:0] [T:20 WIP:0 TODO:19 DONE:1]
```

**`ptsd status --feature <id>`** — Per-feature gap analysis: scenario count vs test count, review scores, pipeline stage.

**`ptsd validate`** — Full pipeline validation. Errors on violations. Used by git hook.

**`ptsd review <feature-id> --stage <stage>`** — Record review score for a feature at a stage. Stores in state.

---

## State Tracking

<!-- feature:state-tracking -->

PTSD maintains state at `.ptsd/state.yaml`. For each feature: pipeline stage, file hashes (seed, BDD, test files), review scores, last test results.

### Hash-Based Regression Detection

On every state-reading operation (`status`, `validate`, `task next`), ptsd compares current file hashes against stored hashes.

| File Changed | Feature Stage | Action |
|-------------|---------------|--------|
| Seed | beyond seed | WARN — downstream may be stale |
| BDD | beyond bdd | WARN — tests may be stale |
| PRD section | beyond prd | ERROR — downgrade stage, create redo tasks |
| Test file | implemented | WARN — re-run tests |

When PRD changes for a feature with downstream artifacts:
1. Feature stage downgraded.
2. Affected BDD/tests flagged as potentially stale.
3. Refactoring task auto-created.
4. Feature blocked until re-validated through pipeline.

### State Format

```yaml
features:
  user-auth:
    stage: bdd
    hashes:
      ".ptsd/seeds/user-auth/user.json": "sha256:a3f2..."
      ".ptsd/bdd/user-auth.feature": "sha256:e1a2..."
    scores:
      prd: { score: 8, at: "2026-02-26T10:00:00" }
      seed: { score: 9, at: "2026-02-26T11:00:00" }
    tests: null
```

---

## Skills

<!-- feature:skills -->

PTSD ships skills for every pipeline stage — both creation and review. Skills are structured instructions in a universal format understood by Claude Code, Cursor, OpenCode, and other AI tools.

### Creation Skills

Guide the LLM through producing artifacts correctly:

| Skill | Purpose |
|-------|---------|
| `write-prd.md` | How to write a PRD section: structure, anchors, edge cases, non-goals |
| `write-seed.md` | How to create golden seed: realistic data, happy + edge cases, manifest |
| `write-bdd.md` | How to write Gherkin scenarios from PRD + seed, all paths |
| `write-tests.md` | How to write tests from BDD: 1:1 mapping, no mocks, real assertions |
| `write-impl.md` | How to implement from tests: no extra code, no over-engineering |
| `create-tasks.md` | How to create tasks: feature link, priority, clear checklist |

### Review Skills

Guide the LLM through evaluating artifacts, output score 0-10:

| Skill | Checks |
|-------|--------|
| `review-prd.md` | Completeness, edge cases, non-goals, clarity |
| `review-seed.md` | Data coverage, realism, happy + edge cases |
| `review-bdd.md` | Scenarios match PRD, all paths, no gaps |
| `review-tests.md` | 1:1 BDD mapping, no mocks, real assertions |
| `review-impl.md` | All tests pass, code matches design, no cheating |

### Workflow Skill

`workflow.md` — Full pipeline, mandatory order, what to invoke when.

### Skill Format

```markdown
---
name: write-bdd
description: Guide for writing Gherkin BDD scenarios from PRD and seed data
trigger: When creating .feature files for a feature
---

[Structured instructions here]
```

Generated into `.ptsd/skills/` by `ptsd init`. Optionally symlinked to `.claude/skills/` or tool-specific location for auto-discovery.

---

## AI Integration

<!-- feature:ai-integration -->

### CLAUDE.md / AGENTS.md Generation

On `ptsd init`, generate agent instruction file containing:

- **Workflow**: always run `ptsd task next --agent` at session start, follow its output.
- **Commands**: all ptsd commands with `--agent` flag, when to use each.
- **Forbidden**: mocking, hardcoding, skipping steps, hiding errors, generating garbage files, corrupting existing code.
- **Required**: explain every error encountered, run `ptsd validate --agent` before commit, use commit prefixes.
- **Philosophy**: think more, do less. Better to ask than to guess. Better to redo than to ship garbage.
- **Transparency**: if something fails, explain WHY. Never suppress errors. Never pretend something works when it doesn't.

### Anti-Cheat Enforcement

If it can be validated automatically, it MUST be:

- Pre-commit hook runs `ptsd validate`. Violation = ERROR = commit blocked. No bypass flag.
- `ptsd bdd add` refuses without seed. `ptsd test map` refuses without BDD.
- Stage advancement requires passing review score (≥ threshold).
- Review skills force LLM to list explicit issues (can't output empty review).
- `ptsd validate` scans test files for mock/stub patterns and flags them.
- No `--force`, `--skip-validation`, or `--no-verify` flags exist.

---

## Git Integration

<!-- feature:git-integration -->

### Hooks

`ptsd init` generates `.git/hooks/pre-commit`:

```sh
#!/bin/sh
ptsd validate
```

Commit blocked on any pipeline violation. No bypass.

### Commit Message Format

```
[SCOPE] type: message
```

**Scopes** (what pipeline artifact the commit touches):

| Scope | Covers |
|-------|--------|
| `[PRD]` | PRD document |
| `[SEED]` | Seed data |
| `[BDD]` | Gherkin scenarios |
| `[TEST]` | Test files |
| `[IMPL]` | Implementation code |
| `[TASK]` | Task management |
| `[STATUS]` | State/status updates |

**Types** (conventional commit actions): `feat`, `add`, `fix`, `refactor`, `remove`, `update`.

**Validation logic:**

ptsd knows which files belong to which stage by their paths. On commit:

1. Reads `git diff --staged` — classifies each file by stage.
2. Reads commit message scope `[SCOPE]`.
3. If staged files don't match the declared scope — **ERROR, commit blocked**.
4. If scope is missing — **ERROR, commit blocked**.
5. Scope determines which pipeline checks to run (e.g. `[BDD]` skips test/impl validation).

**Examples:**
```
[PRD] add: user authentication feature section
[SEED] add: sample user data for auth feature
[BDD] feat: login and registration scenarios
[TEST] add: auth endpoint tests matching BDD scenarios
[IMPL] feat: implement user authentication
[TASK] add: catalog API implementation task
[STATUS] update: mark auth feature as implemented
```

Pre-commit hook enforces this automatically. No bypass.

---

## Configuration

<!-- feature:config -->

`.ptsd/ptsd.yaml`:

```yaml
project:
  name: "MyApp"

testing:
  runner: "npx vitest run"
  patterns:
    files: ["**/*.test.ts"]
  result_parser:
    format: json
    root: "testResults"
    status_field: "status"
    passed_value: "passed"
    failed_value: "failed"

review:
  min_score: 7
  auto_redo: true

hooks:
  pre_commit: true
  scopes: [PRD, SEED, BDD, TEST, IMPL, TASK, STATUS]
  types: [feat, add, fix, refactor, remove, update]
```

### Test Adapter Selection

1. If `result_parser` present — use generic configurable adapter.
2. If `runner` contains known keyword (vitest, jest, pytest, "go test") — use built-in preset.
3. Otherwise — exit-code adapter (pass if exit 0).

---

## Output Design

<!-- feature:output -->

All rendering via Bubbletea. `--agent` flag switches render mode.

### Human Mode (default)

Interactive TUI dashboard. Tables with colors, navigation between features, live test status, pipeline progress visualization. Rich and informative.

### Agent Mode (`--agent`)

Ultra-compact. Zero decoration. Exact coordinates. Every token counts.

```
# ptsd status --agent
[FEAT:5 FAIL:0] [BDD:5 FAIL:0] [TESTS:5 FAIL:0] [T:20 WIP:0 TODO:19 DONE:1]

# ptsd task next --agent --limit 3
T-1 [TODO] [A] [PRD:l30-40 BDD:l30-100 TEST:l0-200]: Implement user auth
T-2 [TODO] [B] [PRD:l50-60 BDD:l110-150]: Write catalog tests
T-3 [WIP] [A] [PRD:l70-80]: Create payment seed data

# ptsd test run --agent
pass:7 fail:2 fail:tests/auth.test.ts:42,tests/catalog.test.ts:18

# ptsd feature show user-auth --agent
user-auth [in-progress] PRD:l30-40 SEED:ok BDD:3scn TEST:2/3 SCORE:prd=8,seed=9,bdd=7

# ptsd validate --agent
err:pipeline user-auth has bdd but no tests
err:pipeline catalog has no seed
```

### Error Format

All errors: `err:<category> <message>`. Single line. No stack traces.

Categories: `pipeline`, `config`, `io`, `user`, `test`.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success, no issues |
| 1 | Validation failure, pipeline violation |
| 2 | Bad arguments, unknown command |
| 3 | Config error |
| 4 | I/O error |
| 5 | Test runner failure |

---

## Common Issues Registry

<!-- feature:common-issues -->

PTSD maintains a per-project **common issues** file at `.ptsd/issues.yaml`. This is a compressed knowledge base of recurring problems encountered during development.

### Purpose

LLMs lose context between sessions. The same mistakes repeat: wrong venv, missing keys, reading stale files, misconfigured tools. Common issues file breaks this cycle — LLM reads it at session start and avoids known traps.

### Format

```yaml
issues:
  - id: venv-wrong-python
    category: env
    summary: "venv uses system python instead of project python"
    fix: "rm -rf .venv && python3.12 -m venv .venv"
  - id: missing-api-key
    category: access
    summary: "OPENAI_API_KEY not set in .env"
    fix: "cp .env.example .env && fill keys"
  - id: stale-lock
    category: io
    summary: "LLM reads package-lock.json instead of pnpm-lock.yaml"
    fix: "always check which package manager is configured in ptsd.yaml"
```

Fields:
- `id` — slug, unique per issue
- `category` — `env` | `access` | `io` | `config` | `test` | `llm`
- `summary` — one line, max 80 chars
- `fix` — one line, concrete action

### Rules

1. **LLM MUST read `.ptsd/issues.yaml` at session start** — before any work.
2. **Add issue immediately** when a problem occurs for the second time.
3. **Remove issue** when root cause is permanently fixed (not just worked around).
4. **No duplicates.** Check existing issues before adding.
5. **No essays.** Summary + fix = two short lines. If it needs more, it's a task, not an issue.

### Commands

**`ptsd issues list [--category C]`** — List all issues, optionally filtered.

**`ptsd issues add --category C "summary" --fix "fix"`** — Add new issue.

**`ptsd issues remove <id>`** — Remove resolved issue.

### Categories

| Category | Covers |
|----------|--------|
| `env` | Wrong runtime, venv, node version, PATH |
| `access` | Missing API keys, tokens, credentials, permissions |
| `io` | Reading wrong files, stale caches, wrong paths |
| `config` | Misconfigured tools, wrong settings |
| `test` | Flaky tests, wrong test runner, missing fixtures |
| `llm` | LLM-specific: hallucinated paths, wrong assumptions, repeated mistakes |

---

## Architecture

Go module. Single binary. Three-layer structure:

```
cmd/ptsd/main.go          → entry point, arg routing
internal/
  cli/                    → command handlers (thin: parse args → call core → render)
    init.go
    feature.go
    task.go
    status.go
    validate.go
    ...
  core/                   → domain logic (testable, no TUI deps)
    config.go             → load/merge ptsd.yaml
    registry.go           → feature CRUD on features.yaml
    pipeline.go           → stage gates, transitions, validation
    state.go              → state.yaml, hashes, scores
    tasks.go              → task CRUD on tasks.yaml
    prd.go                → PRD anchor extraction
    bdd.go                → .feature file parsing
    testrunner.go         → execute configured runner, parse results
    seed.go               → seed management
    review.go             → score storage and threshold checks
    hooks.go              → pre-commit hook generation and validation
    classify.go           → file path → pipeline stage classification
  render/                 → Bubbletea output layer
    agent.go              → compact text renderer (--agent mode)
    tui.go                → interactive TUI renderer
    models.go             → shared view models
  yaml/                   → hand-rolled YAML parser (no deps)
    parse.go
    serialize.go
```

### Dependency Flow

```
cmd/ptsd → cli/* → core/* → yaml/*
                 → render/*
```

- `core/` has zero TUI imports. Pure domain logic.
- `render/` has zero core logic. Only formatting.
- `cli/` is the glue: calls core, passes results to render.
- `yaml/` is a leaf package. No imports from project.

### Key Interfaces

```go
// Renderer — cli/ calls this, doesn't know if TUI or agent
type Renderer interface {
    Status(data StatusData)
    TaskList(tasks []Task)
    Error(category string, message string)
    // ...
}

// TestRunner — core/ calls configured runner
type TestRunner interface {
    Run(patterns []string) (TestResults, error)
}
```

### Build

```
go build -o ptsd cmd/ptsd/main.go
```

Single binary. No CGO. Cross-compile for linux/darwin/windows.
