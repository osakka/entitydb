package api

import (
	"entitydb/models"
	"net/http"
)

// DataspaceEntityHandlerRBAC wraps EntityHandler with RBAC and Dataspace permission checks
type DataspaceEntityHandlerRBAC struct {
	handler        *EntityHandler
	repo           models.EntityRepository
	sessionManager *models.SessionManager
}

// NewDataspaceEntityHandlerRBAC creates a new RBAC-enabled dataspace entity handler
func NewDataspaceEntityHandlerRBAC(handler *EntityHandler, repo models.EntityRepository, sessionManager *models.SessionManager) *DataspaceEntityHandlerRBAC {
	return &DataspaceEntityHandlerRBAC{
		handler:        handler,
		repo:           repo,
		sessionManager: sessionManager,
	}
}

// CreateDataspaceEntityWithRBAC wraps CreateEntity with dataspace context
func (h *DataspaceEntityHandlerRBAC) CreateDataspaceEntityWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermEntityCreate)(h.handler.CreateEntity)
}

// QueryDataspaceEntitiesWithRBAC wraps QueryEntities with dataspace context
func (h *DataspaceEntityHandlerRBAC) QueryDataspaceEntitiesWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermEntityView)(h.handler.QueryEntities)
}