package render

// Stubs for render package. Delete when agent.go is implemented.

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
	panic("not implemented")
}

func (r *AgentRenderer) RenderTaskNext(tasks []TaskView) string {
	panic("not implemented")
}

func (r *AgentRenderer) RenderError(category string, message string) string {
	panic("not implemented")
}

func (r *AgentRenderer) RenderFeatureShow(feature FeatureView) string {
	panic("not implemented")
}

func (r *AgentRenderer) RenderTestResults(results TestResultsView) string {
	panic("not implemented")
}
