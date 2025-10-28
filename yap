#!/usr/bin/env bash
# yap - voice typing for Linux

# Project location (actual code and venv live here)
PROJECT_DIR="$HOME/Workspace/tools/homemade/yap"
VENV_PYTHON="$PROJECT_DIR/venv/bin/python"
SCRIPT="$PROJECT_DIR/main.py"

get_pid() {
    pgrep -f "$VENV_PYTHON.*$SCRIPT" 2>/dev/null
}

notify_started() {
    notify-send -i audio-input-microphone -u critical "Yap" "Yapping started" 2>/dev/null &
}

show_help() {
    cat << 'EOF'
usage: yap <command> [options]

commands:
  start [options]     Start voice typing (default: --model tiny)
  toggle [options]    Smart pause/resume/start
  pause               Pause listening
  resume              Resume listening
  stop (kill)         Stop voice typing
  models              Show installed models

options:
  --model X           Model size: tiny, base, small, medium, large
  --device X          Device: cpu, cuda
  --language X        Language code (default: en)
  --tcp [PORT]        Enable TCP server (default port: 12322)

models automatically download on first use
EOF
}

show_models() {
    CACHE_DIR="$HOME/.cache/huggingface/hub"
    if [ ! -d "$CACHE_DIR" ]; then
        echo "no models installed"
        return
    fi

    INSTALLED=$(ls "$CACHE_DIR" 2>/dev/null | grep "faster-whisper-" | sed 's/.*faster-whisper-//' | tr '\n' ', ' | sed 's/,$//')

    if [ -z "$INSTALLED" ]; then
        echo "no models installed"
    else
        echo "installed: $INSTALLED"
    fi
}

case "$1" in
    help|--help|-h|"")
        show_help
        ;;

    models)
        show_models
        ;;

    start)
        shift  # Remove 'start' from args
        PID=$(get_pid)
        if [ -n "$PID" ]; then
            echo "already running (pid $PID)"
            exit 0
        fi

        notify_started

        # Start with provided args or defaults
        if [ $# -eq 0 ]; then
            exec "$VENV_PYTHON" "$SCRIPT" --model tiny
        else
            exec "$VENV_PYTHON" "$SCRIPT" "$@"
        fi
        ;;

    toggle)
        shift  # Remove 'toggle' from args
        PID=$(get_pid)
        if [ -z "$PID" ]; then
            # Not running - clean up stale state file and start
            rm -f /tmp/yap-state
            notify_started
            if [ $# -eq 0 ]; then
                exec "$VENV_PYTHON" "$SCRIPT" --model tiny
            else
                exec "$VENV_PYTHON" "$SCRIPT" "$@"
            fi
        else
            # Running - check if paused
            STATE_FILE="/tmp/yap-state"

            if [ -f "$STATE_FILE" ] && [ "$(cat "$STATE_FILE")" = "paused" ]; then
                # Currently paused - resume
                notify_started
                kill -SIGUSR2 "$PID"
                echo "active" > "$STATE_FILE"
            else
                # Currently active - pause
                kill -SIGUSR1 "$PID"
                echo "paused" > "$STATE_FILE"
            fi
        fi
        ;;

    pause)
        PID=$(get_pid)
        if [ -z "$PID" ]; then
            echo "not running"
            exit 1
        fi
        kill -SIGUSR1 "$PID"
        echo "paused" > /tmp/yap-state
        ;;

    resume)
        PID=$(get_pid)
        if [ -z "$PID" ]; then
            echo "not running"
            exit 1
        fi
        notify_started
        kill -SIGUSR2 "$PID"
        echo "active" > /tmp/yap-state
        ;;

    stop|kill)
        PID=$(get_pid)
        if [ -z "$PID" ]; then
            echo "not running"
            exit 1
        fi
        kill "$PID"
        rm -f /tmp/yap-state
        ;;

    *)
        echo "unknown command: $1"
        echo "run 'yap help' for usage"
        exit 1
        ;;
esac
