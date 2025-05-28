package api

import (
	"entitydb/models"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// RBACMetricsHandler handles RBAC and session metrics
type RBACMetricsHandler struct {
	entityRepo     models.EntityRepository
	sessionManager *models.SessionManager
}

// NewRBACMetricsHandler creates a new RBAC metrics handler
func NewRBACMetricsHandler(entityRepo models.EntityRepository, sessionManager *models.SessionManager) *RBACMetricsHandler {
	return &RBACMetricsHandler{
		entityRepo:     entityRepo,
		sessionManager: sessionManager,
	}
}

// RBACMetricsResponse represents the complete RBAC metrics response
type RBACMetricsResponse struct {
	Users           *SimplifiedUserMetrics        `json:"users"`
	Sessions        *SimplifiedSessionMetrics     `json:"sessions"`
	Auth            *SimplifiedAuthMetrics        `json:"auth"`
	Permissions     *SimplifiedPermissionMetrics  `json:"permissions"`
	SecurityEvents  []SecurityEvent               `json:"security_events"`
	Timestamp       string                        `json:"timestamp"`
}

// SimplifiedUserMetrics for frontend compatibility
type SimplifiedUserMetrics struct {
	TotalUsers  int `json:"total_users"`
	AdminCount  int `json:"admin_count"`
}

// SimplifiedSessionMetrics for frontend compatibility  
type SimplifiedSessionMetrics struct {
	ActiveCount    int     `json:"active_count"`
	TotalToday     int     `json:"total_today"`
	AvgDurationMs  float64 `json:"avg_duration_ms"`
}

// SimplifiedAuthMetrics for frontend compatibility
type SimplifiedAuthMetrics struct {
	SuccessfulLogins int     `json:"successful_logins"`
	FailedLogins     int     `json:"failed_logins"`
	SuccessRate      float64 `json:"success_rate"`
}

// SimplifiedPermissionMetrics for frontend compatibility
type SimplifiedPermissionMetrics struct {
	ChecksPerSecond float64 `json:"checks_per_second"`
	TotalChecks     int     `json:"total_checks"`
	CacheHitRate    float64 `json:"cache_hit_rate"`
}

// SecurityEvent represents a security event
type SecurityEvent struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	Details   string    `json:"details"`
	Status    string    `json:"status"`
}

// UserMetrics contains user-related statistics
type UserMetrics struct {
	TotalUsers    int                    `json:"total_users"`
	ActiveUsers   int                    `json:"active_users"`
	AdminUsers    int                    `json:"admin_users"`
	RegularUsers  int                    `json:"regular_users"`
	UsersByRole   map[string]int         `json:"users_by_role"`
	UserCreation  []UserCreationPoint    `json:"user_creation_timeline"`
}

// RoleMetrics contains role and permission statistics
type RoleMetrics struct {
	TotalRoles       int                       `json:"total_roles"`
	RoleDistribution map[string]int            `json:"role_distribution"`
	Permissions      map[string]PermissionInfo `json:"permissions"`
}

// SessionMetrics contains session-related statistics
type SessionMetrics struct {
	TotalActiveSessions   int                   `json:"total_active_sessions"`
	SessionsPerUser       map[string]int        `json:"sessions_per_user"`
	AverageSessionTime    float64               `json:"average_session_duration_minutes"`
	PeakConcurrentSessions int                  `json:"peak_concurrent_sessions"`
	SessionTimeline       []SessionTimePoint    `json:"session_timeline"`
	SessionsByDuration    map[string]int        `json:"sessions_by_duration"`
}

// AuthenticationMetrics contains authentication success/failure statistics
type AuthenticationMetrics struct {
	TotalAttempts    int                    `json:"total_attempts"`
	SuccessfulLogins int                    `json:"successful_logins"`
	FailedLogins     int                    `json:"failed_logins"`
	SuccessRate      float64                `json:"success_rate"`
	Timeline         []AuthTimePoint        `json:"timeline"`
	FailureReasons   map[string]int         `json:"failure_reasons"`
}

// SessionInfo represents an active session
type SessionInfo struct {
	Token      string    `json:"token"`
	Username   string    `json:"username"`
	UserID     string    `json:"user_id"`
	Roles      []string  `json:"roles"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsed   time.Time `json:"last_used"`
	ExpiresAt  time.Time `json:"expires_at"`
	Duration   string    `json:"duration"`
	Status     string    `json:"status"`
}

// ActivityInfo represents recent user activity
type ActivityInfo struct {
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details"`
}

