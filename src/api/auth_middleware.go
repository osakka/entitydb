package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	
	"entitydb/models"
)

// contextKey is used for storing data in request context
type authContextKey struct{}

// AuthContext holds authentication information
type AuthContext struct {
	Session     *models.Session
	User        *models.Entity
	Permissions []string
	IsAdmin     bool
}

// SessionAuthMiddleware creates middleware that validates sessions
func SessionAuthMiddleware(sessionManager *models.SessionManager, entityRepo models.EntityRepository) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				RespondError(w, http.StatusUnauthorized, "Authentication required")
				return
			}
			
			// Expected format: "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				RespondError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}
			
			token := parts[1]
			
			// Get session
			session, exists := sessionManager.GetSession(token)
			if !exists {
				RespondError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}
			
			// Get user entity
			user, err := entityRepo.GetByID(session.UserID)
			if err != nil {
				RespondError(w, http.StatusUnauthorized, "User not found")
				return
			}
			
			// Extract permissions from user tags
			permissions := models.GetTagsByNamespace(user.Tags, "rbac")
			isAdmin := hasAdminRole(permissions)
			
			// Create auth context
			authCtx := &AuthContext{
				Session:     session,
				User:        user,
				Permissions: permissions,
				IsAdmin:     isAdmin,
			}
			
			// Add to request context
			ctx := context.WithValue(r.Context(), authContextKey{}, authCtx)
			next(w, r.WithContext(ctx))
		}
	}
}

// RequirePermission creates middleware that checks for specific permissions
func RequirePermission(sessionManager *models.SessionManager, entityRepo models.EntityRepository, requiredPerm RBACPermission) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		// First apply auth middleware
		authHandler := SessionAuthMiddleware(sessionManager, entityRepo)(func(w http.ResponseWriter, r *http.Request) {
			// Get auth context
			authCtx, ok := GetAuthContext(r)
			if !ok {
				RespondError(w, http.StatusInternalServerError, "Authentication context not found")
				return
			}
			
			// Check permission
			requiredPermTag := formatPermissionTag(requiredPerm)
			hasPermission := authCtx.IsAdmin || models.HasPermission(authCtx.Permissions, requiredPermTag)
			
			if !hasPermission {
				RespondError(w, http.StatusForbidden, 
					fmt.Sprintf("Insufficient permissions: %s required", requiredPermTag))
				return
			}
			
			next(w, r)
		})
		
		return authHandler
	}
}

// GetAuthContext retrieves the auth context from the request
func GetAuthContext(r *http.Request) (*AuthContext, bool) {
	ctx, ok := r.Context().Value(authContextKey{}).(*AuthContext)
	return ctx, ok
}

// RefreshSession middleware refreshes the session expiration on each request
func RefreshSession(sessionManager *models.SessionManager) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := GetAuthContext(r)
			if ok && authCtx.Session != nil {
				sessionManager.RefreshSession(authCtx.Session.Token)
			}
			next(w, r)
		}
	}
}

// SessionStats returns session statistics
func SessionStats(sessionManager *models.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := map[string]interface{}{
			"active_sessions": sessionManager.GetActiveSessions(),
		}
		
		// If user is admin, provide more details
		authCtx, ok := GetAuthContext(r)
		if ok && authCtx.IsAdmin {
			// Could add more detailed stats here
			stats["session_ttl"] = "configurable"
		}
		
		RespondJSON(w, http.StatusOK, stats)
	}
}