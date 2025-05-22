package api

import (
	"encoding/json"
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"net/http"
	"time"
)

// HubManagementHandler handles hub management operations
type HubManagementHandler struct {
	repo models.EntityRepository
}

// NewHubManagementHandler creates a new hub management handler
func NewHubManagementHandler(repo models.EntityRepository) *HubManagementHandler {
	return &HubManagementHandler{
		repo: repo,
	}
}

// CreateHubRequest represents a request to create a new hub
type CreateHubRequest struct {
	Name        string `json:"name"`        // Hub name (required)
	Description string `json:"description"` // Hub description
	AdminUser   string `json:"admin_user"`  // Initial admin user ID (optional)
}

// CreateHubResponse represents the response from creating a hub
type CreateHubResponse struct {
	Hub     HubInfo `json:"hub"`
	Message string  `json:"message"`
}

// HubInfo represents hub information
type HubInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	AdminUser   string `json:"admin_user,omitempty"`
}

// ListHubsResponse represents the response from listing hubs
type ListHubsResponse struct {
	Hubs  []HubInfo `json:"hubs"`
	Total int       `json:"total"`
}

// CreateHub handles creating a new hub
func (h *HubManagementHandler) CreateHub(w http.ResponseWriter, r *http.Request) {
	// Get RBAC context
	rbacCtx, hasRBAC := GetRBACContext(r)
	if !hasRBAC {
		RespondError(w, http.StatusUnauthorized, "RBAC context required")
		return
	}

	// Check hub creation permission
	if !CheckHubManagementPermission(rbacCtx, "create", "") {
		RespondError(w, http.StatusForbidden, "No hub creation permission")
		return
	}

	// Parse request
	var req CreateHubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate hub name
	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, "hub name is required")
		return
	}

	// Check if hub already exists
	existingHubs, err := h.repo.ListByTags([]string{FormatHubTag(req.Name)}, true)
	if err != nil {
		logger.Error("Failed to check existing hubs: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to check existing hubs")
		return
	}

	if len(existingHubs) > 0 {
		RespondError(w, http.StatusConflict, fmt.Sprintf("hub already exists: %s", req.Name))
		return
	}

	// Create hub configuration entity
	hubEntity := &models.Entity{
		Tags:      []string{},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Add hub metadata tags
	hubEntity.AddTag(fmt.Sprintf("type:hub"))
	hubEntity.AddTag(FormatHubTag(req.Name))
	hubEntity.AddTag(fmt.Sprintf("hub_name:%s", req.Name))
	if req.Description != "" {
		hubEntity.AddTag(fmt.Sprintf("description:%s", req.Description))
	}
	hubEntity.AddTag(fmt.Sprintf("created_by:%s", rbacCtx.User.ID))

	// Set content as JSON
	hubContent := map[string]interface{}{
		"name":        req.Name,
		"description": req.Description,
		"created_by":  rbacCtx.User.ID,
		"created_at":  hubEntity.CreatedAt,
		"status":      "active",
	}

	contentBytes, _ := json.Marshal(hubContent)
	hubEntity.Content = contentBytes

	// Save hub entity
	err = h.repo.Create(hubEntity)
	if err != nil {
		logger.Error("Failed to create hub entity: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to create hub")
		return
	}

	// If admin user specified, create/update their permissions
	if req.AdminUser != "" {
		err = h.assignHubAdmin(req.AdminUser, req.Name)
		if err != nil {
			logger.Warn("Failed to assign hub admin: %v", err)
			// Don't fail the hub creation, just log the warning
		}
	}

	logger.Info("Hub created: %s by user %s", req.Name, rbacCtx.User.ID)

	// Return response
	response := CreateHubResponse{
		Hub: HubInfo{
			Name:        req.Name,
			Description: req.Description,
			CreatedAt:   hubEntity.CreatedAt,
			AdminUser:   req.AdminUser,
		},
		Message: fmt.Sprintf("Hub '%s' created successfully", req.Name),
	}

	RespondJSON(w, http.StatusCreated, response)
}

