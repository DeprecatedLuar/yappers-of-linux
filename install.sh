#!/usr/bin/env bash

# ===== CONFIGURATION =====
PROJECT_NAME="Yappers of Linux"
REPO_USER="DeprecatedLuar"
REPO_NAME="yappers-of-linux"
BRANCH="main"
BINARY_NAME="yap"

# Installation directory
INSTALL_DIR="/usr/local/bin"  # System-wide (requires sudo)
# INSTALL_DIR="$HOME/.local/bin"  # User install (no sudo)

# Build configuration (for fallback)
BUILD_LANG="go"  # Language: go, rust, make
GO_MODULE_PATH="github.com/$REPO_USER/$REPO_NAME"  # For Go projects
BUILD_CMD="go build -o $BINARY_NAME cmd/main.go"  # Custom build command
# BUILD_CMD="cargo build --release && cp target/release/$BINARY_NAME ."  # Rust example
# BUILD_CMD="make"  # Make example

# Custom messages
MSG_DETECTING="Detecting system..."
MSG_NO_RELEASE="No release found, building from source..."
MSG_DOWNLOAD="Downloading binary"
MSG_DOWNLOAD_FAILED="Release download failed, building from source..."
MSG_CLONE="Cloning repository..."
MSG_BUILD="Building from source..."
MSG_INSTALL="Installing to"
MSG_INSTALL_COMPLETE="Installation complete"
MSG_TRY_COMMAND="Try running: $BINARY_NAME start"
MSG_NO_GO="Go not found. Install Go and try again."
MSG_NO_RUST="Rust/Cargo not found. Install Rust and try again."
MSG_NO_GIT="Git not found. Install Git and try again."
MSG_BUILD_FAILED="Build failed"
NEXT_STEPS=("Try running: $BINARY_NAME start" "First run will auto-install Python dependencies (~2 minutes)") # Pre-baked next steps

# ASCII art (leave empty to disable)
ASCII_ART=''

# ===== END CONFIGURATION =====

set -e

# ===========================================
# MESSAGES LIBRARY
# ===========================================

# Color detection: disable colors if not outputting to a terminal
if [ -t 1 ] && [ -n "$TERM" ]; then
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
else
    BLUE=''
    CYAN=''
    GREEN=''
    RED=''
    YELLOW=''
    NC=''
fi

action() {
    echo -e "${BLUE}→${NC} $1"
}

info() {
    echo -e "${CYAN}ℹ${NC} $1"
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

error() {
    echo -e "${RED}✗ $1${NC}" >&2
    exit 1
}

warn() {
    echo -e "${YELLOW}! $1${NC}"
}

separator() {
    local text="$1"
    local width=$(tput cols 2>/dev/null || echo 80)
    local text_length=$((${#text} + 4))
    local dash_count=$((width - text_length))

    if [ $dash_count -lt 0 ]; then
        dash_count=0
    fi

    local dashes=$(printf '%*s' "$dash_count" '' | tr ' ' '-')
    echo -e "${BLUE}-- ${text} ${dashes}${NC}"
}

# ===========================================
# OS DETECTION LIBRARY
# ===========================================

detect_os() {
    local os_type
    os_type=$(uname -s | tr '[:upper:]' '[:lower:]')

    case "$os_type" in
        linux*)
            echo "linux"
            ;;
        darwin*)
            echo "darwin"
            ;;
        mingw* | msys* | cygwin*)
            echo "windows"
            ;;
        freebsd*)
            echo "freebsd"
            ;;
        openbsd*)
            echo "openbsd"
            ;;
        netbsd*)
            echo "netbsd"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

detect_arch() {
    local arch
    arch=$(uname -m)

    case "$arch" in
        x86_64 | x86-64 | x64 | amd64)
            echo "amd64"
            ;;
        aarch64 | arm64)
            echo "arm64"
            ;;
        armv7* | armv8l)
            echo "arm"
            ;;
        armv6*)
            echo "armv6"
            ;;
        i386 | i686)
            echo "386"
            ;;
        *)
            echo "$arch"
            ;;
    esac
}

is_nixos() {
    [[ -f /etc/NIXOS ]]
}

parse_os_release() {
    local key="$1"
    local value=""

    if [[ -f /etc/os-release ]]; then
        value=$(grep -E "^${key}=" /etc/os-release | cut -d= -f2- | tr -d '"')
    fi

    echo "$value"
}

detect_distro() {
    local os="$1"

    if [[ "$os" != "linux" ]]; then
        echo "none"
        return
    fi

    if is_nixos; then
        echo "nixos"
        return
    fi

    local distro_id
    distro_id=$(parse_os_release "ID")

    if [[ -n "$distro_id" ]]; then
        echo "$distro_id"
    else
        echo "unknown"
    fi
}

detect_distro_family() {
    local distro="$1"

    case "$distro" in
        nixos)
            echo "nixos"
            ;;
        ubuntu | debian | pop | linuxmint | raspbian)
            echo "debian"
            ;;
        arch | manjaro | endeavouros)
            echo "arch"
            ;;
        fedora | rhel | centos | rocky | alma)
            echo "rhel"
            ;;
        alpine)
            echo "alpine"
            ;;
        gentoo)
            echo "gentoo"
            ;;
        opensuse* | sles)
            echo "suse"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

