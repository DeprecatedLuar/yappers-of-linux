package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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

	// Clean up previous output file (ephemeral, always fresh)
	configDir, err := internal.GetConfigDir()
	if err == nil {
		outputFile := filepath.Join(configDir, "output.txt")
		os.Remove(outputFile) // Ignore error if doesn't exist
	}

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
	enableTyping := cfg.EnableTyping

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--model" && i+1 < len(args) {
			model = args[i+1]
		} else if arg == "--device" && i+1 < len(args) {
			device = args[i+1]
		} else if arg == "--language" && i+1 < len(args) {
			language = args[i+1]
		} else if arg == "--lang" && i+1 < len(args) {
			language = args[i+1]
		} else if arg == "--fast" {
			fastMode = true
		} else if arg == "--no-typing" {
			enableTyping = false
		} else if arg == "--gpu" || arg == "--cuda" {
			device = "cuda"
		} else if arg == "--cpu" {
			device = "cpu"
		}
	}

	if enableTyping {
		if err := internal.EnsureYdotoold(); err != nil {
			fmt.Fprintf(os.Stderr, "ydotoold setup failed: %v\n", err)
			os.Exit(1)
		}
	}

	pythonArgs := []string{script, "--model", model, "--device", device, "--language", language}
	if fastMode {
		pythonArgs = append(pythonArgs, "--fast")
	}
	if !enableTyping {
		pythonArgs = append(pythonArgs, "--no-typing")
	}
	if cfg.OutputFile {
		pythonArgs = append(pythonArgs, "--output-file")
	}

	cmd := exec.Command(venvPython, pythonArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	// Capture stderr to watch for state markers
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create stderr pipe: %v\n", err)
		os.Exit(1)
	}

	// Set LD_LIBRARY_PATH for CUDA libraries (cuBLAS, cuDNN)
	cmd.Env = os.Environ()
	sitePackages := filepath.Join(systemDir, "venv", "lib", "python3.10", "site-packages")
	cudaLibPaths := filepath.Join(sitePackages, "nvidia", "cublas", "lib") + ":" +
		filepath.Join(sitePackages, "nvidia", "cudnn", "lib")
	currentLdPath := os.Getenv("LD_LIBRARY_PATH")
	if currentLdPath != "" {
		cudaLibPaths = cudaLibPaths + ":" + currentLdPath
	}
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+cudaLibPaths)

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start: %v\n", err)
		os.Exit(1)
	}

	pidData := []byte(strconv.Itoa(cmd.Process.Pid))
	if err := os.WriteFile(internal.PIDFile, pidData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to write pid file: %v\n", err)
	}

	// Read stderr line by line, watching for state markers
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()

			// Watch for state markers
			if strings.Contains(line, "SYSTEM_READY") {
				internal.Notify("Ready to listen", "start", cfg)
			} else {
				// Print other stderr output
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}()

	cmd.Wait()
	os.Remove(internal.PIDFile)
	os.Remove(internal.StateFile)
}
