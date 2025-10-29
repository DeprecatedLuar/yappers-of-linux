"""
Text output to terminal and active window.

Handles:
- Printing transcribed text to terminal
- Typing text into active window via ydotool (Wayland) or xdotool (X11)
- Clearing ephemeral status lines
"""

import subprocess
from .config import DisplayConfig


class TextOutput:
    """Manages text output to terminal and active window."""

    def __init__(self, enable_typing=True):
        """
        Initialize text output.

        Args:
            enable_typing: Enable keyboard typing (default: True, set False to only print to terminal)
        """
        self.enable_typing = enable_typing

    def clear_status_line(self):
        """Clear ephemeral status line in terminal."""
        print(f"\r{' ' * DisplayConfig.STATUS_LINE_WIDTH}\r", end='', flush=True)

    def print_status(self, status):
        """
        Print ephemeral status (overwrites previous line).

        Args:
            status: Status text to display (e.g., "ready", "recording", "processing")
        """
        print(f"\r{status:<{DisplayConfig.STATUS_LINE_WIDTH}}", end='', flush=True)

    def print_text(self, text):
        """
        Print transcribed text to terminal.

        Args:
            text: Text to print
        """
        self.clear_status_line()
        print(f"{text}\n", flush=True)

    def type_text(self, text):
        """
        Type text into active window using ydotool or xdotool.

        Args:
            text: Text to type (space will be appended automatically)
        """
        # Print to terminal first for immediate feedback
        self.print_text(text)

        # Skip keyboard typing if disabled
        if not self.enable_typing:
            return

        # Type into active window
        try:
            # Try ydotool first (Wayland-compatible)
            subprocess.run(
                ['ydotool', '--socket-path', '/run/ydotoold/socket', 'type', text + ' '],
                capture_output=True,
                text=True,
                check=True
            )
        except (subprocess.CalledProcessError, FileNotFoundError):
            # Fallback to xdotool (X11 only)
            try:
                subprocess.run(
                    ['xdotool', 'type', '--delay', '10', text + ' '],
                    capture_output=True,
                    check=True
                )
            except (subprocess.CalledProcessError, FileNotFoundError):
                print(f"\rerror: failed to type")
