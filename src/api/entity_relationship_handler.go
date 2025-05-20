package api

import (
	"entitydb/models"
	"entitydb/logger"
	"encoding/json"
	"net/http"
)

// EntityRelationshipHandler provides API endpoints for working with entity relationships
type EntityRelationshipHandler struct {
	repo models.EntityRelationshipRepository
}

// NewEntityRelationshipHandler creates a new entity relationship handler
func NewEntityRelationshipHandler(repo models.EntityRelationshipRepository) *EntityRelationshipHandler {
	return &EntityRelationshipHandler{
		repo: repo,
	}
}

// CreateRelationship creates a new entity relationship
// @Summary Create relationship
// @Description Create a new relationship between two entities
// @Tags relationships
// @Accept json
// @Produce json
// @Param body body RelationshipRequest true "Relationship data"
// @Success 201 {object} RelationshipResponse
// @Router /api/v1/entity-relationships [post]
func (h *EntityRelationshipHandler) CreateRelationship(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		SourceID         string                 `json:"source_id"`
		RelationshipType string                 `json:"relationship_type"`
		TargetID         string                 `json:"target_id"`
		Metadata         map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SourceID == "" || req.RelationshipType == "" || req.TargetID == "" {
		RespondError(w, http.StatusBadRequest, "Source ID, relationship type, and target ID are required")
		return
	}

	// Create a new relationship
	relationship := models.NewEntityRelationship(
		req.SourceID,
		req.RelationshipType,
		req.TargetID,
	)

	// Extract and set created by from authorization
	username := ""
	// Extract username from context instead
	// Skip auth for now
	username = "system"

	if username != "" {
		relationship.SetCreatedBy(username)
	} else {
		relationship.SetCreatedBy("system")
	}

	// Add metadata if provided
	if req.Metadata != nil && len(req.Metadata) > 0 {
		err := relationship.AddMetadata(req.Metadata)
		if err != nil {
			logger.Debug("Error adding metadata to relationship: %v", err)
		}
	}

	// Save the relationship
	err := h.repo.Create(relationship)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create relationship: "+err.Error())
		return
	}

	// Return the created relationship
	RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"relationship": relationship,
		"success":      true,
	})
}

// DeleteRelationship deletes an entity relationship
// @Summary Delete relationship
// @Description Delete an existing relationship between two entities
// @Tags relationships
// @Accept json
// @Produce json
// @Param source_id query string true "Source entity ID"
// @Param relationship_type query string true "Type of relationship"
// @Param target_id query string true "Target entity ID"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/entity-relationships [delete]
func (h *EntityRelationshipHandler) DeleteRelationship(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	sourceID := r.URL.Query().Get("source_id")
	relationshipType := r.URL.Query().Get("relationship_type")
	targetID := r.URL.Query().Get("target_id")

	if sourceID == "" || relationshipType == "" || targetID == "" {
		RespondError(w, http.StatusBadRequest, "Source ID, relationship type, and target ID are required")
		return
	}

	// Delete the relationship
	err := h.repo.Delete(sourceID, relationshipType, targetID)
	if err != nil {
		if err == models.ErrNotFound {
			RespondError(w, http.StatusNotFound, "Relationship not found")
		} else {
			RespondError(w, http.StatusInternalServerError, "Failed to delete relationship: "+err.Error())
		}
		return
	}

	// Return success
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// GetRelationship gets a specific entity relationship
// @Summary Get relationship
// @Description Get a specific relationship between two entities
// @Tags relationships
// @Accept json
// @Produce json
// @Param source_id query string true "Source entity ID"
// @Param relationship_type query string true "Type of relationship"
// @Param target_id query string true "Target entity ID"
// @Success 200 {object} RelationshipResponse
// @Router /api/v1/entity-relationships [get]
func (h *EntityRelationshipHandler) GetRelationship(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	sourceID := r.URL.Query().Get("source_id")
	relationshipType := r.URL.Query().Get("relationship_type")
	targetID := r.URL.Query().Get("target_id")

	if sourceID == "" || relationshipType == "" || targetID == "" {
		RespondError(w, http.StatusBadRequest, "Source ID, relationship type, and target ID are required")
		return
	}

	// Get the relationship
	relationship, err := h.repo.GetRelationship(sourceID, relationshipType, targetID)
	if err != nil {
		if err == models.ErrNotFound {
			RespondError(w, http.StatusNotFound, "Relationship not found")
		} else {
			RespondError(w, http.StatusInternalServerError, "Failed to get relationship: "+err.Error())
		}
		return
	}

	// Get metadata
	metadata, err := relationship.GetMetadata()
	if err != nil {
		logger.Debug("Error parsing relationship metadata: %v", err)
		metadata = make(map[string]interface{})
	}

	// Return the relationship
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"relationship": map[string]interface{}{
			"source_id":         relationship.SourceID,
			"relationship_type": relationship.RelationshipType,
			"target_id":         relationship.TargetID,
			"created_at":        relationship.CreatedAt,
			"created_by":        relationship.CreatedBy,
			"metadata":          metadata,
		},
	})
}

