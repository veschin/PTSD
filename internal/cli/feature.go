package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/veschin/ptsd/internal/core"
	"github.com/veschin/ptsd/internal/render"
)

func RunFeature(args []string, agentMode bool) int {
	if len(args) == 0 {
		return usageError(agentMode, "feature", "subcommand required: add|list|remove|status|show|pipeline")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return renderError(agentMode, "io", err.Error())
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "add":
		if len(rest) < 2 {
			return usageError(agentMode, "feature add", "usage: feature add <id> <title> [--pipeline full|standard|lite]")
		}
		id := rest[0]
		pipeline := ""
		var titleParts []string
		for i := 1; i < len(rest); i++ {
			if rest[i] == "--pipeline" && i+1 < len(rest) {
				pipeline = rest[i+1]
				i++
			} else {
				titleParts = append(titleParts, rest[i])
			}
		}
		title := strings.Join(titleParts, " ")
		if title == "" {
			return usageError(agentMode, "feature add", "usage: feature add <id> <title> [--pipeline full|standard|lite]")
		}
		if err := core.AddFeatureWithPipeline(cwd, id, title, pipeline); err != nil {
			return coreError(agentMode, err)
		}
		if agentMode {
			if pipeline != "" {
				fmt.Printf("feature.add id=%s pipeline=%s\n", id, pipeline)
			} else {
				fmt.Printf("feature.add id=%s\n", id)
			}
		} else {
			fmt.Printf("Added feature: %s\n", id)
		}
		return 0

	case "list":
		filter := ""
		if len(rest) > 0 {
			filter = rest[0]
		}
		features, err := core.ListFeatures(cwd, filter)
		if err != nil {
			return coreError(agentMode, err)
		}
		if agentMode {
			for _, f := range features {
				p := f.Pipeline
				if p == "" {
					p = "standard"
				}
				fmt.Printf("%s [%s:%s] %s\n", f.ID, f.Status, p, f.Title)
			}
		} else {
			for _, f := range features {
				fmt.Printf("%-30s %-15s %s\n", f.ID, f.Status, f.Title)
			}
		}
		return 0

	case "remove":
		if len(rest) < 1 {
			return usageError(agentMode, "feature remove", "usage: feature remove <id>")
		}
		id := rest[0]
		if err := core.RemoveFeature(cwd, id); err != nil {
			return coreError(agentMode, err)
		}
		if agentMode {
			fmt.Printf("feature.remove id=%s\n", id)
		} else {
			fmt.Printf("Removed feature: %s\n", id)
		}
		return 0

	case "status":
		if len(rest) < 2 {
			return usageError(agentMode, "feature status", "usage: feature status <id> <status>")
		}
		id := rest[0]
		status := rest[1]
		if err := core.UpdateFeatureStatus(cwd, id, status); err != nil {
			return coreError(agentMode, err)
		}
		if agentMode {
			fmt.Printf("feature.status id=%s status=%s\n", id, status)
		} else {
			fmt.Printf("Updated feature %s status to %s\n", id, status)
		}
		return 0

	case "show":
		if len(rest) < 1 {
			return usageError(agentMode, "feature show", "usage: feature show <id>")
		}
		id := rest[0]
		detail, err := core.ShowFeature(cwd, id)
		if err != nil {
			return coreError(agentMode, err)
		}
		fv := render.FeatureView{
			ID:          detail.ID,
			Status:      detail.Status,
			Pipeline:    detail.Pipeline,
			PRDRange:    detail.PRDAnchor,
			SeedStatus:  detail.SeedStatus,
			BDDCount:    detail.ScenarioCount,
			TestTotal:   detail.TestCount,
			TestCovered: detail.TestPassed,
		}
		r := newRenderer(agentMode)
		fmt.Println(r.RenderFeatureShow(fv))
		return 0

	case "pipeline":
		if len(rest) < 2 {
			return usageError(agentMode, "feature pipeline", "usage: feature pipeline <id> <full|standard|lite>")
		}
		id := rest[0]
		pipeline := rest[1]
		if err := core.UpdateFeaturePipeline(cwd, id, pipeline); err != nil {
			return coreError(agentMode, err)
		}
		if agentMode {
			fmt.Printf("feature.pipeline id=%s pipeline=%s\n", id, pipeline)
		} else {
			fmt.Printf("Updated feature %s pipeline to %s\n", id, pipeline)
		}
		return 0

	default:
		return usageError(agentMode, "feature", fmt.Sprintf("unknown subcommand %q: use add|list|remove|status|show|pipeline", sub))
	}
}
