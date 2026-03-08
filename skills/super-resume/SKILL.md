---
name: super-resume
description: Launch an interactive TUI to manage Claude Code sessions - browse, pin, tag, filter, and resume sessions
user-invocable: true
argument-hint: ""
---

# Super Resume

Launch an interactive terminal interface for managing Claude Code sessions.

## Features

- **Browse sessions** - View all sessions or filter by current directory
- **Resume directly** - Press Enter to jump straight into a session
- **Pin sessions** - Pinned sessions appear at the top
- **Tag sessions** - Add, edit, and remove tags for organization
- **Filter sessions** - Search by name, ID, directory, or tag
- **Preview messages** - Navigate through conversation history
- **Agent sessions** - Expand/collapse agent sub-sessions

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑/k` `↓/j` | Navigate |
| `Enter` | Resume session |
| `→/l` | Preview session |
| `A` | Toggle all/current directory |
| `S` | Show/hide agent sessions |
| `P` | Pin/unpin |
| `T` | Add tag |
| `U` | Manage tags |
| `D` | Delete |
| `/` | Filter |
| `Q` | Quit |

## Usage

Launch the TUI:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume
```
