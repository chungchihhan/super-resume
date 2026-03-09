// super-resume is a TUI and CLI for managing Claude Code sessions.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chungchihhan/super-resume/internal/metadata"
	"github.com/chungchihhan/super-resume/internal/session"
	"github.com/chungchihhan/super-resume/internal/tui"
)

// listFlags holds parsed flags for the list command.
type listFlags struct {
	allDirs bool
	limit   int
	pinned  bool
	tag     string
	jsonOut bool
}

// parseListFlags parses flags for the list command.
// Supports: -a, -N (e.g., -10), --pinned, --tagged <tag>, --json
func parseListFlags(args []string) listFlags {
	flags := listFlags{
		limit: 5, // default
	}

	numberRegex := regexp.MustCompile(`^-(\d+)$`)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "-a":
			flags.allDirs = true
		case arg == "--pinned":
			flags.pinned = true
		case arg == "--json":
			flags.jsonOut = true
		case arg == "--tagged":
			if i+1 < len(args) {
				i++
				flags.tag = args[i]
			}
		default:
			// Check for -N pattern (e.g., -10)
			if matches := numberRegex.FindStringSubmatch(arg); len(matches) == 2 {
				if n, err := strconv.Atoi(matches[1]); err == nil {
					flags.limit = n
				}
			}
		}
	}

	return flags
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Check for CLI commands
	if len(os.Args) > 1 {
		return handleCommand(os.Args[1:])
	}

	// Launch TUI
	return runTUI()
}

func runTUI() error {
	meta, err := metadata.New()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	mgr, err := session.NewManager(meta)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	model := tui.New(mgr, meta)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	// Check if a session was selected for resume
	if m, ok := finalModel.(tui.Model); ok {
		sessionID, cwd, _ := m.GetResumeInfo()
		if sessionID != "" {
			// Change to the session's working directory first
			if cwd != "" {
				if err := os.Chdir(cwd); err != nil {
					return fmt.Errorf("failed to change to session directory %s: %w", cwd, err)
				}
			}

			// Execute claude --resume directly
			claudePath, err := exec.LookPath("claude")
			if err != nil {
				return fmt.Errorf("claude not found in PATH: %w", err)
			}
			args := []string{"claude", "--resume", sessionID}
			return syscall.Exec(claudePath, args, os.Environ())
		}
	}

	return nil
}

func handleCommand(args []string) error {
	meta, err := metadata.New()
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	mgr, err := session.NewManager(meta)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "pin":
		return cmdPin(mgr, meta, cmdArgs)
	case "unpin":
		return cmdUnpin(meta, cmdArgs)
	case "delete":
		return cmdDelete(mgr, cmdArgs)
	case "tag":
		return cmdTag(meta, cmdArgs)
	case "untag":
		return cmdUntag(meta, cmdArgs)
	case "list":
		flags := parseListFlags(cmdArgs)
		return cmdList(mgr, flags)
	case "config":
		return cmdConfig(meta, cmdArgs)
	case "resume":
		return cmdResume(mgr, meta, cmdArgs)
	case "help", "-h", "--help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s\nRun 'super-resume help' for usage", cmd)
	}
}

func cmdPin(mgr *session.Manager, meta *metadata.Store, args []string) error {
	if len(args) == 0 {
		// Try to pin current session from env
		sessionID := os.Getenv("CLAUDE_SESSION_ID")
		if sessionID == "" {
			return fmt.Errorf("no session specified and CLAUDE_SESSION_ID not set")
		}
		args = []string{sessionID}
	}

	s, err := mgr.Find(args[0])
	if err != nil {
		return fmt.Errorf("session not found: %s", args[0])
	}

	// Check if session has actual content
	if s.Name == "<Empty session>" || s.Name == "Warmup" {
		return fmt.Errorf("cannot pin this session: no user messages yet (only slash commands don't count)")
	}

	if err := meta.Pin(s.ID); err != nil {
		return err
	}

	fmt.Printf("📌 Pinned: %s\n", s.Name)
	return nil
}

