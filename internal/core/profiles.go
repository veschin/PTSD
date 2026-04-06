package core

// ValidPipelines lists accepted profile names.
var ValidPipelines = map[string]bool{
	"full":     true,
	"standard": true,
	"lite":     true,
}

// PipelineStages defines which stages each profile requires, in order.
var PipelineStages = map[string][]string{
	"full":     {"prd", "seed", "bdd", "tests", "impl"},
	"standard": {"prd", "bdd", "tests", "impl"},
	"lite":     {"prd", "tests", "impl"},
}

// StageRequired checks if a stage is part of a pipeline profile.
func StageRequired(pipeline, stage string) bool {
	stages := resolvePipeline(pipeline)
	for _, s := range stages {
		if s == stage {
			return true
		}
	}
	return false
}

// NextStage returns the next stage in the pipeline for a given profile.
// Returns "" if currentStage is the last stage or not found.
func NextStage(pipeline, currentStage string) string {
	stages := resolvePipeline(pipeline)
	for i, s := range stages {
		if s == currentStage && i+1 < len(stages) {
			return stages[i+1]
		}
	}
	return ""
}

// NextAction returns the skill/action name for the next stage transition.
// E.g., for standard pipeline at "prd" stage -> "write-bdd".
func NextAction(pipeline, currentStage string) string {
	next := NextStage(pipeline, currentStage)
	if next == "" {
		return ""
	}
	return "write-" + next
}

// resolvePipeline returns stages for a pipeline, defaulting to "standard".
func resolvePipeline(pipeline string) []string {
	if stages, ok := PipelineStages[pipeline]; ok {
		return stages
	}
	return PipelineStages["standard"]
}

// ResolveFeaturePipeline returns the effective pipeline for a feature.
// Uses feature's own pipeline field if set, otherwise falls back to config default,
// and ultimately to "standard".
func ResolveFeaturePipeline(projectDir, featureID string) string {
	features, err := loadFeatures(projectDir)
	if err != nil {
		return resolveDefaultPipeline(projectDir)
	}
	for _, f := range features {
		if f.ID == featureID {
			if f.Pipeline != "" {
				return f.Pipeline
			}
			return resolveDefaultPipeline(projectDir)
		}
	}
	return resolveDefaultPipeline(projectDir)
}

// resolveDefaultPipeline reads the default pipeline from config, falling back to "standard".
func resolveDefaultPipeline(projectDir string) string {
	cfg, err := LoadConfig(projectDir)
	if err != nil {
		return "standard"
	}
	if cfg.Pipeline.Default != "" {
		return cfg.Pipeline.Default
	}
	return "standard"
}
