package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddBDDWithSeed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")
	if err := InitSeed(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}

	err := AddBDD(dir, "user-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bddPath := filepath.Join(dir, ".ptsd", "bdd", "user-auth.feature")
	if _, err := os.Stat(bddPath); os.IsNotExist(err) {
		t.Fatal("user-auth.feature not created")
	}

	data, err := os.ReadFile(bddPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "@feature:user-auth") {
		t.Error("feature file missing @feature tag")
	}
}

func TestAddBDDWithoutSeed(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")

	err := AddBDD(dir, "user-auth")
	if err == nil {
		t.Fatal("expected error for missing seed")
	}
	if !strings.HasPrefix(err.Error(), "err:pipeline") {
		t.Errorf("expected err:pipeline, got: %v", err)
	}
}

func TestCheckBDD(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress", "catalog:in-progress")
	if err := InitSeed(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}
	if err := InitSeed(dir, "catalog"); err != nil {
		t.Fatal(err)
	}
	if err := AddBDD(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}

	missing, err := CheckBDD(dir)
	if err == nil {
		t.Fatal("expected error for missing BDD")
	}
	if len(missing) != 1 || missing[0] != "catalog" {
		t.Errorf("expected [catalog] missing, got: %v", missing)
	}
	if !strings.HasPrefix(err.Error(), "err:pipeline") {
		t.Errorf("expected err:pipeline, got: %v", err)
	}
}

func TestCheckBDDUnknownFeatureTag(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")
	if err := InitSeed(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}
	if err := AddBDD(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}

	bddDir := filepath.Join(dir, ".ptsd", "bdd")
	ghostFeature := `@feature:ghost
Feature: Ghost
  Scenario: X
    Given Y
`
	if err := os.WriteFile(filepath.Join(bddDir, "ghost.feature"), []byte(ghostFeature), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := CheckBDD(dir)
	if err == nil {
		t.Fatal("expected error for unknown feature tag")
	}
	if !strings.HasPrefix(err.Error(), "err:validation") {
		t.Errorf("expected err:validation, got: %v", err)
	}
}

func TestShowBDDCompact(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")
	if err := InitSeed(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}
	if err := AddBDD(dir, "user-auth"); err != nil {
		t.Fatal(err)
	}

	bddPath := filepath.Join(dir, ".ptsd", "bdd", "user-auth.feature")
	content := `@feature:user-auth
Feature: User Auth
  Scenario: Login success
    Given user exists
    When I login
    Then I see dashboard

  Scenario: Login fail
    Given invalid creds
    When I login
    Then error shown

  Scenario: Logout
    Given logged in
    When I logout
    Then redirected
`
	if err := os.WriteFile(bddPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := ShowBDD(dir, "user-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 3 {
		t.Errorf("expected 3 scenario lines, got %d", len(lines))
	}
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			t.Errorf("line not in compact format: %s", line)
		}
	}
}

func TestShowBDDNotFound(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")

	_, err := ShowBDD(dir, "ghost")
	if err == nil {
		t.Fatal("expected error for nonexistent feature")
	}
	if !strings.HasPrefix(err.Error(), "err:validation") {
		t.Errorf("expected err:validation, got: %v", err)
	}
}

func TestParseFeatureFile(t *testing.T) {
	dir := t.TempDir()
	featurePath := filepath.Join(dir, "test.feature")
	content := `@feature:test
Feature: Test Feature
  Scenario: First
    Given A
    When B
    Then C

  Scenario: Second
    Given X
    When Y
    Then Z
`
	if err := os.WriteFile(featurePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	ff, err := ParseFeatureFile(featurePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ff.Tag != "test" {
		t.Errorf("expected tag 'test', got: %s", ff.Tag)
	}
	if len(ff.Scenarios) != 2 {
		t.Fatalf("expected 2 scenarios, got %d", len(ff.Scenarios))
	}
	if ff.Scenarios[0].Name != "First" {
		t.Errorf("expected scenario name 'First', got: %s", ff.Scenarios[0].Name)
	}
	if len(ff.Scenarios[0].Steps) != 3 {
		t.Errorf("expected 3 steps, got %d", len(ff.Scenarios[0].Steps))
	}
}

func TestParseFeatureFileNotFound(t *testing.T) {
	_, err := ParseFeatureFile("/nonexistent/path.feature")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if !strings.HasPrefix(err.Error(), "err:validation") {
		t.Errorf("expected err:validation, got: %v", err)
	}
}
