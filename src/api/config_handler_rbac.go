package api

import (
	"entitydb/models"
	"net/http"
)

// EntityConfigHandlerRBAC wraps EntityConfigHandler with RBAC permission checks
type EntityConfigHandlerRBAC struct {
	handler *EntityConfigHandler
	repo    models.EntityRepository
}

// NewEntityConfigHandlerRBAC creates a new RBAC-enabled config handler
func NewEntityConfigHandlerRBAC(handler *EntityConfigHandler, repo models.EntityRepository) *EntityConfigHandlerRBAC {
	return &EntityConfigHandlerRBAC{
		handler: handler,
		repo:    repo,
	}
}

// GetConfigWithRBAC wraps GetConfig with permission check
func (h *EntityConfigHandlerRBAC) GetConfigWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, PermConfigView)(h.handler.GetConfig)
}

// SetConfigWithRBAC wraps SetConfig with permission check
func (h *EntityConfigHandlerRBAC) SetConfigWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, PermConfigUpdate)(h.handler.SetConfig)
}

// GetFeatureFlagsWithRBAC wraps GetFeatureFlags with permission check
func (h *EntityConfigHandlerRBAC) GetFeatureFlagsWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, PermConfigView)(h.handler.GetFeatureFlags)
}

// SetFeatureFlagWithRBAC wraps SetFeatureFlag with permission check
func (h *EntityConfigHandlerRBAC) SetFeatureFlagWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, PermConfigUpdate)(h.handler.SetFeatureFlag)
}

// DashboardHandlerRBAC wraps dashboard operations with RBAC permission checks
type DashboardHandlerRBAC struct {
	handler *DashboardHandler
	repo    models.EntityRepository
}

// NewDashboardHandlerRBAC creates a new RBAC-enabled dashboard handler
func NewDashboardHandlerRBAC(handler *DashboardHandler, repo models.EntityRepository) *DashboardHandlerRBAC {
	return &DashboardHandlerRBAC{
		handler: handler,
		repo:    repo,
	}
}

// GetDashboardStatsWithRBAC wraps DashboardStats with permission check
func (h *DashboardHandlerRBAC) GetDashboardStatsWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, PermSystemView)(h.handler.DashboardStats)
}