#!/usr/bin/env python3
"""
Enhanced voice typing with pre-recording buffer
Combines faster-whisper with pre-buffer technique from RealtimeSTT
"""

import argparse
import numpy as np
import pyaudio
import webrtcvad
import collections
import subprocess
import sys
import signal
import time
import threading
import queue
import os
import socket
import json
from faster_whisper import WhisperModel

class VoiceTyping:
    def __init__(self, model_size="small", device="cpu", language="en", tcp_port=None):
        # Store config for state reporting
        self.model_size = model_size
        self.device = device
        self.language = language
        self.tcp_port = tcp_port

        # State tracking with lock for thread safety
        self._state = "initializing"
        self._state_lock = threading.Lock()

        # Initialize running flag early (needed by TCP server thread)
        self.running = True

        # Audio settings
        self.RATE = 16000
        self.CHUNK_DURATION_MS = 30  # ms
        self.CHUNK_SIZE = int(self.RATE * self.CHUNK_DURATION_MS / 1000)
        self.PRE_BUFFER_DURATION_SEC = 1.5  # Pre-recording buffer
        self.BUFFER_DURATION_SEC = 4.0      # Main buffer (longer for context)
        self.SILENCE_DURATION_SEC = 0.8     # Silence to trigger processing

        # VAD settings - less aggressive
        self.vad = webrtcvad.Vad(1)  # Least aggressive mode

        # Initialize Whisper
        compute_type = "int8" if device == "cpu" else "float16"
        self.model = WhisperModel(model_size, device=device, compute_type=compute_type)

        # Warm up the model
        dummy_audio = np.zeros(16000, dtype=np.float32)
        list(self.model.transcribe(dummy_audio, language=language))

        print(f"loaded: {model_size}")

        # Start TCP server if requested
        if tcp_port:
            self.start_tcp_server(tcp_port)
        
        # Audio setup (suppress ALSA warnings)
        devnull = os.open(os.devnull, os.O_WRONLY)
        stderr_fd = os.dup(2)
        os.dup2(devnull, 2)
        try:
            self.audio = pyaudio.PyAudio()
            self.stream = self.audio.open(
                format=pyaudio.paInt16,
                channels=1,
                rate=self.RATE,
                input=True,
                frames_per_buffer=self.CHUNK_SIZE
            )
        finally:
            os.dup2(stderr_fd, 2)
            os.close(devnull)
            os.close(stderr_fd)
        
        # Pre-recording circular buffer (always recording)
        self.pre_buffer = collections.deque(
            maxlen=int(self.PRE_BUFFER_DURATION_SEC * self.RATE / self.CHUNK_SIZE)
        )
        
        # Main recording buffer
        self.recording_buffer = []
        
        # State
        self.is_recording = False
        self.silence_chunks = 0
        self.speech_detected = False
        self.paused = False  # Pause state for external control

        # Threading for continuous pre-buffering
        self.audio_queue = queue.Queue()

        # Set initial state to ready after initialization
        self.state = "ready"

    @property
    def state(self):
        """Get current state"""
        with self._state_lock:
            return self._state

    @state.setter
    def state(self, new_state):
        """Set state and update display"""
        with self._state_lock:
            self._state = new_state
        # Update terminal display (ephemeral) with padding to clear previous content
        if new_state in ["ready", "recording", "processing", "paused"]:
            print(f"\r{new_state:<20}", end='', flush=True)

    def get_state_json(self):
        """Get current state as JSON for TCP clients"""
        return json.dumps({
            "state": self.state,
            "model": self.model_size,
            "device": self.device,
            "language": self.language,
            "timestamp": int(time.time())
        })

    def start_tcp_server(self, port):
        """Start TCP server for state monitoring"""
        def tcp_server_thread():
            try:
                server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
                server.bind(('127.0.0.1', port))
                server.listen(5)
                server.settimeout(1.0)  # Non-blocking with timeout
                print(f"tcp: {port}")

                while self.running:
                    try:
                        client, _ = server.accept()
                        response = self.get_state_json() + "\n"
                        client.send(response.encode('utf-8'))
                        client.close()
                    except socket.timeout:
                        continue
                    except:
                        break
            except Exception as e:
                print(f"tcp error: {e}")
            finally:
                try:
                    server.close()
                except:
                    pass

        thread = threading.Thread(target=tcp_server_thread, daemon=True)
        thread.start()

    def pause_listening(self, signum=None, frame=None):
        """Pause audio processing"""
        if not self.paused:
            self.paused = True
            self.state = "paused"

    def resume_listening(self, signum=None, frame=None):
        """Resume audio processing"""
        if self.paused:
            self.paused = False
            self.pre_buffer.clear()
            self.recording_buffer = []
            self.is_recording = False
            self.silence_chunks = 0
            while not self.audio_queue.empty():
                try:
                    self.audio_queue.get_nowait()
                except queue.Empty:
                    break
            self.state = "ready"

    def audio_reader_thread(self):
        """Continuously read audio in a separate thread"""
        while self.running:
            try:
                chunk = self.stream.read(self.CHUNK_SIZE, exception_on_overflow=False)
                self.audio_queue.put(chunk)
            except:
                pass
                
    def process_audio(self, audio_data):
        """Transcribe audio and type it out"""
        # Convert to numpy array
        audio_np = np.frombuffer(b''.join(audio_data), dtype=np.int16).astype(np.float32) / 32768.0
        
        # Skip if too short
        if len(audio_np) < 0.5 * self.RATE:  # Less than 0.5 seconds
            return
        
        # Transcribe with optimized settings
        segments, _ = self.model.transcribe(
            audio_np,
            language=self.language,
            beam_size=5,  # Better accuracy
            best_of=3,    # Multiple attempts
            temperature=0.0,
            without_timestamps=True,
            vad_filter=True,
            vad_parameters=dict(
                min_silence_duration_ms=300,
                speech_pad_ms=400  # Padding around speech
            )
        )
        
        # Collect all text
        full_text = " ".join(segment.text.strip() for segment in segments if segment.text.strip())

        if full_text:
            # Type it out
            self.type_text(full_text)
        else:
            # Clear the "processing" message
            print("\r" + " " * 20 + "\r", end='', flush=True)
            
    def type_text(self, text):
        """Type the text using ydotool or xdotool"""
        # Print first for immediate feedback (clear status line first)
        print(f"\r{' ' * 20}\r{text}\n", flush=True)

        try:
            # Try ydotool first
            subprocess.run(['ydotool', '--socket-path', '/run/ydotoold/socket', 'type', text + ' '],
                         capture_output=True, text=True, check=True)
        except:
            # Fallback to xdotool
            try:
                subprocess.run(['xdotool', 'type', '--delay', '10', text + ' '],
                             capture_output=True, check=True)
            except:
                print(f"\rerror: failed to type")
                
    def run(self):
        """Main recording loop"""
        # Register signal handlers
        signal.signal(signal.SIGUSR1, self.pause_listening)
        signal.signal(signal.SIGUSR2, self.resume_listening)

        # Start audio reader thread
        reader_thread = threading.Thread(target=self.audio_reader_thread)
        reader_thread.daemon = True
        reader_thread.start()

        try:
            while self.running:
                # Get audio chunk from queue
                try:
                    chunk = self.audio_queue.get(timeout=0.1)
                except queue.Empty:
                    continue

                # Skip processing if paused
                if self.paused:
                    continue

                # Always add to pre-buffer (circular buffer)
                self.pre_buffer.append(chunk)

                # Check for speech
                is_speech = self.vad.is_speech(chunk, self.RATE)

                if is_speech and not self.is_recording:
                    # Start recording - include pre-buffer
                    self.state = "recording"
                    self.is_recording = True
                    self.recording_buffer = list(self.pre_buffer)
                    self.recording_buffer.append(chunk)
                    self.silence_chunks = 0

                elif self.is_recording:
                    # Continue recording
                    self.recording_buffer.append(chunk)

                    if not is_speech:
                        self.silence_chunks += 1
                        silence_duration = self.silence_chunks * self.CHUNK_DURATION_MS / 1000

                        if silence_duration >= self.SILENCE_DURATION_SEC:
                            # Process the recording
                            self.state = "processing"
                            self.process_audio(self.recording_buffer)

                            # Reset
                            self.state = "ready"
                            self.is_recording = False
                            self.recording_buffer = []
                            self.silence_chunks = 0
                    else:
                        self.silence_chunks = 0

        except KeyboardInterrupt:
            pass
        finally:
            # Clear status line before exit
            print("\r" + " " * 20 + "\r", end='', flush=True)
            self.cleanup()
            
    def cleanup(self):
        """Clean up resources"""
        self.running = False
        time.sleep(0.1)
        self.stream.stop_stream()
        self.stream.close()
        self.audio.terminate()

