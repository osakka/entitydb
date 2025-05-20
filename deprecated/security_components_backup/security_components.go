package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Security components for the EntityDB server
// These components provide secure password handling, input validation,
// audit logging, and other security features.

// SecurityManager manages all security components
type SecurityManager struct {
	validator   *InputValidator
	auditLogger *AuditLogger
	server      *EntityDBServer
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(server *EntityDBServer) *SecurityManager {
	sm := &SecurityManager{
		server: server,
	}
	
	// Initialize the input validator
	sm.validator = NewInputValidator()
	log.Printf("EntityDB Server: Input validator initialized")
	
	// Initialize the audit logger
	auditLogger, err := NewAuditLogger("/opt/entitydb/var/log/audit", server.entities)
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
		return
	}
	sm.auditLogger.LogAuthEvent(userID, username, action, status, ip, details)
}

// LogAccessEvent logs an access control event
func (sm *SecurityManager) LogAccessEvent(userID, username, action, status, path string, details map[string]interface{}) {
	if sm.auditLogger == nil {
		return
	}
	sm.auditLogger.LogAccessEvent(userID, username, action, status, path, details)
}

// LogEntityEvent logs an entity-related event
func (sm *SecurityManager) LogEntityEvent(userID, username, entityID, entityType, action, status string, details map[string]interface{}) {
	if sm.auditLogger == nil {
		return
	}
	sm.auditLogger.LogEntityEvent(userID, username, entityID, entityType, action, status, details)
}

// LogAdminEvent logs an administrative event
func (sm *SecurityManager) LogAdminEvent(userID, username, action, status string, details map[string]interface{}) {
	if sm.auditLogger == nil {
		return
	}
	sm.auditLogger.LogAdminEvent(userID, username, action, status, details)
}

// SecureMiddleware provides a security middleware for HTTP handlers
type SecureMiddleware struct {
	sm     *SecurityManager
	server *EntityDBServer
}

// NewSecureMiddleware creates a new security middleware
func NewSecureMiddleware(sm *SecurityManager, server *EntityDBServer) *SecureMiddleware {
	return &SecureMiddleware{
		sm:     sm,
		server: server,
	}
}

// Wrap wraps an HTTP handler with security features
func (m *SecureMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Log request metadata
		log.Printf("EntityDB Server: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		// Get user from authentication header if present
		var user *User
		var userID, username string
		
		// Check for Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Parse token from Authorization header (Bearer token)
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				token := tokenParts[1]
				user = m.server.validateToken(token)
				
				if user != nil {
					userID = user.ID
					username = user.Username
					
					// Log access event for authenticated requests
					m.sm.LogAccessEvent(userID, username, "access", "success", r.URL.Path, map[string]interface{}{
						"method": r.Method,
						"ip":     r.RemoteAddr,
					})
				} else {
					// Log invalid token attempt
					m.sm.LogAccessEvent("unknown", "unknown", "invalid_token", "failure", r.URL.Path, map[string]interface{}{
						"method": r.Method,
						"ip":     r.RemoteAddr,
					})
				}
			}
		}
		
		// Call the next handler
		next(w, r)
	}
}

// AuditLogger manages security audit logging
type AuditLogger struct {
	logFile   *os.File
	enabled   bool
	logPath   string
	entityMap map[string]map[string]interface{}
}

