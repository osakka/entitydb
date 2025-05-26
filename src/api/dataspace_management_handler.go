package api

import (
	"encoding/json"
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"net/http"
	"time"
)

// DataspaceManagementHandler handles hub management operations
type DataspaceManagementHandler struct {
	repo models.EntityRepository
}

// NewDataspaceManagementHandler creates a new hub management handler
func NewDataspaceManagementHandler(repo models.EntityRepository) *DataspaceManagementHandler {
	return &DataspaceManagementHandler{
		repo: repo,
	}
}

// CreateDataspaceRequest represents a request to create a new hub
type CreateDataspaceRequest struct {
	Name        string `json:"name"`        // Hub name (required)
	Description string `json:"description"` // Hub description
	AdminUser   string `json:"admin_user"`  // Initial admin user ID (optional)
}

// CreateDataspaceResponse represents the response from creating a hub
type CreateDataspaceResponse struct {
	Hub     DataspaceInfo `json:"hub"`
	Message string  `json:"message"`
}

// DataspaceInfo represents hub information
type DataspaceInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	AdminUser   string `json:"admin_user,omitempty"`
}

// ListDataspacesResponse represents the response from listing hubs
type ListDataspacesResponse struct {
	Hubs  []DataspaceInfo `json:"hubs"`
	Total int       `json:"total"`
}

// CreateDataspace handles creating a new hub
func (h *DataspaceManagementHandler) CreateDataspace(w http.ResponseWriter, r *http.Request) {
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
	var req CreateDataspaceRequest
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
	existingHubs, err := h.repo.ListByTags([]string{FormatDataspaceTag(req.Name)}, true)
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
		CreatedAt: models.Now(),
		UpdatedAt: models.Now(),
	}

	// Add hub metadata tags
	hubEntity.AddTag(fmt.Sprintf("type:dataspace"))
	hubEntity.AddTag(FormatDataspaceTag(req.Name))
	hubEntity.AddTag(fmt.Sprintf("dataspace_name:%s", req.Name))
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
		logger.Error("Failed to create dataspace entity: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to create dataspace")
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
	response := CreateDataspaceResponse{
		Hub: DataspaceInfo{
			Name:        req.Name,
			Description: req.Description,
			CreatedAt:   time.Unix(0, hubEntity.CreatedAt).Format(time.RFC3339),
			AdminUser:   req.AdminUser,
		},
		Message: fmt.Sprintf("Hub '%s' created successfully", req.Name),
	}

	RespondJSON(w, http.StatusCreated, response)
}

// ListDataspaces handles listing all hubs user has access to
func (h *DataspaceManagementHandler) ListDataspaces(w http.ResponseWriter, r *http.Request) {
	// Get RBAC context
	rbacCtx, hasRBAC := GetRBACContext(r)
	if !hasRBAC {
		RespondError(w, http.StatusUnauthorized, "RBAC context required")
		return
	}

	// Query all dataspace entities
	hubEntities, err := h.repo.ListByTags([]string{"type:dataspace"}, false) // matchAll = false for single tag
	if err != nil {
		logger.Error("Failed to list hubs: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to list hubs")
		return
	}

	// Filter hubs user has access to
	var accessibleHubs []DataspaceInfo
	for _, entity := range hubEntities {
		dataspaceName := h.extractDataspaceName(entity)
		if dataspaceName == "" {
			continue
		}

		// Check if user can view this hub
		if rbacCtx.IsAdmin || CheckDataspacePermission(rbacCtx, dataspaceName, "view") {
			hubInfo := h.entityToDataspaceInfo(entity)
			accessibleHubs = append(accessibleHubs, hubInfo)
		}
	}

	// Return response
	response := ListDataspacesResponse{
		Hubs:  accessibleHubs,
		Total: len(accessibleHubs),
	}

	RespondJSON(w, http.StatusOK, response)
}

// DeleteDataspace handles deleting a hub
func (h *DataspaceManagementHandler) DeleteDataspace(w http.ResponseWriter, r *http.Request) {
	// Get RBAC context
	rbacCtx, hasRBAC := GetRBACContext(r)
	if !hasRBAC {
		RespondError(w, http.StatusUnauthorized, "RBAC context required")
		return
	}

	// Get hub name from query parameter
	dataspaceName := r.URL.Query().Get("name")
	if dataspaceName == "" {
		RespondError(w, http.StatusBadRequest, "hub name is required")
		return
	}

	// Check hub deletion permission
	if !CheckHubManagementPermission(rbacCtx, "delete", dataspaceName) {
		RespondError(w, http.StatusForbidden, fmt.Sprintf("No delete permission for hub: %s", dataspaceName))
		return
	}

	// Find hub entity
	hubEntities, err := h.repo.ListByTags([]string{"type:dataspace", FormatDataspaceTag(dataspaceName)}, true)
	if err != nil {
		logger.Error("Failed to find hub: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to find hub")
		return
	}

	if len(hubEntities) == 0 {
		RespondError(w, http.StatusNotFound, fmt.Sprintf("dataspace not found: %s", dataspaceName))
		return
	}

	// Check if hub has any entities (prevent deletion of non-empty hubs)
	hubData, err := h.repo.ListByTags([]string{FormatDataspaceTag(dataspaceName)}, false)
	if err != nil {
		logger.Error("Failed to check hub contents: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to check hub contents")
		return
	}

	// Count non-dataspace entities (exclude the hub entity itself)
	nonHubEntities := 0
	for _, entity := range hubData {
		isHubEntity := false
		for _, tag := range entity.GetTagsWithoutTimestamp() {
			if tag == "type:dataspace" {
				isHubEntity = true
				break
			}
		}
		if !isHubEntity {
			nonHubEntities++
		}
	}

	if nonHubEntities > 0 { // Any non-dataspace entities exist
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

	logger.Info("Hub deleted: %s by user %s", dataspaceName, rbacCtx.User.ID)

	RespondJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Hub '%s' deleted successfully", dataspaceName),
	})
}

// Helper functions

// assignHubAdmin assigns hub admin permissions to a user
func (h *DataspaceManagementHandler) assignHubAdmin(userID, dataspaceName string) error {
	// Get user entity
	user, err := h.repo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %s", userID)
	}

	// Add hub admin permission tags
	user.AddTag(fmt.Sprintf("rbac:perm:entity:*:dataspace:%s", dataspaceName))
	user.AddTag(fmt.Sprintf("rbac:perm:dataspace:manage:%s", dataspaceName))

	// Update user entity
	return h.repo.Update(user)
}

// extractDataspaceName extracts hub name from entity tags
func (h *DataspaceManagementHandler) extractDataspaceName(entity *models.Entity) string {
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if dataspaceName, ok := ParseDataspaceTag(tag); ok {
			return dataspaceName
		}
	}
	return ""
}

// entityToDataspaceInfo converts entity to hub info
func (h *DataspaceManagementHandler) entityToDataspaceInfo(entity *models.Entity) DataspaceInfo {
	info := DataspaceInfo{
		CreatedAt: time.Unix(0, entity.CreatedAt).Format(time.RFC3339),
	}

	// Parse tags for hub information
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if dataspaceName, ok := ParseDataspaceTag(tag); ok {
			info.Name = dataspaceName
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