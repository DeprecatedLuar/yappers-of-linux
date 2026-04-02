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

