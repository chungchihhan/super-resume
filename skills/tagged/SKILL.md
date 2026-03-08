---
name: tagged
description: Show sessions with a specific tag and resume one
user-invocable: true
argument-hint: "<tag>"
---

# Show Tagged Sessions

List sessions with a specific tag and let the user select one to resume.

## Task

Show sessions tagged with: $ARGUMENTS

## Steps

1. Launch the TUI:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume
```

2. Press `/` to open the filter, then type the tag name to filter sessions.

3. Press `Enter` on any session to resume it.

## Tips

- The filter searches across name, ID, directory, and tags
- Press `Esc` to clear the filter
- Press `A` to toggle between all sessions and current directory
