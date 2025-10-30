package commands

import (
	"fmt"
	"os"
	"syscall"

	"yappers-of-linux/internal"
)

func Resume() {
	pid, err := internal.GetPID()
	if err != nil {
		fmt.Println("not running")
		os.Exit(1)
	}

	if err := syscall.Kill(pid, syscall.SIGUSR2); err != nil {
		fmt.Fprintf(os.Stderr, "failed to resume: %v\n", err)
		os.Exit(1)
	}

	os.WriteFile(internal.StateFile, []byte("active"), 0644)

	cfg := internal.LoadConfig()
	internal.Notify("Yapping started", "start", cfg)
}
