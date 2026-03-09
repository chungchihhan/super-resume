---
name: setup
description: Configure Super Resume settings (terminal preference)
user-invocable: true
argument-hint: ""
---

# Setup Super Resume

Configure Super Resume settings.

## Steps

1. Check current terminal configuration:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume config terminal
```

2. If not configured, ask the user which terminal they use:

**Which terminal do you use?**

| Option | Terminal |
|--------|----------|
| 1 | Terminal.app (macOS default) |
| 2 | iTerm2 |
| 3 | Warp |
| 4 | Kitty |
| 5 | Alacritty |

3. Once user selects, run:

```bash
${CLAUDE_PLUGIN_ROOT}/bin/super-resume config terminal <terminal-name>
```

Where `<terminal-name>` is one of: `terminal`, `iterm`, `warp`, `kitty`, `alacritty`

4. Confirm the setting was saved.

5. **Important for Warp users:** Remind them to enable accessibility permissions:
   - Open System Settings > Privacy & Security > Accessibility
   - Enable access for Warp

## Notes

- This only needs to be done once
- You can re-run `/setup` anytime to change your terminal preference
