package api

import (
	"entitydb/models"
	"net/http"
)

// HubEntityHandlerRBAC wraps EntityHandler with RBAC and Hub permission checks
type HubEntityHandlerRBAC struct {
	handler        *EntityHandler
	repo           models.EntityRepository
	sessionManager *models.SessionManager
}

// NewHubEntityHandlerRBAC creates a new RBAC-enabled hub entity handler
func NewHubEntityHandlerRBAC(handler *EntityHandler, repo models.EntityRepository, sessionManager *models.SessionManager) *HubEntityHandlerRBAC {
	return &HubEntityHandlerRBAC{
		handler:        handler,
		repo:           repo,
		sessionManager: sessionManager,
	}
}

// CreateHubEntityWithRBAC wraps CreateHubEntity with permission check
func (h *HubEntityHandlerRBAC) CreateHubEntityWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermEntityCreate)(h.handler.CreateHubEntity)
}

// QueryHubEntitiesWithRBAC wraps QueryHubEntities with permission check
func (h *HubEntityHandlerRBAC) QueryHubEntitiesWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermEntityView)(h.handler.QueryHubEntities)
}