func cmdUnpin(meta *metadata.Store, args []string) error {
	if len(args) == 0 {
		sessionID := os.Getenv("CLAUDE_SESSION_ID")
		if sessionID == "" {
			return fmt.Errorf("no session specified and CLAUDE_SESSION_ID not set")
		}
		args = []string{sessionID}
	}

	if err := meta.Unpin(args[0]); err != nil {
		return err
	}

	fmt.Printf("Unpinned: %s\n", args[0])
	return nil
}

func cmdDelete(mgr *session.Manager, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("session ID required")
	}

	s, err := mgr.Find(args[0])
	if err != nil {
		return fmt.Errorf("session not found: %s", args[0])
	}

	// Confirm
	fmt.Printf("Delete session '%s'? (y/N): ", s.Name)
	var response string
	fmt.Scanln(&response)
	if response != "y" && response != "Y" {
		fmt.Println("Cancelled")
		return nil
	}

	if err := mgr.Delete(s.ID); err != nil {
		return err
	}

	fmt.Printf("🗑️  Deleted: %s\n", s.Name)
	return nil
}

func cmdTag(meta *metadata.Store, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: super-resume tag <session-id> <tag>")
	}

	sessionID := args[0]
	tag := args[1]

	if err := meta.AddTag(sessionID, tag); err != nil {
		return err
	}

	fmt.Printf("🏷️  Added tag '%s' to %s\n", tag, sessionID)
	return nil
}

func cmdUntag(meta *metadata.Store, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: super-resume untag <session-id> <tag>")
	}

	sessionID := args[0]
	tag := args[1]

	if err := meta.RemoveTag(sessionID, tag); err != nil {
		return err
	}

	fmt.Printf("🏷️  Removed tag '%s' from %s\n", tag, sessionID)
	return nil
}

// listOutput represents the JSON output structure for list command.
type listOutput struct {
	Sessions []sessionOutput `json:"sessions"`
	Total    int             `json:"total"`
	Showing  int             `json:"showing"`
	Filters  filterInfo      `json:"filters"`
}

type sessionOutput struct {
	Number       int      `json:"number"`
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Pinned       bool     `json:"pinned"`
	Tags         []string `json:"tags"`
	Modified     string   `json:"modified"`
	MessageCount int      `json:"message_count"`
}

type filterInfo struct {
	AllDirs    bool   `json:"all_dirs"`
	PinnedOnly bool   `json:"pinned_only"`
	Tag        string `json:"tag"`
	Limit      int    `json:"limit"`
}

func cmdList(mgr *session.Manager, flags listFlags) error {
	// Get sessions based on scope
	var sessions []*session.Session
	var err error

	if flags.allDirs {
		sessions, err = mgr.List()
	} else {
		sessions, err = mgr.ListForCurrentDir()
	}
	if err != nil {
		return err
	}

	totalBeforeFilter := len(sessions)

	// Filter out agent sessions and empty/warmup sessions
	{
		var filtered []*session.Session
		for _, s := range sessions {
			// Skip agent sessions
			if s.IsAgent {
				continue
			}
			// Skip empty sessions and warmup sessions
			if s.Name == "<Empty session>" || s.Name == "Warmup" {
				continue
			}
			filtered = append(filtered, s)
		}
		sessions = filtered
	}

	// Filter by pinned
	if flags.pinned {
		var filtered []*session.Session
		for _, s := range sessions {
			if s.IsPinned {
				filtered = append(filtered, s)
			}
		}
		sessions = filtered
	}

	// Filter by tag
	if flags.tag != "" {
		var filtered []*session.Session
		for _, s := range sessions {
			for _, t := range s.Tags {
				if t == flags.tag {
					filtered = append(filtered, s)
					break
				}
			}
		}
		sessions = filtered
	}

	// Apply limit
	if flags.limit > 0 && len(sessions) > flags.limit {
		sessions = sessions[:flags.limit]
	}

	// Output
	if flags.jsonOut {
		return outputJSON(sessions, totalBeforeFilter, flags)
	}
	return outputHuman(sessions)
}

