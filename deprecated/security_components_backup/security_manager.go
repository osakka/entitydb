package main

import (
	"log"
	"net/http"
)

// SecurityManager manages all security components
type SecurityManager struct {
	validator   *InputValidator
	auditLogger *AuditLogger
	server      interface{} // Generic interface to allow both real and mock servers
}

// NewSecurityManager creates a new security manager for any server type
func NewSecurityManager(server interface{}) *SecurityManager {
	sm := &SecurityManager{
		server: server,
	}

	// Initialize the input validator
	sm.validator = NewInputValidator()
	log.Printf("EntityDB Server: Input validator initialized")

	// Initialize the audit logger
	var entities map[string]map[string]interface{}

	// Try to extract entities from either server type
	switch s := server.(type) {
	case *MockServer:
		entities = s.Entities
	default:
		// For the real server, we can't extract entities yet
		entities = make(map[string]map[string]interface{})
	}

	auditLogger, err := NewAuditLogger("/opt/entitydb/var/log/audit", entities)
	if err != nil {
		log.Printf("EntityDB Server: Warning - Failed to initialize audit logger: %v", err)
	} else {
		sm.auditLogger = auditLogger
		log.Printf("EntityDB Server: Audit logger initialized")
	}

	return sm
}

// Close closes all security components
func (sm *SecurityManager) Close() {
	if sm.auditLogger != nil {
		if err := sm.auditLogger.Close(); err != nil {
			log.Printf("EntityDB Server: Error closing audit logger: %v", err)
		} else {
			log.Printf("EntityDB Server: Audit logger closed successfully")
		}
	}
}

// ValidateLoginRequest validates a login request
func (sm *SecurityManager) ValidateLoginRequest(w http.ResponseWriter, req map[string]interface{}) bool {
	if sm.validator == nil {
		return true
	}
	return sm.validator.ValidateLogin(w, req)
}

// ValidateEntityCreate validates an entity creation request
func (sm *SecurityManager) ValidateEntityCreate(w http.ResponseWriter, req map[string]interface{}) bool {
	if sm.validator == nil {
		return true
	}
	return sm.validator.ValidateEntityCreate(w, req)
}

// ValidateEntityUpdate validates an entity update request
func (sm *SecurityManager) ValidateEntityUpdate(w http.ResponseWriter, req map[string]interface{}) bool {
	if sm.validator == nil {
		return true
	}
	return sm.validator.ValidateEntityUpdate(w, req)
}

// ValidateRelationshipCreate validates a relationship creation request
func (sm *SecurityManager) ValidateRelationshipCreate(w http.ResponseWriter, req map[string]interface{}) bool {
	if sm.validator == nil {
		return true
	}
	return sm.validator.ValidateRelationshipCreate(w, req)
}

// LogAuthEvent logs an authentication event
func (sm *SecurityManager) LogAuthEvent(userID, username, action, status, ip string, details map[string]interface{}) {
	if sm.auditLogger == nil {
		// If no audit logger, just log to standard log
		log.Printf("SECURITY: Auth event: %s %s %s %s", userID, username, action, status)
		return
	}
	sm.auditLogger.LogAuthEvent(userID, username, action, status, ip, details)
}

// LogAccessEvent logs an access control event
func (sm *SecurityManager) LogAccessEvent(userID, username, action, status, path string, details map[string]interface{}) {
	if sm.auditLogger == nil {
		// If no audit logger, just log to standard log
		log.Printf("SECURITY: Access event: %s %s %s %s %s", userID, username, action, status, path)
		return
	}
	sm.auditLogger.LogAccessEvent(userID, username, action, status, path, details)
}

// LogEntityEvent logs an entity-related event
func (sm *SecurityManager) LogEntityEvent(userID, username, entityID, entityType, action, status string, details map[string]interface{}) {
	if sm.auditLogger == nil {
		// If no audit logger, just log to standard log
		log.Printf("SECURITY: Entity event: %s %s %s %s %s %s", userID, username, entityID, entityType, action, status)
		return
	}
	sm.auditLogger.LogEntityEvent(userID, username, entityID, entityType, action, status, details)
}

// LogAdminEvent logs an administrative event
func (sm *SecurityManager) LogAdminEvent(userID, username, action, status string, details map[string]interface{}) {
	if sm.auditLogger == nil {
		// If no audit logger, just log to standard log
		log.Printf("SECURITY: Admin event: %s %s %s %s", userID, username, action, status)
		return
	}
	sm.auditLogger.LogAdminEvent(userID, username, action, status, details)
}

// SecureMiddleware provides a security middleware for HTTP handlers
type SecureMiddleware struct {
	sm     *SecurityManager
	server interface{}
	rbac   *RBACMiddleware
}

// NewSecureMiddleware creates a new security middleware
func NewSecureMiddleware(sm *SecurityManager, server interface{}) *SecureMiddleware {
	middleware := &SecureMiddleware{
		sm:     sm,
		server: server,
	}

	// Initialize RBAC middleware - use improved version if available
	if containsRole != nil {
		middleware.rbac = NewRBACMiddleware(sm, server)
	} else {
		log.Printf("EntityDB Server: Using basic RBAC middleware (consider upgrading to improved version)")
		middleware.rbac = NewRBACMiddleware(sm, server)
	}

	return middleware
}

// Wrap wraps an HTTP handler with security features
func (m *SecureMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for login and status endpoints
		if r.URL.Path == "/api/v1/login" || r.URL.Path == "/api/v1/auth/login" ||
		   r.URL.Path == "/api/v1/status" || r.URL.Path == "/health" || r.URL.Path == "/api/v1/health" {
			// Log request metadata
			log.Printf("EntityDB Server: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

			// Create basic logging for any request
			m.sm.LogAccessEvent("unknown", "unknown", "access", "info", r.URL.Path, map[string]interface{}{
				"method": r.Method,
				"ip":     r.RemoteAddr,
			})

			// Call the next handler directly
			next(w, r)
			return
		}

		// Log request metadata
		log.Printf("EntityDB Server: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Create basic logging for any request
		m.sm.LogAccessEvent("unknown", "unknown", "access", "info", r.URL.Path, map[string]interface{}{
			"method": r.Method,
			"ip":     r.RemoteAddr,
		})

		// Apply RBAC middleware
		permission := GetRequiredPermission(r.Method, r.URL.Path)
		rbacHandler := m.rbac.RequirePermission(permission, next)
		rbacHandler(w, r)
	}
}