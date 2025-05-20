package api

// This file contains the bug fix for chunked content retrieval

import (
	"fmt"
	"entitydb/logger"
)

// HandleChunkedContent checks if an entity has chunked content and reassembles it if needed
// This function is called from the GetEntity handler
func (h *EntityHandler) HandleChunkedContent(entityID string, includeContent bool) ([]byte, error) {
	// If we're not including content, just return nil
	if !includeContent {
		return nil, nil
	}
	
	// Get the main entity
	entity, err := h.repo.GetByID(entityID)
	if err != nil {
		logger.Error("Failed to get entity %s: %v", entityID, err)
		return nil, err
	}
	
	// Check if this entity has chunked content
	metadata := entity.GetContentMetadata()
	numChunks, hasChunkCount := metadata["chunks"]
	
	if !hasChunkCount {
		// Not chunked, return the original content
		return entity.Content, nil
	}
	
	logger.Debug("Entity %s is chunked with %s chunks", entity.ID, numChunks)
	
	// Parse number of chunks
	chunkCount := 0
	fmt.Sscanf(numChunks, "%d", &chunkCount)
	
	if chunkCount <= 0 {
		logger.Error("Invalid chunk count: %s", numChunks)
		return entity.Content, nil
	}
	
	// Allocate buffer for reassembled content
	var reassembledContent []byte
	
	// Fetch all chunks and reassemble content
	for i := 0; i < chunkCount; i++ {
		chunkID := fmt.Sprintf("%s-chunk-%d", entity.ID, i)
		chunkEntity, err := h.repo.GetByID(chunkID)
		
		if err != nil {
			logger.Error("Failed to get chunk %s: %v", chunkID, err)
			continue
		}
		
		logger.Debug("Retrieved chunk %s with %d bytes", chunkID, len(chunkEntity.Content))
		reassembledContent = append(reassembledContent, chunkEntity.Content...)
	}
	
	logger.Info("Reassembled chunked content for entity %s, total size: %d bytes", 
		entity.ID, len(reassembledContent))
	
	return reassembledContent, nil
}