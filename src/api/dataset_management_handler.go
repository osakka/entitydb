package api

import (
	"encoding/json"
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DatasetManagementHandler handles hub management operations
type DatasetManagementHandler struct {
	repo models.EntityRepository
}

// NewDatasetManagementHandler creates a new hub management handler
func NewDatasetManagementHandler(repo models.EntityRepository) *DatasetManagementHandler {
	return &DatasetManagementHandler{
		repo: repo,
	}
}

// CreateDatasetRequest represents a request to create a new hub
type CreateDatasetRequest struct {
	Name        string `json:"name"`        // Hub name (required)
	Description string `json:"description"` // Hub description
	AdminUser   string `json:"admin_user"`  // Initial admin user ID (optional)
}

// CreateDatasetResponse represents the response from creating a hub
type CreateDatasetResponse struct {
	Hub     DatasetInfo `json:"hub"`
	Message string  `json:"message"`
}

// DatasetInfo represents hub information
type DatasetInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	AdminUser   string `json:"admin_user,omitempty"`
}

// ListDatasetsResponse represents the response from listing hubs
type ListDatasetsResponse struct {
	Hubs  []DatasetInfo `json:"hubs"`
	Total int       `json:"total"`
}

// CreateDataset handles creating a new hub
func (h *DatasetManagementHandler) CreateDataset(w http.ResponseWriter, r *http.Request) {
	// Get RBAC context
	rbacCtx, hasRBAC := GetRBACContext(r)
	if !hasRBAC {
		RespondError(w, http.StatusUnauthorized, "RBAC context required")
		return
	}

	// Check hub creation permission
	if !CheckHubManagementPermission(rbacCtx) {
		RespondError(w, http.StatusForbidden, "No hub creation permission")
		return
	}

	// Parse request
	var req CreateDatasetRequest
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
	existingHubs, err := h.repo.ListByTags([]string{FormatDatasetTag(req.Name)}, true)
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
	hubEntity.AddTag(fmt.Sprintf("type:dataset"))
	hubEntity.AddTag(FormatDatasetTag(req.Name))
	hubEntity.AddTag(fmt.Sprintf("dataset_name:%s", req.Name))
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
		logger.Error("Failed to create dataset entity: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to create dataset")
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
	response := CreateDatasetResponse{
		Hub: DatasetInfo{
			Name:        req.Name,
			Description: req.Description,
			CreatedAt:   time.Unix(0, hubEntity.CreatedAt).Format(time.RFC3339),
			AdminUser:   req.AdminUser,
		},
		Message: fmt.Sprintf("Hub '%s' created successfully", req.Name),
	}

	RespondJSON(w, http.StatusCreated, response)
}

// ListDatasets handles listing all hubs user has access to
func (h *DatasetManagementHandler) ListDatasets(w http.ResponseWriter, r *http.Request) {
	// Get RBAC context
	rbacCtx, hasRBAC := GetRBACContext(r)
	if !hasRBAC {
		RespondError(w, http.StatusUnauthorized, "RBAC context required")
		return
	}

	// Query all dataset entities
	hubEntities, err := h.repo.ListByTags([]string{"type:dataset"}, false) // matchAll = false for single tag
	if err != nil {
		logger.Error("Failed to list hubs: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to list hubs")
		return
	}

	// Filter hubs user has access to
	var accessibleHubs []DatasetInfo
	for _, entity := range hubEntities {
		datasetName := h.extractDatasetName(entity)
		if datasetName == "" {
			continue
		}

		// Check if user can view this hub
		if rbacCtx.IsAdmin || CheckDatasetPermission(rbacCtx, datasetName, "view") {
			hubInfo := h.entityToDatasetInfo(entity)
			accessibleHubs = append(accessibleHubs, hubInfo)
		}
	}

	// Return response
	response := ListDatasetsResponse{
		Hubs:  accessibleHubs,
		Total: len(accessibleHubs),
	}

	RespondJSON(w, http.StatusOK, response)
}

