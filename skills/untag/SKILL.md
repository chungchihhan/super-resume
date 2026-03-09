---
name: untag
description: Remove a tag from a session (current session or by number from list)
user-invocable: true
argument-hint: "<tag> or <number> <tag>"
---

# Untag Session

Remove a tag from a session.

## Task

Remove tag: $ARGUMENTS

## Steps

1. Parse $ARGUMENTS:
   - If one argument: remove that tag from current session
   - If two arguments: first is session number from list, second is the tag

2. Run the untag command:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume untag "<session-id>" "<tag>"
```

3. Confirm the tag was removed.

## Examples

- `/untag work` - Remove "work" tag from current session
- `/untag 2 old` - Remove "old" tag from second session from list
