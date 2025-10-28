package commands

import (
	"fmt"
	"os"
	"syscall"

	"yappers-of-linux/internal"
)

func Stop() {
	pid, err := internal.GetPID()
	if err != nil {
		fmt.Println("not running")
		os.Exit(1)
	}

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "failed to stop: %v\n", err)
		os.Exit(1)
	}

	os.Remove(internal.PIDFile)
	os.Remove(internal.StateFile)
}
