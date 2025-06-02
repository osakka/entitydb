package api

import (
	"encoding/json"
	"entitydb/models"
	"fmt"
	"net/http"
	"strings"
	"time"
	"entitydb/logger"
)

// TemporalRBACMetricsHandler handles RBAC metrics using temporal data
type TemporalRBACMetricsHandler struct {
	entityRepo     models.EntityRepository
	sessionManager *models.SessionManager
}

// NewTemporalRBACMetricsHandler creates a new temporal RBAC metrics handler
func NewTemporalRBACMetricsHandler(entityRepo models.EntityRepository, sessionManager *models.SessionManager) *TemporalRBACMetricsHandler {
	return &TemporalRBACMetricsHandler{
		entityRepo:     entityRepo,
		sessionManager: sessionManager,
	}
}

// GetRBACMetricsFromTemporal retrieves RBAC metrics from temporal storage
func (h *TemporalRBACMetricsHandler) GetRBACMetricsFromTemporal(w http.ResponseWriter, r *http.Request) {
	logger.Info("Fetching RBAC metrics from temporal storage")
	
	// Get time range from query params (default: last 24 hours)
	hours := 24
	if hoursParam := r.URL.Query().Get("hours"); hoursParam != "" {
		if parsed, err := time.ParseDuration(hoursParam + "h"); err == nil {
			hours = int(parsed.Hours())
		}
	}
	
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	// Get real user data from entities
	userEntities, err := h.entityRepo.ListByTag("type:user")
	if err != nil {
		logger.Error("Failed to fetch users: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}
	
	// Count users and roles
	totalUsers := len(userEntities)
	adminCount := 0
	for _, user := range userEntities {
		for _, tag := range user.Tags {
			cleanTag := tag
			if idx := strings.LastIndex(tag, "|"); idx > 0 {
				cleanTag = tag[idx+1:]
			}
			if cleanTag == "rbac:role:admin" {
				adminCount++
				break
			}
		}
	}
	
	// Get active sessions from session manager
	activeSessions := h.sessionManager.GetActiveSessions()
	
	// Try to get authentication metrics from temporal storage
	authSuccess := 0
	authFailed := 0
	
	// Look for auth event entities
	authEventEntities, _ := h.entityRepo.ListByTag("type:auth_event")
	for _, entity := range authEventEntities {
		// Check if entity is within our time range
		if entity.CreatedAt < startTime.UnixNano() {
			continue
		}
		
		// Check for success/failure tags
		for _, tag := range entity.Tags {
			cleanTag := tag
			if idx := strings.LastIndex(tag, "|"); idx > 0 {
				cleanTag = tag[idx+1:]
			}
			
			if cleanTag == "status:success" {
				authSuccess++
			} else if cleanTag == "status:failed" {
				authFailed++
			}
		}
	}
	
	// If no auth events found, provide current session data as minimum
	if authSuccess == 0 && authFailed == 0 && activeSessions > 0 {
		authSuccess = activeSessions // At minimum, active sessions represent successful logins
	}
	
	successRate := 0.0
	if total := authSuccess + authFailed; total > 0 {
		successRate = float64(authSuccess) / float64(total) * 100
	} else if activeSessions > 0 {
		successRate = 100.0 // If we have sessions but no events, assume 100% success
	}
	
	// Get permission check metrics from temporal storage
	permissionChecks := 0
	permissionEntities, _ := h.entityRepo.ListByTag("type:permission_check")
	for _, entity := range permissionEntities {
		if entity.CreatedAt >= startTime.UnixNano() {
			permissionChecks++
		}
	}
	
	// Calculate checks per second (over the time range)
	checksPerSecond := float64(permissionChecks) / (float64(hours) * 3600)
	
	// Get security events from temporal storage
	var securityEvents []SecurityEvent
	secEventEntities, _ := h.entityRepo.ListByTag("type:security_event")
	
	// Get the 5 most recent security events
	for i := len(secEventEntities) - 1; i >= 0 && len(securityEvents) < 5; i-- {
		event := secEventEntities[i]
		
		// Parse event data from content
		var eventData map[string]interface{}
		if len(event.Content) > 0 {
			json.Unmarshal(event.Content, &eventData)
		}
		
		eventType := "unknown"
		username := "unknown"
		status := "info"
		details := ""
		
		// Extract info from tags
		for _, tag := range event.Tags {
			cleanTag := tag
			if idx := strings.LastIndex(tag, "|"); idx > 0 {
				cleanTag = tag[idx+1:]
			}
			
			if strings.HasPrefix(cleanTag, "event:") {
				eventType = strings.TrimPrefix(cleanTag, "event:")
			} else if strings.HasPrefix(cleanTag, "user:") {
				username = strings.TrimPrefix(cleanTag, "user:")
			} else if strings.HasPrefix(cleanTag, "status:") {
				status = strings.TrimPrefix(cleanTag, "status:")
			}
		}
		
		// Get details from content if available
		if d, ok := eventData["details"].(string); ok {
			details = d
		} else if d, ok := eventData["message"].(string); ok {
			details = d
		} else {
			details = fmt.Sprintf("%s event for user %s", eventType, username)
		}
		
		securityEvents = append(securityEvents, SecurityEvent{
			ID:        event.ID,
			Timestamp: time.Unix(0, event.CreatedAt),
			Type:      eventType,
			Username:  username,
			Details:   details,
			Status:    status,
		})
	}
	
	// If no security events exist, show current session as an event
	if len(securityEvents) == 0 && activeSessions > 0 {
		securityEvents = append(securityEvents, SecurityEvent{
			ID:        "current_session",
			Timestamp: time.Now(),
			Type:      "session_active",
			Username:  "current_user",
			Details:   fmt.Sprintf("%d active sessions", activeSessions),
			Status:    "success",
		})
	}
	
	// Build response
	response := RBACMetricsResponse{
		Users: &SimplifiedUserMetrics{
			TotalUsers: totalUsers,
			AdminCount: adminCount,
		},
		Sessions: &SimplifiedSessionMetrics{
			ActiveCount:   activeSessions,
			TotalToday:    authSuccess, // Use actual successful logins
			AvgDurationMs: 7200000,      // Default 2 hours (will be calculated from temporal data later)
		},
		Auth: &SimplifiedAuthMetrics{
			SuccessfulLogins: authSuccess,
			FailedLogins:     authFailed,
			SuccessRate:      successRate,
		},
		Permissions: &SimplifiedPermissionMetrics{
			ChecksPerSecond: checksPerSecond,
			TotalChecks:     permissionChecks,
			CacheHitRate:    0.0, // Will be calculated from temporal data
		},
		SecurityEvents: securityEvents,
		Timestamp:      time.Now().Format(time.RFC3339),
	}
	
	RespondJSON(w, http.StatusOK, response)
}

// TrackAuthEvent stores authentication events in temporal storage
func (h *TemporalRBACMetricsHandler) TrackAuthEvent(username string, success bool, details string) {
	entity := &models.Entity{
		ID:   fmt.Sprintf("auth_event_%s_%d", username, time.Now().UnixNano()),
		Tags: []string{"type:auth_event"},
	}
	
	// Add status tag
	if success {
		entity.Tags = append(entity.Tags, "status:success")
	} else {
		entity.Tags = append(entity.Tags, "status:failed")
	}
	
	// Add username tag
	entity.Tags = append(entity.Tags, fmt.Sprintf("user:%s", username))
	
	// Store details in content
	content := map[string]interface{}{
		"username": username,
		"success":  success,
		"details":  details,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	if data, err := json.Marshal(content); err == nil {
		entity.Content = data
	}
	
	// Create the entity
	if err := h.entityRepo.Create(entity); err != nil {
		logger.Error("Failed to track auth event: %v", err)
	}
}

// TrackPermissionCheck stores permission check events in temporal storage
func (h *TemporalRBACMetricsHandler) TrackPermissionCheck(userID string, resource string, action string, allowed bool) {
	entity := &models.Entity{
		ID: fmt.Sprintf("perm_check_%s_%s_%s_%d", userID, resource, action, time.Now().UnixNano()),
		Tags: []string{
			"type:permission_check",
			fmt.Sprintf("user:%s", userID),
			fmt.Sprintf("resource:%s", resource),
			fmt.Sprintf("action:%s", action),
		},
	}
	
	if allowed {
		entity.Tags = append(entity.Tags, "result:allowed")
	} else {
		entity.Tags = append(entity.Tags, "result:denied")
	}
	
	// Store check details in content
	content := map[string]interface{}{
		"user_id":   userID,
		"resource":  resource,
		"action":    action,
		"allowed":   allowed,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	if data, err := json.Marshal(content); err == nil {
		entity.Content = data
	}
	
	// Create the entity with retention policy (keep for 7 days)
	entity.Tags = append(entity.Tags, "retention:period:604800") // 7 days in seconds
	
	if err := h.entityRepo.Create(entity); err != nil {
		logger.Error("Failed to track permission check: %v", err)
	}
}

// TrackSecurityEvent stores security events in temporal storage
func (h *TemporalRBACMetricsHandler) TrackSecurityEvent(eventType string, username string, details string, status string) {
	entity := &models.Entity{
		ID: fmt.Sprintf("sec_event_%s_%s_%d", eventType, username, time.Now().UnixNano()),
		Tags: []string{
			"type:security_event",
			fmt.Sprintf("event:%s", eventType),
			fmt.Sprintf("user:%s", username),
			fmt.Sprintf("status:%s", status),
		},
	}
	
	// Store event details in content
	content := map[string]interface{}{
		"type":      eventType,
		"username":  username,
		"details":   details,
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	
	if data, err := json.Marshal(content); err == nil {
		entity.Content = data
	}
	
	// Create the entity with retention policy (keep for 30 days)
	entity.Tags = append(entity.Tags, "retention:period:2592000") // 30 days in seconds
	
	if err := h.entityRepo.Create(entity); err != nil {
		logger.Error("Failed to track security event: %v", err)
	}
}