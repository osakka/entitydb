package models

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Session represents a user session
type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	LastUsed  time.Time `json:"last_used"`
}

// SessionManager manages user sessions with expiration
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	ttl      time.Duration // session time-to-live
}

// NewSessionManager creates a new session manager
func NewSessionManager(ttl time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
	
	// Start cleanup goroutine
	go sm.cleanupExpiredSessions()
	
	return sm
}

// CreateSession creates a new session for a user
func (sm *SessionManager) CreateSession(userID, username string, roles []string) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}
	
	now := time.Now()
	session := &Session{
		Token:     token,
		UserID:    userID,
		Username:  username,
		Roles:     roles,
		CreatedAt: now,
		ExpiresAt: now.Add(sm.ttl),
		LastUsed:  now,
	}
	
	sm.mu.Lock()
	sm.sessions[token] = session
	sm.mu.Unlock()
	
	return session, nil
}

// GetSession retrieves a session by token
func (sm *SessionManager) GetSession(token string) (*Session, bool) {
	sm.mu.RLock()
	session, exists := sm.sessions[token]
	sm.mu.RUnlock()
	
	if !exists {
		return nil, false
	}
	
	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		sm.DeleteSession(token)
		return nil, false
	}
	
	// Update last used time
	sm.mu.Lock()
	session.LastUsed = time.Now()
	sm.mu.Unlock()
	
	return session, true
}

// RefreshSession extends the expiration time of a session
func (sm *SessionManager) RefreshSession(token string) (*Session, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[token]
	if !exists || time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	
	session.ExpiresAt = time.Now().Add(sm.ttl)
	session.LastUsed = time.Now()
	
	return session, true
}

// DeleteSession removes a session
func (sm *SessionManager) DeleteSession(token string) {
	sm.mu.Lock()
	delete(sm.sessions, token)
	sm.mu.Unlock()
}

// cleanupExpiredSessions periodically removes expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for token, session := range sm.sessions {
			if now.After(session.ExpiresAt) {
				delete(sm.sessions, token)
			}
		}
		sm.mu.Unlock()
	}
}

// generateToken creates a secure random token
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "token_" + hex.EncodeToString(bytes), nil
}

// GetActiveSessions returns count of active sessions
func (sm *SessionManager) GetActiveSessions() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	count := 0
	now := time.Now()
	for _, session := range sm.sessions {
		if now.Before(session.ExpiresAt) {
			count++
		}
	}
	return count
}

// GetUserSessions returns all sessions for a specific user
func (sm *SessionManager) GetUserSessions(userID string) []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	var userSessions []*Session
	now := time.Now()
	
	for _, session := range sm.sessions {
		if session.UserID == userID && now.Before(session.ExpiresAt) {
			userSessions = append(userSessions, session)
		}
	}
	
	return userSessions
}

// RevokeUserSessions removes all sessions for a specific user
func (sm *SessionManager) RevokeUserSessions(userID string) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	count := 0
	for token, session := range sm.sessions {
		if session.UserID == userID {
			delete(sm.sessions, token)
			count++
		}
	}
	
	return count
}