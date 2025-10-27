# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Voice typing solution for Linux (Wayland/X11) using OpenAI's Whisper model via faster-whisper. Key innovation is a **pre-recording circular buffer** that captures speech before VAD triggers, ensuring no words are missed at the beginning.

## Core Architecture

**Single-file Implementation** (`main.py`):
- `VoiceTyping` class manages entire pipeline
- Threaded audio reader (`audio_reader_thread`) continuously reads from microphone into queue
- Main loop processes audio from queue, manages recording state machine
- Pre-buffer: 1.5s circular buffer (always recording)
- Main buffer: 4.0s duration for context
- Silence detection: 0.8s triggers transcription

**Audio Pipeline**:
```
Microphone → PyAudio → Audio Queue → Circular Pre-buffer
                                    ↓
                            VAD detects speech
                                    ↓
                    Pre-buffer + new chunks → Recording Buffer
                                    ↓
                            Silence detected
                                    ↓
                    faster-whisper transcription
                                    ↓
                    ydotool/xdotool types text
```

**State Machine** (5 states):
1. `initializing` - Loading Whisper model on startup
2. `ready` - Pre-buffer filling, waiting for VAD to detect speech
3. `recording` - VAD detected speech, actively capturing audio (includes pre-buffer)
4. `processing` - Silence detected, transcribing audio with Whisper
5. `paused` - User paused listening (SIGUSR1), model loaded but not processing

States cycle: ready → recording → processing → ready (or paused from any state)

## Running the Project

**CLI Wrapper** (recommended):
```bash
yap                    # Show help and available commands
yap start              # Start with default model (tiny)
yap start --model small  # Use specific model
yap toggle             # Smart pause/resume/start (single command for everything)
yap pause              # Pause listening (keeps model loaded)
yap resume             # Resume listening
yap stop               # Stop voice typing completely
yap models             # Show installed models on disk
```

**Direct Python**:
```bash
source venv/bin/activate
python main.py --model tiny
```

**Available model sizes**:
- `tiny` (default) - ~75MB, fastest, least accurate
- `base` - ~150MB
- `small` - ~500MB, good balance
- `medium` - ~1.5GB
- `large` - ~3GB, best accuracy, slowest

**Other options**:
- `--device`: cpu (default), cuda
- `--language`: en (default), es, fr, etc.
- `--tcp [PORT]`: Enable TCP server for state monitoring (default port: 12322)

**Output Behavior**:
- Silent startup (only shows `loaded: <model>`)
- If TCP enabled, shows `tcp: <port>` on startup
- Ephemeral status indicators: `ready` → `recording` → `processing` (overwritten in-place)
- When paused, displays `paused` (was stuck on `listening` before)
- Transcribed text prints to terminal FIRST, then types into active window
- ALSA warnings suppressed

## Key Technical Details

**Critical Settings** (in `VoiceTyping.__init__`):
- `RATE = 16000` - Sample rate required by Whisper
- `CHUNK_DURATION_MS = 30` - WebRTC VAD requirement
- `PRE_BUFFER_DURATION_SEC = 1.5` - Captures beginning of speech
- `SILENCE_DURATION_SEC = 0.8` - How long to wait before transcribing
- `vad = webrtcvad.Vad(1)` - Mode 1 = least aggressive (fewer false positives)

**Whisper Transcription** (`process_audio` method):
- `beam_size=5` - Better accuracy vs speed tradeoff
- `vad_filter=True` - Additional VAD filtering in Whisper
- `temperature=0.0` - Deterministic output
- Model warmup on startup with dummy audio prevents first-run delay

**Keyboard Input** (`type_text` method):
- Prints transcribed text to terminal FIRST (immediate feedback)
- Then types into active window using ydotool/xdotool
- Primary: ydotool (Wayland-compatible, requires ydotoold service)
- Fallback: xdotool (X11 only)
- Adds space after text automatically
- Character-by-character typing can be interrupted if you press keys during typing

