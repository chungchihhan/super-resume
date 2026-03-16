# Super Resume

A session manager for Claude Code - browse, filter, pin, tag, and resume sessions from both the TUI and directly inside Claude Code via slash commands.
<video controls src="super-resume-deom.mp4" title="demo"></video>

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![Claude Code](https://img.shields.io/badge/Claude_Code-Plugin-D97706?logo=anthropic&logoColor=white)](https://github.com/chungchihhan/super-resume)
[![TUI](https://img.shields.io/badge/TUI-Bubbletea-FF75B7)](https://github.com/charmbracelet/bubbletea)

## Features

- **Browse sessions** - View all sessions or filter by current directory
- **Resume directly** - Jump straight into any session from TUI or Claude Code
- **Pin sessions** - Pinned sessions appear first
- **Tag sessions** - Add, edit, and remove tags for organization
- **Filter sessions** - Search by name, ID, directory, or tag
- **Preview messages** - Navigate through conversation history
- **Slash commands** - Manage sessions without leaving Claude Code

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- Claude Code CLI

## Installation

### Option 1: Claude Code Plugin Marketplace (Recommended)

```bash
# Add the marketplace
/plugin marketplace add chungchihhan/super-resume

# Install the plugin
/plugin install super-resume
```

### Option 2(TUI only): One-liner Install

If you only want the `super-resume` TUI in your terminal without Claude Code skills:

```bash
curl -fsSL https://raw.githubusercontent.com/chungchihhan/super-resume/main/install.sh | bash
```

### Option3: Build from Source

```bash
git clone https://github.com/chungchihhan/super-resume.git
cd super-resume
make build
```

## Usage

### TUI (Terminal)

```bash
super-resume
```

### CLI Commands

```bash
# List sessions
super-resume list                        # Current directory, 5 sessions
super-resume list -10                    # Current directory, 10 sessions
super-resume list -a                     # All directories
super-resume list -a -10 --pinned        # All dirs, pinned only
super-resume list --tagged work          # Sessions tagged "work"
super-resume list --json                 # JSON output

# Session management
super-resume pin [session-id]            # Pin a session (defaults to current)
super-resume unpin [session-id]          # Unpin a session
super-resume tag <session-id> <tag>      # Add a tag
super-resume untag <session-id> <tag>    # Remove a tag
super-resume delete <session-id>         # Delete a session

# Resume in terminal
super-resume config terminal <type>      # Set terminal (warp, iterm, terminal, kitty, alacritty)
super-resume resume <session-id>         # Open session in configured terminal
```

### Slash Commands (inside Claude Code)

First time setup - configure your terminal:

```
/setup
```

| Command              | Description                             |
| -------------------- | --------------------------------------- |
| `/list-session`      | List sessions in current directory      |
| `/list-session -a`   | List sessions from all directories      |
| `/list-session -10`  | List 10 sessions                        |
| `/list-pinned`       | List pinned sessions                    |
| `/list-tagged <tag>` | List sessions with a tag                |
| `/go <n>`            | Resume session by number from last list |
| `/pin`               | Pin current session                     |
| `/pin <n>`           | Pin session by number from list         |
| `/unpin`             | Unpin current session                   |
| `/tag <tag>`         | Tag current session                     |
| `/tag <n> <tag>`     | Tag session by number from list         |
| `/untag <tag>`       | Remove tag from current session         |
| `/setup`             | Configure terminal preference           |
| `/help`              | Show all commands                       |

**Example workflow:**

```
/list-session          # See sessions in current directory
/go 1                  # Resume the first one (opens in your terminal)
```

```
/list-pinned           # See pinned sessions
/tag 2 work            # Tag the second one
/go 1                  # Resume the first one
```

> **Note for Warp users:** Enable Warp in System Settings → Privacy & Security → Accessibility for `/go` to work.

## TUI Keyboard Shortcuts

### List View

| Key     | Action                       |
| ------- | ---------------------------- |
| `↑/k`   | Move up                      |
| `↓/j`   | Move down                    |
| `PgUp`  | Page up                      |
| `PgDn`  | Page down                    |
| `Enter` | **Resume session**           |
| `→/l`   | Preview session              |
| `A`     | Toggle all/current directory |
| `S`     | Show/hide agent sessions     |
| `P`     | Pin/unpin session            |
| `T`     | Add tag                      |
| `U`     | Manage tags (edit/delete)    |
| `D`     | Delete session               |
| `/`     | Filter sessions              |
| `Esc`   | Clear filter or quit         |
| `Q`     | Quit                         |

### Preview View

| Key       | Action                     |
| --------- | -------------------------- |
| `↑/k`     | Scroll up                  |
| `↓/j`     | Scroll down                |
| `Enter`   | Resume at selected message |
| `←/h/Esc` | Back to list               |

### Tag Management (U)

| Key     | Action              |
| ------- | ------------------- |
| `←/→`   | Select tag          |
| `D`     | Delete selected tag |
| `Enter` | Edit selected tag   |
| `Esc`   | Cancel              |

## Display

Sessions are displayed with:

```
 Pinned   tag1   tag2
▶ Session name from first message...
  5 minutes ago · 42 msgs · ~/path/to/project
```

- **Badges** - Pinned status and tags shown above the session
- **Path** - Full working directory with `~` for home
- **Agent sessions** - Indented under their parent when expanded

## Data Storage

- **Sessions**: `~/.claude/projects/**/*.jsonl`
- **Metadata** (pins, tags, config): `~/.claude/session-metadata.json`

## License

MIT
