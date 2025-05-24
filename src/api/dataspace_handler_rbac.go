package api

import (
	"entitydb/models"
	"net/http"
)

// DataspaceHandlerRBAC wraps DataspaceHandler with RBAC checks
type DataspaceHandlerRBAC struct {
	handler *DataspaceHandler
	repo    models.EntityRepository
	sm      *models.SessionManager
}

// NewDataspaceHandlerRBAC creates a new RBAC-wrapped dataspace handler
func NewDataspaceHandlerRBAC(handler *DataspaceHandler, repo models.EntityRepository, sm *models.SessionManager) *DataspaceHandlerRBAC {
	return &DataspaceHandlerRBAC{
		handler: handler,
		repo:    repo,
		sm:      sm,
	}
}

// ListDataspaces with RBAC check
func (h *DataspaceHandlerRBAC) ListDataspaces(w http.ResponseWriter, r *http.Request) {
	RBACMiddleware(h.repo, h.sm, RBACPermission{Resource: "dataspace", Action: "view"})(h.handler.ListDataspaces)(w, r)
}

// GetDataspace with RBAC check
func (h *DataspaceHandlerRBAC) GetDataspace(w http.ResponseWriter, r *http.Request) {
	RBACMiddleware(h.repo, h.sm, RBACPermission{Resource: "dataspace", Action: "view"})(h.handler.GetDataspace)(w, r)
}

// CreateDataspace with RBAC check
func (h *DataspaceHandlerRBAC) CreateDataspace(w http.ResponseWriter, r *http.Request) {
	RBACMiddleware(h.repo, h.sm, RBACPermission{Resource: "dataspace", Action: "create"})(h.handler.CreateDataspace)(w, r)
}

// UpdateDataspace with RBAC check
func (h *DataspaceHandlerRBAC) UpdateDataspace(w http.ResponseWriter, r *http.Request) {
	RBACMiddleware(h.repo, h.sm, RBACPermission{Resource: "dataspace", Action: "update"})(h.handler.UpdateDataspace)(w, r)
}

// DeleteDataspace with RBAC check
func (h *DataspaceHandlerRBAC) DeleteDataspace(w http.ResponseWriter, r *http.Request) {
	RBACMiddleware(h.repo, h.sm, RBACPermission{Resource: "dataspace", Action: "delete"})(h.handler.DeleteDataspace)(w, r)
}