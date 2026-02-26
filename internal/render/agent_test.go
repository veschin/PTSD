package render

import (
	"strings"
	"testing"
)

func TestRenderStatus(t *testing.T) {
	tests := []struct {
		name     string
		data     StatusData
		expected string
	}{
		{
			name: "typical project status",
			data: StatusData{
				FeatTotal: 5, FeatFail: 0,
				BDDTotal: 5, BDDFail: 0,
				TestTotal: 5, TestFail: 0,
				TaskTotal: 20, TaskWIP: 0, TaskTodo: 19, TaskDone: 1,
			},
			expected: "[FEAT:5 FAIL:0] [BDD:5 FAIL:0] [TESTS:5 FAIL:0] [T:20 WIP:0 TODO:19 DONE:1]",
		},
		{
			name: "partial progress",
			data: StatusData{
				FeatTotal: 5, FeatFail: 0,
				BDDTotal: 3, BDDFail: 0,
				TestTotal: 2, TestFail: 0,
				TaskTotal: 20, TaskWIP: 0, TaskTodo: 19, TaskDone: 1,
			},
			expected: "[FEAT:5 FAIL:0] [BDD:3 FAIL:0] [TESTS:2 FAIL:0] [T:20 WIP:0 TODO:19 DONE:1]",
		},
		{
			name: "empty project",
			data: StatusData{
				FeatTotal: 0, FeatFail: 0,
				BDDTotal: 0, BDDFail: 0,
				TestTotal: 0, TestFail: 0,
				TaskTotal: 0, TaskWIP: 0, TaskTodo: 0, TaskDone: 0,
			},
			expected: "[FEAT:0 FAIL:0] [BDD:0 FAIL:0] [TESTS:0 FAIL:0] [T:0 WIP:0 TODO:0 DONE:0]",
		},
		{
			name: "with failures",
			data: StatusData{
				FeatTotal: 5, FeatFail: 1,
				BDDTotal: 3, BDDFail: 0,
				TestTotal: 2, TestFail: 1,
				TaskTotal: 10, TaskWIP: 2, TaskTodo: 5, TaskDone: 3,
			},
			expected: "[FEAT:5 FAIL:1] [BDD:3 FAIL:0] [TESTS:2 FAIL:1] [T:10 WIP:2 TODO:5 DONE:3]",
		},
	}

	r := &AgentRenderer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.RenderStatus(tt.data)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestRenderTaskNext(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []TaskView
		expected string
	}{
		{
			name: "single todo task with all ranges",
			tasks: []TaskView{
				{
					ID: "T-1", Status: "TODO", Priority: "A",
					PRDRange: "l30-40", BDDRange: "l30-100", TestRange: "l0-200",
					Title: "Implement user auth",
				},
			},
			expected: "T-1 [TODO] [A] [PRD:l30-40 BDD:l30-100 TEST:l0-200]: Implement user auth",
		},
		{
			name: "wip task",
			tasks: []TaskView{
				{
					ID: "T-2", Status: "WIP", Priority: "B",
					PRDRange: "l50-60", BDDRange: "", TestRange: "",
					Title: "Add validation",
				},
			},
			expected: "T-2 [WIP] [B] [PRD:l50-60]: Add validation",
		},
		{
			name: "no tasks",
			tasks: []TaskView{},
			expected: "",
		},
		{
			name: "multiple tasks renders all",
			tasks: []TaskView{
				{ID: "T-1", Status: "TODO", Priority: "A", PRDRange: "l1-10", Title: "First"},
				{ID: "T-2", Status: "TODO", Priority: "B", PRDRange: "l20-30", Title: "Second"},
			},
			expected: "T-1 [TODO] [A] [PRD:l1-10]: First\nT-2 [TODO] [B] [PRD:l20-30]: Second",
		},
	}

	r := &AgentRenderer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.RenderTaskNext(tt.tasks)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestRenderError(t *testing.T) {
	tests := []struct {
		name     string
		category string
		message  string
		expected string
	}{
		{
			name:     "pipeline error",
			category: "pipeline",
			message:  "user-auth has bdd but no tests",
			expected: "err:pipeline user-auth has bdd but no tests",
		},
		{
			name:     "config error",
			category: "config",
			message:  "missing ptsd.yaml",
			expected: "err:config missing ptsd.yaml",
		},
		{
			name:     "validation error",
			category: "validation",
			message:  "feature catalog missing prd section",
			expected: "err:validation feature catalog missing prd section",
		},
	}

	r := &AgentRenderer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.RenderError(tt.category, tt.message)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestRenderFeatureShow(t *testing.T) {
	tests := []struct {
		name     string
		feature  FeatureView
		contains []string
	}{
		{
			name: "feature with full details",
			feature: FeatureView{
				ID: "user-auth", Status: "in-progress",
				PRDRange: "l30-40", SeedStatus: "ok",
				BDDCount: 3, TestCovered: 2, TestTotal: 3,
				Scores: map[string]int{"prd": 8, "seed": 9, "bdd": 7},
			},
			contains: []string{"user-auth", "in-progress", "PRD:l30-40", "SEED:ok", "BDD:3scn", "TEST:2/3", "prd=8", "seed=9", "bdd=7"},
		},
		{
			name: "feature missing seed",
			feature: FeatureView{
				ID: "catalog", Status: "seed",
				PRDRange: "l100-120", SeedStatus: "missing",
				BDDCount: 0, TestCovered: 0, TestTotal: 0,
				Scores: map[string]int{"prd": 6},
			},
			contains: []string{"catalog", "seed", "PRD:l100-120", "SEED:missing", "BDD:0scn", "TEST:0/0"},
		},
	}

	r := &AgentRenderer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.RenderFeatureShow(tt.feature)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("expected output to contain %q, got %q", want, got)
				}
			}
		})
	}
}

func TestRenderTestResults(t *testing.T) {
	tests := []struct {
		name     string
		results  TestResultsView
		contains []string
	}{
		{
			name: "all passing",
			results: TestResultsView{
				Total: 5, Passed: 5, Failed: 0,
				Duration: "1.2s",
			},
			contains: []string{"pass:5", "fail:0"},
		},
		{
			name: "some failures",
			results: TestResultsView{
				Total: 5, Passed: 3, Failed: 2,
				Duration: "2.5s",
				Failures: []string{"TestAuth", "TestLogin"},
			},
			contains: []string{"pass:3", "fail:2", "fail:TestAuth,TestLogin"},
		},
	}

	r := &AgentRenderer{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.RenderTestResults(tt.results)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("expected output to contain %q, got %q", want, got)
				}
			}
		})
	}
}

func TestAgentRendererImplementsRenderer(t *testing.T) {
	var _ Renderer = &AgentRenderer{}
}
