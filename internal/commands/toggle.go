package commands

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"yappers-of-linux/internal"
)

func Toggle(args []string) {
	cfg := internal.LoadConfig()

	pid, err := internal.GetPID()
	if err != nil {
		os.Remove(internal.StateFile)
		internal.Notify("Yapping started", cfg)
		Start(args)
		return
	}

	stateData, err := os.ReadFile(internal.StateFile)
	state := strings.TrimSpace(string(stateData))

	if err == nil && state == "paused" {
		internal.Notify("Yapping started", cfg)
		if err := syscall.Kill(pid, syscall.SIGUSR2); err != nil {
			fmt.Fprintf(os.Stderr, "failed to resume: %v\n", err)
			os.Exit(1)
		}
		os.WriteFile(internal.StateFile, []byte("active"), 0644)
	} else {
		if err := syscall.Kill(pid, syscall.SIGUSR1); err != nil {
			fmt.Fprintf(os.Stderr, "failed to pause: %v\n", err)
			os.Exit(1)
		}
		os.WriteFile(internal.StateFile, []byte("paused"), 0644)
	}
}
