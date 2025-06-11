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

// GetEntityRepo returns the entity repository
func (sm *SecurityManager) GetEntityRepo() EntityRepository {
	return sm.entityRepo
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
	RelationshipCanAccess       = "can_access"      // User/Role can access Dataset
	RelationshipOwns            = "owns"            // User owns Dataset
	RelationshipBelongsTo       = "belongs_to"      // Entity belongs to Dataset
	RelationshipDelegates       = "delegates"       // Role delegates to another Role in Dataset
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
	logger.TraceIf("auth", "creating user for username: %s", username)
	
	// Check if user already exists
	existingUsers, err := sm.entityRepo.ListByTag("identity:username:" + username)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing user: %v", err)
	}
	
	if len(existingUsers) > 0 {
		logger.TraceIf("auth", "user %s already exists with id: %s", username, existingUsers[0].ID)
		return nil, fmt.Errorf("user with username '%s' already exists", username)
	}
	
	// Generate secure UUID for user
	userID := "user_" + generateSecureUUID()
	logger.TraceIf("auth", "generated user id: %s", userID)
	
	// Generate password hash and salt
	salt := generateSalt()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	
	// Create user entity with credentials in content
	tags := []string{
		"type:" + EntityTypeUser,
		"dataset:_system",
		"identity:username:" + username,
		"identity:uuid:" + userID,
		"status:active",
		"profile:email:" + email,
		"created:" + NowString(),
		"has:credentials", // Tag to indicate this user has embedded credentials
	}
	
	// Add rbac:role:admin tag directly for admin user
	if username == "admin" {
		tags = append(tags, "rbac:role:admin")
	}
	
	// Store credentials in content as a simple format: salt|hash
	// This keeps it simple and efficient
	credentialContent := fmt.Sprintf("%s|%s", salt, string(hashedPassword))
	
	userEntity := &Entity{
		ID:        userID,
		Tags:      tags,
		Content:   []byte(credentialContent),
		CreatedAt: Now(),
		UpdatedAt: Now(),
	}
	
	// Create user entity with embedded credentials
	if err := sm.entityRepo.Create(userEntity); err != nil {
		logger.Error("failed to create user entity: %v", err)
		return nil, fmt.Errorf("failed to create user entity: %v", err)
	}
	logger.TraceIf("auth", "successfully created user entity with embedded credentials")
	
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
	logger.TraceIf("auth", "looking for user with tag: identity:username:%s", username)
	logger.TraceIf("auth", "about to call ListByTag for user lookup")
	userEntities, err := sm.entityRepo.ListByTag("identity:username:" + username)
	logger.TraceIf("auth", "ListByTag returned %d entities", len(userEntities))
	if err != nil {
		logger.Error("error finding user: %v", err)
		return nil, fmt.Errorf("user not found: %v", err)
	}
	if len(userEntities) == 0 {
		logger.TraceIf("auth", "no user entities found with username: %s", username)
		return nil, fmt.Errorf("user not found")
	}
	logger.TraceIf("auth", "found %d user entities for username: %s", len(userEntities), username)
	
	userEntity := userEntities[0]
	
	// Check if user is active and has credentials
	userTags := userEntity.GetTagsWithoutTimestamp()
	isActive := false
	hasCredentials := false
	for _, tag := range userTags {
		if tag == "status:active" {
			isActive = true
		}
		if tag == "has:credentials" {
			hasCredentials = true
		}
	}
	
	if !isActive {
		return nil, fmt.Errorf("user account is not active")
	}
	
	if !hasCredentials {
		logger.TraceIf("auth", "user does not have embedded credentials")
		return nil, fmt.Errorf("no credentials found for user")
	}
	
	// Extract credentials from user entity content
	logger.TraceIf("auth", "extracting credentials from user entity content")
	if len(userEntity.Content) == 0 {
		logger.Error("user entity has no content for user %s", username)
		return nil, fmt.Errorf("invalid credentials")
	}
	
	// Parse content format: salt|hash
	credentialParts := strings.SplitN(string(userEntity.Content), "|", 2)
	if len(credentialParts) != 2 {
		logger.Error("invalid credential format in user entity for user %s", username)
		return nil, fmt.Errorf("invalid credential format")
	}
	
	salt := credentialParts[0]
	hashedPassword := []byte(credentialParts[1])
	
	// Verify password
	logger.TraceIf("auth", "starting password verification for user %s", username)
	logger.TraceIf("auth", "credential hash length: %d bytes", len(hashedPassword))
	logger.TraceIf("auth", "salt: %s", salt)
	logger.TraceIf("auth", "password+salt length: %d", len(password+salt))
	
	// Check if content looks like a bcrypt hash (should start with $2a$, $2b$, or $2y$)
	if len(hashedPassword) < 4 || hashedPassword[0] != '$' {
		logger.Error("credential content does not appear to be a bcrypt hash for user %s (first bytes: %v)", 
			username, hashedPassword[:min(4, len(hashedPassword))])
		return nil, fmt.Errorf("invalid credential format")
	}
	
	// Use a goroutine with timeout to prevent indefinite hang
	type bcryptResult struct {
		err error
	}
	resultChan := make(chan bcryptResult, 1)
	
	go func() {
		startTime := time.Now()
		err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password+salt))
		elapsed := time.Since(startTime)
		logger.TraceIf("auth", "password verification completed in %v", elapsed)
		resultChan <- bcryptResult{err: err}
	}()
	
	// Wait for bcrypt result with timeout
	select {
	case result := <-resultChan:
		if result.err != nil {
			logger.TraceIf("auth", "password verification failed: %v", result.err)
			return nil, fmt.Errorf("invalid password")
		}
		logger.TraceIf("auth", "password verification successful")
	case <-time.After(5 * time.Second):
		logger.Error("password verification timed out after 5 seconds for user %s", username)
		return nil, fmt.Errorf("authentication timeout")
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
			"dataset:_system",
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
	return sm.HasPermissionInDataset(user, resource, action, "")
}

