package models

import (
	"fmt"
	"strings"
	
	"entitydb/logger"
)

// SecurityInitializer handles initial setup of security entities and relationships
type SecurityInitializer struct {
	securityManager *SecurityManager
	entityRepo      EntityRepository
}

// NewSecurityInitializer creates a new security initializer
func NewSecurityInitializer(securityManager *SecurityManager, entityRepo EntityRepository) *SecurityInitializer {
	return &SecurityInitializer{
		securityManager: securityManager,
		entityRepo:      entityRepo,
	}
}

// InitializeDefaultSecurityEntities creates the default roles, permissions, and admin user
func (si *SecurityInitializer) InitializeDefaultSecurityEntities() error {
	// Create default dataspaces first (system infrastructure)
	if err := si.createDefaultDataspaces(); err != nil {
		return fmt.Errorf("failed to create default dataspaces: %v", err)
	}
	
	// Force sync after creating dataspaces
	if err := si.forceSync(); err != nil {
		return fmt.Errorf("failed to sync after creating dataspaces: %v", err)
	}

	// Create default permissions
	if err := si.createDefaultPermissions(); err != nil {
		return fmt.Errorf("failed to create default permissions: %v", err)
	}
	
	// Force sync after creating permissions
	if err := si.forceSync(); err != nil {
		return fmt.Errorf("failed to sync after creating permissions: %v", err)
	}

	// Create default roles
	if err := si.createDefaultRoles(); err != nil {
		return fmt.Errorf("failed to create default roles: %v", err)
	}
	
	// Force sync after creating roles
	if err := si.forceSync(); err != nil {
		return fmt.Errorf("failed to sync after creating roles: %v", err)
	}

	// Create default groups
	if err := si.createDefaultGroups(); err != nil {
		return fmt.Errorf("failed to create default groups: %v", err)
	}
	
	// Force sync after creating groups
	if err := si.forceSync(); err != nil {
		return fmt.Errorf("failed to sync after creating groups: %v", err)
	}

	// Create admin user if it doesn't exist
	if err := si.createDefaultAdminUser(); err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	return nil
}

// createDefaultPermissions creates the basic permissions needed for the system
func (si *SecurityInitializer) createDefaultPermissions() error {
	permissions := []struct {
		id       string
		resource string
		action   string
		scope    string
	}{
		// Entity permissions
		{"perm_entity_view", "entity", "view", "global"},
		{"perm_entity_create", "entity", "create", "global"},
		{"perm_entity_update", "entity", "update", "global"},
		{"perm_entity_delete", "entity", "delete", "global"},
		{"perm_entity_all", "entity", "*", "global"},

		// User permissions
		{"perm_user_view", "user", "view", "global"},
		{"perm_user_create", "user", "create", "global"},
		{"perm_user_update", "user", "update", "global"},
		{"perm_user_delete", "user", "delete", "global"},
		{"perm_user_all", "user", "*", "global"},

		// Admin permissions
		{"perm_admin_view", "admin", "view", "global"},
		{"perm_admin_update", "admin", "update", "global"},
		{"perm_admin_all", "admin", "*", "global"},

		// System permissions
		{"perm_system_view", "system", "view", "global"},
		{"perm_system_update", "system", "update", "global"},
		{"perm_system_all", "system", "*", "global"},

		// Config permissions
		{"perm_config_view", "config", "view", "global"},
		{"perm_config_update", "config", "update", "global"},
		{"perm_config_all", "config", "*", "global"},

		// Relationship permissions
		{"perm_relation_view", "relation", "view", "global"},
		{"perm_relation_create", "relation", "create", "global"},
		{"perm_relation_update", "relation", "update", "global"},
		{"perm_relation_delete", "relation", "delete", "global"},
		{"perm_relation_all", "relation", "*", "global"},

		// Dataspace permissions
		{"perm_dataspace_view", "dataspace", "view", "global"},
		{"perm_dataspace_create", "dataspace", "create", "global"},
		{"perm_dataspace_update", "dataspace", "update", "global"},
		{"perm_dataspace_delete", "dataspace", "delete", "global"},
		{"perm_dataspace_manage", "dataspace", "manage", "global"},
		{"perm_dataspace_all", "dataspace", "*", "global"},
		
		// Metrics permissions
		{"perm_metrics_view", "metrics", "view", "global"},
		{"perm_metrics_write", "metrics", "write", "global"},
		{"perm_metrics_all", "metrics", "*", "global"},

		// Global wildcard permission
		{"perm_all", "*", "*", "global"},
	}

	for _, perm := range permissions {
		// During initial setup, skip existence check to avoid timing issues
		// The repository will handle duplicate creation gracefully
		
		permissionEntity := &Entity{
			ID: perm.id,
			Tags: []string{
				"type:" + EntityTypePermission,
				"dataspace:_system",
				"resource:" + perm.resource,
				"action:" + perm.action,
				"scope:" + perm.scope,
				"created:" + NowString(),
			},
			Content:   nil,
			CreatedAt: Now(),
			UpdatedAt: Now(),
		}

		if err := si.entityRepo.Create(permissionEntity); err != nil {
			return fmt.Errorf("failed to create permission %s: %v", perm.id, err)
		}
	}

	return nil
}

