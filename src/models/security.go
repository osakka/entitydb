package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"golang.org/x/crypto/bcrypt"
	
	"entitydb/logger"
)

// SecurityManager handles all relationship-based security operations
type SecurityManager struct {
	entityRepo EntityRepository
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(entityRepo EntityRepository) *SecurityManager {
	return &SecurityManager{
		entityRepo: entityRepo,
	}
}

// SecurityEntity types for type safety
const (
	EntityTypeUser       = "user"
	EntityTypeCredential = "credential"
	EntityTypeSession    = "session"
	EntityTypeRole       = "role"
	EntityTypePermission = "permission"
	EntityTypeGroup      = "group"
)

// Relationship types for security graph
const (
	RelationshipHasCredential   = "has_credential"
	RelationshipAuthenticatedAs = "authenticated_as"
	RelationshipMemberOf        = "member_of"
	RelationshipHasRole         = "has_role"
	RelationshipGrants          = "grants"
	RelationshipCanAccess       = "can_access"      // User/Role can access Dataspace
	RelationshipOwns            = "owns"            // User owns Dataspace
	RelationshipBelongsTo       = "belongs_to"      // Entity belongs to Dataspace
	RelationshipDelegates       = "delegates"       // Role delegates to another Role in Dataspace
)

// SecurityUser represents a user in the security system
type SecurityUser struct {
	ID       string
	Username string
	Email    string
	Status   string
	Entity   *Entity
}

// SecuritySession represents an active session
type SecuritySession struct {
	ID        string
	Token     string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
	IPAddress string
	UserAgent string
	Entity    *Entity
}

// SecurityRole represents a role in the system
type SecurityRole struct {
	ID          string
	Name        string
	Level       int
	Scope       string
	Permissions []string
	Entity      *Entity
}

// SecurityPermission represents an atomic permission
type SecurityPermission struct {
	ID       string
	Resource string
	Action   string
	Scope    string
	Entity   *Entity
}

// CreateUser creates a new user entity with separate credential entity
func (sm *SecurityManager) CreateUser(username, password, email string) (*SecurityUser, error) {
	logger.Debug("CreateUser called for username: %s", username)
	
	// Check if user already exists
	existingUsers, err := sm.entityRepo.ListByTag("identity:username:" + username)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing user: %v", err)
	}
	
	if len(existingUsers) > 0 {
		logger.Debug("User %s already exists with ID: %s", username, existingUsers[0].ID)
		return nil, fmt.Errorf("user with username '%s' already exists", username)
	}
	
	// Generate secure UUID for user
	userID := "user_" + generateSecureUUID()
	logger.Debug("Generated user ID: %s", userID)
	
	// Create user entity (no sensitive data)
	tags := []string{
		"type:" + EntityTypeUser,
		"dataspace:_system",
		"identity:username:" + username,
		"identity:uuid:" + userID,
		"status:active",
		"profile:email:" + email,
		"created:" + NowString(),
	}
	
	// Add rbac:role:admin tag directly for admin user
	if username == "admin" {
		tags = append(tags, "rbac:role:admin")
	}
	
	userEntity := &Entity{
		ID:        userID,
		Tags:      tags,
		Content:   nil, // No content for user entities
		CreatedAt: Now(),
		UpdatedAt: Now(),
	}
	
	// Create user entity
	if err := sm.entityRepo.Create(userEntity); err != nil {
		logger.Error("Failed to create user entity: %v", err)
		return nil, fmt.Errorf("failed to create user entity: %v", err)
	}
	logger.Debug("Successfully created user entity")
	
	// Create separate credential entity
	credentialID := "cred_" + generateSecureUUID()
	salt := generateSalt()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	
	credentialEntity := &Entity{
		ID: credentialID,
		Tags: []string{
			"type:" + EntityTypeCredential,
			"dataspace:_system",
			"algorithm:bcrypt",
			"user:" + userID,
			"salt:" + salt,
			"created:" + NowString(),
		},
		Content:   hashedPassword,
		CreatedAt: Now(),
		UpdatedAt: Now(),
	}
	
	// Create credential entity
	if err := sm.entityRepo.Create(credentialEntity); err != nil {
		return nil, fmt.Errorf("failed to create credential entity: %v", err)
	}
	
	// Create relationship between user and credential
	relationship := &EntityRelationship{
		ID:               "rel_" + generateSecureUUID(),
		SourceID:         userID,
		TargetID:         credentialID,
		Type:             RelationshipHasCredential,
		RelationshipType: RelationshipHasCredential,
		Properties:       map[string]string{"primary": "true"},
		CreatedAt:        Now(),
	}
	
	if err := sm.entityRepo.CreateRelationship(relationship); err != nil {
		return nil, fmt.Errorf("failed to create user-credential relationship: %v", err)
	}
	
	return &SecurityUser{
		ID:       userID,
		Username: username,
		Email:    email,
		Status:   "active",
		Entity:   userEntity,
	}, nil
}