// ListHubs handles listing all hubs user has access to
func (h *HubManagementHandler) ListHubs(w http.ResponseWriter, r *http.Request) {
	// Get RBAC context
	rbacCtx, hasRBAC := GetRBACContext(r)
	if !hasRBAC {
		RespondError(w, http.StatusUnauthorized, "RBAC context required")
		return
	}

	// Query all hub entities
	hubEntities, err := h.repo.ListByTags([]string{"type:hub"}, false) // matchAll = false for single tag
	if err != nil {
		logger.Error("Failed to list hubs: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to list hubs")
		return
	}

	// Filter hubs user has access to
	var accessibleHubs []HubInfo
	for _, entity := range hubEntities {
		hubName := h.extractHubName(entity)
		if hubName == "" {
			continue
		}

		// Check if user can view this hub
		if rbacCtx.IsAdmin || CheckHubPermission(rbacCtx, hubName, "view") {
			hubInfo := h.entityToHubInfo(entity)
			accessibleHubs = append(accessibleHubs, hubInfo)
		}
	}

	// Return response
	response := ListHubsResponse{
		Hubs:  accessibleHubs,
		Total: len(accessibleHubs),
	}

	RespondJSON(w, http.StatusOK, response)
}

// DeleteHub handles deleting a hub
func (h *HubManagementHandler) DeleteHub(w http.ResponseWriter, r *http.Request) {
	// Get RBAC context
	rbacCtx, hasRBAC := GetRBACContext(r)
	if !hasRBAC {
		RespondError(w, http.StatusUnauthorized, "RBAC context required")
		return
	}

	// Get hub name from query parameter
	hubName := r.URL.Query().Get("name")
	if hubName == "" {
		RespondError(w, http.StatusBadRequest, "hub name is required")
		return
	}

	// Check hub deletion permission
	if !CheckHubManagementPermission(rbacCtx, "delete", hubName) {
		RespondError(w, http.StatusForbidden, fmt.Sprintf("No delete permission for hub: %s", hubName))
		return
	}

	// Find hub entity
	hubEntities, err := h.repo.ListByTags([]string{"type:hub", FormatHubTag(hubName)}, true)
	if err != nil {
		logger.Error("Failed to find hub: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to find hub")
		return
	}

	if len(hubEntities) == 0 {
		RespondError(w, http.StatusNotFound, fmt.Sprintf("hub not found: %s", hubName))
		return
	}

	// Check if hub has any entities (prevent deletion of non-empty hubs)
	hubData, err := h.repo.ListByTags([]string{FormatHubTag(hubName)}, false)
	if err != nil {
		logger.Error("Failed to check hub contents: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to check hub contents")
		return
	}

	// Count non-hub entities (exclude the hub entity itself)
	nonHubEntities := 0
	for _, entity := range hubData {
		isHubEntity := false
		for _, tag := range entity.GetTagsWithoutTimestamp() {
			if tag == "type:hub" {
				isHubEntity = true
				break
			}
		}
		if !isHubEntity {
			nonHubEntities++
		}
	}

	if nonHubEntities > 0 { // Any non-hub entities exist
		RespondError(w, http.StatusConflict, "cannot delete non-empty hub")
		return
	}

	// Delete hub entity
	err = h.repo.Delete(hubEntities[0].ID)
	if err != nil {
		logger.Error("Failed to delete hub: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to delete hub")
		return
	}

	logger.Info("Hub deleted: %s by user %s", hubName, rbacCtx.User.ID)

	RespondJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Hub '%s' deleted successfully", hubName),
	})
}

// Helper functions

// assignHubAdmin assigns hub admin permissions to a user
func (h *HubManagementHandler) assignHubAdmin(userID, hubName string) error {
	// Get user entity
	user, err := h.repo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %s", userID)
	}

	// Add hub admin permission tags
	user.AddTag(fmt.Sprintf("rbac:perm:entity:*:hub:%s", hubName))
	user.AddTag(fmt.Sprintf("rbac:perm:hub:manage:%s", hubName))

	// Update user entity
	return h.repo.Update(user)
}

// extractHubName extracts hub name from entity tags
func (h *HubManagementHandler) extractHubName(entity *models.Entity) string {
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if hubName, ok := ParseHubTag(tag); ok {
			return hubName
		}
	}
	return ""
}

// entityToHubInfo converts entity to hub info
func (h *HubManagementHandler) entityToHubInfo(entity *models.Entity) HubInfo {
	info := HubInfo{
		CreatedAt: entity.CreatedAt,
	}

	// Parse tags for hub information
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if hubName, ok := ParseHubTag(tag); ok {
			info.Name = hubName
		} else if len(tag) > 12 && tag[:12] == "description:" {
			info.Description = tag[12:]
		}
	}

	// Parse content for additional info
	if len(entity.Content) > 0 {
		var content map[string]interface{}
		if err := json.Unmarshal(entity.Content, &content); err == nil {
			if desc, ok := content["description"].(string); ok && info.Description == "" {
				info.Description = desc
			}
			if adminUser, ok := content["admin_user"].(string); ok {
				info.AdminUser = adminUser
			}
		}
	}

	return info
}