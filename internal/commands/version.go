package commands

import (
	"fmt"
)

func ShowVersion() {
	fmt.Printf("Yapping on version %s\n", Version)
	fmt.Println("\033[0;34mby Luar\033[0m")

	// Silently check for updates
	newVersion, err := updater.CheckForUpdate(Version)
	if err != nil {
		return
	}

	if newVersion != "" {
		fmt.Printf("\nA new version is available: %s\n", newVersion)
		fmt.Println("Run 'yap update' to install.")
	}
}
