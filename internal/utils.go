package internal

import (
	"fmt"
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

func Notify(message string, event string, cfg *Config) {
	notifCfg := ParseNotifications(cfg.Notifications)

	if !notifCfg.ShouldNotify(event) {
		return
	}

	urgency := "normal"
	if notifCfg.Urgent {
		urgency = "critical"
	}

	cmd := exec.Command("notify-send", "-i", "audio-input-microphone", "-u", urgency, "Yap", message)
	if err := cmd.Start(); err == nil {
		go cmd.Wait()
	}
}

func IsWayland() bool {
	sessionType := strings.ToLower(os.Getenv("XDG_SESSION_TYPE"))
	return sessionType == "wayland"
}

func HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func CheckTypingDependencies() error {
	if IsWayland() {
		if !HasCommand("wtype") {
			fmt.Println("wtype not found (required for Wayland typing).")
			fmt.Println("install it or start without typing: yap start --no-typing")
			return fmt.Errorf("missing dependencies")
		}
	} else {
		if !HasCommand("xdotool") {
			fmt.Println("xdotool not found (required for X11 typing).")
			fmt.Println("install it or start without typing: yap start --no-typing")
			return fmt.Errorf("missing dependencies")
		}
	}

	return nil
}
