#!/bin/bash
set -e

# Super Resume Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/chungchihhan/super-resume/main/install.sh | bash

REPO="chungchihhan/super-resume"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
PLUGIN_DIR="${PLUGIN_DIR:-$HOME/.claude/plugins/super-resume}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux*) OS="linux" ;;
        darwin*) OS="darwin" ;;
        mingw*|msys*|cygwin*) OS="windows" ;;
        *) error "Unsupported OS: $OS" ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac

    echo "${OS}_${ARCH}"
}

# Get latest release version
get_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/'
}

# Download and install
install() {
    PLATFORM=$(detect_platform)
    info "Detected platform: $PLATFORM"

    VERSION=$(get_latest_version)
    if [ -z "$VERSION" ]; then
        error "Could not determine latest version. Please check https://github.com/${REPO}/releases"
    fi
    info "Latest version: $VERSION"

    # Construct download URL
    EXT="tar.gz"
    if [ "$OS" = "windows" ]; then
        EXT="zip"
    fi

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/super-resume_${VERSION#v}_${PLATFORM}.${EXT}"
    info "Downloading from: $DOWNLOAD_URL"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download
    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/super-resume.${EXT}"

    # Extract
    cd "$TMP_DIR"
    if [ "$EXT" = "tar.gz" ]; then
        tar -xzf "super-resume.${EXT}"
    else
        unzip -q "super-resume.${EXT}"
    fi

    # Install binary
    mkdir -p "$INSTALL_DIR"
    cp super-resume "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/super-resume"
    info "Binary installed to: $INSTALL_DIR/super-resume"

    # Install plugin files
    mkdir -p "$PLUGIN_DIR"
    mkdir -p "$PLUGIN_DIR/.claude-plugin"
    mkdir -p "$PLUGIN_DIR/skills"

    if [ -d "skills" ]; then
        cp -r skills/* "$PLUGIN_DIR/skills/"
    fi
    if [ -d ".claude-plugin" ]; then
        cp -r .claude-plugin/* "$PLUGIN_DIR/.claude-plugin/"
    fi
    if [ -f "README.md" ]; then
        cp README.md "$PLUGIN_DIR/"
    fi
    info "Plugin files installed to: $PLUGIN_DIR"

    # Check if INSTALL_DIR is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "$INSTALL_DIR is not in your PATH"
        echo ""
        echo "Add it to your shell config:"
        echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc"
        echo "  # or for zsh:"
        echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc"
        echo ""
    fi

    echo ""
    info "Installation complete!"
    echo ""
    echo "To use with Claude Code:"
    echo "  claude --plugin-dir $PLUGIN_DIR"
    echo ""
    echo "Or add to your Claude settings to make it permanent."
    echo ""
    echo "Available skills:"
    echo "  /setup            - First-time setup (configure terminal)"
    echo "  /super-resume     - Open the TUI in your terminal"
    echo "  /list-session     - List recent sessions"
    echo "  /go <n>           - Resume session by number (opens new tab)"
    echo "  /pin              - Pin current or numbered session"
    echo "  /tag              - Tag current or numbered session"
    echo "  /help             - Show all available commands"
}

install
