package commands

import (
	"fmt"
	"os"
)

func Parse(args []string) {
	if len(args) < 2 {
		Help()
		return
	}

	command := args[1]

	switch command {
	case "help", "--help", "-h":
		Help()
	case "models":
		Models()
	case "config":
		Config()
	case "start":
		Start(args[2:])
	case "toggle":
		Toggle(args[2:])
	case "pause":
		Pause()
	case "resume":
		Resume()
	case "stop", "kill":
		Stop()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		fmt.Fprintln(os.Stderr, "run 'yap help' for usage")
		os.Exit(1)
	}
}
