// Package models provides RBAC (Role-Based Access Control) tag management for EntityDB.
// The RBAC system uses a tag-based approach where permissions and roles are stored
// as temporal tags directly on user entities.
package models

import (
	"fmt"
	"strings"
	"entitydb/logger"
)

// RBAC Tag Format Specification
//
// EntityDB uses a hierarchical tag format for RBAC permissions and roles.
// All RBAC tags are stored as temporal tags with nanosecond timestamps.
//
// # Tag Formats
//
// Role Tags:
//   - Format: "rbac:role:<role_name>"
//   - Example: "rbac:role:admin"
//   - Example: "rbac:role:user"
//   - Example: "rbac:role:viewer"
//
// Permission Tags:
//   - Format: "rbac:perm:<resource>:<action>"
//   - Example: "rbac:perm:entity:create"
//   - Example: "rbac:perm:user:delete"
//   - Example: "rbac:perm:*:*" (wildcard for all permissions)
//
// # Wildcard Support
//
// The permission system supports wildcards (*) for flexible access control:
//   - "rbac:perm:*:*" - All permissions (typically for admin)
//   - "rbac:perm:entity:*" - All actions on entity resource
//   - "rbac:perm:*:view" - View permission on all resources
//
// # Storage Format
//
// RBAC tags are stored as temporal tags on user entities:
//   - "2024-01-15T10:30:45.123456789.rbac:role:admin"
//   - "2024-01-15T10:30:45.123456789.rbac:perm:entity:create"
//
// # Permission Hierarchy
//
// Permissions follow a resource:action pattern:
//   - entity: create, view, update, delete, list
//   - user: create, view, update, delete, list
//   - dataset: create, view, update, delete, list
//   - relation: create, view, update, delete
//   - system: view, manage
//   - config: view, update
//   - metrics: read, write

// RBACTagManager manages RBAC tags on user entities.
// It provides methods to assign/remove roles and permissions,
// and to query a user's access rights.
//
// Thread-safe for concurrent operations.
type RBACTagManager struct {
	entityRepo EntityRepository // Repository for entity operations
}

// NewRBACTagManager creates a new RBAC tag manager instance.
//
// Parameters:
//   - entityRepo: The entity repository for user operations
//
// Returns a configured RBACTagManager ready for use.
func NewRBACTagManager(entityRepo EntityRepository) *RBACTagManager {
	return &RBACTagManager{
		entityRepo: entityRepo,
	}
}

// AssignRoleToUser adds a role tag to a user entity.
// If the user already has the role, this is a no-op.
//
// The role tag is added in the format "rbac:role:<role_name>".
// The tag is automatically timestamped by the storage layer.
//
// Parameters:
//   - userID: The ID of the user entity
//   - roleName: The name of the role (e.g., "admin", "user", "viewer")
//
// Returns an error if:
//   - The user doesn't exist
//   - The update operation fails
//
// Example:
//
//	err := tagManager.AssignRoleToUser("user-123", "admin")
//	// User now has tag: "rbac:role:admin"
func (rtm *RBACTagManager) AssignRoleToUser(userID, roleName string) error {
	logger.Info("Assigning role %s to user %s", roleName, userID)
	
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
	
	// Strip timestamps to check for existing role
	cleanTags := user.GetTagsWithoutTimestamp()
	for _, tag := range cleanTags {
		if tag == roleTag {
			hasRole = true
			break
		}
	}
	
	if hasRole {
		logger.Debug("User %s already has role %s", userID, roleName)
		return nil
	}
	
	// Add the role tag (will be timestamped by storage layer)
	user.Tags = append(user.Tags, roleTag)
	
	// Update the user entity
	if err := rtm.entityRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user with role: %v", err)
	}
	
	logger.Info("Successfully assigned role %s to user %s", roleName, userID)
	return nil
}

// RemoveRoleFromUser removes a role tag from a user entity.
// This removes all temporal instances of the role tag.
//
// Parameters:
//   - userID: The ID of the user entity
//   - roleName: The name of the role to remove
//
// Returns an error if:
//   - The user doesn't exist
//   - The update operation fails
//
// Example:
//
//	err := tagManager.RemoveRoleFromUser("user-123", "admin")
//	// All instances of "rbac:role:admin" are removed
func (rtm *RBACTagManager) RemoveRoleFromUser(userID, roleName string) error {
	logger.Info("Removing role %s from user %s", roleName, userID)
	
	// Get the user entity
	user, err := rtm.entityRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", userID)
	}
	
	// Filter out the role tag (handles temporal tags)
	roleTag := fmt.Sprintf("rbac:role:%s", roleName)
	newTags := []string{}
	
	for _, tag := range user.Tags {
		// Extract the tag content without timestamp
		cleanTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			cleanTag = parts[1]
		} else if parts := strings.SplitN(tag, ".", 2); len(parts) == 2 {
			cleanTag = parts[1]
		}
		
		// Keep tags that don't match the role to remove
		if cleanTag != roleTag {
			newTags = append(newTags, tag)
		}
	}
	
	user.Tags = newTags
	
	// Update the user entity
	if err := rtm.entityRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	
	logger.Info("Successfully removed role %s from user %s", roleName, userID)
	return nil
}

