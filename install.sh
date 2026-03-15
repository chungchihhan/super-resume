#!/bin/bash
set -e

# Super Resume Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/chungchihhan/super-resume/main/install.sh | bash

REPO="chungchihhan/super-resume"
MARKETPLACE_NAME="chungchihhan"
PLUGIN_NAME="super-resume"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
PLUGINS_DIR="$HOME/.claude/plugins"
PLUGINS_JSON="$PLUGINS_DIR/installed_plugins.json"
KNOWN_MARKETPLACES_JSON="$PLUGINS_DIR/known_marketplaces.json"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

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

install() {
    PLATFORM=$(detect_platform)
    info "Detected platform: $PLATFORM"

    VERSION=$(get_latest_version)
    [ -z "$VERSION" ] && error "Could not determine latest version. Check https://github.com/${REPO}/releases"
    info "Latest version: $VERSION"
    VERSION_NUM="${VERSION#v}"

    # Install path matching Claude Code's plugin cache structure
    CACHE_DIR="$PLUGINS_DIR/cache/${MARKETPLACE_NAME}/${PLUGIN_NAME}/${VERSION_NUM}"

    EXT="tar.gz"
    [ "$OS" = "windows" ] && EXT="zip"

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/super-resume_${VERSION_NUM}_${PLATFORM}.${EXT}"
    info "Downloading from: $DOWNLOAD_URL"

    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/archive.${EXT}"

    cd "$TMP_DIR"
    if [ "$EXT" = "tar.gz" ]; then
        tar -xzf "archive.${EXT}"
    else
        unzip -q "archive.${EXT}"
    fi

    # Install plugin files to Claude Code cache directory
    mkdir -p "$CACHE_DIR"
    [ -d "skills" ] && cp -r skills "$CACHE_DIR/"
    [ -d ".claude-plugin" ] && cp -r .claude-plugin "$CACHE_DIR/"
    [ -f "README.md" ] && cp README.md "$CACHE_DIR/"
    [ -f "LICENSE" ] && cp LICENSE "$CACHE_DIR/"

    # Place binary in cache dir's bin/
    mkdir -p "$CACHE_DIR/bin"
    cp super-resume "$CACHE_DIR/bin/"
    chmod +x "$CACHE_DIR/bin/super-resume"

    # Symlink to INSTALL_DIR for terminal use
    mkdir -p "$INSTALL_DIR"
    ln -sf "$CACHE_DIR/bin/super-resume" "$INSTALL_DIR/super-resume"
    info "Binary available at: $INSTALL_DIR/super-resume"
    info "Plugin files installed at: $CACHE_DIR"

    # Register marketplace in known_marketplaces.json (if not already registered)
    python3 - <<PYEOF
import json, os

known_file = "$KNOWN_MARKETPLACES_JSON"
try:
    with open(known_file) as f:
        data = json.load(f)
except:
    data = {}

if "$MARKETPLACE_NAME" not in data:
    data["$MARKETPLACE_NAME"] = {
        "source": {"source": "github", "repo": "$REPO"},
        "installLocation": "$PLUGINS_DIR/marketplaces/$MARKETPLACE_NAME",
        "lastUpdated": "2026-01-01T00:00:00.000Z"
    }
    with open(known_file, 'w') as f:
        json.dump(data, f, indent=2)
    print("Marketplace registered")
PYEOF

    # Register plugin in installed_plugins.json
    python3 - <<PYEOF
import json
from datetime import datetime, timezone

plugins_file = "$PLUGINS_JSON"
cache_path = "$CACHE_DIR"
version = "$VERSION_NUM"
now = datetime.now(timezone.utc).strftime('%Y-%m-%dT%H:%M:%S.000Z')

try:
    with open(plugins_file) as f:
        data = json.load(f)
except:
    data = {"version": 2, "plugins": {}}

key = "${PLUGIN_NAME}@${MARKETPLACE_NAME}"
data["plugins"][key] = [{
    "scope": "user",
    "installPath": cache_path,
    "version": version,
    "installedAt": now,
    "lastUpdated": now
}]

with open(plugins_file, 'w') as f:
    json.dump(data, f, indent=2)
print("Plugin registered in Claude Code")
PYEOF

    # Check PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        warn "$INSTALL_DIR is not in your PATH"
        echo ""
        echo "Add to your shell config:"
        echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc"
        echo "  source ~/.zshrc"
        echo ""
    fi

    echo ""
    info "Installation complete!"
    echo ""
    echo "Next steps:"
    echo "  1. Restart Claude Code"
    echo "  2. Run: /plugin install super-resume"
    echo "  3. Run: /setup  (to configure your terminal)"
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
