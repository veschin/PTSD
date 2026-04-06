# CLAUDE.md -- Maintainer System Prompt

You are maintaining PTSD (PRD-Test-Seed Dashboard) -- a Go CLI that enforces structured AI-driven development. Single binary, zero deps (stdlib only, no `go.sum`). This project dogfoods itself.

## Hard Constraints

- **stdlib only.** No third-party imports. No `go.sum`. No exceptions.
- **No YAML library.** All YAML is hand-parsed: `strings.Split`/`HasPrefix`/`TrimPrefix`. Every new YAML field = manual parser update.
- **No mocks in tests.** Real files, real CLI, temp directories. `setupProjectWithFeatures(t, "id:status")` is the central helper.
- **Error format: `err:<category> <message>`.** Categories: validation, pipeline, user, config, io, test. `coreError()` in `cli/helpers.go` routes to exit codes 1-5.
- **Commit format: `[SCOPE] type: message`.** Scopes: PRD, SEED, BDD, TEST, IMPL, TASK, STATUS.

## Build & Test

```bash
go build ./cmd/ptsd/...      # must always pass
go test ./...                # must always pass
go test -run TestName ./path # single test
```

## Architecture

```
cmd/ptsd/main.go      flat switch dispatcher, no cobra/flag
  -> internal/cli/*    args -> core -> render, func RunX(args, agentMode) int
  -> internal/core/*   domain logic, zero TUI imports
  -> internal/render/* AgentRenderer only (HumanRenderer not yet built)
```

### Key Files -- Know These

| File | What It Does | When You Touch It |
|------|-------------|-------------------|
| `core/profiles.go` | Pipeline profiles: full/standard/lite. `StageRequired()`, `NextStage()`, `NextAction()`, `ResolveFeaturePipeline()` | Adding/changing pipeline stages |
| `core/registry.go` | `Feature{ID,Title,Status,Pipeline}`, CRUD on `features.yaml`, `loadFeatures()`/`saveFeatures()` | Changing feature fields |
| `core/gatecheck.go` | `GateCheck()` -- blocks invalid AI writes. Profile-aware. `alwaysAllowed` map, `isImplFile()` | Adding allowed file paths, changing gates |
| `core/context.go` | `BuildContext()` -- emits next/blocked lines with pipeline. Drives AI behavior at session start | Changing what AI sees |
| `core/autotrack.go` | `AutoTrack()` -- PostToolUse hook, advances stage. `stageOrder` map lives here | Changing stage advancement logic |
| `core/pipeline.go` | `Validate()` -- pre-commit checks. Profile-aware. `scanForMocks()` | Adding validation rules |
| `core/testdetect.go` | `IsTestFile()`, `StripTestSuffix()`, `DefaultTestPatterns()`. **Single source of truth** for test patterns across all languages | Adding language support |
| `core/init.go` | `InitProject()`, `ReInitProject()`, `ReInitProjectWithTool()`. Tool adapters: claude/opencode/generic. `detectTool()`, `detectTestRunner()` | Adding tool adapters, changing init |
| `core/state.go` | `State`/`FeatureState`, hashes, scores, `CheckRegressions()`, `ComputeStageFromArtifacts()` | Changing state tracking |
| `core/review.go` | `RecordReview()`, `CheckReviewGate()`. Validates stage against feature's pipeline profile | Changing review logic |
| `core/templates.go` | `//go:embed templates/*`. `renderTemplate()`, `readTemplate()`. All templates auto-embedded | Adding templates |
| `core/migrate.go` | `MigrateProject()` -- adds pipeline/tool fields to old projects | Changing config format |

### Adding a New Command

1. `case "cmd":` in `cmd/ptsd/main.go`
2. `func RunCmd(args []string, agentMode bool) int` in `cli/cmd.go`
3. Core logic in `core/cmd.go`
4. Add to `cli/help.go`

### Adding a New Tool Adapter

1. Generator function in `core/init.go` (e.g., `generateCursorConfig()`)
2. Template files in `core/templates/<tool>/`
3. Wire into both `switch tool` blocks in `InitProject` and `ReInitProjectWithTool`
4. Add to `detectTool()` auto-detection

## Pipeline Profiles -- The Core Concept

Every feature has a pipeline profile that determines required stages:

| Profile | Stages | Default |
|---------|--------|---------|
| `full` | PRD -> Seed -> BDD -> Tests -> Impl | No |
| `standard` | PRD -> BDD -> Tests -> Impl | **Yes** |
| `lite` | PRD -> Tests -> Impl | No |

**Profile drives EVERYTHING:** gates, context output, validation, review, autotrack, stage computation. When you add logic that checks seed/bdd/test prerequisites, it MUST call `StageRequired(pipeline, stage)` or it will break non-full pipelines.

`Feature.Pipeline` field in `registry.go`. Empty = config default (`ptsd.yaml` -> `pipeline.default`). Config default empty = `"standard"`.

## Tool Adapters

`ptsd init --tool <name>` generates tool-specific AI integration:

