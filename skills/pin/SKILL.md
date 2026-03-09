---
name: pin
description: Pin a session (current session or by number from list)
user-invocable: true
argument-hint: "[number]"
---

# Pin Session

Pin a session. Pinned sessions appear at the top of the list.

## Task

Pin session: $ARGUMENTS

## Steps

1. Determine which session to pin:
   - If no argument: use current session (`${CLAUDE_SESSION_ID}`)
   - If number (1, 2, 3...): get session ID from the most recent list

2. Run the pin command:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume pin "<session-id>"
```

3. Confirm the session was pinned.

## Examples

- `/pin` - Pin the current session
- `/pin 2` - Pin the second session from the last list
