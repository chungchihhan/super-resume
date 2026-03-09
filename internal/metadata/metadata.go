// Package metadata handles persistent storage for session pins and tags.
package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Store manages session metadata (pins, tags).
type Store struct {
	path string
	data Data
	mu   sync.RWMutex
}

// Data represents the metadata structure.
type Data struct {
	Pinned map[string]bool     `json:"pinned"`
	Tags   map[string][]string `json:"tags"`
	Config Config              `json:"config"`
}

// Config holds user configuration.
type Config struct {
	Terminal string `json:"terminal"` // terminal, iterm, warp, kitty, alacritty
}

// New creates a new metadata store.
func New() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(home, ".claude", "session-metadata.json")
	s := &Store{
		path: path,
		data: Data{
			Pinned: make(map[string]bool),
			Tags:   make(map[string][]string),
		},
	}

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := json.Unmarshal(data, &s.data); err != nil {
		return err
	}

	// Initialize maps if nil
	if s.data.Pinned == nil {
		s.data.Pinned = make(map[string]bool)
	}
	if s.data.Tags == nil {
		s.data.Tags = make(map[string][]string)
	}

	return nil
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// IsPinned checks if a session is pinned.
func (s *Store) IsPinned(sessionID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Pinned[sessionID]
}

// TogglePin toggles the pin status of a session.
func (s *Store) TogglePin(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data.Pinned[sessionID] {
		delete(s.data.Pinned, sessionID)
	} else {
		s.data.Pinned[sessionID] = true
	}

	return s.save()
}

// Pin pins a session.
func (s *Store) Pin(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Pinned[sessionID] = true
	return s.save()
}

// Unpin unpins a session.
func (s *Store) Unpin(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data.Pinned, sessionID)
	return s.save()
}

// GetTags returns tags for a session.
func (s *Store) GetTags(sessionID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Tags[sessionID]
}

// AddTag adds a tag to a session.
func (s *Store) AddTag(sessionID, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tags := s.data.Tags[sessionID]
	for _, t := range tags {
		if t == tag {
			return nil // Already has this tag
		}
	}

	s.data.Tags[sessionID] = append(tags, tag)
	return s.save()
}

// RemoveTag removes a tag from a session.
func (s *Store) RemoveTag(sessionID, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tags := s.data.Tags[sessionID]
	for i, t := range tags {
		if t == tag {
			s.data.Tags[sessionID] = append(tags[:i], tags[i+1:]...)
			break
		}
	}

	return s.save()
}

// SetTags sets all tags for a session.
func (s *Store) SetTags(sessionID string, tags []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(tags) == 0 {
		delete(s.data.Tags, sessionID)
	} else {
		s.data.Tags[sessionID] = tags
	}

	return s.save()
}

// RemoveSession removes all metadata for a session (called on delete).
func (s *Store) RemoveSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data.Pinned, sessionID)
	delete(s.data.Tags, sessionID)

	return s.save()
}

// AllPinned returns all pinned session IDs.
func (s *Store) AllPinned() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pinned []string
	for id, isPinned := range s.data.Pinned {
		if isPinned {
			pinned = append(pinned, id)
		}
	}
	return pinned
}

// GetTerminal returns the configured terminal type.
func (s *Store) GetTerminal() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Config.Terminal
}

// SetTerminal sets the terminal type.
func (s *Store) SetTerminal(terminal string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Config.Terminal = terminal
	return s.save()
}

// SupportedTerminals returns a list of supported terminal types.
func SupportedTerminals() []string {
	return []string{"terminal", "iterm", "warp", "kitty", "alacritty"}
}
