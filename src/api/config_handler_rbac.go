package api

import (
	"entitydb/models"
	"net/http"
)

// EntityConfigHandlerRBAC wraps EntityConfigHandler with RBAC permission checks
type EntityConfigHandlerRBAC struct {
	handler        *EntityConfigHandler
	repo           models.EntityRepository
	sessionManager *models.SessionManager
}

// NewEntityConfigHandlerRBAC creates a new RBAC-enabled config handler
func NewEntityConfigHandlerRBAC(handler *EntityConfigHandler, repo models.EntityRepository, sessionManager *models.SessionManager) *EntityConfigHandlerRBAC {
	return &EntityConfigHandlerRBAC{
		handler:        handler,
		repo:           repo,
		sessionManager: sessionManager,
	}
}

// GetConfig wraps GetConfig with permission check
func (h *EntityConfigHandlerRBAC) GetConfig() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermConfigView)(h.handler.GetConfig)
}

// SetConfig wraps SetConfig with permission check
func (h *EntityConfigHandlerRBAC) SetConfig() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermConfigUpdate)(h.handler.SetConfig)
}

// GetFeatureFlags wraps GetFeatureFlags with permission check
func (h *EntityConfigHandlerRBAC) GetFeatureFlags() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermConfigView)(h.handler.GetFeatureFlags)
}

// SetFeatureFlag wraps SetFeatureFlag with permission check
func (h *EntityConfigHandlerRBAC) SetFeatureFlag() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermConfigUpdate)(h.handler.SetFeatureFlag)
}

// DashboardHandlerRBAC wraps dashboard operations with RBAC permission checks
type DashboardHandlerRBAC struct {
	handler        *DashboardHandler
	repo           models.EntityRepository
	sessionManager *models.SessionManager
}

// NewDashboardHandlerRBAC creates a new RBAC-enabled dashboard handler
func NewDashboardHandlerRBAC(handler *DashboardHandler, repo models.EntityRepository, sessionManager *models.SessionManager) *DashboardHandlerRBAC {
	return &DashboardHandlerRBAC{
		handler:        handler,
		repo:           repo,
		sessionManager: sessionManager,
	}
}

// GetDashboardStats wraps DashboardStats with permission check
func (h *DashboardHandlerRBAC) GetDashboardStats() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermSystemView)(h.handler.DashboardStats)
}