package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckPRDAnchorsAllPresent(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress", "catalog:in-progress")
	docsDir := filepath.Join(dir, ".ptsd", "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}

	prd := "# PRD\n<!-- feature:user-auth -->\nAuth section\n<!-- feature:catalog -->\nCatalog section\n"
	if err := os.WriteFile(filepath.Join(docsDir, "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	errors, err := CheckPRDAnchors(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(errors), errors)
	}
}

func TestCheckPRDAnchorsFeatureMissingAnchor(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")
	docsDir := filepath.Join(dir, ".ptsd", "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}

	prd := "# PRD\nNo anchors here\n"
	if err := os.WriteFile(filepath.Join(docsDir, "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	errors, err := CheckPRDAnchors(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errors), errors)
	}
	if errors[0].Type != "missing-anchor" {
		t.Errorf("expected type missing-anchor, got %s", errors[0].Type)
	}
	if errors[0].FeatureID != "user-auth" {
		t.Errorf("expected feature user-auth, got %s", errors[0].FeatureID)
	}
}

func TestCheckPRDAnchorsOrphanedAnchor(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")
	docsDir := filepath.Join(dir, ".ptsd", "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}

	prd := "# PRD\n<!-- feature:user-auth -->\nAuth\n<!-- feature:ghost -->\nGhost section\n"
	if err := os.WriteFile(filepath.Join(docsDir, "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	errors, err := CheckPRDAnchors(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var orphaned []PRDError
	for _, e := range errors {
		if e.Type == "orphaned-anchor" {
			orphaned = append(orphaned, e)
		}
	}
	if len(orphaned) != 1 {
		t.Fatalf("expected 1 orphaned-anchor error, got %d: %v", len(orphaned), errors)
	}
	if orphaned[0].FeatureID != "ghost" {
		t.Errorf("expected feature ghost, got %s", orphaned[0].FeatureID)
	}
}

func TestExtractPRDSection(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress", "catalog:in-progress")
	docsDir := filepath.Join(dir, ".ptsd", "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}

	prd := "# PRD\n<!-- feature:user-auth -->\nLine one\nLine two\nLine three\n<!-- feature:catalog -->\nCatalog stuff\n"
	if err := os.WriteFile(filepath.Join(docsDir, "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	section, err := ExtractPRDSection(dir, "user-auth")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if section.FeatureID != "user-auth" {
		t.Errorf("expected feature user-auth, got %s", section.FeatureID)
	}
	if section.StartLine < 1 {
		t.Errorf("expected StartLine >= 1, got %d", section.StartLine)
	}
	if section.EndLine <= section.StartLine {
		t.Errorf("expected EndLine > StartLine, got start=%d end=%d", section.StartLine, section.EndLine)
	}
	if !strings.Contains(section.Content, "Line one") || !strings.Contains(section.Content, "Line three") {
		t.Errorf("expected content to contain section lines, got: %q", section.Content)
	}
	if strings.Contains(section.Content, "Catalog stuff") {
		t.Error("content should not contain next feature's section")
	}
}

func TestExtractPRDSectionNotFound(t *testing.T) {
	dir := setupProjectWithFeatures(t, "user-auth:in-progress")
	docsDir := filepath.Join(dir, ".ptsd", "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatal(err)
	}

	prd := "# PRD\nNo anchors\n"
	if err := os.WriteFile(filepath.Join(docsDir, "PRD.md"), []byte(prd), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := ExtractPRDSection(dir, "user-auth")
	if err == nil {
		t.Fatal("expected error for missing anchor")
	}
}