| Tool | What's Generated | Enforcement |
|------|-----------------|-------------|
| `claude` | `.claude/settings.json` hooks, `.claude/skills/*/SKILL.md`, `CLAUDE.md` | Full: gate, track, context |
| `opencode` | `.opencode/plugins/ptsd.ts`, `.opencode/commands/*.md`, `AGENTS.md` | Full: gate, track, context |
| `generic` | `AGENTS.md` only | Advisory only (git hooks still enforce) |

Auto-detection: `.claude/` exists -> claude, `.opencode/` exists -> opencode, else -> claude default.

`ptsd init` on existing project = reinit: regenerates hooks/skills/instructions, preserves data. `ReInitProjectWithTool(dir, tool)` accepts explicit tool override for switching tools.

## Hooks Architecture

**Claude Code**: shell scripts in `.claude/hooks/`, wired via `.claude/settings.json`:
- `SessionStart` -> `ptsd context --agent` (context injection, runs ONCE per session)
- `PreToolUse` (Edit|Write) -> `ptsd gate-check` (blocks invalid writes)
- `PostToolUse` (Edit|Write) -> `ptsd auto-track` (advances stage)

**OpenCode**: TypeScript plugin in `.opencode/plugins/ptsd.ts`:
- `session.start` -> context injection
- `tool.execute.before` -> gate check
- `tool.execute.after` -> auto-track

**Git hooks** (all tools): `pre-commit` -> `ptsd validate`, `commit-msg` -> scope validation.

Hook scripts use `{{.Bin}}` template -- renders to ptsd binary path at init time.

## Polyglot Test Detection

`core/testdetect.go` is the **single source of truth**. All test file checks use `IsTestFile()` -- never inline suffix checks.

Supported: `_test.go`, `.test.ts`, `.test.js`, `.test.tsx`, `.test.jsx`, `.spec.ts`, `.spec.js`, `_test.py`, `test_*.py`, `_spec.rb`, `_test.rs`, `Test.java`, `Tests.cs`.

`DefaultTestPatterns(runner)` returns appropriate globs per detected runner. `config.go:applyDefaults` uses this when patterns are empty.

**Trap:** If you add a new test pattern to `knownTestSuffixes`, ALL consumers automatically pick it up (gatecheck, autotrack, pipeline, hooks, adopt). But `DefaultTestPatterns()` also needs updating for the corresponding runner.

## Testing Patterns

```go
// Central helper -- creates temp dir with .ptsd/ structure
dir := setupProjectWithFeatures(t, "auth:in-progress", "config:planned")

// Integration tests -- build real binary, run in temp dir
cmd := exec.Command(bin, "validate", "--agent")
cmd.Dir = dir
```

`assertHasError(t, errors, feature, category, contains)` -- searches error slice.

**Critical:** When testing profile-dependent behavior, set `pipeline: full` in features.yaml explicitly. Default is `standard` which skips seed/bdd checks.

## State Tracking

`state.yaml` stores per-feature: stage, hashes (SHA256 of PRD/seed/bdd/test files), scores, test mappings.

Test mappings: `<bdd-file>::<test-file>` or `feature:<id>::<test-file>` (lite pipeline, no BDD).

`CheckRegressions()` compares stored vs current hashes. PRD change = stage downgrade. Seed/BDD change = warning only.

## Release

```bash
go test ./... && go build ./cmd/ptsd/...
git tag vX.Y.Z && git push origin vX.Y.Z
```

- **NEVER delete/move a pushed tag** -- Go proxy caches are immutable
- Minor bump for features, patch for fixes
- Update pinned version in README.md

## Landmines

1. **`loadFeatures()` is called from many places** -- gatecheck, context, validation, adopt, review. It re-reads `features.yaml` each time. Don't assume caching.
2. **`stageOrder` map in `autotrack.go`** is used by context, review, and state. Changing it affects stage comparison everywhere.
3. **Template changes require rebuild** -- `//go:embed` bakes templates into binary at compile time.
4. **`review-status.yaml` is gate-blocked** -- AI cannot edit it directly. Only `ptsd review` and `AutoTrack()` write it.
5. **`ptsd init` reinit path** returns early before the fresh-init code. Changes to fresh init don't affect reinit unless you also update `ReInitProjectWithTool`.
6. **`adopt` uses `deferred` status** for test-discovered features. `deferred` = excluded from all pipeline checks. User activates with `feature status <id> in-progress`.
7. **Empty `Pipeline` field on Feature** = resolve from config default. Don't assume it's always set.

<!-- ---ptsd--- -->
# Claude Agent Instructions

## Authority Hierarchy (ENFORCED BY HOOKS)

PTSD (iron law) > User (context provider) > Assistant (executor)

- PTSD decides what CAN and CANNOT be done. Pipeline, gates, validation -- non-negotiable.
  Hooks enforce this automatically -- writes that violate pipeline are BLOCKED.
- User provides context and requirements. User also follows ptsd rules.
- Assistant executes within ptsd constraints. Writes code, docs, tests on behalf of user.

## Session Start Protocol

EVERY session, BEFORE any work:
1. Run: ptsd context --agent -- see full pipeline state
2. Run: ptsd task next --agent -- get next task
3. Follow output exactly.

## Commands (always use --agent flag)

