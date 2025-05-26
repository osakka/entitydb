package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"entitydb/models"
)

// SecurityMiddleware replaces the old RBAC middleware with relationship-based security
type SecurityMiddleware struct {
	securityManager *models.SecurityManager
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(securityManager *models.SecurityManager) *SecurityMiddleware {
	return &SecurityMiddleware{
		securityManager: securityManager,
	}
}

// SecurityContext stores security information in the request context
type SecurityContext struct {
	User    *models.SecurityUser
	Session *models.SecuritySession
	Token   string
}

// Context key for security data
type securityContextKey struct{}

// RequireAuthentication ensures the request has a valid session
func (sm *SecurityMiddleware) RequireAuthentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			RespondError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Extract token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			RespondError(w, http.StatusUnauthorized, "Invalid token format")
			return
		}
		token := parts[1]

		// Validate session using relationship-based authentication
		user, err := sm.securityManager.ValidateSession(token)
		if err != nil {
			RespondError(w, http.StatusUnauthorized, "Invalid or expired session")
			return
		}

		// Create security context
		securityCtx := &SecurityContext{
			User:  user,
			Token: token,
		}

		// Add to request context
		ctx := context.WithValue(r.Context(), securityContextKey{}, securityCtx)
		next(w, r.WithContext(ctx))
	}
}

// RequirePermission creates middleware that checks for specific permissions using relationship traversal
func (sm *SecurityMiddleware) RequirePermission(resource, action string) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		// First apply authentication middleware
		authHandler := sm.RequireAuthentication(func(w http.ResponseWriter, r *http.Request) {
			// Get security context
			securityCtx, ok := GetSecurityContext(r)
			if !ok {
				RespondError(w, http.StatusInternalServerError, "Security context not found")
				return
			}

			// Check permission using relationship traversal
			hasPermission, err := sm.securityManager.HasPermission(securityCtx.User, resource, action)
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "Failed to check permissions")
				return
			}

			if !hasPermission {
				RespondError(w, http.StatusForbidden,
					fmt.Sprintf("Insufficient permissions: %s:%s required", resource, action))
				return
			}

			next(w, r)
		})

		return authHandler
	}
}

// RequirePermissionInDataspace creates middleware that checks for specific permissions in a dataspace
func (sm *SecurityMiddleware) RequirePermissionInDataspace(resource, action string) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		// First apply authentication middleware
		authHandler := sm.RequireAuthentication(func(w http.ResponseWriter, r *http.Request) {
			// Get security context
			securityCtx, ok := GetSecurityContext(r)
			if !ok {
				RespondError(w, http.StatusInternalServerError, "Security context not found")
				return
			}

			// Extract dataspace ID from request path or query parameters
			dataspaceID := r.URL.Query().Get("dataspace_id")
			if dataspaceID == "" {
				// Try to extract from path for REST-style URLs like /dataspaces/{id}/entities
				pathParts := strings.Split(r.URL.Path, "/")
				for i, part := range pathParts {
					if part == "dataspaces" && i+1 < len(pathParts) {
						dataspaceID = pathParts[i+1]
						break
					}
				}
			}

			// Check permission using relationship traversal with dataspace context
			hasPermission, err := sm.securityManager.HasPermissionInDataspace(securityCtx.User, resource, action, dataspaceID)
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "Failed to check permissions")
				return
			}

			if !hasPermission {
				if dataspaceID != "" {
					RespondError(w, http.StatusForbidden,
						fmt.Sprintf("Insufficient permissions: %s:%s required in dataspace %s", resource, action, dataspaceID))
				} else {
					RespondError(w, http.StatusForbidden,
						fmt.Sprintf("Insufficient permissions: %s:%s required", resource, action))
				}
				return
			}

			next(w, r)
		})

		return authHandler
	}
}

// RequireDataspaceAccess creates middleware that checks if user can access a specific dataspace
func (sm *SecurityMiddleware) RequireDataspaceAccess() MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		authHandler := sm.RequireAuthentication(func(w http.ResponseWriter, r *http.Request) {
			securityCtx, ok := GetSecurityContext(r)
			if !ok {
				RespondError(w, http.StatusInternalServerError, "Security context not found")
				return
			}

			// Extract dataspace ID from request
			dataspaceID := r.URL.Query().Get("dataspace_id")
			if dataspaceID == "" {
				// Try to extract from path
				pathParts := strings.Split(r.URL.Path, "/")
				for i, part := range pathParts {
					if part == "dataspaces" && i+1 < len(pathParts) {
						dataspaceID = pathParts[i+1]
						break
					}
				}
			}

			if dataspaceID == "" {
				RespondError(w, http.StatusBadRequest, "Dataspace ID required")
				return
			}

			// Check dataspace access
			hasAccess, err := sm.securityManager.CanAccessDataspace(securityCtx.User, dataspaceID)
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "Failed to check dataspace access")
				return
			}

			if !hasAccess {
				RespondError(w, http.StatusForbidden, 
					fmt.Sprintf("Access denied to dataspace %s", dataspaceID))
				return
			}

			next(w, r)
		})

		return authHandler
	}
}

