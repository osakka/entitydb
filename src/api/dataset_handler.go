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

// DatasetHandler handles dataset management operations
type DatasetHandler struct {
	repo models.EntityRepository
}

// NewDatasetHandler creates a new handler for dataset management
func NewDatasetHandler(repo models.EntityRepository) *DatasetHandler {
	return &DatasetHandler{repo: repo}
}

// DatasetRequest represents a request to create or update a dataset
type DatasetRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Settings    map[string]string `json:"settings"`
}

// DatasetResponse represents a dataset in API responses
type DatasetResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Settings    map[string]string `json:"settings"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// ListDatasets returns all configured datasets
func (h *DatasetHandler) ListDatasets(w http.ResponseWriter, r *http.Request) {
	// Get all entities with type:dataset tag
	entities, err := h.repo.ListByTags([]string{"type:dataset"}, true)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to list datasets: " + err.Error())
		return
	}

	responses := make([]DatasetResponse, 0, len(entities))
	for _, entity := range entities {
		resp := h.entityToDatasetResponse(entity)
		responses = append(responses, resp)
	}

	RespondJSON(w, http.StatusOK, responses)
}

// GetDataset returns a specific dataset by ID
func (h *DatasetHandler) GetDataset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datasetID := vars["id"]

	entity, err := h.repo.GetByID(datasetID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Dataset not found")
		return
	}

	// Verify it's a dataset entity
	if !h.hasTag(entity.Tags, "type:dataset") {
		RespondError(w, http.StatusNotFound, "Entity is not a dataset")
		return
	}

	resp := h.entityToDatasetResponse(entity)
	RespondJSON(w, http.StatusOK, resp)
}

// CreateDataset creates a new dataset
func (h *DatasetHandler) CreateDataset(w http.ResponseWriter, r *http.Request) {
	var req DatasetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate name
	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, "Dataset name is required")
		return
	}

	// Check if dataset already exists
	existing, err := h.repo.ListByTags([]string{"type:dataset", "dataset:" + req.Name}, true)
	if err == nil && len(existing) > 0 {
		RespondError(w, http.StatusConflict, "Dataset already exists")
		return
	}

	// Create dataset entity (datasets themselves belong to _system dataset)
	entity := &models.Entity{
		ID:   uuid.New().String(),
		Tags: []string{
			"type:dataset",
			"dataset:_system", // Dataset entities belong to the system dataset
			"name:" + req.Name,
			"id:" + req.Name,
		},
		Content: h.marshalDatasetContent(req),
	}

	if err := h.repo.Create(entity); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create dataset")
		return
	}

	resp := h.entityToDatasetResponse(entity)
	RespondJSON(w, http.StatusCreated, resp)
}

// UpdateDataset updates an existing dataset
func (h *DatasetHandler) UpdateDataset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datasetID := vars["id"]

	var req DatasetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing dataset
	entity, err := h.repo.GetByID(datasetID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Dataset not found")
		return
	}

	// Verify it's a dataset entity
	if !h.hasTag(entity.Tags, "type:dataset") {
		RespondError(w, http.StatusNotFound, "Entity is not a dataset")
		return
	}

	// Update content
	entity.Content = h.marshalDatasetContent(req)

	if err := h.repo.Update(entity); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update dataset")
		return
	}

	resp := h.entityToDatasetResponse(entity)
	RespondJSON(w, http.StatusOK, resp)
}

// DeleteDataset deletes a dataset
func (h *DatasetHandler) DeleteDataset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datasetID := vars["id"]

	// Get existing dataset
	entity, err := h.repo.GetByID(datasetID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Dataset not found")
		return
	}

	// Verify it's a dataset entity
	if !h.hasTag(entity.Tags, "type:dataset") {
		RespondError(w, http.StatusNotFound, "Entity is not a dataset")
		return
	}

	// Extract dataset name
	var datasetName string
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "dataset:") {
			datasetName = strings.TrimPrefix(tag, "dataset:")
			break
		}
	}

	if datasetName != "" {
		// Check if there are any entities in this dataset
		entities, err := h.repo.ListByTags([]string{"dataset:" + datasetName}, true)
		if err == nil && len(entities) > 1 { // > 1 because the dataset entity itself has this tag
			RespondError(w, http.StatusConflict, "Cannot delete dataset with existing entities")
			return
		}
	}

	if err := h.repo.Delete(datasetID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to delete dataset")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func (h *DatasetHandler) hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (h *DatasetHandler) marshalDatasetContent(req DatasetRequest) []byte {
	content := map[string]interface{}{
		"description": req.Description,
		"settings":    req.Settings,
	}
	data, _ := json.Marshal(content)
	return data
}

func (h *DatasetHandler) entityToDatasetResponse(entity *models.Entity) DatasetResponse {
	// Parse timestamps
	createdAt := time.Unix(0, entity.CreatedAt)
	updatedAt := time.Unix(0, entity.UpdatedAt)
	
	resp := DatasetResponse{
		ID:        entity.ID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	// Extract dataset name and description from tags
	for _, tag := range entity.Tags {
		// Handle temporal tags (TIMESTAMP|tag format)
		actualTag := tag
		if idx := strings.LastIndex(tag, "|"); idx != -1 {
			actualTag = tag[idx+1:]
		}
		
		if strings.HasPrefix(actualTag, "name:") {
			resp.Name = strings.TrimPrefix(actualTag, "name:")
		} else if strings.HasPrefix(actualTag, "description:") {
			resp.Description = strings.TrimPrefix(actualTag, "description:")
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