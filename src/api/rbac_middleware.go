package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	
	"entitydb/models"
)

// RBACPermission represents a required permission for an operation
type RBACPermission struct {
	Resource string // e.g., "entity", "issue", "user", "system"
	Action   string // e.g., "view", "create", "update", "delete"
}

// RBACContext stores user permissions in the request context
type RBACContext struct {
	User        *models.Entity
	Permissions []string
	IsAdmin     bool
}

// Context key for RBAC data
type rbacContextKey struct{}

// RBACMiddleware creates middleware that enforces permissions
func RBACMiddleware(entityRepo models.EntityRepository, requiredPerm RBACPermission) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
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
			
			// Extract username from token
			// Token format is either "token_username_nanos" or "user:username" for testing
			var username string
			if strings.HasPrefix(token, "token_") {
				// Real token format: token_username_nanos
				parts := strings.Split(token, "_")
				if len(parts) < 3 {
					RespondError(w, http.StatusUnauthorized, "Invalid token format")
					return
				}
				username = parts[1]
			} else if strings.HasPrefix(token, "user:") {
				// Test token format: user:username
				username = strings.TrimPrefix(token, "user:")
			} else {
				RespondError(w, http.StatusUnauthorized, "Invalid token")
				return
			}
			
			// Find user entity using List method
			usersAll, err := entityRepo.List()
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "Failed to load users")
				return
			}
			
			// Filter to find the user
			var user *models.Entity
			for _, u := range usersAll {
				var hasUserType, hasUsername bool
				for _, tag := range u.Tags {
					if tag == "type:user" {
						hasUserType = true
					}
					if tag == fmt.Sprintf("id:username:%s", username) {
						hasUsername = true
					}
				}
				if hasUserType && hasUsername {
					user = u
					break
				}
			}
			
			if user == nil {
				RespondError(w, http.StatusUnauthorized, "User not found")
				return
			}
			
			// Extract permissions from user tags
			permissions := models.GetTagsByNamespace(user.Tags, "rbac")
			isAdmin := hasAdminRole(permissions)
			
			// Check required permission
			requiredPermTag := formatPermissionTag(requiredPerm)
			hasPermission := isAdmin || models.HasPermission(permissions, requiredPermTag)
			
			if !hasPermission {
				RespondError(w, http.StatusForbidden, 
					fmt.Sprintf("Insufficient permissions: %s required", requiredPermTag))
				return
			}
			
			// Add RBAC context to request
			rbacCtx := &RBACContext{
				User:        user,
				Permissions: permissions,
				IsAdmin:     isAdmin,
			}
			
			ctx := context.WithValue(r.Context(), rbacContextKey{}, rbacCtx)
			next(w, r.WithContext(ctx))
		}
	}
}

// GetRBACContext retrieves the RBAC context from the request
func GetRBACContext(r *http.Request) (*RBACContext, bool) {
	ctx, ok := r.Context().Value(rbacContextKey{}).(*RBACContext)
	return ctx, ok
}

// formatPermissionTag formats a permission into a tag
func formatPermissionTag(perm RBACPermission) string {
	return fmt.Sprintf("rbac:perm:%s:%s", perm.Resource, perm.Action)
}

// hasAdminRole checks if user has admin role
func hasAdminRole(permissions []string) bool {
	for _, perm := range permissions {
		if perm == "rbac:role:admin" || perm == "rbac:perm:*" {
			return true
		}
	}
	return false
}

// Common permission definitions
var (
	PermEntityView   = RBACPermission{Resource: "entity", Action: "view"}
	PermEntityCreate = RBACPermission{Resource: "entity", Action: "create"}
	PermEntityUpdate = RBACPermission{Resource: "entity", Action: "update"}
	PermEntityDelete = RBACPermission{Resource: "entity", Action: "delete"}
	
	PermRelationView   = RBACPermission{Resource: "relation", Action: "view"}
	PermRelationCreate = RBACPermission{Resource: "relation", Action: "create"}
	PermRelationUpdate = RBACPermission{Resource: "relation", Action: "update"}
	PermRelationDelete = RBACPermission{Resource: "relation", Action: "delete"}
	
	PermUserView   = RBACPermission{Resource: "user", Action: "view"}
	PermUserCreate = RBACPermission{Resource: "user", Action: "create"}
	PermUserUpdate = RBACPermission{Resource: "user", Action: "update"}
	PermUserDelete = RBACPermission{Resource: "user", Action: "delete"}
	
	PermSystemView   = RBACPermission{Resource: "system", Action: "view"}
	PermSystemUpdate = RBACPermission{Resource: "system", Action: "update"}
	
	PermConfigView   = RBACPermission{Resource: "config", Action: "view"}
	PermConfigUpdate = RBACPermission{Resource: "config", Action: "update"}
)

// CheckEntityPermission checks if user has permission for a specific entity
func CheckEntityPermission(rbacCtx *RBACContext, entity *models.Entity, action string) bool {
	// Admins have all permissions
	if rbacCtx.IsAdmin {
		return true
	}
	
	// Check specific permission for entity type
	entityType := getEntityType(entity)
	requiredPerm := fmt.Sprintf("rbac:perm:%s:%s", entityType, action)
	
	return models.HasPermission(rbacCtx.Permissions, requiredPerm)
}

// getEntityType extracts the entity type from tags
func getEntityType(entity *models.Entity) string {
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "type:") {
			return strings.TrimPrefix(tag, "type:")
		}
	}
	return "entity"
}