package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	StateFile = "/tmp/yap-state"
	PIDFile   = "/tmp/yap.pid"
)

func ShowHelp() {
	help := `usage: yap <command> [options]

commands:
  start [options]     Start voice typing (default: --model tiny)
  toggle [options]    Smart pause/resume/start
  pause               Pause listening
  resume              Resume listening
  stop (kill)         Stop voice typing
  models              Show installed models

options:
  --model X           Model size: tiny, base, small, medium, large
  --device X          Device: cpu, cuda
  --language X        Language code (default: en)
  --tcp [PORT]        Enable TCP server (default port: 12322)
  --fast              Use fast mode (int8, less accurate but faster)

modes:
  default             Accurate mode (float32, better quality)
  --fast              Fast mode (int8, lower quality but faster)

models automatically download on first use`

	fmt.Println(help)
}

func ShowModels() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("no models installed")
		return
	}

	cacheDir := filepath.Join(homeDir, ".cache", "huggingface", "hub")
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		fmt.Println("no models installed")
		return
	}

	var models []string
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "faster-whisper-") {
			model := strings.TrimPrefix(entry.Name(), "models--Systran--faster-whisper-")
			models = append(models, model)
		}
	}

	if len(models) == 0 {
		fmt.Println("no models installed")
	} else {
		fmt.Printf("installed: %s\n", strings.Join(models, ", "))
	}
}

func GetProjectDir() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return "", err
	}

	return filepath.Dir(execPath), nil
}

func GetPID() (int, error) {
	data, err := os.ReadFile(PIDFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, err
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		os.Remove(PIDFile)
		return 0, err
	}

	return pid, nil
}

func Notify(message string) {
	cmd := exec.Command("notify-send", "-i", "audio-input-microphone", "-u", "critical", "Yap", message)
	cmd.Start()
}

func Start(args []string) {
	pid, err := GetPID()
	if err == nil {
		fmt.Printf("already running (pid %d)\n", pid)
		return
	}

	projectDir, err := GetProjectDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to find project directory: %v\n", err)
		os.Exit(1)
	}

	venvPython := filepath.Join(projectDir, "venv", "bin", "python")
	script := filepath.Join(projectDir, "main.py")

	Notify("Yapping started")

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
	if err := os.WriteFile(PIDFile, pidData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to write pid file: %v\n", err)
	}

	cmd.Wait()
	os.Remove(PIDFile)
	os.Remove(StateFile)
}

func Toggle(args []string) {
	pid, err := GetPID()
	if err != nil {
		os.Remove(StateFile)
		Notify("Yapping started")
		Start(args)
		return
	}

	stateData, err := os.ReadFile(StateFile)
	state := strings.TrimSpace(string(stateData))

	if err == nil && state == "paused" {
		Notify("Yapping started")
		if err := syscall.Kill(pid, syscall.SIGUSR2); err != nil {
			fmt.Fprintf(os.Stderr, "failed to resume: %v\n", err)
			os.Exit(1)
		}
		os.WriteFile(StateFile, []byte("active"), 0644)
	} else {
		if err := syscall.Kill(pid, syscall.SIGUSR1); err != nil {
			fmt.Fprintf(os.Stderr, "failed to pause: %v\n", err)
			os.Exit(1)
		}
		os.WriteFile(StateFile, []byte("paused"), 0644)
	}
}

func Pause() {
	pid, err := GetPID()
	if err != nil {
		fmt.Println("not running")
		os.Exit(1)
	}

	if err := syscall.Kill(pid, syscall.SIGUSR1); err != nil {
		fmt.Fprintf(os.Stderr, "failed to pause: %v\n", err)
		os.Exit(1)
	}

	os.WriteFile(StateFile, []byte("paused"), 0644)
}

func Resume() {
	pid, err := GetPID()
	if err != nil {
		fmt.Println("not running")
		os.Exit(1)
	}

	Notify("Yapping started")

	if err := syscall.Kill(pid, syscall.SIGUSR2); err != nil {
		fmt.Fprintf(os.Stderr, "failed to resume: %v\n", err)
		os.Exit(1)
	}

	os.WriteFile(StateFile, []byte("active"), 0644)
}

func Stop() {
	pid, err := GetPID()
	if err != nil {
		fmt.Println("not running")
		os.Exit(1)
	}

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "failed to stop: %v\n", err)
		os.Exit(1)
	}

	os.Remove(PIDFile)
	os.Remove(StateFile)
}
