package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"entitydb/logger"
	"entitydb/models"
)

// ImprovedChunkHandler handles chunked entity retrieval with better error handling and validation
func (h *EntityHandler) ImprovedChunkHandler(w http.ResponseWriter, r *http.Request) {
	// Get entity ID from query parameter
	id := r.URL.Query().Get("id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}

	// Check if streaming is requested
	streamMode := r.URL.Query().Get("stream") == "true"
	
	// Get entity from repository
	entity, err := h.repo.GetByID(id)
	if err != nil {
		logger.Error("Failed to get entity %s: %v", id, err)
		RespondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	// Only proceed if this is actually a chunked entity
	if !entity.IsChunked() {
		logger.Warn("Entity %s is not chunked, cannot reassemble chunks", id)
		RespondJSON(w, http.StatusOK, entity)
		return
	}

	// Extract chunk info
	chunkInfo, err := extractChunkInfo(entity)
	if err != nil {
		logger.Error("Failed to extract chunk info for entity %s: %v", id, err)
		RespondError(w, http.StatusInternalServerError, "Invalid chunk metadata")
		return
	}

	// Based on mode, either stream or reassemble
	if streamMode {
		h.streamChunks(w, r, entity, chunkInfo)
	} else {
		h.reassembleAndServeChunks(w, r, entity, chunkInfo)
	}
}

// ChunkMetadata contains details about a chunked entity
type ChunkMetadata struct {
	ChunkCount   int
	ChunkSize    int64
	TotalSize    int64
	ContentType  string
	Checksum     string
	ChunkMapping map[int]string // Maps chunk index to chunk ID
}

// Extract chunk info from entity tags
func extractChunkInfo(entity *models.Entity) (*ChunkMetadata, error) {
	info := &ChunkMetadata{
		ChunkMapping: make(map[int]string),
	}
	
	// Get metadata from entity tags
	metadata := entity.GetContentMetadata()
	
	// Get chunk count
	chunksStr, ok := metadata["chunks"]
	if !ok {
		return nil, errors.New("missing chunks tag")
	}
	chunkCount, err := strconv.Atoi(chunksStr)
	if err != nil || chunkCount <= 0 {
		return nil, fmt.Errorf("invalid chunk count: %s", chunksStr)
	}
	info.ChunkCount = chunkCount
	
	// Get chunk size
	chunkSizeStr, ok := metadata["chunk-size"]
	if !ok {
		return nil, errors.New("missing chunk-size tag")
	}
	chunkSize, err := strconv.ParseInt(chunkSizeStr, 10, 64)
	if err != nil || chunkSize <= 0 {
		return nil, fmt.Errorf("invalid chunk size: %s", chunkSizeStr)
	}
	info.ChunkSize = chunkSize
	
	// Get total size
	sizeStr, ok := metadata["size"]
	if !ok {
		return nil, errors.New("missing content size tag")
	}
	totalSize, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || totalSize <= 0 {
		return nil, fmt.Errorf("invalid total size: %s", sizeStr)
	}
	info.TotalSize = totalSize
	
	// Get content type
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if strings.HasPrefix(tag, "content:type:") {
			info.ContentType = strings.TrimPrefix(tag, "content:type:")
			break
		}
	}
	if info.ContentType == "" {
		info.ContentType = "application/octet-stream" // default
	}
	
	// Get checksum if available
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if strings.HasPrefix(tag, "content:checksum:sha256:") {
			info.Checksum = strings.TrimPrefix(tag, "content:checksum:sha256:")
			break
		}
	}
	
	// Create chunk mapping
	for i := 0; i < chunkCount; i++ {
		info.ChunkMapping[i] = fmt.Sprintf("%s-chunk-%d", entity.ID, i)
	}
	
	return info, nil
}

// Check if entity is properly chunked
func isEntityChunked(entity *models.Entity) bool {
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if strings.HasPrefix(tag, "content:chunks:") {
			// Get chunk count
			chunksStr := strings.TrimPrefix(tag, "content:chunks:")
			chunkCount, err := strconv.Atoi(chunksStr)
			return err == nil && chunkCount > 0
		}
	}
	return false
}

