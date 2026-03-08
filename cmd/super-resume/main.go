// super-resume is a TUI for managing Claude Code sessions.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/chungchihhan/super-resume/internal/metadata"
	"github.com/chungchihhan/super-resume/internal/session"
	"github.com/chungchihhan/super-resume/internal/tui"
)

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
	case "list":
		return cmdList(mgr)
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

func cmdList(mgr *session.Manager) error {
	sessions, err := mgr.List()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	for _, s := range sessions {
		pin := "  "
		if s.IsPinned {
			pin = "📌"
		}

		tags := ""
		if len(s.Tags) > 0 {
			tags = fmt.Sprintf(" [%s]", joinTags(s.Tags))
		}

		// Decode directory path for display (- back to /)
		displayDir := session.DecodeDirPath(s.Directory)

		fmt.Printf("%s %-30s │ %s │ %-20s │ %d msgs%s\n",
			pin,
			truncate(s.Name, 30),
			s.Modified.Format("2006-01-02 15:04"),
			truncate(displayDir, 20),
			s.MessageCount,
			tags,
		)
	}

	return nil
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
	fmt.Println(`Claude Code Session Manager

USAGE:
    super-resume              Launch interactive TUI
    super-resume <command>    Run a command

COMMANDS:
    pin [session]     Pin a session (uses CLAUDE_SESSION_ID if not specified)
    unpin [session]   Unpin a session
    delete <session>  Delete a session
    tag <session> <tag>  Add a tag to a session
    list              List all sessions
    help              Show this help

TUI KEYBINDINGS:
    ↑/k, ↓/j    Navigate
    p           Pin/unpin session
    d           Delete session
    t           Add tag
    /           Filter sessions
    Tab         Toggle preview
    Enter       Show resume command
    ?           Toggle help
    q           Quit`)
}
