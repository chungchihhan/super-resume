// Package tui provides the bubbletea-based terminal user interface.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harrychung/session-manager/internal/metadata"
	"github.com/harrychung/session-manager/internal/session"
)

// Styles - using AdaptiveColor for light/dark background support
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "229"}).
			Background(lipgloss.AdaptiveColor{Light: "153", Dark: "57"}).
			Bold(true)

	pinnedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "130", Dark: "220"})

	// Normal text: black on light bg, white on dark bg
	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "255"})

	// Dim text: dark gray on light bg, light gray on dark bg
	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "250"})

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.AdaptiveColor{Light: "254", Dark: "236"}).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "250"})

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
)

// keyMap defines keybindings.
type keyMap struct {
	Up           key.Binding
	Down         key.Binding
	Pin          key.Binding
	Delete       key.Binding
	Tag          key.Binding
	Filter       key.Binding
	Right        key.Binding
	Left         key.Binding
	Enter        key.Binding
	Escape       key.Binding
	Help         key.Binding
	Quit         key.Binding
	Confirm      key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	ToggleAll    key.Binding
	ToggleAgents key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Pin: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pin/unpin"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "x"),
		key.WithHelp("d", "delete"),
	),
	Tag: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "add tag"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→", "preview"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←", "back"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+u"),
		key.WithHelp("pgup", "page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+d"),
		key.WithHelp("pgdn", "page down"),
	),
	ToggleAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "all/current dir"),
	),
	ToggleAgents: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "show/hide agents"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Pin, k.Delete, k.Filter, k.ToggleAll, k.ToggleAgents, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Pin, k.Delete, k.Tag},
		{k.Filter, k.Right, k.Enter},
		{k.ToggleAll, k.ToggleAgents, k.Help, k.Quit},
	}
}

// Mode represents the current UI mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeFilter
	ModeConfirmDelete
	ModeAddTag
	ModePreview
)

// Model is the bubbletea model.
type Model struct {
	sessions        []*session.Session
	filtered        []*session.Session
	cursor          int
	manager         *session.Manager
	metadata        *metadata.Store
	width           int
	height          int
	mode            Mode
	filterInput     textinput.Model
	tagInput        textinput.Model
	help             help.Model
	showHelp         bool
	showAllSessions  bool            // false = current dir only, true = all sessions
	expandedSessions map[string]bool // Set of session IDs with agents expanded
	statusMessage    string
	selectedSession *session.Session
	currentDir      string
}

