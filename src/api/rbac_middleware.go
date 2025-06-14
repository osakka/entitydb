// Package api provides HTTP handlers for the EntityDB REST API.
// This file implements RBAC (Role-Based Access Control) middleware.
package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"entitydb/logger"
	"entitydb/models"
)

// RBACPermission represents a required permission for an operation.
// Permissions follow the pattern "resource:action" where:
//   - Resource: The entity type being accessed (entity, user, system, etc.)
//   - Action: The operation being performed (view, create, update, delete)
//
// Special permissions:
//   - "*": Grants all permissions (admin only)
//   - "resource:*": Grants all actions on a resource
type RBACPermission struct {
	Resource string // e.g., "entity", "issue", "user", "system"
	Action   string // e.g., "view", "create", "update", "delete"
}

// RBACContext stores user permissions in the request context.
// This is added to the request context by the RBAC middleware and can be
// accessed by handlers to make authorization decisions.
type RBACContext struct {
	User        *models.Entity  // The authenticated user entity
	Permissions []string        // List of permissions extracted from user tags
	IsAdmin     bool           // True if user has admin role
}

// Context key for RBAC data
type rbacContextKey struct{}

// RBACMiddleware creates middleware that enforces permissions.
//
// This middleware:
//   1. Extracts the Bearer token from the Authorization header
//   2. Validates the session token with the security manager (database-based)
//   3. Loads the user entity from the database
//   4. Extracts RBAC permissions from user tags
//   5. Checks if the user has the required permission
//   6. Tracks permission events for auditing
//   7. Adds user context to the request for handlers
//
// Permission Format:
//   Tags use the format "rbac:perm:resource:action"
//   Example: "rbac:perm:entity:create" grants entity creation
//
// Admin Override:
//   Users with "rbac:role:admin" or "rbac:perm:*" bypass all checks
//
// Usage:
//   handler := RBACMiddleware(repo, securityManager, RBACPermission{
//       Resource: "entity",
//       Action: "create",
//   })(actualHandler)
func RBACMiddleware(entityRepo models.EntityRepository, securityManager *models.SecurityManager, requiredPerm RBACPermission) MiddlewareFunc {
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
			
			// Validate token using security manager (database-based sessions)
			securityUser, err := securityManager.ValidateSession(token)
			if err != nil {
				RespondError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}
			
			// SecurityUser already contains the entity, so we can use it directly
			user := securityUser.Entity
			
			// Extract permissions from user tags
			permissions := models.GetTagsByNamespace(user.Tags, "rbac")
			isAdmin := hasAdminRole(permissions)
			
			// Check required permission
			requiredPermTag := formatPermissionTag(requiredPerm)
			hasPermission := isAdmin || models.HasPermission(permissions, requiredPermTag)
			
			if !hasPermission {
				// Track permission denied event
				go func() {
					permEvent := &models.Entity{
						ID: fmt.Sprintf("perm_event_%s_%d", securityUser.ID, time.Now().UnixNano()),
						Tags: []string{
							"type:permission_event",
							"status:denied",
							"user:" + securityUser.ID,
							"permission:" + requiredPermTag,
							"resource:" + requiredPerm.Resource,
							"action:" + requiredPerm.Action,
						},
						Content: []byte(fmt.Sprintf(`{"user":"%s","permission":"%s","granted":false,"timestamp":"%s"}`, 
							securityUser.ID, requiredPermTag, time.Now().Format(time.RFC3339))),
					}
					if err := entityRepo.Create(permEvent); err != nil {
						logger.Error("Failed to track permission event: %v", err)
					}
				}()
				
				RespondError(w, http.StatusForbidden, 
					fmt.Sprintf("Insufficient permissions: %s required", requiredPermTag))
				return
			}
			
			// Track permission granted event
			go func() {
				permEvent := &models.Entity{
					ID: fmt.Sprintf("perm_event_%s_%d", securityUser.ID, time.Now().UnixNano()),
					Tags: []string{
						"type:permission_event", 
						"status:granted",
						"user:" + securityUser.ID,
						"permission:" + requiredPermTag,
						"resource:" + requiredPerm.Resource,
						"action:" + requiredPerm.Action,
					},
					Content: []byte(fmt.Sprintf(`{"user":"%s","permission":"%s","granted":true,"timestamp":"%s"}`, 
						securityUser.ID, requiredPermTag, time.Now().Format(time.RFC3339))),
				}
				if err := entityRepo.Create(permEvent); err != nil {
					logger.Error("Failed to track permission event: %v", err)
				}
			}()
			
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

// GetRBACContext retrieves the RBAC context from the request.
// Returns the context and true if found, or nil and false if not found.
// This should be called by handlers that need to access user permissions
// after the RBAC middleware has run.
func GetRBACContext(r *http.Request) (*RBACContext, bool) {
	ctx, ok := r.Context().Value(rbacContextKey{}).(*RBACContext)
	return ctx, ok
}

// formatPermissionTag formats a permission into a tag string.
// Converts RBACPermission{Resource: "entity", Action: "create"}
// into "rbac:perm:entity:create" for tag-based permission checking.
func formatPermissionTag(perm RBACPermission) string {
	return fmt.Sprintf("rbac:perm:%s:%s", perm.Resource, perm.Action)
}

// hasAdminRole checks if user has admin role privileges.
// Returns true if the user has either:
//   - rbac:role:admin - explicit admin role
//   - rbac:perm:* - wildcard permission (grants all permissions)
func hasAdminRole(permissions []string) bool {
	for _, perm := range permissions {
		if perm == "rbac:role:admin" || perm == "rbac:perm:*" {
			return true
		}
	}
	return false
}

// Common permission definitions for use throughout the API.
// These constants define the standard permissions used by EntityDB.
// Each permission follows the pattern: Resource + Action
var (
	// Entity permissions - for general entity CRUD operations
	PermEntityView   = RBACPermission{Resource: "entity", Action: "view"}
	PermEntityCreate = RBACPermission{Resource: "entity", Action: "create"}
	PermEntityUpdate = RBACPermission{Resource: "entity", Action: "update"}
	PermEntityDelete = RBACPermission{Resource: "entity", Action: "delete"}
	
	// Relationship permissions - for entity relationship management
	PermRelationView   = RBACPermission{Resource: "relation", Action: "view"}
	PermRelationCreate = RBACPermission{Resource: "relation", Action: "create"}
	PermRelationUpdate = RBACPermission{Resource: "relation", Action: "update"}
	PermRelationDelete = RBACPermission{Resource: "relation", Action: "delete"}
	
	// User permissions - for user management (admin only)
	PermUserView   = RBACPermission{Resource: "user", Action: "view"}
	PermUserCreate = RBACPermission{Resource: "user", Action: "create"}
	PermUserUpdate = RBACPermission{Resource: "user", Action: "update"}
	PermUserDelete = RBACPermission{Resource: "user", Action: "delete"}
	
	// System permissions - for system monitoring and configuration
	PermSystemView   = RBACPermission{Resource: "system", Action: "view"}
	PermSystemUpdate = RBACPermission{Resource: "system", Action: "update"}
	
	// Configuration permissions - for feature flags and settings
	PermConfigView   = RBACPermission{Resource: "config", Action: "view"}
	PermConfigUpdate = RBACPermission{Resource: "config", Action: "update"}
)

// CheckEntityPermission checks if user has permission for a specific entity.
// This function considers the entity's type tag to determine the required permission.
// For example, an entity with "type:document" requires "document:view" permission.
//
// Parameters:
//   - rbacCtx: The RBAC context containing user permissions
//   - entity: The entity to check permissions for
//   - action: The action to perform (view, create, update, delete)
//
// Returns true if the user has permission, false otherwise.
// Admin users always return true.
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

// getEntityType extracts the entity type from tags.
// Looks for tags with the "type:" prefix and returns the type value.
// If no type tag is found, returns "entity" as the default type.
//
// Examples:
//   - Tags ["type:user", "status:active"] returns "user"
//   - Tags ["status:active"] returns "entity"
func getEntityType(entity *models.Entity) string {
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "type:") {
			return strings.TrimPrefix(tag, "type:")
		}
	}
	return "entity"
}