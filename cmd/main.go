package main

import (
	"fmt"
	"os"

	"yappers-of-linux/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		cli.ShowHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "help", "--help", "-h":
		cli.ShowHelp()
	case "models":
		cli.ShowModels()
	case "start":
		cli.Start(os.Args[2:])
	case "toggle":
		cli.Toggle(os.Args[2:])
	case "pause":
		cli.Pause()
	case "resume":
		cli.Resume()
	case "stop", "kill":
		cli.Stop()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		fmt.Fprintln(os.Stderr, "run 'yap help' for usage")
		os.Exit(1)
	}
}
