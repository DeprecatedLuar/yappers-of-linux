#!/bin/bash

# Yappers of Linux Local Install Script
# Builds and installs to ~/Workspace/tools/bin/

set -e

INSTALL_DIR="$HOME/Workspace/tools/bin"

echo "Building yap binary..."
go build -o yap cmd/main.go

if [ $? -eq 0 ]; then
    echo "Build successful!"

    mkdir -p "$INSTALL_DIR"

    echo "Installing to $INSTALL_DIR..."
    cp yap "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/yap"

    echo ""
    echo "Installation complete!"
    echo "Binary installed to: $INSTALL_DIR/yap"
    echo ""
    echo "Test with: yap"
    echo "First run will auto-extract Python code and install dependencies"
    echo ""
    echo "Note: Make sure $INSTALL_DIR is in your PATH"
else
    echo "Build failed!"
    exit 1
fi