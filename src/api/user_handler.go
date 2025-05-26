package api

import (
	"entitydb/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"entitydb/logger"
	
	"golang.org/x/crypto/bcrypt"
)

// UserHandler handles user-related API endpoints through entity system
type UserHandler struct {
	entityRepo models.EntityRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(entityRepo models.EntityRepository) *UserHandler {
	return &UserHandler{
		entityRepo: entityRepo,
	}
}

// ChangePasswordRequest represents a request to change current user's password
type ChangePasswordRequest struct {
	Username        string `json:"username"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ResetPasswordRequest represents a request to reset another user's password (admin only)
type ResetPasswordRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

// CreateUser handles user creation as an entity
// @Summary Create a new user
// @Description Create a new user entity with authentication credentials
// @Tags users
// @Accept json
// @Produce json
// @Param body body CreateUserRequest true "User data"
// @Success 201 {object} models.Entity
// @Router /api/v1/users/create [post]
// POST /api/v1/users/create
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Username == "" {
		RespondError(w, http.StatusBadRequest, "Username is required")
		return
	}

	if req.Password == "" {
		RespondError(w, http.StatusBadRequest, "Password is required")
		return
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	passwordHash := string(hashedPassword)

	// Create user as entity
	tags := []string{
		"type:user",
		"id:username:" + req.Username,
		"status:active",
		"rbac:role:" + getRole(req.Role),
	}
	
	// Add default permissions based on role
	if getRole(req.Role) == "admin" {
		tags = append(tags, "rbac:perm:*")
	} else if getRole(req.Role) == "user" {
		// Regular users get basic permissions
		tags = append(tags, 
			"rbac:perm:entity:view",
			"rbac:perm:entity:create",
			"rbac:perm:entity:update",
			"rbac:perm:entity:delete",
		)
	}

	// Create content
	contentData := map[string]string{
		"username":     req.Username,
		"email":        req.Email,
		"full_name":    req.FullName,
		"password_hash": passwordHash,
		"created_at":   models.NowString(),
	}
	jsonData, _ := json.Marshal(contentData)

	// Create entity
	entity := &models.Entity{
		ID:        "user_" + req.Username,
		Tags:      append(tags, "content:type:json"),
		Content:   jsonData,
		CreatedAt: models.Now(),
		UpdatedAt: models.Now(),
	}

	// Save to repository
	if err := h.entityRepo.Create(entity); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Return created entity (without password hash)
	// For the new model, we need to redact the password hash in the JSON content
	if len(entity.Content) > 0 {
		var contentData map[string]string
		if err := json.Unmarshal(entity.Content, &contentData); err == nil {
			contentData["password_hash"] = "[REDACTED]"
			jsonData, _ := json.Marshal(contentData)
			entity.Content = jsonData
		}
	}

	RespondJSON(w, http.StatusCreated, entity)
}

// ChangePassword handles changing a user's own password
// @Summary Change user's own password
// @Description Change the current user's password
// @Tags users
// @Accept json
// @Produce json
// @Param body body ChangePasswordRequest true "Password change data"
// @Success 200 {object} StatusResponse
// @Router /api/v1/users/change-password [post]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate request
	if req.Username == "" {
		RespondError(w, http.StatusBadRequest, "Username is required")
		return
	}
	
	if req.CurrentPassword == "" || req.NewPassword == "" {
		RespondError(w, http.StatusBadRequest, "Current and new passwords are required")
		return
	}
	
	// Get RBAC context to verify identity
	rbacCtx, ok := GetRBACContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	// Extract username from current user
	userID := rbacCtx.User.ID
	userEntity := rbacCtx.User
	
	// For normal users, verify they're changing their own password
	requestUserEntityID := "user_" + req.Username
	if !rbacCtx.IsAdmin && userID != requestUserEntityID {
		RespondError(w, http.StatusForbidden, "You can only change your own password")
		return
	}
	
	// If admin is changing someone else's password, get that user's entity
	if rbacCtx.IsAdmin && userID != requestUserEntityID {
		fetchedEntity, err := h.entityRepo.GetByID(requestUserEntityID)
		if err != nil {
			RespondError(w, http.StatusNotFound, "User not found")
			return
		}
		userEntity = fetchedEntity
	}
	
	// Parse user data from content
	userData, err := parseUserData(userEntity.Content)
	if err != nil {
		logger.Error("Failed to parse user data: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to process user data")
		return
	}
	
	// Verify current password
	storedHash, ok := userData["password_hash"]
	if !ok {
		RespondError(w, http.StatusInternalServerError, "Invalid user data: missing password hash")
		return
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.CurrentPassword))
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "Current password is incorrect")
		return
	}
	
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to hash new password")
		return
	}
	
	// Update password in user data
	userData["password_hash"] = string(hashedPassword)
	userData["updated_at"] = models.NowString()
	
	// Marshal updated user data
	updatedContent, err := json.Marshal(userData)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to encode user data")
		return
	}
	
	// Update entity
	userEntity.Content = updatedContent
	userEntity.UpdatedAt = models.Now()
	
	// Save to repository
	if err := h.entityRepo.Update(userEntity); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}
	
	// Return success
	RespondJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Password changed successfully"})
}

// ResetPassword handles admin resetting a user's password
// @Summary Reset user password (admin only)
// @Description Reset a user's password (requires admin permission)
// @Tags users
// @Accept json
// @Produce json
// @Param body body ResetPasswordRequest true "Password reset data"
// @Success 200 {object} StatusResponse
// @Router /api/v1/users/reset-password [post]
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	
	// Validate request
	if req.UserID == "" && req.Username == "" {
		RespondError(w, http.StatusBadRequest, "Either user_id or username is required")
		return
	}
	
	if req.Password == "" {
		RespondError(w, http.StatusBadRequest, "Password is required")
		return
	}
	
	// Get RBAC context to verify admin rights
	rbacCtx, ok := GetRBACContext(r)
	if !ok {
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}
	
	// Security check: Only admins can reset passwords
	if !rbacCtx.IsAdmin {
		RespondError(w, http.StatusForbidden, "Administrator privileges required")
		return
	}
	
	// Find the user entity
	var userEntity *models.Entity
	var err error
	
	if req.UserID != "" {
		// Find by ID
		userEntity, err = h.entityRepo.GetByID(req.UserID)
	} else {
		// Find by username
		entityID := "user_" + req.Username
		userEntity, err = h.entityRepo.GetByID(entityID)
		
		if err != nil {
			// Try finding by tag
			userTag := "id:username:" + req.Username
			userEntities, err := h.entityRepo.ListByTag(userTag)
			if err != nil || len(userEntities) == 0 {
				RespondError(w, http.StatusNotFound, "User not found")
				return
			}
			userEntity = userEntities[0]
		}
	}
	
	if err != nil || userEntity == nil {
		RespondError(w, http.StatusNotFound, "User not found")
		return
	}
	
	// Verify this is a user entity
	isUserEntity := false
	for _, tag := range userEntity.Tags {
		if tag == "type:user" || strings.HasSuffix(tag, "|type:user") {
			isUserEntity = true
			break
		}
	}
	
	if !isUserEntity {
		RespondError(w, http.StatusBadRequest, "Invalid user entity")
		return
	}
	
	// Parse user data from content
	userData, err := parseUserData(userEntity.Content)
	if err != nil {
		logger.Error("Failed to parse user data: %v", err)
		RespondError(w, http.StatusInternalServerError, "Failed to process user data")
		return
	}
	
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to hash new password")
		return
	}
	
	// Update password in user data
	userData["password_hash"] = string(hashedPassword)
	userData["updated_at"] = models.NowString()
	
	// Marshal updated user data
	updatedContent, err := json.Marshal(userData)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to encode user data")
		return
	}
	
	// Update entity
	userEntity.Content = updatedContent
	userEntity.UpdatedAt = models.Now()
	
	// Save to repository
	if err := h.entityRepo.Update(userEntity); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}
	
	// Return success
	RespondJSON(w, http.StatusOK, map[string]string{"status": "success", "message": "Password reset successfully"})
}

// parseUserData extracts user data from entity content
func parseUserData(content []byte) (map[string]string, error) {
	userData := make(map[string]string)
	
	// With root cause fixed, content should be clean JSON
	if err := json.Unmarshal(content, &userData); err == nil {
		return userData, nil
	}
	
	// Fallback for existing wrapped content
	var wrapper map[string]interface{}
	if err := json.Unmarshal(content, &wrapper); err == nil {
		if innerContent, ok := wrapper["application/octet-stream"]; ok {
			if innerStr, ok := innerContent.(string); ok {
				if err := json.Unmarshal([]byte(innerStr), &userData); err == nil {
					return userData, nil
				}
			}
		}
	}
	
	return nil, fmt.Errorf("failed to parse user data")
}

// Helper function to determine user role
func getRole(role string) string {
	if role == "" {
		return "user"
	}
	if role == "admin" || role == "user" {
		return role
	}
	return "user"
}