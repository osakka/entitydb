package api

import (
	"net/http"
	"time"
)

// PublicRBACMetricsResponse represents basic metrics available without authentication
type PublicRBACMetricsResponse struct {
	Auth      *SimplifiedAuthMetrics `json:"auth"`
	Sessions  *PublicSessionMetrics  `json:"sessions"`
	Timestamp string                 `json:"timestamp"`
}

// PublicSessionMetrics contains public session information
type PublicSessionMetrics struct {
	ActiveCount int `json:"active_count"`
}

// GetPublicRBACMetrics returns basic RBAC metrics without authentication
// @Summary Get public RBAC metrics
// @Description Get basic RBAC metrics available without authentication
// @Tags rbac
// @Produce json
// @Success 200 {object} PublicRBACMetricsResponse
// @Router /api/v1/rbac/metrics/public [get]
func (h *RBACMetricsHandler) GetPublicRBACMetrics(w http.ResponseWriter, r *http.Request) {
	// Get basic session count
	activeSessions := h.sessionManager.GetActiveSessions()
	
	// Generate basic authentication metrics
	authMetrics := h.generateAuthenticationMetrics()
	
	// Create public response with limited data
	response := PublicRBACMetricsResponse{
		Auth: &SimplifiedAuthMetrics{
			SuccessfulLogins: authMetrics.SuccessfulLogins,
			FailedLogins:     authMetrics.FailedLogins,
			SuccessRate:      authMetrics.SuccessRate,
		},
		Sessions: &PublicSessionMetrics{
			ActiveCount: activeSessions,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}
	
	RespondJSON(w, http.StatusOK, response)
}

// GetAuthenticatedRBACMetrics returns comprehensive metrics for authenticated users
// This is a wrapper around GetRBACMetrics that can be used with less restrictive permissions
func (h *RBACMetricsHandler) GetAuthenticatedRBACMetrics(w http.ResponseWriter, r *http.Request) {
	// For authenticated users (not just admins), provide full metrics
	h.GetRBACMetrics(w, r)
}