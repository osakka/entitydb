package api

import (
	"net/http"
	"strings"
	"entitydb/logger"
)

// ListEntitiesDataspaceAware handles listing entities with dataspace support
func (h *EntityHandler) ListEntitiesDataspaceAware(w http.ResponseWriter, r *http.Request) {
	// Check if timestamps should be included in response
	includeTimestamps := r.URL.Query().Get("include_timestamps") == "true"
	
	// Get query parameters
	tagsParam := r.URL.Query().Get("tags")
	matchAll := r.URL.Query().Get("matchAll") == "true"
	
	// If no tags specified, fall back to original handler
	if tagsParam == "" {
		h.ListEntities(w, r)
		return
	}
	
	// Parse tags (comma-separated)
	tags := strings.Split(tagsParam, ",")
	for i := range tags {
		tags[i] = strings.TrimSpace(tags[i])
	}
	
	// Use ListByTags which supports dataspace queries
	entities, err := h.repo.ListByTags(tags, matchAll)
	if err != nil {
		logger.Error("Failed to list entities by tags: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to list entities")
		return
	}
	
	// Convert to response format
	response := make([]map[string]interface{}, len(entities))
	for i, entity := range entities {
		if includeTimestamps {
			response[i] = map[string]interface{}{
				"id":         entity.ID,
				"tags":       entity.Tags,
				"content":    entity.Content,
				"created_at": entity.CreatedAt,
				"updated_at": entity.UpdatedAt,
			}
		} else {
			// Filter out temporal tags
			filteredTags := make([]string, 0, len(entity.Tags))
			for _, tag := range entity.Tags {
				// Remove timestamp prefix if present
				if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
					filteredTags = append(filteredTags, parts[1])
				} else {
					filteredTags = append(filteredTags, tag)
				}
			}
			response[i] = map[string]interface{}{
				"id":         entity.ID,
				"tags":       filteredTags,
				"content":    entity.Content,
				"created_at": entity.CreatedAt,
				"updated_at": entity.UpdatedAt,
			}
		}
	}
	
	RespondJSON(w, http.StatusOK, response)
}