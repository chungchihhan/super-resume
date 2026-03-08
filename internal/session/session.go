// Package session handles Claude Code session discovery and parsing.
package session

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/harrychung/super-resume/internal/metadata"
)

// Session represents a Claude Code session.
type Session struct {
	ID              string
	Path            string
	Directory       string // Encoded directory name (folder name in .claude/projects)
	Cwd             string // Actual working directory path from session file
	Name            string
	Created         time.Time
	Modified        time.Time
	MessageCount    int
	IsPinned        bool
	IsAgent         bool             // True if this is a sub-agent session
	ParentSessionID string           // For agents: the parent session ID
	Tags            []string
	Preview         []PreviewMessage // First few messages for preview
}

// Message represents a single message in the transcript.
type Message struct {
	Type    string `json:"type"`
	IsMeta  bool   `json:"isMeta"`
	Message struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	} `json:"message"`
	Summary string `json:"summary"`
}

// PreviewMessage represents a message for preview display.
type PreviewMessage struct {
	Role string // "user", "assistant", or "summary"
	Text string
}

// Manager handles session operations.
type Manager struct {
	sessionsDir string
	currentDir  string // Current working directory for filtering
	metadata    *metadata.Store
}

// NewManager creates a new session manager.
func NewManager(meta *metadata.Store) (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cwd, _ := os.Getwd()

	return &Manager{
		sessionsDir: filepath.Join(home, ".claude", "projects"),
		currentDir:  cwd,
		metadata:    meta,
	}, nil
}

// SetCurrentDir sets the directory to filter sessions by.
func (m *Manager) SetCurrentDir(dir string) {
	m.currentDir = dir
}

// GetCurrentDir returns the current directory.
func (m *Manager) GetCurrentDir() string {
	return m.currentDir
}

// encodeDirPath encodes a directory path the way Claude Code does it.
func encodeDirPath(path string) string {
	// Claude Code replaces / with - in directory names
	return strings.ReplaceAll(path, "/", "-")
}

// DecodeDirPath decodes a Claude Code directory path back to normal format.
func DecodeDirPath(encoded string) string {
	// Replace - with / to restore the original path
	return strings.ReplaceAll(encoded, "-", "/")
}

// ListForCurrentDir returns sessions for the current directory only.
func (m *Manager) ListForCurrentDir() ([]*Session, error) {
	all, err := m.List()
	if err != nil {
		return nil, err
	}

	if m.currentDir == "" {
		return all, nil
	}

	var filtered []*Session
	for _, s := range all {
		// Compare using actual Cwd from session file (more reliable than encoded paths)
		if s.Cwd != "" {
			if s.Cwd == m.currentDir || strings.HasPrefix(s.Cwd, m.currentDir+"/") {
				filtered = append(filtered, s)
			}
		}
	}

	return filtered, nil
}

// List returns all sessions, sorted by pinned status then modified time.
func (m *Manager) List() ([]*Session, error) {
	var sessions []*Session

	if _, err := os.Stat(m.sessionsDir); os.IsNotExist(err) {
		return sessions, nil
	}

	err := filepath.Walk(m.sessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return nil
		}

		session, err := m.parseSession(path, info)
		if err != nil {
			return nil // Skip unparseable sessions
		}

		sessions = append(sessions, session)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort: pinned first, then by modified time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].IsPinned != sessions[j].IsPinned {
			return sessions[i].IsPinned
		}
		return sessions[i].Modified.After(sessions[j].Modified)
	})

	return sessions, nil
}

