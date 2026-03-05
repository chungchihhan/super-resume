// Package tui provides the bubbletea-based terminal user interface.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/harrychung/session-manager/internal/metadata"
	"github.com/harrychung/session-manager/internal/session"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(true)

	pinnedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	previewStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			MarginLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)
)

// keyMap defines keybindings.
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Pin       key.Binding
	Delete    key.Binding
	Tag       key.Binding
	Filter    key.Binding
	Preview   key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Help      key.Binding
	Quit      key.Binding
	Confirm   key.Binding
	PageUp    key.Binding
	PageDown  key.Binding
	ToggleAll key.Binding
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
	Preview: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "toggle preview"),
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
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Pin, k.Delete, k.Filter, k.ToggleAll, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Pin, k.Delete, k.Tag},
		{k.Filter, k.Preview, k.Enter, k.ToggleAll},
		{k.Help, k.Quit, k.Escape},
	}
}

// Mode represents the current UI mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeFilter
	ModeConfirmDelete
	ModeAddTag
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
	help            help.Model
	showHelp        bool
	showPreview     bool
	showAllSessions bool // false = current dir only, true = all sessions
	statusMessage   string
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
		manager:         mgr,
		metadata:        meta,
		help:            help.New(),
		filterInput:     filterInput,
		tagInput:        tagInput,
		showPreview:     true,
		showAllSessions: false, // Default: current directory only
		currentDir:      mgr.GetCurrentDir(),
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

	case key.Matches(msg, keys.Preview):
		m.showPreview = !m.showPreview

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
	if filter == "" {
		m.filtered = m.sessions
		return
	}

	var filtered []*session.Session
	for _, s := range m.sessions {
		if strings.Contains(strings.ToLower(s.Name), filter) ||
			strings.Contains(strings.ToLower(s.ID), filter) ||
			strings.Contains(strings.ToLower(s.Directory), filter) ||
			containsTag(s.Tags, filter) {
			filtered = append(filtered, s)
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

	var b strings.Builder

	// Title with current directory
	title := "📋 Claude Code Session Manager"
	if m.showAllSessions {
		title += " (all sessions)"
	} else {
		// Show shortened current directory
		dir := m.currentDir
		if len(dir) > 40 {
			dir = "..." + dir[len(dir)-37:]
		}
		title += fmt.Sprintf(" 📁 %s", dir)
	}
	b.WriteString(titleStyle.Render(title))
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

	// Calculate list height
	listHeight := m.height - 10
	if m.showHelp {
		listHeight -= 6
	}
	if listHeight < 5 {
		listHeight = 5
	}

	// Calculate preview width
	listWidth := m.width
	previewWidth := 0
	if m.showPreview && m.width > 80 {
		listWidth = m.width * 55 / 100
		previewWidth = m.width - listWidth - 4
	}

	// Session list
	listContent := m.renderList(listHeight, listWidth)

	// Preview panel
	if m.showPreview && previewWidth > 0 && len(m.filtered) > 0 && m.cursor < len(m.filtered) {
		previewContent := m.renderPreview(listHeight, previewWidth)
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listContent, previewContent))
	} else {
		b.WriteString(listContent)
	}

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

	// Help
	b.WriteString("\n")
	if m.showHelp {
		b.WriteString(m.help.View(keys))
	} else {
		b.WriteString(helpStyle.Render("Press ? for help"))
	}

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

		// Build session line
		pin := "  "
		if s.IsPinned {
			pin = "📌"
		}

		name := truncateStr(s.Name, 30)
		dir := truncateStr(s.Directory, 20)
		date := s.Modified.Format("01/02 15:04")
		msgs := fmt.Sprintf("%3d msgs", s.MessageCount)

		line := fmt.Sprintf("%s %-30s │ %s │ %-20s │ %s",
			pin, name, date, dir, msgs)

		// Add tags
		if len(s.Tags) > 0 {
			tagStr := " "
			for _, t := range s.Tags {
				tagStr += tagStyle.Render(t) + " "
			}
			line += tagStr
		}

		// Truncate to width
		if len(line) > width-2 {
			line = line[:width-5] + "..."
		}

		// Apply style
		if i == m.cursor {
			line = selectedStyle.Render(line)
		} else if s.IsPinned {
			line = pinnedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.filtered) > height {
		b.WriteString(dimStyle.Render(fmt.Sprintf("\n(%d/%d)", m.cursor+1, len(m.filtered))))
	}

	return b.String()
}

// Style for command display
var commandStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("82")).
	Background(lipgloss.Color("236")).
	Padding(0, 1)

func (m Model) renderPreview(height, width int) string {
	if len(m.filtered) == 0 || m.cursor >= len(m.filtered) {
		return ""
	}

	s := m.filtered[m.cursor]

	var content strings.Builder
	content.WriteString(lipgloss.NewStyle().Bold(true).Render(s.Name))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render(fmt.Sprintf("ID: %s", s.ID[:min(20, len(s.ID))])))
	content.WriteString("\n")
	content.WriteString(dimStyle.Render(fmt.Sprintf("Messages: %d", s.MessageCount)))
	content.WriteString("\n")

	if len(s.Tags) > 0 {
		content.WriteString("Tags: ")
		for _, t := range s.Tags {
			content.WriteString(tagStyle.Render(t) + " ")
		}
		content.WriteString("\n")
	}

	// Resume command
	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("82")).Render("Resume Command:"))
	content.WriteString("\n")
	resumeCmd := fmt.Sprintf("claude --resume %s", s.ID)
	content.WriteString(commandStyle.Render(resumeCmd))
	content.WriteString("\n")

	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Preview:"))
	content.WriteString("\n")

	// Get preview - reduce lines to make room for command
	preview, err := m.manager.GetPreview(s.ID, height-12)
	if err == nil {
		for _, line := range preview {
			wrapped := wrapText(line, width-4)
			content.WriteString(dimStyle.Render(wrapped))
			content.WriteString("\n")
		}
	}

	return previewStyle.Width(width).Height(height).Render(content.String())
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func wrapText(s string, width int) string {
	if len(s) <= width {
		return s
	}
	return s[:width-3] + "..."
}
