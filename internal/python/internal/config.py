"""
Configuration constants for voice typing engine.

All tunable parameters are centralized here for easy adjustment.
"""


class AudioConfig:
    """Audio capture and processing parameters."""

    RATE = 16000
    CHUNK_DURATION_MS = 30
    CHUNK_SIZE = int(RATE * CHUNK_DURATION_MS / 1000)
    PRE_BUFFER_DURATION_SEC = 1.5
    BUFFER_DURATION_SEC = 4.0
    SILENCE_DURATION_SEC = 0.8


class VADConfig:
    """Voice Activity Detection parameters."""

    AGGRESSIVENESS = 2


class TranscriptionConfig:
    """Whisper transcription parameters."""

    MIN_AUDIO_DURATION_SEC = 0.7
    BEAM_SIZE = 5
    BEST_OF = 3
    TEMPERATURE = 0.0
    VAD_MIN_SILENCE_MS = 400
    VAD_SPEECH_PAD_MS = 400
    COMPUTE_TYPE_CPU = "int8"
    COMPUTE_TYPE_GPU = "float16"


class DisplayConfig:
    """UI and display parameters."""

    STATUS_LINE_WIDTH = 20


class TCPConfig:
    """TCP server parameters."""

    LISTEN_BACKLOG = 5
    TIMEOUT_SEC = 1.0


class ThreadConfig:
    """Thread synchronization parameters."""

    AUDIO_QUEUE_TIMEOUT_SEC = 0.1
    CLEANUP_SLEEP_SEC = 0.1
