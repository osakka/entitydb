package api

import (
	"entitydb/models"
	"net/http"
)

// EntityRelationshipHandlerRBAC wraps EntityRelationshipHandler with RBAC permission checks
type EntityRelationshipHandlerRBAC struct {
	handler *EntityRelationshipHandler
	repo    models.EntityRepository
}

// NewEntityRelationshipHandlerRBAC creates a new RBAC-enabled relationship handler
func NewEntityRelationshipHandlerRBAC(handler *EntityRelationshipHandler, repo models.EntityRepository) *EntityRelationshipHandlerRBAC {
	return &EntityRelationshipHandlerRBAC{
		handler: handler,
		repo:    repo,
	}
}

// CreateRelationshipWithRBAC wraps CreateRelationship with permission check
func (h *EntityRelationshipHandlerRBAC) CreateRelationshipWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, PermRelationCreate)(h.handler.CreateRelationship)
}

// GetRelationshipWithRBAC wraps GetRelationship with permission check
func (h *EntityRelationshipHandlerRBAC) GetRelationshipWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, PermRelationView)(h.handler.GetRelationship)
}

// ListRelationshipsBySourceWithRBAC wraps ListRelationshipsBySource with permission check
func (h *EntityRelationshipHandlerRBAC) ListRelationshipsBySourceWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, PermRelationView)(h.handler.ListRelationshipsBySource)
}

// UpdateRelationships doesn't exist, so we'll use the general handler
// HandleEntityRelationshipsWithRBAC wraps HandleEntityRelationships with permission check
func (h *EntityRelationshipHandlerRBAC) HandleEntityRelationshipsWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, PermRelationView)(h.handler.HandleEntityRelationships)
}

// DeleteRelationshipWithRBAC wraps DeleteRelationship with permission check
func (h *EntityRelationshipHandlerRBAC) DeleteRelationshipWithRBAC() http.HandlerFunc {
	return RBACMiddleware(h.repo, PermRelationDelete)(h.handler.DeleteRelationship)
}