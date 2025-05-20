package api

import (
	"entitydb/models"
	"encoding/json"
	"net/http"
	"time"
)

// EntityConfigHandler handles configuration and feature flag requests via entities
type EntityConfigHandler struct {
	entityRepo *models.RepositoryQueryWrapper
}

// NewEntityConfigHandler creates a new entity-based config handler
func NewEntityConfigHandler(entityRepo models.EntityRepository) *EntityConfigHandler {
	return &EntityConfigHandler{
		entityRepo: models.NewRepositoryQueryWrapper(entityRepo),
	}
}

// GetConfig retrieves configuration entries
// @Summary Get configuration
// @Description Retrieve current system configuration
// @Tags configuration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param namespace query string false "Configuration namespace"
// @Param key query string false "Configuration key"
// @Success 200 {array} models.Entity
// @Router /api/v1/config [get]
func (h *EntityConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	key := r.URL.Query().Get("key")
	
	query := h.entityRepo.Query().HasTag("type:config")
	
	if namespace != "" {
		query = query.HasWildcardTag("conf:" + namespace + ":*")
	}
	
	if key != "" {
		query = query.HasTag("conf:" + namespace + ":" + key)
	}
	
	entities, err := query.Execute()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve config")
		return
	}
	
	RespondJSON(w, http.StatusOK, entities)
}

// SetConfig creates or updates a configuration entry
// @Summary Set configuration
// @Description Update system configuration values
// @Tags configuration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ConfigSetRequest true "Configuration data"
// @Success 200 {object} SuccessResponse
// @Router /api/v1/config/set [post]
func (h *EntityConfigHandler) SetConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key       string      `json:"key"`
		Value     interface{} `json:"value"`
		Namespace string      `json:"namespace"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if req.Key == "" {
		RespondError(w, http.StatusBadRequest, "Key is required")
		return
	}
	
	if req.Namespace == "" {
		req.Namespace = "system"
	}
	
	// Create config entity
	entityID := "config_" + req.Namespace + "_" + req.Key
	
	tags := []string{
		"type:config",
		"conf:" + req.Namespace + ":" + req.Key,
	}
	
	valueBytes, _ := json.Marshal(req.Value)
	
	entity := &models.Entity{
		ID:        entityID,
		Tags:      append(tags, 
			"content:type:json",
			"key:" + req.Key,
			"namespace:" + req.Namespace,
			"updated_at:" + time.Now().UTC().Format(time.RFC3339),
		),
		Content:   valueBytes, // Store value directly as bytes
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	
	// Check if exists
	existing, err := h.entityRepo.GetByID(entityID)
	if err == nil && existing != nil {
		// Update existing
		entity.CreatedAt = existing.CreatedAt
		if err := h.entityRepo.Update(entity); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to update config")
			return
		}
	} else {
		// Create new
		if err := h.entityRepo.Create(entity); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to create config")
			return
		}
	}
	
	RespondJSON(w, http.StatusOK, entity)
}

// GetFeatureFlags retrieves feature flags
// @Summary Get feature flags
// @Description Retrieve all feature flags
// @Tags configuration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param stage query string false "Stage filter (alpha, beta, stable)"
// @Success 200 {array} models.Entity
// @Router /api/v1/feature-flags [get]
func (h *EntityConfigHandler) GetFeatureFlags(w http.ResponseWriter, r *http.Request) {
	stage := r.URL.Query().Get("stage")
	
	query := h.entityRepo.Query().HasTag("type:feature_flag")
	
	if stage != "" {
		query = query.HasWildcardTag("feat:" + stage + ":*")
	}
	
	entities, err := query.Execute()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve feature flags")
		return
	}
	
	RespondJSON(w, http.StatusOK, entities)
}

// SetFeatureFlag creates or updates a feature flag
// @Summary Set feature flag
// @Description Update a feature flag value
// @Tags configuration
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body FeatureFlagRequest true "Feature flag data"
// @Success 200 {object} models.Entity
// @Router /api/v1/feature-flags/set [post]
func (h *EntityConfigHandler) SetFeatureFlag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Flag    string `json:"flag"`
		Enabled bool   `json:"enabled"`
		Stage   string `json:"stage"` // alpha, beta, stable
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	if req.Flag == "" {
		RespondError(w, http.StatusBadRequest, "Flag name is required")
		return
	}
	
	if req.Stage == "" {
		req.Stage = "stable"
	}
	
	// Create feature flag entity
	entityID := "feature_" + req.Flag
	
	tags := []string{
		"type:feature_flag",
		"feat:" + req.Stage + ":" + req.Flag,
	}
	
	if req.Enabled {
		tags = append(tags, "status:enabled")
	} else {
		tags = append(tags, "status:disabled")
	}
	
	// Store flag information as structured content
	flagData, _ := json.Marshal(map[string]interface{}{
		"flag":    req.Flag,
		"enabled": req.Enabled,
		"stage":   req.Stage,
	})
	
	entity := &models.Entity{
		ID:        entityID,
		Tags:      append(tags,
			"content:type:json",
			"flag:" + req.Flag,
			"stage:" + req.Stage,
			"updated_at:" + time.Now().UTC().Format(time.RFC3339),
		),
		Content:   flagData, // Store as JSON bytes
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	
	// Check if exists
	existing, err := h.entityRepo.GetByID(entityID)
	if err == nil && existing != nil {
		// Update existing
		entity.CreatedAt = existing.CreatedAt
		if err := h.entityRepo.Update(entity); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to update feature flag")
			return
		}
	} else {
		// Create new
		if err := h.entityRepo.Create(entity); err != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to create feature flag")
			return
		}
	}
	
	RespondJSON(w, http.StatusOK, entity)
}
