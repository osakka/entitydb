// Package api provides deletion management endpoints with RBAC integration
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"entitydb/logger"
	"entitydb/models"
	"entitydb/services"
	"entitydb/storage/binary"

	"github.com/gorilla/mux"
)

// DeletionHandler manages entity deletion operations with lifecycle states
type DeletionHandler struct {
	repository        models.EntityRepository
	collector         *services.DeletionCollector
	securityMiddleware *SecurityMiddleware
}

// NewDeletionHandler creates a new deletion handler instance
func NewDeletionHandler(repo models.EntityRepository, collector *services.DeletionCollector, security *SecurityMiddleware) *DeletionHandler {
	return &DeletionHandler{
		repository:        repo,
		collector:         collector,
		securityMiddleware: security,
	}
}

// =============================================================================
// Request/Response Types
// =============================================================================

// SoftDeleteRequest represents a request to soft delete an entity
// @Description Request body for soft deleting an entity
type SoftDeleteRequest struct {
	// Reason for deletion (required)
	Reason string `json:"reason" binding:"required" example:"No longer needed"`
	
	// Policy to apply (optional, uses default if not specified)
	Policy string `json:"policy,omitempty" example:"standard-cleanup"`
	
	// Force deletion even if entity has relationships (default: false)
	Force bool `json:"force,omitempty" example:"false"`
}

// RestoreRequest represents a request to restore a deleted entity
// @Description Request body for restoring a deleted entity
type RestoreRequest struct {
	// Reason for restoration (required)
	Reason string `json:"reason" binding:"required" example:"Accidentally deleted"`
}

// DeletionStatusResponse represents the deletion status of an entity
// @Description Response containing entity deletion status and audit trail
type DeletionStatusResponse struct {
	// Entity ID
	EntityID string `json:"entity_id" example:"user-123"`
	
	// Current lifecycle state
	LifecycleState models.EntityLifecycleState `json:"lifecycle_state" example:"soft_deleted"`
	
	// When the entity was created
	CreatedAt time.Time `json:"created_at" example:"2023-01-15T10:30:00Z"`
	
	// When the entity was last updated
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-20T14:45:00Z"`
	
	// When the entity was deleted (if applicable)
	DeletedAt *time.Time `json:"deleted_at,omitempty" example:"2023-01-25T09:15:00Z"`
	
	// Who deleted the entity (if applicable)
	DeletedBy string `json:"deleted_by,omitempty" example:"admin"`
	
	// Reason for deletion (if applicable)
	DeleteReason string `json:"delete_reason,omitempty" example:"Data cleanup"`
	
	// When the entity was archived (if applicable)
	ArchivedAt *time.Time `json:"archived_at,omitempty" example:"2023-02-25T09:15:00Z"`
	
	// Retention policy applied (if applicable)
	RetentionPolicy string `json:"retention_policy,omitempty" example:"standard-cleanup"`
	
	// Whether the entity can be restored
	CanRestore bool `json:"can_restore" example:"true"`
	
	// Whether the entity can be purged
	CanPurge bool `json:"can_purge" example:"false"`
}

// DeletionListResponse represents a list of deleted entities
// @Description Response containing a list of deleted entities with pagination
type DeletionListResponse struct {
	// List of deleted entities
	Entities []DeletionStatusResponse `json:"entities"`
	
	// Total number of deleted entities
	Total int `json:"total" example:"150"`
	
	// Number of entities returned in this response
	Count int `json:"count" example:"25"`
	
	// Offset used for pagination
	Offset int `json:"offset" example:"0"`
	
	// Limit used for pagination
	Limit int `json:"limit" example:"25"`
}

// PurgeRequest represents a request to permanently purge an entity
// @Description Request body for permanently purging an entity (irreversible)
type PurgeRequest struct {
	// Confirmation string (must be "PURGE" to proceed)
	Confirmation string `json:"confirmation" binding:"required" example:"PURGE"`
	
	// Reason for purging (required)
	Reason string `json:"reason" binding:"required" example:"Legal requirement"`
}

// =============================================================================
// API Endpoints
// =============================================================================

