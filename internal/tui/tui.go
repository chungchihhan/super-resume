// Package tui provides the bubbletea-based terminal user interface.
package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harrychung/super-resume/internal/metadata"
	"github.com/harrychung/super-resume/internal/session"
)

// Styles - using AdaptiveColor for light/dark background support
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

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
			Padding(0, 1).
			Bold(true)

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
	RemoveTag    key.Binding
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
		key.WithKeys("up", "k", "K"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j", "J"),
		key.WithHelp("↓/j", "down"),
	),
	Pin: key.NewBinding(
		key.WithKeys("p", "P"),
		key.WithHelp("p", "pin/unpin"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "D", "x", "X"),
		key.WithHelp("d", "delete"),
	),
	Tag: key.NewBinding(
		key.WithKeys("t", "T"),
		key.WithHelp("t", "add tag"),
	),
	RemoveTag: key.NewBinding(
		key.WithKeys("u", "U"),
		key.WithHelp("u", "manage tags"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l", "L"),
		key.WithHelp("→", "preview"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h", "H"),
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
		key.WithKeys("q", "Q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("y", "Y"),
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
		key.WithKeys("a", "A"),
		key.WithHelp("a", "all/current dir"),
	),
	ToggleAgents: key.NewBinding(
		key.WithKeys("s", "S"),
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
	ModeRemoveTag
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
	previewScroll    int                      // Scroll offset in preview mode
	previewCursor    int                      // Current highlighted message in preview
	previewCache     []session.PreviewMessage // Cached preview messages
	previewSessionID string                   // ID of cached preview session
	selectedSession *session.Session
	currentDir      string
	tagCursor       int // For selecting tag to remove

	// Resume info - set when user selects a session with Enter
	ResumeSessionID    string // Session ID to resume
	ResumeDirectory    string // Directory where the session was created (encoded)
	ResumeMessageIndex int    // Message index (0 = newest/bottom, >0 = specific message from preview)
}

// GetResumeInfo returns the session ID, directory, and message index if a session was selected for resume.
func (m Model) GetResumeInfo() (sessionID string, directory string, messageIndex int) {
	return m.ResumeSessionID, m.ResumeDirectory, m.ResumeMessageIndex
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

	case ModeRemoveTag:
		if m.selectedSession == nil || len(m.selectedSession.Tags) == 0 {
			m.mode = ModeNormal
			return m, nil
		}
		switch {
		case key.Matches(msg, keys.Escape):
			m.mode = ModeNormal
			m.selectedSession = nil
			m.tagCursor = 0
		case key.Matches(msg, keys.Left), key.Matches(msg, keys.Up):
			if m.tagCursor > 0 {
				m.tagCursor--
			}
		case key.Matches(msg, keys.Right), key.Matches(msg, keys.Down):
			if m.tagCursor < len(m.selectedSession.Tags)-1 {
				m.tagCursor++
			}
		case key.Matches(msg, keys.Delete): // D to delete
			tag := m.selectedSession.Tags[m.tagCursor]
			if err := m.metadata.RemoveTag(m.selectedSession.ID, tag); err != nil {
				m.statusMessage = "Error removing tag: " + err.Error()
			} else {
				m.statusMessage = "Removed tag: " + tag
			}
			m.mode = ModeNormal
			m.selectedSession = nil
			m.tagCursor = 0
			return m, m.loadSessions
		case key.Matches(msg, keys.Enter): // Enter to edit
			tag := m.selectedSession.Tags[m.tagCursor]
			// Switch to add tag mode with the current tag value for editing
			m.tagInput.SetValue(tag)
			// Remove the old tag first, then add the new one when done
			if err := m.metadata.RemoveTag(m.selectedSession.ID, tag); err != nil {
				m.statusMessage = "Error: " + err.Error()
				m.mode = ModeNormal
				m.selectedSession = nil
				m.tagCursor = 0
				return m, nil
			}
			m.mode = ModeAddTag
			m.tagInput.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case ModePreview:
		switch msg.String() {
		case "esc", "left", "h", "q", "backspace":
			m.mode = ModeNormal
			m.previewScroll = 0
			m.previewCursor = 0
			m.previewCache = nil
			m.previewSessionID = ""
			return m, nil
		case "enter":
			if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
				s := m.filtered[m.cursor]
				m.ResumeSessionID = s.ID
				m.ResumeDirectory = s.Cwd // Use actual cwd from session file
				// Store message position from bottom (previewCursor is from top)
				// len-1-previewCursor gives position from bottom, +1 for 1-indexed
				m.ResumeMessageIndex = m.previewCursor + 1
				return m, tea.Quit
			}
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.previewCursor > 0 {
				m.previewCursor--
				// Scroll up if cursor goes above visible area
				if m.previewCursor < m.previewScroll {
					m.previewScroll = m.previewCursor
				}
			}
			return m, nil
		case "down", "j":
			if m.previewCursor < len(m.previewCache)-1 {
				m.previewCursor++
				// Scroll down if cursor goes below visible area
				previewLines := m.height - 12
				if previewLines < 5 {
					previewLines = 5
				}
				if m.previewCursor >= m.previewScroll+previewLines {
					m.previewScroll = m.previewCursor - previewLines + 1
				}
			}
			return m, nil
		default:
			// Ignore other keys but don't get stuck
			return m, nil
		}
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

	case key.Matches(msg, keys.RemoveTag):
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			s := m.filtered[m.cursor]
			if len(s.Tags) > 0 {
				m.selectedSession = s
				m.tagCursor = 0
				m.mode = ModeRemoveTag
			} else {
				m.statusMessage = "No tags to remove"
			}
		}

	case key.Matches(msg, keys.Filter):
		m.mode = ModeFilter
		m.filterInput.Focus()
		return m, textinput.Blink

	case key.Matches(msg, keys.Right):
		if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
			s := m.filtered[m.cursor]
			// Load preview cache
			preview, _ := m.manager.GetPreview(s.ID, 100)
			m.previewCache = preview
			m.previewSessionID = s.ID
			m.previewScroll = 0
			m.previewCursor = 0
			m.mode = ModePreview
		}

	case key.Matches(msg, keys.Help):
		m.showHelp = !m.showHelp

	case key.Matches(msg, keys.ToggleAll):
		m.showAllSessions = !m.showAllSessions
		m.statusMessage = "" // Clear status, title shows the mode
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
				m.statusMessage = "Hidden agents"
			} else {
				// Check if this session has any agents first
				hasAgents := false
				for _, sess := range m.sessions {
					if sess.IsAgent && sess.ParentSessionID == s.ID {
						hasAgents = true
						break
					}
				}
				if hasAgents {
					m.expandedSessions[s.ID] = true
					m.statusMessage = "Showing agents"
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
			m.ResumeSessionID = s.ID
			m.ResumeDirectory = s.Cwd // Use actual cwd from session file
			m.ResumeMessageIndex = 0  // 0 = newest message (bottom)
			return m, tea.Quit
		}

	case key.Matches(msg, keys.Escape):
		// If filter is active, clear it; otherwise quit
		if m.filterInput.Value() != "" {
			m.filterInput.SetValue("")
			m.applyFilter()
			m.statusMessage = ""
		} else {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *Model) applyFilter() {
	filter := strings.ToLower(m.filterInput.Value())

	// Separate parent sessions and agents
	var parentSessions []*session.Session
	agentsByParent := make(map[string][]*session.Session)

	for _, s := range m.sessions {
		if s.IsAgent {
			agentsByParent[s.ParentSessionID] = append(agentsByParent[s.ParentSessionID], s)
		} else {
			parentSessions = append(parentSessions, s)
		}
	}

	// Build filtered list: parent sessions with their agents inserted directly after
	var filtered []*session.Session
	for _, s := range parentSessions {
		// Apply text filter to parent session
		if filter != "" {
			if !strings.Contains(strings.ToLower(s.Name), filter) &&
				!strings.Contains(strings.ToLower(s.ID), filter) &&
				!strings.Contains(strings.ToLower(s.Directory), filter) &&
				!containsTag(s.Tags, filter) {
				continue
			}
		}

		filtered = append(filtered, s)

		// Add agents right after their parent if expanded
		if m.expandedSessions[s.ID] {
			for _, agent := range agentsByParent[s.ID] {
				filtered = append(filtered, agent)
			}
		}
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

	// Title with scope indicator
	if m.showAllSessions {
		b.WriteString(titleStyle.Render("Sessions") + dimStyle.Render(" · All"))
	} else {
		// Shorten path: replace home dir with ~
		displayDir := m.currentDir
		if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(displayDir, home) {
			displayDir = "~" + strings.TrimPrefix(displayDir, home)
		}
		b.WriteString(titleStyle.Render("Sessions") + dimStyle.Render(" · "+displayDir))
	}
	b.WriteString("\n\n")

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
			Render("Delete this session? (y/n)"))
	}

	if m.mode == ModeAddTag && m.selectedSession != nil {
		b.WriteString("\n")
		b.WriteString("Add tag: ")
		b.WriteString(m.tagInput.View())
	}

	if m.mode == ModeRemoveTag && m.selectedSession != nil && len(m.selectedSession.Tags) > 0 {
		b.WriteString("\n")
		b.WriteString("Manage tags: ")
		for i, tag := range m.selectedSession.Tags {
			if i == m.tagCursor {
				b.WriteString(selectedStyle.Render("[" + tag + "]"))
			} else {
				b.WriteString(tagStyle.Render(tag))
			}
			b.WriteString(" ")
		}
		b.WriteString(dimStyle.Render("(←/→ select, D delete, Enter edit, Esc cancel)"))
	}

	// Help bar
	b.WriteString("\n")
	helpItems := []string{
		"↑↓ navigate",
		"→ preview",
		"A all/current",
		"P pin",
		"T tag",
		"U untag",
		"D delete",
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

	// Badge styles - match tagStyle format
	pinnedBadgeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "208", Dark: "214"}).
		Background(lipgloss.AdaptiveColor{Light: "254", Dark: "236"}).
		Padding(0, 1).
		Bold(true)

	for i := start; i < end; i++ {
		s := m.filtered[i]

		// Indent for agent sessions
		indent := ""
		metaIndent := "  "
		if s.IsAgent {
			indent = "    "
			metaIndent = "      "
		}

		// Build badges line (pinned + tags)
		var badges []string
		if s.IsPinned {
			badges = append(badges, pinnedBadgeStyle.Render("Pinned"))
		}
		for _, tag := range s.Tags {
			badges = append(badges, tagStyle.Render(tag))
		}
		if len(badges) > 0 {
			b.WriteString(metaIndent + strings.Join(badges, " "))
			b.WriteString("\n")
		}

		// Build session line - Claude Code style
		cursor := "  "
		if i == m.cursor {
			cursor = "▶ "
		}

		// Session summary (first part of name, truncated to fit screen)
		summary := truncateStr(s.Name, width-4-len(indent)) // 4 = cursor(2) + right margin(2)

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

		// Meta line: time · messages · path
		relTime := relativeTime(s.Modified)
		displayPath := s.Cwd
		if home, err := os.UserHomeDir(); err == nil && strings.HasPrefix(displayPath, home) {
			displayPath = "~" + strings.TrimPrefix(displayPath, home)
		}
		meta := fmt.Sprintf("%s · %d msgs · %s", relTime, s.MessageCount, displayPath)

		b.WriteString(metaIndent + dimStyle.Render(meta))
		b.WriteString("\n\n")
	}

	return b.String()
}

