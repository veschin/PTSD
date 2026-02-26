package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Project ProjectConfig
	Testing TestingConfig
	Review  ReviewConfig
	Hooks   HooksConfig
}

type ProjectConfig struct {
	Name string
}

type TestingConfig struct {
	Runner       string
	Patterns     PatternsConfig
	ResultParser ResultParserConfig
}

type PatternsConfig struct {
	Files []string
}

type ResultParserConfig struct {
	Format      string
	Root        string
	StatusField string
	PassedValue string
	FailedValue string
}

type ReviewConfig struct {
	MinScore int
	AutoRedo bool
}

type HooksConfig struct {
	PreCommit bool
	Scopes    []string
	Types     []string
}

func LoadConfig(dir string) (*Config, error) {
	cfgPath, err := findConfigPath(dir)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("err:config: %w", err)
	}

	cfg, err := parseConfig(string(content))
	if err != nil {
		return nil, err
	}

	applyDefaults(cfg)

	return cfg, nil
}

func findConfigPath(dir string) (string, error) {
	for {
		cfgPath := filepath.Join(dir, ".ptsd", "ptsd.yaml")
		if _, err := os.Stat(cfgPath); err == nil {
			return cfgPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("err:config: not found")
}

func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func parseInlineArray(value string) []string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		inner := value[1 : len(value)-1]
		var result []string
		for _, item := range strings.Split(inner, ",") {
			item = strings.TrimSpace(item)
			item = stripQuotes(item)
			if item != "" {
				result = append(result, item)
			}
		}
		return result
	}
	return nil
}

func parseConfig(content string) (*Config, error) {
	cfg := &Config{
		Hooks: HooksConfig{PreCommit: true},
	}

	lines := strings.Split(content, "\n")
	var currentSection string
	var currentSubSection string
	preCommitExplicit := false

	for i, line := range lines {
		line = strings.TrimRight(line, " ")

		if strings.Contains(line, "[") && !strings.Contains(line, "]") {
			return nil, fmt.Errorf("err:config: invalid YAML at line %d: unclosed bracket", i+1)
		}

		if strings.HasSuffix(line, ":") && !strings.HasPrefix(line, " ") {
			currentSection = strings.TrimSuffix(line, ":")
			currentSubSection = ""
			continue
		}

		if strings.HasPrefix(line, "  ") && strings.HasSuffix(line, ":") && !strings.HasPrefix(line, "    ") {
			currentSubSection = strings.TrimSpace(strings.TrimSuffix(line, ":"))
			continue
		}

		if strings.Contains(line, ": ") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = stripQuotes(value)

			if currentSection == "project" {
				if key == "name" {
					cfg.Project.Name = value
				}
			} else if currentSection == "testing" {
				if currentSubSection == "patterns" && key == "files" {
					if inline := parseInlineArray(parts[1]); inline != nil {
						cfg.Testing.Patterns.Files = inline
					} else {
						cfg.Testing.Patterns.Files = parseArray(lines, i)
					}
				} else if currentSubSection == "result_parser" {
					switch key {
					case "format":
						cfg.Testing.ResultParser.Format = value
					case "root":
						cfg.Testing.ResultParser.Root = value
					case "status_field":
						cfg.Testing.ResultParser.StatusField = value
					case "passed_value":
						cfg.Testing.ResultParser.PassedValue = value
					case "failed_value":
						cfg.Testing.ResultParser.FailedValue = value
					}
				} else if key == "runner" {
					cfg.Testing.Runner = value
				}
			} else if currentSection == "review" {
				switch key {
				case "min_score":
					n, err := strconv.Atoi(value)
					if err != nil {
						return nil, fmt.Errorf("err:config invalid min_score: %s", value)
					}
					cfg.Review.MinScore = n
				case "auto_redo":
					cfg.Review.AutoRedo = value == "true"
				}
			} else if currentSection == "hooks" {
				switch key {
				case "pre_commit":
					cfg.Hooks.PreCommit = value == "true"
					preCommitExplicit = true
				case "scopes":
					if inline := parseInlineArray(parts[1]); inline != nil {
						cfg.Hooks.Scopes = inline
					} else {
						cfg.Hooks.Scopes = parseArray(lines, i)
					}
				case "types":
					if inline := parseInlineArray(parts[1]); inline != nil {
						cfg.Hooks.Types = inline
					} else {
						cfg.Hooks.Types = parseArray(lines, i)
					}
				}
			}
		}
	}

	_ = preCommitExplicit
	return cfg, nil
}

func parseArray(lines []string, startIdx int) []string {
	var result []string
	indent := "    - "

	for i := startIdx + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, indent) {
			val := strings.TrimPrefix(line, indent)
			val = strings.TrimSpace(val)
			val = stripQuotes(val)
			result = append(result, val)
		} else if line == "" {
			continue
		} else {
			break
		}
	}

	return result
}

func applyDefaults(cfg *Config) {
	if len(cfg.Testing.Patterns.Files) == 0 {
		cfg.Testing.Patterns.Files = []string{"**/*_test.go"}
	}
	if cfg.Review.MinScore == 0 {
		cfg.Review.MinScore = 7
	}
}
