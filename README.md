# Voice Typing for Linux - Technical Reference

Voice typing for Linux (Wayland/X11) using OpenAI Whisper via faster-whisper. Distributed as a **portable Go binary** with embedded Python engine and self-healing dependency management. Pre-recording circular buffer ensures no speech loss at utterance start.

## Architecture

**Hybrid Design**: Go CLI wrapper + embedded Python engine

**Entry Point**: `cmd/main.go` → Go binary (`yap`)
- Minimal Go wrapper (command routing only)
- Embeds entire Python codebase and config in binary
- Self-healing system: auto-extracts to `~/.config/yappers-of-linux/`
- Hash-based dependency management (SHA256 of requirements.txt)
- Zero manual setup required

**Python Engine**: Modular architecture in `internal/python/`
- `main.py` - Entry point (argument parsing, instantiation)
- `internal/engine.py` - Core `VoiceTyping` class (orchestration, state machine)
- `internal/capture.py` - Audio capture, circular buffer, WebRTC VAD
- `internal/transcribe.py` - Whisper model loading, transcription
- `internal/output.py` - Terminal display, keyboard injection
- `internal/server.py` - Optional TCP server for state monitoring
- `internal/config.py` - Centralized configuration constants

**Audio Pipeline**:
```
yap binary (Go)
    ↓
SelfHeal() → extract Python to ~/.config/yappers-of-linux/.system/
    ↓
Spawn: python ~/.config/yappers-of-linux/.system/main.py
    ↓
VoiceTyping.run()
    ↓
AudioCapture thread → mic (16kHz mono) → queue → pre-buffer (1.5s circular)
    ↓ (VAD detects speech)
Pre-buffer + new chunks → recording buffer
    ↓ (0.8s silence)
Transcriber.transcribe() → faster-whisper
    ↓
TextOutput.type_text() → terminal + ydotool/xdotool
```

**State Machine** (`VoiceTyping.state` property):
1. `initializing` - Loading Whisper model
2. `ready` - Pre-buffer filling, waiting for VAD trigger
3. `recording` - VAD detected speech, capturing audio
4. `processing` - Transcribing with Whisper
5. `paused` - SIGUSR1 received, processing skipped, model stays loaded

Thread-safe state management via locks (`_state_lock`, `_running_lock`, `_paused_lock`)

## Critical Parameters

All tunable parameters centralized in `internal/python/internal/config.py`:

**AudioConfig** (most important for tuning):
```python
RATE = 16000                      # Whisper requirement
CHUNK_DURATION_MS = 30            # WebRTC VAD requirement (30ms frames)
CHUNK_SIZE = 480                  # 16000 * 30 / 1000
PRE_BUFFER_DURATION_SEC = 1.5     # Captures speech beginning
SILENCE_DURATION_SEC = 0.8        # Silence before transcription trigger
MIN_AUDIO_DURATION_SEC = 0.7      # Discard recordings shorter than this
```

**VADConfig**:
```python
AGGRESSIVENESS = 2                # WebRTC VAD mode (0=least, 3=most aggressive)
```

**TranscriptionConfig**:
```python
BEAM_SIZE = 5                     # Accuracy vs speed tradeoff
TEMPERATURE = 0.0                 # Deterministic output
VAD_MIN_SILENCE_MS = 400          # Additional VAD in Whisper
COMPUTE_TYPE_CPU = "int8"         # Fast mode (float32 in accurate mode)
COMPUTE_TYPE_GPU = "float16"      # GPU compute type
```

**Model warmup** (`internal/python/internal/transcribe.py`):
- Dummy 1s audio transcribed on init to prevent first-run delay
- Audio normalization: Peak normalization applied before transcription

## Usage

**Binary**: `yap` (portable Go executable)

**Commands**:
```bash
yap                        # Show help
yap start                  # Start with default model (tiny), accurate mode
yap start --model small    # Use specific model
yap start --fast           # Fast mode (int8, lower accuracy)
yap start --no-typing      # Only print to terminal, no keyboard injection
yap start --tcp            # Enable TCP server on port 12322
yap start --tcp 9999       # Custom TCP port
yap toggle                 # Smart pause/resume/start
yap pause                  # SIGUSR1 → paused=True, processing skipped
yap resume                 # SIGUSR2 → paused=False, clear buffers
yap stop (or kill)         # SIGTERM → cleanup and exit
yap models                 # Show installed models on disk
```

