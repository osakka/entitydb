package api

import (
	"entitydb/logger"
	"fmt"
	"net/http"
	"strings"
)

// Direct chunk streaming approach to fix the chunked content retrieval issue

// StreamEntity handles direct streaming of entity content, including chunked entities
func (h *EntityHandler) StreamEntity(w http.ResponseWriter, r *http.Request) {
	// Get entity ID
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
	includeContent := r.URL.Query().Get("include_content") == "true" || r.URL.Query().Get("stream") == "true"
	if !includeContent {
		RespondError(w, http.StatusBadRequest, "Include content parameter is required")
		return
	}

	// Get content type from entity tags
	contentType := "application/octet-stream"
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "content:type:") {
			parts := strings.SplitN(tag, "content:type:", 2)
			if len(parts) == 2 {
				contentType = parts[1]
				break
			}
		}
	}

	// Check if this is a chunked entity by looking for chunks tag
	isChunked := false
	chunkCount := 0
	chunkSize := int64(0)
	totalSize := int64(0)

	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "content:chunks:") {
			parts := strings.SplitN(tag, "content:chunks:", 2)
			if len(parts) == 2 {
				isChunked = true
				fmt.Sscanf(parts[1], "%d", &chunkCount)
			}
		} else if strings.HasPrefix(tag, "content:chunk-size:") {
			parts := strings.SplitN(tag, "content:chunk-size:", 2)
			if len(parts) == 2 {
				fmt.Sscanf(parts[1], "%d", &chunkSize)
			}
		} else if strings.HasPrefix(tag, "content:size:") {
			parts := strings.SplitN(tag, "content:size:", 2)
			if len(parts) == 2 {
				fmt.Sscanf(parts[1], "%d", &totalSize)
			}
		}
	}

	logger.Debug("Entity %s: isChunked=%v, chunkCount=%d, chunkSize=%d, totalSize=%d",
		id, isChunked, chunkCount, chunkSize, totalSize)

	// Set response headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", id))

	if isChunked && chunkCount > 0 {
		// This is a chunked entity - stream chunks
		logger.Info("Streaming chunked entity: id=%s, chunks=%d, chunkSize=%d, totalSize=%d",
			id, chunkCount, chunkSize, totalSize)

		if totalSize > 0 {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", totalSize))
		}

		// Stream each chunk
		for i := 0; i < chunkCount; i++ {
			chunkID := fmt.Sprintf("%s-chunk-%d", entity.ID, i)
			logger.Debug("Fetching chunk %d/%d: %s", i+1, chunkCount, chunkID)
			
			chunkEntity, err := h.repo.GetByID(chunkID)
			if err != nil {
				logger.Error("Failed to get chunk %s: %v", chunkID, err)
				continue
			}
			
			logger.Debug("Retrieved chunk %d/%d with %d bytes", 
				i+1, chunkCount, len(chunkEntity.Content))
			
			// Write chunk content directly to response
			if _, err := w.Write(chunkEntity.Content); err != nil {
				logger.Error("Failed to write chunk to response: %v", err)
				return
			}
			
			// Flush after each chunk
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	} else {
		// Not chunked - stream the main entity's content
		if len(entity.Content) == 0 {
			RespondError(w, http.StatusNotFound, "Entity has no content")
			return
		}
		
		logger.Info("Streaming non-chunked entity: id=%s, contentSize=%d bytes", 
			id, len(entity.Content))
		
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(entity.Content)))
		
		// Write content directly
		if _, err := w.Write(entity.Content); err != nil {
			logger.Error("Failed to write content to response: %v", err)
			return
		}
	}
}