// AuthenticateUser performs relationship-based authentication
func (sm *SecurityManager) AuthenticateUser(username, password string) (*SecurityUser, error) {
	// Find user by username tag
	logger.Debug("Looking for user with tag: identity:username:%s", username)
	userEntities, err := sm.entityRepo.ListByTag("identity:username:" + username)
	if err != nil {
		logger.Error("Error finding user: %v", err)
		return nil, fmt.Errorf("user not found: %v", err)
	}
	if len(userEntities) == 0 {
		logger.Debug("No user entities found with username: %s", username)
		return nil, fmt.Errorf("user not found")
	}
	logger.Debug("Found %d user entities for username: %s", len(userEntities), username)
	
	userEntity := userEntities[0]
	
	// Check if user is active
	userTags := userEntity.GetTagsWithoutTimestamp()
	isActive := false
	for _, tag := range userTags {
		if tag == "status:active" {
			isActive = true
			break
		}
	}
	
	if !isActive {
		return nil, fmt.Errorf("user account is not active")
	}
	
	// Get credential entity via relationship
	logger.Debug("Getting relationships for user ID: %s", userEntity.ID)
	credentialEntities, err := sm.entityRepo.GetRelationshipsBySource(userEntity.ID)
	if err != nil {
		logger.Error("Failed to get relationships for user %s: %v", userEntity.ID, err)
		return nil, fmt.Errorf("failed to get user credentials: %v", err)
	}
	logger.Debug("Found %d relationships for user %s", len(credentialEntities), userEntity.ID)
	
	var credentialEntity *Entity
	for _, rel := range credentialEntities {
		if relationship, ok := rel.(*EntityRelationship); ok {
			logger.Debug("Checking relationship %s of type %s/%s", relationship.ID, relationship.Type, relationship.RelationshipType)
			if relationship.Type == RelationshipHasCredential || relationship.RelationshipType == RelationshipHasCredential {
				logger.Debug("Found has_credential relationship, fetching credential entity %s", relationship.TargetID)
				credEntity, err := sm.entityRepo.GetByID(relationship.TargetID)
				if err == nil {
					credentialEntity = credEntity
					logger.Debug("Successfully fetched credential entity %s", relationship.TargetID)
					break
				} else {
					logger.Error("Failed to fetch credential entity %s: %v", relationship.TargetID, err)
				}
			}
		}
	}
	
	if credentialEntity == nil {
		return nil, fmt.Errorf("no credentials found for user")
	}
	
	// Extract salt from credential tags
	credTags := credentialEntity.GetTagsWithoutTimestamp()
	var salt string
	for _, tag := range credTags {
		if strings.HasPrefix(tag, "salt:") {
			salt = strings.TrimPrefix(tag, "salt:")
			break
		}
	}
	
	// Verify password
	err = bcrypt.CompareHashAndPassword(credentialEntity.Content, []byte(password+salt))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}
	
	// Extract user details
	var email string
	for _, tag := range userTags {
		if strings.HasPrefix(tag, "profile:email:") {
			email = strings.TrimPrefix(tag, "profile:email:")
			break
		}
	}
	
	return &SecurityUser{
		ID:       userEntity.ID,
		Username: username,
		Email:    email,
		Status:   "active",
		Entity:   userEntity,
	}, nil
}

