package api

import (
	"entitydb/models"
	"net/http"
)

// DatasetManagementHandlerRBAC wraps DatasetManagementHandler with RBAC permission checks
type DatasetManagementHandlerRBAC struct {
	handler        *DatasetManagementHandler
	repo           models.EntityRepository
	sessionManager *models.SessionManager
}

// NewDatasetManagementHandlerRBAC creates a new RBAC-enabled dataset management handler
func NewDatasetManagementHandlerRBAC(handler *DatasetManagementHandler, repo models.EntityRepository, sessionManager *models.SessionManager) *DatasetManagementHandlerRBAC {
	return &DatasetManagementHandlerRBAC{
		handler:        handler,
		repo:           repo,
		sessionManager: sessionManager,
	}
}

// Define dataset management permissions
var (
	PermDatasetCreate = RBACPermission{Resource: "dataset", Action: "create"}
	PermDatasetView   = RBACPermission{Resource: "dataset", Action: "view"}
	PermDatasetDelete = RBACPermission{Resource: "dataset", Action: "delete"}
)

// CreateDataset wraps CreateDataset with permission check
func (h *DatasetManagementHandlerRBAC) CreateDataset() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermDatasetCreate)(h.handler.CreateDataset)
}

// ListDatasets wraps ListDatasets with permission check
func (h *DatasetManagementHandlerRBAC) ListDatasets() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermDatasetView)(h.handler.ListDatasets)
}

// DeleteDataset wraps DeleteDataset with permission check
func (h *DatasetManagementHandlerRBAC) DeleteDataset() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermDatasetDelete)(h.handler.DeleteDataset)
}