package api

import (
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"net/http"
	"time"
	"strings"
)

// Utility: Assert that a repository is a TemporalRepository
func asTemporalRepository(repo models.EntityRepository) (*binary.TemporalRepository, error) {
	if temporalRepo, ok := repo.(*binary.TemporalRepository); ok {
		return temporalRepo, nil
	}
	return nil, fmt.Errorf("repository does not support temporal features")
}

// GetEntityAsOfFixed is an improved implementation of GetEntityAsOf
func (h *EntityHandler) GetEntityAsOfFixed(w http.ResponseWriter, r *http.Request) {
	// Debug logs
	logger.Debug("GetEntityAsOfFixed called with params: %v", r.URL.Query())
	
	// Check if timestamps should be included in response
	includeTimestamps := r.URL.Query().Get("include_timestamps") == "true"
	
	// Get entity ID from query
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		logger.Error("Entity ID is missing in request")
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}
	
	// Get timestamp from query - handle different parameter names
	asOfStr := r.URL.Query().Get("as_of")
	if asOfStr == "" {
		asOfStr = r.URL.Query().Get("timestamp")
	}
	if asOfStr == "" {
		logger.Error("Timestamp is missing in request")
		RespondError(w, http.StatusBadRequest, "Timestamp is required")
		return
	}
	
	logger.Debug("Using timestamp: %s", asOfStr)
	
	// Parse timestamp with flexible format handling
	var asOf time.Time
	var err error
	
	// Try multiple timestamp formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	}
	
	for _, format := range formats {
		asOf, err = time.Parse(format, asOfStr)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		logger.Error("Failed to parse timestamp %s: %v", asOfStr, err)
		RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid timestamp format. Try format like '2025-05-21T08:45:20Z'. Error: %v", err))
		return
	}
	
	logger.Debug("Parsed timestamp: %v", asOf)
	
	// Get entity repository
	temporalRepo, err := asTemporalRepository(h.repo)
	if err != nil {
		logger.Error("Repository doesn't support temporal features: %v", err)
		RespondError(w, http.StatusInternalServerError, "Temporal features not available")
		return
	}
	
	// Convert timestamp to UTC to avoid timezone issues
	asOf = asOf.UTC()
	logger.Debug("Using UTC timestamp: %v", asOf)
	
	// Get entity as of timestamp with better error reporting
	entity, err := temporalRepo.GetEntityAsOfFixed(entityID, asOf)
	if err != nil {
		logger.Error("Failed to get entity %s as of %v: %v", entityID, asOf, err)
		
		if strings.Contains(err.Error(), "entity not found") {
			RespondError(w, http.StatusNotFound, fmt.Sprintf("Entity %s not found at timestamp %v", entityID, asOf))
		} else if strings.Contains(err.Error(), "did not exist at") {
			RespondError(w, http.StatusNotFound, fmt.Sprintf("Entity %s did not exist at timestamp %v", entityID, asOf))
		} else {
			RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get historical entity: %v", err))
		}
		return
	}
	
	// Return entity with timestamps stripped unless requested
	response := h.stripTimestampsFromEntity(entity, includeTimestamps)
	logger.Debug("Returning entity as of %v: %+v", asOf, response)
	RespondJSON(w, http.StatusOK, response)
}

// GetEntityHistoryFixed is an improved implementation of GetEntityHistory
func (h *EntityHandler) GetEntityHistoryFixed(w http.ResponseWriter, r *http.Request) {
	// Debug logs
	logger.Debug("GetEntityHistoryFixed called with params: %v", r.URL.Query())
	
	// Get entity ID from query
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		logger.Error("Entity ID is missing in request")
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}
	
	// Get optional limit
	limit := 100 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := parseInt(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	logger.Debug("Getting history for entity %s with limit %d", entityID, limit)
	
	// Get entity repository
	temporalRepo, err := asTemporalRepository(h.repo)
	if err != nil {
		logger.Error("Repository doesn't support temporal features: %v", err)
		RespondError(w, http.StatusInternalServerError, "Temporal features not available")
		return
	}
	
	// Check if entity exists first
	_, err = temporalRepo.GetByID(entityID)
	if err != nil {
		logger.Error("Entity %s not found: %v", entityID, err)
		RespondError(w, http.StatusNotFound, fmt.Sprintf("Entity %s not found", entityID))
		return
	}
	
	// Get entity history
	history, err := temporalRepo.GetEntityHistoryFixed(entityID, limit)
	if err != nil {
		logger.Error("Failed to get entity history: %v", err)
		RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get entity history: %v", err))
		return
	}
	
	logger.Debug("Found %d history entries for entity %s", len(history), entityID)
	RespondJSON(w, http.StatusOK, history)
}