// createDefaultRoles creates the basic roles needed for the system
func (si *SecurityInitializer) createDefaultRoles() error {
	roles := []struct {
		id    string
		name  string
		level int
		scope string
	}{
		{"role_admin", "admin", 100, "global"},
		{"role_user", "user", 10, "global"},
		{"role_guest", "guest", 1, "global"},
	}

	for _, role := range roles {
		// During initial setup, skip existence check to avoid timing issues
		// The repository will handle duplicate creation gracefully
		
		roleEntity := &Entity{
			ID: role.id,
			Tags: []string{
				"type:" + EntityTypeRole,
				"dataspace:_system",
				"name:" + role.name,
				fmt.Sprintf("level:%d", role.level),
				"scope:" + role.scope,
				"created:" + NowString(),
			},
			Content:   nil,
			CreatedAt: Now(),
			UpdatedAt: Now(),
		}

		if err := si.entityRepo.Create(roleEntity); err != nil {
			return fmt.Errorf("failed to create role %s: %v", role.id, err)
		}
	}

	// No relationship creation needed - RBAC is tag-based
	logger.Debug("Roles created successfully - permissions handled via tags")

	return nil
}

// createRolePermissionRelationships is deprecated in v2.29.0+
// RBAC is now handled via direct tags on user entities
func (si *SecurityInitializer) createRolePermissionRelationships() error {
	// This function is no longer used in v2.29.0+
	// All role permissions are handled via direct RBAC tags on entities
	logger.Debug("Role-permission relationships are handled via tags in v2.29.0+")
	return nil
}

// createDefaultGroups creates the basic groups needed for the system
func (si *SecurityInitializer) createDefaultGroups() error {
	groups := []struct {
		id    string
		name  string
		level string
	}{
		{"group_administrators", "administrators", "organizational"},
		{"group_users", "users", "organizational"},
		{"group_guests", "guests", "organizational"},
	}

	for _, group := range groups {
		// During initial setup, skip existence check to avoid timing issues
		// The repository will handle duplicate creation gracefully
		
		groupEntity := &Entity{
			ID: group.id,
			Tags: []string{
				"type:" + EntityTypeGroup,
				"dataspace:_system",
				"name:" + group.name,
				"level:" + group.level,
				"created:" + NowString(),
			},
			Content:   nil,
			CreatedAt: Now(),
			UpdatedAt: Now(),
		}

		if err := si.entityRepo.Create(groupEntity); err != nil {
			return fmt.Errorf("failed to create group %s: %v", group.id, err)
		}
	}

	// No relationship creation needed - RBAC is tag-based
	logger.Debug("Groups created successfully - roles handled via tags")

	return nil
}

// createGroupRoleRelationships is deprecated in v2.29.0+
// RBAC is now handled via direct tags on user entities
func (si *SecurityInitializer) createGroupRoleRelationships() error {
	// This function is no longer used in v2.29.0+
	// Group-role associations are handled via direct RBAC tags on entities
	logger.Debug("Group-role relationships are handled via tags in v2.29.0+")
	return nil
}

// createDefaultAdminUser creates the default admin user if it doesn't exist
func (si *SecurityInitializer) createDefaultAdminUser() error {
	logger.Debug("Creating default admin user...")
	// Get or create admin user (with uniqueness check)
	_, err := si.getOrCreateAdminUser()
	if err != nil {
		return fmt.Errorf("failed to get or create admin user: %v", err)
	}

	// Admin user now has embedded credentials and direct RBAC tags
	// No relationships needed - everything is in the user entity
	logger.Debug("Admin user created successfully with embedded credentials and RBAC tags")

	return nil
}

// getOrCreateAdminUser safely gets existing admin user or creates one if it doesn't exist
func (si *SecurityInitializer) getOrCreateAdminUser() (*SecurityUser, error) {
	// First, check if admin user already exists
	adminEntities, err := si.entityRepo.ListByTag("identity:username:admin")
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing admin user: %v", err)
	}
	
	if len(adminEntities) > 0 {
		// User exists, return the first one
		logger.Debug("Found existing admin user: %s", adminEntities[0].ID)
		return &SecurityUser{
			ID:       adminEntities[0].ID,
			Username: "admin",
			Email:    "admin@entitydb.local",
			Status:   "active",
			Entity:   adminEntities[0],
		}, nil
	}
	
	// User doesn't exist, create it
	logger.Debug("Creating new admin user...")
	adminUser, err := si.securityManager.CreateUser("admin", "admin", "admin@entitydb.local")
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %v", err)
	}
	
	logger.Info("Created new admin user: %s", adminUser.ID)
	return adminUser, nil
}

