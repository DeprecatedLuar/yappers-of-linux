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

	cfg := internal.LoadConfig()

	systemDir, err := internal.GetSystemDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get system directory: %v\n", err)
		os.Exit(1)
	}

	venvPython := filepath.Join(systemDir, "venv", "bin", "python")
	script := filepath.Join(systemDir, "main.py")

	model := cfg.Model
	device := cfg.Device
	language := cfg.Language
	fastMode := cfg.FastMode

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--model" && i+1 < len(args) {
			model = args[i+1]
		} else if arg == "--device" && i+1 < len(args) {
			device = args[i+1]
		} else if arg == "--language" && i+1 < len(args) {
			language = args[i+1]
		} else if arg == "--fast" {
			fastMode = true
		}
	}

	pythonArgs := []string{script, "--model", model, "--device", device, "--language", language}
	if fastMode {
		pythonArgs = append(pythonArgs, "--fast")
	}

	mode := "accurate"
	if fastMode {
		mode = "fast"
	}
	fmt.Printf("device: %s | language: %s | mode: %s\n", device, language, mode)

	internal.Notify("Yapping started", cfg)

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
