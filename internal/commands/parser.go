package commands

import (
	"fmt"
	"os"
)

func Parse(args []string) {
	if len(args) < 2 {
		Help([]string{})
		return
	}

	command := args[1]

	switch command {
	case "help", "--help", "-h":
		Help(args[2:])
	case "version", "--version", "-v":
		ShowVersion()
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
	case "update":
		Update(args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		fmt.Fprintln(os.Stderr, "run 'yap help' for usage")
		os.Exit(1)
	}
}
