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

// CreateRelationship wraps CreateRelationship with permission check
func (h *EntityRelationshipHandlerRBAC) CreateRelationship() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermRelationCreate)(h.handler.CreateRelationship)
}

// GetRelationship wraps GetRelationship with permission check
func (h *EntityRelationshipHandlerRBAC) GetRelationship() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermRelationView)(h.handler.GetRelationship)
}

// ListRelationshipsBySource wraps ListRelationshipsBySource with permission check
func (h *EntityRelationshipHandlerRBAC) ListRelationshipsBySource() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermRelationView)(h.handler.ListRelationshipsBySource)
}

// UpdateRelationships doesn't exist, so we'll use the general handler
// HandleEntityRelationships wraps HandleEntityRelationships with permission check
func (h *EntityRelationshipHandlerRBAC) HandleEntityRelationships() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermRelationView)(h.handler.HandleEntityRelationships)
}

// DeleteRelationship wraps DeleteRelationship with permission check
func (h *EntityRelationshipHandlerRBAC) DeleteRelationship() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.sessionManager, PermRelationDelete)(h.handler.DeleteRelationship)
}