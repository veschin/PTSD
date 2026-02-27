package core

import (
	"os"
	"path/filepath"
	"strings"
)

type GateCheckResult struct {
	Allowed bool
	Reason  string
	Feature string
}

// alwaysAllowed lists file paths that never require gate checks.
// NOTE: review-status.yaml is NOT here — direct AI edits are blocked.
// AutoTrack (PostToolUse) updates it via Go code, bypassing the gate.
// Use `ptsd review` to set review verdicts.
var alwaysAllowed = map[string]bool{
	".ptsd/docs/PRD.md":     true,
	".ptsd/tasks.yaml":      true,
	".ptsd/state.yaml":      true,
	".ptsd/features.yaml":   true,
	".ptsd/ptsd.yaml":       true,
	".ptsd/issues.yaml":     true,
	"CLAUDE.md":             true,
	".claude/settings.json": true,
}

func GateCheck(projectDir, filePath string) GateCheckResult {
	// Normalize path: make relative to project
	rel := filePath
	if filepath.IsAbs(filePath) {
		r, err := filepath.Rel(projectDir, filePath)
		if err == nil {
			rel = r
		}
	}

	// Always-allowed files
	if alwaysAllowed[rel] {
		return GateCheckResult{Allowed: true}
	}

	// review-status.yaml: blocked for direct AI edits.
	// AutoTrack (PostToolUse) writes via Go code, bypassing the gate.
	// Use `ptsd review` to set review verdicts.
	if rel == ".ptsd/review-status.yaml" {
		return GateCheckResult{
			Allowed: false,
			Reason:  "direct edits to review-status.yaml are blocked — use ptsd review",
		}
	}

	// Skills are always allowed
	if strings.HasPrefix(rel, ".ptsd/skills/") {
		return GateCheckResult{Allowed: true}
	}

	// Claude hooks are always allowed
	if strings.HasPrefix(rel, ".claude/hooks/") {
		return GateCheckResult{Allowed: true}
	}

	// BDD file → requires seed
	if strings.HasPrefix(rel, ".ptsd/bdd/") && strings.HasSuffix(rel, ".feature") {
		featureID := strings.TrimSuffix(filepath.Base(rel), ".feature")
		seedPath := filepath.Join(projectDir, ".ptsd", "seeds", featureID, "seed.yaml")
		if _, err := os.Stat(seedPath); os.IsNotExist(err) {
			return GateCheckResult{
				Allowed: false,
				Reason:  "no seed for " + featureID + " — run: ptsd seed init " + featureID,
				Feature: featureID,
			}
		}
		return GateCheckResult{Allowed: true, Feature: featureID}
	}

	// Seed file → requires PRD anchor
	if strings.HasPrefix(rel, ".ptsd/seeds/") {
		parts := strings.Split(rel, "/")
		if len(parts) >= 3 {
			featureID := parts[2] // .ptsd/seeds/<id>/...
			anchors, err := extractAnchors(projectDir)
			if err == nil {
				found := false
				for _, a := range anchors {
					if a == featureID {
						found = true
						break
					}
				}
				if !found {
					return GateCheckResult{
						Allowed: false,
						Reason:  "no PRD anchor for " + featureID,
						Feature: featureID,
					}
				}
			}
			return GateCheckResult{Allowed: true, Feature: featureID}
		}
	}

	// Test file → requires BDD
	if strings.HasSuffix(rel, "_test.go") || strings.HasSuffix(rel, ".test.ts") || strings.HasSuffix(rel, ".test.js") {
		featureID := inferFeatureFromTestFile(projectDir, rel)
		if featureID != "" {
			bddPath := filepath.Join(projectDir, ".ptsd", "bdd", featureID+".feature")
			if _, err := os.Stat(bddPath); os.IsNotExist(err) {
				return GateCheckResult{
					Allowed: false,
					Reason:  "no BDD scenarios for " + featureID + " — run: ptsd bdd add " + featureID,
					Feature: featureID,
				}
			}
		}
		return GateCheckResult{Allowed: true, Feature: featureID}
	}

	// Impl code → requires tests exist
	if isImplFile(rel) {
		featureID := inferFeatureFromImplFile(projectDir, rel)
		if featureID != "" {
			state, _ := LoadState(projectDir)
			if !hasTestsForFeature(projectDir, featureID, state) {
				return GateCheckResult{
					Allowed: false,
					Reason:  "no tests for " + featureID,
					Feature: featureID,
				}
			}
		}
		return GateCheckResult{Allowed: true, Feature: featureID}
	}

	return GateCheckResult{Allowed: true}
}

func inferFeatureFromTestFile(projectDir, rel string) string {
	base := filepath.Base(rel)
	// Strip test suffixes
	name := strings.TrimSuffix(base, "_test.go")
	name = strings.TrimSuffix(name, ".test.ts")
	name = strings.TrimSuffix(name, ".test.js")

	// Check if this name matches a feature ID
	features, err := loadFeatures(projectDir)
	if err != nil {
		return ""
	}

	match := matchFeatureID(name, features)
	if match != "" {
		return match
	}

	// Check state test mappings
	state, _ := LoadState(projectDir)
	if state != nil {
		for fid, fs := range state.Features {
			if tests, ok := fs.Tests.([]string); ok {
				for _, t := range tests {
					if strings.Contains(t, rel) {
						return fid
					}
				}
			}
		}
	}

	return ""
}

func inferFeatureFromImplFile(projectDir, rel string) string {
	base := filepath.Base(rel)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	features, err := loadFeatures(projectDir)
	if err != nil {
		return ""
	}

	return matchFeatureID(name, features)
}

// matchFeatureID finds the best matching feature ID for a filename.
// Prefers exact match, then longest substring match to avoid "auth" matching "authorization".
func matchFeatureID(name string, features []Feature) string {
	// Exact match first
	for _, f := range features {
		if name == f.ID {
			return f.ID
		}
	}

	// Longest substring match — prevents "auth" from matching "authorization"
	bestMatch := ""
	bestLen := 0
	for _, f := range features {
		if strings.Contains(name, f.ID) && len(f.ID) > bestLen {
			bestMatch = f.ID
			bestLen = len(f.ID)
		}
	}
	return bestMatch
}

func isImplFile(rel string) bool {
	if strings.HasPrefix(rel, ".ptsd/") || strings.HasPrefix(rel, ".claude/") || strings.HasPrefix(rel, ".git/") {
		return false
	}
	ext := filepath.Ext(rel)
	implExts := map[string]bool{
		".go": true, ".ts": true, ".js": true, ".py": true,
		".rs": true, ".java": true, ".c": true, ".cpp": true,
	}
	return implExts[ext]
}
