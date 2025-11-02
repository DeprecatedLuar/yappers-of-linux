"""
Audio capture, buffering, and voice activity detection.

Handles:
- Microphone audio stream setup
- Continuous audio reading into queue
- Pre-buffer (circular buffer to capture speech before VAD triggers)
- Recording buffer (main buffer during active recording)
- Voice Activity Detection (VAD) using WebRTC
"""

import os
import queue
import threading
import collections
import pyaudio
import webrtcvad

from .config import AudioConfig, VADConfig, ThreadConfig


class AudioCapture:
    """Manages audio stream and buffering."""

    def __init__(self):
        self.audio_queue = queue.Queue()
        self._running = False
        self._reader_thread = None

        self.vad = webrtcvad.Vad(VADConfig.AGGRESSIVENESS)

        self.pre_buffer = collections.deque(
            maxlen=int(AudioConfig.PRE_BUFFER_DURATION_SEC * AudioConfig.RATE / AudioConfig.CHUNK_SIZE)
        )
        self.recording_buffer = []
        self.is_recording = False
        self.silence_chunks = 0

        # Suppress ALSA warnings during PyAudio initialization
        devnull = os.open(os.devnull, os.O_WRONLY)
        stderr_fd = os.dup(2)
        os.dup2(devnull, 2)
        try:
            self.audio = pyaudio.PyAudio()
            self.stream = self.audio.open(
                format=pyaudio.paInt16,
                channels=1,
                rate=AudioConfig.RATE,
                input=True,
                frames_per_buffer=AudioConfig.CHUNK_SIZE
            )
        finally:
            os.dup2(stderr_fd, 2)
            os.close(devnull)
            os.close(stderr_fd)

    def start(self):
        """Start audio capture thread."""
        self._running = True
        self._reader_thread = threading.Thread(target=self._audio_reader_thread)
        self._reader_thread.daemon = True
        self._reader_thread.start()

    def pause_capture(self):
        """Pause capture (stop thread, close stream, keep PyAudio alive)."""
        self._running = False
        if self._reader_thread:
            self._reader_thread.join(timeout=1.0)
            self._reader_thread = None
        if self.stream:
            self.stream.stop_stream()
            self.stream.close()
            self.stream = None

    def resume_capture(self):
        """Resume capture (reopen stream if needed, restart thread)."""
        # Reopen stream if it was closed
        if not self.stream:
            self.stream = self.audio.open(
                format=pyaudio.paInt16,
                channels=1,
                rate=AudioConfig.RATE,
                input=True,
                frames_per_buffer=AudioConfig.CHUNK_SIZE
            )
        # Start capture thread
        self.start()

    def stop(self):
        """Stop audio capture and cleanup (full teardown)."""
        self._running = False
        if self._reader_thread:
            self._reader_thread.join(timeout=1.0)
        if self.stream:
            self.stream.stop_stream()
            self.stream.close()
        self.audio.terminate()

    def _audio_reader_thread(self):
        """Continuously read audio from microphone into queue."""
        while self._running:
            try:
                chunk = self.stream.read(AudioConfig.CHUNK_SIZE, exception_on_overflow=False)
                self.audio_queue.put(chunk)
            except (OSError, IOError):
                pass

    def get_chunk(self, timeout=None):
        """
        Get next audio chunk from queue.

        Args:
            timeout: Queue timeout in seconds (None = block forever)

        Returns:
            Audio chunk bytes, or None if timeout/empty

        Raises:
            queue.Empty: If timeout expires
        """
        if timeout is None:
            timeout = ThreadConfig.AUDIO_QUEUE_TIMEOUT_SEC
        return self.audio_queue.get(timeout=timeout)

    def is_speech(self, chunk):
        """Check if audio chunk contains speech using VAD."""
        return self.vad.is_speech(chunk, AudioConfig.RATE)

    def add_to_pre_buffer(self, chunk):
        """Add chunk to circular pre-buffer."""
        self.pre_buffer.append(chunk)

    def start_recording(self):
        """Start recording (copies pre-buffer to recording buffer)."""
        self.is_recording = True
        self.recording_buffer = list(self.pre_buffer)
        self.silence_chunks = 0

    def add_to_recording(self, chunk):
        """Add chunk to recording buffer."""
        if self.is_recording:
            self.recording_buffer.append(chunk)

    def increment_silence(self):
        """Increment silence counter."""
        self.silence_chunks += 1

    def reset_silence(self):
        """Reset silence counter."""
        self.silence_chunks = 0

    def get_silence_duration(self):
        """Get current silence duration in seconds."""
        return self.silence_chunks * AudioConfig.CHUNK_DURATION_MS / 1000

    def should_stop_recording(self):
        """Check if silence duration exceeds threshold."""
        return self.get_silence_duration() >= AudioConfig.SILENCE_DURATION_SEC

    def get_recording(self):
        """Get current recording buffer."""
        return self.recording_buffer

    def reset_buffers(self):
        """Clear all buffers and reset recording state."""
        self.pre_buffer.clear()
        self.recording_buffer = []
        self.is_recording = False
        self.silence_chunks = 0

    def clear_queue(self):
        """Empty the audio queue."""
        while not self.audio_queue.empty():
            try:
                self.audio_queue.get_nowait()
            except queue.Empty:
                break
