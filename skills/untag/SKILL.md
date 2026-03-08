---
name: untag
description: Remove a tag from a Claude Code session. Uses current session if no ID specified.
user-invocable: true
argument-hint: "[session-id] <tag>"
---

# Remove Tag from Session

Remove a tag from a session.

## Task

Remove tag from session: $ARGUMENTS

## Steps

1. Parse the arguments:
   - If only one argument: use current session (CLAUDE_SESSION_ID) and the argument as tag
   - If two arguments: first is session ID, second is tag

2. Run the untag command:

```bash
# If only tag provided (use current session)
${CLAUDE_PLUGIN_ROOT}/bin/super-resume untag "${CLAUDE_SESSION_ID}" "<tag>"

# If session ID and tag provided
${CLAUDE_PLUGIN_ROOT}/bin/super-resume untag "<session-id>" "<tag>"
```

3. Confirm the tag was removed

## Usage Examples

- `/untag work` - Remove "work" tag from current session
- `/untag abc123 old` - Remove "old" tag from session abc123

## Notes

- Use `/super-resume` and press `U` to interactively manage tags
