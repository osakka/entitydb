package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"entitydb/models"
)

// DatasetContext stores hub information in the request context
type DatasetContext struct {
	DatasetName        string
	UserHubs       []string // Hubs user has access to
	CanAccessHub   bool
	IsGlobalAdmin  bool
}

// Context key for hub data
type hubContextKey struct{}

// DatasetMiddleware creates middleware that enforces hub-based access control
func DatasetMiddleware(entityRepo models.EntityRepository, sessionManager *models.SessionManager) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get RBAC context (should be set by RBACMiddleware first)
			rbacCtx, hasRBAC := GetRBACContext(r)
			if !hasRBAC {
				RespondError(w, http.StatusUnauthorized, "RBAC context required")
				return
			}

			// Extract hub from query parameter or request body
			datasetName := extractHubFromRequest(r)
			if datasetName == "" {
				// For create operations, hub might be in request body
				// This will be handled in individual handlers
				datasetName = "default" // Allow for now, handlers will validate
			}

			// Check if user has access to this hub
			userHubs := getUserHubs(rbacCtx.Permissions)
			canAccessHub := rbacCtx.IsAdmin || containsHub(userHubs, datasetName) || datasetName == "default"

			// Create hub context
			hubCtx := &DatasetContext{
				DatasetName:       datasetName,
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

// GetDatasetContext retrieves the hub context from the request
func GetDatasetContext(r *http.Request) (*DatasetContext, bool) {
	ctx, ok := r.Context().Value(hubContextKey{}).(*DatasetContext)
	return ctx, ok
}

// ValidateEntityDataset validates that an entity belongs to a hub the user can access
func ValidateEntityDataset(rbacCtx *RBACContext, entity *models.Entity) error {
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
		if strings.HasPrefix(tag, "dataset:") {
			return tags // Hub tag already exists
		}
	}

	// Add default hub tag
	return append(tags, fmt.Sprintf("dataset:%s", defaultHub))
}

// CheckDatasetPermission checks if user has specific permission for a hub
func CheckDatasetPermission(rbacCtx *RBACContext, datasetName string, action string) bool {
	// Global admin has all permissions
	if rbacCtx.IsAdmin {
		return true
	}

	// Check hub-specific permission
	hubPerm := fmt.Sprintf("rbac:perm:entity:%s:dataset:%s", action, datasetName)
	if models.HasPermission(rbacCtx.Permissions, hubPerm) {
		return true
	}

	// Check general hub permission
	generalPerm := fmt.Sprintf("rbac:perm:entity:%s:dataset:*", action)
	return models.HasPermission(rbacCtx.Permissions, generalPerm)
}

// CheckHubManagementPermission checks hub management permissions
func CheckHubManagementPermission(rbacCtx *RBACContext, action string, datasetName string) bool {
	// Global admin has all permissions
	if rbacCtx.IsAdmin {
		return true
	}

	// Check specific hub management permission
	if datasetName != "" {
		hubPerm := fmt.Sprintf("rbac:perm:dataset:%s:%s", action, datasetName)
		if models.HasPermission(rbacCtx.Permissions, hubPerm) {
			return true
		}
	}

	// Check general hub management permission
	generalPerm := fmt.Sprintf("rbac:perm:dataset:%s", action)
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
		// Look for permissions like rbac:perm:entity:*:dataset:worcha
		if strings.Contains(perm, ":dataset:") {
			parts := strings.Split(perm, ":")
			if len(parts) >= 6 && parts[4] == "hub" {
				datasetName := parts[5]
				if datasetName != "*" && !containsHub(hubs, datasetName) {
					hubs = append(hubs, datasetName)
				}
			}
		}
	}
	return hubs
}

// getEntityHub extracts hub name from entity tags
func getEntityHub(entity *models.Entity) string {
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "dataset:") {
			return strings.TrimPrefix(tag, "dataset:")
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

// FormatDatasetTag creates a hub tag
func FormatDatasetTag(datasetName string) string {
	return fmt.Sprintf("dataset:%s", datasetName)
}

// FormatTraitTag creates a trait tag for a hub
func FormatTraitTag(datasetName, namespace, value string) string {
	return fmt.Sprintf("%s:trait:%s:%s", datasetName, namespace, value)
}

// FormatSelfTag creates a self tag for a hub
func FormatSelfTag(datasetName, namespace, value string) string {
	return fmt.Sprintf("%s:self:%s:%s", datasetName, namespace, value)
}

// ParseDatasetTag parses a hub tag and returns the hub name
func ParseDatasetTag(tag string) (string, bool) {
	if strings.HasPrefix(tag, "dataset:") {
		return strings.TrimPrefix(tag, "dataset:"), true
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