// ensureAdminUserRelationships is deprecated in v2.29.0+
// Admin users now have embedded credentials and direct RBAC tags
func (si *SecurityInitializer) ensureAdminUserRelationships(adminUserID string) error {
	// This function is no longer used in v2.29.0+
	// Admin users are created with embedded credentials and direct RBAC tags
	logger.Debug("Admin user relationships are handled via embedded credentials and RBAC tags in v2.29.0+")
	
	// Ensure admin user has direct RBAC tags (this is still needed)
	rbacTagManager := NewRBACTagManager(si.entityRepo)
	if err := rbacTagManager.AssignRoleToUser(adminUserID, "admin"); err != nil {
		return fmt.Errorf("failed to assign admin role tag: %v", err)
	}

	return nil
}

// MigrateExistingUsers migrates existing users to the new security model
func (si *SecurityInitializer) MigrateExistingUsers() error {
	// Get all existing user entities
	userEntities, err := si.entityRepo.ListByTag("type:user")
	if err != nil {
		return fmt.Errorf("failed to get existing users: %v", err)
	}

	for _, userEntity := range userEntities {
		// Skip if already migrated (has identity:uuid tag)
		userTags := userEntity.GetTagsWithoutTimestamp()
		hasUUID := false
		for _, tag := range userTags {
			if strings.HasPrefix(tag, "identity:uuid:") {
				hasUUID = true
				break
			}
		}

		if hasUUID {
			continue // Already migrated
		}

		// Migrate this user
		if err := si.migrateUser(userEntity); err != nil {
			return fmt.Errorf("failed to migrate user %s: %v", userEntity.ID, err)
		}
	}

	return nil
}

// migrateUser migrates a single user to the new security model
func (si *SecurityInitializer) migrateUser(oldUserEntity *Entity) error {
	// Extract username from tags or content
	userTags := oldUserEntity.GetTagsWithoutTimestamp()
	var username string

	for _, tag := range userTags {
		if strings.HasPrefix(tag, "id:username:") {
			username = strings.TrimPrefix(tag, "id:username:")
			break
		} else if strings.HasPrefix(tag, "identity:username:") {
			username = strings.TrimPrefix(tag, "identity:username:")
			break
		}
	}

	if username == "" {
		return fmt.Errorf("could not extract username from user entity %s", oldUserEntity.ID)
	}

	// Check if this user already exists in new format
	newUserEntities, err := si.entityRepo.ListByTag("identity:username:" + username)
	if err == nil && len(newUserEntities) > 0 {
		return nil // Already migrated
	}

	// Extract password hash and other details from old entity content
	// This is a complex migration that depends on the old format
	// For now, we'll require manual migration or recreation of users

	return fmt.Errorf("automatic migration not implemented - please recreate user %s", username)
}

// forceSync forces the repository to sync and make entities immediately available
func (si *SecurityInitializer) forceSync() error {
	// If the repository has a sync method, call it
	if syncable, ok := si.entityRepo.(interface{ Sync() error }); ok {
		return syncable.Sync()
	}
	
	// Alternative: try to force a flush/checkpoint if available
	if flushable, ok := si.entityRepo.(interface{ Flush() error }); ok {
		if err := flushable.Flush(); err != nil {
			return fmt.Errorf("failed to flush repository: %v", err)
		}
	}
	
	if checkpointable, ok := si.entityRepo.(interface{ Checkpoint() error }); ok {
		if err := checkpointable.Checkpoint(); err != nil {
			return fmt.Errorf("failed to checkpoint repository: %v", err)
		}
	}
	
	return nil
}

// createDefaultDataspaces creates the system and default dataspaces
func (si *SecurityInitializer) createDefaultDataspaces() error {
	logger.Info("Creating default dataspaces...")
	
	// Create system dataspace for system entities (users, permissions, etc.)
	systemDataspace := &Entity{
		ID: "dataspace_system",
		Tags: []string{
			"type:dataspace",
			"dataspace:_system",
			"name:_system",
			"description:System dataspace for internal entities",
			"system:true",
			"created:" + NowString(),
		},
		Content:   nil,
		CreatedAt: Now(),
		UpdatedAt: Now(),
	}
	
	if err := si.entityRepo.Create(systemDataspace); err != nil {
		// Check if already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create system dataspace: %v", err)
		}
		logger.Debug("System dataspace already exists")
	} else {
		logger.Info("Created system dataspace: _system")
	}
	
	// Create default dataspace for user data
	defaultDataspace := &Entity{
		ID: "dataspace_default",
		Tags: []string{
			"type:dataspace",
			"dataspace:default",
			"name:default",
			"description:Default dataspace for user entities",
			"system:false",
			"created:" + NowString(),
		},
		Content:   nil,
		CreatedAt: Now(),
		UpdatedAt: Now(),
	}
	
	if err := si.entityRepo.Create(defaultDataspace); err != nil {
		// Check if already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create default dataspace: %v", err)
		}
		logger.Debug("Default dataspace already exists")
	} else {
		logger.Info("Created default dataspace: default")
	}
	
	return nil
}