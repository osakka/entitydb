// security_input_audit.go
// This file contains the InputValidator and AuditLogger components.

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
)

// InputValidator validates API inputs
type InputValidator struct {
	patterns map[string]*regexp.Regexp
}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	validator := &InputValidator{
		patterns: make(map[string]*regexp.Regexp),
	}
	
	// Add default patterns with stricter validation
	validator.patterns["id"] = regexp.MustCompile(`^[a-zA-Z0-9_\-]{3,40}$`)
	validator.patterns["type"] = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]{1,19}$`) // Must start with letter, no special chars
	validator.patterns["title"] = regexp.MustCompile(`^[\w\s\.\-,:;!?()]{1,100}$`) // More restricted chars for title
	validator.patterns["description"] = regexp.MustCompile(`^[\w\s\.\-,:;!?()]{0,1000}$`) // Same for description
	validator.patterns["status"] = regexp.MustCompile(`^[a-z][a-z0-9_\-]{1,19}$`) // Lowercase for status codes
	validator.patterns["tag"] = regexp.MustCompile(`^[a-z][a-z0-9_\-:]{1,39}$`) // Tags should be lowercase
	validator.patterns["username"] = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]{2,19}$`) // Must start with letter
	validator.patterns["password"] = regexp.MustCompile(`^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*?&]{8,32}$`) // Require letter and number
	validator.patterns["token"] = regexp.MustCompile(`^tk_[a-zA-Z0-9_\-]{6,40}$`)
	
	return validator
}

// ValidatePattern validates a string against a named pattern
func (v *InputValidator) ValidatePattern(name, value string) bool {
	pattern, exists := v.patterns[name]
	if !exists {
		log.Printf("EntityDB Server: No pattern found for %s", name)
		return true
	}
	
	return pattern.MatchString(value)
}

// ValidateLogin validates a login request
func (v *InputValidator) ValidateLogin(w http.ResponseWriter, req map[string]interface{}) bool {
	// Check for required fields
	username, ok := req["username"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Username is required",
		})
		return false
	}
	
	password, ok := req["password"].(string)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Password is required",
		})
		return false
	}
	
	// Validate username pattern
	if !v.ValidatePattern("username", username) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid username format",
		})
		return false
	}
	
	// Validate password pattern
	if !v.ValidatePattern("password", password) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid password format",
		})
		return false
	}
	
	return true
}

