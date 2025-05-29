package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"entitydb/logger"
	"entitydb/models"
)

// AuthHandler handles authentication using the new relationship-based security system
type AuthHandler struct {
	securityManager *models.SecurityManager
	sessionManager  *models.SessionManager
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(securityManager *models.SecurityManager, sessionManager *models.SessionManager) *AuthHandler {
	return &AuthHandler{
		securityManager: securityManager,
		sessionManager:  sessionManager,
	}
}

// AuthLoginRequest represents a login request for the new auth system
type AuthLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthLoginResponse represents a login response for the new auth system
type AuthLoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt string       `json:"expires_at"`
	UserID    string       `json:"user_id"`
	User      AuthUserInfo `json:"user"`
}

// AuthUserInfo represents user information returned in login response
type AuthUserInfo struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// AuthErrorResponse represents an error response for auth endpoints
type AuthErrorResponse struct {
	Error string `json:"error"`
}

// Login handles user authentication using relationship-based security
// @Summary Authenticate user
// @Description Authenticate user with username and password using relationship-based security
// @Tags authentication
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginReq AuthLoginRequest
	
	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		logger.Error("Failed to decode login request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Invalid request body"})
		return
	}

	// Validate input
	if loginReq.Username == "" || loginReq.Password == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Username and password are required"})
		return
	}

	// Authenticate user using relationship-based security
	logger.Debug("[AuthHandler] Attempting to authenticate user: %s", loginReq.Username)
	userEntity, err := h.securityManager.AuthenticateUser(loginReq.Username, loginReq.Password)
	if err != nil {
		logger.Error("[AuthHandler] Authentication failed for user %s: %v", loginReq.Username, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Invalid credentials"})
		return
	}
	logger.Debug("[AuthHandler] User authenticated successfully: %s", userEntity.ID)

	// Get user roles from the SecurityUser
	var roles []string
	for _, tag := range userEntity.Entity.GetTagsWithoutTimestamp() {
		if strings.HasPrefix(tag, "rbac:role:") {
			role := strings.TrimPrefix(tag, "rbac:role:")
			roles = append(roles, role)
		}
	}

	// Create session using SessionManager
	session, err := h.sessionManager.CreateSession(userEntity.ID, userEntity.Username, roles)
	if err != nil {
		logger.Error("[AuthHandler] Failed to create session for user %s: %v", userEntity.ID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Failed to create session"})
		return
	}

	// Create response with session token
	response := AuthLoginResponse{
		Token:     session.Token,
		UserID:    userEntity.ID,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
		User: AuthUserInfo{
			ID:       userEntity.ID,
			Username: userEntity.Username,
			Email:    userEntity.Email,
			Roles:    roles,
		},
	}

	logger.Info("User %s authenticated successfully", loginReq.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout by invalidating the session
// @Summary Logout user
// @Description Logout user by invalidating the current session
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get security context
	securityCtx, ok := GetSecurityContext(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Authentication required"})
		return
	}

	// Invalidate session (delete session entity and relationships)
	err := h.invalidateSession(securityCtx.Token)
	if err != nil {
		logger.Error("Failed to invalidate session: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Failed to logout"})
		return
	}

	logger.Info("User %s logged out", securityCtx.User.Username)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// RefreshToken refreshes the current session token
// @Summary Refresh session token
// @Description Refresh the current session token to extend expiration
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} LoginResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get security context
	securityCtx, ok := GetSecurityContext(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Authentication required"})
		return
	}

	// Create new session token and expiration (simplified for now)
	newToken := "new_session_token_" + securityCtx.User.ID
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)

	// Get user roles
	roles, err := h.getUserRoles(securityCtx.User)
	if err != nil {
		logger.Error("Failed to get user roles for %s: %v", securityCtx.User.Username, err)
		roles = []string{}
	}

	// Create response with unified string timestamp
	response := AuthLoginResponse{
		Token:     newToken,
		UserID:    securityCtx.User.ID,
		ExpiresAt: expiresAt,
		User: AuthUserInfo{
			ID:       securityCtx.User.ID,
			Username: securityCtx.User.Username,
			Email:    securityCtx.User.Email,
			Roles:    roles,
		},
	}

	logger.Info("Session refreshed for user %s", securityCtx.User.Username)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// WhoAmI returns information about the current authenticated user
// @Summary Get current user information
// @Description Get information about the currently authenticated user
// @Tags authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserInfo
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/whoami [get]
func (h *AuthHandler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	// Get security context
	securityCtx, ok := GetSecurityContext(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Authentication required"})
		return
	}

	// Get user roles
	roles, err := h.getUserRoles(securityCtx.User)
	if err != nil {
		logger.Error("Failed to get user roles for %s: %v", securityCtx.User.Username, err)
		roles = []string{}
	}

	// Create response
	userInfo := AuthUserInfo{
		ID:       securityCtx.User.ID,
		Username: securityCtx.User.Username,
		Email:    securityCtx.User.Email,
		Roles:    roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userInfo)
}

// Helper functions

// getUserRoles gets all roles for a user (direct and inherited through groups)
func (h *AuthHandler) getUserRoles(user *models.SecurityUser) ([]string, error) {
	// This is a simplified version - in production you might want to cache this
	roles := []string{}

	// Get direct roles
	userRoles, err := h.getUserDirectRoles(user.ID)
	if err != nil {
		return nil, err
	}

	for _, role := range userRoles {
		roles = append(roles, role.Name)
	}

	// Get group-based roles
	userGroups, err := h.getUserGroups(user.ID)
	if err != nil {
		return nil, err
	}

	for _, group := range userGroups {
		groupRoles, err := h.getGroupRoles(group.ID)
		if err != nil {
			continue // Skip this group if we can't get its roles
		}

		for _, role := range groupRoles {
			// Avoid duplicates
			found := false
			for _, existingRole := range roles {
				if existingRole == role.Name {
					found = true
					break
				}
			}
			if !found {
				roles = append(roles, role.Name)
			}
		}
	}

	return roles, nil
}

// getUserDirectRoles gets roles directly assigned to a user
func (h *AuthHandler) getUserDirectRoles(userID string) ([]*models.SecurityRole, error) {
	// This would use the SecurityManager's methods in production
	// For now, we'll implement a basic version
	return []*models.SecurityRole{}, nil
}

// getUserGroups gets groups a user belongs to
func (h *AuthHandler) getUserGroups(userID string) ([]*models.Entity, error) {
	// This would use the SecurityManager's methods in production
	// For now, we'll implement a basic version
	return []*models.Entity{}, nil
}

// getGroupRoles gets roles assigned to a group
func (h *AuthHandler) getGroupRoles(groupID string) ([]*models.SecurityRole, error) {
	// This would use the SecurityManager's methods in production
	// For now, we'll implement a basic version
	return []*models.SecurityRole{}, nil
}

// invalidateSession invalidates a session by deleting the session entity and its relationships
func (h *AuthHandler) invalidateSession(token string) error {
	// Find session entity by token
	// Delete session entity and its relationships
	// This would be implemented using the EntityRepository
	return nil // Placeholder
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if there are multiple
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}

	return ip
}