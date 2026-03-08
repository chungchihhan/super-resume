# Super Resume

A TUI for managing Claude Code sessions - browse, filter, pin, tag, and resume sessions directly.

![Demo](https://img.shields.io/badge/TUI-bubbletea-blue)

## Features

- **Browse sessions** - View all sessions or filter by current directory
- **Resume directly** - Press Enter to jump straight into a session
- **Pin sessions** - Pinned sessions appear first
- **Tag sessions** - Add, edit, and remove tags for organization
- **Filter sessions** - Search by name, ID, directory, or tag
- **Preview messages** - Navigate through conversation history
- **Agent sessions** - Expand/collapse agent sub-sessions per parent

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- Claude Code CLI

## Installation

### One-liner Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/chungchihhan/super-resume/main/install.sh | bash
```

### Claude Code Plugin Marketplace

```bash
# Add the marketplace
/plugin marketplace add chungchihhan/super-resume

# Install the plugin
/plugin install super-resume
```

### Go Install

```bash
go install github.com/chungchihhan/super-resume/cmd/super-resume@latest
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/chungchihhan/super-resume.git
cd super-resume

# Build the binary
make build

# Or manually:
go mod tidy
go build -o bin/super-resume ./cmd/super-resume
```

## Usage

### Launch the TUI

```bash
./bin/super-resume
```

### CLI Commands

```bash
./bin/super-resume list              # List all sessions
./bin/super-resume pin <session-id>  # Pin a session
./bin/super-resume unpin <session-id> # Unpin a session
./bin/super-resume delete <session-id> # Delete a session
./bin/super-resume tag <session-id> <tag> # Add a tag
```

### Use as a Claude Code Plugin

```bash
claude --plugin-dir ~/path/to/super-resume
```

Then use `/manage-sessions` to launch the TUI.

## Keyboard Shortcuts

### List View

| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `PgUp` | Page up |
| `PgDn` | Page down |
| `Enter` | **Resume session** (launches Claude) |
| `→/l` | Preview session |
| `A` | Toggle all/current directory |
| `S` | Show/hide agent sessions |
| `P` | Pin/unpin session |
| `T` | Add tag |
| `U` | Manage tags (edit/delete) |
| `D` | Delete session |
| `/` | Filter sessions |
| `Esc` | Clear filter or quit |
| `Q` | Quit |

### Preview View

| Key | Action |
|-----|--------|
| `↑/k` | Scroll up |
| `↓/j` | Scroll down |
| `Enter` | Resume at selected message |
| `←/h/Esc` | Back to list |

### Tag Management (U)

| Key | Action |
|-----|--------|
| `←/→` | Select tag |
| `D` | Delete selected tag |
| `Enter` | Edit selected tag |
| `Esc` | Cancel |

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
- **Metadata** (pins, tags): `~/.claude/session-metadata.json`

## Development

```bash
make fmt    # Format code
make test   # Run tests
make build  # Build binary
make run    # Build and run
```

## License

MIT
