---
name: manage-sessions
description: Launch an interactive TUI to manage Claude Code sessions - view, pin, delete, tag, and preview sessions
user-invocable: true
argument-hint: ""
---

# Session Manager TUI

Launch an interactive terminal interface for managing Claude Code sessions.

## Features

- **List all sessions** with metadata (name, date, directory, message count)
- **Pin/unpin sessions** (📌) for quick access - pinned sessions appear first
- **Delete sessions** with confirmation prompt
- **Add tags** to organize sessions
- **Filter sessions** by name, ID, directory, or tag
- **Preview panel** showing session details and transcript preview
- **Vim-style navigation** (j/k) or arrow keys

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| ↑/k | Move up |
| ↓/j | Move down |
| p | Pin/unpin session |
| d | Delete session |
| t | Add tag |
| / | Filter sessions |
| Tab | Toggle preview panel |
| Enter | Show resume command |
| ? | Toggle help |
| q | Quit |

## Usage

Launch the TUI:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume
```