// New creates a new TUI model.
func New(mgr *session.Manager, meta *metadata.Store) Model {
	filterInput := textinput.New()
	filterInput.Placeholder = "Type to filter..."
	filterInput.CharLimit = 50

	tagInput := textinput.New()
	tagInput.Placeholder = "Enter tag name..."
	tagInput.CharLimit = 30

	return Model{
		manager:          mgr,
		metadata:         meta,
		help:             help.New(),
		filterInput:      filterInput,
		tagInput:         tagInput,
		showAllSessions:  false,              // Default: current directory only
		expandedSessions: make(map[string]bool),
		currentDir:       mgr.GetCurrentDir(),
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return m.loadSessions
}

func (m Model) loadSessions() tea.Msg {
	var sessions []*session.Session
	var err error

	if m.showAllSessions {
		sessions, err = m.manager.List()
	} else {
		sessions, err = m.manager.ListForCurrentDir()
	}

	if err != nil {
		return errMsg{err}
	}
	return sessionsLoadedMsg{sessions}
}

type sessionsLoadedMsg struct {
	sessions []*session.Session
}

type errMsg struct {
	err error
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case sessionsLoadedMsg:
		m.sessions = msg.sessions
		m.applyFilter()
		return m, nil

	case errMsg:
		m.statusMessage = "Error: " + msg.err.Error()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Handle text input updates
	if m.mode == ModeFilter {
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		m.applyFilter()
		return m, cmd
	}

	if m.mode == ModeAddTag {
		var cmd tea.Cmd
		m.tagInput, cmd = m.tagInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeFilter:
		switch {
		case key.Matches(msg, keys.Escape):
			m.mode = ModeNormal
			m.filterInput.SetValue("")
			m.applyFilter()
		case key.Matches(msg, keys.Enter):
			m.mode = ModeNormal
			m.filterInput.Blur()
		default:
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			m.applyFilter()
			return m, cmd
		}
		return m, nil

	case ModeConfirmDelete:
		switch {
		case key.Matches(msg, keys.Confirm):
			if m.selectedSession != nil {
				if err := m.manager.Delete(m.selectedSession.ID); err != nil {
					m.statusMessage = "Error deleting: " + err.Error()
				} else {
					m.statusMessage = "Deleted: " + m.selectedSession.Name
				}
				m.selectedSession = nil
			}
			m.mode = ModeNormal
			return m, m.loadSessions
		case key.Matches(msg, keys.Escape), key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
			m.mode = ModeNormal
			m.selectedSession = nil
			m.statusMessage = "Delete cancelled"
		}
		return m, nil

	case ModeAddTag:
		switch {
		case key.Matches(msg, keys.Escape):
			m.mode = ModeNormal
			m.tagInput.SetValue("")
			m.selectedSession = nil
		case key.Matches(msg, keys.Enter):
			if m.selectedSession != nil && m.tagInput.Value() != "" {
				if err := m.metadata.AddTag(m.selectedSession.ID, m.tagInput.Value()); err != nil {
					m.statusMessage = "Error adding tag: " + err.Error()
				} else {
					m.statusMessage = "Added tag: " + m.tagInput.Value()
				}
			}
			m.mode = ModeNormal
			m.tagInput.SetValue("")
			m.selectedSession = nil
			return m, m.loadSessions
		default:
			var cmd tea.Cmd
			m.tagInput, cmd = m.tagInput.Update(msg)
			return m, cmd
		}
		return m, nil

	case ModePreview:
		switch {
		case key.Matches(msg, keys.Escape), key.Matches(msg, keys.Left):
			m.mode = ModeNormal
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Enter):
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				s := m.filtered[m.cursor]
				fmt.Printf("\n\nTo resume this session:\n  claude --resume %s\n\n", s.ID)
				return m, tea.Quit
			}
		}
		return m, nil
	}

	// Normal mode
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}

	case key.Matches(msg, keys.PageUp):
		m.cursor -= 10
		if m.cursor < 0 {
			m.cursor = 0
		}

	case key.Matches(msg, keys.PageDown):
		m.cursor += 10
		if m.cursor >= len(m.filtered) {
			m.cursor = len(m.filtered) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}

	case key.Matches(msg, keys.Pin):
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			s := m.filtered[m.cursor]
			if err := m.metadata.TogglePin(s.ID); err != nil {
				m.statusMessage = "Error: " + err.Error()
			} else {
				if s.IsPinned {
					m.statusMessage = "Unpinned: " + s.Name
				} else {
					m.statusMessage = "Pinned: " + s.Name
				}
			}
			return m, m.loadSessions
		}

	case key.Matches(msg, keys.Delete):
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			m.selectedSession = m.filtered[m.cursor]
			m.mode = ModeConfirmDelete
		}

	case key.Matches(msg, keys.Tag):
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			m.selectedSession = m.filtered[m.cursor]
			m.mode = ModeAddTag
			m.tagInput.Focus()
		}

	case key.Matches(msg, keys.Filter):
		m.mode = ModeFilter
		m.filterInput.Focus()
		return m, textinput.Blink

	case key.Matches(msg, keys.Right):
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			m.mode = ModePreview
		}

	case key.Matches(msg, keys.Help):
		m.showHelp = !m.showHelp

	case key.Matches(msg, keys.ToggleAll):
		m.showAllSessions = !m.showAllSessions
		if m.showAllSessions {
			m.statusMessage = "Showing all sessions"
		} else {
			m.statusMessage = "Showing current directory only"
		}
		return m, m.loadSessions

	case key.Matches(msg, keys.ToggleAgents):
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			s := m.filtered[m.cursor]
			if s.IsAgent {
				// Can't expand agents of an agent
				return m, nil
			}
			// Toggle this session's agents
			if m.expandedSessions[s.ID] {
				delete(m.expandedSessions, s.ID)
				m.statusMessage = "Hidden agents for: " + s.Name
			} else {
				// Check if this session has any agents first
				hasAgents := false
				for _, sess := range m.sessions {
					if sess.IsAgent && sess.Directory == s.Directory {
						hasAgents = true
						break
					}
				}
				if hasAgents {
					m.expandedSessions[s.ID] = true
					m.statusMessage = "Showing agents for: " + s.Name
				} else {
					m.statusMessage = "No agents for this session"
				}
			}
			m.applyFilter()
		}
		return m, nil

	case key.Matches(msg, keys.Enter):
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			s := m.filtered[m.cursor]
			// Return with session info for resume
			fmt.Printf("\n\nTo resume this session:\n  claude --resume %s\n\n", s.ID)
			return m, tea.Quit
		}

	case key.Matches(msg, keys.Escape):
		m.filterInput.SetValue("")
		m.applyFilter()
		m.statusMessage = ""
	}

	return m, nil
}

