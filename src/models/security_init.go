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

	// Create role-permission relationships
	if err := si.createRolePermissionRelationships(); err != nil {
		return fmt.Errorf("failed to create role-permission relationships: %v", err)
	}

	return nil
}

// createRolePermissionRelationships creates the relationships between roles and permissions
func (si *SecurityInitializer) createRolePermissionRelationships() error {
	// Admin role gets all permissions
	adminPermissions := []string{
		"perm_all", // Global wildcard permission
	}

	for _, permID := range adminPermissions {
		relationship := &EntityRelationship{
			ID:       "rel_role_admin_grants_" + permID,
			SourceID: "role_admin",
			TargetID: permID,
			Type:     RelationshipGrants,
			Properties: map[string]string{
				"inherited": "false",
				"scope":     "global",
			},
			CreatedAt: Now(),
		}

		if err := si.entityRepo.CreateRelationship(relationship); err != nil {
			return fmt.Errorf("failed to create admin-permission relationship: %v", err)
		}
	}

	// User role gets basic permissions
	userPermissions := []string{
		"perm_entity_view",
		"perm_entity_create",
		"perm_entity_update",
		"perm_entity_delete",
		"perm_relation_view",
		"perm_relation_create",
		"perm_system_view",
	}

	for _, permID := range userPermissions {
		relationship := &EntityRelationship{
			ID:       "rel_role_user_grants_" + permID,
			SourceID: "role_user",
			TargetID: permID,
			Type:     RelationshipGrants,
			Properties: map[string]string{
				"inherited": "false",
				"scope":     "user",
			},
			CreatedAt: Now(),
		}

		if err := si.entityRepo.CreateRelationship(relationship); err != nil {
			return fmt.Errorf("failed to create user-permission relationship: %v", err)
		}
	}

	// Guest role gets minimal permissions
	guestPermissions := []string{
		"perm_entity_view",
		"perm_system_view",
	}

	for _, permID := range guestPermissions {
		relationship := &EntityRelationship{
			ID:       "rel_role_guest_grants_" + permID,
			SourceID: "role_guest",
			TargetID: permID,
			Type:     RelationshipGrants,
			Properties: map[string]string{
				"inherited": "false",
				"scope":     "read_only",
			},
			CreatedAt: Now(),
		}

		if err := si.entityRepo.CreateRelationship(relationship); err != nil {
			return fmt.Errorf("failed to create guest-permission relationship: %v", err)
		}
	}

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

	// Create group-role relationships
	if err := si.createGroupRoleRelationships(); err != nil {
		return fmt.Errorf("failed to create group-role relationships: %v", err)
	}

	return nil
}

// createGroupRoleRelationships creates the relationships between groups and roles
func (si *SecurityInitializer) createGroupRoleRelationships() error {
	groupRoleMap := map[string]string{
		"group_administrators": "role_admin",
		"group_users":          "role_user",
		"group_guests":         "role_guest",
	}

	for groupID, roleID := range groupRoleMap {
		relationship := &EntityRelationship{
			ID:       "rel_" + groupID + "_has_" + roleID,
			SourceID: groupID,
			TargetID: roleID,
			Type:     RelationshipHasRole,
			Properties: map[string]string{
				"inherited": "true",
				"default":   "true",
			},
			CreatedAt: Now(),
		}

		if err := si.entityRepo.CreateRelationship(relationship); err != nil {
			return fmt.Errorf("failed to create group-role relationship: %v", err)
		}
	}

	return nil
}

