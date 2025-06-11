package api

import (
	"entitydb/models"
	"net/http"
)

// DatasetHandlerRBAC wraps DatasetHandler with RBAC checks
type DatasetHandlerRBAC struct {
	handler *DatasetHandler
	repo    models.EntityRepository
	sm      *models.SessionManager
}

// NewDatasetHandlerRBAC creates a new RBAC-wrapped dataset handler
func NewDatasetHandlerRBAC(handler *DatasetHandler, repo models.EntityRepository, sm *models.SessionManager) *DatasetHandlerRBAC {
	return &DatasetHandlerRBAC{
		handler: handler,
		repo:    repo,
		sm:      sm,
	}
}

// ListDatasets with RBAC check
func (h *DatasetHandlerRBAC) ListDatasets(w http.ResponseWriter, r *http.Request) {
	RBACMiddleware(h.repo, h.sm, RBACPermission{Resource: "dataset", Action: "view"})(h.handler.ListDatasets)(w, r)
}

// GetDataset with RBAC check
func (h *DatasetHandlerRBAC) GetDataset(w http.ResponseWriter, r *http.Request) {
	RBACMiddleware(h.repo, h.sm, RBACPermission{Resource: "dataset", Action: "view"})(h.handler.GetDataset)(w, r)
}

// CreateDataset with RBAC check
func (h *DatasetHandlerRBAC) CreateDataset(w http.ResponseWriter, r *http.Request) {
	RBACMiddleware(h.repo, h.sm, RBACPermission{Resource: "dataset", Action: "create"})(h.handler.CreateDataset)(w, r)
}

// UpdateDataset with RBAC check
func (h *DatasetHandlerRBAC) UpdateDataset(w http.ResponseWriter, r *http.Request) {
	RBACMiddleware(h.repo, h.sm, RBACPermission{Resource: "dataset", Action: "update"})(h.handler.UpdateDataset)(w, r)
}

// DeleteDataset with RBAC check
func (h *DatasetHandlerRBAC) DeleteDataset(w http.ResponseWriter, r *http.Request) {
	RBACMiddleware(h.repo, h.sm, RBACPermission{Resource: "dataset", Action: "delete"})(h.handler.DeleteDataset)(w, r)
}