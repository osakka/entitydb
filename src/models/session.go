// Package models provides core data structures and business logic for EntityDB.
// This file implements session management for user authentication and authorization.
package models

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Session represents an authenticated user session in the system.
// Sessions are used to maintain authentication state between requests
// and track user activity. Each session has a unique token and expiration time.
//
// Example usage:
//
//	session := &Session{
//	    Token:     "token_abc123...",
//	    UserID:    "user-uuid",
//	    Username:  "johndoe",
//	    Roles:     []string{"rbac:role:user", "rbac:role:admin"},
//	    CreatedAt: time.Now(),
//	    ExpiresAt: time.Now().Add(24 * time.Hour),
//	    LastUsed:  time.Now(),
//	}
type Session struct {
	// Token is the unique session identifier used for authentication.
	// Format: "token_" + 64 character hex string (256 bits of entropy)
	Token string `json:"token"`
	
	// UserID is the unique identifier of the authenticated user entity
	UserID string `json:"user_id"`
	
	// Username is the human-readable identifier for the user
	Username string `json:"username"`
	
	// Roles contains the RBAC roles assigned to this session.
	// These are copied from the user entity at session creation time.
	// Format: ["rbac:role:admin", "rbac:role:user", etc.]
	Roles []string `json:"roles"`
	
	// CreatedAt is the timestamp when the session was created
	CreatedAt time.Time `json:"created_at"`
	
	// ExpiresAt is the timestamp when the session will become invalid.
	// Sessions are automatically cleaned up after expiration.
	ExpiresAt time.Time `json:"expires_at"`
	
	// LastUsed is updated on each authenticated request to track activity.
	// This can be used for idle timeout policies.
	LastUsed time.Time `json:"last_used"`
}

// SessionManager manages user sessions with automatic expiration and cleanup.
// It provides thread-safe operations for creating, retrieving, refreshing,
// and revoking sessions. Sessions are stored in-memory and automatically
// cleaned up after expiration.
//
// The manager starts a background goroutine that runs every 5 minutes
// to remove expired sessions and prevent memory leaks.
//
// Example usage:
//
//	// Create a session manager with 24-hour session lifetime
//	sm := NewSessionManager(24 * time.Hour)
//	
//	// Create a new session
//	session, err := sm.CreateSession("user-123", "johndoe", []string{"rbac:role:user"})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	// Validate and retrieve a session
//	if session, exists := sm.GetSession(token); exists {
//	    // Session is valid
//	}
type SessionManager struct {
	// sessions maps session tokens to session objects
	sessions map[string]*Session
	
	// mu provides thread-safe access to the sessions map
	mu sync.RWMutex
	
	// ttl is the time-to-live for new sessions.
	// Sessions expire after this duration from creation or last refresh.
	ttl time.Duration
}

// NewSessionManager creates a new session manager with the specified time-to-live.
// The TTL determines how long sessions remain valid without refresh.
// A background cleanup goroutine is automatically started to remove expired sessions.
//
// Parameters:
//   - ttl: Duration that sessions remain valid (e.g., 24*time.Hour)
//
// Returns:
//   - *SessionManager: The initialized session manager
//
// Example:
//
//	// Create a session manager with 24-hour sessions
//	sm := NewSessionManager(24 * time.Hour)
func NewSessionManager(ttl time.Duration) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
	
	// Start cleanup goroutine to remove expired sessions every 5 minutes
	go sm.cleanupExpiredSessions()
	
	return sm
}

// CreateSession creates a new authenticated session for a user.
// The session is assigned a cryptographically secure random token
// and set to expire after the manager's TTL duration.
//
// Parameters:
//   - userID: The unique identifier of the user entity
//   - username: The human-readable username
//   - roles: RBAC roles to assign to this session (e.g., ["rbac:role:admin"])
//
// Returns:
//   - *Session: The newly created session with token
//   - error: Any error that occurred during token generation
//
// Example:
//
//	session, err := sm.CreateSession(
//	    "550e8400-e29b-41d4-a716-446655440000",
//	    "admin",
//	    []string{"rbac:role:admin", "rbac:perm:*"},
//	)
//	if err != nil {
//	    return fmt.Errorf("failed to create session: %w", err)
//	}
//	// Use session.Token for authentication
func (sm *SessionManager) CreateSession(userID, username string, roles []string) (*Session, error) {
	// Generate cryptographically secure token
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
	
	// Thread-safe storage of the new session
	sm.mu.Lock()
	sm.sessions[token] = session
	sm.mu.Unlock()
	
	return session, nil
}

