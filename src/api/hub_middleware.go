package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"entitydb/models"
)

// HubContext stores hub information in the request context
type HubContext struct {
	HubName        string
	UserHubs       []string // Hubs user has access to
	CanAccessHub   bool
	IsGlobalAdmin  bool
}

// Context key for hub data
type hubContextKey struct{}

// HubMiddleware creates middleware that enforces hub-based access control
func HubMiddleware(entityRepo models.EntityRepository, sessionManager *models.SessionManager) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get RBAC context (should be set by RBACMiddleware first)
			rbacCtx, hasRBAC := GetRBACContext(r)
			if !hasRBAC {
				RespondError(w, http.StatusUnauthorized, "RBAC context required")
				return
			}

			// Extract hub from query parameter or request body
			hubName := extractHubFromRequest(r)
			if hubName == "" {
				// For create operations, hub might be in request body
				// This will be handled in individual handlers
				hubName = "default" // Allow for now, handlers will validate
			}

			// Check if user has access to this hub
			userHubs := getUserHubs(rbacCtx.Permissions)
			canAccessHub := rbacCtx.IsAdmin || containsHub(userHubs, hubName) || hubName == "default"

			// Create hub context
			hubCtx := &HubContext{
				HubName:       hubName,
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

// GetHubContext retrieves the hub context from the request
func GetHubContext(r *http.Request) (*HubContext, bool) {
	ctx, ok := r.Context().Value(hubContextKey{}).(*HubContext)
	return ctx, ok
}

// ValidateEntityHub validates that an entity belongs to a hub the user can access
func ValidateEntityHub(rbacCtx *RBACContext, entity *models.Entity) error {
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
		if strings.HasPrefix(tag, "hub:") {
			return tags // Hub tag already exists
		}
	}

	// Add default hub tag
	return append(tags, fmt.Sprintf("hub:%s", defaultHub))
}

// CheckHubPermission checks if user has specific permission for a hub
func CheckHubPermission(rbacCtx *RBACContext, hubName string, action string) bool {
	// Global admin has all permissions
	if rbacCtx.IsAdmin {
		return true
	}

	// Check hub-specific permission
	hubPerm := fmt.Sprintf("rbac:perm:entity:%s:hub:%s", action, hubName)
	if models.HasPermission(rbacCtx.Permissions, hubPerm) {
		return true
	}

	// Check general hub permission
	generalPerm := fmt.Sprintf("rbac:perm:entity:%s:hub:*", action)
	return models.HasPermission(rbacCtx.Permissions, generalPerm)
}

// CheckHubManagementPermission checks hub management permissions
func CheckHubManagementPermission(rbacCtx *RBACContext, action string, hubName string) bool {
	// Global admin has all permissions
	if rbacCtx.IsAdmin {
		return true
	}

	// Check specific hub management permission
	if hubName != "" {
		hubPerm := fmt.Sprintf("rbac:perm:hub:%s:%s", action, hubName)
		if models.HasPermission(rbacCtx.Permissions, hubPerm) {
			return true
		}
	}

	// Check general hub management permission
	generalPerm := fmt.Sprintf("rbac:perm:hub:%s", action)
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
		// Look for permissions like rbac:perm:entity:*:hub:worcha
		if strings.Contains(perm, ":hub:") {
			parts := strings.Split(perm, ":")
			if len(parts) >= 6 && parts[4] == "hub" {
				hubName := parts[5]
				if hubName != "*" && !containsHub(hubs, hubName) {
					hubs = append(hubs, hubName)
				}
			}
		}
	}
	return hubs
}

// getEntityHub extracts hub name from entity tags
func getEntityHub(entity *models.Entity) string {
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "hub:") {
			return strings.TrimPrefix(tag, "hub:")
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

// FormatHubTag creates a hub tag
func FormatHubTag(hubName string) string {
	return fmt.Sprintf("hub:%s", hubName)
}

// FormatTraitTag creates a trait tag for a hub
func FormatTraitTag(hubName, namespace, value string) string {
	return fmt.Sprintf("%s:trait:%s:%s", hubName, namespace, value)
}

// FormatSelfTag creates a self tag for a hub
func FormatSelfTag(hubName, namespace, value string) string {
	return fmt.Sprintf("%s:self:%s:%s", hubName, namespace, value)
}

// ParseHubTag parses a hub tag and returns the hub name
func ParseHubTag(tag string) (string, bool) {
	if strings.HasPrefix(tag, "hub:") {
		return strings.TrimPrefix(tag, "hub:"), true
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