// GetRecentChangesFixed is an improved implementation of GetRecentChanges
func (h *EntityHandler) GetRecentChangesFixed(w http.ResponseWriter, r *http.Request) {
	// Debug logs
	logger.Debug("GetRecentChangesFixed called with params: %v", r.URL.Query())
	
	// Get optional limit
	limit := 100 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := parseInt(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	// Get entity ID if specified (for entity-specific changes)
	entityID := r.URL.Query().Get("id")
	logger.Debug("Getting recent changes with limit %d, entity ID: %s", limit, entityID)
	
	// Get recent changes
	temporalRepo, err := asTemporalRepository(h.repo)
	if err != nil {
		logger.Error("Repository doesn't support temporal features: %v", err)
		RespondError(w, http.StatusInternalServerError, "Temporal features not available")
		return
	}
	
	var changes []*models.EntityChange
	
	if entityID != "" {
		// Check if entity exists first
		_, err = temporalRepo.GetByID(entityID)
		if err != nil {
			logger.Error("Entity %s not found: %v", entityID, err)
			RespondError(w, http.StatusNotFound, fmt.Sprintf("Entity %s not found", entityID))
			return
		}
		
		// Get changes for specific entity
		changes, err = temporalRepo.GetEntityHistoryFixed(entityID, limit)
	} else {
		// Get global changes
		changes, err = temporalRepo.GetRecentChangesFixed(limit)
	}
	
	if err != nil {
		logger.Error("Failed to get recent changes: %v", err)
		RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get recent changes: %v", err))
		return
	}
	
	logger.Debug("Found %d change entries", len(changes))
	RespondJSON(w, http.StatusOK, changes)
}

// GetEntityDiffFixed is an improved implementation of GetEntityDiff
func (h *EntityHandler) GetEntityDiffFixed(w http.ResponseWriter, r *http.Request) {
	// Debug logs
	logger.Debug("GetEntityDiffFixed called with params: %v", r.URL.Query())
	
	// Check if timestamps should be included in response
	includeTimestamps := r.URL.Query().Get("include_timestamps") == "true"
	
	// Get entity ID from query
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		logger.Error("Entity ID is missing in request")
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}
	
	// Get timestamps from query (support multiple parameter names)
	t1Str := r.URL.Query().Get("from_timestamp")
	if t1Str == "" {
		t1Str = r.URL.Query().Get("t1")
	}
	if t1Str == "" {
		t1Str = r.URL.Query().Get("from")
	}
	
	t2Str := r.URL.Query().Get("to_timestamp")
	if t2Str == "" {
		t2Str = r.URL.Query().Get("t2")
	}
	if t2Str == "" {
		t2Str = r.URL.Query().Get("to")
	}
	
	if t1Str == "" || t2Str == "" {
		logger.Error("Missing from or to timestamp in request")
		RespondError(w, http.StatusBadRequest, "Both from and to timestamps are required")
		return
	}
	
	logger.Debug("Using timestamps: from=%s, to=%s", t1Str, t2Str)
	
	// Parse timestamps with multiple format support
	var t1, t2 time.Time
	var err error
	
	// Try multiple timestamp formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	}
	
	// Parse first timestamp
	for _, format := range formats {
		t1, err = time.Parse(format, t1Str)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		logger.Error("Failed to parse from timestamp %s: %v", t1Str, err)
		RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid from timestamp format. Try format like '2025-05-21T08:45:20Z'. Error: %v", err))
		return
	}
	
	// Parse second timestamp
	for _, format := range formats {
		t2, err = time.Parse(format, t2Str)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		logger.Error("Failed to parse to timestamp %s: %v", t2Str, err)
		RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid to timestamp format. Try format like '2025-05-21T08:45:20Z'. Error: %v", err))
		return
	}
	
	// Convert to UTC for consistency
	t1 = t1.UTC()
	t2 = t2.UTC()
	logger.Debug("Parsed and converted timestamps: from=%v, to=%v", t1, t2)
	
	// Get entity repository
	temporalRepo, err := asTemporalRepository(h.repo)
	if err != nil {
		logger.Error("Repository doesn't support temporal features: %v", err)
		RespondError(w, http.StatusInternalServerError, "Temporal features not available")
		return
	}
	
	// Check if entity exists first
	_, err = temporalRepo.GetByID(entityID)
	if err != nil {
		logger.Error("Entity %s not found: %v", entityID, err)
		RespondError(w, http.StatusNotFound, fmt.Sprintf("Entity %s not found", entityID))
		return
	}
	
	beforeEntity, afterEntity, err := temporalRepo.GetEntityDiffFixed(entityID, t1, t2)
	if err != nil {
		logger.Error("Failed to get entity diff: %v", err)
		RespondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get entity diff: %v", err))
		return
	}
	
	// Strip timestamps if not requested
	if !includeTimestamps {
		if beforeEntity != nil {
			beforeEntity = h.stripTimestampsFromEntity(beforeEntity, false)
		}
		if afterEntity != nil {
			afterEntity = h.stripTimestampsFromEntity(afterEntity, false)
		}
	}
	
	// Construct the diff response
	diff := map[string]interface{}{
		"entity_id": entityID,
		"from_time": t1.Format(time.RFC3339),
		"to_time":   t2.Format(time.RFC3339),
		"before":    beforeEntity,
		"after":     afterEntity,
	}
	
	// Add a helpful summary of changes
	if beforeEntity != nil && afterEntity != nil {
		// Build a summary of changes
		var addedTags, removedTags []string
		
		// Get simple tags (without timestamps)
		beforeSimpleTags := beforeEntity.GetTagsWithoutTimestamp()
		afterSimpleTags := afterEntity.GetTagsWithoutTimestamp()
		
		// Find added tags
		for _, tag := range afterSimpleTags {
			found := false
			for _, beforeTag := range beforeSimpleTags {
				if tag == beforeTag {
					found = true
					break
				}
			}
			if !found {
				addedTags = append(addedTags, tag)
			}
		}
		
		// Find removed tags
		for _, tag := range beforeSimpleTags {
			found := false
			for _, afterTag := range afterSimpleTags {
				if tag == afterTag {
					found = true
					break
				}
			}
			if !found {
				removedTags = append(removedTags, tag)
			}
		}
		
		diff["added_tags"] = addedTags
		diff["removed_tags"] = removedTags
	}
	
	logger.Debug("Returning diff result with %d added tags and %d removed tags", 
		len(diff["added_tags"].([]string)), len(diff["removed_tags"].([]string)))
	RespondJSON(w, http.StatusOK, diff)
}

// Helper to parse integer safely
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}