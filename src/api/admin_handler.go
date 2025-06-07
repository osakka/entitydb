package api

import (
	"encoding/json"
	"entitydb/logger"
	"entitydb/models"
	"net/http"
	"time"
)

// AdminHandler handles administrative operations
type AdminHandler struct {
	repo models.EntityRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(repo models.EntityRepository) *AdminHandler {
	return &AdminHandler{
		repo: repo,
	}
}

// ReindexRequest represents a request to reindex the database
type ReindexRequest struct {
	Force bool `json:"force"` // Force reindex even if healthy
}

// ReindexResponse represents the response from a reindex operation
type ReindexResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	EntitiesIndexed int    `json:"entities_indexed"`
	Duration       string `json:"duration"`
	Errors         []string `json:"errors,omitempty"`
}

// ReindexHandler handles the reindex request
func (h *AdminHandler) ReindexHandler(w http.ResponseWriter, r *http.Request) {
	// This should be wrapped with RBAC middleware requiring admin permission
	// For now, we'll add a basic check
	
	start := time.Now()
	
	// Decode request
	var req ReindexRequest
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}
	
	logger.Info("admin reindex requested with force=%v", req.Force)
	
	// Call the repository's reindex method
	if reindexer, ok := h.repo.(interface{ ReindexTags() error }); ok {
		err := reindexer.ReindexTags()
		if err != nil {
			logger.Error("reindex failed: %v", err)
			RespondJSON(w, http.StatusInternalServerError, ReindexResponse{
				Success: false,
				Message: "Reindex failed",
				Errors:  []string{err.Error()},
			})
			return
		}
		
		// Get entity count
		entities, _ := h.repo.List()
		
		RespondJSON(w, http.StatusOK, ReindexResponse{
			Success:         true,
			Message:         "Reindex completed successfully",
			EntitiesIndexed: len(entities),
			Duration:        time.Since(start).String(),
		})
	} else {
		RespondJSON(w, http.StatusNotImplemented, ReindexResponse{
			Success: false,
			Message: "Repository does not support reindexing",
		})
	}
}

// HealthCheckHandler provides detailed health information including index health
func (h *AdminHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Get basic health from repository
	health := map[string]interface{}{
		"status": "ok",
		"timestamp": models.Now(),
	}
	
	// Add index health if supported
	if healthChecker, ok := h.repo.(interface{ VerifyIndexHealth() error }); ok {
		err := healthChecker.VerifyIndexHealth()
		if err != nil {
			health["index_health"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			health["index_health"] = map[string]interface{}{
				"status": "healthy",
			}
		}
	}
	
	// Get entity and tag counts
	entities, _ := h.repo.List()
	health["entity_count"] = len(entities)
	
	RespondJSON(w, http.StatusOK, health)
}