// AuditEvent represents a security-relevant event
type AuditEvent struct {
	Timestamp   string                 `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username,omitempty"`
	EntityID    string                 `json:"entity_id,omitempty"`
	EntityType  string                 `json:"entity_type,omitempty"`
	Action      string                 `json:"action"`
	Status      string                 `json:"status"`
	IP          string                 `json:"ip,omitempty"`
	RequestPath string                 `json:"request_path,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logDir string, entityMap map[string]map[string]interface{}) (*AuditLogger, error) {
	if logDir == "" {
		logDir = "/opt/entitydb/var/log/audit"
	}

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create audit log directory: %v", err)
	}

	// Create log file with today's date
	logPath := filepath.Join(logDir, fmt.Sprintf("entitydb_audit_%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit log file: %v", err)
	}

	logger := &AuditLogger{
		logFile:   logFile,
		enabled:   true,
		logPath:   logPath,
		entityMap: entityMap,
	}

	// Log initialization
	initEvent := AuditEvent{
		Timestamp: time.Now().Format(time.RFC3339),
		EventType: "system",
		Action:    "audit_logger_init",
		Status:    "success",
		Details: map[string]interface{}{
			"log_path": logPath,
		},
	}
	logger.LogEvent(initEvent)

	log.Printf("EntityDB Server: Audit logger initialized with log path: %s", logPath)

	return logger, nil
}

// LogEvent logs a security-relevant event
func (a *AuditLogger) LogEvent(event AuditEvent) error {
	if !a.enabled {
		return nil
	}

	// Ensure timestamp is set
	if event.Timestamp == "" {
		event.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Add entity type if we have entity ID but no type
	if event.EntityID != "" && event.EntityType == "" {
		if entity, exists := a.entityMap[event.EntityID]; exists {
			if entityType, ok := entity["type"].(string); ok {
				event.EntityType = entityType
			}
		}
	}

	// Add username if we have user ID but no username
	if event.UserID != "" && event.Username == "" {
		userEntityID := fmt.Sprintf("entity_user_%s", event.UserID)
		if entity, exists := a.entityMap[userEntityID]; exists {
			if username, ok := entity["title"].(string); ok {
				event.Username = username
			}
		}
	}

	// Convert to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("EntityDB Server: Error marshalling audit event: %v", err)
		return err
	}

	// Write to log file
	if _, err := a.logFile.Write(append(eventJSON, '\n')); err != nil {
		log.Printf("EntityDB Server: Error writing to audit log: %v", err)
		return err
	}

	return nil
}

// LogAuthEvent logs authentication-related events
func (a *AuditLogger) LogAuthEvent(userID, username, action, status, ip string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp: time.Now().Format(time.RFC3339),
		EventType: "authentication",
		UserID:    userID,
		Username:  username,
		Action:    action,
		Status:    status,
		IP:        ip,
		Details:   details,
	}
	a.LogEvent(event)
}

// LogAccessEvent logs access control events
func (a *AuditLogger) LogAccessEvent(userID, username, action, status, path string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp:   time.Now().Format(time.RFC3339),
		EventType:   "access_control",
		UserID:      userID,
		Username:    username,
		Action:      action,
		Status:      status,
		RequestPath: path,
		Details:     details,
	}
	a.LogEvent(event)
}

// LogEntityEvent logs entity-related events
func (a *AuditLogger) LogEntityEvent(userID, username, entityID, entityType, action, status string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp:  time.Now().Format(time.RFC3339),
		EventType:  "entity",
		UserID:     userID,
		Username:   username,
		EntityID:   entityID,
		EntityType: entityType,
		Action:     action,
		Status:     status,
		Details:    details,
	}
	a.LogEvent(event)
}

// LogAdminEvent logs administrative events
func (a *AuditLogger) LogAdminEvent(userID, username, action, status string, details map[string]interface{}) {
	event := AuditEvent{
		Timestamp: time.Now().Format(time.RFC3339),
		EventType: "administrative",
		UserID:    userID,
		Username:  username,
		Action:    action,
		Status:    status,
		Details:   details,
	}
	a.LogEvent(event)
}

// Close closes the audit logger
func (a *AuditLogger) Close() error {
	if a.logFile != nil {
		// Log closure
		closeEvent := AuditEvent{
			Timestamp: time.Now().Format(time.RFC3339),
			EventType: "system",
			Action:    "audit_logger_close",
			Status:    "success",
		}
		a.LogEvent(closeEvent)

		return a.logFile.Close()
	}
	return nil
}

// RotateLog rotates the log file
func (a *AuditLogger) RotateLog() error {
	// Close current log file
	if err := a.logFile.Close(); err != nil {
		return err
	}

	// Create new log file with today's date
	logPath := filepath.Join(filepath.Dir(a.logPath), fmt.Sprintf("entitydb_audit_%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new audit log file during rotation: %v", err)
	}

	// Update logger
	a.logFile = logFile
	a.logPath = logPath

	// Log rotation
	rotateEvent := AuditEvent{
		Timestamp: time.Now().Format(time.RFC3339),
		EventType: "system",
		Action:    "audit_log_rotation",
		Status:    "success",
		Details: map[string]interface{}{
			"new_log_path": logPath,
		},
	}
	a.LogEvent(rotateEvent)

	log.Printf("EntityDB Server: Audit log rotated to: %s", logPath)

	return nil
}

// InputValidator handles input validation for API endpoints
type InputValidator struct {
	// Validation patterns
	patterns map[string]*regexp.Regexp
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	validator := &InputValidator{
		patterns: make(map[string]*regexp.Regexp),
	}

	// Initialize common validation patterns
	validator.patterns["username"] = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
	validator.patterns["password"] = regexp.MustCompile(`^.{8,}$`) // Minimum 8 chars
	validator.patterns["entityID"] = regexp.MustCompile(`^entity_[a-zA-Z0-9_]{1,64}$`)
	validator.patterns["relID"] = regexp.MustCompile(`^rel_[a-zA-Z0-9_]{1,64}$`)
	validator.patterns["status"] = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)
	validator.patterns["type"] = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)
	validator.patterns["title"] = regexp.MustCompile(`^.{1,256}$`) // Non-empty, max 256
	validator.patterns["tag"] = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
	validator.patterns["role"] = regexp.MustCompile(`^(admin|user|readonly)$`)

	return validator
}

// Validate validates input based on the specified rules
func (v *InputValidator) Validate(input map[string]interface{}, rules map[string]string) []ValidationError {
	var errors []ValidationError

	for field, rule := range rules {
		// Skip validation if field is not required and not present
		value, exists := input[field]
		if !exists {
			if strings.Contains(rule, "required") {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: "Field is required",
				})
			}
			continue
		}

		// Handle different validation rules
		if strings.Contains(rule, "required") {
			// Already checked above
		}

		// Type validations
		if strings.Contains(rule, "string") {
			if _, ok := value.(string); !ok {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: "Field must be a string",
				})
				continue
			}
		}

		if strings.Contains(rule, "array") {
			if _, ok := value.([]interface{}); !ok {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: "Field must be an array",
				})
				continue
			}
		}

		if strings.Contains(rule, "object") {
			if _, ok := value.(map[string]interface{}); !ok {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: "Field must be an object",
				})
				continue
			}
		}

		// Pattern validations
		for pattern, regex := range v.patterns {
			if strings.Contains(rule, pattern) {
				if strValue, ok := value.(string); ok {
					if !regex.MatchString(strValue) {
						errors = append(errors, ValidationError{
							Field:   field,
							Message: fmt.Sprintf("Field does not match %s pattern", pattern),
						})
					}
				}
			}
		}

		// Array item validations
		if strings.Contains(rule, "array:") {
			if arr, ok := value.([]interface{}); ok {
				// Extract array item validation
				parts := strings.Split(rule, "array:")
				if len(parts) > 1 {
					itemRules := parts[1]
					for i, item := range arr {
						if strings.Contains(itemRules, "string") {
							if _, ok := item.(string); !ok {
								errors = append(errors, ValidationError{
									Field:   fmt.Sprintf("%s[%d]", field, i),
									Message: "Array item must be a string",
								})
								continue
							}
						}

						for pattern, regex := range v.patterns {
							if strings.Contains(itemRules, pattern) {
								if strItem, ok := item.(string); ok {
									if !regex.MatchString(strItem) {
										errors = append(errors, ValidationError{
											Field:   fmt.Sprintf("%s[%d]", field, i),
											Message: fmt.Sprintf("Array item does not match %s pattern", pattern),
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return errors
}

// ValidateAndRespond validates input and writes errors to response if any
func (v *InputValidator) ValidateAndRespond(w http.ResponseWriter, input map[string]interface{}, rules map[string]string) bool {
	errors := v.Validate(input, rules)
	if len(errors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Validation failed",
			"errors":  errors,
		})
		return false
	}
	return true
}

// ValidateEntityCreate validates entity creation input
func (v *InputValidator) ValidateEntityCreate(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"type":        "required|string|type",
		"title":       "required|string|title",
		"description": "string",
		"status":      "string|status",
		"tags":        "array:string|tag",
		"properties":  "object",
	}
	return v.ValidateAndRespond(w, input, rules)
}

// ValidateEntityUpdate validates entity update input
func (v *InputValidator) ValidateEntityUpdate(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"title":       "string|title",
		"description": "string",
		"status":      "string|status",
		"tags":        "array:string|tag",
		"properties":  "object",
	}
	return v.ValidateAndRespond(w, input, rules)
}

// ValidateRelationshipCreate validates relationship creation input
func (v *InputValidator) ValidateRelationshipCreate(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"source_id":  "required|string|entityID",
		"target_id":  "required|string|entityID",
		"type":       "required|string|type",
		"properties": "object",
	}
	return v.ValidateAndRespond(w, input, rules)
}

// ValidateRelationshipUpdate validates relationship update input
func (v *InputValidator) ValidateRelationshipUpdate(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"properties": "required|object",
	}
	return v.ValidateAndRespond(w, input, rules)
}

// ValidateLogin validates login input
func (v *InputValidator) ValidateLogin(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"username": "required|string|username",
		"password": "required|string|password",
	}
	return v.ValidateAndRespond(w, input, rules)
}

// ValidateUserCreate validates user creation input
func (v *InputValidator) ValidateUserCreate(w http.ResponseWriter, input map[string]interface{}) bool {
	rules := map[string]string{
		"type":        "required|string|type",
		"title":       "required|string|username",
		"description": "string",
		"status":      "string|status",
		"tags":        "array:string|tag",
		"properties": "required|object",
	}

	// Basic validation
	if !v.ValidateAndRespond(w, input, rules) {
		return false
	}

	// Special validation for user properties
	properties, ok := input["properties"].(map[string]interface{})
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Validation failed",
			"errors": []ValidationError{{
				Field:   "properties",
				Message: "Must be a valid properties object with username, roles and password_hash",
			}},
		})
		return false
	}

	// Validate properties
	propRules := map[string]string{
		"username":      "required|string|username",
		"roles":         "required|array:string|role",
		"password_hash": "required|string",
	}

	propErrors := v.Validate(properties, propRules)
	if len(propErrors) > 0 {
		// Prefix field names with "properties."
		for i := range propErrors {
			propErrors[i].Field = "properties." + propErrors[i].Field
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Validation failed",
			"errors":  propErrors,
		})
		return false
	}

	return true
}