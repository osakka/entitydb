package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
	"golang.org/x/crypto/bcrypt"
	
	"entitydb/logger"
)

// sessionValidationResult caches session validation results (safe copy, no pointers)
type sessionValidationResult struct {
	userID    string
	username  string
	email     string
	timestamp time.Time
	expiry    time.Time
}

// SecurityManager handles all relationship-based security operations
type SecurityManager struct {
	entityRepo          EntityRepository
	sessionCache        sync.Map // map[string]*sessionValidationResult
	sessionCacheTTL     time.Duration
	sessionCacheMutex   sync.RWMutex // Prevents race conditions in session validation
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(entityRepo EntityRepository) *SecurityManager {
	return &SecurityManager{
		entityRepo:      entityRepo,
		sessionCacheTTL: 30 * time.Second, // Cache validation results for 30 seconds
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

// CreateUser creates a new user entity using the new UUID-based architecture
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
	
	// Generate password hash and salt
	salt := generateSalt()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}
	
	// Prepare additional tags for the user
	additionalTags := []string{
		"identity:username:" + username,
		"name:" + username, // Friendly display name
		"status:active",
		"profile:email:" + email,
		"has:credentials", // Tag to indicate this user has embedded credentials
	}
	
	// Add comprehensive RBAC permissions for admin user
	if username == "admin" {
		additionalTags = append(additionalTags, 
			"rbac:role:admin",
			"rbac:perm:*:*", // All permissions
		)
	} else {
		// Regular users get basic permissions
		additionalTags = append(additionalTags,
			"rbac:role:user",
			"rbac:perm:entity:view",
			"rbac:perm:entity:create",
			"rbac:perm:entity:update",
		)
	}
	
	// Create user entity with mandatory tags (owned by system user)
	userEntity, err := NewEntityWithMandatoryTags(
		EntityTypeUser,  // entityType
		"system",        // dataset
		SystemUserID,    // createdBy (system user owns all users)
		additionalTags,  // additional tags
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user entity structure: %v", err)
	}
	
	// Store credentials in content as salt|hash format
	credentialContent := fmt.Sprintf("%s|%s", salt, string(hashedPassword))
	userEntity.Content = []byte(credentialContent)
	
	// Create user entity with embedded credentials
	if err := sm.entityRepo.Create(userEntity); err != nil {
		logger.Error("failed to create user entity: %v", err)
		return nil, fmt.Errorf("failed to create user entity: %v", err)
	}
	logger.TraceIf("auth", "successfully created user entity with UUID: %s", userEntity.ID)
	
	return &SecurityUser{
		ID:       userEntity.ID,
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
	
	logger.TraceIf("auth", "AuthenticateUser returning user with ID: %s (length: %d)", userEntity.ID, len(userEntity.ID))
	
	return &SecurityUser{
		ID:       userEntity.ID,
		Username: username,
		Email:    email,
		Status:   "active",
		Entity:   userEntity,
	}, nil
}

// CreateSession creates a new session entity using the new UUID-based architecture
func (sm *SecurityManager) CreateSession(user *SecurityUser, ipAddress, userAgent string) (*SecuritySession, error) {
	token := generateSecureToken()
	expiresAt := time.Now().Add(2 * time.Hour) // 2 hour sessions
	
	// Prepare additional tags for the session
	additionalTags := []string{
		"token:" + token,
		"expires:" + expiresAt.Format(time.RFC3339),
		"ip:" + ipAddress,
		"user_agent:" + userAgent,
		"authenticated_as:" + user.ID,
		"session:active",
	}
	
	// Create session entity with mandatory tags (owned by the user)
	sessionEntity, err := NewEntityWithMandatoryTags(
		EntityTypeSession, // entityType
		"system",          // dataset
		user.ID,           // createdBy (user owns their sessions)
		additionalTags,    // additional tags
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session entity structure: %v", err)
	}
	
	// No content for sessions
	sessionEntity.Content = nil
	
	// Create session entity with relationship tags
	logger.TraceIf("auth", "creating session entity %s with %d tags", sessionEntity.ID, len(sessionEntity.Tags))
	if err := sm.entityRepo.Create(sessionEntity); err != nil {
		logger.Error("Failed to create session entity: %v", err)
		return nil, fmt.Errorf("failed to create session entity: %v", err)
	}
	logger.Debug("session created for user %s", user.Username)
	
	// Verify session is immediately findable to prevent test suite failures
	// Wait for indexing to complete by attempting to find the session
	var sessionEntities []*Entity
	for i := 0; i < 5; i++ {
		sessionEntities, err = sm.entityRepo.ListByTag("token:" + token)
		if err == nil && len(sessionEntities) > 0 {
			break // Session is findable
		}
		if i < 4 {
			logger.Debug("CreateSession: Session not immediately findable, waiting 10ms (attempt %d)", i+1)
			time.Sleep(10 * time.Millisecond) // Short delay for indexing completion
		}
	}
	
	if len(sessionEntities) == 0 {
		logger.Warn("CreateSession: Session created but not immediately findable - may cause validation failures")
	} else {
		logger.Debug("CreateSession: Session verified as findable after %d attempts", len(sessionEntities))
	}
	
	return &SecuritySession{
		ID:        sessionEntity.ID,
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: expiresAt,
		CreatedAt: time.Unix(0, sessionEntity.CreatedAt),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Entity:    sessionEntity,
	}, nil
}

// ValidateSession validates a session token and returns the associated user (with caching)
func (sm *SecurityManager) ValidateSession(token string) (*SecurityUser, error) {
	logger.Debug("ValidateSession: Looking for session with token: %s", token)
	
	// SURGICAL FIX: Use read lock for cache check to prevent race conditions
	sm.sessionCacheMutex.RLock()
	
	// Check cache first
	if cached, ok := sm.sessionCache.Load(token); ok {
		result := cached.(*sessionValidationResult)
		
		// Check if cache entry is still valid and session hasn't expired
		now := time.Now()
		if now.Sub(result.timestamp) < sm.sessionCacheTTL && now.Before(result.expiry) {
			logger.Debug("ValidateSession: Cache hit for token: %s", token)
			sm.sessionCacheMutex.RUnlock() // Release read lock before user entity fetch
			
			// CRITICAL FIX: Must fetch entity for permission checking
			userEntity, err := sm.entityRepo.GetByID(result.userID)
			if err != nil {
				// Cache invalidated due to missing entity - use write lock for deletion
				sm.sessionCacheMutex.Lock()
				sm.sessionCache.Delete(token)
				sm.sessionCacheMutex.Unlock()
				logger.Debug("ValidateSession: Cache invalidated due to missing user entity: %s", result.userID)
				// SURGICAL FIX: Fall through to database lookup instead of returning error
			} else {
				return &SecurityUser{
					ID:       result.userID,
					Username: result.username,
					Email:    result.email,
					Status:   "active",
					Entity:   userEntity, // Must include entity for RBAC permission checking
				}, nil
			}
		} else {
			// Cache expired or session expired, remove it - upgrade to write lock
			sm.sessionCacheMutex.RUnlock()
			sm.sessionCacheMutex.Lock()
			sm.sessionCache.Delete(token)
			sm.sessionCacheMutex.Unlock()
			logger.Debug("ValidateSession: Cache expired for token: %s", token)
		}
	} else {
		sm.sessionCacheMutex.RUnlock()
	}
	
	logger.Debug("ValidateSession: Cache miss, performing database lookup for token: %s", token)
	
	// Find session by token tag using the fixed temporal tag search
	// Retry mechanism to handle indexing delays
	var sessionEntities []*Entity
	var err error
	
	for i := 0; i < 3; i++ { // SURGICAL FIX: Increased retries to handle indexing delays
		logger.Debug("ValidateSession: Attempt %d to find session with token: %s", i+1, token)
		sessionEntities, err = sm.entityRepo.ListByTag("token:" + token)
		if err != nil {
			logger.Error("ValidateSession: Error finding session (attempt %d): %v", i+1, err)
			if i < 2 {
				time.Sleep(25 * time.Millisecond) // SURGICAL FIX: Increased delay for proper indexing
				continue
			}
			return nil, fmt.Errorf("session lookup failed: %v", err)
		}
		logger.Debug("ValidateSession: Found %d session entities on attempt %d", len(sessionEntities), i+1)
		if len(sessionEntities) > 0 {
			break // Found session
		}
		if i < 2 {
			logger.Debug("ValidateSession: Session not found, retrying (attempt %d) after 25ms", i+1)
			time.Sleep(25 * time.Millisecond) // SURGICAL FIX: Proper delay for indexing completion
		}
	}
	
	if len(sessionEntities) == 0 {
		logger.Debug("ValidateSession: No session found with token after retries: %s", token)
		return nil, fmt.Errorf("session not found")
	}
	
	sessionEntity := sessionEntities[0]
	logger.Debug("ValidateSession: Found session entity: %s", sessionEntity.ID)
	
	// Check if session is expired or invalidated
	sessionTags := sessionEntity.GetTagsWithoutTimestamp()
	var expiresAt time.Time
	var expirationStr string
	var isInvalidated bool
	
	logger.Trace("auth", "ValidateSession: Processing session %s with %d tags", sessionEntity.ID, len(sessionTags))
	
	for _, tag := range sessionTags {
		if strings.HasPrefix(tag, "expires:") {
			expirationStr = strings.TrimPrefix(tag, "expires:")
			expiresAt, err = time.Parse(time.RFC3339, expirationStr)
			if err != nil {
				logger.Error("ValidateSession: Failed to parse expiration time '%s': %v", expirationStr, err)
				return nil, fmt.Errorf("invalid session expiration format")
			}
		} else if tag == "status:invalidated" {
			isInvalidated = true
			logger.Error("ValidateSession: FOUND status:invalidated tag for session: %s", sessionEntity.ID)
		}
	}
	
	// Check for invalidated status first (logout)
	if isInvalidated {
		logger.Error("ValidateSession: Session invalidated via logout: %s", sessionEntity.ID)
		return nil, fmt.Errorf("session invalidated")
	}
	
	// Check for time-based expiration
	currentTime := time.Now()
	logger.Info("ValidateSession: Current time: %s, Session expires: %s (raw: %s)", 
		currentTime.Format(time.RFC3339), expiresAt.Format(time.RFC3339), expirationStr)
	
	if currentTime.After(expiresAt) {
		logger.Error("ValidateSession: Session expired - current: %s, expires: %s", 
			currentTime.Format(time.RFC3339), expiresAt.Format(time.RFC3339))
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
	
	// Cache the validation result for future requests (safe copy, no entity pointers)
	// SURGICAL FIX: Use write lock for cache update to prevent race conditions
	sm.sessionCacheMutex.Lock()
	sm.sessionCache.Store(token, &sessionValidationResult{
		userID:    userEntity.ID,
		username:  username,
		email:     email,
		timestamp: time.Now(),
		expiry:    expiresAt,
	})
	sm.sessionCacheMutex.Unlock()
	logger.Debug("ValidateSession: Cached validation result for token: %s", token)
	
	return &SecurityUser{
		ID:       userEntity.ID,
		Username: username,
		Email:    email,
		Status:   "active",
		Entity:   userEntity,
	}, nil
}

// InvalidateSessionCache invalidates a session from the cache (called during logout)
func (sm *SecurityManager) InvalidateSessionCache(token string) {
	// SURGICAL FIX: Use write lock for cache invalidation to prevent race conditions
	sm.sessionCacheMutex.Lock()
	sm.sessionCache.Delete(token)
	sm.sessionCacheMutex.Unlock()
	logger.Debug("InvalidateSessionCache: Removed token from cache: %s", token)
}

// HasPermission checks if a user has a specific permission via tag-based RBAC
func (sm *SecurityManager) HasPermission(user *SecurityUser, resource, action string) (bool, error) {
	return sm.HasPermissionInDataset(user, resource, action, "")
}

// HasPermissionInDataset checks if a user has a specific permission in a dataset via tag-based RBAC
func (sm *SecurityManager) HasPermissionInDataset(user *SecurityUser, resource, action, datasetID string) (bool, error) {
	// CRITICAL FIX: Prevent nil pointer dereference in production
	if user == nil {
		logger.Error("HasPermissionInDataset: user is nil")
		return false, fmt.Errorf("user cannot be nil")
	}
	if user.Entity == nil {
		logger.Error("HasPermissionInDataset: user.Entity is nil for user %s", user.ID)
		return false, fmt.Errorf("user entity not loaded")
	}
	
	userTags := user.Entity.GetTagsWithoutTimestamp()
	logger.Debug("HasPermissionInDataset: checking permission %s:%s for user %s with tags: %v", resource, action, user.ID, userTags)
	
	// Check for admin role (has all permissions)
	for _, tag := range userTags {
		if tag == "rbac:role:admin" {
			logger.Debug("HasPermissionInDataset: user %s has admin role, granting permission", user.ID)
			return true, nil
		}
	}
	
	// Modern permission checking - v2.32.0+ format
	// Check for admin wildcard permission (grants all access)
	for _, tag := range userTags {
		if tag == "rbac:perm:*" || tag == "rbac:perm:*:*" {
			logger.Debug("HasPermissionInDataset: user %s has wildcard permission, granting access", user.ID)
			return true, nil
		}
	}
	
	// Check for specific permission
	requiredPerm := fmt.Sprintf("rbac:perm:%s:%s", resource, action)
	for _, tag := range userTags {
		if tag == requiredPerm {
			logger.Debug("HasPermissionInDataset: found specific permission tag: %s", tag)
			return true, nil
		}
	}
	
	// Check for resource wildcard (e.g., rbac:perm:entity:*)
	resourceWildcard := fmt.Sprintf("rbac:perm:%s:*", resource)
	for _, tag := range userTags {
		if tag == resourceWildcard {
			logger.Debug("HasPermissionInDataset: found resource wildcard permission: %s", tag)
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
	
	logger.Debug("InvalidateSession: Original tags: %v", sessionEntity.Tags)
	logger.Debug("InvalidateSession: Updated tags: %v", updatedTags)
	
	// CRITICAL FIX: Use SetTags to properly invalidate entity tag cache
	// Direct assignment sessionEntity.Tags = updatedTags bypasses cache invalidation
	sessionEntity.SetTags(updatedTags)
	sessionEntity.UpdatedAt = Now()
	
	// Update the session entity
	if err := sm.entityRepo.Update(sessionEntity); err != nil {
		logger.Error("InvalidateSession: Failed to update session entity: %v", err)
		return fmt.Errorf("failed to invalidate session: %v", err)
	}
	
	// CRITICAL: Clear the session from cache to ensure immediate invalidation
	sm.InvalidateSessionCache(token)
	
	logger.Debug("InvalidateSession: Successfully invalidated session: %s", sessionEntity.ID)
	return nil
}

// RefreshSession extends the expiration time of an existing session
func (sm *SecurityManager) RefreshSession(token string) (*SecuritySession, error) {
	logger.TraceIf("auth", "refreshing session with token prefix: %s", token[:8])
	
	// Find session by token tag
	searchTag := "token:" + token
	logger.TraceIf("auth", "searching for session tag: %s", searchTag)
	sessionEntities, err := sm.entityRepo.ListByTag(searchTag)
	if err != nil {
		logger.Error("session refresh failed to find session: %v", err)
		return nil, fmt.Errorf("failed to find session: %v", err)
	}
	logger.TraceIf("auth", "found %d session entities", len(sessionEntities))
	if len(sessionEntities) == 0 {
		logger.Warn("session refresh failed: session not found")
		return nil, fmt.Errorf("session not found")
	}
	
	sessionEntity := sessionEntities[0]
	logger.Debug("RefreshSession: Found session entity: %s", sessionEntity.ID)
	
	// Check if session is still valid (not already expired)
	currentTime := time.Now()
	var expiresAt time.Time
	var expiryStr string
	
	// Since we're in temporal-only mode, all tags have timestamps
	// We need to look at the clean tags (without timestamps)
	cleanTags := sessionEntity.GetTagsWithoutTimestamp()
	
	for _, tag := range cleanTags {
		if strings.HasPrefix(tag, "expires:") {
			expiryStr = strings.TrimPrefix(tag, "expires:")
			if parsedTime, err := time.Parse(time.RFC3339, expiryStr); err == nil {
				expiresAt = parsedTime
				break
			}
		}
	}
	
	logger.Debug("RefreshSession: Current time: %s, Session expires: %s (raw: %s)", 
		currentTime.Format(time.RFC3339), expiresAt.Format(time.RFC3339), expiryStr)
	
	if currentTime.After(expiresAt) {
		logger.Debug("RefreshSession: Session already expired: %s", sessionEntity.ID)
		return nil, fmt.Errorf("session expired")
	}
	
	// Extract session details BEFORE updating (to ensure we have the data)
	cleanSessionTags := sessionEntity.GetTagsWithoutTimestamp()
	logger.Info("RefreshSession: Clean session tags before update: %v", cleanSessionTags)
	
	var userID, ipAddress, userAgent string
	for _, tag := range cleanSessionTags {
		if strings.HasPrefix(tag, "authenticated_as:") {
			userID = strings.TrimPrefix(tag, "authenticated_as:")
			logger.Info("RefreshSession: Extracted UserID: %s", userID)
		} else if strings.HasPrefix(tag, "ip:") {
			ipAddress = strings.TrimPrefix(tag, "ip:")
		} else if strings.HasPrefix(tag, "user_agent:") {
			userAgent = strings.TrimPrefix(tag, "user_agent:")
		}
	}
	
	if userID == "" {
		logger.Error("RefreshSession: Failed to extract UserID from session tags")
		return nil, fmt.Errorf("session missing user information")
	}
	
	// Generate new expiration time (2 hours from now)
	newExpiresAt := time.Now().Add(2 * time.Hour)
	
	// Update session tags with new expiration
	updatedTags := []string{}
	for _, tag := range sessionEntity.Tags {
		// Skip old expires tag - we'll add a new one
		// Handle temporal tags with format TIMESTAMP|tag
		actualTag := tag
		if pipePos := strings.Index(tag, "|"); pipePos != -1 {
			actualTag = tag[pipePos+1:]
		}
		
		if !strings.HasPrefix(actualTag, "expires:") {
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
	
	// Create SecuritySession struct to return with extracted data
	securitySession := &SecuritySession{
		ID:        sessionEntity.ID,
		Token:     token,
		UserID:    userID,
		ExpiresAt: newExpiresAt,
		CreatedAt: time.Unix(0, sessionEntity.CreatedAt),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Entity:    sessionEntity,
	}
	
	logger.Info("RefreshSession: Final UserID: %s", securitySession.UserID)
	logger.Info("RefreshSession: Final session data - ID: %s, UserID: %s, Token: %s", securitySession.ID, securitySession.UserID, securitySession.Token[:8])
	
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