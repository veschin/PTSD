<p align="center">
  <img src="ptsd.png" width="600" alt="PTSD — a ringmaster putting AI-driven development through its paces" />
</p>

<h1 align="center">PTSD</h1>

<p align="center">
  <strong>PRD &rarr; Seed &rarr; BDD &rarr; Tests &rarr; Implementation. Every feature earns its place.</strong>
</p>

<p align="center">
  CLI tool that enforces structured AI development with a strict pipeline.<br>
  No skipping stages. No orphan files. No vibes-based shipping.<br>
  Go. Single binary. Zero third-party dependencies.
</p>

---

## What It Does

PTSD puts every feature through a gauntlet before it ships:

```
PRD → Seed → BDD → Tests → Implementation
```

Each stage is gated. Each gate requires review (score 0-10, min 7 to pass). No exceptions, no `--force`, no bypass flags.

Features are the atom. Nothing exists without a feature link — no orphan tests, no stray files, no loose code.

When used with Claude Code, PTSD auto-wires hooks that **enforce the pipeline in real time**: gate-check blocks out-of-order writes, auto-track advances stages on file creation, context injection shows the AI what to do next.

## Table of Contents

- [Pipeline](#pipeline)
- [Install](#install)
- [Quick Start](#quick-start)
- [Claude Code Integration](#claude-code-integration)
- [Commands](#commands)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [Output Modes](#output-modes)
- [Exit Codes](#exit-codes)
- [Dogfooding](#dogfooding)

## Pipeline

Every feature moves independently through five stages in strict order:

| Stage | Gate | What Happens |
|---|---|---|
| **PRD** | `<!-- feature:id -->` anchor exists | Requirements written with feature anchors |
| **Seed** | PRD anchor exists | Golden test data created in `seeds/<id>/` |
| **BDD** | Seed exists | Gherkin scenarios written in `bdd/<id>.feature` |
| **Tests** | BDD exists | Real test cases (no mocks), mapped to scenarios |
| **Impl** | Tests exist | Code written to make tests pass |

Skip a stage — blocked. Miss a review — blocked. Score below 7 — redo.

Git hooks enforce commit scopes (`[SCOPE] type: message`). Staged files must match declared scope.

## Install

Requires: Go 1.25+

```bash
go install github.com/veschin/ptsd/cmd/ptsd@latest
```

Or build from source:

```bash
git clone https://github.com/veschin/ptsd.git
cd ptsd
go build -o ptsd ./cmd/ptsd
```

## Quick Start

```bash
mkdir my-project && cd my-project
git init
ptsd init --name my-project

# Add features
ptsd feature add auth "User Authentication"
ptsd feature add api "REST API"
ptsd feature status auth in-progress

# Check what to do next
ptsd context --agent
# next: auth stage=prd action=write-seed

# Work through the pipeline: seed → bdd → tests → impl
# Each stage = separate commit with [SCOPE] type: message

# Review after each stage
ptsd review auth seed 8
ptsd review auth bdd 9

# Validate before commit
ptsd validate --agent
```

## Claude Code Integration

`ptsd init` generates `.claude/settings.json` with 4 hooks that enforce the pipeline automatically:

| Hook | Event | What It Does |
|---|---|---|
| **Context** | SessionStart, UserPromptSubmit | Runs `ptsd context --agent` — injects pipeline state into AI context |
| **GateCheck** | PreToolUse (Edit/Write) | Runs `ptsd hooks pre-tool-use` — blocks writes that violate pipeline order |
| **AutoTrack** | PostToolUse (Edit/Write) | Runs `ptsd hooks post-tool-use` — auto-advances feature stage on file writes |
| **commit-msg** | Git commit | Validates `[SCOPE] type:` format, checks staged files match scope |

Generated files:
```
.claude/
  settings.json                    # 4 hook events wired to ptsd
  hooks/
    ptsd-context.sh                # SessionStart + UserPromptSubmit
    ptsd-gate.sh                   # PreToolUse gate-check
    ptsd-track.sh                  # PostToolUse auto-track
  skills/<name>/SKILL.md           # 13 skills for Claude Code auto-discovery
```

The AI sees `next:`, `blocked:`, or `done:` for each feature — and skills tell it HOW to do each stage.

### Hook token overhead

Hooks add ~3-4% token overhead (~3K on a 100K session). Latency: ~100ms per hook call. Negligible.

## Commands

**Project setup:**
```bash
ptsd init [--name <name>]          # initialize .ptsd/, .claude/, git hooks
ptsd adopt                         # bootstrap ptsd onto existing project
```

**Features:**
```bash
ptsd feature list                  # all features and their status
ptsd feature add <id> <title>      # register a new feature
ptsd feature status <id> <status>  # set feature status (planned/in-progress/done)
ptsd feature show <id>             # show feature details
ptsd feature remove <id>           # remove a feature
```

**Pipeline:**
```bash
ptsd seed add <feature>            # initialize seed data
ptsd bdd add <feature>             # initialize BDD scenarios
ptsd prd check                     # validate PRD anchors
ptsd review <feature> <stage> <score>  # record review (score 0-10)
ptsd validate                      # check all pipeline gates
```

**Context and tracking:**
```bash
ptsd context --agent               # show pipeline state (next/blocked/done per feature)
ptsd status                        # project-wide pipeline overview
ptsd task next                     # next task to work on
```

**Hooks (called by Claude Code, not manually):**
```bash
ptsd hooks pre-tool-use            # reads tool JSON from stdin, gate-checks file path
ptsd hooks post-tool-use           # reads tool JSON from stdin, auto-tracks stage
ptsd hooks validate-commit --msg-file <path>  # commit-msg hook, validates scope + staged files
```

**Other:**
```bash
ptsd config                        # show/edit ptsd.yaml
ptsd skills                        # list pipeline skills
ptsd issues                        # common issues registry
ptsd gate-check                    # manual gate-check for a file path
ptsd auto-track                    # manual auto-track for a file path
```

## Project Structure

```
.ptsd/                             # all artifacts (git-tracked)
  ptsd.yaml                        # config (review.min_score, test patterns, etc.)
  features.yaml                    # feature registry (source of truth)
  state.yaml                       # hashes, scores, test results
  review-status.yaml               # per-feature: stage, tests, review, issues
  tasks.yaml                       # task queue
  issues.yaml                      # common issues registry
  docs/PRD.md                      # product requirements with <!-- feature:id --> anchors
  seeds/<id>/                      # golden seed data per feature
  bdd/<id>.feature                 # Gherkin scenarios per feature
  skills/                          # pipeline skill docs per stage

.claude/                           # Claude Code integration (git-tracked)
  settings.json                    # hook wiring (4 events)
  hooks/                           # shell scripts for hooks
  skills/<name>/SKILL.md           # 13 auto-discovery skills

.git/hooks/
  commit-msg                       # scope + staged file validation
```

## Architecture

```
cmd/ptsd/main.go → internal/cli/* → internal/core/* → internal/yaml/*
                                   → internal/render/*
```

| Package | Responsibility | Rule |
|---|---|---|
| `core/` | Domain logic — pipeline, validation, state, hooks | Zero TUI imports |
| `render/` | Output formatting — agent mode | Zero domain logic |
| `cli/` | Glue — args &rarr; core &rarr; render | One file per command |
| `yaml/` | Hand-rolled YAML parser | Leaf package, no deps |

`main.go` is a flat `switch` dispatcher (~75 lines). No cobra, no flag package. Manual arg parsing.

18 commands: init, adopt, feature, config, task, prd, seed, bdd, test, status, validate, hooks, review, skills, issues, context, gate-check, auto-track.

## Output Modes

**Agent mode** (`--agent`) — ultra-compact, zero decoration, exact `file:line` coordinates. Designed for LLM consumption. This is the primary mode — human TUI is not yet implemented.

Error format:
```
err:<category> <message>
```
Categories: `pipeline`, `config`, `io`, `user`, `test`

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Validation failure / pipeline violation |
| 2 | Bad arguments |
| 3 | Config error |
| 4 | I/O error |
| 5 | Test runner failure |

## Dogfooding

PTSD dogfoods itself — all development follows the ptsd pipeline. 4 rounds of AI agent testing documented in [FEEDBACK.md](FEEDBACK.md):

| Round | Agent | Features | Result |
|---|---|---|---|
| R1 | Sonnet (pre-hooks) | 2 | Found 4 bypasses (stage batching, fake reviews, structure divergence, global skills) |
| R2 | Sonnet (hooks enabled) | 2 | 32/44 pass. Found commit-msg bug, confirmed BYPASS-2 still open |
| R3 | GLM-5 (post-fixes) | 2 | 17/18 pass. All R2 fixes verified. Workflow skill missing `ptsd review` |
| R4 | Sonnet (5 features) | 5 | 9/10 pass. 21 tests, realistic data. Found BYPASS-5 (cross-feature batching), seed/bdd wipe bug |
