package api

import (
	"entitydb/models"
	"entitydb/logger"
)

// AuthEventTracker tracks authentication events
type AuthEventTracker struct {
	rbacMetricsHandler *TemporalRBACMetricsHandler
}

// NewAuthEventTracker creates a new auth event tracker
func NewAuthEventTracker(entityRepo models.EntityRepository, sessionManager *models.SessionManager) *AuthEventTracker {
	return &AuthEventTracker{
		rbacMetricsHandler: NewTemporalRBACMetricsHandler(entityRepo, sessionManager),
	}
}

// TrackLoginSuccess tracks successful login
func (t *AuthEventTracker) TrackLoginSuccess(username string, userID string) {
	logger.Info("Tracking successful login for user %s", username)
	t.rbacMetricsHandler.TrackAuthEvent(username, true, "Successful login")
	t.rbacMetricsHandler.TrackSecurityEvent("login_success", username, "User authenticated successfully", "success")
}

// TrackLoginFailure tracks failed login
func (t *AuthEventTracker) TrackLoginFailure(username string, reason string) {
	logger.Warn("Tracking failed login for user %s: %s", username, reason)
	t.rbacMetricsHandler.TrackAuthEvent(username, false, reason)
	t.rbacMetricsHandler.TrackSecurityEvent("login_failed", username, reason, "failed")
}

// TrackPermissionCheck tracks permission checks
func (t *AuthEventTracker) TrackPermissionCheck(userID string, resource string, action string, allowed bool) {
	t.rbacMetricsHandler.TrackPermissionCheck(userID, resource, action, allowed)
	
	// Track significant denials as security events
	if !allowed {
		t.rbacMetricsHandler.TrackSecurityEvent("permission_denied", userID, 
			"Access denied to "+resource+":"+action, "blocked")
	}
}

// TrackSessionExpired tracks session expiration
func (t *AuthEventTracker) TrackSessionExpired(username string, sessionID string) {
	logger.Info("Tracking session expiration for user %s", username)
	t.rbacMetricsHandler.TrackSecurityEvent("session_expired", username, 
		"Session expired after timeout", "info")
}