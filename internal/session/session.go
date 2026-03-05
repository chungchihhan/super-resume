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

	"github.com/harrychung/session-manager/internal/metadata"
)

// Session represents a Claude Code session.
type Session struct {
	ID           string
	Path         string
	Directory    string
	Name         string
	Created      time.Time
	Modified     time.Time
	MessageCount int
	IsPinned     bool
	Tags         []string
	Preview      []string // First few messages for preview
}

// Message represents a single message in the transcript.
type Message struct {
	Type    string `json:"type"`
	Message struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	} `json:"message"`
	Summary string `json:"summary"`
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

// ListForCurrentDir returns sessions for the current directory only.
func (m *Manager) ListForCurrentDir() ([]*Session, error) {
	all, err := m.List()
	if err != nil {
		return nil, err
	}

	if m.currentDir == "" {
		return all, nil
	}

	// Encode the current directory path like Claude Code does
	encodedDir := encodeDirPath(m.currentDir)

	var filtered []*Session
	for _, s := range all {
		// Match if the session's directory matches the encoded current dir
		if s.Directory == encodedDir || strings.HasPrefix(s.Directory, encodedDir) {
			filtered = append(filtered, s)
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

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()

		if lineCount == 1 {
			// Try to extract session name from first message
			var msg Message
			if err := json.Unmarshal([]byte(line), &msg); err == nil {
				if msg.Summary != "" {
					session.Name = truncate(msg.Summary, 50)
				}
			}
		}

		// Collect preview messages
		if len(session.Preview) < previewLines {
			var msg Message
			if err := json.Unmarshal([]byte(line), &msg); err == nil {
				preview := extractPreviewText(msg)
				if preview != "" {
					session.Preview = append(session.Preview, preview)
				}
			}
		}
	}

	session.MessageCount = lineCount
	return session, nil
}

func extractPreviewText(msg Message) string {
	if msg.Summary != "" {
		return truncate(msg.Summary, 80)
	}

	if msg.Message.Content == nil {
		return ""
	}

	// Handle different content types
	switch content := msg.Message.Content.(type) {
	case string:
		return truncate(content, 80)
	case []any:
		// Array of content blocks
		for _, block := range content {
			if m, ok := block.(map[string]any); ok {
				if text, ok := m["text"].(string); ok {
					return truncate(text, 80)
				}
			}
		}
	}

	return ""
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
func (m *Manager) GetPreview(sessionID string, maxLines int) ([]string, error) {
	session, err := m.Find(sessionID)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(session.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() && len(lines) < maxLines {
		var msg Message
		if err := json.Unmarshal([]byte(scanner.Text()), &msg); err != nil {
			continue
		}

		role := msg.Message.Role
		if role == "" {
			role = msg.Type
		}

		preview := extractPreviewText(msg)
		if preview != "" {
			lines = append(lines, role+": "+preview)
		}
	}

	return lines, nil
}
