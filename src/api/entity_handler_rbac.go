package api

import (
	"entitydb/models"
	"net/http"
)

// EntityHandlerRBAC wraps EntityHandler with RBAC permission checks
type EntityHandlerRBAC struct {
	handler        *EntityHandler
	repo           models.EntityRepository
	sessionManager *models.SessionManager
}

// NewEntityHandlerRBAC creates a new RBAC-enabled entity handler
func NewEntityHandlerRBAC(handler *EntityHandler, repo models.EntityRepository, sessionManager *models.SessionManager) *EntityHandlerRBAC {
	return &EntityHandlerRBAC{
		handler:        handler,
		repo:           repo,
		sessionManager: sessionManager,
	}
}

// CreateEntityWithRBAC wraps CreateEntity with permission check
func (h *EntityHandlerRBAC) CreateEntityWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityCreate)(h.handler.CreateEntity)
}

// GetEntityWithRBAC wraps GetEntity with permission check
func (h *EntityHandlerRBAC) GetEntityWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetEntity)
}

// ListEntitiesWithRBAC wraps ListEntities with permission check
func (h *EntityHandlerRBAC) ListEntitiesWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.ListEntitiesDataspaceAware)
}

// UpdateEntityWithRBAC wraps UpdateEntity with permission check
func (h *EntityHandlerRBAC) UpdateEntityWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityUpdate)(h.handler.UpdateEntity)
}

// QueryEntitiesWithRBAC wraps QueryEntities with permission check
func (h *EntityHandlerRBAC) QueryEntitiesWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.QueryEntities)
}

// GetEntityAsOfWithRBAC wraps GetEntityAsOf with permission check
func (h *EntityHandlerRBAC) GetEntityAsOfWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetEntityAsOf)
}

// GetEntityHistoryWithRBAC wraps GetEntityHistory with permission check
func (h *EntityHandlerRBAC) GetEntityHistoryWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetEntityHistory)
}

// GetRecentChangesWithRBAC wraps GetRecentChanges with permission check
func (h *EntityHandlerRBAC) GetRecentChangesWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetRecentChanges)
}

// GetEntityDiffWithRBAC wraps GetEntityDiff with permission check
func (h *EntityHandlerRBAC) GetEntityDiffWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetEntityDiff)
}

// GetEntityTimeseriesWithRBAC wraps GetEntityTimeseries with permission check
func (h *EntityHandlerRBAC) GetEntityTimeseriesWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetEntityTimeseries)
}

// SimpleCreateEntityWithRBAC wraps SimpleCreateEntity with permission check
func (h *EntityHandlerRBAC) SimpleCreateEntityWithRBAC() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityCreate)(h.handler.SimpleCreateEntity)
}