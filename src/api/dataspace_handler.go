package api

import (
	"encoding/json"
	"entitydb/models"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// DataspaceHandler handles dataspace management operations
type DataspaceHandler struct {
	repo models.EntityRepository
}

// NewDataspaceHandler creates a new handler for dataspace management
func NewDataspaceHandler(repo models.EntityRepository) *DataspaceHandler {
	return &DataspaceHandler{repo: repo}
}

// DataspaceRequest represents a request to create or update a dataspace
type DataspaceRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Settings    map[string]string `json:"settings"`
}

// DataspaceResponse represents a dataspace in API responses
type DataspaceResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Settings    map[string]string `json:"settings"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ListDataspaces returns all configured dataspaces
func (h *DataspaceHandler) ListDataspaces(w http.ResponseWriter, r *http.Request) {
	// Get all entities with type:dataspace tag
	entities, err := h.repo.ListByTags([]string{"type:dataspace"}, true)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to list dataspaces: " + err.Error())
		return
	}

	responses := make([]DataspaceResponse, 0, len(entities))
	for _, entity := range entities {
		resp := h.entityToDataspaceResponse(entity)
		responses = append(responses, resp)
	}

	RespondJSON(w, http.StatusOK, responses)
}

// GetDataspace returns a specific dataspace by ID
func (h *DataspaceHandler) GetDataspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataspaceID := vars["id"]

	entity, err := h.repo.GetByID(dataspaceID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Dataspace not found")
		return
	}

	// Verify it's a dataspace entity
	if !h.hasTag(entity.Tags, "type:dataspace") {
		RespondError(w, http.StatusNotFound, "Entity is not a dataspace")
		return
	}

	resp := h.entityToDataspaceResponse(entity)
	RespondJSON(w, http.StatusOK, resp)
}

// CreateDataspace creates a new dataspace
func (h *DataspaceHandler) CreateDataspace(w http.ResponseWriter, r *http.Request) {
	var req DataspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate name
	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, "Dataspace name is required")
		return
	}

	// Check if dataspace already exists
	existing, err := h.repo.ListByTags([]string{"type:dataspace", "dataspace:" + req.Name}, true)
	if err == nil && len(existing) > 0 {
		RespondError(w, http.StatusConflict, "Dataspace already exists")
		return
	}

	// Create dataspace entity
	entity := &models.Entity{
		ID:   uuid.New().String(),
		Tags: []string{
			"type:dataspace",
			"dataspace:" + req.Name,
			"id:" + req.Name,
		},
		Content: h.marshalDataspaceContent(req),
	}

	if err := h.repo.Create(entity); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create dataspace")
		return
	}

	resp := h.entityToDataspaceResponse(entity)
	RespondJSON(w, http.StatusCreated, resp)
}

// UpdateDataspace updates an existing dataspace
func (h *DataspaceHandler) UpdateDataspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataspaceID := vars["id"]

	var req DataspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing dataspace
	entity, err := h.repo.GetByID(dataspaceID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Dataspace not found")
		return
	}

	// Verify it's a dataspace entity
	if !h.hasTag(entity.Tags, "type:dataspace") {
		RespondError(w, http.StatusNotFound, "Entity is not a dataspace")
		return
	}

	// Update content
	entity.Content = h.marshalDataspaceContent(req)

	if err := h.repo.Update(entity); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update dataspace")
		return
	}

	resp := h.entityToDataspaceResponse(entity)
	RespondJSON(w, http.StatusOK, resp)
}

// DeleteDataspace deletes a dataspace
func (h *DataspaceHandler) DeleteDataspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dataspaceID := vars["id"]

	// Get existing dataspace
	entity, err := h.repo.GetByID(dataspaceID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Dataspace not found")
		return
	}

	// Verify it's a dataspace entity
	if !h.hasTag(entity.Tags, "type:dataspace") {
		RespondError(w, http.StatusNotFound, "Entity is not a dataspace")
		return
	}

	// Extract dataspace name
	var dataspaceName string
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "dataspace:") {
			dataspaceName = strings.TrimPrefix(tag, "dataspace:")
			break
		}
	}

	if dataspaceName != "" {
		// Check if there are any entities in this dataspace
		entities, err := h.repo.ListByTags([]string{"dataspace:" + dataspaceName}, true)
		if err == nil && len(entities) > 1 { // > 1 because the dataspace entity itself has this tag
			RespondError(w, http.StatusConflict, "Cannot delete dataspace with existing entities")
			return
		}
	}

	if err := h.repo.Delete(dataspaceID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to delete dataspace")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func (h *DataspaceHandler) hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (h *DataspaceHandler) marshalDataspaceContent(req DataspaceRequest) []byte {
	content := map[string]interface{}{
		"description": req.Description,
		"settings":    req.Settings,
	}
	data, _ := json.Marshal(content)
	return data
}

func (h *DataspaceHandler) entityToDataspaceResponse(entity *models.Entity) DataspaceResponse {
	// Parse timestamps
	createdAt := time.Unix(0, entity.CreatedAt)
	updatedAt := time.Unix(0, entity.UpdatedAt)
	
	resp := DataspaceResponse{
		ID:        entity.ID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	// Extract dataspace name from tags
	for _, tag := range entity.Tags {
		// Handle temporal tags (TIMESTAMP|tag format)
		actualTag := tag
		if idx := strings.LastIndex(tag, "|"); idx != -1 {
			actualTag = tag[idx+1:]
		}
		
		if strings.HasPrefix(actualTag, "dataspace:") {
			resp.Name = strings.TrimPrefix(actualTag, "dataspace:")
			break
		}
	}

	// Parse content
	if len(entity.Content) > 0 {
		var content map[string]interface{}
		if err := json.Unmarshal(entity.Content, &content); err == nil {
			if desc, ok := content["description"].(string); ok {
				resp.Description = desc
			}
			if settings, ok := content["settings"].(map[string]interface{}); ok {
				resp.Settings = make(map[string]string)
				for k, v := range settings {
					if str, ok := v.(string); ok {
						resp.Settings[k] = str
					}
				}
			}
		}
	}

	return resp
}