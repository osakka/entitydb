package api

import (
	"entitydb/logger"
	"fmt"
	"net/http"
	"strings"
)

// HandleChunkedEntityRetrieval is an enhanced handler for chunked entity retrieval
func (h *EntityHandler) HandleChunkedEntityRetrieval(w http.ResponseWriter, r *http.Request) {
	// Get entity ID
	id := r.URL.Query().Get("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}

	// Get main entity
	entity, err := h.repo.GetByID(id)
	if err != nil {
		logger.Error("Failed to get entity %s: %v", id, err)
		RespondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	// Check if this is a chunked entity
	if !entity.IsChunked() {
		RespondError(w, http.StatusBadRequest, "Entity is not chunked")
		return
	}

	// Get chunking info
	chunkInfo := GetChunkInfo(entity)
	logger.Info("Retrieving chunked entity: id=%s, chunks=%d, chunkSize=%d, totalSize=%d",
		id, chunkInfo.ChunkCount, chunkInfo.ChunkSize, chunkInfo.TotalSize)

	// Set headers for binary content streaming
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", chunkInfo.TotalSize))
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Stream chunks directly to response
	for i := 0; i < chunkInfo.ChunkCount; i++ {
		chunkID := fmt.Sprintf("%s-chunk-%d", entity.ID, i)
		chunkEntity, err := h.repo.GetByID(chunkID)
		
		if err != nil {
			logger.Error("Failed to get chunk %s: %v", chunkID, err)
			continue
		}
		
		logger.Debug("Retrieved chunk %d/%d with %d bytes", 
			i+1, chunkInfo.ChunkCount, len(chunkEntity.Content))
		
		// Write chunk content directly to response
		if _, err := w.Write(chunkEntity.Content); err != nil {
			logger.Error("Failed to write chunk to response: %v", err)
			return
		}
	}
}

// HandleRawChunkRetrieval handles retrieving a single chunk entity
func (h *EntityHandler) HandleRawChunkRetrieval(w http.ResponseWriter, r *http.Request) {
	// Get parent entity ID and chunk index
	parentID := r.URL.Query().Get("parent_id")
	chunkIndexStr := r.URL.Query().Get("chunk_index")
	
	if parentID == "" || chunkIndexStr == "" {
		RespondError(w, http.StatusBadRequest, "Parent ID and chunk index are required")
		return
	}
	
	// Parse chunk index
	var chunkIndex int
	fmt.Sscanf(chunkIndexStr, "%d", &chunkIndex)
	
	// Construct chunk ID
	chunkID := fmt.Sprintf("%s-chunk-%d", parentID, chunkIndex)
	
	// Get chunk entity
	chunkEntity, err := h.repo.GetByID(chunkID)
	if err != nil {
		logger.Error("Failed to get chunk %s: %v", chunkID, err)
		RespondError(w, http.StatusNotFound, "Chunk not found")
		return
	}
	
	// Set headers for binary content
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(chunkEntity.Content)))
	
	// Write chunk content to response
	if _, err := w.Write(chunkEntity.Content); err != nil {
		logger.Error("Failed to write chunk to response: %v", err)
		return
	}
}

// StreamChunkedEntityContent streams the content of a chunked entity directly to the client
func (h *EntityHandler) StreamChunkedEntityContent(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}
	
	// Get the main entity
	entity, err := h.repo.GetByID(id)
	if err != nil {
		logger.Error("Failed to get entity %s: %v", id, err)
		RespondError(w, http.StatusNotFound, "Entity not found")
		return
	}
	
	// Check if this is a chunked entity
	logger.Debug("Checking if entity %s is chunked: %v", id, entity.IsChunked())
	logger.Debug("Entity tags: %v", entity.Tags)
	
	// Get content metadata
	metadata := GetChunkInfo(entity)
	logger.Debug("Chunk info: chunks=%d, size=%d, totalSize=%d", 
		metadata.ChunkCount, metadata.ChunkSize, metadata.TotalSize)
	
	if metadata.ChunkCount <= 0 {
		// If not chunked, fall back to normal content response
		logger.Debug("Entity %s is not chunked or has invalid chunk count: %d", id, metadata.ChunkCount)
		if len(entity.Content) == 0 {
			RespondError(w, http.StatusNotFound, "Entity has no content")
			return
		}
		
		// Set headers for binary content
		contentType := "application/octet-stream"
		for _, tag := range entity.Tags {
			if strings.HasPrefix(tag, "content:type:") {
				contentType = strings.TrimPrefix(tag, "content:type:")
				break
			}
		}
		
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(entity.Content)))
		
		// Write content
		w.Write(entity.Content)
		return
	}
	
	// This is a chunked entity - stream it
	chunkInfo := GetChunkInfo(entity)
	
	// Set response headers
	contentType := "application/octet-stream"
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "content:type:") {
			contentType = strings.TrimPrefix(tag, "content:type:")
			break
		}
	}
	
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", chunkInfo.TotalSize))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", id))
	
	// Stream chunks directly to response
	for i := 0; i < chunkInfo.ChunkCount; i++ {
		chunkID := fmt.Sprintf("%s-chunk-%d", entity.ID, i)
		chunkEntity, err := h.repo.GetByID(chunkID)
		
		if err != nil {
			logger.Error("Failed to get chunk %s: %v", chunkID, err)
			continue
		}
		
		logger.Debug("Streaming chunk %d/%d with %d bytes", 
			i+1, chunkInfo.ChunkCount, len(chunkEntity.Content))
		
		if _, err := w.Write(chunkEntity.Content); err != nil {
			logger.Error("Failed to write chunk to response: %v", err)
			return
		}
		
		// Flush response to ensure client receives data
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}