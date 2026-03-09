---
name: list-tagged
description: List sessions with a specific tag
user-invocable: true
argument-hint: "<tag> [-N]"
---

# List Tagged Sessions

Show sessions that have a specific tag.

## Task

List sessions with tag: $ARGUMENTS

## Steps

1. Parse $ARGUMENTS:
   - First argument is the tag name (required)
   - `-N` = limit to N results (e.g., `-10`)
   - Default: 5 results from all directories

2. Run the list command with `--json --tagged <tag> -a`:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume list --json --tagged <tag> -a [limit flag]
```

3. Parse the JSON response and format a table like this:

```
┌────┬────┬──────────────────────────────────────────────────────────────────────────────────────────────────────┬──────────────┬──────────────┐
│ #  │ 📌 │ Name                                                                                                 │ Tags         │ Time         │
├────┼────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┼──────────────┼──────────────┤
│ 1  │ 📌 │ Fix authentication bug in the login flow that was causing users to be logged out unexpectedly        │ work, urgent │ Mar 09 14:30 │
│ 2  │    │ Add user dashboard feature with analytics and reporting capabilities                                 │ work         │ Mar 08 10:15 │
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
Showing X sessions tagged "work".

Commands:
  /go <n>          Resume session (e.g., /go 1)
  /pin <n>         Pin a session
  /untag <n> <tag> Remove tag from session
```

5. Remember the session list for subsequent `/go` commands.

## Examples

- `/list-tagged work` - List up to 5 sessions tagged "work"
- `/list-tagged bug-fix -10` - List up to 10 sessions tagged "bug-fix"
