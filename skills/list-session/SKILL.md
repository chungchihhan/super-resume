---
name: list-session
description: List Claude Code sessions in current directory (use -a for all)
user-invocable: true
argument-hint: "[-a] [-N]"
---

# List Sessions

List Claude Code sessions. By default, shows sessions in the current directory only.

## Task

List sessions with arguments: $ARGUMENTS

## Steps

1. Parse $ARGUMENTS:
   - `-a` = all directories (not just current)
   - `-N` = limit to N results (e.g., `-10`)
   - Default: current directory, 5 results

2. Run the list command with `--json`:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume list --json [parsed flags]
```

3. Parse the JSON response and format a table like this:

```
┌────┬────┬──────────────────────────────────────────────────────────────────────────────────────────────────────┬──────────────┬──────────────┐
│ #  │ 📌 │ Name                                                                                                 │ Tags         │ Time         │
├────┼────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┼──────────────┼──────────────┤
│ 1  │ 📌 │ Fix authentication bug in the login flow that was causing users to be logged out unexpectedly        │ work, urgent │ Mar 09 14:30 │
│ 2  │    │ Add dark mode toggle to settings page with system preference detection                               │ feature      │ Mar 08 10:15 │
│ 3  │    │ Refactor database layer for better performance and connection pooling                                │ -            │ Mar 07 09:00 │
└────┴────┴──────────────────────────────────────────────────────────────────────────────────────────────────────┴──────────────┴──────────────┘
```

**Table rules:**

- Use box-drawing characters (┌ ┬ ┐ ├ ┼ ┤ └ ┴ ┘ │ ─)
- Name column: 100 characters wide, truncate with "..." if longer
- Tags column: show "-" if no tags, otherwise comma-separated
- Time column: format as "Mon DD HH:MM" (e.g., "Mar 09 14:30")
- Show 📌 emoji for pinned sessions, empty for unpinned

4. After the table, show:

```
Showing X of Y sessions.

Commands:
  /go <n>        Resume session (e.g., /go 1)
  /pin <n>       Pin a session
  /tag <n> <tag> Add tag to session
```

5. Remember the session list for subsequent `/go` commands.

## Examples

- `/list-session` - List 5 sessions in current directory
- `/list-session -a` - List 5 sessions from all directories
- `/list-session -10` - List 10 sessions in current directory
- `/list-session -a -10` - List 10 sessions from all directories