// RequireRole creates middleware that checks for specific roles using relationship traversal
func (sm *SecurityMiddleware) RequireRole(roleName string) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		authHandler := sm.RequireAuthentication(func(w http.ResponseWriter, r *http.Request) {
			securityCtx, ok := GetSecurityContext(r)
			if !ok {
				RespondError(w, http.StatusInternalServerError, "Security context not found")
				return
			}

			// Check if user has the required role
			hasRole, err := sm.hasRole(securityCtx.User, roleName)
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "Failed to check role")
				return
			}

			if !hasRole {
				RespondError(w, http.StatusForbidden,
					fmt.Sprintf("Role required: %s", roleName))
				return
			}

			next(w, r)
		})

		return authHandler
	}
}

// GetSecurityContext retrieves the security context from the request
func GetSecurityContext(r *http.Request) (*SecurityContext, bool) {
	ctx, ok := r.Context().Value(securityContextKey{}).(*SecurityContext)
	return ctx, ok
}

// hasRole checks if a user has a specific role (helper function)
func (sm *SecurityMiddleware) hasRole(user *models.SecurityUser, roleName string) (bool, error) {
	// This could be optimized with caching, but for now we'll traverse the graph
	// Get user roles directly
	userRoles, err := sm.getUserRoles(user.ID)
	if err != nil {
		return false, err
	}

	for _, role := range userRoles {
		if role.Name == roleName {
			return true, nil
		}
	}

	// Check group-based roles
	userGroups, err := sm.getUserGroups(user.ID)
	if err != nil {
		return false, err
	}

	for _, group := range userGroups {
		groupRoles, err := sm.getGroupRoles(group.ID)
		if err != nil {
			continue
		}

		for _, role := range groupRoles {
			if role.Name == roleName {
				return true, nil
			}
		}
	}

	return false, nil
}

// Helper functions (these mirror the SecurityManager but are needed here)

func (sm *SecurityMiddleware) getUserRoles(userID string) ([]*models.SecurityRole, error) {
	// This should ideally be moved to SecurityManager or use the SecurityManager methods
	// For now, duplicating the logic to keep middleware independent
	return nil, fmt.Errorf("not implemented - use SecurityManager.HasPermission instead")
}

func (sm *SecurityMiddleware) getUserGroups(userID string) ([]*models.Entity, error) {
	return nil, fmt.Errorf("not implemented - use SecurityManager.HasPermission instead")
}

func (sm *SecurityMiddleware) getGroupRoles(groupID string) ([]*models.SecurityRole, error) {
	return nil, fmt.Errorf("not implemented - use SecurityManager.HasPermission instead")
}

// Common permission definitions for convenience
var (
	PermissionEntityView   = []string{"entity", "view"}
	PermissionEntityCreate = []string{"entity", "create"}
	PermissionEntityUpdate = []string{"entity", "update"}
	PermissionEntityDelete = []string{"entity", "delete"}

	PermissionUserView   = []string{"user", "view"}
	PermissionUserCreate = []string{"user", "create"}
	PermissionUserUpdate = []string{"user", "update"}
	PermissionUserDelete = []string{"user", "delete"}

	PermissionAdminView   = []string{"admin", "view"}
	PermissionAdminUpdate = []string{"admin", "update"}

	PermissionSystemView   = []string{"system", "view"}
	PermissionSystemUpdate = []string{"system", "update"}
)

// Convenience functions for common permission checks
func (sm *SecurityMiddleware) RequireEntityView() MiddlewareFunc {
	return sm.RequirePermission("entity", "view")
}

func (sm *SecurityMiddleware) RequireEntityCreate() MiddlewareFunc {
	return sm.RequirePermission("entity", "create")
}

func (sm *SecurityMiddleware) RequireEntityUpdate() MiddlewareFunc {
	return sm.RequirePermission("entity", "update")
}

func (sm *SecurityMiddleware) RequireUserCreate() MiddlewareFunc {
	return sm.RequirePermission("user", "create")
}

func (sm *SecurityMiddleware) RequireAdminView() MiddlewareFunc {
	return sm.RequirePermission("admin", "view")
}

func (sm *SecurityMiddleware) RequireSystemView() MiddlewareFunc {
	return sm.RequirePermission("system", "view")
}

func (sm *SecurityMiddleware) RequireAdminRole() MiddlewareFunc {
	return sm.RequireRole("admin")
}