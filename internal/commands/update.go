package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/DeprecatedLuar/yappers-of-linux/lib/satellite"
)

var Version = "0.0.0"

var updater = satellite.New("DeprecatedLuar", "yappers-of-linux")

func Update(args []string) {
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
		runInstaller()
		return
	}

	// Check for updates
	newVersion, err := updater.CheckForUpdate(Version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to check for updates: %v\n", err)
		return
	}

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
		runInstaller()
	} else {
		fmt.Println("Update cancelled. You can update later by running:")
		fmt.Printf("  yap update --force\n")
	}
}

func runInstaller() {
	if err := updater.RunInstaller(); err != nil {
		fmt.Fprintf(os.Stderr, "\nUpdate failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Update complete!")
}
