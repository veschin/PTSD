<p align="center">
  <img src="ptsd.png" width="600" alt="PTSD — a ringmaster putting AI-driven development through its paces" />
</p>

<h1 align="center">PTSD</h1>

<p align="center">
  <strong>PRD &rarr; Test &rarr; Seed &rarr; Dashboard. Every feature earns its place.</strong>
</p>

<p align="center">
  CLI tool that enforces structured AI development with a strict pipeline.<br>
  No skipping stages. No orphan files. No vibes-based shipping.<br>
  Go + Bubbletea. Single binary. Zero runtime deps.
</p>

---

## What It Does

PTSD puts every feature through a gauntlet before it ships:

```
PRD → Seed → BDD → Tests → Implementation
```

Each stage is gated. Each gate requires review. Score below threshold — redo. No exceptions, no `--force`, no bypass flags.

Features are the atom. Nothing exists without a feature link — no orphan tests, no stray files, no loose code.

## Table of Contents

- [Pipeline](#pipeline)
- [Install](#install)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [Features](#features)
- [Output Modes](#output-modes)
- [Exit Codes](#exit-codes)

## Pipeline

Every feature moves independently through five stages in strict order:

| Stage | Gate | What Happens |
|---|---|---|
| **PRD** | Review score &ge; 7 | Requirements written, validated |
| **Seed** | PRD passed | Golden test data created |
| **BDD** | Seed exists | Gherkin scenarios written |
| **Tests** | BDD exists | Test cases mapped to scenarios |
| **Impl** | All tests pass | Code written, validated, shipped |

Skip a stage — blocked. Miss a review — blocked. Score below threshold — redo.

Git hooks enforce commit scopes. `ptsd validate` blocks commits on any pipeline violation.

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

## Usage

```bash
ptsd init                          # initialize .ptsd/ in current project
ptsd status                        # pipeline overview — who's where
ptsd task next                     # next task to work on
ptsd validate                      # check all pipeline gates
```

**Pipeline commands:**
```bash
ptsd feature list                  # all features and their stages
ptsd feature add <id> <title>      # register a new feature
ptsd seed add <feature-id>         # create seed data
ptsd bdd add <feature-id>          # write BDD scenarios
ptsd test map <feature-id>         # map tests to BDD
ptsd review <feature-id>           # run quality review (score 0-10)
```

**Agent mode** (for LLM-driven workflows):
```bash
ptsd status --agent                # compact output, no decoration
ptsd task next --agent             # machine-readable, file:line coords
ptsd validate --agent              # err:<category> format
```

## Project Structure

```
.ptsd/                             # all artifacts (git-tracked)
  ptsd.yaml                        # config
  features.yaml                    # feature registry (source of truth)
  state.yaml                       # hashes, scores, test results
  review-status.yaml               # per-feature review verdicts
  tasks.yaml                       # task queue
  issues.yaml                      # common issues registry
  docs/PRD.md                      # product requirements
  seeds/<id>/                      # golden seed data per feature
  bdd/<id>.feature                 # Gherkin scenarios per feature
  skills/                          # pipeline skills per stage
```

## Architecture

```
cmd/ptsd/main.go → internal/cli/* → internal/core/* → internal/yaml/*
                                   → internal/render/*
```

| Package | Responsibility | Rule |
|---|---|---|
| `core/` | Domain logic — pipeline, validation, state | Zero TUI imports |
| `render/` | Bubbletea output — tables, colors, agent mode | Zero domain logic |
| `cli/` | Glue — args &rarr; core &rarr; render | Thin as possible |
| `yaml/` | Hand-rolled YAML parser | Leaf package, no deps |

## Features

17 features across three tiers:

**Core** — the engine
| Feature | Description |
|---|---|
| `config` | Configuration system |
| `feature-mgmt` | Feature registry and lifecycle |
| `task-mgmt` | Task queue and assignment |
| `validate` | Pipeline gate enforcement |
| `git-hooks` | Commit scope validation |
| `status` | Project-wide pipeline overview |
| `output` | Dual render mode (human + agent) |
| `init` | Project initialization |

**Pipeline** — the stages
| Feature | Description |
|---|---|
| `prd-check` | PRD validation and completeness |
| `seed-mgmt` | Golden seed data management |
| `bdd-mgmt` | Gherkin scenario management |
| `test-integration` | Test runner and mapping |
| `state-tracking` | Hash-based regression detection |

**Advanced** — the extras
| Feature | Description |
|---|---|
| `review` | Quality scoring (0-10 scale) |
| `skills` | Pipeline skill generation |
| `adopt` | Bootstrap existing projects |
| `common-issues` | Recurring issue registry |

## Output Modes

**Human mode** (default) — interactive TUI with tables, colors, navigation via Bubbletea.

**Agent mode** (`--agent`) — ultra-compact, zero decoration, exact `file:line` coordinates. Designed for LLM consumption.

Error format in agent mode:
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
