# Voice Typing for Linux - Technical Reference

Voice typing for Linux (Wayland/X11) using OpenAI Whisper via faster-whisper. Pre-recording circular buffer ensures no speech loss at utterance start.

## Architecture

**Single-file implementation**: `main.py` (267 lines)

**Core class**: `VoiceTyping`
- Threaded audio reader populates queue continuously
- Main loop processes queue, manages state machine
- Pre-buffer: 1.5s circular deque (always recording)
- Main buffer: 4.0s for context
- Silence threshold: 0.8s triggers transcription

**Audio pipeline**:
```
Microphone → PyAudio(16kHz mono) → Queue → Circular pre-buffer(1.5s)
                                              ↓ (VAD detects speech)
                                    Pre-buffer + new chunks → Recording buffer
                                              ↓ (0.8s silence)
                                    faster-whisper transcription
                                              ↓
                                    ydotool/xdotool types text + space
```

**State machine**:
1. Continuous: Fill pre-buffer (deque with maxlen)
2. VAD speech detected → `is_recording=True`, copy pre-buffer to recording_buffer
3. Continue recording while speech detected
4. 0.8s silence → process & transcribe, reset to step 1
5. Pause state (SIGUSR1): queue fills, processing skipped, model stays loaded

## Critical Parameters

**`VoiceTyping.__init__` constants** (main.py:22-32):
```python
RATE = 16000                      # Whisper requirement
CHUNK_DURATION_MS = 30            # WebRTC VAD requirement (30ms frames)
CHUNK_SIZE = 480                  # 16000 * 30 / 1000
PRE_BUFFER_DURATION_SEC = 1.5     # Captures speech beginning
BUFFER_DURATION_SEC = 4.0         # Main recording context window
SILENCE_DURATION_SEC = 0.8        # Silence before transcription trigger
vad = webrtcvad.Vad(1)           # Mode 1 = least aggressive (fewer false positives)
```

**Whisper transcription settings** (main.py:119-131):
```python
beam_size=5                       # Accuracy vs speed tradeoff
best_of=3                         # Multiple decoding attempts
temperature=0.0                   # Deterministic output
vad_filter=True                   # Additional VAD in Whisper
vad_parameters:
  min_silence_duration_ms=300
  speech_pad_ms=400              # Padding around detected speech
```

**Model warmup** (main.py:39-41):
- Dummy 1s audio transcribed on init to prevent first-run delay

## CLI Wrapper

**File**: `voice` bash script
- Default model: `tiny` (not small)
- Auto-detects script directory (portable)
- Uses project venv Python interpreter
- Process discovery: `pgrep -f "python.*main.py"`

**Commands**:
```bash
voice                  # Start with --model tiny (default)
voice --model base     # Use specific model
voice --device cuda    # GPU acceleration (requires torch + CUDA)
voice --language es    # Set language (default: en)
voice pause            # SIGUSR1 → paused=True, processing skipped, model loaded
voice resume           # SIGUSR2 → paused=False, clear buffers, resume
voice stop             # SIGTERM → cleanup and exit
```

**Available models**: tiny, base, small, medium, large (code only accepts these 5)

## Dependencies

**Python** (requirements.txt):
```
faster-whisper==1.1.1    # CTranslate2-optimized Whisper
numpy>=1.24.0            # Audio array processing
pyaudio==0.2.14          # Microphone access
webrtcvad==2.0.10        # Voice activity detection (30ms frames only)
torch>=2.0.0             # Optional: GPU support
```

**System**:
- `portaudio19-dev` - Required by PyAudio
- `ydotool` + `ydotoold` (Wayland) - Keyboard input injection
- `xdotool` (X11) - Keyboard input fallback
- Microphone access via ALSA/PulseAudio/PipeWire

**Wayland requirement**: ydotoold daemon MUST be running before script starts
- Check: `ps aux | grep ydotoold`
- Start: `ydotoold &`
- Socket: `/run/ydotoold/socket` (hardcoded in main.py:150)

## Output Behavior

**Silent startup**: Only prints `loaded: <model>`
- ALSA warnings suppressed (stderr redirected during PyAudio init)
- No startup banner or status messages

**Runtime output**:
- `listening` - Ephemeral (overwritten with `\r`)
- `processing` - Ephemeral (overwritten with `\r`)
- Transcribed text - Persistent, followed by newline
- `error: failed to type` - If both ydotool and xdotool fail

## Signal-Based Control

**Registered handlers** (main.py:163-164):
- `SIGUSR1` → `pause_listening()`: Set paused flag, stop processing
- `SIGUSR2` → `resume_listening()`: Clear paused flag, clear buffers, empty queue
- `SIGINT` → Graceful exit via KeyboardInterrupt

**Paused state behavior**:
- Audio queue continues filling (audio_reader_thread still runs)
- Main loop skips processing when `self.paused=True`
- Model stays loaded in memory (no re-initialization on resume)
- Resume clears all buffers to prevent stale audio

## Key Implementation Details

**Threading model** (main.py:100-107):
- Daemon thread continuously reads audio chunks into queue
- Main thread processes queue in 0.1s timeout loop
- Exception suppression in reader thread prevents crashes

**Keyboard typing** (main.py:143-159):
1. Print to stdout first (immediate feedback)
2. Try ydotool with explicit socket path
3. Fallback to xdotool with 10ms delay between keys
4. Append space to all typed text
5. Silent failure if both methods fail (only error message)

**Audio processing skip logic** (main.py:114-116):
- Discard recordings < 0.5s (too short for useful transcription)
- Convert int16 PCM to float32 normalized [-1, 1]
- Join all buffer chunks before numpy conversion

## Project Files

```
main.py                    # Main implementation (267 lines)
yap                        # Bash wrapper
requirements.txt           # Python dependencies
shell.nix                  # Nix development shell
ydotool-service.nix       # NixOS systemd service for ydotoold
venv/                      # Python virtual environment
```

## Known Issues

**Pre-buffer not working**: Increase `PRE_BUFFER_DURATION_SEC` if initial words still missed (should not happen with current 1.5s)

**ydotool socket error**: ydotoold not running or wrong socket path
- Current path: `/run/ydotoold/socket` (hardcoded)
- Check: `ls -la /run/ydotoold/socket`

**High CPU**:
- Use smaller model (`--model tiny`)
- Enable GPU (`--device cuda` requires torch with CUDA support)
- VAD mode 1 is already least aggressive (modes: 0-3)

**Poor accuracy**:
- Larger model (`--model base`, `medium`, `large`)
- Check mic sample rate: `arecord -V stereo -r 16000 -f S16_LE -d 5 test.wav`
- Reduce background noise (VAD mode 1 may miss soft speech)