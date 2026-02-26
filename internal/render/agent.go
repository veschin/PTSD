package render

import (
	"fmt"
	"sort"
	"strings"
)

type StatusData struct {
	FeatTotal int
	FeatFail  int
	BDDTotal  int
	BDDFail   int
	TestTotal int
	TestFail  int
	TaskTotal int
	TaskWIP   int
	TaskTodo  int
	TaskDone  int
}

type TaskView struct {
	ID        string
	Status    string
	Priority  string
	PRDRange  string
	BDDRange  string
	TestRange string
	Title     string
}

type FeatureView struct {
	ID          string
	Status      string
	PRDRange    string
	SeedStatus  string
	BDDCount    int
	TestCovered int
	TestTotal   int
	Scores      map[string]int
}

type TestResultsView struct {
	Total    int
	Passed   int
	Failed   int
	Duration string
	Failures []string
}

type Renderer interface {
	RenderStatus(data StatusData) string
	RenderTaskNext(tasks []TaskView) string
	RenderError(category string, message string) string
	RenderFeatureShow(feature FeatureView) string
	RenderTestResults(results TestResultsView) string
}

type AgentRenderer struct{}

func (r *AgentRenderer) RenderStatus(data StatusData) string {
	return fmt.Sprintf("[FEAT:%d FAIL:%d] [BDD:%d FAIL:%d] [TESTS:%d FAIL:%d] [T:%d WIP:%d TODO:%d DONE:%d]",
		data.FeatTotal, data.FeatFail,
		data.BDDTotal, data.BDDFail,
		data.TestTotal, data.TestFail,
		data.TaskTotal, data.TaskWIP, data.TaskTodo, data.TaskDone)
}

func (r *AgentRenderer) RenderTaskNext(tasks []TaskView) string {
	if len(tasks) == 0 {
		return ""
	}

	var lines []string
	for _, t := range tasks {
		var ranges []string
		if t.PRDRange != "" {
			ranges = append(ranges, "PRD:"+t.PRDRange)
		}
		if t.BDDRange != "" {
			ranges = append(ranges, "BDD:"+t.BDDRange)
		}
		if t.TestRange != "" {
			ranges = append(ranges, "TEST:"+t.TestRange)
		}
		rangeStr := strings.Join(ranges, " ")
		line := fmt.Sprintf("%s [%s] [%s] [%s]: %s", t.ID, t.Status, t.Priority, rangeStr, t.Title)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (r *AgentRenderer) RenderError(category string, message string) string {
	return fmt.Sprintf("err:%s %s", category, message)
}

func (r *AgentRenderer) RenderFeatureShow(feature FeatureView) string {
	var scoresParts []string
	keys := make([]string, 0, len(feature.Scores))
	for k := range feature.Scores {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		scoresParts = append(scoresParts, fmt.Sprintf("%s=%d", k, feature.Scores[k]))
	}
	scoresStr := strings.Join(scoresParts, ",")

	result := fmt.Sprintf("%s [%s] PRD:%s SEED:%s BDD:%dscn TEST:%d/%d",
		feature.ID, feature.Status,
		feature.PRDRange, feature.SeedStatus,
		feature.BDDCount, feature.TestCovered, feature.TestTotal)

	if scoresStr != "" {
		result += " SCORE:" + scoresStr
	}

	return result
}

func (r *AgentRenderer) RenderTestResults(results TestResultsView) string {
	out := fmt.Sprintf("pass:%d fail:%d", results.Passed, results.Failed)
	if len(results.Failures) > 0 {
		out += " fail:" + strings.Join(results.Failures, ",")
	}
	return out
}
