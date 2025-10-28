package commands

import (
	"fmt"
	"os"
	"syscall"

	"yappers-of-linux/internal"
)

func Pause() {
	pid, err := internal.GetPID()
	if err != nil {
		fmt.Println("not running")
		os.Exit(1)
	}

	if err := syscall.Kill(pid, syscall.SIGUSR1); err != nil {
		fmt.Fprintf(os.Stderr, "failed to pause: %v\n", err)
		os.Exit(1)
	}

	os.WriteFile(internal.StateFile, []byte("paused"), 0644)
}
