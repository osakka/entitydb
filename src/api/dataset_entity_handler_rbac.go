package api

import (
	"entitydb/models"
	"net/http"
)

// DatasetEntityHandlerRBAC wraps EntityHandler with RBAC and Dataset permission checks
type DatasetEntityHandlerRBAC struct {
	handler        *EntityHandler
	repo           models.EntityRepository
	sessionManager *models.SessionManager
}

// NewDatasetEntityHandlerRBAC creates a new RBAC-enabled dataset entity handler
func NewDatasetEntityHandlerRBAC(handler *EntityHandler, repo models.EntityRepository, sessionManager *models.SessionManager) *DatasetEntityHandlerRBAC {
	return &DatasetEntityHandlerRBAC{
		handler:        handler,
		repo:           repo,
		sessionManager: sessionManager,
	}
}

// CreateDatasetEntity wraps CreateEntity with dataset context
func (h *DatasetEntityHandlerRBAC) CreateDatasetEntity() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermEntityCreate)(h.handler.CreateEntity)
}

// QueryDatasetEntities wraps QueryEntities with dataset context
func (h *DatasetEntityHandlerRBAC) QueryDatasetEntities() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermEntityView)(h.handler.QueryEntities)
}