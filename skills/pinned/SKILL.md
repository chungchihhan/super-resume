---
name: pinned
description: Show all pinned sessions and resume one
user-invocable: true
argument-hint: ""
---

# Show Pinned Sessions

List all pinned sessions and let the user select one to resume.

## Steps

1. Launch the TUI which shows pinned sessions at the top:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume
```

2. Pinned sessions appear first in the list with a "Pinned" badge.

3. Press `Enter` on any session to resume it.

## Tips

- Press `/` and type to filter sessions
- Press `A` to toggle between all sessions and current directory
- Press `P` to pin/unpin sessions
