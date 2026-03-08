---
name: pin-session
description: Pin a Claude Code session for quick access. Uses current session if no ID specified.
user-invocable: true
argument-hint: "[session-id]"
---

# Pin Session

Pin a session to mark it as important or frequently accessed. Pinned sessions appear at the top of the session list.

## Task

Pin the session: $ARGUMENTS

If no session is specified, pin the current session using CLAUDE_SESSION_ID.

## Steps

1. Run the pin command:

```bash
# If no argument, use current session
${CLAUDE_PLUGIN_ROOT}/bin/super-resume pin "${CLAUDE_SESSION_ID}"

# If session ID provided
${CLAUDE_PLUGIN_ROOT}/bin/super-resume pin "$ARGUMENTS"
```

2. Confirm the session was pinned successfully.

## Usage Examples

- `/pin-session` - Pin the current session
- `/pin-session abc123` - Pin session abc123

## Notes

- Pinned sessions are stored in `~/.claude/session-metadata.json`
- Use `/super-resume` to view all sessions (pinned appear first)
- Use `/unpin-session` to remove the pin
