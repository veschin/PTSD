// ===== Scroll Reveal =====
function initReveal() {
  if (!("IntersectionObserver" in window)) {
    document.querySelectorAll(".reveal").forEach(function(el) { el.classList.add("visible"); });
    return;
  }
  var obs = new IntersectionObserver(function(entries) {
    entries.forEach(function(e) {
      if (e.isIntersecting) { e.target.classList.add("visible"); obs.unobserve(e.target); }
    });
  }, { threshold: 0.08 });
  document.querySelectorAll(".reveal").forEach(function(el) { obs.observe(el); });
}

// ===== Nav scroll =====
function initNav() {
  var nav = document.getElementById("nav");
  if (!nav) return;
  window.addEventListener("scroll", function() {
    nav.classList.toggle("scrolled", window.scrollY > 10);
  }, { passive: true });
}

// ===== TOC =====
function initTOC() {
  var defs = [
    { id: "philosophy", label: "Philosophy" },
    { id: "pipeline", label: "Pipeline" },
    { id: "profiles", label: "Profiles" },
    { id: "enforcement", label: "Enforcement" },
    { id: "skills", label: "Skills" },
    { id: "integration", label: "Integration" },
    { id: "comparison", label: "Comparison" },
    { id: "start", label: "Start" }
  ];
  var toc = document.createElement("div");
  toc.className = "toc";
  defs.forEach(function(d) {
    var a = document.createElement("a");
    a.href = "#" + d.id; a.textContent = d.label; a.dataset.sec = d.id;
    toc.appendChild(a);
  });
  document.body.appendChild(toc);
  function checkWidth() { toc.style.display = window.innerWidth >= 1100 ? "block" : "none"; }
  checkWidth();
  window.addEventListener("resize", checkWidth);
  if (!("IntersectionObserver" in window)) return;
  var obs = new IntersectionObserver(function(entries) {
    entries.forEach(function(e) {
      var link = toc.querySelector('[data-sec="' + e.target.id + '"]');
      if (link) link.classList.toggle("active", e.isIntersecting);
    });
  }, { rootMargin: "-15% 0px -65% 0px" });
  defs.forEach(function(d) { var el = document.getElementById(d.id); if (el) obs.observe(el); });
}