// DeleteDataset handles deleting a hub
func (h *DatasetManagementHandler) DeleteDataset(w http.ResponseWriter, r *http.Request) {
	// Get RBAC context
	rbacCtx, hasRBAC := GetRBACContext(r)
	if !hasRBAC {
		RespondError(w, http.StatusUnauthorized, "RBAC context required")
		return
	}

	// Get hub name from query parameter
	datasetName := r.URL.Query().Get("name")
	if datasetName == "" {
		RespondError(w, http.StatusBadRequest, "hub name is required")
		return
	}

	// Check hub deletion permission
	if !CheckHubManagementPermission(rbacCtx) {
		RespondError(w, http.StatusForbidden, fmt.Sprintf("No delete permission for hub: %s", datasetName))
		return
	}

	// Find hub entity
	hubEntities, err := h.repo.ListByTags([]string{"type:dataset", FormatDatasetTag(datasetName)}, true)
	if err != nil {
		logger.Error("Failed to find hub: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to find hub")
		return
	}

	if len(hubEntities) == 0 {
		RespondError(w, http.StatusNotFound, fmt.Sprintf("dataset not found: %s", datasetName))
		return
	}

	// Check if hub has any entities (prevent deletion of non-empty hubs)
	hubData, err := h.repo.ListByTags([]string{FormatDatasetTag(datasetName)}, false)
	if err != nil {
		logger.Error("Failed to check hub contents: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to check hub contents")
		return
	}

	// Count non-dataset entities (exclude the hub entity itself)
	nonHubEntities := 0
	for _, entity := range hubData {
		isHubEntity := false
		for _, tag := range entity.GetTagsWithoutTimestamp() {
			if tag == "type:dataset" {
				isHubEntity = true
				break
			}
		}
		if !isHubEntity {
			nonHubEntities++
		}
	}

	if nonHubEntities > 0 { // Any non-dataset entities exist
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

	logger.Info("Hub deleted: %s by user %s", datasetName, rbacCtx.User.ID)

	RespondJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Hub '%s' deleted successfully", datasetName),
	})
}

// Helper functions

// assignHubAdmin assigns hub admin permissions to a user
func (h *DatasetManagementHandler) assignHubAdmin(userID, datasetName string) error {
	// Get user entity
	user, err := h.repo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %s", userID)
	}

	// Add hub admin permission tags
	user.AddTag(fmt.Sprintf("rbac:perm:entity:*:dataset:%s", datasetName))
	user.AddTag(fmt.Sprintf("rbac:perm:dataset:manage:%s", datasetName))

	// Update user entity
	return h.repo.Update(user)
}

// extractDatasetName extracts hub name from entity tags
func (h *DatasetManagementHandler) extractDatasetName(entity *models.Entity) string {
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if datasetName, ok := ParseDatasetTag(tag); ok {
			return datasetName
		}
	}
	return ""
}

// entityToDatasetInfo converts entity to hub info
func (h *DatasetManagementHandler) entityToDatasetInfo(entity *models.Entity) DatasetInfo {
	info := DatasetInfo{
		CreatedAt: time.Unix(0, entity.CreatedAt).Format(time.RFC3339),
	}

	// Parse tags for hub information
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if datasetName, ok := ParseDatasetTag(tag); ok {
			info.Name = datasetName
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

// CheckDatasetPermission checks if a user has permission for a specific dataset
func CheckDatasetPermission(rbacCtx *RBACContext, datasetName, action string) bool {
	// Check if user has global dataset permissions or specific dataset permissions
	requiredPerm := "dataset:" + action
	globalPerm := "dataset:*"
	
	for _, perm := range rbacCtx.Permissions {
		if perm == requiredPerm || perm == globalPerm || perm == "*" {
			return true
		}
		// Check dataset-specific permissions
		if perm == "dataset:"+datasetName+":"+action || perm == "dataset:"+datasetName+":*" {
			return true
		}
	}
	return false
}

// ParseDatasetTag parses a dataset tag and returns the dataset name
func ParseDatasetTag(tag string) (string, bool) {
	if strings.HasPrefix(tag, "dataset:") {
		datasetName := strings.TrimPrefix(tag, "dataset:")
		if datasetName != "" && datasetName != "system" {
			return datasetName, true
		}
	}
	return "", false
}

// FormatDatasetTag creates a properly formatted dataset tag
func FormatDatasetTag(datasetName string) string {
	return "dataset:" + datasetName
}

// CheckHubManagementPermission checks if user can manage datasets/hubs
func CheckHubManagementPermission(rbacCtx *RBACContext) bool {
	requiredPerms := []string{"dataset:manage", "dataset:*", "admin:*", "*"}
	
	for _, perm := range rbacCtx.Permissions {
		for _, required := range requiredPerms {
			if perm == required {
				return true
			}
		}
	}
	return rbacCtx.IsAdmin
}