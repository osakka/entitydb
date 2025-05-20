package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

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