// ValidateEntityCreate validates an entity creation request
func (v *InputValidator) ValidateEntityCreate(w http.ResponseWriter, req map[string]interface{}) bool {
	// Validation errors
	var errors []ValidationError

	// Check for required fields
	entityType, hasType := req["type"].(string)
	if !hasType {
		errors = append(errors, ValidationError{
			Field:   "type",
			Message: "Entity type is required",
		})
	}

	title, hasTitle := req["title"].(string)
	if !hasTitle {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "Entity title is required",
		})
	}

	// Return early if required fields are missing
	if len(errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Type and title are required for entity creation",
			"errors":  errors,
		})
		return false
	}

	// Validate type pattern
	if !v.ValidatePattern("type", entityType) {
		errors = append(errors, ValidationError{
			Field:   "type",
			Message: "Entity type must start with a letter and contain only letters, numbers, underscores and hyphens",
		})
	}

	// Check reserved types
	reservedTypes := map[string]bool{
		"system":   true,
		"internal": true,
		"admin":    true,
		"security": true,
	}
	if reservedTypes[strings.ToLower(entityType)] {
		errors = append(errors, ValidationError{
			Field:   "type",
			Message: "This entity type is reserved for system use",
		})
	}

	// Validate title pattern
	if !v.ValidatePattern("title", title) {
		errors = append(errors, ValidationError{
			Field:   "title",
			Message: "Title contains invalid characters",
		})
	}

	// Validate description if present
	if description, ok := req["description"].(string); ok {
		if !v.ValidatePattern("description", description) {
			errors = append(errors, ValidationError{
				Field:   "description",
				Message: "Description contains invalid characters",
			})
		}
	}

	// Validate status if present
	if status, ok := req["status"].(string); ok {
		if !v.ValidatePattern("status", status) {
			errors = append(errors, ValidationError{
				Field:   "status",
				Message: "Status must start with a lowercase letter and contain only lowercase letters, numbers, underscores and hyphens",
			})
		}
	}

	// Validate tags if present
	if tagsInterface, ok := req["tags"].([]interface{}); ok {
		for i, tagInterface := range tagsInterface {
			if tag, ok := tagInterface.(string); ok {
				if !v.ValidatePattern("tag", tag) {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("tags[%d]", i),
						Message: "Tag must start with a lowercase letter and contain only lowercase letters, numbers, underscores, hyphens and colons",
					})
				}
			} else {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("tags[%d]", i),
					Message: "Tag must be a string",
				})
			}
		}
	}

	// Validate properties if present
	if propertiesInterface, ok := req["properties"].(map[string]interface{}); ok {
		for key, value := range propertiesInterface {
			// Validate property key
			if !regexp.MustCompile(`^[a-z][a-z0-9_]{0,29}$`).MatchString(key) {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("properties.%s", key),
					Message: "Property keys must start with a lowercase letter and contain only lowercase letters, numbers and underscores",
				})
			}

			// Validate property values based on type
			switch v := value.(type) {
			case string:
				if len(v) > 1000 {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("properties.%s", key),
						Message: "String property values must be less than 1000 characters",
					})
				}
			case float64:
				// Numbers are valid
			case bool:
				// Booleans are valid
			case []interface{}:
				if len(v) > 100 {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("properties.%s", key),
						Message: "Array property values must have fewer than 100 items",
					})
				}
			case map[string]interface{}:
				if len(v) > 20 {
					errors = append(errors, ValidationError{
						Field:   fmt.Sprintf("properties.%s", key),
						Message: "Object property values must have fewer than 20 properties",
					})
				}
			}
		}
	}

	// Return any validation errors
	if len(errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Entity validation failed",
			"errors":  errors,
		})
		return false
	}

	return true
}

// ValidateEntityUpdate validates an entity update request
func (v *InputValidator) ValidateEntityUpdate(w http.ResponseWriter, req map[string]interface{}) bool {
	// Check for ID field
	id, hasID := req["id"].(string)
	if !hasID {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Entity ID is required for updates",
		})
		return false
	}
	
	// Validate ID pattern
	if !v.ValidatePattern("id", id) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid entity ID format",
		})
		return false
	}
	
	// Validate title if present
	if title, ok := req["title"].(string); ok {
		if !v.ValidatePattern("title", title) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "Invalid title format",
			})
			return false
		}
	}
	
	// Validate description if present
	if description, ok := req["description"].(string); ok {
		if !v.ValidatePattern("description", description) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "error",
				"message": "Invalid description format",
			})
			return false
		}
	}
	
	return true
}

// ValidateRelationshipCreate validates a relationship creation request
func (v *InputValidator) ValidateRelationshipCreate(w http.ResponseWriter, req map[string]interface{}) bool {
	// Check for required fields
	sourceID, hasSource := req["source_id"].(string)
	targetID, hasTarget := req["target_id"].(string)
	relType, hasType := req["type"].(string)
	
	if !hasSource || !hasTarget || !hasType {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Source ID, target ID, and type are required for relationship creation",
		})
		return false
	}
	
	// Validate ID patterns
	if !v.ValidatePattern("id", sourceID) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid source ID format",
		})
		return false
	}
	
	if !v.ValidatePattern("id", targetID) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid target ID format",
		})
		return false
	}
	
	// Validate type pattern
	if !v.ValidatePattern("type", relType) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Invalid relationship type format",
		})
		return false
	}
	
	return true
}

