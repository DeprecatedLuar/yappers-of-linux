package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"yappers-of-linux/internal"
)

func Start(args []string) {
	pid, err := internal.GetPID()
	if err == nil {
		fmt.Printf("already running (pid %d)\n", pid)
		return
	}

	if err := internal.SelfHeal(); err != nil {
		fmt.Fprintf(os.Stderr, "setup failed: %v\n", err)
		os.Exit(1)
	}

	systemDir, err := internal.GetSystemDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get system directory: %v\n", err)
		os.Exit(1)
	}

	venvPython := filepath.Join(systemDir, "venv", "bin", "python")
	script := filepath.Join(systemDir, "main.py")

	internal.Notify("Yapping started")

	pythonArgs := []string{script}
	if len(args) == 0 {
		pythonArgs = append(pythonArgs, "--model", "tiny")
	} else {
		pythonArgs = append(pythonArgs, args...)
	}

	cmd := exec.Command(venvPython, pythonArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start: %v\n", err)
		os.Exit(1)
	}

	pidData := []byte(strconv.Itoa(cmd.Process.Pid))
	if err := os.WriteFile(internal.PIDFile, pidData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to write pid file: %v\n", err)
	}

	cmd.Wait()
	os.Remove(internal.PIDFile)
	os.Remove(internal.StateFile)
}
