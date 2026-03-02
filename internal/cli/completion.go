package cli

import (
	"fmt"
	"os"

	"github.com/veschin/ptsd/internal/core"
)

func RunCompletion(args []string, agentMode bool) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "err:user usage: ptsd completion <bash|fish>")
		return 2
	}

	var path string
	switch args[0] {
	case "bash":
		path = "templates/completion/completion.bash"
	case "fish":
		path = "templates/completion/completion.fish"
	default:
		fmt.Fprintf(os.Stderr, "err:user unsupported shell: %s (use bash or fish)\n", args[0])
		return 2
	}

	content, err := core.ReadEmbeddedTemplate(path)
	if err != nil {
		return coreError(agentMode, err)
	}

	fmt.Print(content)
	return 0
}
