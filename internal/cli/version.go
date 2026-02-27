package cli

import (
	"fmt"
	"runtime/debug"
)

func RunVersion(args []string, agentMode bool) int {
	version := "dev"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
	fmt.Printf("ptsd %s\n", version)
	return 0
}
