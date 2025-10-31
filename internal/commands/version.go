package commands

import (
	"fmt"
	"os/exec"
	"strings"
)

func ShowVersion() {
	fmt.Printf("Yapping on version %s\n", Version)
	fmt.Println("\033[0;34mby Luar\033[0m")

	// Silently check for updates
	scriptURL := "https://raw.githubusercontent.com/DeprecatedLuar/yappers-of-linux/main/install.sh"
	curlCmd := fmt.Sprintf("curl -sSL %s | bash -s check-update %s", scriptURL, Version)

	cmd := exec.Command("bash", "-c", curlCmd)
	output, err := cmd.Output()

	// If check fails or no update, exit silently
	if err != nil {
		return
	}

	newVersion := strings.TrimSpace(string(output))
	if newVersion != "" {
		fmt.Printf("\nA new version is available: %s\n", newVersion)
		fmt.Println("Run 'yap update' to install.")
	}
}
