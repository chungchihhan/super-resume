---
name: go
description: Resume a Claude Code session by number from the previous list
user-invocable: true
argument-hint: "<number>"
---

# Resume Session

Resume a session from the previous list. Use the number from `/list-session`, `/list-pinned`, or `/list-tagged`.

## Task

Resume session: $ARGUMENTS

## Steps

1. Parse $ARGUMENTS to identify the session:
   - If a number (1, 2, 3...), find the session at that position from the most recent list
   - If a session ID (UUID format), use it directly

2. Get the session ID from the numbered list shown previously in this conversation.

3. Run the resume command:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume resume <session-id>
```

4. Confirm to the user that the session is opening in their configured terminal.

## Examples

- `/go 1` - Resume the first session from the last list
- `/go 3` - Resume the third session from the last list

## Prerequisites

- Run `/list-session`, `/list-pinned`, or `/list-tagged` first to see available sessions
- Configure your terminal with `/setup` if not already done

## Notes

- This command opens the session in a new terminal window/tab
- For Warp users: Warp must have accessibility permissions enabled in System Settings
