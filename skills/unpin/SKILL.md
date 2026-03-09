---
name: unpin
description: Unpin a session (current session or by number from list)
user-invocable: true
argument-hint: "[number]"
---

# Unpin Session

Remove the pin from a session.

## Task

Unpin session: $ARGUMENTS

## Steps

1. Determine which session to unpin:
   - If no argument: use current session (`${CLAUDE_SESSION_ID}`)
   - If number (1, 2, 3...): get session ID from the most recent list

2. Run the unpin command:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume unpin "<session-id>"
```

3. Confirm the session was unpinned.

## Examples

- `/unpin` - Unpin the current session
- `/unpin 1` - Unpin the first session from the last list