**Pause/Resume Control** (signal-based IPC):
- `SIGUSR1` → pause listening (stops processing audio, keeps model loaded)
- `SIGUSR2` → resume listening (clears buffers, resumes processing)
- CLI wrapper (`yap pause`/`resume`) sends signals to running process
- Paused state: audio queue still fills but processing skipped (minimal CPU)
- State tracked in `/tmp/yap-state` for toggle functionality
- Now correctly displays `paused` state in terminal and TCP server

**TCP Server for State Monitoring** (optional):
- Inspired by Kanata's TCP port implementation (port 12321)
- Enable with `--tcp` flag (default port 12322) or `--tcp 9999` for custom port
- Poll-based design: clients connect → get JSON state → disconnect
- JSON response format:
  ```json
  {
    "state": "ready",
    "model": "tiny",
    "device": "cpu",
    "language": "en",
    "timestamp": 1761481697
  }
  ```
- Use case: Integration with status bars, border color systems, desktop notifications
- Test connection: `nc 127.0.0.1 12322`
- Thread-safe state tracking with automatic updates
- Server runs on localhost (127.0.0.1) only for security

## Wayland Requirements

**Critical**: ydotoold service must be running before voice typing works on Wayland
- Check: `systemctl status ydotoold` or `ps aux | grep ydotoold`
- Start manually: `ydotoold &`
- NixOS: Import `ydotool-service.nix` into system configuration

## Dependencies

**Core** (requirements.txt):
- `faster-whisper==1.1.1` - Optimized Whisper implementation
- `numpy>=1.24.0` - Audio array processing
- `pyaudio==0.2.14` - Microphone access (requires portaudio system lib)
- `webrtcvad==2.0.10` - Voice activity detection

**System dependencies**:
- portaudio19-dev (for PyAudio)
- ydotool + ydotoold (Wayland) or xdotool (X11)

## Model Management

**How models work**:
- Models are downloaded from HuggingFace (Systran/faster-whisper) on first use
- Cached locally in `~/.cache/huggingface/hub/models--Systran--faster-whisper-<size>/`
- First run with new model will pause briefly for download (30s - few minutes)
- Subsequent runs load instantly from cache
- Use `yap models` to see what's already downloaded

**To use a new model**: Just run `yap start --model medium` - it will auto-download if needed

## Common Issues

**"No words at beginning"**: Should never happen due to pre-buffer. If it does, increase `PRE_BUFFER_DURATION_SEC`

**"ydotool: failed to connect socket"**: ydotoold service not running. Start with `ydotoold &` or via systemd

**"already running"**: Another instance is active. Use `yap stop` to kill it, or `yap toggle` to pause/resume

**High CPU usage**: Use smaller model (`--model tiny`) or enable GPU (`--device cuda` requires PyTorch with CUDA)

**Poor accuracy**:
- Try larger model (`--model base` or `medium`)
- Check microphone quality
- Reduce background noise (VAD mode 1 is least aggressive)

**Text types into wrong window**: Ensure focus is in text input field before speaking. Character-by-character typing can be interrupted if you switch focus or press keys during transcription.

## CLI Wrapper Implementation

**Location**: `yap` bash script in project root
- Auto-detects script directory (no hardcoded paths)
- Uses project venv Python interpreter
- Process discovery via `pgrep -f "python.*main.py"`
- Signal-based control (SIGUSR1/SIGUSR2 for pause/resume, SIGTERM for stop)
- Can be copied to `~/Workspace/tools/bin/` for global PATH access

**Features**:
- **Multi-instance protection**: `yap start` checks if already running, prevents duplicate instances
- **Smart toggle**: Single command to start/pause/resume based on current state
- **Flag passthrough**: `yap toggle --tcp 9999` passes flags to start command if not running
- **State tracking**: Uses `/tmp/yap-state` file to remember pause/active state
- **Model discovery**: `yap models` lists installed models from `~/.cache/huggingface/hub/`
- **Help system**: Running `yap` with no args shows usage and available models
- **Auto-download**: Models automatically download from HuggingFace on first use (via faster-whisper library)
