---
name: adopt
description: Use when bootstrapping an existing project into PTSD pipeline
---

## Instructions

1. Run `ptsd adopt --dry-run` first to see what will be discovered (BDD files, test files, Go packages). Review the output before proceeding.
2. Run `ptsd adopt` to create `.ptsd/` structure and register discovered features.
3. Run `ptsd status --agent` to verify what was created.

## Post-Adopt Checklist

After adopt completes, work through these steps for each discovered feature:

1. **Assess status.** Run `ptsd feature list --agent`. Features from BDD files start as `in-progress`. Features from test files start as `deferred` -- activate them with `ptsd feature status <id> in-progress` when ready.
2. **Choose pipeline profile.** Default is `standard` (PRD -> BDD -> Tests -> Impl). Use `ptsd feature pipeline <id> lite` for simple features that already have tests but no BDD. Use `full` for data-heavy features that need seed data.
3. **Map tests to features.** For each test file, run `ptsd test map --feature <id> <test-file>` (lite pipeline) or `ptsd test map <bdd-file> <test-file>` (standard/full). Without mapping, PTSD cannot track test results per feature.
4. **Write PRD anchors.** Every feature needs a section in `.ptsd/docs/PRD.md` with `<!-- feature:<id> -->` anchor. Document what the feature does, acceptance criteria, edge cases.
5. **Verify.** Run `ptsd validate --agent` -- fix any pipeline violations before committing.

## When the project has no tests

If `--dry-run` finds no test files:
- Adopt creates empty `.ptsd/` structure with no features
- Add features manually: `ptsd feature add <id> "Title"`
- Start from PRD stage -- the pipeline will guide you through BDD and tests

## Working with deferred features

Features discovered from test files get status `deferred` and pipeline `lite`. This means:
- They are excluded from pipeline checks until activated
- Activate when ready: `ptsd feature status <id> in-progress`
- The `lite` pipeline skips BDD -- write tests directly from PRD

## Common Mistakes

- Running `ptsd adopt` without `--dry-run` first -- always preview what will be discovered.
- Forgetting to map test files -- without mapping, `ptsd status` shows TESTS:0 even if tests exist.
- Setting all features to the same pipeline instead of assessing each individually.
- Not writing PRD anchors -- features without PRD sections will block at the first gate.
- Rewriting working code to fit PTSD patterns -- adopt tracks what exists, it does not require refactoring.