detect_distro_version() {
    local os="$1"

    if [[ "$os" != "linux" ]]; then
        echo ""
        return
    fi

    parse_os_release "VERSION_ID"
}

detect_kernel() {
    uname -r
}

get_system_info() {
    local os arch distro distro_family distro_version kernel

    os=$(detect_os)
    arch=$(detect_arch)
    distro=$(detect_distro "$os")
    distro_family=$(detect_distro_family "$distro")
    distro_version=$(detect_distro_version "$os")
    kernel=$(detect_kernel)

    cat <<EOF
{
  "os": "$os",
  "arch": "$arch",
  "distro": "$distro",
  "distro_family": "$distro_family",
  "distro_version": "$distro_version",
  "kernel": "$kernel"
}
EOF
}

# ===========================================
# PATH MANAGEMENT LIBRARY
# ===========================================

ensure_in_path() {
    local install_dir="$1"

    # Check if already in PATH
    if [[ ":$PATH:" == *":$install_dir:"* ]]; then
        return 0
    fi

    echo ""
    warn "$install_dir is not in your PATH"
    echo ""

    # Detect distro for NixOS handling
    local distro=""
    if is_nixos; then
        distro="nixos"
    fi

    # NixOS special handling
    if [[ "$distro" == "nixos" ]]; then
        handle_nixos_path "$install_dir"
        return 0
    fi

    # Detect shell
    local user_shell=$(basename "$SHELL")
    local rc_file=""

    case "$user_shell" in
        bash)
            rc_file="$HOME/.bashrc"
            ;;
        zsh)
            rc_file="$HOME/.zshrc"
            ;;
        fish)
            rc_file="$HOME/.config/fish/config.fish"
            ;;
        *)
            warn "Unknown shell: $user_shell"
            info "Add this to your shell config manually:"
            echo "  export PATH=\"$install_dir:\$PATH\""
            return 1
            ;;
    esac

    # Create rc file if it doesn't exist
    if [[ ! -f "$rc_file" ]]; then
        touch "$rc_file"
    fi

    # Check if PATH export already exists
    if grep -q "$install_dir" "$rc_file" 2>/dev/null; then
        info "PATH export already in $rc_file"
        info "Reload your shell: source $rc_file"
        return 0
    fi

    # Add to rc file
    echo "" >> "$rc_file"
    echo "# Added by installer" >> "$rc_file"
    echo "export PATH=\"$install_dir:\$PATH\"" >> "$rc_file"

    success "Added $install_dir to PATH in $rc_file"
    info "Reload your shell: source $rc_file"
}

handle_nixos_path() {
    local install_dir="$1"

    echo "NixOS detected. Choose installation method:"
    echo ""
    echo "  1) Quick way - Add to .bashrc (works immediately)"
    echo "  2) NixOS way - Use declarative configuration (proper NixOS style)"
    echo ""
    read -p "Choice [1/2]: " choice

    case "$choice" in
        1)
            # Add to .bashrc even on NixOS
            local rc_file="$HOME/.bashrc"
            if [[ ! -f "$rc_file" ]]; then
                touch "$rc_file"
            fi

            if grep -q "$install_dir" "$rc_file" 2>/dev/null; then
                info "PATH export already in $rc_file"
            else
                echo "" >> "$rc_file"
                echo "# Added by installer" >> "$rc_file"
                echo "export PATH=\"$install_dir:\$PATH\"" >> "$rc_file"
                success "Added $install_dir to PATH in $rc_file"
            fi
            info "Reload your shell: source $rc_file"
            ;;
        2)
            # Show declarative instructions
            echo ""
            info "For declarative configuration, add to your home-manager config:"
            echo ""
            echo "  home.sessionPath = [ \"$install_dir\" ];"
            echo ""
            info "Or in configuration.nix (system-wide):"
            echo ""
            echo "  environment.sessionVariables = {"
            echo "    PATH = [ \"$install_dir\" ];"
            echo "  };"
            echo ""
            info "Then run: nixos-rebuild switch"
            echo ""
            ;;
        *)
            error "Invalid choice. Exiting."
            ;;
    esac
}

# ===========================================
# MAIN INSTALLATION LOGIC
# ===========================================

# Detect OS and architecture
action "$MSG_DETECTING"
SYSTEM_INFO=$(get_system_info)
OS=$(echo "$SYSTEM_INFO" | grep -o '"os": "[^"]*"' | cut -d'"' -f4)
ARCH=$(echo "$SYSTEM_INFO" | grep -o '"arch": "[^"]*"' | cut -d'"' -f4)
echo "Detected: $OS $ARCH"

# Check for dependencies function
require_command() {
    local cmd="$1"
    local msg="$2"

    if ! command -v "$cmd" &> /dev/null; then
        error "$msg"
    fi
}

# Try GitHub releases first
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO_USER/$REPO_NAME/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' || echo "")

