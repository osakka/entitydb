package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"entitydb/models"
)

// DataspaceContext stores hub information in the request context
type DataspaceContext struct {
	DataspaceName        string
	UserHubs       []string // Hubs user has access to
	CanAccessHub   bool
	IsGlobalAdmin  bool
}

// Context key for hub data
type hubContextKey struct{}

// DataspaceMiddleware creates middleware that enforces hub-based access control
func DataspaceMiddleware(entityRepo models.EntityRepository, sessionManager *models.SessionManager) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get RBAC context (should be set by RBACMiddleware first)
			rbacCtx, hasRBAC := GetRBACContext(r)
			if !hasRBAC {
				RespondError(w, http.StatusUnauthorized, "RBAC context required")
				return
			}

			// Extract hub from query parameter or request body
			dataspaceName := extractHubFromRequest(r)
			if dataspaceName == "" {
				// For create operations, hub might be in request body
				// This will be handled in individual handlers
				dataspaceName = "default" // Allow for now, handlers will validate
			}

			// Check if user has access to this hub
			userHubs := getUserHubs(rbacCtx.Permissions)
			canAccessHub := rbacCtx.IsAdmin || containsHub(userHubs, dataspaceName) || dataspaceName == "default"

			// Create hub context
			hubCtx := &DataspaceContext{
				DataspaceName:       dataspaceName,
				UserHubs:      userHubs,
				CanAccessHub:  canAccessHub,
				IsGlobalAdmin: rbacCtx.IsAdmin,
			}

			// Add hub context to request
			ctx := context.WithValue(r.Context(), hubContextKey{}, hubCtx)
			next(w, r.WithContext(ctx))
		}
	}
}

// GetDataspaceContext retrieves the hub context from the request
func GetDataspaceContext(r *http.Request) (*DataspaceContext, bool) {
	ctx, ok := r.Context().Value(hubContextKey{}).(*DataspaceContext)
	return ctx, ok
}

// ValidateEntityDataspace validates that an entity belongs to a hub the user can access
func ValidateEntityDataspace(rbacCtx *RBACContext, entity *models.Entity) error {
	// Global admin can access all hubs
	if rbacCtx.IsAdmin {
		return nil
	}

	// Extract hub from entity tags
	entityHub := getEntityHub(entity)
	if entityHub == "" {
		return fmt.Errorf("entity missing required hub tag")
	}

	// Check if user has access to this hub
	userHubs := getUserHubs(rbacCtx.Permissions)
	if !containsHub(userHubs, entityHub) {
		return fmt.Errorf("access denied to hub: %s", entityHub)
	}

	return nil
}

// RequireHubTag ensures entity has a hub tag
func RequireHubTag(entity *models.Entity) error {
	hub := getEntityHub(entity)
	if hub == "" {
		return fmt.Errorf("entity must have a hub tag (hub:name)")
	}
	return nil
}

// AddDefaultHubTag adds a default hub tag if none exists
func AddDefaultHubTag(tags []string, defaultHub string) []string {
	// Check if hub tag already exists
	for _, tag := range tags {
		if strings.HasPrefix(tag, "dataspace:") {
			return tags // Hub tag already exists
		}
	}

	// Add default hub tag
	return append(tags, fmt.Sprintf("dataspace:%s", defaultHub))
}

// CheckDataspacePermission checks if user has specific permission for a hub
func CheckDataspacePermission(rbacCtx *RBACContext, dataspaceName string, action string) bool {
	// Global admin has all permissions
	if rbacCtx.IsAdmin {
		return true
	}

	// Check hub-specific permission
	hubPerm := fmt.Sprintf("rbac:perm:entity:%s:dataspace:%s", action, dataspaceName)
	if models.HasPermission(rbacCtx.Permissions, hubPerm) {
		return true
	}

	// Check general hub permission
	generalPerm := fmt.Sprintf("rbac:perm:entity:%s:dataspace:*", action)
	return models.HasPermission(rbacCtx.Permissions, generalPerm)
}

// CheckHubManagementPermission checks hub management permissions
func CheckHubManagementPermission(rbacCtx *RBACContext, action string, dataspaceName string) bool {
	// Global admin has all permissions
	if rbacCtx.IsAdmin {
		return true
	}

	// Check specific hub management permission
	if dataspaceName != "" {
		hubPerm := fmt.Sprintf("rbac:perm:dataspace:%s:%s", action, dataspaceName)
		if models.HasPermission(rbacCtx.Permissions, hubPerm) {
			return true
		}
	}

	// Check general hub management permission
	generalPerm := fmt.Sprintf("rbac:perm:dataspace:%s", action)
	return models.HasPermission(rbacCtx.Permissions, generalPerm)
}

// Helper functions

// extractHubFromRequest gets hub name from query parameters
func extractHubFromRequest(r *http.Request) string {
	// Try query parameter first
	if hub := r.URL.Query().Get("hub"); hub != "" {
		return hub
	}

	// Try header
	if hub := r.Header.Get("X-Hub"); hub != "" {
		return hub
	}

	return ""
}

// getUserHubs extracts hub access from user permissions
func getUserHubs(permissions []string) []string {
	var hubs []string
	for _, perm := range permissions {
		// Look for permissions like rbac:perm:entity:*:dataspace:worcha
		if strings.Contains(perm, ":dataspace:") {
			parts := strings.Split(perm, ":")
			if len(parts) >= 6 && parts[4] == "hub" {
				dataspaceName := parts[5]
				if dataspaceName != "*" && !containsHub(hubs, dataspaceName) {
					hubs = append(hubs, dataspaceName)
				}
			}
		}
	}
	return hubs
}

// getEntityHub extracts hub name from entity tags
func getEntityHub(entity *models.Entity) string {
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "dataspace:") {
			return strings.TrimPrefix(tag, "dataspace:")
		}
	}
	return ""
}

// containsHub checks if hub list contains specific hub
func containsHub(hubs []string, hub string) bool {
	for _, h := range hubs {
		if h == hub {
			return true
		}
	}
	return false
}

// FormatDataspaceTag creates a hub tag
func FormatDataspaceTag(dataspaceName string) string {
	return fmt.Sprintf("dataspace:%s", dataspaceName)
}

// FormatTraitTag creates a trait tag for a hub
func FormatTraitTag(dataspaceName, namespace, value string) string {
	return fmt.Sprintf("%s:trait:%s:%s", dataspaceName, namespace, value)
}

// FormatSelfTag creates a self tag for a hub
func FormatSelfTag(dataspaceName, namespace, value string) string {
	return fmt.Sprintf("%s:self:%s:%s", dataspaceName, namespace, value)
}

// ParseDataspaceTag parses a hub tag and returns the hub name
func ParseDataspaceTag(tag string) (string, bool) {
	if strings.HasPrefix(tag, "dataspace:") {
		return strings.TrimPrefix(tag, "dataspace:"), true
	}
	return "", false
}

// ParseTraitTag parses a trait tag and returns hub, namespace, value
func ParseTraitTag(tag string) (hub, namespace, value string, ok bool) {
	parts := strings.Split(tag, ":")
	if len(parts) == 4 && parts[1] == "trait" {
		return parts[0], parts[2], parts[3], true
	}
	return "", "", "", false
}

// ParseSelfTag parses a self tag and returns hub, namespace, value
func ParseSelfTag(tag string) (hub, namespace, value string, ok bool) {
	parts := strings.Split(tag, ":")
	if len(parts) == 4 && parts[1] == "self" {
		return parts[0], parts[2], parts[3], true
	}
	return "", "", "", false
}