// AuditLogger logs security events
type AuditLogger struct {
	logFile   *os.File
	enabled   bool
	logPath   string
	entityMap map[string]map[string]interface{}
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logPath string, entityMap map[string]map[string]interface{}) (*AuditLogger, error) {
	// Create logger
	logger := &AuditLogger{
		enabled:   true,
		logPath:   logPath,
		entityMap: entityMap,
	}
	
	// Create log directory if it doesn't exist
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		if err := os.MkdirAll(logPath, 0755); err != nil {
			log.Printf("EntityDB Server: Failed to create audit log directory: %v", err)
			return nil, err
		}
	}
	
	// Open log file
	logFilePath := filepath.Join(logPath, fmt.Sprintf("audit_%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("EntityDB Server: Failed to open audit log file: %v", err)
		return nil, err
	}
	
	logger.logFile = logFile
	
	return logger, nil
}

// Close closes the audit logger
func (a *AuditLogger) Close() error {
	if a.logFile != nil {
		return a.logFile.Close()
	}
	return nil
}

// LogEvent logs an event
func (a *AuditLogger) LogEvent(event map[string]interface{}) {
	if !a.enabled || a.logFile == nil {
		return
	}
	
	// Add timestamp
	event["timestamp"] = time.Now().Format(time.RFC3339)
	
	// Convert to JSON
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("EntityDB Server: Failed to marshal audit event: %v", err)
		return
	}
	
	// Write to log file
	if _, err := a.logFile.Write(append(jsonBytes, '\n')); err != nil {
		log.Printf("EntityDB Server: Failed to write audit log: %v", err)
	}
}

// LogAuthEvent logs an authentication event
func (a *AuditLogger) LogAuthEvent(userID, username, action, status, ip string, details map[string]interface{}) {
	event := map[string]interface{}{
		"type":     "auth",
		"user_id":  userID,
		"username": username,
		"action":   action,
		"status":   status,
		"ip":       ip,
	}
	
	// Add details if provided
	if details != nil {
		for key, value := range details {
			event[key] = value
		}
	}
	
	a.LogEvent(event)
}

// LogAccessEvent logs an access control event
func (a *AuditLogger) LogAccessEvent(userID, username, action, status, path string, details map[string]interface{}) {
	event := map[string]interface{}{
		"type":     "access",
		"user_id":  userID,
		"username": username,
		"action":   action,
		"status":   status,
		"path":     path,
	}
	
	// Add details if provided
	if details != nil {
		for key, value := range details {
			event[key] = value
		}
	}
	
	a.LogEvent(event)
}

// LogEntityEvent logs an entity-related event
func (a *AuditLogger) LogEntityEvent(userID, username, entityID, entityType, action, status string, details map[string]interface{}) {
	event := map[string]interface{}{
		"type":        "entity",
		"user_id":     userID,
		"username":    username,
		"entity_id":   entityID,
		"entity_type": entityType,
		"action":      action,
		"status":      status,
	}
	
	// Add details if provided
	if details != nil {
		for key, value := range details {
			event[key] = value
		}
	}
	
	// Add entity details if available
	if a.entityMap != nil {
		if entity, ok := a.entityMap[entityID]; ok {
			// Only include essential entity properties to avoid large log entries
			essentialProps := map[string]interface{}{
				"title": entity["title"],
				"type":  entity["type"],
			}
			event["entity"] = essentialProps
		}
	}
	
	a.LogEvent(event)
}

// LogAdminEvent logs an administrative event
func (a *AuditLogger) LogAdminEvent(userID, username, action, status string, details map[string]interface{}) {
	event := map[string]interface{}{
		"type":     "admin",
		"user_id":  userID,
		"username": username,
		"action":   action,
		"status":   status,
	}
	
	// Add details if provided
	if details != nil {
		for key, value := range details {
			event[key] = value
		}
	}
	
	a.LogEvent(event)
}