// ===== Skills Browser (inline data) =====
var SKILL_DATA = {
  "adopt": `---
name: adopt
description: Use when bootstrapping an existing project into PTSD pipeline
---

## Instructions

1. Run ptsd adopt --name <name> in the project root.
2. PTSD creates .ptsd/ with config, features.yaml, and empty state.
3. Register existing features with realistic status values.
4. For each feature, assess current stage: which pipeline steps are complete.
5. Create seed data from existing tests or documentation.
6. Write BDD scenarios to capture existing behavior.
7. Do not rewrite working code -- document and track it.

## Common Mistakes

- Setting all features to the same stage instead of assessing each individually.
- Rewriting working code to fit ptsd patterns -- adopt tracks what exists.
- Forgetting to create seed data from existing test fixtures.
- Not checking for existing test files when setting the tests stage.
- Skipping BDD scenarios for features that already have passing tests.
`,
  "create-tasks": `---
name: create-tasks
description: Use when adding tasks to tasks.yaml
---

## Instructions

1. Every task must link to a feature via the feature field.
2. Use IDs in format T-<n>, incrementing from last existing ID.
3. Priority: A (urgent), B (normal), C (low).
4. Title must be a clear action: "Implement X", "Fix Y", "Add Z".
5. Add a checklist of subtasks if the task has multiple steps.
6. Status: TODO -> IN-PROGRESS -> DONE.

## Common Mistakes

- Creating tasks without a feature link -- every task must belong to a feature.
- Duplicate task IDs -- always check the last existing ID before creating.
- Vague titles like "Update code" instead of specific "Add validation to seed init".
- Missing priority field -- defaults are ambiguous, always set explicitly.
- Creating tasks for work already done -- check review-status.yaml first.
`,
  "review-bdd": `---
name: review-bdd
description: Use when reviewing BDD scenarios before advancing to tests stage
---

## Review Checklist

Score 0-10 based on how many items pass.

- [ ] One scenario per PRD acceptance criterion
- [ ] Happy path covered
- [ ] Error paths covered
- [ ] Edge cases from seed data used
- [ ] Each scenario independently runnable
- [ ] Gherkin syntax correct
- [ ] Feature tag present

Output: score and list of specific issues found.

## Common Mistakes

- Missing scenarios for error paths defined in the PRD.
- Scenarios that depend on execution order or shared state.
- Using abstract values instead of concrete seed data in Given steps.
- Accepting scenarios that test multiple behaviors in a single scenario.
`,
  "review-impl": `---
name: review-impl
description: Use when reviewing implementation after all tests pass
---

## Review Checklist

Score 0-10 based on how many items pass.

- [ ] All tests pass (go test ./...)
- [ ] No skipped or disabled tests
- [ ] No mock or stub patterns in implementation
- [ ] Error format: err:<category>
- [ ] No code outside the task scope
- [ ] Package boundaries respected (core/render/cli/yaml)
- [ ] No premature abstractions

Output: score and list of specific issues found.

## Common Mistakes

- Not running the full test suite -- a passing subset does not prove correctness.
- Accepting code that adds features beyond the task scope.
- Missing err:<category> prefix on error returns.
- Domain logic placed in cli/ or render/ instead of core/.
- Premature abstractions -- extracting helpers for code used only once.
`,
  "review-prd": `---
name: review-prd
description: Use when reviewing a PRD section before advancing to seed stage
---

## Review Checklist

Score 0-10 based on how many items pass.

- [ ] One-line summary present and accurate
- [ ] Problem statement is clear
- [ ] Acceptance criteria are testable
- [ ] Non-goals explicitly listed
- [ ] Edge cases covered
- [ ] No ambiguous language
- [ ] Feature anchor comment present

Output: score and list of specific issues found.

## Common Mistakes

- Accepting vague acceptance criteria ("works correctly") without pushing for specifics.
- Not verifying that edge cases cover empty, missing, and invalid inputs.
- Skipping non-goals check -- missing non-goals cause scope creep later.
- Giving a passing score when the feature anchor is missing.
`,
  "review-seed": `---
name: review-seed
description: Use when reviewing seed data before advancing to bdd stage
---

## Review Checklist

Score 0-10 based on how many items pass.

- [ ] seed.yaml has feature field
- [ ] At least one happy-path data file
- [ ] Edge case data present (empty, boundary, invalid)
- [ ] All files in manifest exist on disk
- [ ] Data is realistic (not placeholder values)
- [ ] File formats match what the feature consumes

Output: score and list of specific issues found.

## Common Mistakes

- Not checking that every file listed in seed.yaml actually exists on disk.
- Accepting placeholder data ("test", "example") as realistic.
- Missing boundary value data -- 0, max int, empty string, single element.
- Not verifying the data format matches what the implementation will parse.
`,
  "review-tests": `---
name: review-tests
description: Use when reviewing tests before advancing to impl stage
---

## Review Checklist

Score 0-10 based on how many items pass.

- [ ] One test per BDD scenario
- [ ] Test names match scenarios
- [ ] Real files used, no mocks
- [ ] Error messages checked (err:<category> prefix)
- [ ] Assertions on actual values not just err==nil
- [ ] t.TempDir() used for isolation
- [ ] Tests pass independently

Output: score and list of specific issues found.

## Common Mistakes

- Accepting tests that only check err == nil without verifying return values.
- Not verifying 1:1 mapping between BDD scenarios and test functions.
- Missing assertions on error message prefixes (err:validation, err:io, etc.).
- Tests that pass in sequence but fail when run individually (shared state).
`,
  "workflow": `---
name: workflow
description: Use at session start or when unsure what to do next
---

## Pipeline Profiles

Each feature has a pipeline profile. Check: \`ptsd feature show <id> --agent\`

| Profile | Stages |
|---------|--------|
| full | PRD -> Seed -> BDD -> Tests -> Impl |
| standard | PRD -> BDD -> Tests -> Impl |
| lite | PRD -> Tests -> Impl |

Context tells you what to do: \`ptsd context --agent\`
Follow the \`action=\` field. Skip stages not in the feature's profile.

### At each stage

| Stage | Create skill | Review skill |
|-------|-------------|--------------|
| PRD | write-prd | review-prd |
| Seed | write-seed | review-seed |
| BDD | write-bdd | review-bdd |
| Tests | write-tests | review-tests |
| Impl | write-impl | review-impl |

### Session protocol

1. Run \`ptsd context --agent\` -- see where each feature is and what to do next.
2. Pick the next feature/stage from the \`next:\` lines.
3. Apply the write-<stage> skill -> create artifacts.
4. Commit with \`[SCOPE] type: message\` format.
5. Run \`ptsd review <feature> <stage> <score>\` -- score 0-10, honest self-assessment.
6. Move to the next stage or feature.

### Stage cycle (repeat for every feature × every stage)

\`\`\`
write artifacts -> commit [SCOPE] -> ptsd review <feature> <stage> <score> -> next
\`\`\`

Do NOT skip the \`ptsd review\` step. It records review verdicts. Without it the feature stays \`review: pending\` forever.

### Gate rules

- No BDD without seed initialized (full pipeline only)
- No tests without BDD written (full/standard only)
- No impl without passing test review
- No stage advance without review score >= min_score (default 7)

## Common Mistakes

- Starting implementation without checking review-status.yaml first.
- Skipping the review skill after the create skill -- both are required at each stage.
- Forgetting to update review-status.yaml immediately after completing work.
- Working on a feature that is blocked by a gate (e.g., writing tests before BDD exists).
`,
  "write-bdd": `---
name: write-bdd
description: Use when writing Gherkin BDD scenarios for a feature
---

> **Profile gate:** Only for \`full\` and \`standard\` pipelines. Skip for \`lite\`. Check: \`ptsd context --agent\`

## Instructions

1. Write one scenario per acceptance criterion in the PRD.
2. Cover happy path, edge cases, and error paths.
3. Use seed data values in Given steps.
4. Each scenario must be independently runnable.
5. Use standard Gherkin: Given/When/Then. No And/But stacking.
6. Tag the feature: @feature:<id> at top of file.

## Common Mistakes

- Writing scenarios that depend on each other or share state.
- Missing error path scenarios -- every error condition in the PRD needs a scenario.
- Using abstract values instead of concrete seed data in Given steps.
- Stacking And/But steps instead of writing focused Given/When/Then.
- Forgetting the @feature:<id> tag -- ptsd uses this to link BDD to features.
`,
  "write-impl": `---
name: write-impl
description: Use when implementing code to make failing tests pass
---

## Instructions

1. Make each failing test pass, one at a time.
2. Write only the code required -- no speculative features.
3. Follow the project's package/module structure.
4. Error format: use the project's error format (e.g., err:<category> <message>).
5. No mocks in implementation. Use real I/O.
6. Run the configured test runner after each change.

## Common Mistakes

- Writing more code than the tests require -- no speculative features.
- Putting logic in the wrong layer -- follow the project's module structure.
- Printing errors instead of returning them with the project's error format.
- Adding dependencies that violate the project's dependency policy.
- Not running tests after each change -- catching failures early is cheaper.
- Forgetting to run \`ptsd review <feature> impl <score>\` after committing -- the feature stays unreviewed.
`,
  "write-prd": `---
name: write-prd
description: Use when creating or updating a PRD section for a feature
---

## Instructions

1. Start with a one-line summary of the feature purpose.
2. Define the problem being solved and who it affects.
3. List acceptance criteria as testable statements.
4. Define non-goals explicitly -- what is out of scope.
5. Cover edge cases: empty input, missing files, invalid state.
6. Add a feature anchor comment: <!-- feature:<id> -->
7. Keep language precise. No ambiguity.

## Common Mistakes

- Writing acceptance criteria that cannot be tested ("it should be fast", "user-friendly").
- Forgetting non-goals -- every PRD must state what is NOT in scope.
- Missing edge cases for empty/missing/invalid inputs.
- Using vague language: "should handle errors" instead of "returns err:validation when input is empty".
- Omitting the feature anchor comment -- without it, ptsd cannot link the PRD section to the feature.
`,
  "write-seed": `---
name: write-seed
description: Use when creating golden seed data for a feature
---

> **Profile gate:** Only for \`full\` pipeline. Skip for \`standard\` and \`lite\`. Check: \`ptsd context --agent\`

## Instructions

1. Create seed.yaml with feature field and files list.
2. Include at least one happy-path data file.
3. Include edge-case data: empty collections, boundary values, invalid inputs.
4. Use realistic data -- not "test" or "foo".
5. Every file referenced in seed.yaml must exist on disk.
6. Formats: JSON, YAML, or CSV depending on what the feature consumes.

## Common Mistakes

- Using placeholder data ("test", "foo", "bar") instead of realistic values.
- Listing files in seed.yaml that do not exist on disk.
- Only covering the happy path -- missing empty, boundary, and invalid cases.
- Creating seed data that does not match the format the feature actually consumes.
- Forgetting the feature field in seed.yaml -- ptsd cannot link it without this.
`,
  "write-tests": `---
name: write-tests
description: Use when writing tests from BDD scenarios for a feature
---

> For \`lite\` pipeline: write tests directly from PRD (no BDD). Map with: \`ptsd test map --feature <id> <test-file>\`
> For \`standard\`/\`full\`: write tests from BDD scenarios. Map with: \`ptsd test map <bdd-file> <test-file>\`

## Instructions

1. One test function per BDD scenario, named after the scenario.
2. Use real files in temp directories -- no mocks for internal code.
3. Assert exact output values, not just "no error".
4. Test error paths: verify error message prefix (err:<category>).
5. Use t.TempDir() for isolation.
6. No test helpers that obscure what is being tested.

## Common Mistakes

- Asserting only err == nil instead of checking actual return values.
- Using mocks for internal code -- ptsd requires real files and real I/O.
- Test names that do not match BDD scenario names -- breaks traceability.
- Sharing state between tests via package-level variables.
- Forgetting to test error message prefixes (err:validation, err:io, etc.).
`
};
var skillIndex = 0;
var skillNames = [];

