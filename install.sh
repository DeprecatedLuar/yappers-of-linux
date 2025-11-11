#!/usr/bin/env bash

# ===== CONFIGURATION =====
# Required parameters
PROJECT_NAME="Yappers of Linux"
BINARY_NAME="yap"
REPO_USER="DeprecatedLuar"
REPO_NAME="yappers-of-linux"
INSTALL_DIR="$HOME/.local/bin"
BUILD_CMD="go build -o $BINARY_NAME cmd/main.go"

# Optional branding (leave empty to disable)
ASCII_ART=''
MSG_FINAL="big hug from Luar"
NEXT_STEPS="Try running: $BINARY_NAME start|First run will auto-install Python dependencies (~2 minutes)"

# ===== END CONFIGURATION =====

set -e

# Command routing
case "${1:-}" in
    local)
        # Local installation - no satellite needed
        echo "Installing locally..."

        # Find binary in current directory
        if [ ! -f "$BINARY_NAME" ]; then
            echo "Error: $BINARY_NAME not found. Build it first (e.g., go build)"
            exit 1
        fi

        # Stop running instance
        if pgrep -x "$BINARY_NAME" > /dev/null 2>&1; then
            echo "Stopping running instance..."
            pkill -TERM -x "$BINARY_NAME" 2>/dev/null || true
            sleep 1
        fi

        # Install
        mkdir -p "$INSTALL_DIR"
        cp "$BINARY_NAME" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"

        echo "Installed to $INSTALL_DIR/$BINARY_NAME"
        ;;

    *)
        # Standard installation via satellite
        curl -sSL https://raw.githubusercontent.com/$REPO_USER/the-satellite/main/satellite.sh | \
            bash -s -- install \
                "$PROJECT_NAME" \
                "$BINARY_NAME" \
                "$REPO_USER" \
                "$REPO_NAME" \
                "$INSTALL_DIR" \
                "$BUILD_CMD" \
                "$ASCII_ART" \
                "$MSG_FINAL" \
                "$NEXT_STEPS"
        ;;
esac

