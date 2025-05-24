package api

import (
	"encoding/json"
	"entitydb/models"
	"entitydb/logger"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// OptimizedEntityHandler handles entity operations with performance optimizations
type OptimizedEntityHandler struct {
	repo models.EntityRepository
}

// NewOptimizedEntityHandler creates an optimized entity handler
func NewOptimizedEntityHandler(repo models.EntityRepository) *OptimizedEntityHandler {
	return &OptimizedEntityHandler{
		repo: repo,
	}
}

// ListEntitiesOptimized lists entities with minimal overhead
func (h *OptimizedEntityHandler) ListEntitiesOptimized(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Set response headers early
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Response-Time", "0") // Will update later
	
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	// sortBy := r.URL.Query().Get("sort") // TODO: implement sorting
	includeTimestamps := r.URL.Query().Get("include_timestamps") == "true"
	
	limit := 1000 // Default limit
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 10000 {
				limit = 10000 // Max limit
			}
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	// Use List() which should be cached
	entities, err := h.repo.List()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to list entities")
		return
	}
	
	// Apply pagination in memory (faster than repeated DB queries)
	totalCount := len(entities)
	
	// Apply offset and limit
	if offset >= totalCount {
		entities = []*models.Entity{}
	} else {
		end := offset + limit
		if end > totalCount {
			end = totalCount
		}
		entities = entities[offset:end]
	}
	
	// Strip timestamps if not requested (saves bandwidth)
	if !includeTimestamps {
		for i, entity := range entities {
			entities[i] = h.stripTimestamps(entity)
		}
	}
	
	// Build response
	response := map[string]interface{}{
		"entities": entities,
		"metadata": map[string]interface{}{
			"total":  totalCount,
			"offset": offset,
			"limit":  limit,
			"count":  len(entities),
		},
	}
	
	// Update response time header
	elapsed := time.Since(start)
	w.Header().Set("X-Response-Time", elapsed.String())
	
	// Log performance metric
	if elapsed > 100*time.Millisecond {
		logger.Warn("Slow list request: %v for %d entities", elapsed, len(entities))
	} else {
		logger.Debug("List request completed in %v", elapsed)
	}
	
	// Use json.Encoder for streaming response
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false) // Faster encoding
	
	if err := encoder.Encode(response); err != nil {
		logger.Error("Failed to encode response: %v", err)
	}
}

// QueryEntitiesOptimized handles entity queries with caching
func (h *OptimizedEntityHandler) QueryEntitiesOptimized(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Parse query parameters
	tags := r.URL.Query().Get("tags")
	matchAll := r.URL.Query().Get("match") == "all"
	limit := 1000
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	var entities []*models.Entity
	var err error
	
	if tags == "" {
		// No tags specified, use List (which is cached)
		entities, err = h.repo.List()
	} else {
		// Parse tags
		tagList := strings.Split(tags, ",")
		for i := range tagList {
			tagList[i] = strings.TrimSpace(tagList[i])
		}
		
		// Use ListByTags (which should also be cached)
		entities, err = h.repo.ListByTags(tagList, matchAll)
	}
	
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Query failed")
		return
	}
	
	// Apply limit
	if len(entities) > limit {
		entities = entities[:limit]
	}
	
	// Strip timestamps for performance
	for i, entity := range entities {
		entities[i] = h.stripTimestamps(entity)
	}
	
	response := map[string]interface{}{
		"entities": entities,
		"count":    len(entities),
		"query": map[string]interface{}{
			"tags":      tags,
			"match_all": matchAll,
			"limit":     limit,
		},
	}
	
	// Log performance
	elapsed := time.Since(start)
	logger.Debug("Query completed in %v, returned %d entities", elapsed, len(entities))
	
	RespondJSON(w, http.StatusOK, response)
}

// stripTimestamps removes temporal timestamps from tags for better performance
func (h *OptimizedEntityHandler) stripTimestamps(entity *models.Entity) *models.Entity {
	// Create a shallow copy
	result := &models.Entity{
		ID:        entity.ID,
		Content:   entity.Content, // Don't copy content, just reference
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		Tags:      make([]string, 0, len(entity.Tags)),
	}
	
	// Strip timestamps from tags
	for _, tag := range entity.Tags {
		if idx := strings.LastIndex(tag, "|"); idx != -1 {
			result.Tags = append(result.Tags, tag[idx+1:])
		} else {
			result.Tags = append(result.Tags, tag)
		}
	}
	
	return result
}