function highlightSkill(raw) {
  var escaped = raw
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

  // Split into sections to apply context-aware coloring
  var inMistakes = false;
  var lines = escaped.split('\n').map(function(line) {
    // Track if we're in Common Mistakes section
    if (/^#{1,3} .*[Mm]istakes/.test(line)) inMistakes = true;
    else if (/^#{1,3} /.test(line)) inMistakes = false;

    // Frontmatter
    if (line === '---') return '<span class="cl-c">' + line + '</span>';
    // Headings
    if (/^#{1,3} /.test(line)) return '<span class="cl-k">' + line + '</span>';
    // Blockquotes
    if (/^&gt;/.test(line)) return '<span class="cl-k">' + line + '</span>';
    // Checklist items
    if (/^- \[[ x]\]/.test(line)) return '<span class="cl-t">' + line + '</span>';
    // Bullet items: red only in mistakes section
    if (/^- /.test(line)) return '<span class="' + (inMistakes ? 'cl-e' : 'cl-s') + '">' + line + '</span>';
    // Numbered instructions
    if (/^\d+\./.test(line)) return '<span class="cl-s">' + line + '</span>';
    // Frontmatter key: value
    if (/^[a-z]+:/.test(line)) return '<span class="cl-c">' + line + '</span>';
    return line;
  });

  return lines.join('\n').replace(/`([^`]+)`/g, '<code>$1</code>');
}

function showSkill(i) {
  var view = document.getElementById("skill-view");
  var label = document.getElementById("skill-label");
  var dots = document.querySelectorAll(".skill-dot");
  if (!view || !skillNames[i]) return;

  skillIndex = i;
  label.textContent = skillNames[i];
  view.innerHTML = highlightSkill(SKILL_DATA[skillNames[i]]);
  dots.forEach(function(d, j) { d.classList.toggle("active", j === i); });
}

function initSkills() {
  var view = document.getElementById("skill-view");
  var dotsEl = document.getElementById("skill-dots");
  var prev = document.getElementById("skill-prev");
  var next = document.getElementById("skill-next");
  if (!view || !dotsEl) return;

  skillNames = Object.keys(SKILL_DATA);

  // build dots
  skillNames.forEach(function(_, i) {
    var dot = document.createElement("button");
    dot.className = "skill-dot";
    dot.addEventListener("click", function() { showSkill(i); });
    dotsEl.appendChild(dot);
  });

  prev.addEventListener("click", function() {
    showSkill((skillIndex - 1 + skillNames.length) % skillNames.length);
  });
  next.addEventListener("click", function() {
    showSkill((skillIndex + 1) % skillNames.length);
  });

  showSkill(0);
}

// ===== Init =====
document.addEventListener("DOMContentLoaded", function() {
  initReveal();
  initNav();
  initTOC();
  initSkills();
});
