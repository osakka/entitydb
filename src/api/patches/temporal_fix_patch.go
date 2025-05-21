package api

import (
	"entitydb/logger"
	"fmt"
	"net/http"
	"time"
	"strings"
)

// GetEntityAsOfPatch is a patched implementation of GetEntityAsOf addressing timestamp parsing issues
func (h *EntityHandler) GetEntityAsOfPatch(w http.ResponseWriter, r *http.Request) {
	// Debug logs
	logger.Debug("GetEntityAsOfPatch called with params: %v", r.URL.Query())
	
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
	
	// PATCH: Convert timestamp to UTC to avoid timezone issues
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

// GetEntityHistoryPatch is a patched implementation of GetEntityHistory
func (h *EntityHandler) GetEntityHistoryPatch(w http.ResponseWriter, r *http.Request) {
	// Debug logs
	logger.Debug("GetEntityHistoryPatch called with params: %v", r.URL.Query())
	
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

// RegisterPatchedHandlers registers patched temporal handlers to the router
func RegisterPatchedHandlers(router *http.ServeMux, handler *EntityHandler) {
	router.HandleFunc("/api/v1/entities/as-of-patched", handler.GetEntityAsOfPatch)
	router.HandleFunc("/api/v1/entities/history-patched", handler.GetEntityHistoryPatch)
}