"""
Speech transcription using Whisper model.

Handles:
- Whisper model loading and initialization
- Model warmup to avoid first-run delay
- Audio transcription with configurable parameters
"""

import numpy as np
from faster_whisper import WhisperModel

from .config import AudioConfig, TranscriptionConfig


class Transcriber:
    """Whisper-based speech transcription."""

    def __init__(self, model_size="small", device="cpu", language="en", fast=False):
        """
        Initialize Whisper model.

        Args:
            model_size: Model size (tiny, base, small, medium, large)
            device: Compute device (cpu, gpu)
            language: Language code (en, es, fr, etc.)
            fast: Use fast mode (int8) instead of accurate mode (float32) on CPU
        """
        self.model_size = model_size
        self.device = device
        self.language = language
        self.fast = fast

        # Map 'gpu' to 'cuda' for faster-whisper backend
        whisper_device = "cuda" if device == "gpu" else "cpu"

        if device == "cpu":
            compute_type = TranscriptionConfig.COMPUTE_TYPE_CPU if fast else "float32"
        else:
            compute_type = TranscriptionConfig.COMPUTE_TYPE_GPU

        self.model = WhisperModel(model_size, device=whisper_device, compute_type=compute_type)

        # Warm up model to avoid first-run delay
        self._warmup()

    def _warmup(self):
        """Run dummy transcription to initialize model."""
        dummy_audio = np.zeros(AudioConfig.RATE, dtype=np.float32)
        list(self.model.transcribe(dummy_audio, language=self.language))

    def transcribe(self, audio_data):
        """
        Transcribe audio to text.

        Args:
            audio_data: List of audio chunks (bytes)

        Returns:
            Transcribed text string (empty if no speech detected)
        """
        # Convert bytes to numpy array
        audio_np = np.frombuffer(b''.join(audio_data), dtype=np.int16).astype(np.float32) / 32768.0

        # Normalize audio to consistent amplitude (peak normalization)
        max_val = np.abs(audio_np).max()
        if max_val > 0:
            audio_np = audio_np / max_val

        # Skip if audio too short
        if len(audio_np) < TranscriptionConfig.MIN_AUDIO_DURATION_SEC * AudioConfig.RATE:
            return ""

        # Transcribe
        segments, _ = self.model.transcribe(
            audio_np,
            language=self.language,
            beam_size=TranscriptionConfig.BEAM_SIZE,
            best_of=TranscriptionConfig.BEST_OF,
            temperature=TranscriptionConfig.TEMPERATURE,
            without_timestamps=True,
            vad_filter=True,
            vad_parameters=dict(
                min_silence_duration_ms=TranscriptionConfig.VAD_MIN_SILENCE_MS,
                speech_pad_ms=TranscriptionConfig.VAD_SPEECH_PAD_MS
            ),
            condition_on_previous_text=False
        )

        # Join segments (filter by confidence to reduce hallucinations)
        full_text = " ".join(
            segment.text.strip()
            for segment in segments
            if segment.text.strip() and segment.avg_logprob > TranscriptionConfig.MIN_CONFIDENCE
        )

        return full_text
