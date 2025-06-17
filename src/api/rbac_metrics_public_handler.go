package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"entitydb/logger"
	"entitydb/models"
)

// getStringFromContent safely converts byte content to string
func getStringFromContent(content []byte) string {
	if content == nil {
		return ""
	}
	
	// Try to unmarshal as JSON string first
	var str string
	if err := json.Unmarshal(content, &str); err == nil {
		return str
	}
	
	// Otherwise return as raw string
	return string(content)
}

// GetPublicRBACMetrics returns basic RBAC metrics without authentication
// @Summary Get public RBAC metrics
// @Description Get basic RBAC metrics available without authentication
// @Tags rbac
// @Produce json
// @Success 200 {object} PublicRBACMetricsResponse
// @Router /api/v1/rbac/metrics/public [get]
func (h *TemporalRBACMetricsHandler) GetPublicRBACMetrics(w http.ResponseWriter, r *http.Request) {
	// Get basic session count (tag-based)
	activeSessionEntities, err := h.entityRepo.ListByTag("type:session")
	if err != nil {
		logger.Error("Failed to get session entities: %v", err)
		activeSessionEntities = []*models.Entity{} // Default to empty list
	}
	
	// Count non-expired sessions
	activeSessions := 0
	now := time.Now()
	for _, sessionEntity := range activeSessionEntities {
		// Check if session is expired by looking for expires tag
		for _, tag := range sessionEntity.Tags {
			if strings.HasPrefix(tag, "expires:") {
				expiresStr := strings.TrimPrefix(tag, "expires:")
				if expiresTime, err := time.Parse(time.RFC3339, expiresStr); err == nil {
					if now.Before(expiresTime) {
						activeSessions++
					}
				}
				break
			}
		}
	}
	
	// Get user count
	userEntities, _ := h.entityRepo.ListByTag("type:user")
	activeUsers := len(userEntities)
	
	// Get authentication events from temporal storage
	authEventEntities, _ := h.entityRepo.ListByTag("type:auth_event")
	
	successCount := 0
	failureCount := 0
	
	for _, entity := range authEventEntities {
		// Check tags for status (handle temporal tags)
		for _, tag := range entity.Tags {
			// Remove temporal timestamp if present
			cleanTag := tag
			if idx := strings.Index(tag, "|"); idx != -1 {
				cleanTag = tag[idx+1:]
			}
			
			if cleanTag == "status:success" {
				successCount++
				break
			} else if cleanTag == "status:failed" {
				failureCount++
				break
			}
		}
	}
	
	totalLogins := successCount + failureCount
	successRate := 0.0
	if totalLogins > 0 {
		successRate = float64(successCount) / float64(totalLogins) * 100
	}
	
	// Create public response with limited data
	response := PublicRBACMetricsResponse{
		Auth: &SimplifiedAuthMetrics{
			SuccessfulLogins: successCount,
			FailedLogins:     failureCount,
			SuccessRate:      successRate,
		},
		Sessions: &PublicSessionMetrics{
			ActiveCount: activeSessions,
		},
		ActiveUsers: activeUsers,
		Timestamp:   time.Now(),
	}
	
	RespondJSON(w, http.StatusOK, response)
}

// GetAuthenticatedRBACMetrics returns comprehensive metrics for authenticated users
// This is a wrapper around GetRBACMetrics that can be used with less restrictive permissions
func (h *TemporalRBACMetricsHandler) GetAuthenticatedRBACMetrics(w http.ResponseWriter, r *http.Request) {
	// For authenticated users (not just admins), provide full metrics
	h.GetRBACMetricsFromTemporal(w, r)
}