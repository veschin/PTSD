package cli

import "fmt"

// Version is set at build time via -ldflags.
var Version = "dev"

func RunVersion(args []string, agentMode bool) int {
	fmt.Printf("ptsd %s\n", Version)
	return 0
}
