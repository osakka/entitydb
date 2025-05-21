package api

import (
	"entitydb/logger"
	"entitydb/storage/binary"
	"net/http"
	"time"
	"strings"
)

// Simple handler for testing fixed temporal features
func (h *EntityHandler) TestTemporalFixHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the requested operation from the URL path
	path := r.URL.Path
	operation := ""
	if parts := strings.Split(path, "/"); len(parts) > 0 {
		operation = parts[len(parts)-1]
	}
	
	logger.Debug("Test temporal fix handler called for operation: %s", operation)
	
	// Get repository as temporal repository
	temporalRepo, ok := h.repo.(*binary.TemporalRepository)
	if !ok {
		RespondError(w, http.StatusInternalServerError, "Repository does not support temporal features")
		return
	}
	
	switch operation {
	case "as-of-test":
		testAsOf(w, r, temporalRepo)
	case "history-test":
		testHistory(w, r, temporalRepo)
	case "changes-test":
		testChanges(w, r, temporalRepo)
	case "diff-test":
		testDiff(w, r, temporalRepo)
	default:
		RespondError(w, http.StatusBadRequest, "Unknown test operation")
	}
}

// Test the as-of functionality
func testAsOf(w http.ResponseWriter, r *http.Request, repo *binary.TemporalRepository) {
	// Get parameters
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}
	
	timestampStr := r.URL.Query().Get("timestamp")
	if timestampStr == "" {
		RespondError(w, http.StatusBadRequest, "Timestamp is required")
		return
	}
	
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid timestamp format")
		return
	}
	
	// Get entity as of timestamp using fixed implementation
	entity, err := repo.GetEntityAsOfFixed(entityID, timestamp)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get entity as of timestamp: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, entity)
}

// Test the history functionality
func testHistory(w http.ResponseWriter, r *http.Request, repo *binary.TemporalRepository) {
	// Get parameters
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}
	
	// Use fixed implementation
	history, err := repo.GetEntityHistoryFixed(entityID, 100)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get entity history: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, history)
}

// Test the changes functionality
func testChanges(w http.ResponseWriter, r *http.Request, repo *binary.TemporalRepository) {
	// Use fixed implementation
	changes, err := repo.GetRecentChangesFixed(100)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get recent changes: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, changes)
}

// Test the diff functionality
func testDiff(w http.ResponseWriter, r *http.Request, repo *binary.TemporalRepository) {
	// Get parameters
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}
	
	t1Str := r.URL.Query().Get("t1")
	t2Str := r.URL.Query().Get("t2")
	if t1Str == "" || t2Str == "" {
		RespondError(w, http.StatusBadRequest, "Both t1 and t2 timestamps are required")
		return
	}
	
	t1, err := time.Parse(time.RFC3339, t1Str)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid t1 timestamp format")
		return
	}
	
	t2, err := time.Parse(time.RFC3339, t2Str)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid t2 timestamp format")
		return
	}
	
	// Use fixed implementation
	before, after, err := repo.GetEntityDiffFixed(entityID, t1, t2)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get entity diff: "+err.Error())
		return
	}
	
	// Construct useful response
	response := map[string]interface{}{
		"before":       before,
		"after":        after,
		"from":         t1.Format(time.RFC3339),
		"to":           t2.Format(time.RFC3339),
		"entity_id":    entityID,
	}
	
	// Add helpful diff information
	if before != nil && after != nil {
		beforeTags := before.GetTagsWithoutTimestamp()
		afterTags := after.GetTagsWithoutTimestamp()
		
		// Find added tags
		addedTags := []string{}
		for _, tag := range afterTags {
			found := false
			for _, beforeTag := range beforeTags {
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
		removedTags := []string{}
		for _, tag := range beforeTags {
			found := false
			for _, afterTag := range afterTags {
				if tag == afterTag {
					found = true
					break
				}
			}
			if !found {
				removedTags = append(removedTags, tag)
			}
		}
		
		response["added_tags"] = addedTags
		response["removed_tags"] = removedTags
	}
	
	RespondJSON(w, http.StatusOK, response)
}
