package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session represents an active shell session
type Session struct {
	ID           string
	ClientAddr  string
	CreatedAt    time.Time
	LastActivity time.Time
	IsActive     bool
	Process      *Process
	mu           sync.RWMutex
}

// Process represents a running process
type Process struct {
	PID     int
	Cmd     string
	Args    []string
	Env     map[string]string
	WorkDir string
}

// Manager manages all active sessions
type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewManager creates a new session manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session
func (m *Manager) CreateSession(clientAddr string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &Session{
		ID:           uuid.New().String(),
		ClientAddr:   clientAddr,
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		IsActive:     true,
	}

	m.sessions[session.ID] = session
	return session
}

// GetSession retrieves a session by ID
func (m *Manager) GetSession(sessionID string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[sessionID]
	return session, exists
}

// UpdateActivity updates the last activity time for a session
func (m *Manager) UpdateActivity(sessionID string) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if exists {
		session.mu.Lock()
		session.LastActivity = time.Now()
		session.mu.Unlock()
	}
}

// KillSession terminates a session
func (m *Manager) KillSession(sessionID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return false
	}

	session.mu.Lock()
	session.IsActive = false
	session.mu.Unlock()

	delete(m.sessions, sessionID)
	return true
}

// ListSessions returns all active sessions
func (m *Manager) ListSessions() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// SetProcess sets the process for a session
func (s *Session) SetProcess(proc *Process) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Process = proc
}

// GetProcess returns the process for a session
func (s *Session) GetProcess() *Process {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Process
}