def main():
    parser = argparse.ArgumentParser(description='Enhanced voice typing with pre-buffer')
    parser.add_argument('--model', default='small',
                       choices=['tiny', 'base', 'small', 'medium', 'large'],
                       help='Model size (default: small)')
    parser.add_argument('--device', default='cpu',
                       choices=['cpu', 'cuda'],
                       help='Device (default: cpu)')
    parser.add_argument('--language', default='en',
                       help='Language code (default: en)')
    parser.add_argument('--tcp', nargs='?', const=12322, type=int, metavar='PORT',
                       help='Enable TCP server for state monitoring (default port: 12322)')

    args = parser.parse_args()
    
    # Check CUDA if requested
    if args.device == 'cuda':
        try:
            import torch
            if not torch.cuda.is_available():
                print("CUDA not available, using CPU")
                args.device = 'cpu'
        except ImportError:
            print("PyTorch not installed, using CPU")
            args.device = 'cpu'
    
    # Run voice typing
    vt = VoiceTyping(
        model_size=args.model,
        device=args.device,
        language=args.language,
        tcp_port=args.tcp
    )
    
    # Handle graceful shutdown
    signal.signal(signal.SIGINT, lambda s, f: sys.exit(0))
    
    vt.run()

if __name__ == '__main__':
    main()