// CreateSession creates a new session entity linked to user
func (sm *SecurityManager) CreateSession(user *SecurityUser, ipAddress, userAgent string) (*SecuritySession, error) {
	sessionID := "session_" + generateSecureUUID()
	token := generateSecureToken()
	expiresAt := time.Now().Add(2 * time.Hour) // 2 hour sessions
	
	sessionEntity := &Entity{
		ID: sessionID,
		Tags: []string{
			"type:" + EntityTypeSession,
			"dataspace:_system",
			"token:" + token,
			"expires:" + expiresAt.Format(time.RFC3339),
			"ip:" + ipAddress,
			"user_agent:" + userAgent,
			"created:" + NowString(),
		},
		Content:   nil,
		CreatedAt: Now(),
		UpdatedAt: Now(),
	}
	
	// Create session entity
	if err := sm.entityRepo.Create(sessionEntity); err != nil {
		return nil, fmt.Errorf("failed to create session entity: %v", err)
	}
	
	// Create relationship between session and user
	relationship := &EntityRelationship{
		ID:         "rel_" + generateSecureUUID(),
		SourceID:   sessionID,
		TargetID:   user.ID,
		Type:       RelationshipAuthenticatedAs,
		Properties: map[string]string{"active": "true"},
		CreatedAt:  Now(),
	}
	
	if err := sm.entityRepo.CreateRelationship(relationship); err != nil {
		return nil, fmt.Errorf("failed to create session-user relationship: %v", err)
	}
	
	return &SecuritySession{
		ID:        sessionID,
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Entity:    sessionEntity,
	}, nil
}

// ValidateSession validates a session token and returns the associated user
func (sm *SecurityManager) ValidateSession(token string) (*SecurityUser, error) {
	// Find session by token tag
	sessionEntities, err := sm.entityRepo.ListByTag("token:" + token)
	if err != nil || len(sessionEntities) == 0 {
		return nil, fmt.Errorf("session not found")
	}
	
	sessionEntity := sessionEntities[0]
	
	// Check if session is expired
	sessionTags := sessionEntity.GetTagsWithoutTimestamp()
	var expiresAt time.Time
	for _, tag := range sessionTags {
		if strings.HasPrefix(tag, "expires:") {
			expirationStr := strings.TrimPrefix(tag, "expires:")
			expiresAt, err = time.Parse(time.RFC3339, expirationStr)
			if err != nil {
				return nil, fmt.Errorf("invalid session expiration format")
			}
			break
		}
	}
	
	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("session expired")
	}
	
	// Get user via relationship
	userRelationships, err := sm.entityRepo.GetRelationshipsBySource(sessionEntity.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session relationships: %v", err)
	}
	
	var userEntity *Entity
	for _, rel := range userRelationships {
		if relationship, ok := rel.(*EntityRelationship); ok {
			if relationship.Type == RelationshipAuthenticatedAs {
				userEnt, err := sm.entityRepo.GetByID(relationship.TargetID)
				if err == nil {
					userEntity = userEnt
					break
				}
			}
		}
	}
	
	if userEntity == nil {
		return nil, fmt.Errorf("no user found for session")
	}
	
	// Extract user details from tags
	userTags := userEntity.GetTagsWithoutTimestamp()
	var username, email string
	for _, tag := range userTags {
		if strings.HasPrefix(tag, "identity:username:") {
			username = strings.TrimPrefix(tag, "identity:username:")
		} else if strings.HasPrefix(tag, "profile:email:") {
			email = strings.TrimPrefix(tag, "profile:email:")
		}
	}
	
	return &SecurityUser{
		ID:       userEntity.ID,
		Username: username,
		Email:    email,
		Status:   "active",
		Entity:   userEntity,
	}, nil
}

// HasPermission checks if a user has a specific permission via relationship traversal
func (sm *SecurityManager) HasPermission(user *SecurityUser, resource, action string) (bool, error) {
	return sm.HasPermissionInDataspace(user, resource, action, "")
}

