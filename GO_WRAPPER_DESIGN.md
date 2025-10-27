# Go Wrapper Design

## Problem
Current Python implementation requires:
- Bash script wrapper with path tracking
- Manual venv management
- Copying script to ~/Workspace/tools/bin/
- Bootstrap complexity

## Solution
Single Go binary that manages Python worker automatically.

## Architecture

```
~/Workspace/tools/bin/yap        # Single Go binary (only thing user downloads)
~/.config/yap/
    ├── venv/                    # Auto-created Python venv
    ├── main.py                  # Auto-extracted from embedded Go binary
    └── requirements.txt         # Auto-extracted
```

## Key Features

### 1. Self-Healing Setup
Every command run checks if environment exists:
- Missing `~/.config/yap/venv/`? → Create it
- Missing dependencies? → `pip install -r requirements.txt`
- Outdated Python script? → Re-extract from embedded binary
- **No explicit setup command needed** - just works

### 2. Zero-Config Distribution
```bash
# User installation (single command)
curl -L https://github.com/user/yap/releases/latest/download/yap -o ~/Workspace/tools/bin/yap
chmod +x ~/Workspace/tools/bin/yap

# First run auto-sets up everything
yap start
```

### 3. Embedded Resources
Go binary embeds via `go:embed`:
- `main.py` (Python worker script)
- `requirements.txt` (Python dependencies)
- Extracted to `~/.config/yap/` on first run

## Implementation Notes

### Go Binary Responsibilities
- CLI argument parsing and command routing
- Process lifecycle management (start/stop/pause/resume)
- Signal handling (SIGUSR1/SIGUSR2/SIGTERM to Python worker)
- State file management (`/tmp/yap-state`)
- Self-healing environment check on EVERY command
- Spawn Python worker: `~/.config/yap/venv/bin/python ~/.config/yap/main.py [flags]`

### Python Script Responsibilities
- Audio capture and VAD (unchanged from current implementation)
- Whisper transcription
- Keyboard input via ydotool/xdotool
- TCP server for state monitoring (optional)

### Commands
```bash
yap start [--model tiny] [--device cpu] [--tcp 12322]  # Start worker (auto-setup if needed)
yap toggle [flags]                                      # Smart pause/resume/start
yap stop                                                # Kill worker
yap pause                                               # SIGUSR1
yap resume                                              # SIGUSR2
yap models                                              # List installed Whisper models
yap doctor                                              # Check environment health
```

## Self-Healing Logic (runs on every command)

```go
func ensureEnvironment() error {
    configDir := "~/.config/yap"

    // 1. Create config dir if missing
    if !exists(configDir) {
        mkdir(configDir)
    }

    // 2. Extract embedded Python files if missing or outdated
    if !exists(configDir + "/main.py") || needsUpdate() {
        extractEmbedded("main.py", configDir)
        extractEmbedded("requirements.txt", configDir)
    }

    // 3. Create venv if missing
    if !exists(configDir + "/venv") {
        exec("python3 -m venv " + configDir + "/venv")
    }

    // 4. Install/update dependencies
    exec(configDir + "/venv/bin/pip install -r " + configDir + "/requirements.txt")

    return nil
}
```

## Build Process

```bash
# Embed Python files in Go binary
go build -o yap cmd/yap/main.go

# Binary now contains everything needed
# Distribute single file: ~/Workspace/tools/bin/yap
```

## Benefits

✅ Single binary distribution (Go ergonomics)
✅ No path tracking hell (fixed config location)
✅ Auto-setup on first run (zero config)
✅ Leverages Python ML ecosystem (no AI rewrite needed)
✅ Self-healing (survives config corruption)
✅ Easy updates (just replace binary)

## Future Enhancements

- `yap update` → Re-download latest Python script from GitHub
- Version pinning for Python dependencies
- Multi-platform builds (Linux/macOS/Windows)
- Man page generation
- Bash/zsh completion

## Migration from Current Setup

1. Build Go binary with embedded Python script
2. Replace bash wrapper in ~/Workspace/tools/bin/yap
3. Delete old project folder (optional - Go binary is now self-contained)
4. First run auto-creates ~/.config/yap/

No user data loss - state file location unchanged.
