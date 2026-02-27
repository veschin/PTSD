package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMapBDDToTestFile(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(filepath.Join(ptsdDir, "bdd"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0755); err != nil {
		t.Fatal(err)
	}

	bddFile := ".ptsd/bdd/user-auth.feature"
	testFile := "tests/auth.test.ts"
	bddPath := filepath.Join(dir, bddFile)
	testPath := filepath.Join(dir, testFile)

	bddContent := `@feature:user-auth
Feature: User Auth
  Scenario: Login
    Given user exists
`
	if err := os.WriteFile(bddPath, []byte(bddContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testPath, []byte("// test file"), 0644); err != nil {
		t.Fatal(err)
	}

	stateYAML := `features: {}
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	err := MapTest(dir, bddFile, testFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	state, err := os.ReadFile(filepath.Join(ptsdDir, "state.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	stateStr := string(state)
	if !strings.Contains(stateStr, "user-auth") || !strings.Contains(stateStr, bddFile) || !strings.Contains(stateStr, testFile) {
		t.Errorf("expected state.yaml to contain mapping from %q to %q for feature user-auth, got:\n%s", bddFile, testFile, stateStr)
	}
}

func TestCheckTestCoverage(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	bddDir := filepath.Join(ptsdDir, "bdd")
	if err := os.MkdirAll(bddDir, 0755); err != nil {
		t.Fatal(err)
	}

	bddContent := `@feature:user-auth
Feature: User Auth
  Scenario: Login
  Scenario: Logout
  Scenario: Register
`
	if err := os.WriteFile(filepath.Join(bddDir, "user-auth.feature"), []byte(bddContent), 0644); err != nil {
		t.Fatal(err)
	}

	stateYAML := `features:
  user-auth:
    tests:
      - .ptsd/bdd/user-auth.feature::tests/auth.test.ts
      - .ptsd/bdd/user-auth.feature::tests/auth2.test.ts
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	coverage, err := CheckTestCoverage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(coverage) == 0 {
		t.Fatal("expected coverage entries")
	}
	found := false
	for _, c := range coverage {
		if c.Feature == "user-auth" {
			found = true
			if c.Status != "partial" {
				t.Errorf("expected partial coverage, got %s", c.Status)
			}
		}
	}
	if !found {
		t.Error("expected user-auth in coverage")
	}
}

func TestCheckTestCoverageNoTests(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	bddDir := filepath.Join(ptsdDir, "bdd")
	if err := os.MkdirAll(bddDir, 0755); err != nil {
		t.Fatal(err)
	}

	bddContent := `@feature:orphan
Feature: Orphan
  Scenario: Something
`
	if err := os.WriteFile(filepath.Join(bddDir, "orphan.feature"), []byte(bddContent), 0644); err != nil {
		t.Fatal(err)
	}

	stateYAML := `features: {}
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	coverage, err := CheckTestCoverage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, c := range coverage {
		if c.Feature == "orphan" && c.Status != "no-tests" {
			t.Errorf("expected no-tests status, got %s", c.Status)
		}
	}
}

func TestRunTestsForFeature(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0755); err != nil {
		t.Fatal(err)
	}

	// Script echoes its arguments so we can verify filter was applied
	testScript := `#!/bin/sh
echo "# args: $@"
echo "ok 1 - test pass"
echo "not ok 2 - test fail"
echo "# Failed at tests/auth.test.ts:42"
exit 1
`
	testPath := filepath.Join(dir, "tests", "run.sh")
	if err := os.WriteFile(testPath, []byte(testScript), 0755); err != nil {
		t.Fatal(err)
	}

	configYAML := `project:
  name: TestApp
testing:
  runner: ./tests/run.sh
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	stateYAML := `features:
  user-auth:
    tests:
      - .ptsd/bdd/user-auth.feature::tests/auth.test.ts
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := RunTests(dir, "user-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.Total == 0 {
		t.Error("expected some tests to run")
	}
	if len(results.Failures) == 0 && results.Failed > 0 {
		t.Error("expected failure details when tests fail")
	}
}

func TestRunAllTests(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0755); err != nil {
		t.Fatal(err)
	}

	// Script echoes args; with no filter, no args should be passed
	testScript := `#!/bin/sh
echo "# args: $@"
echo "ok 1 - pass"
echo "ok 2 - pass"
exit 0
`
	testPath := filepath.Join(dir, "tests", "run.sh")
	if err := os.WriteFile(testPath, []byte(testScript), 0755); err != nil {
		t.Fatal(err)
	}

	configYAML := `project:
  name: TestApp
testing:
  runner: ./tests/run.sh
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	stateYAML := `features:
  user-auth:
    tests:
      - .ptsd/bdd/user-auth.feature::tests/auth.test.ts
  data-sync:
    tests:
      - .ptsd/bdd/data-sync.feature::tests/sync.test.ts
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := RunTests(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.Total == 0 {
		t.Error("expected tests to run for all features")
	}
}

func TestRunTestsFilterPassesOnlyFeatureFiles(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0755); err != nil {
		t.Fatal(err)
	}

	// Script writes received arguments to a file for verification
	testScript := `#!/bin/sh
echo "$@" > "` + filepath.Join(dir, "received_args.txt") + `"
echo "ok 1 - pass"
exit 0
`
	if err := os.WriteFile(filepath.Join(dir, "tests", "run.sh"), []byte(testScript), 0755); err != nil {
		t.Fatal(err)
	}

	configYAML := `project:
  name: TestApp
testing:
  runner: ./tests/run.sh
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Two features with different test files
	stateYAML := `features:
  user-auth:
    tests:
      - .ptsd/bdd/user-auth.feature::tests/auth.test.ts
  data-sync:
    tests:
      - .ptsd/bdd/data-sync.feature::tests/sync.test.ts
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Run with feature filter "user-auth"
	_, err := RunTests(dir, "user-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsBytes, err := os.ReadFile(filepath.Join(dir, "received_args.txt"))
	if err != nil {
		t.Fatal(err)
	}
	args := strings.TrimSpace(string(argsBytes))

	// Should contain only user-auth's test file, not data-sync's
	if !strings.Contains(args, "tests/auth.test.ts") {
		t.Errorf("expected runner to receive tests/auth.test.ts, got: %q", args)
	}
	if strings.Contains(args, "tests/sync.test.ts") {
		t.Errorf("runner should NOT receive tests/sync.test.ts when filtering by user-auth, got: %q", args)
	}
}

func TestRunTestsNoFilterRunsAll(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0755); err != nil {
		t.Fatal(err)
	}

	// Script writes received arguments to a file for verification
	testScript := `#!/bin/sh
echo "$@" > "` + filepath.Join(dir, "received_args.txt") + `"
echo "ok 1 - pass"
exit 0
`
	if err := os.WriteFile(filepath.Join(dir, "tests", "run.sh"), []byte(testScript), 0755); err != nil {
		t.Fatal(err)
	}

	configYAML := `project:
  name: TestApp
testing:
  runner: ./tests/run.sh
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	stateYAML := `features:
  user-auth:
    tests:
      - .ptsd/bdd/user-auth.feature::tests/auth.test.ts
  data-sync:
    tests:
      - .ptsd/bdd/data-sync.feature::tests/sync.test.ts
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	// Run without feature filter — should run all
	_, err := RunTests(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	argsBytes, err := os.ReadFile(filepath.Join(dir, "received_args.txt"))
	if err != nil {
		t.Fatal(err)
	}
	args := strings.TrimSpace(string(argsBytes))

	// No filter = no file arguments appended to runner
	if args != "" {
		t.Errorf("expected no arguments when running without filter, got: %q", args)
	}
}

func TestResultsUpdateState(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0755); err != nil {
		t.Fatal(err)
	}

	testScript := `#!/bin/sh
echo "ok 1 - pass"
echo "ok 2 - pass"
echo "ok 3 - pass"
echo "ok 4 - pass"
echo "ok 5 - pass"
echo "not ok 6 - fail"
echo "# Failed at tests/auth.test.ts:99"
exit 1
`
	testPath := filepath.Join(dir, "tests", "run.sh")
	if err := os.WriteFile(testPath, []byte(testScript), 0755); err != nil {
		t.Fatal(err)
	}

	configYAML := `project:
  name: TestApp
testing:
  runner: ./tests/run.sh
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	stateYAML := `features:
  user-auth:
    tests:
      - .ptsd/bdd/user-auth.feature::tests/auth.test.ts
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := RunTests(dir, "user-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results.Passed != 5 {
		t.Errorf("expected 5 passed, got %d", results.Passed)
	}
	if results.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", results.Failed)
	}

	stateBytes, err := os.ReadFile(filepath.Join(ptsdDir, "state.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	state := string(stateBytes)
	if !strings.Contains(state, "passed:") || !strings.Contains(state, "failed:") {
		t.Errorf("expected state.yaml to contain test results (passed/failed counts) for user-auth, got:\n%s", state)
	}
}

func TestNoTestRunnerConfigured(t *testing.T) {
	dir := t.TempDir()
	ptsdDir := filepath.Join(dir, ".ptsd")
	if err := os.MkdirAll(ptsdDir, 0755); err != nil {
		t.Fatal(err)
	}

	configYAML := `project:
  name: TestApp
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "ptsd.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatal(err)
	}

	stateYAML := `features:
  user-auth:
    tests:
      - .ptsd/bdd/user-auth.feature::tests/auth.test.ts
`
	if err := os.WriteFile(filepath.Join(ptsdDir, "state.yaml"), []byte(stateYAML), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := RunTests(dir, "user-auth")
	if err == nil {
		t.Fatal("expected error for missing test runner")
	}
	if err.Error() != "err:config no test runner configured" {
		t.Errorf("expected 'err:config no test runner configured', got: %v", err)
	}
}

