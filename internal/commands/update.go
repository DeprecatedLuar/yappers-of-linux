package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var Version = "0.0.0"

func Update(args []string) {
	scriptURL := "https://raw.githubusercontent.com/DeprecatedLuar/yappers-of-linux/main/install.sh"

	// Check for --force flag
	force := false
	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
			break
		}
	}

	if force {
		fmt.Println("Force update enabled. Installing latest version...")
		runInstallScript(scriptURL)
		return
	}

	// Check for updates
	curlCmd := fmt.Sprintf("curl -sSL %s | bash -s check-update %s", scriptURL, Version)
	cmd := exec.Command("bash", "-c", curlCmd)
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to check for updates: %v\n", err)
		return
	}

	newVersion := strings.TrimSpace(string(output))
	if newVersion == "" {
		fmt.Println("You're golden. Already on the latest version.")
		return
	}

	// New version available - prompt user
	fmt.Printf("A new version is available: %s (current: %s)\n", newVersion, Version)
	fmt.Print("Do you want to install it? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input: %v\n", err)
		return
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response == "y" || response == "yes" {
		fmt.Println("\nInstalling update...")
		runInstallScript(scriptURL)
	} else {
		fmt.Println("Update cancelled. You can update later by running:")
		fmt.Printf("  curl -sSL %s | bash\n", scriptURL)
	}
}

func runInstallScript(scriptURL string) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("curl -sSL %s | bash", scriptURL))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\nUpdate failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Update complete!")
}
