package api

import (
	"entitydb/models"
	"net/http"
)

// EntityRelationshipHandlerRBAC wraps EntityRelationshipHandler with RBAC permission checks
type EntityRelationshipHandlerRBAC struct {
	handler        *EntityRelationshipHandler
	repo           models.EntityRepository
	sessionManager *models.SessionManager
}

// NewEntityRelationshipHandlerRBAC creates a new RBAC-enabled relationship handler
func NewEntityRelationshipHandlerRBAC(handler *EntityRelationshipHandler, repo models.EntityRepository, sessionManager *models.SessionManager) *EntityRelationshipHandlerRBAC {
	return &EntityRelationshipHandlerRBAC{
		handler:        handler,
		repo:           repo,
		sessionManager: sessionManager,
	}
}

// CreateRelationshipWithRBAC wraps CreateRelationship with permission check
func (h *EntityRelationshipHandlerRBAC) CreateRelationshipWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermRelationCreate)(h.handler.CreateRelationship)
}

// GetRelationshipWithRBAC wraps GetRelationship with permission check
func (h *EntityRelationshipHandlerRBAC) GetRelationshipWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermRelationView)(h.handler.GetRelationship)
}

// ListRelationshipsBySourceWithRBAC wraps ListRelationshipsBySource with permission check
func (h *EntityRelationshipHandlerRBAC) ListRelationshipsBySourceWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermRelationView)(h.handler.ListRelationshipsBySource)
}

// UpdateRelationships doesn't exist, so we'll use the general handler
// HandleEntityRelationshipsWithRBAC wraps HandleEntityRelationships with permission check
func (h *EntityRelationshipHandlerRBAC) HandleEntityRelationshipsWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermRelationView)(h.handler.HandleEntityRelationships)
}

// DeleteRelationshipWithRBAC wraps DeleteRelationship with permission check
func (h *EntityRelationshipHandlerRBAC) DeleteRelationshipWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermRelationDelete)(h.handler.DeleteRelationship)
}