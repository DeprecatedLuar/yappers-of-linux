package commands

import (
	"fmt"
	"os"
	"syscall"

	"yappers-of-linux/internal"
)

func Toggle(args []string) {
	pid, err := internal.GetPID()
	if err != nil {
		Start(args)
		return
	}

	if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
		fmt.Fprintf(os.Stderr, "failed to toggle: %v\n", err)
		os.Exit(1)
	}
}
