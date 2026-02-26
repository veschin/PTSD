package core

// Stubs for unimplemented functions. Delete when all implementations are complete.

import "time"

// --- bdd.go ---

type FeatureFileData struct {
	Tag       string
	Title     string
	Scenarios []ScenarioData
}

type ScenarioData struct {
	Name  string
	Title string
	Steps []string
}

func InitSeed(projectDir string, featureID string) error {
	panic("not implemented: InitSeed")
}

func AddBDD(projectDir string, featureID string) error {
	panic("not implemented: AddBDD")
}

func CheckBDD(projectDir string) ([]string, error) {
	panic("not implemented: CheckBDD")
}

func ShowBDD(projectDir string, featureID string) ([]string, error) {
	panic("not implemented: ShowBDD")
}

func ParseFeatureFile(path string) (FeatureFileData, error) {
	panic("not implemented: ParseFeatureFile")
}

// --- seed.go ---

func CheckSeeds(projectDir string) ([]string, error) {
	panic("not implemented: CheckSeeds")
}

func AddSeedFile(projectDir string, featureID string, filePath string, fileType string) error {
	panic("not implemented: AddSeedFile")
}

// --- tasks.go ---

type Task struct {
	ID       string
	Feature  string
	Title    string
	Status   string
	Priority string
}

func AddTask(projectDir string, featureID string, title string, priority string) (Task, error) {
	panic("not implemented: AddTask")
}

func ListTasks(projectDir string, featureFilter string, statusFilter string) ([]Task, error) {
	panic("not implemented: ListTasks")
}

func UpdateTask(projectDir string, id string, status string) error {
	panic("not implemented: UpdateTask")
}

func TaskNext(projectDir string, limit int) ([]Task, error) {
	panic("not implemented: TaskNext")
}

// --- pipeline.go ---

type ValidationError struct {
	Feature  string
	Category string
	Message  string
}

func Validate(projectDir string) ([]ValidationError, error) {
	panic("not implemented: Validate")
}

// --- state.go ---

type FeatureState struct {
	Stage  string
	Hashes map[string]string
	Scores map[string]ScoreEntry
	Tests  interface{}
}

type ScoreEntry struct {
	Value     int
	Timestamp time.Time
}

type State struct {
	Features map[string]FeatureState
}

type RegressionWarning struct {
	Feature  string
	File     string
	FileType string
	Category string
	Message  string
}

func LoadState(projectDir string) (*State, error) {
	panic("not implemented: LoadState")
}

func SyncState(projectDir string) error {
	panic("not implemented: SyncState")
}

func CheckRegressions(projectDir string) ([]RegressionWarning, error) {
	panic("not implemented: CheckRegressions")
}

// --- review.go ---

func RecordReview(projectDir string, featureID string, stage string, score int) error {
	panic("not implemented: RecordReview")
}

func CheckReviewGate(projectDir string, featureID string, stage string) (bool, error) {
	panic("not implemented: CheckReviewGate")
}

// --- testrunner.go ---

type TestResults struct {
	Total    int
	Passed   int
	Failed   int
	Failures []string
}

type CoverageEntry struct {
	Feature string
	Status  string
}

func MapTest(projectDir string, bddFile string, testFile string) error {
	panic("not implemented: MapTest")
}

func CheckTestCoverage(projectDir string) ([]CoverageEntry, error) {
	panic("not implemented: CheckTestCoverage")
}

func RunTests(projectDir string, featureFilter string) (TestResults, error) {
	panic("not implemented: RunTests")
}
