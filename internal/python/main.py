#!/usr/bin/env python3
"""
Voice typing for Linux using Whisper.

Entry point for the voice typing engine.
"""

import argparse
import signal
import sys

from internal import VoiceTyping


def main():
    parser = argparse.ArgumentParser(description='Voice typing with pre-buffer')
    parser.add_argument(
        '--model',
        default='small',
        choices=['tiny', 'base', 'small', 'medium', 'large'],
        help='Whisper model size (default: small)'
    )
    parser.add_argument(
        '--device',
        default='cpu',
        choices=['cpu', 'gpu', 'cuda'],
        help='Compute device: cpu, gpu, or cuda (gpu and cuda are aliases, default: cpu)'
    )
    parser.add_argument(
        '--language',
        default='en',
        help='Language code (default: en)'
    )
    parser.add_argument(
        '--tcp',
        nargs='?',
        const=12322,
        type=int,
        metavar='PORT',
        help='Enable TCP server for state monitoring (default port: 12322)'
    )
    parser.add_argument(
        '--fast',
        action='store_true',
        help='Use fast mode (int8, faster but less accurate) instead of accurate mode (float32)'
    )

    args = parser.parse_args()

    # Normalize cuda -> gpu (they're aliases)
    if args.device == 'cuda':
        args.device = 'gpu'

    # Validate GPU availability
    if args.device == 'gpu':
        try:
            import torch
            if not torch.cuda.is_available():
                print("GPU not available, using CPU")
                args.device = 'cpu'
        except ImportError:
            print("PyTorch not installed, using CPU")
            args.device = 'cpu'

    # Create and run engine
    vt = VoiceTyping(
        model_size=args.model,
        device=args.device,
        language=args.language,
        tcp_port=args.tcp,
        fast=args.fast
    )

    # Handle Ctrl+C gracefully
    signal.signal(signal.SIGINT, lambda _s, _f: sys.exit(0))

    vt.run()


if __name__ == '__main__':
    main()
