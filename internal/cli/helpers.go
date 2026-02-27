package cli

import (
	"fmt"
	"strings"

	"github.com/veschin/ptsd/internal/render"
)

func newRenderer(agentMode bool) render.Renderer {
	return &render.AgentRenderer{}
}

func renderError(agentMode bool, category string, message string) int {
	r := newRenderer(agentMode)
	fmt.Println(r.RenderError(category, message))
	return errCategoryCode(category)
}

func coreError(agentMode bool, err error) int {
	msg := err.Error()
	category := "io"
	if strings.HasPrefix(msg, "err:") {
		parts := strings.SplitN(msg[4:], " ", 2)
		if len(parts) == 2 {
			category = parts[0]
			msg = parts[1]
		} else if len(parts) == 1 {
			category = strings.TrimRight(parts[0], ":")
			msg = ""
		}
	}
	return renderError(agentMode, category, msg)
}

func usageError(agentMode bool, cmd string, message string) int {
	return renderError(agentMode, "user", fmt.Sprintf("%s: %s", cmd, message))
}

func errCategoryCode(category string) int {
	switch category {
	case "validation", "pipeline":
		return 1
	case "user":
		return 2
	case "config":
		return 3
	case "io":
		return 4
	case "test":
		return 5
	default:
		return 1
	}
}