# If no stable release, try first release (including prereleases)
if [ -z "$LATEST_RELEASE" ]; then
    LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO_USER/$REPO_NAME/releases" | grep '"tag_name":' | head -1 | sed -E 's/.*"([^"]+)".*/\1/' || echo "")
fi

BINARY_INSTALLED=false

if [ -n "$LATEST_RELEASE" ]; then
    # Try common binary naming patterns
    BINARY_PATTERNS=(
        "${BINARY_NAME}-${OS}-${ARCH}"
        "${BINARY_NAME}_${OS}_${ARCH}"
        "${BINARY_NAME}-${LATEST_RELEASE}-${OS}-${ARCH}"
        "${BINARY_NAME}"
    )

    # Also try with .tar.gz and .zip extensions
    ARCHIVE_PATTERNS=(
        "${BINARY_NAME}-${OS}-${ARCH}.tar.gz"
        "${BINARY_NAME}-${OS}-${ARCH}.zip"
        "${BINARY_NAME}_${OS}_${ARCH}.tar.gz"
        "${BINARY_NAME}_${OS}_${ARCH}.zip"
    )

    action "$MSG_DOWNLOAD ($LATEST_RELEASE)..."

    # Try direct binary download
    for pattern in "${BINARY_PATTERNS[@]}"; do
        DOWNLOAD_URL="https://github.com/$REPO_USER/$REPO_NAME/releases/download/$LATEST_RELEASE/$pattern"

        if curl -fsSL -o "$BINARY_NAME" "$DOWNLOAD_URL" 2>/dev/null; then
            chmod +x "$BINARY_NAME"
            BINARY_INSTALLED=true
            break
        fi
    done

    # Try archive download and extraction
    if [ "$BINARY_INSTALLED" = false ]; then
        for pattern in "${ARCHIVE_PATTERNS[@]}"; do
            DOWNLOAD_URL="https://github.com/$REPO_USER/$REPO_NAME/releases/download/$LATEST_RELEASE/$pattern"

            if curl -fsSL -o "/tmp/archive-$$" "$DOWNLOAD_URL" 2>/dev/null; then
                if [[ "$pattern" == *.tar.gz ]]; then
                    tar -xzf "/tmp/archive-$$" -C /tmp
                elif [[ "$pattern" == *.zip ]]; then
                    unzip -q "/tmp/archive-$$" -d /tmp
                fi

                # Try to find binary in extracted files
                if [ -f "/tmp/$BINARY_NAME" ]; then
                    mv "/tmp/$BINARY_NAME" "./$BINARY_NAME"
                    chmod +x "$BINARY_NAME"
                    BINARY_INSTALLED=true
                    rm -f "/tmp/archive-$$"
                    break
                fi
                rm -f "/tmp/archive-$$"
            fi
        done
    fi

    if [ "$BINARY_INSTALLED" = false ]; then
        warn "$MSG_DOWNLOAD_FAILED"
    fi
fi

# Fallback: build from source
if [ "$BINARY_INSTALLED" = false ]; then
    action "$MSG_NO_RELEASE"

    # Check build dependencies
    case "$BUILD_LANG" in
        go)
            require_command go "$MSG_NO_GO"
            ;;
        rust)
            require_command cargo "$MSG_NO_RUST"
            ;;
        make)
            require_command make "Make not found. Install build tools and try again."
            ;;
    esac

    require_command git "$MSG_NO_GIT"

    # Check if we're in the repo directory (local usage)
    if [ -d ".git" ] && ([ -f "go.mod" ] || [ -f "Cargo.toml" ] || [ -f "Makefile" ]); then
        action "Building from current directory..."

        eval "$BUILD_CMD"

        if [ $? -ne 0 ]; then
            error "$MSG_BUILD_FAILED"
        fi
    else
        # Remote usage - clone and build
        action "$MSG_CLONE"

        TEMP_DIR=$(mktemp -d)
        cd "$TEMP_DIR"

        if ! git clone "https://github.com/$REPO_USER/$REPO_NAME.git" .; then
            error "Failed to clone repository"
        fi

        action "$MSG_BUILD"
        eval "$BUILD_CMD"

        if [ $? -ne 0 ]; then
            rm -rf "$TEMP_DIR"
            error "$MSG_BUILD_FAILED"
        fi

        # Move binary to original directory
        mv "$BINARY_NAME" "$OLDPWD/"
        cd "$OLDPWD"
        rm -rf "$TEMP_DIR"
    fi
fi

# Install binary
action "$MSG_INSTALL $INSTALL_DIR..."

# Check if sudo is needed
if [[ "$INSTALL_DIR" == /usr/* ]] || [[ "$INSTALL_DIR" == /opt/* ]]; then
    sudo cp "$BINARY_NAME" "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    mkdir -p "$INSTALL_DIR"
    cp "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Clean up
rm -f "$BINARY_NAME"

# Ensure install directory is in PATH
ensure_in_path "$INSTALL_DIR"

echo ""
success "$MSG_INSTALL_COMPLETE"

# Show ASCII art if configured
if [ -n "$ASCII_ART" ]; then
    echo ""
    echo "$ASCII_ART"
fi

# Show next steps
for step in "${NEXT_STEPS[@]}"; do
    echo "$step"
done