- ptsd context --agent              -- full pipeline state (auto-injected by hooks)
- ptsd status --agent               -- project overview
- ptsd task next --agent            -- next task to work on
- ptsd task update <id> --status WIP -- mark task in progress
- ptsd validate --agent             -- check pipeline before commit
- ptsd feature list --agent         -- list all features
- ptsd seed init <id> --agent       -- initialize seed directory
- ptsd gate-check --file <path> --agent -- check if file write is allowed
- ptsd test map --feature <id> <test-file> -- map test without BDD (for lite pipeline)
- ptsd feature pipeline <id> <profile> -- change feature pipeline
- ptsd migrate --agent            -- migrate project to current version

## Skills

PTSD pipeline skills are in `.claude/skills/` -- auto-loaded when relevant.

| Skill | When to Use |
|-------|------------|
| write-prd | Creating or updating a PRD section |
| write-seed | Creating seed data for a feature |
| write-bdd | Writing Gherkin BDD scenarios |
| write-tests | Writing tests from BDD scenarios |
| write-impl | Implementing to make tests pass |
| create-tasks | Adding tasks to tasks.yaml |
| review-prd | Reviewing PRD before advancing to seed |
| review-seed | Reviewing seed data before advancing to bdd |
| review-bdd | Reviewing BDD before advancing to tests |
| review-tests | Reviewing tests before advancing to impl |
| review-impl | Reviewing implementation after tests pass |
| workflow | Session start or when unsure what to do next |
| adopt | Bootstrapping existing project into PTSD |

Use the corresponding write skill, then review skill at each pipeline stage.

Note: write-seed is only required for full pipeline. write-bdd is required for full and standard pipelines. Lite pipeline skips both -- write tests directly from PRD.

## Pipeline Profiles

Each feature has a pipeline profile that determines required stages:

| Profile | Stages | Use For |
|---------|--------|---------|
| full | PRD -> Seed -> BDD -> Tests -> Impl | Complex, data-heavy features |
| standard | PRD -> BDD -> Tests -> Impl | Default. Most features |
| lite | PRD -> Tests -> Impl | Simple utilities, config |

Check feature pipeline: `ptsd feature show <id> --agent`
Change pipeline: `ptsd feature pipeline <id> full|standard|lite`

Each required stage needs review score >= 7 before advancing.
Hooks enforce gates automatically -- blocked writes show the reason.

## Rules

- NO mocks for internal code. Real tests, real files, temp directories.
- NO garbage files. Every file must link to a feature.
- NO hiding errors. Explain WHY something failed.
- NO over-engineering. Minimum code for the current task.
- ALWAYS run: ptsd validate --agent before committing.
- COMMIT FORMAT: [SCOPE] type: message
  Scopes: PRD, SEED, BDD, TEST, IMPL, TASK, STATUS
  Types: feat, add, fix, refactor, remove, update

## Troubleshooting

When ptsd status/validate shows unexpected results, debug with these steps:

| Symptom | Cause | Fix |
|---------|-------|-----|
| TESTS:0 but test files exist | Tests not mapped to features | `ptsd test map .ptsd/bdd/<id>.feature <test-file>` or `ptsd test map --feature <id> <test-file>` (lite pipeline) |
| BDD:0 but .feature files exist | State hashes empty, SyncState not run | `ptsd status --agent` triggers sync; if still 0, check `.ptsd/bdd/<id>.feature` has `@feature:<id>` tag on line 1 |
| Feature stuck at wrong stage | review-status.yaml stale or stage not advanced | Run `ptsd review <id> <stage> <score>` to advance; check `ptsd context --agent` for blockers |
| "no test files mapped" on `ptsd test run` | Test mapping missing in state.yaml | `ptsd test map .ptsd/bdd/<id>.feature <test-file>` or `--feature <id> <test-file>` |
| Gate blocks file write | File not in allowed list for current stage | Check `ptsd gate-check --file <path> --agent`; advance feature to correct stage first |
| Validate shows "mock detected" | Test file contains mock/stub patterns | Replace mocks with real file-based tests in temp directories |
| Regression warning on status | Artifact file changed after stage was reviewed | Re-review the stage: `ptsd review <id> <stage> <score>` |

### Debug flow
1. `ptsd context --agent` -- shows next action, blockers, stage per feature
2. `ptsd feature show <id> --agent` -- shows artifact counts and test stats
3. `ptsd validate --agent` -- shows all pipeline violations
4. Check `.ptsd/state.yaml` -- hashes, test mappings, stages
5. Check `.ptsd/review-status.yaml` -- review verdicts per feature

### Test mapping
Features need test files mapped to track results:
- Standard/full pipeline: `ptsd test map .ptsd/bdd/<id>.feature <test-file>` (reads @feature tag from BDD)
- Lite pipeline (no BDD): `ptsd test map --feature <id> <test-file>` (direct mapping)
Without mapping, ptsd cannot track test results per feature.

## Forbidden

- Mocking internal code
- Skipping pipeline steps
- Hiding errors or pretending something works
- Generating files not linked to a feature
- Using --force, --skip-validation, --no-verify

<!-- ---ptsd--- -->
