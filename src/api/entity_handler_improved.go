package api

import (
	"entitydb/logger"
	"net/http"
	"strings"
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
		// Check if the request prefers streaming (better for large files)
		if r.URL.Query().Get("stream") == "true" {
			// Stream content directly to the client
			logger.Info("Streaming chunked content for entity %s", id)
			h.StreamChunkedEntityContent(w, r)
			return
		}
		
		// New approach: direct chunk retrieval and binary response
		if r.URL.Query().Get("raw") == "true" {
			logger.Info("Raw chunk retrieval for entity %s", id)
			h.HandleChunkedEntityRetrieval(w, r)
			return
		}
		
		// Otherwise, use the standard reassembly approach with enhanced error handling
		reassembledContent, err := h.HandleChunkedContent(id, includeContent)
		if err == nil && len(reassembledContent) > 0 {
			// Direct binary content assignment to prevent JSON serialization issues
			entity.Content = reassembledContent
			logger.Info("Using reassembled content for entity %s: %d bytes", entity.ID, len(entity.Content))
			
			// Ensure that the content type tag is set correctly for binary data
			hasContentTypeTag := false
			for _, tag := range entity.Tags {
				if strings.HasSuffix(tag, "content:type:application/octet-stream") {
					hasContentTypeTag = true
					break
				}
			}
			
			// Add content type tag if not present
			if !hasContentTypeTag {
				entity.AddTag("content:type:application/octet-stream")
			}
		} else {
			logger.Warn("Failed to reassemble content for entity %s: err=%v, contentLen=%d", 
				id, err, len(reassembledContent))
			
			// Set an empty content field instead of nil to prevent JSON encoding issues
			entity.Content = []byte{}
		}
	}

	// Return entity
	response := h.stripTimestampsFromEntity(entity, includeTimestamps)
	
	// Log content details for debugging
	logger.Debug("Retrieved entity %s with %d bytes of content and %d tags", 
		entity.ID, len(entity.Content), len(entity.Tags))
	
	// No need to manually base64 encode - JSON marshaling handles []byte automatically
	RespondJSON(w, http.StatusOK, response)
}