// SoftDeleteEntity marks an entity as soft deleted
// @Summary Soft delete an entity
// @Description Marks an entity as soft deleted while preserving data for potential restoration
// @Tags Entity Deletion
// @Accept json
// @Produce json
// @Param id path string true "Entity ID"
// @Param request body SoftDeleteRequest true "Deletion request"
// @Success 200 {object} DeletionStatusResponse "Entity successfully soft deleted"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 403 {object} ErrorResponse "Insufficient permissions"
// @Failure 404 {object} ErrorResponse "Entity not found"
// @Failure 409 {object} ErrorResponse "Entity already deleted"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /entities/{id}/delete [post]
func (h *DeletionHandler) SoftDeleteEntity(w http.ResponseWriter, r *http.Request) {
	// Extract entity ID from URL
	vars := mux.Vars(r)
	entityID := vars["id"]
	
	if entityID == "" {
		http.Error(w, "Entity ID is required", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var req SoftDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("SoftDeleteEntity.parsing failed for %s: %v", entityID, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if req.Reason == "" {
		http.Error(w, "Deletion reason is required", http.StatusBadRequest)
		return
	}
	
	// Get current user from context
	user, ok := r.Context().Value("user").(*models.Entity)
	if !ok {
		logger.Error("SoftDeleteEntity.context user not found for %s", entityID)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	
	// Get entity to delete
	entity, err := h.repository.GetByID(entityID)
	if err != nil {
		logger.Warn("SoftDeleteEntity.entity_not_found %s: %v", entityID, err)
		http.Error(w, "Entity not found", http.StatusNotFound)
		return
	}
	
	// Check if entity is already deleted
	if entity.GetLifecycleState() != models.StateActive {
		logger.Warn("SoftDeleteEntity.already_deleted %s: current state %s", entityID, entity.GetLifecycleState())
		http.Error(w, fmt.Sprintf("Entity is already %s", entity.GetLifecycleState()), http.StatusConflict)
		return
	}
	
	// Check for relationships if not forced
	if !req.Force {
		// TODO: Implement relationship checking
		// For now, we'll allow deletion but log a warning
		logger.Debug("SoftDeleteEntity.skip_relationship_check %s: force=%v", entityID, req.Force)
	}
	
	// Apply deletion using entity lifecycle methods
	now := time.Now()
	
	// Add lifecycle status tag
	statusTag := fmt.Sprintf("lifecycle:state:%s", models.StateSoftDeleted)
	entity.AddTag(statusTag)
	
	// Add deletion metadata tags
	deletedByTag := fmt.Sprintf("lifecycle:deleted_by:%s", user.ID)
	entity.AddTag(deletedByTag)
	
	reasonTag := fmt.Sprintf("lifecycle:delete_reason:%s", req.Reason)
	entity.AddTag(reasonTag)
	
	deletedAtTag := fmt.Sprintf("lifecycle:deleted_at:%d", now.UnixNano())
	entity.AddTag(deletedAtTag)
	
	// Add policy tag if specified
	if req.Policy != "" {
		policyTag := fmt.Sprintf("lifecycle:policy:%s", req.Policy)
		entity.AddTag(policyTag)
	}
	
	// Update entity in repository
	if err := h.repository.Update(entity); err != nil {
		logger.Error("SoftDeleteEntity.update_failed %s: %v", entityID, err)
		http.Error(w, "Failed to update entity", http.StatusInternalServerError)
		return
	}
	
	logger.Info("SoftDeleteEntity.success %s: deleted by %s, reason: %s", entityID, user.ID, req.Reason)
	
	// Return deletion status
	status := h.buildDeletionStatusResponse(entity)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// RestoreEntity restores a soft deleted entity to active state
// @Summary Restore a deleted entity
// @Description Restores a soft deleted entity back to active state
// @Tags Entity Deletion
// @Accept json
// @Produce json
// @Param id path string true "Entity ID"
// @Param request body RestoreRequest true "Restoration request"
// @Success 200 {object} DeletionStatusResponse "Entity successfully restored"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 403 {object} ErrorResponse "Insufficient permissions"
// @Failure 404 {object} ErrorResponse "Entity not found"
// @Failure 409 {object} ErrorResponse "Entity cannot be restored"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /entities/{id}/restore [post]
func (h *DeletionHandler) RestoreEntity(w http.ResponseWriter, r *http.Request) {
	// Extract entity ID from URL
	vars := mux.Vars(r)
	entityID := vars["id"]
	
	if entityID == "" {
		http.Error(w, "Entity ID is required", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var req RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("RestoreEntity.parsing failed for %s: %v", entityID, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if req.Reason == "" {
		http.Error(w, "Restoration reason is required", http.StatusBadRequest)
		return
	}
	
	// Get current user from context
	user, ok := r.Context().Value("user").(*models.Entity)
	if !ok {
		logger.Error("RestoreEntity.context user not found for %s", entityID)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	
	// Get entity to restore
	entity, err := h.repository.GetByID(entityID)
	if err != nil {
		logger.Warn("RestoreEntity.entity_not_found %s: %v", entityID, err)
		http.Error(w, "Entity not found", http.StatusNotFound)
		return
	}
	
	// Check if entity can be restored
	currentState := entity.GetLifecycleState()
	if currentState != models.StateSoftDeleted {
		logger.Warn("RestoreEntity.cannot_restore %s: current state %s", entityID, currentState)
		if currentState == models.StateActive {
			http.Error(w, "Entity is already active", http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Entity in state %s cannot be restored", currentState), http.StatusConflict)
		}
		return
	}
	
	// Apply restoration using entity lifecycle methods
	now := time.Now()
	
	// Remove existing lifecycle state tags and add active state
	h.updateLifecycleState(entity, models.StateActive)
	
	// Add restoration metadata tags
	restoredByTag := fmt.Sprintf("lifecycle:restored_by:%s", user.ID)
	entity.AddTag(restoredByTag)
	
	restoredAtTag := fmt.Sprintf("lifecycle:restored_at:%d", now.UnixNano())
	entity.AddTag(restoredAtTag)
	
	restoreReasonTag := fmt.Sprintf("lifecycle:restore_reason:%s", req.Reason)
	entity.AddTag(restoreReasonTag)
	
	// Update entity in repository
	if err := h.repository.Update(entity); err != nil {
		logger.Error("RestoreEntity.update_failed %s: %v", entityID, err)
		http.Error(w, "Failed to update entity", http.StatusInternalServerError)
		return
	}
	
	logger.Info("RestoreEntity.success %s: restored by %s, reason: %s", entityID, user.ID, req.Reason)
	
	// Return deletion status
	status := h.buildDeletionStatusResponse(entity)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetDeletionStatus returns the deletion status of an entity
// @Summary Get entity deletion status
// @Description Returns detailed deletion status and audit trail for an entity
// @Tags Entity Deletion
// @Produce json
// @Param id path string true "Entity ID"
// @Success 200 {object} DeletionStatusResponse "Entity deletion status"
// @Failure 403 {object} ErrorResponse "Insufficient permissions"
// @Failure 404 {object} ErrorResponse "Entity not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /entities/{id}/deletion-status [get]
func (h *DeletionHandler) GetDeletionStatus(w http.ResponseWriter, r *http.Request) {
	// Extract entity ID from URL
	vars := mux.Vars(r)
	entityID := vars["id"]
	
	if entityID == "" {
		http.Error(w, "Entity ID is required", http.StatusBadRequest)
		return
	}
	
	// Get entity
	entity, err := h.repository.GetByID(entityID)
	if err != nil {
		logger.Warn("GetDeletionStatus.entity_not_found %s: %v", entityID, err)
		http.Error(w, "Entity not found", http.StatusNotFound)
		return
	}
	
	// Build and return deletion status
	status := h.buildDeletionStatusResponse(entity)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// ListDeletedEntities returns a paginated list of deleted entities
// @Summary List deleted entities
// @Description Returns a paginated list of entities in various deletion states
// @Tags Entity Deletion
// @Produce json
// @Param state query string false "Lifecycle state filter" Enums(soft_deleted,archived,purged) example("soft_deleted")
// @Param limit query int false "Maximum number of entities to return" minimum(1) maximum(100) default(25)
// @Param offset query int false "Number of entities to skip" minimum(0) default(0)
// @Param deleted_by query string false "Filter by who deleted the entities" example("admin")
// @Success 200 {object} DeletionListResponse "List of deleted entities"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 403 {object} ErrorResponse "Insufficient permissions"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /entities/deleted [get]
func (h *DeletionHandler) ListDeletedEntities(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	stateFilter := r.URL.Query().Get("state")
	deletedByFilter := r.URL.Query().Get("deleted_by")
	
	// Parse pagination parameters
	limit := 25
	offset := 0
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	// Build filter tags
	var filterTags []string
	
	// Add state filter
	if stateFilter != "" {
		if !models.IsValidState(stateFilter) {
			http.Error(w, "Invalid lifecycle state", http.StatusBadRequest)
			return
		}
		filterTags = append(filterTags, fmt.Sprintf("lifecycle:state:%s", stateFilter))
	} else {
		// Default to all non-active states
		filterTags = append(filterTags, "lifecycle:state:soft_deleted")
		filterTags = append(filterTags, "lifecycle:state:archived")
		filterTags = append(filterTags, "lifecycle:state:purged")
	}
	
	// Add deleted_by filter
	if deletedByFilter != "" {
		filterTags = append(filterTags, fmt.Sprintf("lifecycle:deleted_by:%s", deletedByFilter))
	}
	
	// Query entities using OR logic for multiple state filters
	allEntities := make([]*models.Entity, 0)
	
	if stateFilter != "" {
		// Single state filter
		entities, err := h.repository.ListByTag(fmt.Sprintf("lifecycle:state:%s", stateFilter))
		if err != nil {
			logger.Error("ListDeletedEntities.query_failed for state %s: %v", stateFilter, err)
			http.Error(w, "Failed to query entities", http.StatusInternalServerError)
			return
		}
		allEntities = entities
	} else {
		// Multiple state filters - query each and merge
		states := []string{"soft_deleted", "archived", "purged"}
		entityMap := make(map[string]*models.Entity)
		
		for _, state := range states {
			entities, err := h.repository.ListByTag(fmt.Sprintf("lifecycle:state:%s", state))
			if err != nil {
				logger.Warn("ListDeletedEntities.query_partial_fail for state %s: %v", state, err)
				continue
			}
			
			for _, entity := range entities {
				entityMap[entity.ID] = entity
			}
		}
		
		// Convert map to slice
		for _, entity := range entityMap {
			allEntities = append(allEntities, entity)
		}
	}
	
	// Apply additional filters
	if deletedByFilter != "" {
		filtered := make([]*models.Entity, 0)
		for _, entity := range allEntities {
			if entity.GetTagValue("lifecycle:deleted_by") == deletedByFilter {
				filtered = append(filtered, entity)
			}
		}
		allEntities = filtered
	}
	
	// Apply pagination
	total := len(allEntities)
	start := offset
	end := offset + limit
	
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	
	paginatedEntities := allEntities[start:end]
	
	// Build response
	response := DeletionListResponse{
		Entities: make([]DeletionStatusResponse, len(paginatedEntities)),
		Total:    total,
		Count:    len(paginatedEntities),
		Offset:   offset,
		Limit:    limit,
	}
	
	for i, entity := range paginatedEntities {
		response.Entities[i] = h.buildDeletionStatusResponse(entity)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PurgeEntity permanently removes an entity and all its data
// @Summary Permanently purge an entity
// @Description Permanently and irreversibly removes an entity and all its data from the system
// @Tags Entity Deletion
// @Accept json
// @Produce json
// @Param id path string true "Entity ID"
// @Param request body PurgeRequest true "Purge request with confirmation"
// @Success 200 {object} SuccessResponse "Entity successfully purged"
// @Failure 400 {object} ErrorResponse "Invalid request or confirmation"
// @Failure 403 {object} ErrorResponse "Insufficient permissions"
// @Failure 404 {object} ErrorResponse "Entity not found"
// @Failure 409 {object} ErrorResponse "Entity cannot be purged"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /entities/{id}/purge [delete]
func (h *DeletionHandler) PurgeEntity(w http.ResponseWriter, r *http.Request) {
	// Extract entity ID from URL
	vars := mux.Vars(r)
	entityID := vars["id"]
	
	if entityID == "" {
		http.Error(w, "Entity ID is required", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var req PurgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("PurgeEntity.parsing failed for %s: %v", entityID, err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate confirmation
	if req.Confirmation != "PURGE" {
		logger.Warn("PurgeEntity.invalid_confirmation %s: got %s", entityID, req.Confirmation)
		http.Error(w, "Invalid confirmation - must be 'PURGE'", http.StatusBadRequest)
		return
	}
	
	// Validate reason
	if req.Reason == "" {
		http.Error(w, "Purge reason is required", http.StatusBadRequest)
		return
	}
	
	// Get current user from context
	user, ok := r.Context().Value("user").(*models.Entity)
	if !ok {
		logger.Error("PurgeEntity.context user not found for %s", entityID)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	
	// Get entity to purge
	entity, err := h.repository.GetByID(entityID)
	if err != nil {
		logger.Warn("PurgeEntity.entity_not_found %s: %v", entityID, err)
		http.Error(w, "Entity not found", http.StatusNotFound)
		return
	}
	
	// Check if entity can be purged (must be archived or soft deleted)
	currentState := entity.GetLifecycleState()
	if currentState != models.StateArchived && currentState != models.StateSoftDeleted {
		logger.Warn("PurgeEntity.cannot_purge %s: current state %s", entityID, currentState)
		http.Error(w, fmt.Sprintf("Entity in state %s cannot be purged", currentState), http.StatusConflict)
		return
	}
	
	// Log purge operation before deletion
	logger.Info("PurgeEntity.executing %s: purged by %s, reason: %s, state: %s", 
		entityID, user.ID, req.Reason, currentState)
	
	// If the repository supports deletion indexing, add purge entry
	if binaryRepo, ok := h.repository.(*binary.EntityRepository); ok {
		// Create deletion entry for audit trail
		deletionEntry := binary.NewDeletionEntry(
			entityID,
			models.StatePurged,
			user.ID,
			req.Reason,
			entity.GetTagValue("lifecycle:policy"),
			time.Now().UnixNano(),
		)
		
		// Add to deletion index (this will be preserved even after entity removal)
		logger.Debug("PurgeEntity.adding_deletion_entry %s", entityID)
		if err := binaryRepo.AddDeletionEntry(deletionEntry); err != nil {
			logger.Warn("PurgeEntity.deletion_entry_failed %s: %v", entityID, err)
			// Continue with purge even if deletion entry fails
		}
	}
	
	// Permanently remove entity from repository
	if err := h.repository.Delete(entityID); err != nil {
		logger.Error("PurgeEntity.delete_failed %s: %v", entityID, err)
		http.Error(w, "Failed to purge entity", http.StatusInternalServerError)
		return
	}
	
	logger.Info("PurgeEntity.success %s: permanently removed by %s", entityID, user.ID)
	
	// Return success response
	response := SuccessResponse{
		Success: true,
		Message: fmt.Sprintf("Entity %s has been permanently purged", entityID),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// =============================================================================
// Helper Methods
// =============================================================================

// buildDeletionStatusResponse creates a DeletionStatusResponse from an entity
func (h *DeletionHandler) buildDeletionStatusResponse(entity *models.Entity) DeletionStatusResponse {
	status := DeletionStatusResponse{
		EntityID:       entity.ID,
		LifecycleState: entity.GetLifecycleState(),
		CreatedAt:      time.Unix(0, entity.CreatedAt),
		UpdatedAt:      time.Unix(0, entity.UpdatedAt),
	}
	
	// Extract deletion metadata
	if deletedBy := entity.GetTagValue("lifecycle:deleted_by"); deletedBy != "" {
		status.DeletedBy = deletedBy
	}
	
	if deleteReason := entity.GetTagValue("lifecycle:delete_reason"); deleteReason != "" {
		status.DeleteReason = deleteReason
	}
	
	if policy := entity.GetTagValue("lifecycle:policy"); policy != "" {
		status.RetentionPolicy = policy
	}
	
	// Extract timestamps
	if deletedAt := entity.GetDeletedAt(); deletedAt != nil {
		status.DeletedAt = deletedAt
	}
	
	if archivedAt := entity.GetArchivedAt(); archivedAt != nil {
		status.ArchivedAt = archivedAt
	}
	
	// Determine what operations are allowed
	switch status.LifecycleState {
	case models.StateActive:
		status.CanRestore = false
		status.CanPurge = false
	case models.StateSoftDeleted:
		status.CanRestore = true
		status.CanPurge = true
	case models.StateArchived:
		status.CanRestore = false
		status.CanPurge = true
	case models.StatePurged:
		status.CanRestore = false
		status.CanPurge = false
	}
	
	return status
}

// updateLifecycleState updates an entity's lifecycle state by removing old state tags and adding new one
func (h *DeletionHandler) updateLifecycleState(entity *models.Entity, newState models.EntityLifecycleState) {
	// Get current tags and filter out lifecycle:state: tags
	currentTags := entity.GetTagsWithoutTimestamp()
	filteredTags := make([]string, 0, len(currentTags))
	
	for _, tag := range currentTags {
		if !strings.HasPrefix(tag, "lifecycle:state:") {
			filteredTags = append(filteredTags, tag)
		}
	}
	
	// Add new lifecycle state tag
	newStateTag := fmt.Sprintf("lifecycle:state:%s", newState)
	filteredTags = append(filteredTags, newStateTag)
	
	// Update entity tags
	entity.SetTags(filteredTags)
}

