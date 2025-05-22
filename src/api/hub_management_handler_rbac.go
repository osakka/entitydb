package api

import (
	"entitydb/models"
	"net/http"
)

// HubManagementHandlerRBAC wraps HubManagementHandler with RBAC permission checks
type HubManagementHandlerRBAC struct {
	handler        *HubManagementHandler
	repo           models.EntityRepository
	sessionManager *models.SessionManager
}

// NewHubManagementHandlerRBAC creates a new RBAC-enabled hub management handler
func NewHubManagementHandlerRBAC(handler *HubManagementHandler, repo models.EntityRepository, sessionManager *models.SessionManager) *HubManagementHandlerRBAC {
	return &HubManagementHandlerRBAC{
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

// CreateHubWithRBAC wraps CreateHub with permission check
func (h *HubManagementHandlerRBAC) CreateHubWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermHubCreate)(h.handler.CreateHub)
}

// ListHubsWithRBAC wraps ListHubs with permission check
func (h *HubManagementHandlerRBAC) ListHubsWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermHubView)(h.handler.ListHubs)
}

// DeleteHubWithRBAC wraps DeleteHub with permission check
func (h *HubManagementHandlerRBAC) DeleteHubWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermHubDelete)(h.handler.DeleteHub)
}