// HasPermissionInDataspace checks if a user has a specific permission in a dataspace via relationship traversal
func (sm *SecurityManager) HasPermissionInDataspace(user *SecurityUser, resource, action, dataspaceID string) (bool, error) {
	// First check direct user permissions via user->role->permission
	userRoles, err := sm.getUserRoles(user.ID)
	if err != nil {
		return false, err
	}
	
	for _, role := range userRoles {
		hasPermission, err := sm.roleHasPermissionInDataspace(role.ID, resource, action, dataspaceID)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}
	
	// Check group-based permissions via user->group->role->permission
	userGroups, err := sm.getUserGroups(user.ID)
	if err != nil {
		return false, err
	}
	
	for _, group := range userGroups {
		groupRoles, err := sm.getGroupRoles(group.ID)
		if err != nil {
			continue // Skip this group if we can't get its roles
		}
		
		for _, role := range groupRoles {
			hasPermission, err := sm.roleHasPermissionInDataspace(role.ID, resource, action, dataspaceID)
			if err != nil {
				continue // Skip this role if we can't check its permissions
			}
			if hasPermission {
				return true, nil
			}
		}
	}
	
	return false, nil
}

// Helper functions

func (sm *SecurityManager) getUserRoles(userID string) ([]*SecurityRole, error) {
	relationships, err := sm.entityRepo.GetRelationshipsBySource(userID)
	if err != nil {
		return nil, err
	}
	
	var roles []*SecurityRole
	for _, rel := range relationships {
		if relationship, ok := rel.(*EntityRelationship); ok {
			if relationship.Type == RelationshipHasRole {
				roleEntity, err := sm.entityRepo.GetByID(relationship.TargetID)
				if err == nil {
					role := &SecurityRole{
						ID:     roleEntity.ID,
						Entity: roleEntity,
					}
					// Extract role name from tags
					roleTags := roleEntity.GetTagsWithoutTimestamp()
					for _, tag := range roleTags {
						if strings.HasPrefix(tag, "name:") {
							role.Name = strings.TrimPrefix(tag, "name:")
							break
						}
					}
					roles = append(roles, role)
				}
			}
		}
	}
	
	return roles, nil
}

func (sm *SecurityManager) getUserGroups(userID string) ([]*Entity, error) {
	relationships, err := sm.entityRepo.GetRelationshipsBySource(userID)
	if err != nil {
		return nil, err
	}
	
	var groups []*Entity
	for _, rel := range relationships {
		if relationship, ok := rel.(*EntityRelationship); ok {
			if relationship.Type == RelationshipMemberOf {
				groupEntity, err := sm.entityRepo.GetByID(relationship.TargetID)
				if err == nil {
					groups = append(groups, groupEntity)
				}
			}
		}
	}
	
	return groups, nil
}

func (sm *SecurityManager) getGroupRoles(groupID string) ([]*SecurityRole, error) {
	relationships, err := sm.entityRepo.GetRelationshipsBySource(groupID)
	if err != nil {
		return nil, err
	}
	
	var roles []*SecurityRole
	for _, rel := range relationships {
		if relationship, ok := rel.(*EntityRelationship); ok {
			if relationship.Type == RelationshipHasRole {
				roleEntity, err := sm.entityRepo.GetByID(relationship.TargetID)
				if err == nil {
					role := &SecurityRole{
						ID:     roleEntity.ID,
						Entity: roleEntity,
					}
					roles = append(roles, role)
				}
			}
		}
	}
	
	return roles, nil
}

func (sm *SecurityManager) roleHasPermission(roleID, resource, action string) (bool, error) {
	relationships, err := sm.entityRepo.GetRelationshipsBySource(roleID)
	if err != nil {
		return false, err
	}
	
	for _, rel := range relationships {
		if relationship, ok := rel.(*EntityRelationship); ok {
			if relationship.Type == RelationshipGrants {
				permissionEntity, err := sm.entityRepo.GetByID(relationship.TargetID)
				if err == nil {
					permTags := permissionEntity.GetTagsWithoutTimestamp()
					var permResource, permAction string
					for _, tag := range permTags {
						if strings.HasPrefix(tag, "resource:") {
							permResource = strings.TrimPrefix(tag, "resource:")
						} else if strings.HasPrefix(tag, "action:") {
							permAction = strings.TrimPrefix(tag, "action:")
						}
					}
					
					// Check for exact match or wildcard permissions
					if (permResource == resource || permResource == "*") &&
						(permAction == action || permAction == "*") {
						return true, nil
					}
				}
			}
		}
	}
	
	return false, nil
}

