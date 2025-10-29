"""
Text output to terminal and active window.

Handles:
- Printing transcribed text to terminal
- Typing text into active window via ydotool (Wayland) or xdotool (X11)
- Clearing ephemeral status lines
"""

import os
import subprocess
from .config import DisplayConfig


class TextOutput:
    """Manages text output to terminal and active window."""

    def __init__(self, enable_typing=True, output_file=False):
        """
        Initialize text output.

        Args:
            enable_typing: Enable keyboard typing (default: True, set False to only print to terminal)
            output_file: Write transcriptions to output.txt (default: False)
        """
        self.enable_typing = enable_typing
        self.output_file = output_file

        # Get output file path if enabled
        self.output_file_path = None
        if output_file:
            config_dir = os.path.expanduser("~/.config/yappers-of-linux")
            self.output_file_path = os.path.join(config_dir, "output.txt")

        # Detect session type (Wayland vs X11)
        self.session_type = os.environ.get('XDG_SESSION_TYPE', '').lower()
        self.is_wayland = self.session_type == 'wayland'
        self.is_x11 = self.session_type == 'x11'

    def clear_status_line(self):
        """Clear ephemeral status line in terminal."""
        print(f"\r{' ' * DisplayConfig.STATUS_LINE_WIDTH}\r", end='', flush=True)

    def print_status(self, status):
        """
        Print ephemeral status (overwrites previous line).

        Args:
            status: Status text to display (e.g., "ready", "recording", "processing")
        """
        # Clear line first, then print status (prevents cursor glitches)
        print(f"\r{' ' * DisplayConfig.STATUS_LINE_WIDTH}\r{status}", end='', flush=True)

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

        # Write to output file if enabled
        if self.output_file_path:
            try:
                with open(self.output_file_path, 'a', encoding='utf-8') as f:
                    f.write(text + '\n')
            except Exception as e:
                # Silent failure - don't interrupt voice typing for file I/O errors
                pass

        # Skip keyboard typing if disabled
        if not self.enable_typing:
            return

        # Type into active window based on session type
        if self.is_wayland:
            self._type_wayland(text)
        elif self.is_x11:
            self._type_x11(text)
        else:
            # Unknown session type, try both
            self._type_with_fallback(text)

    def _type_wayland(self, text):
        """Type text on Wayland using ydotool."""
        try:
            subprocess.run(
                ['ydotool', 'type', text + ' '],
                capture_output=True,
                text=True,
                check=True
            )
        except FileNotFoundError:
            print(f"\rerror: ydotool not installed (required for Wayland)")
        except subprocess.CalledProcessError:
            print(f"\rerror: ydotool failed (is ydotoold running? try: sudo ydotoold &)")

    def _type_x11(self, text):
        """Type text on X11 using xdotool."""
        try:
            subprocess.run(
                ['xdotool', 'type', '--delay', '10', text + ' '],
                capture_output=True,
                text=True,
                check=True
            )
        except FileNotFoundError:
            print(f"\rerror: xdotool not installed (required for X11)")
        except subprocess.CalledProcessError:
            print(f"\rerror: xdotool failed")

    def _type_with_fallback(self, text):
        """Try ydotool first, fallback to xdotool (for unknown session types)."""
        try:
            subprocess.run(
                ['ydotool', 'type', text + ' '],
                capture_output=True,
                text=True,
                check=True
            )
        except (subprocess.CalledProcessError, FileNotFoundError):
            try:
                subprocess.run(
                    ['xdotool', 'type', '--delay', '10', text + ' '],
                    capture_output=True,
                    check=True
                )
            except (subprocess.CalledProcessError, FileNotFoundError):
                print(f"\rerror: failed to type (session: {self.session_type})")
