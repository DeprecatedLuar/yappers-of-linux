package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Models() {
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