// createDefaultAdminUser creates the default admin user if it doesn't exist
func (si *SecurityInitializer) createDefaultAdminUser() error {
	logger.Debug("[SecurityInit] Creating default admin user...")
	// During initial setup, always try to create admin user
	// The security manager will handle if it already exists
	adminUser, err := si.securityManager.CreateUser("admin", "admin", "admin@entitydb.local")
	if err != nil {
		logger.Debug("[SecurityInit] CreateUser failed: %v", err)
		// If user already exists, that's okay, try to get it
		adminEntities, listErr := si.entityRepo.ListByTag("identity:username:admin")
		if listErr != nil || len(adminEntities) == 0 {
			logger.Error("[SecurityInit] Failed to create admin user and couldn't find existing one: createErr=%v, listErr=%v", err, listErr)
			return fmt.Errorf("failed to create admin user and couldn't find existing one: %v", err)
		}
		logger.Debug("[SecurityInit] Found existing admin user: %s", adminEntities[0].ID)
		// Convert Entity to SecurityUser-like structure for relationship creation
		adminUser = &SecurityUser{
			ID:       adminEntities[0].ID,
			Username: "admin",
		}
	}

	// Add admin to administrators group
	adminGroupRelationship := &EntityRelationship{
		ID:       "rel_" + adminUser.ID + "_member_of_group_administrators",
		SourceID: adminUser.ID,
		TargetID: "group_administrators",
		Type:     RelationshipMemberOf,
		Properties: map[string]string{
			"role_in_group": "admin",
			"primary":       "true",
		},
		CreatedAt: Now(),
	}

	if err := si.entityRepo.CreateRelationship(adminGroupRelationship); err != nil {
		return fmt.Errorf("failed to create admin-group relationship: %v", err)
	}

	// Give admin user direct admin role as well
	adminRoleRelationship := &EntityRelationship{
		ID:       "rel_" + adminUser.ID + "_has_role_admin",
		SourceID: adminUser.ID,
		TargetID: "role_admin",
		Type:     RelationshipHasRole,
		Properties: map[string]string{
			"direct":    "true",
			"inherited": "false",
		},
		CreatedAt: Now(),
	}

	if err := si.entityRepo.CreateRelationship(adminRoleRelationship); err != nil {
		return fmt.Errorf("failed to create admin-role relationship: %v", err)
	}

	return nil
}

// ensureAdminUserRelationships ensures existing admin user has proper relationships
func (si *SecurityInitializer) ensureAdminUserRelationships(adminUserID string) error {
	// Check if admin is member of administrators group
	adminGroupRels, err := si.entityRepo.GetRelationshipsBySource(adminUserID)
	if err != nil {
		return err
	}

	hasMemberOfAdminGroup := false
	hasAdminRole := false

	for _, rel := range adminGroupRels {
		if relationship, ok := rel.(*EntityRelationship); ok {
			if relationship.Type == RelationshipMemberOf && relationship.TargetID == "group_administrators" {
				hasMemberOfAdminGroup = true
			}
			if relationship.Type == RelationshipHasRole && relationship.TargetID == "role_admin" {
				hasAdminRole = true
			}
		}
	}

	// Add missing relationships
	if !hasMemberOfAdminGroup {
		adminGroupRelationship := &EntityRelationship{
			ID:       "rel_" + adminUserID + "_member_of_group_administrators",
			SourceID: adminUserID,
			TargetID: "group_administrators",
			Type:     RelationshipMemberOf,
			Properties: map[string]string{
				"role_in_group": "admin",
				"primary":       "true",
			},
			CreatedAt: Now(),
		}

		if err := si.entityRepo.CreateRelationship(adminGroupRelationship); err != nil {
			return fmt.Errorf("failed to create admin-group relationship: %v", err)
		}
	}

	if !hasAdminRole {
		adminRoleRelationship := &EntityRelationship{
			ID:       "rel_" + adminUserID + "_has_role_admin",
			SourceID: adminUserID,
			TargetID: "role_admin",
			Type:     RelationshipHasRole,
			Properties: map[string]string{
				"direct":    "true",
				"inherited": "false",
			},
			CreatedAt: Now(),
		}

		if err := si.entityRepo.CreateRelationship(adminRoleRelationship); err != nil {
			return fmt.Errorf("failed to create admin-role relationship: %v", err)
		}
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