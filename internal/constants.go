package internal

import (
	"os"
	"path/filepath"
)

func GetPIDFile() string {
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		return filepath.Join(xdg, "yap.pid")
	}
	return "/tmp/yap.pid"
}

func GetStateFile() string {
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		return filepath.Join(xdg, "yap-state")
	}
	return "/tmp/yap-state"
}
