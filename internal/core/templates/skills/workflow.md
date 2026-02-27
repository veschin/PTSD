---
name: workflow
description: Use at session start or when unsure what to do next
---

## Pipeline Order (mandatory, no skipping)

PRD → Seed → BDD → Tests → Implementation

### At each stage

| Stage | Create skill | Review skill |
|-------|-------------|--------------|
| PRD | write-prd | review-prd |
| Seed | write-seed | review-seed |
| BDD | write-bdd | review-bdd |
| Tests | write-tests | review-tests |
| Impl | write-impl | review-impl |

### Session protocol

1. Read .ptsd/review-status.yaml — find where each feature is.
2. Pick the next feature/stage with review: pending or tests: absent.
3. Apply the appropriate skill from the table above.
4. Record progress immediately in review-status.yaml after each action.
5. Run ptsd validate --agent before committing.
6. Commit with [SCOPE] type: message format.

### Gate rules

- No BDD without seed initialized
- No tests without BDD written
- No impl without passing test review
- No stage advance without review score >= min_score (default 7)

## Common Mistakes

- Starting implementation without checking review-status.yaml first.
- Skipping the review skill after the create skill — both are required at each stage.
- Forgetting to update review-status.yaml immediately after completing work.
- Working on a feature that is blocked by a gate (e.g., writing tests before BDD exists).
