// Package api provides HTTP handlers for the EntityDB REST API.
// This file implements authentication middleware for session validation.
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

// AuthContext holds authentication information for the current request.
// This is populated by the authentication middleware and made available
// to handlers through the request context.
type AuthContext struct {
	Session     *models.Session  // Active session information
	User        *models.Entity   // Authenticated user entity
	Permissions []string         // User's RBAC permissions
	IsAdmin     bool            // True if user has admin privileges
}

// SessionAuthMiddleware creates middleware that validates sessions.
//
// This middleware:
//   1. Extracts Bearer token from Authorization header
//   2. Validates the token with the session manager
//   3. Loads the user entity associated with the session
//   4. Extracts RBAC permissions from user tags
//   5. Creates an AuthContext and adds it to the request
//
// The middleware does NOT check specific permissions - it only validates
// that the user has a valid session. Use RequirePermission or RBACMiddleware
// for permission checking.
//
// Usage:
//   handler := SessionAuthMiddleware(sessions, repo)(actualHandler)
//
// Error Responses:
//   - 401 Unauthorized: Missing/invalid token or expired session
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

// RequirePermission creates middleware that checks for specific permissions.
//
// This is a convenience function that combines SessionAuthMiddleware with
// permission checking. It first validates the session, then checks if the
// user has the required permission.
//
// Permission Checking:
//   - Admin users (with rbac:role:admin) bypass all permission checks
//   - Specific permissions must match exactly (e.g., entity:create)
//   - Wildcard permissions are supported (e.g., entity:* or *)
//
// Usage:
//   handler := RequirePermission(sessions, repo, RBACPermission{
//       Resource: "entity",
//       Action: "create",
//   })(actualHandler)
//
// Error Responses:
//   - 401 Unauthorized: Invalid session
//   - 403 Forbidden: Valid session but lacks required permission
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