// UserCreationPoint represents user creation over time
type UserCreationPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// SessionTimePoint represents session activity over time
type SessionTimePoint struct {
	Time     string `json:"time"`
	Created  int    `json:"created"`
	Expired  int    `json:"expired"`
	Active   int    `json:"active"`
}

// AuthTimePoint represents authentication attempts over time
type AuthTimePoint struct {
	Time     string `json:"time"`
	Success  int    `json:"success"`
	Failed   int    `json:"failed"`
}

// PermissionInfo contains permission usage statistics
type PermissionInfo struct {
	UsageCount int      `json:"usage_count"`
	Users      []string `json:"users"`
}

// RBACMetricsSummary contains high-level summary statistics
type RBACMetricsSummary struct {
	SecurityScore     float64 `json:"security_score"`
	ActiveSessionRate float64 `json:"active_session_rate"`
	AuthSuccessRate   float64 `json:"auth_success_rate"`
	AdminRatio        float64 `json:"admin_ratio"`
}

// GetRBACMetrics returns comprehensive RBAC metrics
// @Summary Get RBAC metrics
// @Description Get comprehensive RBAC, session, and authentication metrics
// @Tags rbac
// @Produce json
// @Success 200 {object} RBACMetricsResponse
// @Router /api/v1/rbac/metrics [get]
func (h *RBACMetricsHandler) GetRBACMetrics(w http.ResponseWriter, r *http.Request) {
	// Get all users
	userEntities, err := h.entityRepo.ListByTag("type:user")
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to fetch users")
		return
	}

	// Get session metrics from session manager
	activeSessions := h.getActiveSessionsInfo()
	
	// Calculate user metrics
	userMetrics := h.calculateUserMetrics(userEntities)
	
	// Calculate session metrics
	sessionMetrics := h.calculateSessionMetrics(activeSessions)
	
	// Generate mock authentication metrics (since we don't track this yet)
	authMetrics := h.generateAuthenticationMetrics()

	// Create simplified response that matches frontend expectations
	response := RBACMetricsResponse{
		Users: &SimplifiedUserMetrics{
			TotalUsers: userMetrics.TotalUsers,
			AdminCount: userMetrics.AdminUsers,
		},
		Sessions: &SimplifiedSessionMetrics{
			ActiveCount:   h.sessionManager.GetActiveSessions(),
			TotalToday:    sessionMetrics.TotalActiveSessions * 10, // Mock data
			AvgDurationMs: sessionMetrics.AverageSessionTime * 60 * 1000, // Convert to milliseconds
		},
		Auth: &SimplifiedAuthMetrics{
			SuccessfulLogins: authMetrics.SuccessfulLogins,
			FailedLogins:     authMetrics.FailedLogins,
			SuccessRate:      authMetrics.SuccessRate,
		},
		Permissions: &SimplifiedPermissionMetrics{
			ChecksPerSecond: 42.5, // Mock data
			TotalChecks:     150000, // Mock data  
			CacheHitRate:    0.95, // 95% cache hit rate
		},
		SecurityEvents: h.generateSecurityEvents(),
		Timestamp:      time.Now().Format(time.RFC3339),
	}

	RespondJSON(w, http.StatusOK, response)
}

// getActiveSessionsInfo extracts detailed information about active sessions
func (h *RBACMetricsHandler) getActiveSessionsInfo() []SessionInfo {
	var sessions []SessionInfo
	
	// Use reflection to access session manager's sessions map
	// This is a simplified version - in production you'd add a method to SessionManager
	activeSessions := h.sessionManager.GetActiveSessions()
	
	// For now, generate representative session data
	// In a real implementation, you'd add methods to SessionManager to expose session details
	for i := 0; i < activeSessions; i++ {
		session := SessionInfo{
			Token:     "token_" + generateShortID(),
			Username:  "user" + generateShortID(),
			UserID:    "user_" + generateShortID(),
			Roles:     []string{"user"},
			CreatedAt: time.Now().Add(-time.Duration(i*30) * time.Minute),
			LastUsed:  time.Now().Add(-time.Duration(i*5) * time.Minute),
			ExpiresAt: time.Now().Add(2 * time.Hour),
			Duration:  formatDuration(time.Duration(i*30) * time.Minute),
			Status:    "active",
		}
		sessions = append(sessions, session)
	}
	
	return sessions
}

