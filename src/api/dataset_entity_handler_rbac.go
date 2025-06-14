package api

import (
	"entitydb/models"
	"net/http"
)

// DatasetEntityHandlerRBAC wraps EntityHandler with RBAC and Dataset permission checks
type DatasetEntityHandlerRBAC struct {
	handler         *EntityHandler
	repo            models.EntityRepository
	securityManager *models.SecurityManager
}

// NewDatasetEntityHandlerRBAC creates a new RBAC-enabled dataset entity handler
func NewDatasetEntityHandlerRBAC(handler *EntityHandler, repo models.EntityRepository, securityManager *models.SecurityManager) *DatasetEntityHandlerRBAC {
	return &DatasetEntityHandlerRBAC{
		handler:         handler,
		repo:            repo,
		securityManager: securityManager,
	}
}

// CreateDatasetEntity wraps CreateEntity with dataset context
func (h *DatasetEntityHandlerRBAC) CreateDatasetEntity() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.securityManager, PermEntityCreate)(h.handler.CreateEntity)
}

// QueryDatasetEntities wraps QueryEntities with dataset context
func (h *DatasetEntityHandlerRBAC) QueryDatasetEntities() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.securityManager, PermEntityView)(h.handler.QueryEntities)
}