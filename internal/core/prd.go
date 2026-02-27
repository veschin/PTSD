package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PRDError struct {
	Type      string
	FeatureID string
}

type PRDSection struct {
	FeatureID string
	StartLine int
	EndLine   int
	Content   string
}

const anchorPrefix = "<!-- feature:"
const anchorSuffix = " -->"

func CheckPRDAnchors(projectDir string) ([]PRDError, error) {
	anchors, err := extractAnchors(projectDir)
	if err != nil {
		return nil, err
	}

	features, err := readAllFeatureIDs(projectDir)
	if err != nil {
		return nil, err
	}

	anchorSet := make(map[string]bool)
	for _, a := range anchors {
		anchorSet[a] = true
	}

	featureSet := make(map[string]bool)
	for _, f := range features {
		featureSet[f] = true
	}

	var errs []PRDError

	for _, f := range features {
		if !anchorSet[f] {
			errs = append(errs, PRDError{Type: "missing-anchor", FeatureID: f})
		}
	}

	for _, a := range anchors {
		if !featureSet[a] {
			errs = append(errs, PRDError{Type: "orphaned-anchor", FeatureID: a})
		}
	}

	return errs, nil
}

func ExtractPRDSection(projectDir string, featureID string) (PRDSection, error) {
	prdPath := filepath.Join(projectDir, ".ptsd", "docs", "PRD.md")
	f, err := os.Open(prdPath)
	if err != nil {
		return PRDSection{}, fmt.Errorf("err:io %w", err)
	}
	defer f.Close()

	target := anchorPrefix + featureID + anchorSuffix
	scanner := bufio.NewScanner(f)
	lineNum := 0
	found := false
	startLine := 0
	var contentLines []string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if !found {
			if strings.TrimSpace(line) == target {
				found = true
				startLine = lineNum
			}
			continue
		}

		if strings.Contains(line, anchorPrefix) {
			return PRDSection{
				FeatureID: featureID,
				StartLine: startLine,
				EndLine:   lineNum - 1,
				Content:   strings.Join(contentLines, "\n"),
			}, nil
		}

		contentLines = append(contentLines, line)
	}

	if !found {
		return PRDSection{}, fmt.Errorf("err:pipeline anchor not found for %s", featureID)
	}

	return PRDSection{
		FeatureID: featureID,
		StartLine: startLine,
		EndLine:   lineNum,
		Content:   strings.Join(contentLines, "\n"),
	}, nil
}

func extractAnchors(projectDir string) ([]string, error) {
	prdPath := filepath.Join(projectDir, ".ptsd", "docs", "PRD.md")
	data, err := os.ReadFile(prdPath)
	if err != nil {
		return nil, fmt.Errorf("err:io %w", err)
	}

	var anchors []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, anchorPrefix) && strings.HasSuffix(line, anchorSuffix) {
			id := line[len(anchorPrefix) : len(line)-len(anchorSuffix)]
			anchors = append(anchors, id)
		}
	}
	return anchors, nil
}

func readAllFeatureIDs(projectDir string) ([]string, error) {
	featPath := filepath.Join(projectDir, ".ptsd", "features.yaml")
	data, err := os.ReadFile(featPath)
	if err != nil {
		return nil, fmt.Errorf("err:io %w", err)
	}

	var ids []string
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- id: ") {
			ids = append(ids, strings.TrimPrefix(trimmed, "- id: "))
		}
	}
	return ids, nil
}

func readActiveFeatureIDs(projectDir string) ([]string, error) {
	featPath := filepath.Join(projectDir, ".ptsd", "features.yaml")
	data, err := os.ReadFile(featPath)
	if err != nil {
		return nil, fmt.Errorf("err:io %w", err)
	}

	var ids []string
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- id: ") {
			id := strings.TrimPrefix(trimmed, "- id: ")
			status := ""
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if strings.HasPrefix(next, "- id: ") || next == "" {
					break
				}
				if strings.HasPrefix(next, "status: ") {
					status = strings.TrimPrefix(next, "status: ")
					break
				}
			}
			if status != "planned" && status != "deferred" {
				ids = append(ids, id)
			}
		}
	}
	return ids, nil
}
