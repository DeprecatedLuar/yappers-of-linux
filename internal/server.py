"""
TCP server for state monitoring.

Provides a simple TCP server that returns current state as JSON.
Used for integration with status bars, border color systems, etc.

Inspired by Kanata's TCP port implementation.
"""

import socket
import threading
import json
import time

from .config import TCPConfig


class StateServer:
    """TCP server for exposing application state."""

    def __init__(self, port, get_state_callback):
        """
        Initialize TCP server.

        Args:
            port: TCP port to listen on
            get_state_callback: Callable that returns state dict
        """
        self.port = port
        self.get_state_callback = get_state_callback
        self._running = False
        self._server_thread = None
        self._server_socket = None

    def start(self):
        """Start TCP server in background thread."""
        self._running = True
        self._server_thread = threading.Thread(target=self._server_loop)
        self._server_thread.daemon = True
        self._server_thread.start()
        print(f"tcp: {self.port}")

    def stop(self):
        """Stop TCP server."""
        self._running = False
        if self._server_socket:
            try:
                self._server_socket.close()
            except OSError:
                pass
        if self._server_thread:
            self._server_thread.join(timeout=2.0)

    def _server_loop(self):
        """Main TCP server loop (runs in background thread)."""
        try:
            self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self._server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            self._server_socket.bind(('127.0.0.1', self.port))
            self._server_socket.listen(TCPConfig.LISTEN_BACKLOG)
            self._server_socket.settimeout(TCPConfig.TIMEOUT_SEC)

            while self._running:
                try:
                    client, _ = self._server_socket.accept()
                    state_json = self._get_state_json()
                    response = state_json + "\n"
                    client.send(response.encode('utf-8'))
                    client.close()
                except socket.timeout:
                    continue
                except (OSError, ConnectionError):
                    break
        except OSError as e:
            print(f"tcp error: {e}")
        finally:
            if self._server_socket:
                try:
                    self._server_socket.close()
                except OSError:
                    pass

    def _get_state_json(self):
        """Get current state as JSON string."""
        state_dict = self.get_state_callback()
        state_dict['timestamp'] = int(time.time())
        return json.dumps(state_dict)
