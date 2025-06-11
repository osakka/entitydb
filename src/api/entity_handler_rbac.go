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

// CreateEntity wraps CreateEntity with permission check
func (h *EntityHandlerRBAC) CreateEntity() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityCreate)(h.handler.CreateEntity)
}

// GetEntity wraps GetEntity with permission check
func (h *EntityHandlerRBAC) GetEntity() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetEntity)
}

// ListEntities wraps ListEntities with permission check
func (h *EntityHandlerRBAC) ListEntities() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.ListEntitiesDatasetAware)
}

// UpdateEntity wraps UpdateEntity with permission check
func (h *EntityHandlerRBAC) UpdateEntity() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityUpdate)(h.handler.UpdateEntity)
}

// QueryEntities wraps QueryEntities with permission check
func (h *EntityHandlerRBAC) QueryEntities() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.QueryEntities)
}

// GetEntityAsOf wraps GetEntityAsOf with permission check
func (h *EntityHandlerRBAC) GetEntityAsOf() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetEntityAsOf)
}

// GetEntityHistory wraps GetEntityHistory with permission check
func (h *EntityHandlerRBAC) GetEntityHistory() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetEntityHistory)
}

// GetRecentChanges wraps GetRecentChanges with permission check
func (h *EntityHandlerRBAC) GetRecentChanges() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetRecentChanges)
}

// GetEntityDiff wraps GetEntityDiff with permission check
func (h *EntityHandlerRBAC) GetEntityDiff() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetEntityDiff)
}

// GetEntityTimeseries wraps GetEntityTimeseries with permission check
func (h *EntityHandlerRBAC) GetEntityTimeseries() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetEntityTimeseries)
}

// SimpleCreateEntity wraps SimpleCreateEntity with permission check
func (h *EntityHandlerRBAC) SimpleCreateEntity() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityCreate)(h.handler.SimpleCreateEntity)
}

// ListByTag wraps ListByTag with permission check
func (h *EntityHandlerRBAC) ListByTag() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.ListEntities)
}

// GetEntityChanges wraps GetEntityChanges with permission check
func (h *EntityHandlerRBAC) GetEntityChanges() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.GetRecentChanges)
}

// GetChunk wraps GetChunk with permission check
func (h *EntityHandlerRBAC) GetChunk() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.HandleRawChunkRetrieval)
}

// StreamContent wraps StreamContent with permission check
func (h *EntityHandlerRBAC) StreamContent() http.HandlerFunc {
	return RequirePermission(h.sessionManager, h.repo, PermEntityView)(h.handler.StreamChunkedEntityContent)
}