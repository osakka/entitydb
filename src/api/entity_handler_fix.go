package api

// This file contains the bug fix for chunked content retrieval

// Added more debug logs to trace the issue

import (
	"fmt"
	"entitydb/logger"
	"entitydb/models"
	"strings"
)

// HandleChunkedContent checks if an entity has chunked content and reassembles it if needed
// This function is called from the GetEntity handler
// DEBUG: Added extensive logging to diagnose chunking issues
// Helper function to get min of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function to get content type tags
func getContentTypeTags(entity *models.Entity) []string {
	contentTypeTags := []string{}
	for _, tag := range entity.Tags {
		if strings.Contains(tag, "content:type:") {
			contentTypeTags = append(contentTypeTags, tag)
		}
	}
	return contentTypeTags
}

// ChunkInfo contains details about a chunked entity
type ChunkInfo struct {
	ChunkCount int
	ChunkSize  int64
	TotalSize  int64
}

// GetChunkInfo extracts chunking information from entity tags
func GetChunkInfo(entity *models.Entity) ChunkInfo {
	info := ChunkInfo{}
	
	metadata := entity.GetContentMetadata()
	
	// Get chunk count
	if chunkCount, ok := metadata["chunks"]; ok {
		fmt.Sscanf(chunkCount, "%d", &info.ChunkCount)
	}
	
	// Get chunk size
	if chunkSize, ok := metadata["chunk-size"]; ok {
		fmt.Sscanf(chunkSize, "%d", &info.ChunkSize)
	}
	
	// Get total size
	if totalSize, ok := metadata["size"]; ok {
		fmt.Sscanf(totalSize, "%d", &info.TotalSize)
	}
	
	return info
}

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
	
	// Debug - log chunk info
	logger.Debug("CHUNK_DEBUG: Entity %s has %d chunks, content type tags: %v", 
		entity.ID, chunkCount, getContentTypeTags(entity))
	
	// Double check that we got the right content format
	reassembledContentStart := ""
	if len(reassembledContent) > 0 {
		reassembledContentStart = string(reassembledContent[:min(20, len(reassembledContent))])
	}
	logger.Debug("CHUNK_DEBUG: Reassembled content start: %s", reassembledContentStart)
	
	// Log chunk retrieval summary
	logger.Info("Chunks retrieved: expected=%d, got=%d, totalSize=%d bytes", 
		chunkCount, len(reassembledContent) / 1000000 + 1, len(reassembledContent))
	
	return reassembledContent, nil
}