// ListRelationshipsBySource lists all relationships where the entity is the source
// @Summary List relationships by source
// @Description List all relationships where the specified entity is the source
// @Tags relationships
// @Accept json
// @Produce json
// @Param source_id query string true "Source entity ID"
// @Param relationship_type query string false "Filter by relationship type"
// @Success 200 {object} RelationshipListResponse
// @Router /api/v1/entity-relationships/by-source [get]
func (h *EntityRelationshipHandler) ListRelationshipsBySource(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	sourceID := r.URL.Query().Get("source_id")
	if sourceID == "" {
		RespondError(w, http.StatusBadRequest, "Source ID is required")
		return
	}

	// Get the relationships
	var relationships []*models.EntityRelationship
	var err error

	// If relationship type is provided, filter by type
	relationshipType := r.URL.Query().Get("relationship_type")
	if relationshipType != "" {
		relationships, err = h.repo.GetBySourceAndType(sourceID, relationshipType)
	} else {
		relationships, err = h.repo.GetBySource(sourceID)
	}

	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to list relationships: "+err.Error())
		return
	}

	// Format response
	response := make([]map[string]interface{}, 0, len(relationships))
	for _, relationship := range relationships {
		// Get metadata
		metadata, _ := relationship.GetMetadata()
		
		response = append(response, map[string]interface{}{
			"source_id":         relationship.SourceID,
			"relationship_type": relationship.RelationshipType,
			"target_id":         relationship.TargetID,
			"created_at":        relationship.CreatedAt,
			"created_by":        relationship.CreatedBy,
			"metadata":          metadata,
		})
	}

	// Return the relationships
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"relationships": response,
		"total":         len(response),
	})
}

// ListRelationshipsByTarget lists all relationships where the entity is the target
// @Summary List relationships by target
// @Description List all relationships where the specified entity is the target
// @Tags relationships
// @Accept json
// @Produce json
// @Param target_id query string true "Target entity ID"
// @Param relationship_type query string false "Filter by relationship type"
// @Success 200 {object} RelationshipListResponse
// @Router /api/v1/entity-relationships/by-target [get]
func (h *EntityRelationshipHandler) ListRelationshipsByTarget(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	targetID := r.URL.Query().Get("target_id")
	if targetID == "" {
		RespondError(w, http.StatusBadRequest, "Target ID is required")
		return
	}

	// Get the relationships
	var relationships []*models.EntityRelationship
	var err error

	// If relationship type is provided, filter by type
	relationshipType := r.URL.Query().Get("relationship_type")
	if relationshipType != "" {
		relationships, err = h.repo.GetByTargetAndType(targetID, relationshipType)
	} else {
		relationships, err = h.repo.GetByTarget(targetID)
	}

	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to list relationships: "+err.Error())
		return
	}

	// Format response
	response := make([]map[string]interface{}, 0, len(relationships))
	for _, relationship := range relationships {
		// Get metadata
		metadata, _ := relationship.GetMetadata()
		
		response = append(response, map[string]interface{}{
			"source_id":         relationship.SourceID,
			"relationship_type": relationship.RelationshipType,
			"target_id":         relationship.TargetID,
			"created_at":        relationship.CreatedAt,
			"created_by":        relationship.CreatedBy,
			"metadata":          metadata,
		})
	}

	// Return the relationships
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"relationships": response,
		"total":         len(response),
	})
}

// ListRelationshipsByType lists all relationships of a specific type
// @Summary List relationships by type
// @Description List all relationships of a specific type
// @Tags relationships
// @Accept json
// @Produce json
// @Param relationship_type query string true "Type of relationship"
// @Success 200 {object} RelationshipListResponse
// @Router /api/v1/entity-relationships/by-type [get]
func (h *EntityRelationshipHandler) ListRelationshipsByType(w http.ResponseWriter, r *http.Request) {
	// Parse request parameters
	relationshipType := r.URL.Query().Get("relationship_type")
	if relationshipType == "" {
		RespondError(w, http.StatusBadRequest, "Relationship type is required")
		return
	}

	// Get the relationships
	relationships, err := h.repo.GetByType(relationshipType)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to list relationships: "+err.Error())
		return
	}

	// Format response
	response := make([]map[string]interface{}, 0, len(relationships))
	for _, relationship := range relationships {
		// Get metadata
		metadata, _ := relationship.GetMetadata()
		
		response = append(response, map[string]interface{}{
			"source_id":         relationship.SourceID,
			"relationship_type": relationship.RelationshipType,
			"target_id":         relationship.TargetID,
			"created_at":        relationship.CreatedAt,
			"created_by":        relationship.CreatedBy,
			"metadata":          metadata,
		})
	}

	// Return the relationships
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"relationships": response,
		"total":         len(response),
	})
}
// HandleEntityRelationships handles the main entity relationships endpoint
func (h *EntityRelationshipHandler) HandleEntityRelationships(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.GetRelationship(w, r)
	case "POST":
		h.CreateRelationship(w, r)
	case "DELETE":
		h.DeleteRelationship(w, r)
	default:
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