// calculateUserMetrics analyzes user entities to extract user statistics
func (h *RBACMetricsHandler) calculateUserMetrics(userEntities []*models.Entity) UserMetrics {
	totalUsers := len(userEntities)
	adminUsers := 0
	regularUsers := 0
	usersByRole := make(map[string]int)
	
	for _, user := range userEntities {
		isAdmin := false
		for _, tag := range user.Tags {
			cleanTag := strings.Split(tag, "|")[0] // Remove timestamp if present
			if cleanTag == "rbac:role:admin" {
				adminUsers++
				isAdmin = true
				usersByRole["admin"]++
				break
			}
		}
		if !isAdmin {
			regularUsers++
			usersByRole["user"]++
		}
	}
	
	// Generate user creation timeline (mock data for now)
	userCreation := generateUserCreationTimeline(totalUsers)
	
	return UserMetrics{
		TotalUsers:   totalUsers,
		ActiveUsers:  totalUsers, // Assume all users are active for now
		AdminUsers:   adminUsers,
		RegularUsers: regularUsers,
		UsersByRole:  usersByRole,
		UserCreation: userCreation,
	}
}

// calculateRoleMetrics analyzes roles and permissions
func (h *RBACMetricsHandler) calculateRoleMetrics(userEntities []*models.Entity) RoleMetrics {
	roleDistribution := make(map[string]int)
	permissions := make(map[string]PermissionInfo)
	
	for _, user := range userEntities {
		for _, tag := range user.Tags {
			cleanTag := strings.Split(tag, "|")[0] // Remove timestamp if present
			
			if strings.HasPrefix(cleanTag, "rbac:role:") {
				role := strings.TrimPrefix(cleanTag, "rbac:role:")
				roleDistribution[role]++
			}
			
			if strings.HasPrefix(cleanTag, "rbac:perm:") {
				perm := strings.TrimPrefix(cleanTag, "rbac:perm:")
				if permInfo, exists := permissions[perm]; exists {
					permInfo.UsageCount++
					permInfo.Users = append(permInfo.Users, user.ID)
					permissions[perm] = permInfo
				} else {
					permissions[perm] = PermissionInfo{
						UsageCount: 1,
						Users:      []string{user.ID},
					}
				}
			}
		}
	}
	
	return RoleMetrics{
		TotalRoles:       len(roleDistribution),
		RoleDistribution: roleDistribution,
		Permissions:      permissions,
	}
}

// calculateSessionMetrics analyzes session data
func (h *RBACMetricsHandler) calculateSessionMetrics(sessions []SessionInfo) SessionMetrics {
	totalSessions := len(sessions)
	sessionsPerUser := make(map[string]int)
	sessionsByDuration := map[string]int{
		"0-30min":   0,
		"30-60min":  0,
		"1-2hours":  0,
		"2-6hours":  0,
		"6+ hours":  0,
	}
	
	var totalDuration time.Duration
	
	for _, session := range sessions {
		sessionsPerUser[session.Username]++
		
		duration := time.Since(session.CreatedAt)
		totalDuration += duration
		
		// Categorize by duration
		minutes := int(duration.Minutes())
		if minutes <= 30 {
			sessionsByDuration["0-30min"]++
		} else if minutes <= 60 {
			sessionsByDuration["30-60min"]++
		} else if minutes <= 120 {
			sessionsByDuration["1-2hours"]++
		} else if minutes <= 360 {
			sessionsByDuration["2-6hours"]++
		} else {
			sessionsByDuration["6+ hours"]++
		}
	}
	
	avgDuration := 0.0
	if totalSessions > 0 {
		avgDuration = totalDuration.Minutes() / float64(totalSessions)
	}
	
	// Generate session timeline
	timeline := generateSessionTimeline()
	
	return SessionMetrics{
		TotalActiveSessions:    totalSessions,
		SessionsPerUser:        sessionsPerUser,
		AverageSessionTime:     avgDuration,
		PeakConcurrentSessions: totalSessions, // Simplified
		SessionTimeline:        timeline,
		SessionsByDuration:     sessionsByDuration,
	}
}

