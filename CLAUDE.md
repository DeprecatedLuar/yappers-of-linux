# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Voice typing solution for Linux (Wayland/X11) using OpenAI's Whisper model via faster-whisper. Key innovation is a **pre-recording circular buffer** that captures speech before VAD triggers, ensuring no words are missed at the beginning.

## Code Architecture

The project follows a **modular design** with clear separation of concerns:

**Entry Point**: `cmd/main.go`
- Ultra-minimal Go binary (calls parser)
- Portable, self-contained executable

**Go CLI Layer**: `internal/`
- `commands/` - One file per command (start.go, stop.go, pause.go, resume.go, toggle.go, models.go, help.go)
- `commands/parser.go` - Routes commands to handlers
- `selfheal.go` - Embeds Python code, auto-extracts to `~/.config/yappers-of-linux/.system/`
- `utils.go` - Shared helpers (GetPID, Notify, etc.)
- `constants.go` - Shared constants

**Python Entry Point**: `internal/python/main.py`
- Argument parsing and validation
- Instantiates `VoiceTyping` engine
- Minimal logic, delegates to `internal/` modules

**Core Module**: `internal/engine.py` (`VoiceTyping` class)
- Orchestrates all components
- Manages state machine (initializing → ready → recording → processing → paused)
- Coordinates signal handlers (SIGUSR1/SIGUSR2)
- Thread-safe state management via locks
- Main event loop: processes audio chunks, triggers transcription

**Python Component Modules** (`internal/python/internal/`):
- `capture.py` (`AudioCapture`) - PyAudio stream, circular pre-buffer, recording buffer, WebRTC VAD
- `transcribe.py` (`Transcriber`) - Whisper model loading, warmup, audio normalization, transcription with configurable compute types (float32/int8)
- `output.py` (`TextOutput`) - Terminal display, ydotool/xdotool keyboard injection
- `server.py` (`StateServer`) - Optional TCP server for state monitoring
- `config.py` - Centralized configuration constants (all tunable parameters)

**Pipeline Flow**:
```
yap binary (Go)
    ↓
SelfHeal() → extract Python to ~/.config/yappers-of-linux/.system/
    ↓
Spawn: python ~/.config/yappers-of-linux/.system/main.py
    ↓
VoiceTyping.run()
    ↓
AudioCapture starts thread → reads mic → fills queue
    ↓
Main loop gets chunks → pre-buffer (1.5s circular)
    ↓
VAD detects speech → copy pre-buffer + continue recording
    ↓
0.8s silence → Transcriber.transcribe()
    ↓
TextOutput.type_text() → terminal + ydotool/xdotool
```

## Key Design Patterns

**Thread Safety**:
- State properties use locks (`_state_lock`, `_running_lock`, `_paused_lock`)
- Audio capture runs in daemon thread, main loop processes queue
- TCP server runs in daemon thread, polls state via callback

**Configuration Centralization**:
- All magic numbers in `internal/config.py` classes:
  - `AudioConfig` - Sample rates, buffer durations, chunk sizes
  - `VADConfig` - WebRTC aggressiveness level
  - `TranscriptionConfig` - Whisper beam size, temperature, VAD filtering
  - `DisplayConfig`, `TCPConfig`, `ThreadConfig`

**State Machine**:
1. `initializing` - Loading Whisper model
2. `ready` - Pre-buffer filling, waiting for VAD
3. `recording` - VAD detected speech, capturing audio
4. `processing` - Transcribing audio with Whisper
5. `paused` - SIGUSR1 received, processing skipped, model stays loaded

State transitions managed by `VoiceTyping.state` property setter (thread-safe, updates display).

## Running the Project

**Portable Binary** (only method needed):
```bash
./yap                        # Show help
./yap start                  # Start with default model (tiny), accurate mode
./yap start --model small    # Use specific model
./yap start --fast           # Fast mode (int8, lower accuracy)
./yap start --model base --fast --tcp  # Combine options
./yap toggle                 # Smart pause/resume/start
./yap pause / resume / stop  # Control running instance
./yap models                 # Show installed models on disk
```

**First Run**:
- Binary auto-creates `~/.config/yappers-of-linux/.system/`
- Extracts embedded Python code
- Creates Python venv
- Installs dependencies (one-time, ~2 minutes)
- Subsequent runs are instant

**Self-Healing**:
- Delete `~/.config/yappers-of-linux/` → binary recreates automatically
- Works from any location (fully portable)
- No manual setup required

**Available Options**:
- `--model`: tiny (default), base, small, medium, large
- `--device`: cpu (default), cuda
- `--language`: en (default), es, fr, etc.
- `--tcp [PORT]`: Enable TCP server for state monitoring (default: 12322)
- `--fast`: Use fast mode (int8 compute type, trades accuracy for speed)

**Performance Modes** (CPU only):
- **Accurate (default)**: float32 compute type - better transcription quality
- **Fast (`--fast` flag)**: int8 compute type - faster but less accurate

## Critical Configuration Parameters

Located in `internal/config.py`:

**AudioConfig** (most important for tuning):
- `RATE = 16000` - Whisper requirement
- `CHUNK_DURATION_MS = 30` - WebRTC VAD requirement
- `PRE_BUFFER_DURATION_SEC = 1.5` - Captures speech beginning
- `SILENCE_DURATION_SEC = 0.8` - Silence before transcription
- `MIN_AUDIO_DURATION_SEC = 0.7` - Discard recordings shorter than this

**VADConfig**:
- `AGGRESSIVENESS = 2` - WebRTC VAD mode (0=least, 3=most aggressive)

