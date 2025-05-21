package api

import (
	"entitydb/logger"
	"net/http"
)

// GetEntityImproved is an enhanced version of GetEntity that properly handles chunked content
func (h *EntityHandler) GetEntityImproved(w http.ResponseWriter, r *http.Request) {
	// Check if timestamps should be included in response
	includeTimestamps := r.URL.Query().Get("include_timestamps") == "true"
	
	// Get entity ID from query parameter
	id := r.URL.Query().Get("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}

	// Get entity from repository
	entity, err := h.repo.GetByID(id)
	if err != nil {
		logger.Error("Failed to get entity %s: %v", id, err)
		RespondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	// Check if content should be included
	includeContent := r.URL.Query().Get("include_content") == "true"
	
	// Check if this is a chunked entity and if we need to reassemble chunks
	if includeContent && entity.IsChunked() {
		logger.Info("Entity %s is chunked, using improved chunk handler", id)
		
		// Use our improved chunk handler for all chunked entity retrievals
		h.ImprovedChunkHandler(w, r)
		return
	}

	// For non-chunked entities, just return the entity as is
	response := h.stripTimestampsFromEntity(entity, includeTimestamps)
	
	// Log content details for debugging
	logger.Debug("Retrieved non-chunked entity %s with %d bytes of content and %d tags", 
		entity.ID, len(entity.Content), len(entity.Tags))
	
	// No need to manually base64 encode - JSON marshaling handles []byte automatically
	RespondJSON(w, http.StatusOK, response)
}