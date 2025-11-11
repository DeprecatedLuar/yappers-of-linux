# Satellite Go Library

Dead simple Go API for [Satellite](https://github.com/DeprecatedLuar/the-satellite)-powered updates.

## Installation

```bash
go get github.com/DeprecatedLuar/yappers-of-linux/lib/satellite
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/DeprecatedLuar/yappers-of-linux/lib/satellite"
)

func main() {
    // Configure once for your repository
    updater := satellite.New("DeprecatedLuar", "yappers-of-linux")

    // Check for updates
    newVersion, err := updater.CheckForUpdate("v1.0.0")
    if err != nil {
        panic(err)
    }

    if newVersion != "" {
        fmt.Printf("Update available: %s\n", newVersion)

        // Install update
        if err := updater.RunInstaller(); err != nil {
            panic(err)
        }
    }
}
```

## Usage Patterns

### Update Command

```go
package commands

import "github.com/DeprecatedLuar/yappers-of-linux/lib/satellite"

var updater = satellite.New("DeprecatedLuar", "yappers-of-linux")

func Update(currentVersion string) {
    newVersion, err := updater.CheckForUpdate(currentVersion)
    if err != nil {
        // Handle error
        return
    }

    if newVersion == "" {
        fmt.Println("Already on latest version")
        return
    }

    fmt.Printf("Update available: %s\n", newVersion)

    // Prompt user, then:
    if err := updater.RunInstaller(); err != nil {
        // Handle error
    }
}
```

### Version Command with Update Notification

```go
package commands

import "github.com/DeprecatedLuar/yappers-of-linux/lib/satellite"

var updater = satellite.New("DeprecatedLuar", "yappers-of-linux")

func Version(currentVersion string) {
    fmt.Printf("yap version %s\n", currentVersion)

    // Silent check for updates
    if newVersion, err := updater.CheckForUpdate(currentVersion); err == nil && newVersion != "" {
        fmt.Printf("\n→ Update available: %s (run 'yap update')\n", newVersion)
    }
}
```

### Startup Update Check

```go
package main

import "github.com/DeprecatedLuar/yappers-of-linux/lib/satellite"

var updater = satellite.New("DeprecatedLuar", "yappers-of-linux")

func main() {
    // Background update check (don't block startup)
    go func() {
        if newVersion, _ := updater.CheckForUpdate(Version); newVersion != "" {
            fmt.Printf("\n[Update available: %s - run 'yap update']\n\n", newVersion)
        }
    }()

    // Rest of your app...
}
```

## API Reference

### `New(repoUser, repoName string) *Updater`

Creates a new Updater configured for the specified GitHub repository.

**Parameters:**
- `repoUser` - GitHub username (e.g., "DeprecatedLuar")
- `repoName` - Repository name (e.g., "yappers-of-linux")

**Returns:** Configured `*Updater` instance

---

### `(*Updater) CheckForUpdate(currentVersion string) (string, error)`

Queries Satellite to check if a newer version exists.

**Parameters:**
- `currentVersion` - Current version string (e.g., "v1.0.0")

**Returns:**
- New version string if available (e.g., "v1.2.0")
- Empty string if already up-to-date
- Error if check fails

**Example:**
```go
newVersion, err := updater.CheckForUpdate("v1.0.0")
if err != nil {
    // Network error or Satellite unavailable
}
if newVersion != "" {
    // Update is available
}
```

---

### `(*Updater) RunInstaller() error`

Downloads and executes the project's install script via Satellite.

**Returns:** Error if installation fails

**Notes:**
- Streams output to stdout/stderr (interactive)
- Handles user input (for prompts)
- Blocks until installation completes

**Example:**
```go
if err := updater.RunInstaller(); err != nil {
    fmt.Fprintf(os.Stderr, "Installation failed: %v\n", err)
    os.Exit(1)
}
fmt.Println("✓ Update complete!")
```

## How It Works

1. **CheckForUpdate** calls Satellite's `check-update` command via curl
2. Satellite queries GitHub API for latest release/tag
3. Compares versions and returns newer version if available
4. **RunInstaller** downloads your project's `install.sh` script
5. `install.sh` calls Satellite's `install` command
6. Satellite handles binary download, build, and installation

## Requirements

- `curl` - For downloading Satellite
- `bash` - For executing installation scripts
- Internet connection to reach GitHub

## Future Plans

This library will be extracted to a standalone repository (`satellite-go`) for use across all CLI projects.
