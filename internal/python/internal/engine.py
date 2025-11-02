"""
Main voice typing engine orchestrator.

Coordinates all components:
- Audio capture and VAD
- Transcription
- Text output
- TCP server (optional)
- State machine (ready → recording → processing → ready)
- Signal handlers (pause/resume)
"""

import signal
import threading
import queue
import time

from .capture import AudioCapture
from .transcribe import Transcriber
from .output import TextOutput
from .server import StateServer


class VoiceTyping:
    """Main voice typing engine."""

    def __init__(self, model_size="small", device="cpu", language="en", tcp_port=None, fast=False, enable_typing=True, output_file=False):
        """
        Initialize voice typing engine.

        Args:
            model_size: Whisper model size (tiny, base, small, medium, large)
            device: Compute device (cpu, cuda)
            language: Language code (en, es, fr, etc.)
            tcp_port: Optional TCP port for state monitoring
            fast: Use fast mode (int8) instead of accurate mode (float32) on CPU
            enable_typing: Enable keyboard typing (default: True, set False to only print to terminal)
            output_file: Write transcriptions to output.txt (default: False)
        """
        self.model_size = model_size
        self.device = device
        self.language = language
        self.tcp_port = tcp_port
        self.fast = fast
        self.enable_typing = enable_typing
        self.output_file = output_file

        # State management
        self._state = "initializing"
        self._state_lock = threading.Lock()
        self._running = True
        self._running_lock = threading.Lock()
        self._paused = False
        self._paused_lock = threading.Lock()
        self._is_typing = False
        self._is_typing_lock = threading.Lock()

        # Initialize components
        self.capture = AudioCapture()
        self.transcriber = Transcriber(model_size, device, language, fast)
        self.output = TextOutput(enable_typing, output_file)

        mode = "fast" if fast else "accurate"
        print(f"model: {model_size} | device: {device} | language: {language} | mode: {mode}\n")

        # Start TCP server if requested
        self.server = None
        if tcp_port:
            self.server = StateServer(tcp_port, self._get_state_dict)
            self.server.start()

        # Initial state
        self.state = "ready"
        # Signal to Go that system is ready (via stderr to not interfere with stdout display)
        import sys
        print("SYSTEM_READY", file=sys.stderr, flush=True)

    @property
    def state(self):
        """Get current state (thread-safe)."""
        with self._state_lock:
            return self._state

    @state.setter
    def state(self, new_state):
        """Set state and update display (thread-safe)."""
        with self._state_lock:
            self._state = new_state
        # Update terminal display
        if new_state in ["ready", "recording", "processing", "paused", "warming_up"]:
            self.output.print_status(new_state)

    @property
    def running(self):
        """Get running flag (thread-safe)."""
        with self._running_lock:
            return self._running

    @running.setter
    def running(self, value):
        """Set running flag (thread-safe)."""
        with self._running_lock:
            self._running = value

    @property
    def paused(self):
        """Get paused flag (thread-safe)."""
        with self._paused_lock:
            return self._paused

    @paused.setter
    def paused(self, value):
        """Set paused flag (thread-safe)."""
        with self._paused_lock:
            self._paused = value

    @property
    def is_typing(self):
        """Get typing flag (thread-safe)."""
        with self._is_typing_lock:
            return self._is_typing

    @is_typing.setter
    def is_typing(self, value):
        """Set typing flag (thread-safe)."""
        with self._is_typing_lock:
            self._is_typing = value

    def _get_state_dict(self):
        """Get state dictionary for TCP server."""
        return {
            "state": self.state,
            "model": self.model_size,
            "device": self.device,
            "language": self.language,
            "is_typing": self.is_typing
        }

    def pause_listening(self, _signum=None, _frame=None):
        """Pause listening (SIGUSR1 handler)."""
        if not self.paused:
            self.paused = True
            self.capture.pause_capture()
            self.state = "paused"

    def resume_listening(self, _signum=None, _frame=None):
        """Resume listening (SIGUSR2 handler)."""
        if self.paused:
            self.paused = False
            self.capture.reset_buffers()
            self.capture.clear_queue()

            # Restart audio capture (reopens stream if closed)
            self.capture.resume_capture()

            # Wait for pre-buffer to fill
            self.state = "warming_up"
            time.sleep(1.5)

            self.state = "ready"

    def run(self):
        """Main event loop."""
        # Register signal handlers
        signal.signal(signal.SIGUSR1, self.pause_listening)
        signal.signal(signal.SIGUSR2, self.resume_listening)

        # Start audio capture
        self.capture.start()

        try:
            while self.running:
                # Get next audio chunk
                try:
                    chunk = self.capture.get_chunk()
                except queue.Empty:
                    continue

                # Skip processing if paused
                if self.paused:
                    continue

                # Add to pre-buffer
                self.capture.add_to_pre_buffer(chunk)

                # Check for speech
                is_speech = self.capture.is_speech(chunk)

                if is_speech and not self.capture.is_recording:
                    # Speech detected - start recording
                    self.state = "recording"
                    self.capture.start_recording()

                elif self.capture.is_recording:
                    # Currently recording
                    self.capture.add_to_recording(chunk)

                    if not is_speech:
                        # Silence detected
                        self.capture.increment_silence()

                        if self.capture.should_stop_recording():
                            # Silence threshold exceeded - transcribe
                            self.state = "processing"
                            text = self.transcriber.transcribe(self.capture.get_recording())

                            if text:
                                self.is_typing = True
                                self.output.type_text(text)
                                self.is_typing = False
                            else:
                                self.output.clear_status_line()

                            self.state = "ready"
                            self.capture.reset_buffers()
                    else:
                        # Speech continues - reset silence counter
                        self.capture.reset_silence()

        except KeyboardInterrupt:
            pass
        finally:
            self.cleanup()

    def cleanup(self):
        """Clean up resources."""
        self.output.clear_status_line()
        self.running = False
        time.sleep(0.1)  # Let threads finish
        self.capture.stop()
        if self.server:
            self.server.stop()