func (m Model) renderFullPreview() string {
	s := m.filtered[m.cursor]

	var b strings.Builder
	padding := "  " // Left padding

	// Header with session name
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "240", Dark: "60"}).
		Width(m.width - 6).
		PaddingBottom(1)

	b.WriteString(padding + headerStyle.Render(s.Name))
	b.WriteString("\n\n")

	// Directory path
	displayDir := session.DecodeDirPath(s.Directory)
	b.WriteString(padding + dimStyle.Render(displayDir))
	b.WriteString("\n")

	// Tags
	if len(s.Tags) > 0 {
		b.WriteString(padding)
		for _, t := range s.Tags {
			b.WriteString(tagStyle.Render(t) + " ")
		}
		b.WriteString("\n")
	}

	// Conversation preview section
	b.WriteString("\n")

	previewLines := m.height - 12
	if previewLines < 5 {
		previewLines = 5
	}

	// Styles for different message types - Claude Code style
	userStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "255"}).
		Bold(true)
	assistantStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "55", Dark: "141"})
	summaryStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "245"})
	bulletStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "55", Dark: "141"})

	// Use cached preview data
	if len(m.previewCache) > 0 {
		// Calculate visible slice based on scroll
		start := m.previewScroll
		if start >= len(m.previewCache) {
			start = len(m.previewCache) - 1
		}
		if start < 0 {
			start = 0
		}
		end := start + previewLines
		if end > len(m.previewCache) {
			end = len(m.previewCache)
		}

		// Highlight background color
		highlightBg := lipgloss.AdaptiveColor{Light: "254", Dark: "236"}
		maxWidth := m.width - 8 // Leave padding on both sides

		for i, msg := range m.previewCache[start:end] {
			actualIndex := start + i
			isSelected := actualIndex == m.previewCursor

			// Build the line based on role
			var prefix, text string
			var prefixStyle, textStyle lipgloss.Style

			switch msg.Role {
			case "user":
				prefix = "> "
				text = msg.Text
				prefixStyle = userStyle
				textStyle = userStyle
			case "assistant":
				prefix = "⏺ "
				text = msg.Text
				prefixStyle = bulletStyle
				textStyle = assistantStyle
			case "summary":
				prefix = "  ⎿ "
				text = msg.Text
				prefixStyle = summaryStyle
				textStyle = summaryStyle
			default:
				prefix = ""
				text = msg.Text
				prefixStyle = dimStyle
				textStyle = dimStyle
			}

			// Truncate text to fit within maxWidth (leaving room for prefix and right padding)
			textMaxLen := maxWidth - len(prefix) - 1 // 1 char right padding
			if len(text) > textMaxLen {
				text = text[:textMaxLen-1] + "..."
			}

			// Render the line
			if isSelected {
				// Create highlighted versions of the styles
				highlightedPrefix := prefixStyle.Background(highlightBg)
				highlightedText := textStyle.Background(highlightBg)

				line := highlightedPrefix.Render(prefix) + highlightedText.Render(text)
				// Pad to full width with background
				lineLen := len(prefix) + len(text)
				if lineLen < maxWidth {
					padStyle := lipgloss.NewStyle().Background(highlightBg)
					line += padStyle.Render(strings.Repeat(" ", maxWidth-lineLen))
				}
				b.WriteString(padding + line)
			} else {
				line := prefixStyle.Render(prefix) + textStyle.Render(text)
				b.WriteString(padding + line)
			}

			if msg.Role == "user" || msg.Role == "assistant" {
				b.WriteString("\n\n")
			} else {
				b.WriteString("\n")
			}
		}
	} else {
		b.WriteString(padding + dimStyle.Render("No preview available"))
		b.WriteString("\n")
	}

	// Scroll indicator
	b.WriteString("\n")
	scrollStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "25", Dark: "75"})
	if len(m.previewCache) > previewLines {
		end := m.previewScroll + previewLines
		if end > len(m.previewCache) {
			end = len(m.previewCache)
		}
		scrollText := fmt.Sprintf("▲ %d-%d of %d ▼", m.previewScroll+1, end, len(m.previewCache))
		b.WriteString(padding + scrollStyle.Render(scrollText))
		b.WriteString("\n")
	}

	// Metadata at bottom
	metaStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "245"})
	metaItems := []string{
		fmt.Sprintf("%d messages", s.MessageCount),
		relativeTime(s.Modified),
	}
	if s.IsAgent {
		metaItems = append(metaItems, "agent")
	}
	b.WriteString(padding + metaStyle.Render(strings.Join(metaItems, " · ")))

	// Help bar at bottom
	b.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "243"})
	b.WriteString(padding + helpStyle.Render("← back  ·  ↑↓ scroll  ·  Enter resume"))

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

// wrapText wraps text to fit within maxWidth characters
func wrapText(s string, maxWidth int) string {
	if maxWidth <= 0 || len(s) <= maxWidth {
		return s
	}

	var result strings.Builder
	words := strings.Fields(s)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)
		if lineLen+wordLen+1 > maxWidth && lineLen > 0 {
			result.WriteString("\n")
			lineLen = 0
		}
		if lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}
		result.WriteString(word)
		lineLen += wordLen

		// Safety check for very long words
		if i < len(words)-1 && lineLen >= maxWidth {
			result.WriteString("\n")
			lineLen = 0
		}
	}
	return result.String()
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