// HasPermissionInDataset checks if a user has a specific permission in a dataset via relationship traversal
func (sm *SecurityManager) HasPermissionInDataset(user *SecurityUser, resource, action, datasetID string) (bool, error) {
	// First check direct user permissions via user->role->permission
	userRoles, err := sm.getUserRoles(user.ID)
	if err != nil {
		return false, err
	}
	
	for _, role := range userRoles {
		hasPermission, err := sm.roleHasPermissionInDataset(role.ID, resource, action, datasetID)
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
			hasPermission, err := sm.roleHasPermissionInDataset(role.ID, resource, action, datasetID)
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

func (sm *SecurityManager) roleHasPermissionInDataset(roleID, resource, action, datasetID string) (bool, error) {
	// If no dataset specified, check global permissions
	if datasetID == "" {
		return sm.roleHasPermission(roleID, resource, action)
	}
	
	// First check if user has global admin permissions (overrides dataset restrictions)
	hasGlobal, err := sm.roleHasPermission(roleID, "*", "*")
	if err != nil {
		return false, err
	}
	if hasGlobal {
		return true, nil
	}
	
	// Check if role has access to the specific dataset
	hasDatasetAccess, err := sm.roleHasDatasetAccess(roleID, datasetID)
	if err != nil {
		return false, err
	}
	if !hasDatasetAccess {
		return false, nil // No access to dataset at all
	}
	
	// Check for dataset-scoped permissions
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
					var permResource, permAction, permDataset string
					for _, tag := range permTags {
						if strings.HasPrefix(tag, "resource:") {
							permResource = strings.TrimPrefix(tag, "resource:")
						} else if strings.HasPrefix(tag, "action:") {
							permAction = strings.TrimPrefix(tag, "action:")
						} else if strings.HasPrefix(tag, "dataset:") {
							permDataset = strings.TrimPrefix(tag, "dataset:")
						}
					}
					
					// Check for exact match or wildcard permissions in the right dataset
					if (permResource == resource || permResource == "*") &&
						(permAction == action || permAction == "*") &&
						(permDataset == datasetID || permDataset == "*") {
						return true, nil
					}
				}
			}
		}
	}
	
	return false, nil
}

func (sm *SecurityManager) roleHasDatasetAccess(roleID, datasetID string) (bool, error) {
	relationships, err := sm.entityRepo.GetRelationshipsBySource(roleID)
	if err != nil {
		return false, err
	}
	
	for _, rel := range relationships {
		if relationship, ok := rel.(*EntityRelationship); ok {
			if relationship.Type == RelationshipCanAccess && relationship.TargetID == datasetID {
				return true, nil
			}
		}
	}
	
	return false, nil
}

// CanAccessDataset checks if a user can access a specific dataset
func (sm *SecurityManager) CanAccessDataset(user *SecurityUser, datasetID string) (bool, error) {
	// Check direct user access
	userRoles, err := sm.getUserRoles(user.ID)
	if err != nil {
		return false, err
	}
	
	for _, role := range userRoles {
		hasAccess, err := sm.roleHasDatasetAccess(role.ID, datasetID)
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
			hasAccess, err := sm.roleHasDatasetAccess(role.ID, datasetID)
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

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}