func (m *Manager) parseSession(path string, info os.FileInfo) (*Session, error) {
	sessionID := strings.TrimSuffix(filepath.Base(path), ".jsonl")

	// Get the project directory (parent of the session file)
	dir := filepath.Dir(path)
	relDir, err := filepath.Rel(m.sessionsDir, dir)
	if err != nil {
		relDir = dir
	}

	session := &Session{
		ID:        sessionID,
		Path:      path,
		Directory: relDir,
		Name:      sessionID[:min(8, len(sessionID))], // Default name: short ID
		Created:   info.ModTime(),                     // Approximate
		Modified:  info.ModTime(),
		IsPinned:  m.metadata.IsPinned(sessionID),
		IsAgent:   strings.HasPrefix(sessionID, "agent-"),
		Tags:      m.metadata.GetTags(sessionID),
	}

	// Parse the file for more details
	file, err := os.Open(path)
	if err != nil {
		return session, nil // Return with defaults
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for long lines

	lineCount := 0
	previewLines := 3
	var firstUserMessage string

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()

		// Parse cwd and sessionId from any line (keep looking until we find them)
		if session.Cwd == "" || (session.IsAgent && session.ParentSessionID == "") {
			var meta struct {
				Cwd       string `json:"cwd"`
				SessionID string `json:"sessionId"`
			}
			if err := json.Unmarshal([]byte(line), &meta); err == nil {
				if meta.Cwd != "" && session.Cwd == "" {
					session.Cwd = meta.Cwd
				}
				if meta.SessionID != "" && session.ParentSessionID == "" {
					session.ParentSessionID = meta.SessionID
				}
			}
		}

		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err == nil {
			// Skip meta messages (system caveat messages)
			if msg.IsMeta {
				continue
			}

			// Get message text
			text := extractMessageText(msg)

			// Skip internal command messages (XML-tagged commands)
			if strings.HasPrefix(text, "<command-") ||
				strings.HasPrefix(text, "<local-command-") ||
				strings.HasPrefix(text, "<") && strings.Contains(text, "</") {
				continue
			}

			// Capture the first user message for session name
			if firstUserMessage == "" && msg.Message.Role == "user" {
				if text != "" {
					firstUserMessage = text
				}
			}

			// Collect preview messages (use already extracted text)
			if len(session.Preview) < previewLines && text != "" {
				role := msg.Message.Role
				if msg.Summary != "" {
					role = "summary"
					text = msg.Summary
				}
				if role != "" {
					session.Preview = append(session.Preview, PreviewMessage{Role: role, Text: text})
				}
			}
		}
	}

	// Use first user message as session name
	if firstUserMessage != "" {
		session.Name = truncate(firstUserMessage, 150)
	} else {
		session.Name = "<Empty session>"
	}

	session.MessageCount = lineCount
	return session, nil
}

func extractMessageText(msg Message) string {
	if msg.Message.Content == nil {
		return ""
	}

	switch content := msg.Message.Content.(type) {
	case string:
		return content
	case []any:
		for _, block := range content {
			if m, ok := block.(map[string]any); ok {
				if text, ok := m["text"].(string); ok {
					return text
				}
			}
		}
	}
	return ""
}

func extractPreviewMessage(msg Message) PreviewMessage {
	// Check for summary first (but filter out system/error messages)
	if msg.Summary != "" {
		// Skip summaries that look like system messages
		lower := strings.ToLower(msg.Summary)
		if strings.Contains(lower, "no conversations") ||
			strings.Contains(lower, "error") ||
			strings.Contains(lower, "failed") {
			return PreviewMessage{}
		}
		return PreviewMessage{
			Role: "summary",
			Text: truncate(msg.Summary, 500),
		}
	}

	role := msg.Message.Role
	if role == "" {
		return PreviewMessage{}
	}

	if msg.Message.Content == nil {
		return PreviewMessage{}
	}

	// Handle different content types - use longer limit for preview display
	var text string
	switch content := msg.Message.Content.(type) {
	case string:
		text = truncate(content, 500)
	case []any:
		// Array of content blocks
		for _, block := range content {
			if m, ok := block.(map[string]any); ok {
				if t, ok := m["text"].(string); ok {
					text = truncate(t, 500)
					break
				}
			}
		}
	}

	if text == "" {
		return PreviewMessage{}
	}

	return PreviewMessage{
		Role: role,
		Text: text,
	}
}

func truncate(s string, maxLen int) string {
	// Remove newlines
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.TrimSpace(s)

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Find finds a session by ID or name.
func (m *Manager) Find(identifier string) (*Session, error) {
	sessions, err := m.List()
	if err != nil {
		return nil, err
	}

	// Try exact ID match first
	for _, s := range sessions {
		if s.ID == identifier {
			return s, nil
		}
	}

	// Try name match
	identifier = strings.ToLower(identifier)
	for _, s := range sessions {
		if strings.ToLower(s.Name) == identifier {
			return s, nil
		}
	}

	// Try partial ID match
	for _, s := range sessions {
		if strings.HasPrefix(s.ID, identifier) {
			return s, nil
		}
	}

	return nil, os.ErrNotExist
}

// Delete removes a session file and its metadata.
func (m *Manager) Delete(sessionID string) error {
	sessions, err := m.List()
	if err != nil {
		return err
	}

	for _, s := range sessions {
		if s.ID == sessionID {
			if err := os.Remove(s.Path); err != nil {
				return err
			}
			return m.metadata.RemoveSession(sessionID)
		}
	}

	return os.ErrNotExist
}

// GetPreview returns full preview content for a session.
func (m *Manager) GetPreview(sessionID string, maxLines int) ([]PreviewMessage, error) {
	session, err := m.Find(sessionID)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(session.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var messages []PreviewMessage
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() && len(messages) < maxLines {
		var msg Message
		if err := json.Unmarshal([]byte(scanner.Text()), &msg); err != nil {
			continue
		}

		preview := extractPreviewMessage(msg)
		if preview.Text != "" {
			messages = append(messages, preview)
		}
	}

	return messages, nil
}
