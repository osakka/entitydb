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
func RBACMiddleware(entityRepo models.EntityRepository, sessionManager *models.SessionManager, requiredPerm RBACPermission) MiddlewareFunc {
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
			
			// Validate token using session manager
			session, exists := sessionManager.GetSession(token)
			if !exists {
				RespondError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}
			
			// Find user entity by user ID from session
			user, err := entityRepo.GetByID(session.UserID)
			if err != nil || user == nil {
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
				// Track permission denied event
				go func() {
					permEvent := &models.Entity{
						ID: fmt.Sprintf("perm_event_%s_%d", session.UserID, time.Now().UnixNano()),
						Tags: []string{
							"type:permission_event",
							"status:denied",
							"user:" + session.UserID,
							"permission:" + requiredPermTag,
							"resource:" + requiredPerm.Resource,
							"action:" + requiredPerm.Action,
						},
						Content: []byte(fmt.Sprintf(`{"user":"%s","permission":"%s","granted":false,"timestamp":"%s"}`, 
							session.UserID, requiredPermTag, time.Now().Format(time.RFC3339))),
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
					ID: fmt.Sprintf("perm_event_%s_%d", session.UserID, time.Now().UnixNano()),
					Tags: []string{
						"type:permission_event", 
						"status:granted",
						"user:" + session.UserID,
						"permission:" + requiredPermTag,
						"resource:" + requiredPerm.Resource,
						"action:" + requiredPerm.Action,
					},
					Content: []byte(fmt.Sprintf(`{"user":"%s","permission":"%s","granted":true,"timestamp":"%s"}`, 
						session.UserID, requiredPermTag, time.Now().Format(time.RFC3339))),
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