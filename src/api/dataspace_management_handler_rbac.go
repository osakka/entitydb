package api

import (
	"entitydb/models"
	"net/http"
)

// DataspaceManagementHandlerRBAC wraps DataspaceManagementHandler with RBAC permission checks
type DataspaceManagementHandlerRBAC struct {
	handler        *DataspaceManagementHandler
	repo           models.EntityRepository
	sessionManager *models.SessionManager
}

// NewDataspaceManagementHandlerRBAC creates a new RBAC-enabled hub management handler
func NewDataspaceManagementHandlerRBAC(handler *DataspaceManagementHandler, repo models.EntityRepository, sessionManager *models.SessionManager) *DataspaceManagementHandlerRBAC {
	return &DataspaceManagementHandlerRBAC{
		handler:        handler,
		repo:           repo,
		sessionManager: sessionManager,
	}
}

// Define hub management permissions
var (
	PermHubCreate = RBACPermission{Resource: "hub", Action: "create"}
	PermHubView   = RBACPermission{Resource: "hub", Action: "view"}
	PermHubDelete = RBACPermission{Resource: "hub", Action: "delete"}
)

// CreateDataspaceWithRBAC wraps CreateDataspace with permission check
func (h *DataspaceManagementHandlerRBAC) CreateDataspaceWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermHubCreate)(h.handler.CreateDataspace)
}

// ListDataspacesWithRBAC wraps ListDataspaces with permission check
func (h *DataspaceManagementHandlerRBAC) ListDataspacesWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermHubView)(h.handler.ListDataspaces)
}

// DeleteDataspaceWithRBAC wraps DeleteDataspace with permission check
func (h *DataspaceManagementHandlerRBAC) DeleteDataspaceWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermHubDelete)(h.handler.DeleteDataspace)
}