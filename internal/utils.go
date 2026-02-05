package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
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

func IsYdotooldRunning() bool {
	cmd := exec.Command("pgrep", "-x", "ydotoold")
	err := cmd.Run()
	return err == nil
}

func HasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func GetYdotoolVersion() (string, error) {
	cmd := exec.Command("dpkg", "-l")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "ydotool") && !strings.Contains(line, "ydotoold") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[2], nil
			}
		}
	}

	return "", fmt.Errorf("version not found")
}

func PrintYdotoolVersionWarning() {
	version, err := GetYdotoolVersion()
	if err != nil {
		return
	}

	if strings.HasPrefix(version, "0.") {
		fmt.Println()
		fmt.Println("ydotool is absolute bizarre. you currently have version " + version)
		fmt.Println("which causes spamming issues and literally breaks the program.")
		fmt.Println()
		fmt.Println("options:")
		fmt.Println("  - start with typing disabled: yap start --no-typing")
		fmt.Println("  - update ydotool to v1.0.0+")
		fmt.Println()
		fmt.Println("see: https://github.com/ReimuNotMoe/ydotool/issues/99")
	}
}

func IsYdotooldSocketStale() bool {
	socketPath := "/tmp/.ydotool_socket"

	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return false
	}

	cmd := exec.Command("ydotool", "type", "")
	output, err := cmd.CombinedOutput()

	if err != nil || strings.Contains(string(output), "ydotoold backend unavailable") {
		return true
	}

	return false
}

func PromptUser(message string) bool {
	fmt.Print(message)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))

	if response == "" || response == "y" || response == "yes" {
		return true
	}

	return false
}

func CleanYdotooldSocket() error {
	socketPath := "/tmp/.ydotool_socket"

	cmd := exec.Command("sudo", "rm", socketPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func StartYdotoold() error {
	cmd := exec.Command("nohup", "ydotoold")
	cmd.Dir = "/tmp"

	if err := cmd.Start(); err != nil {
		return err
	}

	go cmd.Wait()

	time.Sleep(500 * time.Millisecond)

	if !IsYdotooldRunning() {
		return fmt.Errorf("ydotoold failed to start")
	}

	return nil
}

func CheckTypingDependencies() error {
	isWayland := IsWayland()

	if isWayland {
		if !HasCommand("ydotool") || !HasCommand("ydotoold") {
			fmt.Println("didnt find ydotool nor ydotoold installed so it wont be able to type.")
			fmt.Println("but if you want to start it anyways without typing just use --no-typing")
			fmt.Println("(or edit the config file with: yap config)")
			return fmt.Errorf("missing dependencies")
		}
	} else {
		if !HasCommand("xdotool") {
			fmt.Println("didnt find xdotool installed so it wont be able to type.")
			fmt.Println("but if you want to start it anyways without typing just use --no-typing")
			fmt.Println("(or edit the config file with: yap config)")
			return fmt.Errorf("missing dependencies")
		}
	}

	return nil
}

func EnsureYdotoold() error {
	if !IsWayland() {
		return nil
	}

	if err := CheckTypingDependencies(); err != nil {
		return err
	}

	if IsYdotooldRunning() {
		return nil
	}

	socketPath := "/tmp/.ydotool_socket"
	if _, err := os.Stat(socketPath); err == nil {
		fmt.Println("wayland input handling is kind of messy.")
		fmt.Println("ydotool needs a daemon (ydotoold) to work properly, but there's a stale")
		fmt.Println("socket blocking it. need to clean it up once (requires sudo).")
		fmt.Println("after this, ydotoold will auto-start.\n")

		if !PromptUser("remove stale socket? [Y/n]: ") {
			return fmt.Errorf("cancelled")
		}

		if err := CleanYdotooldSocket(); err != nil {
			return fmt.Errorf("failed to clean socket: %v", err)
		}
	}

	if err := StartYdotoold(); err != nil {
		PrintYdotoolVersionWarning()
		return fmt.Errorf("failed to start ydotoold: %v", err)
	}

	return nil
}
