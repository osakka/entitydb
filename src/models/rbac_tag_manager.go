package models

import (
	"fmt"
	"strings"
	"entitydb/logger"
)

// RBACTagManager manages RBAC tags on user entities
type RBACTagManager struct {
	entityRepo EntityRepository
}

// NewRBACTagManager creates a new RBAC tag manager
func NewRBACTagManager(entityRepo EntityRepository) *RBACTagManager {
	return &RBACTagManager{
		entityRepo: entityRepo,
	}
}

// AssignRoleToUser adds a role tag directly to a user entity
func (rtm *RBACTagManager) AssignRoleToUser(userID, roleName string) error {
	logger.Info("[RBACTagManager] Assigning role %s to user %s", roleName, userID)
	
	// Get the user entity
	user, err := rtm.entityRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", userID)
	}
	
	// Check if user already has this role
	roleTag := fmt.Sprintf("rbac:role:%s", roleName)
	hasRole := false
	
	cleanTags := user.GetTagsWithoutTimestamp()
	for _, tag := range cleanTags {
		if tag == roleTag {
			hasRole = true
			break
		}
	}
	
	if hasRole {
		logger.Debug("[RBACTagManager] User %s already has role %s", userID, roleName)
		return nil
	}
	
	// Add the role tag
	user.Tags = append(user.Tags, roleTag)
	
	// Update the user entity
	if err := rtm.entityRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user with role: %v", err)
	}
	
	logger.Info("[RBACTagManager] Successfully assigned role %s to user %s", roleName, userID)
	return nil
}

// RemoveRoleFromUser removes a role tag from a user entity
func (rtm *RBACTagManager) RemoveRoleFromUser(userID, roleName string) error {
	logger.Info("[RBACTagManager] Removing role %s from user %s", roleName, userID)
	
	// Get the user entity
	user, err := rtm.entityRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", userID)
	}
	
	// Remove the role tag
	roleTag := fmt.Sprintf("rbac:role:%s", roleName)
	newTags := []string{}
	
	for _, tag := range user.Tags {
		cleanTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			cleanTag = parts[1]
		}
		if cleanTag != roleTag {
			newTags = append(newTags, tag)
		}
	}
	
	user.Tags = newTags
	
	// Update the user entity
	if err := rtm.entityRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	
	logger.Info("[RBACTagManager] Successfully removed role %s from user %s", roleName, userID)
	return nil
}

// GetUserRoles returns all roles assigned to a user
func (rtm *RBACTagManager) GetUserRoles(userID string) ([]string, error) {
	// Get the user entity
	user, err := rtm.entityRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	
	// Extract role tags
	roles := []string{}
	cleanTags := user.GetTagsWithoutTimestamp()
	
	for _, tag := range cleanTags {
		if strings.HasPrefix(tag, "rbac:role:") {
			roleName := strings.TrimPrefix(tag, "rbac:role:")
			roles = append(roles, roleName)
		}
	}
	
	return roles, nil
}

// AssignPermissionToUser adds a permission tag directly to a user entity
func (rtm *RBACTagManager) AssignPermissionToUser(userID, resource, action string) error {
	logger.Info("[RBACTagManager] Assigning permission %s:%s to user %s", resource, action, userID)
	
	// Get the user entity
	user, err := rtm.entityRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", userID)
	}
	
	// Check if user already has this permission
	permTag := fmt.Sprintf("rbac:perm:%s:%s", resource, action)
	hasPerm := false
	
	cleanTags := user.GetTagsWithoutTimestamp()
	for _, tag := range cleanTags {
		if tag == permTag {
			hasPerm = true
			break
		}
	}
	
	if hasPerm {
		logger.Debug("[RBACTagManager] User %s already has permission %s:%s", userID, resource, action)
		return nil
	}
	
	// Add the permission tag
	user.Tags = append(user.Tags, permTag)
	
	// Update the user entity
	if err := rtm.entityRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user with permission: %v", err)
	}
	
	logger.Info("[RBACTagManager] Successfully assigned permission %s:%s to user %s", resource, action, userID)
	return nil
}

// SyncRolePermissions ensures a user has all permissions granted by their roles
// This is useful for denormalizing permissions for faster checks
func (rtm *RBACTagManager) SyncRolePermissions(userID string) error {
	logger.Info("[RBACTagManager] Syncing role permissions for user %s", userID)
	
	// Get user roles
	roles, err := rtm.GetUserRoles(userID)
	if err != nil {
		return fmt.Errorf("failed to get user roles: %v", err)
	}
	
	// For each role, get its permissions and assign to user
	for _, roleName := range roles {
		if roleName == "admin" {
			// Admin gets all permissions
			if err := rtm.AssignPermissionToUser(userID, "*", "*"); err != nil {
				return fmt.Errorf("failed to assign admin permissions: %v", err)
			}
		}
		// Add more role-permission mappings as needed
	}
	
	return nil
}

// MigrateUsersToDirectTags migrates all users to have direct RBAC tags
func (rtm *RBACTagManager) MigrateUsersToDirectTags() error {
	logger.Info("[RBACTagManager] Starting user RBAC tag migration")
	
	// Get all users
	users, err := rtm.entityRepo.ListByTag("type:user")
	if err != nil {
		return fmt.Errorf("failed to list users: %v", err)
	}
	
	migratedCount := 0
	for _, user := range users {
		// Check for admin user
		cleanTags := user.GetTagsWithoutTimestamp()
		isAdmin := false
		
		for _, tag := range cleanTags {
			if tag == "identity:username:admin" {
				isAdmin = true
				break
			}
		}
		
		if isAdmin {
			// Assign admin role
			if err := rtm.AssignRoleToUser(user.ID, "admin"); err != nil {
				logger.Error("[RBACTagManager] Failed to assign admin role to user %s: %v", user.ID, err)
			} else {
				migratedCount++
			}
		}
	}
	
	logger.Info("[RBACTagManager] Migration complete. Migrated %d users", migratedCount)
	return nil
}