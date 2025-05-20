package api

import (
	"entitydb/models"
	"log"
	"net/http"
	"strings"
)

// RegisterWorkspaceRoutes registers proper workspace management routes
func RegisterWorkspaceRoutes(router *Router, entityRepo models.EntityRepository, auth *Auth) {
	// Create handler instance
	handler := &WorkspaceHandler{
		entityRepo: entityRepo,
		auth:       auth,
	}
	
	// Register routes
	router.POST("/api/v1/direct/workspace/create", handler.CreateWorkspace)
	router.GET("/api/v1/direct/workspace/list", handler.ListWorkspaces)
	router.GET("/api/v1/direct/workspace/get", handler.GetWorkspace)
	
	log.Println("Registered workspace API routes")
}

// WorkspaceHandler handles workspace-related API endpoints
type WorkspaceHandler struct {
	entityRepo models.EntityRepository
	auth       *Auth
}

// CreateWorkspace creates a new workspace
func (h *WorkspaceHandler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Priority    string   `json:"priority"`
		Tags        []string `json:"tags"`
		CreatorID   string   `json:"creator_id"`
	}
	
	if err := DecodeJSONBody(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}
	
	// Validate request
	if req.Title == "" {
		RespondError(w, http.StatusBadRequest, "Title is required")
		return
	}
	
	// Determine creator
	creatorID := req.CreatorID
	if creatorID == "" {
		// Try to get from auth context
		userID, ok := r.Context().Value(UserIDKey{}).(string)
		if ok && userID != "" {
			creatorID = userID
		} else {
			// Default creator
			creatorID = "system"
		}
	}
	
	// Set default priority if not provided
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	
	// Create entity tags
	tags := []string{
		"type:workspace",
		"status:active",
		"priority:" + priority,
		"creator:" + creatorID,
	}
	
	// Add custom tags if provided
	if len(req.Tags) > 0 {
		tags = append(tags, req.Tags...)
	}
	
	// Create content items
	// Generate entity ID based on title
	workspaceID := "workspace_" + models.SanitizeIDString(req.Title)

	// Create entity with predetermined ID
	entity := models.NewEntity(workspaceID)
	
	// Add content
	entity.AddContent("title", req.Title)

	if req.Description != "" {
		entity.AddContent("description", req.Description)
	}
	
	// ID already generated above
	
	// Add tags
	entity.Tags = tags
	
	// Try to create
	createdEntity, err := h.entityRepo.Create(entity)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			// If entity already exists with this ID, try getting it
			existingEntity, getErr := h.entityRepo.GetByID(workspaceID)
			if getErr == nil {
				RespondJSON(w, http.StatusOK, map[string]interface{}{
					"id":          existingEntity.ID,
					"title":       GetEntityTitle(existingEntity),
					"description": GetEntityDescription(existingEntity),
					"status":      GetEntityStatus(existingEntity),
					"priority":    GetEntityPriority(existingEntity),
					"tags":        existingEntity.Tags,
					"message":     "Workspace already exists",
				})
				return
			}
		}
		
		// If error is not related to existing entity or we couldn't retrieve it
		log.Printf("Error creating workspace entity: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to create workspace: "+err.Error())
		return
	}
	
	// Convert to response format
	workspace := map[string]interface{}{
		"id":          createdEntity.ID,
		"title":       GetEntityTitle(createdEntity),
		"description": GetEntityDescription(createdEntity),
		"status":      GetEntityStatus(createdEntity),
		"priority":    GetEntityPriority(createdEntity),
		"tags":        createdEntity.Tags,
	}
	
	// Return success
	RespondJSON(w, http.StatusCreated, workspace)
}

// ListWorkspaces lists all workspaces
func (h *WorkspaceHandler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	// Find entities with workspace tag
	entities, err := h.entityRepo.ListByTag("type:workspace")
	if err != nil {
		log.Printf("Error listing workspaces: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to list workspaces")
		return
	}
	
	// Convert to response format
	workspaces := make([]map[string]interface{}, 0, len(entities))
	for _, entity := range entities {
		workspace := map[string]interface{}{
			"id":          entity.ID,
			"title":       GetEntityTitle(entity),
			"description": GetEntityDescription(entity),
			"status":      GetEntityStatus(entity),
			"priority":    GetEntityPriority(entity),
		}
		workspaces = append(workspaces, workspace)
	}
	
	// Return workspaces
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"workspaces": workspaces,
		"total":      len(workspaces),
	})
}

// GetWorkspace gets a workspace by ID
func (h *WorkspaceHandler) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	// Get workspace ID from query
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		RespondError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}
	
	// Get entity
	entity, err := h.entityRepo.GetByID(workspaceID)
	if err != nil {
		log.Printf("Error getting workspace %s: %v", workspaceID, err)
		RespondError(w, http.StatusNotFound, "Workspace not found")
		return
	}
	
	// Check if it's a workspace type
	isWorkspace := false
	for _, tag := range entity.Tags {
		if tag == "type:workspace" {
			isWorkspace = true
			break
		}
	}
	
	if !isWorkspace {
		RespondError(w, http.StatusBadRequest, "Entity is not a workspace")
		return
	}
	
	// Convert to response format
	workspace := map[string]interface{}{
		"id":          entity.ID,
		"title":       GetEntityTitle(entity),
		"description": GetEntityDescription(entity),
		"status":      GetEntityStatus(entity),
		"priority":    GetEntityPriority(entity),
		"tags":        entity.Tags,
		"content":     entity.Content,
	}
	
	// Return workspace
	RespondJSON(w, http.StatusOK, workspace)
}

// Helper functions for entity extraction
func GetEntityTitle(entity *models.Entity) string {
	for _, content := range entity.Content {
		if content.Type == "title" {
			return content.Value
		}
	}
	return ""
}

func GetEntityDescription(entity *models.Entity) string {
	for _, content := range entity.Content {
		if content.Type == "description" {
			return content.Value
		}
	}
	return ""
}

func GetEntityStatus(entity *models.Entity) string {
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "status:") {
			return tag[7:]
		}
	}
	return "pending"
}

func GetEntityPriority(entity *models.Entity) string {
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "priority:") {
			return tag[9:]
		}
	}
	return "medium"
}