**Available options**:
- `--model X`: tiny (default), base, small, medium, large
- `--device X`: cpu (default), cuda
- `--language X` / `--lang X`: en (default), es, fr, etc.
- `--tcp [PORT]`: Enable TCP server (default: 12322)
- `--fast`: Use int8 compute type (faster, less accurate)
- `--no-typing`: Disable keyboard typing (terminal only)
- `--cpu` / `--gpu` / `--cuda`: Device shortcuts

**Performance modes** (CPU only):
- **Accurate (default)**: float32 compute type - better transcription quality
- **Fast (`--fast` flag)**: int8 compute type - faster but less accurate

## Dependencies

**System Requirements** (for binary to work):
- `python3` (3.10+) - For venv and running embedded Python code
- `portaudio19-dev` - Required by PyAudio
- `ydotool` + `ydotoold` (Wayland) or `xdotool` (X11) - Keyboard injection
- `go` 1.21+ (only needed for building from source)

**Python Dependencies** (auto-installed by binary):
```
faster-whisper==1.1.1    # CTranslate2-optimized Whisper
numpy>=1.24.0            # Audio array processing
pyaudio==0.2.14          # Microphone access
webrtcvad==2.0.10        # Voice activity detection
requests>=2.31.0         # Required by faster-whisper
```

**Self-Healing System**:
- Binary embeds Python code and default config.toml
- First run: Auto-extracts to `~/.config/yappers-of-linux/.system/`
- Creates Python venv automatically
- Installs dependencies (~2 minutes, one-time)
- Hash-based dependency management:
  - Computes SHA256 of `requirements.txt`
  - Stores hash in `.deps_installed` marker file
  - Subsequent runs: Skip pip install if hash matches
  - Binary updates: Auto-detects changed requirements, reinstalls automatically
- Config extraction: Only if missing (never overwrites user config)
- Python extraction: Always overwrites (system-managed files)
- No manual setup required

**First Run**:
```bash
./yap start              # Auto-creates venv, installs deps, starts
```

**Building from Source**:
```bash
go build -o yap cmd/main.go
```

**Wayland Requirement**:
`ydotoold` daemon MUST be running before voice typing works:
- Check: `ps aux | grep ydotoold` or `systemctl status ydotoold`
- Start: `ydotoold &`
- Socket: `/run/ydotoold/socket` (hardcoded in `internal/python/internal/output.py`)
- NixOS: Configuration in `other/nix/ydotool-service.nix`
- X11 users: Falls back to `xdotool` automatically

## Configuration

**Location**: `~/.config/yappers-of-linux/config.toml` (auto-created on first run)

**Format**: TOML
```toml
notifications = "urgent"     # true/false/urgent
model = "tiny"               # tiny/base/small/medium/large
device = "cpu"               # cpu/cuda
language = "en"              # Language code
fast_mode = false            # true/false
enable_typing = true         # true/false - disable for terminal-only
```

**Notification Modes**:
- `"urgent"` (default) - Critical urgency notifications via notify-send
- `"true"` - Normal urgency notifications
- `"false"` - Disable all notifications

**Auto-Creation**:
- Binary embeds `internal/config.toml` with defaults
- Extracted to `~/.config/yappers-of-linux/config.toml` on first run
- Never overwrites existing user config (respects user edits)
- Delete config.toml and restart binary to restore defaults

**Command-line overrides**: CLI flags take precedence over config.toml

**Example**: `other/examples/config.toml`

## Model Management

Models auto-download from HuggingFace (Systran/faster-whisper) on first use:
- Cache: `~/.cache/huggingface/hub/models--Systran--faster-whisper-<size>/`
- First run: Brief pause for download (30s - few minutes)
- Subsequent runs: Instant load from cache
- Check installed: `yap models`

**Model Sizes**:
- `tiny` (~75MB) - Fastest, least accurate
- `base` (~150MB) - Good for testing
- `small` (~500MB) - Recommended balance
- `medium` (~1.5GB) - Better accuracy
- `large` (~3GB) - Best accuracy, slowest

## TCP Server for State Monitoring

Optional feature inspired by Kanata's TCP port (12321). Enable with `--tcp` flag.

**Design**: Poll-based (not WebSocket)
- Client connects → receives JSON state → disconnects
- Binds to `127.0.0.1` only (localhost, security)
- Default port: 12322 (customizable: `--tcp 9999`)

**JSON Response**:
```json
{
  "state": "ready",
  "model": "tiny",
  "device": "cpu",
  "language": "en",
  "timestamp": 1761481697
}
```

**Use Cases**: Status bars, border color systems (Kanata integration), desktop widgets

**Test**: `nc 127.0.0.1 12322`

**Implementation**: `internal/python/internal/server.py` (`StateServer` class)

## Output Behavior