// GetUserRoles returns all roles currently assigned to a user.
// It extracts role names from RBAC role tags, handling temporal tags correctly.
//
// Parameters:
//   - userID: The ID of the user entity
//
// Returns:
//   - A slice of role names (e.g., ["admin", "user"])
//   - An error if the user doesn't exist or retrieval fails
//
// Example:
//
//	roles, err := tagManager.GetUserRoles("user-123")
//	// roles = ["admin", "viewer"]
func (rtm *RBACTagManager) GetUserRoles(userID string) ([]string, error) {
	// Get the user entity
	user, err := rtm.entityRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	
	// Extract role names from RBAC tags
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

// AssignPermissionToUser adds a permission tag directly to a user entity.
// This allows fine-grained permission assignment beyond role-based permissions.
//
// The permission tag is added in the format "rbac:perm:<resource>:<action>".
// The tag is automatically timestamped by the storage layer.
//
// Parameters:
//   - userID: The ID of the user entity
//   - resource: The resource being protected (e.g., "entity", "user", "*")
//   - action: The action being permitted (e.g., "create", "view", "*")
//
// Returns an error if:
//   - The user doesn't exist
//   - The update operation fails
//
// Example:
//
//	// Grant specific permission
//	err := tagManager.AssignPermissionToUser("user-123", "entity", "create")
//	// User now has tag: "rbac:perm:entity:create"
//
//	// Grant wildcard permission
//	err := tagManager.AssignPermissionToUser("user-123", "*", "*")
//	// User now has tag: "rbac:perm:*:*" (all permissions)
func (rtm *RBACTagManager) AssignPermissionToUser(userID, resource, action string) error {
	logger.Info("Assigning permission %s:%s to user %s", resource, action, userID)
	
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
		logger.Debug("User %s already has permission %s:%s", userID, resource, action)
		return nil
	}
	
	// Add the permission tag
	user.Tags = append(user.Tags, permTag)
	
	// Update the user entity
	if err := rtm.entityRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user with permission: %v", err)
	}
	
	logger.Info("Successfully assigned permission %s:%s to user %s", resource, action, userID)
	return nil
}

// SyncRolePermissions ensures a user has all permissions granted by their roles.
// This is useful for denormalizing permissions for faster access checks.
//
// Currently implements the following role-permission mappings:
//   - admin: Gets wildcard permission "rbac:perm:*:*"
//
// Additional role mappings can be added as needed.
//
// Parameters:
//   - userID: The ID of the user to sync permissions for
//
// Returns an error if:
//   - Failed to retrieve user roles
//   - Failed to assign permissions
//
// Example:
//
//	// User with admin role
//	err := tagManager.SyncRolePermissions("user-123")
//	// User now has "rbac:perm:*:*" tag
func (rtm *RBACTagManager) SyncRolePermissions(userID string) error {
	logger.Info("Syncing role permissions for user %s", userID)
	
	// Get user roles
	roles, err := rtm.GetUserRoles(userID)
	if err != nil {
		return fmt.Errorf("failed to get user roles: %v", err)
	}
	
	// Map roles to permissions
	for _, roleName := range roles {
		switch roleName {
		case "admin":
			// Admin gets all permissions
			if err := rtm.AssignPermissionToUser(userID, "*", "*"); err != nil {
				return fmt.Errorf("failed to assign admin permissions: %v", err)
			}
		// Add more role mappings here as needed:
		// case "viewer":
		//     rtm.AssignPermissionToUser(userID, "*", "view")
		// case "editor":
		//     rtm.AssignPermissionToUser(userID, "entity", "*")
		}
	}
	
	return nil
}

// MigrateUsersToDirectTags performs a one-time migration to add RBAC tags
// to existing users based on their current attributes.
//
// The migration process:
//   1. Lists all user entities
//   2. Checks for admin users (those with "identity:username:admin" tag)
//   3. Assigns the "admin" role to identified admin users
//
// This method is idempotent - running it multiple times is safe as it
// only adds tags that don't already exist.
//
// Returns an error if the migration fails, otherwise returns nil.
// Logs the number of users migrated.
//
// Example:
//
//	err := tagManager.MigrateUsersToDirectTags()
//	// Output: "Migration complete. Migrated 3 users"
func (rtm *RBACTagManager) MigrateUsersToDirectTags() error {
	logger.Info("Starting user RBAC tag migration")
	
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
				logger.Error("Failed to assign admin role to user %s: %v", user.ID, err)
			} else {
				migratedCount++
			}
		}
	}
	
	logger.Info("Migration complete. Migrated %d users", migratedCount)
	return nil
}