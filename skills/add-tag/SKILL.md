---
name: add-tag
description: Add a tag to a Claude Code session for organization. Uses current session if no ID specified.
user-invocable: true
argument-hint: "[session-id] <tag>"
---

# Add Tag to Session

Add a tag to a session to help organize and filter sessions.

## Task

Add tag to session: $ARGUMENTS

## Steps

1. Parse the arguments:
   - If only one argument: use current session (CLAUDE_SESSION_ID) and the argument as tag
   - If two arguments: first is session ID, second is tag

2. Run the tag command:

```bash
# If only tag provided (use current session)
${CLAUDE_PLUGIN_ROOT}/bin/super-resume tag "${CLAUDE_SESSION_ID}" "<tag>"

# If session ID and tag provided
${CLAUDE_PLUGIN_ROOT}/bin/super-resume tag "<session-id>" "<tag>"
```

3. Confirm the tag was added

## Usage Examples

- `/add-tag work` - Add "work" tag to current session
- `/add-tag abc123 important` - Add "important" tag to session abc123

## Notes

- Tags are stored in `~/.claude/session-metadata.json`
- Sessions can have multiple tags
- Use `/super-resume` to filter by tag (press `/` and type the tag name)