// Stream chunks directly to the client
func (h *EntityHandler) streamChunks(w http.ResponseWriter, r *http.Request, entity *models.Entity, info *ChunkMetadata) {
	// Set content type and length headers
	w.Header().Set("Content-Type", info.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(info.TotalSize, 10))
	w.Header().Set("X-Entity-ID", entity.ID)
	w.Header().Set("X-Entity-Chunks", strconv.Itoa(info.ChunkCount))
	
	// Stream chunks
	bytesWritten := int64(0)
	for i := 0; i < info.ChunkCount; i++ {
		chunkID := info.ChunkMapping[i]
		chunkEntity, err := h.repo.GetByID(chunkID)
		if err != nil {
			logger.Error("Failed to get chunk %s: %v", chunkID, err)
			// If we can't get a chunk, we need to abort - partial content is dangerous
			// We've already started writing, so just close the connection
			http.Error(w, "Failed to retrieve chunk", http.StatusInternalServerError)
			return
		}
		
		// Stream chunk content
		n, err := w.Write(chunkEntity.Content)
		if err != nil {
			logger.Error("Failed to write chunk %s: %v", chunkID, err)
			return
		}
		
		bytesWritten += int64(n)
		logger.Debug("Streamed chunk %d/%d, %d bytes", i+1, info.ChunkCount, n)
	}
	
	// Validate total bytes written
	if bytesWritten != info.TotalSize {
		logger.Error("Incomplete stream: wrote %d bytes, expected %d", bytesWritten, info.TotalSize)
		// We can't do much at this point as the response has already been sent
	} else {
		logger.Info("Successfully streamed all %d chunks, total %d bytes", info.ChunkCount, bytesWritten)
	}
}

// Reassemble chunks and serve in a single response
func (h *EntityHandler) reassembleAndServeChunks(w http.ResponseWriter, r *http.Request, entity *models.Entity, info *ChunkMetadata) {
	// Start with an explicit capacity to avoid reallocations
	reassembledContent := bytes.NewBuffer(make([]byte, 0, info.TotalSize))
	
	// Fetch and validate all chunks
	successfulChunks := 0
	
	// Use a waitgroup to fetch chunks concurrently
	var wg sync.WaitGroup
	chunkContents := make([][]byte, info.ChunkCount)
	chunkErrors := make([]error, info.ChunkCount)
	
	// Fetch chunks concurrently with a reasonable limit
	maxConcurrent := 4
	sem := make(chan struct{}, maxConcurrent)
	
	for i := 0; i < info.ChunkCount; i++ {
		wg.Add(1)
		go func(chunkIndex int) {
			defer wg.Done()
			
			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()
			
			chunkID := info.ChunkMapping[chunkIndex]
			startTime := time.Now()
			chunkEntity, err := h.repo.GetByID(chunkID)
			if err != nil {
				logger.Error("Failed to get chunk %d (%s): %v", chunkIndex, chunkID, err)
				chunkErrors[chunkIndex] = err
				return
			}
			
			chunkContents[chunkIndex] = chunkEntity.Content
			logger.Debug("Retrieved chunk %d/%d (%s) with %d bytes in %v", 
				chunkIndex+1, info.ChunkCount, chunkID, len(chunkEntity.Content), time.Since(startTime))
		}(i)
	}
	
	// Wait for all fetches to complete
	wg.Wait()
	
	// Check for errors and assemble in order
	for i := 0; i < info.ChunkCount; i++ {
		if chunkErrors[i] != nil {
			logger.Error("Failed to retrieve all chunks, chunk %d is missing", i)
			RespondError(w, http.StatusInternalServerError, "Failed to retrieve all chunks")
			return
		}
		
		// Write chunk to buffer
		reassembledContent.Write(chunkContents[i])
		successfulChunks++
	}
	
	// Verify total size
	if int64(reassembledContent.Len()) != info.TotalSize {
		logger.Error("Size mismatch: got %d bytes, expected %d bytes", 
			reassembledContent.Len(), info.TotalSize)
		RespondError(w, http.StatusInternalServerError, "Content size mismatch")
		return
	}
	
	// Verify checksum if available
	if info.Checksum != "" {
		hasher := sha256.New()
		hasher.Write(reassembledContent.Bytes())
		actualChecksum := hex.EncodeToString(hasher.Sum(nil))
		
		if actualChecksum != info.Checksum {
			logger.Error("Checksum mismatch: got %s, expected %s", 
				actualChecksum, info.Checksum)
			RespondError(w, http.StatusInternalServerError, "Content checksum mismatch")
			return
		}
	}
	
	logger.Info("Successfully reassembled %d chunks into %d bytes of content",
		successfulChunks, reassembledContent.Len())
	
	// Determine whether to send binary or include in entity
	if r.URL.Query().Get("raw") == "true" {
		// Send raw binary response
		w.Header().Set("Content-Type", info.ContentType)
		w.Header().Set("Content-Length", strconv.FormatInt(info.TotalSize, 10))
		w.Header().Set("X-Entity-ID", entity.ID)
		io.Copy(w, reassembledContent)
	} else {
		// Create a new copy of the entity with the content
		responseEntity := *entity
		responseEntity.Content = reassembledContent.Bytes()
		
		// Return JSON response
		RespondJSON(w, http.StatusOK, &responseEntity)
	}
}