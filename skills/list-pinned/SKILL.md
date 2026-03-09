---
name: list-pinned
description: List pinned sessions
user-invocable: true
argument-hint: "[-N]"
---

# List Pinned Sessions

Show pinned sessions from all directories.

## Task

List pinned sessions: $ARGUMENTS

## Steps

1. Parse $ARGUMENTS:
   - `-N` = limit to N results (e.g., `-10`)
   - Default: 5 results

2. Run the list command with `--json --pinned -a`:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume list --json --pinned -a [limit flag]
```

3. Parse the JSON response and format a table like this:

```
┌────┬────┬──────────────────────────────────────────────────────────────────────────────────────────────────────┬──────────────┬──────────────┐
│ #  │ 📌 │ Name                                                                                                 │ Tags         │ Time         │
├────┼────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┼──────────────┼──────────────┤
│ 1  │ 📌 │ Fix authentication bug in the login flow that was causing users to be logged out unexpectedly        │ work, urgent │ Mar 09 14:30 │
│ 2  │ 📌 │ Important refactor task for the database layer                                                       │ priority     │ Mar 08 10:15 │
└────┴────┴──────────────────────────────────────────────────────────────────────────────────────────────────────┴──────────────┴──────────────┘
```

**Table rules:**

- Use box-drawing characters (┌ ┬ ┐ ├ ┼ ┤ └ ┴ ┘ │ ─)
- Name column: 100 characters wide, truncate with "..." if longer
- Tags column: show "-" if no tags, otherwise comma-separated
- Time column: format as "Mon DD HH:MM" (e.g., "Mar 09 14:30")
- Show 📌 emoji for pinned sessions

4. After the table, show:

```
Showing X pinned sessions.

Commands:
  /go <n>     Resume session (e.g., /go 1)
  /unpin <n>  Unpin a session
```

5. Remember the session list for subsequent `/go` commands.

## Examples

- `/list-pinned` - List up to 5 pinned sessions
- `/list-pinned -10` - List up to 10 pinned sessions
