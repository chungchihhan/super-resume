---
name: tag
description: Add a tag to a session (current session or by number from list)
user-invocable: true
argument-hint: "<tag> or <number> <tag>"
---

# Tag Session

Add a tag to a session for organization.

## Task

Add tag: $ARGUMENTS

## Steps

1. Parse $ARGUMENTS:
   - If one argument: tag the current session with that tag
   - If two arguments: first is session number from list, second is the tag

2. Run the tag command:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume tag "<session-id>" "<tag>"
```

3. Confirm the tag was added.

## Examples

- `/tag work` - Add "work" tag to current session
- `/tag 2 bug-fix` - Add "bug-fix" tag to second session from list
- `/tag important` - Add "important" tag to current session