// GetSession retrieves and validates a session by its token.
// This method checks if the session exists and is not expired.
// If the session is valid, its LastUsed timestamp is updated.
// Expired sessions are automatically removed.
//
// Parameters:
//   - token: The session token to look up
//
// Returns:
//   - *Session: The session if found and valid
//   - bool: true if session exists and is valid, false otherwise
//
// Example:
//
//	session, valid := sm.GetSession("token_abc123...")
//	if !valid {
//	    // Session doesn't exist or has expired
//	    return ErrUnauthorized
//	}
//	// Use session.UserID, session.Roles, etc.
func (sm *SessionManager) GetSession(token string) (*Session, bool) {
	// Read lock for checking existence
	sm.mu.RLock()
	session, exists := sm.sessions[token]
	sm.mu.RUnlock()
	
	if !exists {
		return nil, false
	}
	
	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Remove expired session
		sm.DeleteSession(token)
		return nil, false
	}
	
	// Update last used time to track activity
	sm.mu.Lock()
	session.LastUsed = time.Now()
	sm.mu.Unlock()
	
	return session, true
}

// RefreshSession extends the expiration time of an existing session.
// This is useful for implementing "remember me" functionality or
// extending sessions for active users. The session's TTL is reset
// to the full duration from the current time.
//
// Parameters:
//   - token: The session token to refresh
//
// Returns:
//   - *Session: The refreshed session with new expiration time
//   - bool: true if session was refreshed, false if not found or expired
//
// Example:
//
//	// Refresh session on important operations
//	if session, refreshed := sm.RefreshSession(token); refreshed {
//	    log.Printf("Session extended until %v", session.ExpiresAt)
//	} else {
//	    // Session expired or doesn't exist
//	}
func (sm *SessionManager) RefreshSession(token string) (*Session, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[token]
	// Cannot refresh non-existent or already expired sessions
	if !exists || time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	
	// Extend expiration by full TTL duration
	session.ExpiresAt = time.Now().Add(sm.ttl)
	session.LastUsed = time.Now()
	
	return session, true
}

// DeleteSession immediately removes a session from the manager.
// This is used for logout operations or when revoking access.
// The operation is idempotent - deleting a non-existent session is a no-op.
//
// Parameters:
//   - token: The session token to delete
//
// Example:
//
//	// Logout endpoint
//	sm.DeleteSession(sessionToken)
//	// Session is now invalid
func (sm *SessionManager) DeleteSession(token string) {
	sm.mu.Lock()
	delete(sm.sessions, token)
	sm.mu.Unlock()
}

// cleanupExpiredSessions runs as a background goroutine to periodically
// remove expired sessions from memory. This prevents memory leaks from
// accumulating expired sessions. The cleanup runs every 5 minutes.
//
// This method is automatically started by NewSessionManager and should
// not be called directly.
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		// Iterate through all sessions and remove expired ones
		for token, session := range sm.sessions {
			if now.After(session.ExpiresAt) {
				delete(sm.sessions, token)
			}
		}
		sm.mu.Unlock()
	}
}

// generateToken creates a cryptographically secure random token.
// The token consists of 256 bits (32 bytes) of random data encoded
// as hexadecimal and prefixed with "token_" for easy identification.
//
// Returns:
//   - string: Token in format "token_" + 64 hex characters
//   - error: Any error from the random number generator
//
// The function uses crypto/rand which provides cryptographically
// secure random numbers suitable for session tokens.
func generateToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits of entropy
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "token_" + hex.EncodeToString(bytes), nil
}

// GetActiveSessions returns the count of currently active (non-expired) sessions.
// This is useful for monitoring and statistics.
//
// Returns:
//   - int: Number of active sessions
//
// Example:
//
//	activeCount := sm.GetActiveSessions()
//	log.Printf("Currently %d active sessions", activeCount)
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

// GetUserSessions returns all active sessions for a specific user.
// This is useful for allowing users to see their active sessions
// or for administrators to audit user access.
//
// Parameters:
//   - userID: The user's unique identifier
//
// Returns:
//   - []*Session: Slice of active sessions for the user (may be empty)
//
// Example:
//
//	sessions := sm.GetUserSessions("user-123")
//	for _, session := range sessions {
//	    fmt.Printf("Session from %v, last used %v\n", 
//	        session.CreatedAt, session.LastUsed)
//	}
func (sm *SessionManager) GetUserSessions(userID string) []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	var userSessions []*Session
	now := time.Now()
	
	// Collect all non-expired sessions for this user
	for _, session := range sm.sessions {
		if session.UserID == userID && now.Before(session.ExpiresAt) {
			userSessions = append(userSessions, session)
		}
	}
	
	return userSessions
}

// RevokeUserSessions removes all sessions for a specific user.
// This is typically used when a user's access needs to be immediately
// revoked, such as when their account is disabled or their password
// is changed.
//
// Parameters:
//   - userID: The user's unique identifier
//
// Returns:
//   - int: Number of sessions that were revoked
//
// Example:
//
//	// Revoke all sessions when password changes
//	revokedCount := sm.RevokeUserSessions(userID)
//	log.Printf("Revoked %d sessions for user %s", revokedCount, userID)
func (sm *SessionManager) RevokeUserSessions(userID string) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	count := 0
	// Remove all sessions belonging to this user
	for token, session := range sm.sessions {
		if session.UserID == userID {
			delete(sm.sessions, token)
			count++
		}
	}
	
	return count
}