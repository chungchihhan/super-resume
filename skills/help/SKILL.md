---
name: help
description: Show all Super Resume commands and their parameters
user-invocable: true
argument-hint: ""
---

# Super Resume Help

## Commands (use inside Claude Code)

### Session Management

| Command | Description | Examples |
|---------|-------------|----------|
| `/pin [n]` | Pin a session | `/pin` (current), `/pin 2` (from list) |
| `/unpin [n]` | Unpin a session | `/unpin` (current), `/unpin 1` (from list) |
| `/tag <tag>` | Add tag to current session | `/tag work` |
| `/tag <n> <tag>` | Add tag to session from list | `/tag 2 bug-fix` |
| `/untag <tag>` | Remove tag from current session | `/untag work` |
| `/untag <n> <tag>` | Remove tag from session in list | `/untag 2 old` |

### Listing Sessions

| Command | Description | Examples |
|---------|-------------|----------|
| `/list-session` | List sessions in current dir | `/list-session`, `/list-session -a -10` |
| `/list-pinned` | List pinned sessions | `/list-pinned`, `/list-pinned -10` |
| `/list-tagged <tag>` | List sessions with tag | `/list-tagged work` |

**List flags:**
- `-a` - all directories (not just current)
- `-N` - limit to N results (e.g., `-10`)

### Resuming Sessions

| Command | Description | Examples |
|---------|-------------|----------|
| `/go <n>` | Resume session by number | `/go 1`, `/go 3` |

### Setup & Help

| Command | Description |
|---------|-------------|
| `/setup` | Configure terminal preference |
| `/help` | Show this help |
| `/super-resume` | Show TUI instructions |

## Workflow Example

```
/list-session         # See sessions in current directory
/go 1                 # Resume the first one
```

```
/list-pinned          # See your pinned sessions
/pin 2                # Pin the second one
/go 1                 # Resume the first one
```

```
/list-tagged work     # See sessions tagged "work"
/tag 3 urgent         # Add "urgent" tag to third session
/go 2                 # Resume the second one
```

## First Time Setup

Run `/setup` to configure which terminal to use for resuming sessions.

## TUI (run from terminal, not inside Claude Code)

```bash
super-resume
```

The TUI provides a full interactive experience with keyboard navigation, preview, and direct resume.
