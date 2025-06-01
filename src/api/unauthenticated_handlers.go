package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	
	"entitydb/models"
	"entitydb/logger"
)

// UnauthenticatedHandlers provides endpoints that bypass authentication for testing
type UnauthenticatedHandlers struct {
	entityRepo     models.EntityRepository
	relationshipRepo models.EntityRelationshipRepository
}

// NewUnauthenticatedHandlers creates handlers for unauthenticated endpoints
func NewUnauthenticatedHandlers(entityRepo models.EntityRepository, relationshipRepo models.EntityRelationshipRepository) *UnauthenticatedHandlers {
	return &UnauthenticatedHandlers{
		entityRepo:     entityRepo,
		relationshipRepo: relationshipRepo,
	}
}

// TestStatus returns a simple status response
func (h *UnauthenticatedHandlers) TestStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
		"test":   true,
	})
}

// TestGetEntityAsOf gets entity at a specific time without auth
func (h *UnauthenticatedHandlers) TestGetEntityAsOf(w http.ResponseWriter, r *http.Request) {
	entityID := r.URL.Query().Get("id")
	asOfStr := r.URL.Query().Get("as_of")
	
	if entityID == "" || asOfStr == "" {
		http.Error(w, "ID and as_of timestamp required", http.StatusBadRequest)
		return
	}
	
	asOf, err := time.Parse(time.RFC3339, asOfStr)
	if err != nil {
		http.Error(w, "Invalid timestamp format", http.StatusBadRequest)
		return
	}
	
	entity, err := h.entityRepo.GetEntityAsOf(entityID, asOf)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// TestGetEntityHistory gets entity history without auth
func (h *UnauthenticatedHandlers) TestGetEntityHistory(w http.ResponseWriter, r *http.Request) {
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		http.Error(w, "Entity ID required", http.StatusBadRequest)
		return
	}
	
	// Parse time range
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	
	// Parse timestamps for validation only (not used in current interface)
	if fromStr != "" {
		if _, err := time.Parse(time.RFC3339, fromStr); err != nil {
			http.Error(w, "Invalid from timestamp", http.StatusBadRequest)
			return
		}
	}
	
	if toStr != "" {
		if _, err := time.Parse(time.RFC3339, toStr); err != nil {
			http.Error(w, "Invalid to timestamp", http.StatusBadRequest)
			return
		}
	}
	
	// For now, use a default limit since the interface expects an int
	history, err := h.entityRepo.GetEntityHistory(entityID, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// TestGetRecentChanges gets recent changes without auth
func (h *UnauthenticatedHandlers) TestGetRecentChanges(w http.ResponseWriter, r *http.Request) {
	sinceStr := r.URL.Query().Get("since")
	
	// Parse timestamp for validation only (not used in current interface)
	if sinceStr != "" {
		if _, err := time.Parse(time.RFC3339, sinceStr); err != nil {
			http.Error(w, "Invalid since timestamp", http.StatusBadRequest)
			return
		}
	}
	
	// For now, use a limit of 100 since the interface expects an int
	changes, err := h.entityRepo.GetRecentChanges(100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(changes)
}

// TestGetEntityDiff gets entity diff without auth
func (h *UnauthenticatedHandlers) TestGetEntityDiff(w http.ResponseWriter, r *http.Request) {
	entityID := r.URL.Query().Get("id")
	t1Str := r.URL.Query().Get("t1")
	t2Str := r.URL.Query().Get("t2")
	
	if entityID == "" || t1Str == "" || t2Str == "" {
		http.Error(w, "ID, t1, and t2 timestamps required", http.StatusBadRequest)
		return
	}
	
	t1, err := time.Parse(time.RFC3339, t1Str)
	if err != nil {
		http.Error(w, "Invalid t1 timestamp", http.StatusBadRequest)
		return
	}
	
	t2, err := time.Parse(time.RFC3339, t2Str)
	if err != nil {
		http.Error(w, "Invalid t2 timestamp", http.StatusBadRequest)
		return
	}
	
	beforeEntity, afterEntity, err := h.entityRepo.GetEntityDiff(entityID, t1, t2)
	// Combine the two entities into a diff structure for the response
	diff := map[string]interface{}{
		"before": beforeEntity,
		"after":  afterEntity,
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diff)
}

// TestCreateEntity creates an entity without authentication
func (h *UnauthenticatedHandlers) TestCreateEntity(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID      string   `json:"id,omitempty"`
		Tags    []string `json:"tags"`
		Content []struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"content,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("[TestCreateEntity] Failed to decode request: %v", err)
		http.Error(w, "Invalid request body: " + err.Error(), http.StatusBadRequest)
		return
	}
	
	entity := &models.Entity{
		ID:        req.ID,
		Tags:      []string{},
		CreatedAt: models.Now(),
		UpdatedAt: models.Now(),
	}
	
	// Add tags
	for _, tag := range req.Tags {
		parts := strings.SplitN(tag, ":", 2)
		if len(parts) == 2 {
			entity.AddTagWithValue(parts[0], parts[1])
		} else {
			// If no colon, add the tag as is
			entity.AddTag(tag)
		}
	}
	
	// Add content
	if len(req.Content) > 0 {
		contentData := make(map[string]string)
		for _, c := range req.Content {
			contentData[c.Type] = c.Value
			entity.AddTag("content:type:" + c.Type)
		}
		jsonData, _ := json.Marshal(contentData)
		entity.Content = jsonData
	}
	
	// Save entity
	err := h.entityRepo.Create(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// TestCreateRelationship creates a relationship without authentication
func (h *UnauthenticatedHandlers) TestCreateRelationship(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceID         string `json:"source_id"`
		RelationshipType string `json:"relationship_type"`
		TargetID         string `json:"target_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	rel := models.NewEntityRelationship(req.SourceID, req.RelationshipType, req.TargetID)
	
	if err := h.relationshipRepo.Create(rel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rel)
}

// TestListRelationships lists relationships without authentication
func (h *UnauthenticatedHandlers) TestListRelationships(w http.ResponseWriter, r *http.Request) {
	sourceID := r.URL.Query().Get("source")
	targetID := r.URL.Query().Get("target")
	
	var relationships []*models.EntityRelationship
	var err error
	
	if sourceID != "" {
		relationships, err = h.relationshipRepo.GetBySource(sourceID)
	} else if targetID != "" {
		relationships, err = h.relationshipRepo.GetByTarget(targetID)
	} else {
		http.Error(w, "source or target parameter required", http.StatusBadRequest)
		return
	}
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(relationships)
}

// TestGetEntity gets an entity by ID without authentication
func (h *UnauthenticatedHandlers) TestGetEntity(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Entity ID required", http.StatusBadRequest)
		return
	}
	
	includeTimestamps := r.URL.Query().Get("include_timestamps") == "true"
	
	entity, err := h.entityRepo.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	// Strip timestamps if requested
	if !includeTimestamps {
		strippedEntity := *entity
		strippedEntity.Tags = entity.GetTagsWithoutTimestamp()
		entity = &strippedEntity
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entity)
}

// TestListEntities lists all entities without authentication
func (h *UnauthenticatedHandlers) TestListEntities(w http.ResponseWriter, r *http.Request) {
	includeTimestamps := r.URL.Query().Get("include_timestamps") == "true"
	
	entities, err := h.entityRepo.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Strip timestamps if requested
	if !includeTimestamps {
		for i, entity := range entities {
			strippedEntity := *entity
			strippedEntity.Tags = entity.GetTagsWithoutTimestamp()
			entities[i] = &strippedEntity
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}