**Silent startup**: Only prints `loaded: <model>` or `loaded: <model> (fast)`
- ALSA warnings suppressed (stderr redirected during PyAudio init in `internal/python/internal/capture.py`)
- Optional TCP port announcement if enabled
- No startup banner

**Runtime output**:
- `listening` - Ephemeral (overwritten with `\r` and whitespace padding)
- `processing` - Ephemeral (overwritten with `\r` and whitespace padding)
- Transcribed text - Persistent, followed by newline
- Printed to stdout AND typed via ydotool/xdotool (unless `--no-typing`)
- `error: failed to type` - If both ydotool and xdotool fail

**Keyboard typing** (`internal/python/internal/output.py`):
1. Print to stdout first (immediate feedback)
2. Try ydotool with explicit socket path (`/run/ydotoold/socket`)
3. Fallback to xdotool with 10ms delay between keys
4. Append space to all typed text
5. Silent failure if both methods fail (only error message)

## Signal-Based Control

**Registered handlers** (`internal/python/internal/engine.py`):
- `SIGUSR1` → Pause: Set paused flag, stop processing
- `SIGUSR2` → Resume: Clear paused flag, clear buffers, empty queue
- `SIGINT` / `SIGTERM` → Graceful exit

**Commands** (via `internal/commands/`):
- `yap pause` - Sends SIGUSR1 to running process
- `yap resume` - Sends SIGUSR2 to running process
- `yap stop` - Sends SIGTERM to running process
- `yap toggle` - Smart logic: pause if active, resume if paused, start if stopped

**Paused state behavior**:
- Audio queue continues filling (AudioCapture thread still runs)
- Main loop skips processing when paused
- Model stays loaded in memory (no re-initialization on resume)
- Resume clears all buffers to prevent stale audio

## Key Implementation Details

**Threading model** (`internal/python/internal/capture.py` + `internal/python/internal/engine.py`):
- `AudioCapture` runs in daemon thread, continuously reads mic → queue
- Main event loop (`VoiceTyping.run()`) processes queue with 0.1s timeout
- Thread-safe state management via locks
- Exception suppression in reader thread prevents crashes

**Audio processing** (`internal/python/internal/engine.py` + `internal/python/internal/transcribe.py`):
- Discard recordings < `MIN_AUDIO_DURATION_SEC` (0.7s, too short for transcription)
- Convert int16 PCM to float32 normalized [-1, 1]
- Peak normalization applied before transcription (prevents volume issues)
- Join all buffer chunks before numpy conversion

**Process management** (`internal/utils.go`):
- PID file: `~/.local/state/yappers-of-linux/pid`
- State file: `~/.local/state/yappers-of-linux/state` (tracks pause/active for toggle)
- Process discovery: `pgrep -f "python.*yappers-of-linux.*main.py"`
- Multi-instance protection via PID file check

**CUDA Support** (`internal/commands/start.go`):
- Auto-sets `LD_LIBRARY_PATH` to include cuBLAS and cuDNN from venv site-packages
- Required for GPU acceleration on systems without system-wide CUDA libraries
- Paths: `venv/lib/python3.10/site-packages/nvidia/{cublas,cudnn}/lib`

## Project Structure

**Source Code** (embedded in binary):
```
cmd/
  main.go                           # Entry point (Go)
internal/
  commands/                         # One file per command
    parser.go                       # Command routing
    start.go, stop.go, pause.go, resume.go, toggle.go
    models.go, help.go
  config.go                         # TOML configuration loading
  config.toml                       # Default config (embedded)
  selfheal.go                       # Embedding & extraction logic
  utils.go                          # Shared helpers (GetPID, Notify, etc.)
  constants.go                      # Shared constants
  python/                           # Embedded Python source
    main.py                         # Python entry point
    internal/                       # Python modules
      engine.py                     # VoiceTyping orchestration
      capture.py                    # Audio capture, VAD, buffers
      transcribe.py                 # Whisper model, transcription
      output.py                     # Terminal display, keyboard injection
      server.py                     # TCP state server
      config.py                     # Centralized configuration constants
    requirements.txt                # Python dependencies
other/
  examples/config.toml              # Example configuration
  nix/ydotool-service.nix           # NixOS ydotoold service
  install-local.sh                  # Local installation script
go.mod                              # Go dependencies
```

**Runtime Directory** (auto-created):
```
~/.config/yappers-of-linux/
  .system/                          # System-managed files (auto-extracted)
    main.py
    internal/                       # Python modules
    requirements.txt
    venv/                           # Python virtual environment
    .deps_installed                 # SHA256 hash marker file
  config.toml                       # User configuration (editable)
```