// generateAuthenticationMetrics creates authentication statistics
func (h *RBACMetricsHandler) generateAuthenticationMetrics() AuthenticationMetrics {
	// This is mock data - in a real implementation, you'd track authentication attempts
	successful := 150
	failed := 12
	total := successful + failed
	successRate := float64(successful) / float64(total) * 100
	
	timeline := []AuthTimePoint{
		{Time: "00:00", Success: 8, Failed: 1},
		{Time: "04:00", Success: 2, Failed: 0},
		{Time: "08:00", Success: 25, Failed: 3},
		{Time: "12:00", Success: 30, Failed: 2},
		{Time: "16:00", Success: 28, Failed: 4},
		{Time: "20:00", Success: 15, Failed: 2},
	}
	
	failureReasons := map[string]int{
		"invalid_password": 8,
		"user_not_found":   2,
		"expired_session":  2,
	}
	
	return AuthenticationMetrics{
		TotalAttempts:    total,
		SuccessfulLogins: successful,
		FailedLogins:     failed,
		SuccessRate:      successRate,
		Timeline:         timeline,
		FailureReasons:   failureReasons,
	}
}

// generateRecentActivity creates recent activity log
func (h *RBACMetricsHandler) generateRecentActivity(sessions []SessionInfo) []ActivityInfo {
	var activities []ActivityInfo
	
	for i, session := range sessions {
		if i >= 10 { // Limit to 10 recent activities
			break
		}
		
		activity := ActivityInfo{
			Username:  session.Username,
			Action:    "login",
			Timestamp: session.CreatedAt,
			Details:   "User authenticated successfully",
		}
		activities = append(activities, activity)
	}
	
	return activities
}

// calculateSummary generates high-level summary metrics
func (h *RBACMetricsHandler) calculateSummary(users UserMetrics, sessions SessionMetrics, auth AuthenticationMetrics) RBACMetricsSummary {
	securityScore := auth.SuccessRate * 0.7 + 30.0 // Simplified security score
	activeSessionRate := float64(sessions.TotalActiveSessions) / float64(users.TotalUsers) * 100
	adminRatio := float64(users.AdminUsers) / float64(users.TotalUsers) * 100
	
	return RBACMetricsSummary{
		SecurityScore:     securityScore,
		ActiveSessionRate: activeSessionRate,
		AuthSuccessRate:   auth.SuccessRate,
		AdminRatio:        adminRatio,
	}
}

// generateSecurityEvents creates mock security events
func (h *RBACMetricsHandler) generateSecurityEvents() []SecurityEvent {
	events := []SecurityEvent{
		{
			ID:        "evt_001",
			Timestamp: time.Now().Add(-2 * time.Hour),
			Type:      "login_success",
			Username:  "admin",
			Details:   "Successful login from IP 192.168.1.100",
			Status:    "success",
		},
		{
			ID:        "evt_002",
			Timestamp: time.Now().Add(-90 * time.Minute),
			Type:      "login_failed",
			Username:  "unknown",
			Details:   "Failed login attempt - invalid credentials",
			Status:    "failed",
		},
		{
			ID:        "evt_003",
			Timestamp: time.Now().Add(-45 * time.Minute),
			Type:      "permission_denied",
			Username:  "user123",
			Details:   "Access denied to admin resources",
			Status:    "blocked",
		},
		{
			ID:        "evt_004",
			Timestamp: time.Now().Add(-30 * time.Minute),
			Type:      "session_expired",
			Username:  "developer",
			Details:   "Session expired after 2 hours",
			Status:    "info",
		},
		{
			ID:        "evt_005",
			Timestamp: time.Now().Add(-15 * time.Minute),
			Type:      "login_success",
			Username:  "user456",
			Details:   "Successful login from mobile device",
			Status:    "success",
		},
	}
	
	return events
}

// Helper functions

func generateShortID() string {
	return time.Now().Format("0405")
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func generateUserCreationTimeline(totalUsers int) []UserCreationPoint {
	points := []UserCreationPoint{
		{Date: "2025-05-18", Count: 1},
		{Date: "2025-05-19", Count: 0},
		{Date: "2025-05-20", Count: 0},
		{Date: "2025-05-21", Count: 0},
		{Date: "2025-05-22", Count: 0},
		{Date: "2025-05-23", Count: 0},
		{Date: "2025-05-24", Count: 0},
		{Date: "2025-05-25", Count: totalUsers - 1}, // All other users created today
	}
	return points
}

func generateSessionTimeline() []SessionTimePoint {
	points := []SessionTimePoint{
		{Time: "00:00", Created: 2, Expired: 1, Active: 1},
		{Time: "04:00", Created: 0, Expired: 0, Active: 1},
		{Time: "08:00", Created: 5, Expired: 1, Active: 5},
		{Time: "12:00", Created: 8, Expired: 2, Active: 11},
		{Time: "16:00", Created: 6, Expired: 3, Active: 14},
		{Time: "20:00", Created: 3, Expired: 4, Active: 13},
	}
	return points
}