func (m *Model) applyFilter() {
	filter := strings.ToLower(m.filterInput.Value())

	// Build set of directories that have expanded sessions
	expandedDirs := make(map[string]bool)
	for sessionID := range m.expandedSessions {
		for _, s := range m.sessions {
			if s.ID == sessionID {
				expandedDirs[s.Directory] = true
				break
			}
		}
	}

	var filtered []*session.Session
	for _, s := range m.sessions {
		// Handle agent sessions - only show if parent session is expanded
		if s.IsAgent {
			if !expandedDirs[s.Directory] {
				continue
			}
		}

		// Apply text filter
		if filter != "" {
			if !strings.Contains(strings.ToLower(s.Name), filter) &&
				!strings.Contains(strings.ToLower(s.ID), filter) &&
				!strings.Contains(strings.ToLower(s.Directory), filter) &&
				!containsTag(s.Tags, filter) {
				continue
			}
		}

		filtered = append(filtered, s)
	}
	m.filtered = filtered

	// Adjust cursor if needed
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func containsTag(tags []string, filter string) bool {
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), filter) {
			return true
		}
	}
	return false
}

// View renders the UI.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Full-screen preview mode
	if m.mode == ModePreview && len(m.filtered) > 0 && m.cursor < len(m.filtered) {
		return m.renderFullPreview()
	}

	var b strings.Builder

	// Simple title
	b.WriteString(titleStyle.Render("Sessions"))
	if !m.showAllSessions {
		b.WriteString(dimStyle.Render(fmt.Sprintf(" in %s", m.currentDir)))
	}
	b.WriteString("\n")

	// Filter line
	if m.mode == ModeFilter {
		b.WriteString("Filter: ")
		b.WriteString(m.filterInput.View())
		b.WriteString("\n")
	} else if m.filterInput.Value() != "" {
		b.WriteString(dimStyle.Render(fmt.Sprintf("Filter: %s (press / to change, esc to clear)", m.filterInput.Value())))
		b.WriteString("\n")
	}

	// Status message
	if m.statusMessage != "" {
		b.WriteString(statusStyle.Render(m.statusMessage))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Calculate list height (each session takes ~3 lines)
	listHeight := (m.height - 8) / 3
	if listHeight < 3 {
		listHeight = 3
	}

	// Session list
	b.WriteString(m.renderList(listHeight, m.width))
	b.WriteString("\n")

	// Modal overlays
	if m.mode == ModeConfirmDelete && m.selectedSession != nil {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Render(fmt.Sprintf("Delete '%s'? (y/n)", m.selectedSession.Name)))
	}

	if m.mode == ModeAddTag && m.selectedSession != nil {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Add tag to '%s': ", m.selectedSession.Name))
		b.WriteString(m.tagInput.View())
	}

	// Help bar
	b.WriteString("\n")
	helpItems := []string{
		"↑↓ navigate",
		"→ preview",
		"A all/current",
		"S agents",
		"P pin",
		"D delete",
		"/ filter",
	}
	b.WriteString(dimStyle.Render(strings.Join(helpItems, " · ")))

	return b.String()
}

