// Package api provides HTTP handlers for the EntityDB REST API.
// This file implements authentication endpoints.
package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"entitydb/logger"
	"entitydb/models"
)

// AuthHandler handles authentication endpoints for EntityDB.
// It manages user login, logout, and session validation using the embedded credential system.
//
// As of v2.29.0, user credentials are stored directly in the user entity's content field
// in the format "salt|bcrypt_hash". This eliminates the need for separate credential
// entities and relationships.
//
// Key responsibilities:
//   - User authentication via username/password
//   - Session token generation and management
//   - Session refresh and logout functionality
//   - Integration with RBAC for permission checking
type AuthHandler struct {
	securityManager *models.SecurityManager
}

// NewAuthHandler creates a new authentication handler.
// Parameters:
//   - securityManager: Handles password verification, user authentication, and session management
func NewAuthHandler(securityManager *models.SecurityManager) *AuthHandler {
	return &AuthHandler{
		securityManager: securityManager,
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

// Login handles user authentication using the embedded credential system.
//
// HTTP Method: POST
// Endpoint: /api/v1/auth/login
// Required Permission: None (public endpoint)
//
// Request Body:
//   {
//     "username": "admin",
//     "password": "admin"
//   }
//
// Response:
//   200 OK: Authentication successful
//   {
//     "token": "generated-session-token",
//     "expires_at": "2024-01-01T12:00:00Z",
//     "user_id": "user-entity-id",
//     "user": {
//       "id": "user-entity-id",
//       "username": "admin",
//       "email": "admin@example.com",
//       "roles": ["admin", "user"]
//     }
//   }
//
// Error Responses:
//   - 400 Bad Request: Invalid request body or missing credentials
//   - 401 Unauthorized: Invalid username or password
//   - 500 Internal Server Error: Failed to create session
//
// Authentication Flow:
//   1. Validates username and password are provided
//   2. Looks up user entity by username tag
//   3. Verifies password against embedded bcrypt hash in entity content
//   4. Creates a new session with TTL (default 1 hour)
//   5. Returns session token and user information
//
// Security Notes:
//   - Passwords are hashed using bcrypt with cost 10
//   - Session tokens are generated using crypto/rand
//   - Failed login attempts are tracked (currently disabled due to deadlock)
//   - Sessions expire after TTL and are automatically cleaned up
//
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
		logger.Warn("failed to decode login request: %v", err)
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
	logger.Info("authentication attempt for user %s", loginReq.Username)
	logger.TraceIf("auth", "calling AuthenticateUser")
	userEntity, err := h.securityManager.AuthenticateUser(loginReq.Username, loginReq.Password)
	logger.TraceIf("auth", "AuthenticateUser returned")
	if err != nil {
		logger.Warn("authentication failed for user %s: %v", loginReq.Username, err)
		// Track error asynchronously - no longer causes hangs
		TrackHTTPError("auth_handler.Login", http.StatusUnauthorized, err)
		
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Invalid credentials"})
		return
	}
	logger.Info("user %s authenticated successfully", loginReq.Username)
	

	// Extract user roles from entity tags
	// Roles are stored as tags with the format "rbac:role:rolename"
	var roles []string
	for _, tag := range userEntity.Entity.GetTagsWithoutTimestamp() {
		if strings.HasPrefix(tag, "rbac:role:") {
			role := strings.TrimPrefix(tag, "rbac:role:")
			roles = append(roles, role)
		}
	}

	// Create session in database (this is what security middleware uses)
	ipAddress := r.RemoteAddr
	userAgent := r.Header.Get("User-Agent")
	logger.TraceIf("auth", "creating database session for user %s", userEntity.ID)
	dbSession, err := h.securityManager.CreateSession(userEntity, ipAddress, userAgent)
	if err != nil {
		logger.Error("failed to create database session for user %s: %v", userEntity.ID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Failed to create session"})
		return
	}
	logger.TraceIf("auth", "created database session for user %s", userEntity.ID)

	// Create response with session token
	response := AuthLoginResponse{
		Token:     dbSession.Token,
		UserID:    userEntity.ID,
		ExpiresAt: dbSession.ExpiresAt.Format(time.RFC3339),
		User: AuthUserInfo{
			ID:       userEntity.ID,
			Username: userEntity.Username,
			Email:    userEntity.Email,
			Roles:    roles,
		},
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout by invalidating the session.
//
// HTTP Method: POST
// Endpoint: /api/v1/auth/logout
// Required Permission: None (but requires valid session token)
//
// Headers:
//   Authorization: Bearer <session-token>
//
// Response:
//   200 OK: Successfully logged out
//   {
//     "message": "Logged out successfully"
//   }
//
// Error Responses:
//   - 401 Unauthorized: No token provided or invalid token format
//   - 500 Internal Server Error: Failed to invalidate session
//
// Logout Flow:
//   1. Extracts session token from Authorization header
//   2. Validates token format (Bearer scheme)
//   3. Invalidates the session in session manager
//   4. Returns success message
//
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
	// Get security context (v2.32.0+ modern RBAC - token already validated by SecurityMiddleware)
	securityCtx, ok := GetSecurityContext(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Authentication required"})
		return
	}
	
	token := securityCtx.Token
	
	// Invalidate session in database
	err := h.securityManager.InvalidateSession(token)
	if err != nil {
		logger.Error("failed to invalidate session: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Failed to logout"})
		return
	}

	logger.Debug("session invalidated successfully")
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
	logger.Info("RefreshToken: Method called from IP %s", r.RemoteAddr)
	
	// Get security context (v2.32.0+ modern RBAC - token already validated by SecurityMiddleware)
	securityCtx, ok := GetSecurityContext(r)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Authentication required"})
		return
	}
	
	currentToken := securityCtx.Token

	// Refresh the session in the database (this validates the token internally)
	logger.TraceIf("auth", "refreshing session with token")
	newSession, err := h.securityManager.RefreshSession(currentToken)
	if err != nil {
		logger.Error("failed to refresh session: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Failed to refresh session"})
		return
	}

	// Get user information from the session
	logger.Info("RefreshToken: Attempting to get user entity for UserID: %s", newSession.UserID)
	if newSession.UserID == "" {
		logger.Error("RefreshToken: UserID is empty in refreshed session")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Session missing user information"})
		return
	}
	
	userEntity, err := h.securityManager.GetEntityRepo().GetByID(newSession.UserID)
	if err != nil {
		logger.Error("failed to get user entity for session (UserID: %s): %v", newSession.UserID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AuthErrorResponse{Error: "Failed to get user information"})
		return
	}
	logger.Info("RefreshToken: Successfully got user entity with %d tags", len(userEntity.Tags))

	// Extract user information from entity
	username := ""
	email := ""
	var roles []string
	
	cleanTags := userEntity.GetTagsWithoutTimestamp()
	logger.Info("RefreshToken: Clean user entity tags: %v", cleanTags)
	
	for _, tag := range cleanTags {
		if strings.HasPrefix(tag, "identity:username:") {
			username = strings.TrimPrefix(tag, "identity:username:")
			logger.Info("RefreshToken: Extracted username: %s", username)
		} else if strings.HasPrefix(tag, "profile:email:") {
			email = strings.TrimPrefix(tag, "profile:email:")
			logger.Info("RefreshToken: Extracted email: %s", email)
		} else if strings.HasPrefix(tag, "rbac:role:") {
			role := strings.TrimPrefix(tag, "rbac:role:")
			roles = append(roles, role)
			logger.Info("RefreshToken: Extracted role: %s", role)
		}
	}
	
	// If we didn't extract any user data, there might be an issue with tag format
	if username == "" && email == "" {
		logger.Warn("RefreshToken: No user data extracted from %d clean tags. Raw tags: %v", len(cleanTags), cleanTags)
	}
	
	logger.Info("RefreshToken: Final extracted data - Username: %s, Email: %s, Roles: %v", username, email, roles)

	// Create response with refreshed session
	response := AuthLoginResponse{
		Token:     newSession.Token,
		UserID:    newSession.UserID,
		ExpiresAt: newSession.ExpiresAt.Format(time.RFC3339),
		User: AuthUserInfo{
			ID:       newSession.UserID,
			Username: username,
			Email:    email,
			Roles:    roles,
		},
	}

	logger.Info("Session refreshed for user %s", username)
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

// getUserRoles gets all roles for a user via tag-based RBAC
func (h *AuthHandler) getUserRoles(user *models.SecurityUser) ([]string, error) {
	var roles []string
	userTags := user.Entity.GetTagsWithoutTimestamp()
	
	// Extract roles from rbac:role: tags
	for _, tag := range userTags {
		if strings.HasPrefix(tag, "rbac:role:") {
			roleName := strings.TrimPrefix(tag, "rbac:role:")
			roles = append(roles, roleName)
		}
	}
	
	return roles, nil
}

// Obsolete helper functions removed - using tag-based RBAC now

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