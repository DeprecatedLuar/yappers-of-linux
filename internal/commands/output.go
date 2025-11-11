package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"yappers-of-linux/internal"
)

func Output() {
	cfg := internal.LoadConfig()

	if !cfg.OutputFile {
		fmt.Fprintln(os.Stderr, "output file is disabled (set output_file = true in config.toml)")
		os.Exit(1)
	}

	configDir, err := internal.GetConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get config directory: %v\n", err)
		os.Exit(1)
	}

	outputPath := filepath.Join(configDir, "output.txt")

	data, err := os.ReadFile(outputPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "output file not found (start yapping first)")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "failed to read output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(string(data))
}
