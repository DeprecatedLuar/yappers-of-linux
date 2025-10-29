package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"yappers-of-linux/internal"
)

func Config() {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "yappers-of-linux")
	configFile := filepath.Join(configDir, "config.toml")

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := internal.SelfHeal(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create config: %v\n", err)
			os.Exit(1)
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, configFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to open editor: %v\n", err)
		os.Exit(1)
	}
}
