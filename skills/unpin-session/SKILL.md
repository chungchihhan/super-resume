---
name: unpin-session
description: Unpin a Claude Code session. Uses current session if no ID specified.
user-invocable: true
argument-hint: "[session-id]"
---

# Unpin Session

Remove the pin from a session.

## Task

Unpin the session: $ARGUMENTS

If no session is specified, unpin the current session using CLAUDE_SESSION_ID.

## Steps

1. Run the unpin command:

```bash
# If no argument, use current session
${CLAUDE_PLUGIN_ROOT}/bin/super-resume unpin "${CLAUDE_SESSION_ID}"

# If session ID provided
${CLAUDE_PLUGIN_ROOT}/bin/super-resume unpin "$ARGUMENTS"
```

2. Confirm the session was unpinned successfully.

## Usage Examples

- `/unpin-session` - Unpin the current session
- `/unpin-session abc123` - Unpin session abc123