func outputJSON(sessions []*session.Session, total int, flags listFlags) error {
	output := listOutput{
		Sessions: make([]sessionOutput, len(sessions)),
		Total:    total,
		Showing:  len(sessions),
		Filters: filterInfo{
			AllDirs:    flags.allDirs,
			PinnedOnly: flags.pinned,
			Tag:        flags.tag,
			Limit:      flags.limit,
		},
	}

	for i, s := range sessions {
		tags := s.Tags
		if tags == nil {
			tags = []string{}
		}
		output.Sessions[i] = sessionOutput{
			Number:       i + 1,
			ID:           s.ID,
			Name:         s.Name,
			Path:         s.Cwd,
			Pinned:       s.IsPinned,
			Tags:         tags,
			Modified:     s.Modified.Format("2006-01-02T15:04:05Z07:00"),
			MessageCount: s.MessageCount,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func outputHuman(sessions []*session.Session) error {
	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	// Print header
	fmt.Println("┌────┬────┬──────────────────────────────────────────────────────────────────────────────────────────────────────┬──────────────┬──────────────┐")
	fmt.Printf("│ #  │ 📌 │ %-100s │ %-12s │ %-12s │\n",
		"Name", "Tags", "Time")
	fmt.Println("├────┼────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┼──────────────┼──────────────┤")

	for i, s := range sessions {
		pin := "  "
		if s.IsPinned {
			pin = "📌"
		}

		tags := "-"
		if len(s.Tags) > 0 {
			tags = joinTags(s.Tags)
		}

		fmt.Printf("│ %-2d │ %s │ %-100s │ %-12s │ %-12s │\n",
			i+1,
			pin,
			truncate(s.Name, 100),
			truncate(tags, 12),
			s.Modified.Format("Jan 02 15:04"),
		)
	}

	fmt.Println("└────┴────┴──────────────────────────────────────────────────────────────────────────────────────────────────────┴──────────────┴──────────────┘")

	return nil
}

func cmdConfig(meta *metadata.Store, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: super-resume config <setting> [value]\n  settings: terminal")
	}

	setting := args[0]
	settingArgs := args[1:]

	switch setting {
	case "terminal":
		return cmdConfigTerminal(meta, settingArgs)
	default:
		return fmt.Errorf("unknown setting: %s\n  available: terminal", setting)
	}
}

func cmdConfigTerminal(meta *metadata.Store, args []string) error {
	// Show current value if no argument
	if len(args) == 0 {
		current := meta.GetTerminal()
		if current == "" {
			fmt.Println("Terminal: (not set)")
			fmt.Println("\nSupported terminals:")
			for _, t := range metadata.SupportedTerminals() {
				fmt.Printf("  %s\n", t)
			}
			fmt.Println("\nSet with: super-resume config terminal <name>")
		} else {
			fmt.Printf("Terminal: %s\n", current)
		}
		return nil
	}

	// Validate terminal type
	terminal := args[0]
	valid := false
	for _, t := range metadata.SupportedTerminals() {
		if t == terminal {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("unsupported terminal: %s\n  supported: %v", terminal, metadata.SupportedTerminals())
	}

	if err := meta.SetTerminal(terminal); err != nil {
		return err
	}

	fmt.Printf("Terminal set to: %s\n", terminal)
	return nil
}

func cmdResume(mgr *session.Manager, meta *metadata.Store, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: super-resume resume <session-id>")
	}

	sessionID := args[0]

	// Find the session to verify it exists
	s, err := mgr.Find(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Get configured terminal
	terminal := meta.GetTerminal()
	if terminal == "" {
		return fmt.Errorf("terminal not configured\nRun: super-resume config terminal <terminal-type>")
	}

	// Build the command to run
	claudeCmd := fmt.Sprintf("claude --resume %s", s.ID)

	var err2 error
	switch terminal {
	case "terminal":
		// Terminal.app supports running command directly via AppleScript
		fullCmd := fmt.Sprintf("cd %q && %s", s.Cwd, claudeCmd)
		script := fmt.Sprintf(`tell app "Terminal" to do script "%s"`, escapeAppleScript(fullCmd))
		err2 = exec.Command("osascript", "-e", script).Run()
	case "iterm":
		// iTerm2 supports running command directly
		fullCmd := fmt.Sprintf("cd %q && %s", s.Cwd, claudeCmd)
		script := fmt.Sprintf(`tell app "iTerm" to create window with default profile command "%s"`, escapeAppleScript(fullCmd))
		err2 = exec.Command("osascript", "-e", script).Run()
	case "warp":
		// Warp: open tab at directory, then use keystrokes to type command
		err2 = openWarpWithCommand(s.Cwd, claudeCmd)
	case "kitty":
		err2 = exec.Command("kitty", "@", "launch", "--type=os-window", "--cwd", s.Cwd, "--", "claude", "--resume", s.ID).Run()
	case "alacritty":
		err2 = exec.Command("alacritty", "--working-directory", s.Cwd, "-e", "claude", "--resume", s.ID).Run()
	default:
		return fmt.Errorf("unsupported terminal: %s", terminal)
	}

	if err2 != nil {
		return fmt.Errorf("failed to open terminal: %w", err2)
	}

	fmt.Printf("Opened session in %s: %s\n", terminal, s.Name)
	return nil
}

// openWarpWithCommand opens Warp at the given directory and types the command.
// Requires Warp to have accessibility permissions in System Settings.
func openWarpWithCommand(cwd, command string) error {
	// URL encode the path
	urlPath := "warp://action/new_tab?path=" + urlEncode(cwd)

	// Open Warp tab at directory
	if err := exec.Command("open", urlPath).Run(); err != nil {
		return fmt.Errorf("failed to open Warp: %w", err)
	}

	// Wait for Warp to open
	time.Sleep(500 * time.Millisecond)

	// Type the command using AppleScript keystrokes
	keystrokeScript := fmt.Sprintf(`tell application "System Events" to keystroke "%s"`, escapeAppleScript(command))
	enterScript := `tell application "System Events" to key code 36`

	if err := exec.Command("osascript", "-e", keystrokeScript, "-e", enterScript).Run(); err != nil {
		return fmt.Errorf("failed to type command (ensure Warp has accessibility permissions): %w", err)
	}

	return nil
}

// urlEncode encodes a string for use in a URL.
func urlEncode(s string) string {
	// Simple URL encoding for path
	result := ""
	for _, c := range s {
		switch {
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'):
			result += string(c)
		case c == '-' || c == '_' || c == '.' || c == '~' || c == '/':
			result += string(c)
		default:
			result += fmt.Sprintf("%%%02X", c)
		}
	}
	return result
}

// escapeAppleScript escapes a string for use in AppleScript.
func escapeAppleScript(s string) string {
	// Escape backslashes and double quotes
	s = regexp.MustCompile(`\\`).ReplaceAllString(s, `\\\\`)
	s = regexp.MustCompile(`"`).ReplaceAllString(s, `\"`)
	return s
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func joinTags(tags []string) string {
	result := ""
	for i, t := range tags {
		if i > 0 {
			result += ", "
		}
		result += t
	}
	return result
}

func printHelp() {
	fmt.Println(`Super Resume - Claude Code Session Manager

USAGE:
    super-resume              Launch interactive TUI
    super-resume <command>    Run a command

COMMANDS:
    list [flags]         List sessions
    pin [session]        Pin a session (uses current session if not specified)
    unpin [session]      Unpin a session
    tag <session> <tag>  Add a tag to a session
    untag <session> <tag>  Remove a tag from a session
    delete <session>     Delete a session
    config terminal [type]  Get/set terminal (terminal, iterm, warp, kitty, alacritty)
    resume <session>     Open session in configured terminal
    help                 Show this help

LIST FLAGS:
    -a               Show all directories (default: current directory only)
    -N               Limit to N results, e.g., -10 (default: 5)
    --pinned         Show only pinned sessions
    --tagged <tag>   Show only sessions with specified tag
    --json           Output in JSON format (for scripts/skills)

EXAMPLES:
    super-resume list                    # Current dir, 5 sessions
    super-resume list -10                # Current dir, 10 sessions
    super-resume list -a                 # All dirs, 5 sessions
    super-resume list -a -10 --pinned    # All dirs, 10 pinned sessions
    super-resume list --tagged work      # Sessions tagged "work"
    super-resume list --json             # JSON output

TUI KEYBINDINGS:
    ↑/k, ↓/j    Navigate
    Enter       Resume session
    →/l         Preview session
    A           Toggle all/current directory
    S           Show/hide agent sessions
    P           Pin/unpin session
    T           Add tag
    U           Manage tags
    D           Delete session
    /           Filter sessions
    Q           Quit`)
}
