---
name: setup
description: Configure Super Resume settings (terminal preference, download binary if needed)
user-invocable: true
argument-hint: ""
---

# Setup Super Resume

Download the binary and configure your terminal.

## Steps

### Step 1: Ensure the binary exists

Check if the binary exists:

```bash
ls "${CLAUDE_PLUGIN_ROOT}/bin/super-resume" 2>/dev/null && echo "exists" || echo "missing"
```

If missing, download the pre-built binary from GitHub releases:

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
[ "$ARCH" = "x86_64" ] && ARCH="amd64"
[ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ] && ARCH="arm64"

VERSION=$(grep -o '"version": *"[^"]*"' "${CLAUDE_PLUGIN_ROOT}/.claude-plugin/marketplace.json" | head -1 | grep -o '[0-9][0-9.]*')
FILENAME="super-resume_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/chungchihhan/super-resume/releases/download/v${VERSION}/${FILENAME}"

mkdir -p "${CLAUDE_PLUGIN_ROOT}/bin"
curl -L "$URL" | tar -xz -C "${CLAUDE_PLUGIN_ROOT}/bin" super-resume
chmod +x "${CLAUDE_PLUGIN_ROOT}/bin/super-resume"
```

If download fails, tell the user:
> Could not download the binary. Please check your internet connection or visit https://github.com/chungchihhan/super-resume/releases to download manually.

If successful, confirm: "Binary downloaded successfully."

### Step 2: Symlink binary to ~/.local/bin

So `super-resume` is available in the terminal:

```bash
mkdir -p "$HOME/.local/bin"
ln -sf "${CLAUDE_PLUGIN_ROOT}/bin/super-resume" "$HOME/.local/bin/super-resume"
```

Then check if `~/.local/bin` is in PATH:

```bash
echo $PATH | grep -q "$HOME/.local/bin" && echo "in PATH" || echo "not in PATH"
```

If not in PATH, tell the user to add it to their shell config (e.g. `~/.zshrc` or `~/.bashrc`):

```
export PATH="$HOME/.local/bin:$PATH"
```

And then reload: `source ~/.zshrc` (or `~/.bashrc`).

### Step 3: Check current terminal configuration

```bash
"${CLAUDE_PLUGIN_ROOT}/bin/super-resume" config terminal
```

### Step 4: Configure terminal if not set

If not configured, ask the user which terminal they use:

| Option | Terminal |
|--------|----------|
| 1 | Warp |
| 2 | iTerm2 |
| 3 | Terminal.app (macOS default) |
| 4 | Kitty |
| 5 | Alacritty |

Once user selects, run:

```bash
"${CLAUDE_PLUGIN_ROOT}/bin/super-resume" config terminal <terminal-name>
```

Where `<terminal-name>` is one of: `warp`, `iterm`, `terminal`, `kitty`, `alacritty`

### Step 5: Confirm setup complete

Tell the user setup is complete and they can now use:
- `/list-session` - List sessions
- `/go <n>` - Resume a session (opens new tab in configured terminal)

## Notes for Warp users

Enable Warp in System Settings → Privacy & Security → Accessibility so `/go` can type commands in the new tab.
