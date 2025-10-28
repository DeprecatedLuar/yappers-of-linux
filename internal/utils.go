package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

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
