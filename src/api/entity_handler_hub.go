package api

import (
	"bytes"
	"encoding/json"
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"net/http"
	"time"
)

// HubEntityRequest represents a hub-aware entity creation request
type HubEntityRequest struct {
	Hub     string                 `json:"hub"`                // Required hub name
	Self    map[string]string      `json:"self,omitempty"`     // Self properties: {"type": "task", "assignee": "john"}
	Traits  map[string]string      `json:"traits,omitempty"`   // Trait properties: {"org": "TechCorp", "project": "Mobile"}
	Tags    []string               `json:"tags,omitempty"`     // Additional raw tags (optional)
	Content interface{}            `json:"content,omitempty"`  // Entity content
	ID      string                 `json:"id,omitempty"`       // Optional ID
}

// HubEntityResponse represents a hub-aware entity response
type HubEntityResponse struct {
	ID        string                 `json:"id"`
	Hub       string                 `json:"hub"`
	Self      map[string]string      `json:"self,omitempty"`
	Traits    map[string]string      `json:"traits,omitempty"`
	Tags      []string               `json:"tags,omitempty"`     // Raw tags (if requested)
	Content   interface{}            `json:"content,omitempty"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// CreateHubEntity handles creating entities with hub/trait/self structure
func (h *EntityHandler) CreateHubEntity(w http.ResponseWriter, r *http.Request) {
	// Get RBAC context
	rbacCtx, hasRBAC := GetRBACContext(r)
	if !hasRBAC {
		RespondError(w, http.StatusUnauthorized, "RBAC context required")
		return
	}

	// Parse request
	var req HubEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate hub is provided
	if req.Hub == "" {
		RespondError(w, http.StatusBadRequest, "hub is required")
		return
	}

	// Check hub access permission
	if !CheckHubPermission(rbacCtx, req.Hub, "create") {
		RespondError(w, http.StatusForbidden, fmt.Sprintf("No create permission for hub: %s", req.Hub))
		return
	}

	// Create entity
	entity := &models.Entity{
		ID:        req.ID,
		Tags:      []string{},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Add hub tag
	entity.AddTag(FormatHubTag(req.Hub))

	// Add self tags
	for namespace, value := range req.Self {
		entity.AddTag(FormatSelfTag(req.Hub, namespace, value))
	}

	// Add trait tags
	for namespace, value := range req.Traits {
		entity.AddTag(FormatTraitTag(req.Hub, namespace, value))
	}

	// Add any additional raw tags
	for _, tag := range req.Tags {
		entity.AddTag(tag)
	}

	// Handle content
	if req.Content != nil {
		contentBytes, contentType, err := h.processContent(req.Content)
		if err != nil {
			RespondError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Set content
		if len(contentBytes) > 0 {
			reader := bytes.NewReader(contentBytes)
			config := models.DefaultChunkConfig()
			_, err := entity.SetContent(reader, contentType, config)
			if err != nil {
				RespondError(w, http.StatusInternalServerError, "Failed to set content")
				return
			}
		}
	}

	// Save entity
	err := h.repo.Create(entity)
	if err != nil {
		logger.Error("Failed to create hub entity: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to create entity")
		return
	}

	// Log creation
	logger.Info("Hub entity created: hub=%s, id=%s", req.Hub, entity.ID)

	// Return response
	response := h.entityToHubResponse(entity, r)
	RespondJSON(w, http.StatusCreated, response)
}

// QueryHubEntities handles querying entities with hub filtering
func (h *EntityHandler) QueryHubEntities(w http.ResponseWriter, r *http.Request) {
	// Get RBAC context
	rbacCtx, hasRBAC := GetRBACContext(r)
	if !hasRBAC {
		RespondError(w, http.StatusUnauthorized, "RBAC context required")
		return
	}

	// Get query parameters
	hubName := r.URL.Query().Get("hub")
	selfFilter := r.URL.Query().Get("self")     // Format: "type:task,assignee:john"
	traitFilter := r.URL.Query().Get("traits")  // Format: "org:TechCorp,project:Mobile"
	includeContent := r.URL.Query().Get("include_content") == "true"
	includeRawTags := r.URL.Query().Get("include_raw_tags") == "true"

	// Build query tags
	var queryTags []string

	// Add hub filter
	if hubName != "" {
		// Check hub access permission
		if !CheckHubPermission(rbacCtx, hubName, "view") {
			RespondError(w, http.StatusForbidden, fmt.Sprintf("No view permission for hub: %s", hubName))
			return
		}
		queryTags = append(queryTags, FormatHubTag(hubName))
	} else {
		// If no specific hub, filter by user's accessible hubs
		if !rbacCtx.IsAdmin {
			userHubs := getUserHubs(rbacCtx.Permissions)
			if len(userHubs) == 0 {
				// User has no hub access
				RespondJSON(w, http.StatusOK, map[string]interface{}{
					"entities": []HubEntityResponse{},
					"total":    0,
				})
				return
			}
			// For now, require specific hub to be specified
			RespondError(w, http.StatusBadRequest, "hub parameter required")
			return
		}
	}

	// Add self filters
	if selfFilter != "" {
		selfPairs := parseFilterPairs(selfFilter)
		for namespace, value := range selfPairs {
			queryTags = append(queryTags, FormatSelfTag(hubName, namespace, value))
		}
	}

	// Add trait filters
	if traitFilter != "" {
		traitPairs := parseFilterPairs(traitFilter)
		for namespace, value := range traitPairs {
			queryTags = append(queryTags, FormatTraitTag(hubName, namespace, value))
		}
	}

	// Query entities
	logger.Debug("QueryHubEntities: Querying with tags: %v", queryTags)
	entities, err := h.repo.ListByTags(queryTags, true) // matchAll = true
	if err != nil {
		logger.Error("Failed to query hub entities: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to query entities")
		return
	}
	logger.Debug("QueryHubEntities: Found %d entities", len(entities))

	// Filter entities user has access to
	var accessibleEntities []*models.Entity
	for _, entity := range entities {
		if ValidateEntityHub(rbacCtx, entity) == nil {
			accessibleEntities = append(accessibleEntities, entity)
		}
	}

	// Convert to hub responses
	var responses []HubEntityResponse
	for _, entity := range accessibleEntities {
		response := h.entityToHubResponse(entity, r)
		
		// Include content if requested
		if includeContent && len(entity.Content) > 0 {
			response.Content = entity.Content
		}
		
		// Include raw tags if requested
		if includeRawTags {
			response.Tags = entity.GetTagsWithoutTimestamp()
		}
		
		responses = append(responses, response)
	}

	// Return results
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"entities": responses,
		"total":    len(responses),
		"hub":      hubName,
	})
}

// Helper functions

// processContent handles different content types
func (h *EntityHandler) processContent(content interface{}) ([]byte, string, error) {
	switch c := content.(type) {
	case string:
		return []byte(c), "text/plain", nil
	case map[string]interface{}, []interface{}:
		jsonBytes, err := json.Marshal(c)
		if err != nil {
			return nil, "", fmt.Errorf("invalid JSON content: %v", err)
		}
		return jsonBytes, "application/json", nil
	default:
		return nil, "", fmt.Errorf("unsupported content type")
	}
}

// entityToHubResponse converts entity to hub-aware response
func (h *EntityHandler) entityToHubResponse(entity *models.Entity, r *http.Request) HubEntityResponse {
	response := HubEntityResponse{
		ID:        entity.ID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		Self:      make(map[string]string),
		Traits:    make(map[string]string),
	}

	// Parse tags into hub/self/traits
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if hubName, ok := ParseHubTag(tag); ok {
			response.Hub = hubName
		} else if hub, namespace, value, ok := ParseSelfTag(tag); ok {
			if response.Hub == "" || response.Hub == hub {
				response.Self[namespace] = value
			}
		} else if hub, namespace, value, ok := ParseTraitTag(tag); ok {
			if response.Hub == "" || response.Hub == hub {
				response.Traits[namespace] = value
			}
		}
	}

	return response
}

// parseFilterPairs parses "key1:val1,key2:val2" format
func parseFilterPairs(filter string) map[string]string {
	pairs := make(map[string]string)
	if filter == "" {
		return pairs
	}

	for _, pair := range splitAndTrim(filter, ",") {
		parts := splitAndTrim(pair, ":")
		if len(parts) == 2 {
			pairs[parts[0]] = parts[1]
		}
	}
	return pairs
}

// splitAndTrim splits string and trims whitespace
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimWhitespace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// Helper functions for string operations (avoiding imports)
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	// Simple split implementation
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimWhitespace(s string) string {
	start := 0
	end := len(s)
	
	// Trim leading whitespace
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	
	// Trim trailing whitespace
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	
	return s[start:end]
}