func (m Model) renderList(height, width int) string {
	var b strings.Builder

	if len(m.filtered) == 0 {
		return dimStyle.Render("No sessions found")
	}

	// Calculate visible range
	start := 0
	if m.cursor >= height {
		start = m.cursor - height + 1
	}
	end := start + height
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		s := m.filtered[i]

		// Indent for agent sessions
		indent := ""
		metaIndent := "  "
		if s.IsAgent {
			indent = "    "
			metaIndent = "      "
		}

		// Build session line - Claude Code style
		cursor := "  "
		if i == m.cursor {
			cursor = "▶ "
		}

		// Session summary (first part of name, truncated)
		summary := truncateStr(s.Name, width-20-len(indent))

		// Relative time
		relTime := relativeTime(s.Modified)

		// Message count
		meta := fmt.Sprintf("%s · %d messages", relTime, s.MessageCount)

		// Apply style - consistent alignment
		if i == m.cursor {
			b.WriteString(indent + selectedStyle.Render(cursor+summary))
		} else if s.IsPinned {
			b.WriteString(indent + pinnedStyle.Render(cursor+summary))
		} else if s.IsAgent {
			b.WriteString(indent + dimStyle.Render(cursor+summary))
		} else {
			b.WriteString(indent + normalStyle.Render(cursor+summary))
		}
		b.WriteString("\n")
		b.WriteString(metaIndent + dimStyle.Render(meta))
		b.WriteString("\n\n")
	}

	return b.String()
}

func (m Model) renderFullPreview() string {
	s := m.filtered[m.cursor]

	var b strings.Builder

	// Title bar
	b.WriteString(titleStyle.Render("Session Preview"))
	b.WriteString("\n\n")

	// Session name
	b.WriteString(lipgloss.NewStyle().Bold(true).Render(s.Name))
	b.WriteString("\n\n")

	// Metadata
	b.WriteString(dimStyle.Render(fmt.Sprintf("ID: %s", s.ID)))
	b.WriteString("\n")

	displayDir := session.DecodeDirPath(s.Directory)
	b.WriteString(dimStyle.Render(fmt.Sprintf("Path: %s", displayDir)))
	b.WriteString("\n")

	b.WriteString(dimStyle.Render(fmt.Sprintf("Messages: %d", s.MessageCount)))
	b.WriteString("\n")

	b.WriteString(dimStyle.Render(fmt.Sprintf("Modified: %s", relativeTime(s.Modified))))
	b.WriteString("\n")

	if s.IsAgent {
		b.WriteString(dimStyle.Render("Type: Agent sub-session"))
		b.WriteString("\n")
	}

	if len(s.Tags) > 0 {
		b.WriteString("Tags: ")
		for _, t := range s.Tags {
			b.WriteString(tagStyle.Render(t) + " ")
		}
		b.WriteString("\n")
	}

	// Resume command
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")).Render("Resume Command:"))
	b.WriteString("\n")
	resumeCmd := fmt.Sprintf("claude --resume %s", s.ID)
	b.WriteString(commandStyle.Render(resumeCmd))
	b.WriteString("\n")

	// Preview content
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("Conversation Preview:"))
	b.WriteString("\n\n")

	previewLines := m.height - 18
	if previewLines < 5 {
		previewLines = 5
	}

	preview, err := m.manager.GetPreview(s.ID, previewLines)
	if err == nil {
		for _, line := range preview {
			b.WriteString(dimStyle.Render(line))
			b.WriteString("\n")
		}
	}

	// Help bar
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("← back · Enter to resume · Esc to cancel"))

	return b.String()
}

// Style for command display
var commandStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("82")).
	Background(lipgloss.AdaptiveColor{Light: "254", Dark: "236"}).
	Padding(0, 1)

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// relativeTime returns a human-readable relative time string
func relativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	default:
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
}