**TranscriptionConfig**:
- `BEAM_SIZE = 5` - Accuracy vs speed
- `TEMPERATURE = 0.0` - Deterministic output
- `VAD_MIN_SILENCE_MS = 400` - Additional VAD in Whisper
- `COMPUTE_TYPE_CPU = "int8"` / `COMPUTE_TYPE_GPU = "float16"`
- Audio normalization: Peak normalization applied before transcription (prevents volume-related errors)

## CLI Wrapper Implementation

**Location**: `cmd/yap` (bash script)
- Hardcoded project path: `$HOME/Workspace/tools/homemade/yap`
- Process discovery: `pgrep -f "$VENV_PYTHON.*$SCRIPT"`
- Signal-based control: SIGUSR1 (pause), SIGUSR2 (resume), SIGTERM (stop)
- State tracking: `/tmp/yap-state` file (tracks paused/active for toggle)
- Desktop notifications: Uses `notify-send` for start/resume feedback
- Multi-instance protection: Checks PID before starting

**Important**: If copying to `~/Workspace/tools/bin/`, update `PROJECT_DIR` variable (line 5) to point to correct location.

## Wayland Requirements

**Critical**: `ydotoold` service MUST be running before voice typing works on Wayland.
- Check: `ps aux | grep ydotoold` or `systemctl status ydotoold`
- Start manually: `ydotoold &`
- Socket path: `/run/ydotoold/socket` (hardcoded in `internal/output.py`)
- NixOS: Configuration available in `other/nix/ydotool-service.nix` (requires username update)

X11 users: Falls back to `xdotool` automatically (no special setup).

## Dependencies

**System Requirements**:
- `python3` - Python 3.10+ (for venv and running Python code)
- `portaudio19-dev` - Required by PyAudio
- `ydotool` + `ydotoold` (Wayland) or `xdotool` (X11)
- `go` 1.21+ (only for building from source)

**Python Dependencies** (auto-installed by binary):
- `faster-whisper==1.1.1` - CTranslate2-optimized Whisper
- `numpy>=1.24.0` - Audio array processing
- `pyaudio==0.2.14` - Microphone access
- `webrtcvad==2.0.10` - Voice activity detection

**Go Dependencies** (for development):
- Standard library only (no external dependencies)

**Building from Source**:
```bash
go build -o yap cmd/main.go
```

**No Python Setup Needed**:
- Binary handles everything automatically on first run
- Creates venv and installs dependencies
- No manual pip install required

## Model Management

Models auto-download from HuggingFace (Systran/faster-whisper) on first use:
- Cache: `~/.cache/huggingface/hub/models--Systran--faster-whisper-<size>/`
- First run: Brief pause for download (30s - few minutes depending on size)
- Subsequent runs: Instant load from cache
- Check installed: `yap models`

Model sizes:
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

**Use Cases**: Status bars, border color systems (like Kanata integration), desktop widgets

**Test**: `nc 127.0.0.1 12322`

Implementation: `internal/server.py` (`StateServer` class)

## Common Development Tasks

**Changing VAD sensitivity**: Adjust `VADConfig.AGGRESSIVENESS` in `internal/config.py` (0-3)

**Changing pre-buffer duration**: Adjust `AudioConfig.PRE_BUFFER_DURATION_SEC` in `internal/config.py`

**Changing silence threshold**: Adjust `AudioConfig.SILENCE_DURATION_SEC` in `internal/config.py`

**Testing without typing**: Comment out `self.output.type_text(text)` in `internal/engine.py:174`, text will print to terminal only

**Debugging state transitions**: Add print statements in `VoiceTyping.state` setter (`internal/engine.py:73`)

**Running with different socket path**: Update `socket_path` in `internal/output.py:63` (ydotool call)

## Project Structure

**Source Code**:
```
cmd/
  main.go                   # Entry point (Go)
internal/
  commands/                 # One file per command
    parser.go               # Command routing
    start.go, stop.go, pause.go, resume.go, toggle.go
    models.go, help.go
  selfheal.go               # Embedding & extraction logic
  utils.go                  # Shared helpers
  constants.go              # Shared constants
  python/                   # Embedded Python source
    main.py                 # Python entry point
    internal/               # Python modules
      engine.py, capture.py, transcribe.py
      output.py, server.py, config.py
    requirements.txt        # Python dependencies
other/
  examples/                 # Example configurations
  nix/                      # NixOS configurations
venv/                       # Development venv (not embedded)
go.mod                      # Go dependencies
```

**Runtime Directory** (auto-created):
```
~/.config/yappers-of-linux/
  .system/                  # System-managed files (auto-extracted)
    main.py
    internal/               # Python modules
    requirements.txt
    venv/                   # Python virtual environment
    .deps_installed         # Marker file
  config.toml               # User configuration (optional, future)
```

## Known Issues & Solutions

**"No words at beginning"**: Should not happen with 1.5s pre-buffer. If it does, increase `AudioConfig.PRE_BUFFER_DURATION_SEC` in `internal/config.py`.

**"ydotool: failed to connect socket"**: ydotoold not running. Start with `ydotoold &` or `systemctl start ydotoold`.

**"already running"**: Another instance active. Use `yap stop` or `yap toggle` to pause/resume.

**High CPU usage**:
- Use smaller model (`--model tiny`)
- Enable GPU (`--device cuda`, requires PyTorch with CUDA)
- Reduce VAD aggressiveness in `internal/config.py`

**Poor accuracy**:
- Larger model (`--model medium` or `large`)
- Check mic quality: `arecord -V stereo -r 16000 -f S16_LE -d 5 test.wav`
- Adjust VAD aggressiveness (higher = more sensitive)

**Text types into wrong window**: Ensure text input field has focus before speaking. Typing is character-by-character and can be interrupted.

**ALSA warnings**: Already suppressed in `internal/capture.py:39-55` via stderr redirect during PyAudio init.