func (sm *SecurityManager) roleHasPermissionInDataspace(roleID, resource, action, dataspaceID string) (bool, error) {
	// If no dataspace specified, check global permissions
	if dataspaceID == "" {
		return sm.roleHasPermission(roleID, resource, action)
	}
	
	// First check if user has global admin permissions (overrides dataspace restrictions)
	hasGlobal, err := sm.roleHasPermission(roleID, "*", "*")
	if err != nil {
		return false, err
	}
	if hasGlobal {
		return true, nil
	}
	
	// Check if role has access to the specific dataspace
	hasDataspaceAccess, err := sm.roleHasDataspaceAccess(roleID, dataspaceID)
	if err != nil {
		return false, err
	}
	if !hasDataspaceAccess {
		return false, nil // No access to dataspace at all
	}
	
	// Check for dataspace-scoped permissions
	relationships, err := sm.entityRepo.GetRelationshipsBySource(roleID)
	if err != nil {
		return false, err
	}
	
	for _, rel := range relationships {
		if relationship, ok := rel.(*EntityRelationship); ok {
			if relationship.Type == RelationshipGrants {
				permissionEntity, err := sm.entityRepo.GetByID(relationship.TargetID)
				if err == nil {
					permTags := permissionEntity.GetTagsWithoutTimestamp()
					var permResource, permAction, permDataspace string
					for _, tag := range permTags {
						if strings.HasPrefix(tag, "resource:") {
							permResource = strings.TrimPrefix(tag, "resource:")
						} else if strings.HasPrefix(tag, "action:") {
							permAction = strings.TrimPrefix(tag, "action:")
						} else if strings.HasPrefix(tag, "dataspace:") {
							permDataspace = strings.TrimPrefix(tag, "dataspace:")
						}
					}
					
					// Check for exact match or wildcard permissions in the right dataspace
					if (permResource == resource || permResource == "*") &&
						(permAction == action || permAction == "*") &&
						(permDataspace == dataspaceID || permDataspace == "*") {
						return true, nil
					}
				}
			}
		}
	}
	
	return false, nil
}

func (sm *SecurityManager) roleHasDataspaceAccess(roleID, dataspaceID string) (bool, error) {
	relationships, err := sm.entityRepo.GetRelationshipsBySource(roleID)
	if err != nil {
		return false, err
	}
	
	for _, rel := range relationships {
		if relationship, ok := rel.(*EntityRelationship); ok {
			if relationship.Type == RelationshipCanAccess && relationship.TargetID == dataspaceID {
				return true, nil
			}
		}
	}
	
	return false, nil
}

// CanAccessDataspace checks if a user can access a specific dataspace
func (sm *SecurityManager) CanAccessDataspace(user *SecurityUser, dataspaceID string) (bool, error) {
	// Check direct user access
	userRoles, err := sm.getUserRoles(user.ID)
	if err != nil {
		return false, err
	}
	
	for _, role := range userRoles {
		hasAccess, err := sm.roleHasDataspaceAccess(role.ID, dataspaceID)
		if err != nil {
			return false, err
		}
		if hasAccess {
			return true, nil
		}
	}
	
	// Check group-based access
	userGroups, err := sm.getUserGroups(user.ID)
	if err != nil {
		return false, err
	}
	
	for _, group := range userGroups {
		groupRoles, err := sm.getGroupRoles(group.ID)
		if err != nil {
			continue
		}
		
		for _, role := range groupRoles {
			hasAccess, err := sm.roleHasDataspaceAccess(role.ID, dataspaceID)
			if err != nil {
				continue
			}
			if hasAccess {
				return true, nil
			}
		}
	}
	
	return false, nil
}

// Utility functions

func generateSecureUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateSalt() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}