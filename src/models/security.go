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

// SecurityUser represents a user in the security system with authentication capabilities.
type SecurityUser struct {
	ID       string  // Unique identifier matching the underlying entity ID
	Username string  // Login username (must be unique across the system)
	Email    string  // Contact email address (optional, used for notifications)
	Status   string  // Account status: "active", "inactive", "suspended", "deleted"
	Entity   *Entity // Underlying entity containing user data and permissions
}

// SecuritySession represents an active user session with tracking and expiration.
type SecuritySession struct {
	ID        string    // Unique session identifier (cryptographically secure)
	Token     string    // Authentication token presented by client
	UserID    string    // ID of the authenticated user this session belongs to
	ExpiresAt time.Time // Session expiration timestamp (UTC)
	CreatedAt time.Time // Session creation timestamp (UTC)
	IPAddress string    // Client IP address for security auditing
	UserAgent string    // Client user agent string for device tracking
	Entity    *Entity   // Underlying entity storing session metadata and audit trail
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
		"dataset:system",
		"identity:username:" + username,
		"identity:uuid:" + userID,
		"name:" + username, // Friendly display name
		"status:active",
		"profile:email:" + email,
		"created:" + NowString(),
		"has:credentials", // Tag to indicate this user has embedded credentials
	}
	
	// Add comprehensive RBAC permissions for admin user
	if username == "admin" {
		tags = append(tags, 
			"rbac:role:admin",
			"rbac:perm:*:*", // All permissions
		)
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
			"dataset:system",
			"token:" + token,
			"expires:" + expiresAt.Format(time.RFC3339),
			"ip:" + ipAddress,
			"user_agent:" + userAgent,
			"created:" + NowString(),
			"authenticated_as:" + user.ID,
			"session:active",
		},
		Content:   nil,
		CreatedAt: Now(),
		UpdatedAt: Now(),
	}
	
	// Create session entity with relationship tags
	if err := sm.entityRepo.Create(sessionEntity); err != nil {
		logger.Error("Failed to create session entity: %v", err)
		return nil, fmt.Errorf("failed to create session entity: %v", err)
	}
	logger.Debug("Created session entity with user relationship: %s -> %s", sessionID, user.ID)
	
	// Session is now created without triggering immediate verification
	// This prevents the indexing race condition that caused recovery attempts
	
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
	logger.Debug("ValidateSession: Looking for session with token: %s", token)
	
	// Find session by token tag using the fixed temporal tag search
	// Retry mechanism to handle indexing delays
	var sessionEntities []*Entity
	var err error
	
	for i := 0; i < 3; i++ {
		sessionEntities, err = sm.entityRepo.ListByTag("token:" + token)
		if err != nil {
			logger.Error("ValidateSession: Error finding session (attempt %d): %v", i+1, err)
			if i < 2 {
				time.Sleep(10 * time.Millisecond) // Short delay before retry
				continue
			}
			return nil, fmt.Errorf("session lookup failed: %v", err)
		}
		if len(sessionEntities) > 0 {
			break // Found session
		}
		if i < 2 {
			logger.Debug("ValidateSession: Session not found, retrying (attempt %d)", i+1)
			time.Sleep(10 * time.Millisecond) // Short delay before retry
		}
	}
	
	if len(sessionEntities) == 0 {
		logger.Debug("ValidateSession: No session found with token after retries: %s", token)
		return nil, fmt.Errorf("session not found")
	}
	
	sessionEntity := sessionEntities[0]
	logger.Debug("ValidateSession: Found session entity: %s", sessionEntity.ID)
	
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
	
	// Get user via tag-based relationship
	var userID string
	for _, tag := range sessionTags {
		if strings.HasPrefix(tag, "authenticated_as:") {
			userID = strings.TrimPrefix(tag, "authenticated_as:")
			break
		}
	}
	
	if userID == "" {
		logger.Debug("ValidateSession: No authenticated_as tag found in session")
		return nil, fmt.Errorf("no user found for session")
	}
	
	userEntity, err := sm.entityRepo.GetByID(userID)
	if err != nil {
		logger.Debug("ValidateSession: Failed to get user entity %s: %v", userID, err)
		return nil, fmt.Errorf("user not found: %v", err)
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

// HasPermission checks if a user has a specific permission via tag-based RBAC
func (sm *SecurityManager) HasPermission(user *SecurityUser, resource, action string) (bool, error) {
	return sm.HasPermissionInDataset(user, resource, action, "")
}

// HasPermissionInDataset checks if a user has a specific permission in a dataset via tag-based RBAC
func (sm *SecurityManager) HasPermissionInDataset(user *SecurityUser, resource, action, datasetID string) (bool, error) {
	userTags := user.Entity.GetTagsWithoutTimestamp()
	logger.Debug("HasPermissionInDataset: checking permission %s:%s for user %s with tags: %v", resource, action, user.ID, userTags)
	
	// Check for admin role (has all permissions)
	for _, tag := range userTags {
		if tag == "rbac:role:admin" {
			logger.Debug("HasPermissionInDataset: user %s has admin role, granting permission", user.ID)
			return true, nil
		}
	}
	
	// Check for specific permissions via rbac:perm: tags
	requiredPerm := fmt.Sprintf("rbac:perm:%s:%s", resource, action)
	wildcardResource := fmt.Sprintf("rbac:perm:%s:*", resource)
	wildcardAction := fmt.Sprintf("rbac:perm:*:%s", action)
	wildcardAll := "rbac:perm:*:*"
	
	logger.Debug("HasPermissionInDataset: checking for permission tags: %s, %s, %s, %s", requiredPerm, wildcardResource, wildcardAction, wildcardAll)
	
	for _, tag := range userTags {
		if tag == requiredPerm || tag == wildcardResource || tag == wildcardAction || tag == wildcardAll {
			logger.Debug("HasPermissionInDataset: found matching permission tag: %s", tag)
			return true, nil
		}
	}
	
	logger.Debug("HasPermissionInDataset: no matching permissions found for user %s", user.ID)
	return false, nil
}

// Helper functions

func (sm *SecurityManager) getUserRoles(userID string) ([]*SecurityRole, error) {
	// Get user entity to check role tags
	userEntity, err := sm.entityRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	
	var roles []*SecurityRole
	userTags := userEntity.GetTagsWithoutTimestamp()
	
	// Extract roles from rbac:role: tags
	for _, tag := range userTags {
		if strings.HasPrefix(tag, "rbac:role:") {
			roleName := strings.TrimPrefix(tag, "rbac:role:")
			role := &SecurityRole{
				ID:     "role_" + roleName,
				Name:   roleName,
				Entity: userEntity, // Role is embedded in user, not separate entity
			}
			roles = append(roles, role)
		}
	}
	
	return roles, nil
}

func (sm *SecurityManager) getUserGroups(userID string) ([]*Entity, error) {
	// For now, simplified - groups would be handled via tags like "member_of:group_id"
	// This can be implemented later if needed
	return []*Entity{}, nil
}

func (sm *SecurityManager) getGroupRoles(groupID string) ([]*SecurityRole, error) {
	// Simplified - groups not implemented yet
	return []*SecurityRole{}, nil
}

// Obsolete relationship-based methods removed - using tag-based RBAC now

// CanAccessDataset checks if a user can access a specific dataset via tag-based RBAC
func (sm *SecurityManager) CanAccessDataset(user *SecurityUser, datasetID string) (bool, error) {
	userTags := user.Entity.GetTagsWithoutTimestamp()
	
	// Admin users have access to all datasets
	for _, tag := range userTags {
		if tag == "rbac:role:admin" {
			return true, nil
		}
	}
	
	// Check for specific dataset access tags (can be implemented later)
	// For now, regular users have access to default dataset
	if datasetID == "default" || datasetID == "" {
		return true, nil
	}
	
	return false, nil
}

// InvalidateSession invalidates a session by token
func (sm *SecurityManager) InvalidateSession(token string) error {
	logger.Debug("InvalidateSession: Looking for session with token: %s", token)
	
	// Find session by token tag
	sessionEntities, err := sm.entityRepo.ListByTag("token:" + token)
	if err != nil {
		logger.Error("InvalidateSession: Error finding session: %v", err)
		return fmt.Errorf("failed to find session: %v", err)
	}
	if len(sessionEntities) == 0 {
		logger.Debug("InvalidateSession: No session found with token: %s", token)
		return fmt.Errorf("session not found")
	}
	
	sessionEntity := sessionEntities[0]
	logger.Debug("InvalidateSession: Found session entity: %s", sessionEntity.ID)
	
	// Update session tags to mark as expired
	updatedTags := []string{}
	for _, tag := range sessionEntity.Tags {
		// Skip expires tag - we'll add a new one
		if !strings.Contains(tag, "expires:") {
			updatedTags = append(updatedTags, tag)
		}
	}
	
	// Add expired tag with past timestamp
	expiredTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	updatedTags = append(updatedTags, "expires:"+expiredTime)
	updatedTags = append(updatedTags, "status:invalidated")
	
	sessionEntity.Tags = updatedTags
	sessionEntity.UpdatedAt = Now()
	
	// Update the session entity
	if err := sm.entityRepo.Update(sessionEntity); err != nil {
		logger.Error("InvalidateSession: Failed to update session entity: %v", err)
		return fmt.Errorf("failed to invalidate session: %v", err)
	}
	
	logger.Debug("InvalidateSession: Successfully invalidated session: %s", sessionEntity.ID)
	return nil
}

// RefreshSession extends the expiration time of an existing session
func (sm *SecurityManager) RefreshSession(token string) (*SecuritySession, error) {
	logger.Debug("RefreshSession: Looking for session with token: %s", token)
	
	// Find session by token tag
	sessionEntities, err := sm.entityRepo.ListByTag("token:" + token)
	if err != nil {
		logger.Error("RefreshSession: Error finding session: %v", err)
		return nil, fmt.Errorf("failed to find session: %v", err)
	}
	if len(sessionEntities) == 0 {
		logger.Debug("RefreshSession: No session found with token: %s", token)
		return nil, fmt.Errorf("session not found")
	}
	
	sessionEntity := sessionEntities[0]
	logger.Debug("RefreshSession: Found session entity: %s", sessionEntity.ID)
	
	// Check if session is still valid (not already expired)
	currentTime := time.Now()
	var expiresAt time.Time
	for _, tag := range sessionEntity.Tags {
		if strings.HasPrefix(tag, "expires:") {
			expiryStr := strings.TrimPrefix(tag, "expires:")
			if parsedTime, err := time.Parse(time.RFC3339, expiryStr); err == nil {
				expiresAt = parsedTime
				break
			}
		}
	}
	
	if currentTime.After(expiresAt) {
		logger.Debug("RefreshSession: Session already expired: %s", sessionEntity.ID)
		return nil, fmt.Errorf("session expired")
	}
	
	// Generate new expiration time (2 hours from now)
	newExpiresAt := time.Now().Add(2 * time.Hour)
	
	// Update session tags with new expiration
	updatedTags := []string{}
	for _, tag := range sessionEntity.Tags {
		// Skip old expires tag - we'll add a new one
		if !strings.HasPrefix(tag, "expires:") {
			updatedTags = append(updatedTags, tag)
		}
	}
	
	// Add new expiration tag
	updatedTags = append(updatedTags, "expires:"+newExpiresAt.Format(time.RFC3339))
	
	sessionEntity.Tags = updatedTags
	sessionEntity.UpdatedAt = Now()
	
	// Update the session entity
	if err := sm.entityRepo.Update(sessionEntity); err != nil {
		logger.Error("RefreshSession: Failed to update session entity: %v", err)
		return nil, fmt.Errorf("failed to refresh session: %v", err)
	}
	
	// Create SecuritySession struct to return
	securitySession := &SecuritySession{
		Token:     token,
		ExpiresAt: newExpiresAt,
		UserID:    "",
		IPAddress: "",
		UserAgent: "",
	}
	
	// Extract session details from tags
	for _, tag := range sessionEntity.Tags {
		if strings.HasPrefix(tag, "authenticated_as:") {
			securitySession.UserID = strings.TrimPrefix(tag, "authenticated_as:")
		} else if strings.HasPrefix(tag, "ip:") {
			securitySession.IPAddress = strings.TrimPrefix(tag, "ip:")
		} else if strings.HasPrefix(tag, "user_agent:") {
			securitySession.UserAgent = strings.TrimPrefix(tag, "user_agent:")
		}
	}
	
	logger.Info("RefreshSession: Successfully refreshed session: %s", sessionEntity.ID)
	return securitySession, nil
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