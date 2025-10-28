package internal

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed all:python
var pythonFiles embed.FS

const (
	ConfigDir = ".config/yappers-of-linux"
	SystemDir = ".system"
)

func SelfHeal() error {
	if err := ensureConfigDir(); err != nil {
		return fmt.Errorf("failed to ensure config dir: %w", err)
	}

	if err := extractPythonFiles(); err != nil {
		return fmt.Errorf("failed to extract python files: %w", err)
	}

	if err := ensureVenv(); err != nil {
		return fmt.Errorf("failed to ensure venv: %w", err)
	}

	if err := installDependencies(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	return nil
}

func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ConfigDir), nil
}

func GetSystemDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, SystemDir), nil
}

func ensureConfigDir() error {
	systemDir, err := GetSystemDir()
	if err != nil {
		return err
	}

	return os.MkdirAll(systemDir, 0755)
}

func extractPythonFiles() error {
	systemDir, err := GetSystemDir()
	if err != nil {
		return err
	}

	return fs.WalkDir(pythonFiles, "python", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel("python", path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(systemDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		data, err := pythonFiles.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, 0644)
	})
}

func ensureVenv() error {
	systemDir, err := GetSystemDir()
	if err != nil {
		return err
	}

	venvPath := filepath.Join(systemDir, "venv")
	venvPython := filepath.Join(venvPath, "bin", "python")

	if _, err := os.Stat(venvPython); err == nil {
		return nil
	}

	cmd := exec.Command("python3", "-m", "venv", venvPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func installDependencies() error {
	systemDir, err := GetSystemDir()
	if err != nil {
		return err
	}

	venvPip := filepath.Join(systemDir, "venv", "bin", "pip")
	requirementsPath := filepath.Join(systemDir, "requirements.txt")

	currentHash, err := hashFile(requirementsPath)
	if err != nil {
		return err
	}

	installedMarker := filepath.Join(systemDir, ".deps_installed")
	if installedHash, err := os.ReadFile(installedMarker); err == nil {
		if string(installedHash) == currentHash {
			return nil
		}
	}

	cmd := exec.Command(venvPip, "install", "-r", requirementsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return os.WriteFile(installedMarker, []byte(currentHash), 0644)
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
