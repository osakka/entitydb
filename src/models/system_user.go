package models

import (
	"fmt"
	"strings"
	
	"golang.org/x/crypto/bcrypt"
	"entitydb/logger"
)

var (
	// BcryptCost is the configurable cost for bcrypt password hashing
	// Default: bcrypt.DefaultCost (10)
	BcryptCost = bcrypt.DefaultCost
)

// SystemUserManager handles the immutable system user
type SystemUserManager struct {
	entityRepo EntityRepository
}

// NewSystemUserManager creates a new system user manager
func NewSystemUserManager(entityRepo EntityRepository) *SystemUserManager {
	return &SystemUserManager{
		entityRepo: entityRepo,
	}
}

// InitializeSystemUser creates the immutable system user if it doesn't exist
// This user owns all bootstrapping entities and serves as the root of the ownership chain
func (sum *SystemUserManager) InitializeSystemUser() (*SecurityUser, error) {
	logger.Info("Initializing system user with UUID: %s", SystemUserID)
	
	// Check if system user already exists and validate it
	existingEntity, err := sum.entityRepo.GetByID(SystemUserID)
	if err == nil {
		// System user exists - check if it's a recovery placeholder
		cleanTags := existingEntity.GetTagsWithoutTimestamp()
		isRecoveryPlaceholder := false
		for _, tag := range cleanTags {
			if tag == "status:recovered" || tag == "recovery:placeholder" {
				isRecoveryPlaceholder = true
				logger.Warn("Found recovery placeholder for system user - will replace with proper entity")
				break
			}
		}
		
		// If it's a proper system user (not a placeholder), use it
		if !isRecoveryPlaceholder {
			logger.Info("System user already exists and is valid: %s", SystemUserID)
			return sum.entityToSecurityUser(existingEntity)
		}
		
		// Delete the recovery placeholder to create a proper system user
		logger.Info("Deleting recovery placeholder for system user")
		if deleteErr := sum.entityRepo.Delete(SystemUserID); deleteErr != nil {
			logger.Warn("Failed to delete system user placeholder: %v", deleteErr)
			// Continue anyway - Create() will handle conflicts
		}
	}
	
	// Create new system user with mandatory tags
	logger.Info("Creating new system user with immutable UUID: %s", SystemUserID)
	
	// System user is created by itself (bootstrap scenario)
	systemUserEntity, err := NewEntityWithMandatoryTags(
		"user",           // entityType
		"system",         // dataset
		SystemUserID,     // createdBy (self-reference for bootstrap)
		[]string{
			"identity:username:" + SystemUserUsername,
			"identity:uuid:" + SystemUserID,
			"name:System User",
			"status:active",
			"system:true",                    // Mark as system entity
			"rbac:role:system",              // System role (higher than admin)
			"rbac:perm:*",                   // All permissions wildcard
			"bootstrap:root",                // Root of the ownership chain
			"immutable:true",                // Cannot be deleted or modified
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create system user entity: %v", err)
	}
	
	// Set the ID to the exact system user UUID (no generation needed)
	systemUserEntity.ID = SystemUserID
	
	// No credentials for system user - it's used for ownership only
	systemUserEntity.Content = nil
	
	// Create the system user entity
	if err := sum.entityRepo.Create(systemUserEntity); err != nil {
		// If entity still exists after delete, try to retrieve it
		if strings.Contains(err.Error(), "already exists") {
			logger.Info("System user already exists after recreation attempt, retrieving existing entity")
			existingEntity, getErr := sum.entityRepo.GetByID(SystemUserID)
			if getErr != nil {
				return nil, fmt.Errorf("system user exists but cannot retrieve: %v", getErr)
			}
			return sum.entityToSecurityUser(existingEntity)
		}
		return nil, fmt.Errorf("failed to create system user in repository: %v", err)
	}
	
	logger.Info("Successfully created system user: %s", SystemUserID)
	return sum.entityToSecurityUser(systemUserEntity)
}

// GetSystemUser retrieves the system user
func (sum *SystemUserManager) GetSystemUser() (*SecurityUser, error) {
	systemEntity, err := sum.entityRepo.GetByID(SystemUserID)
	if err != nil {
		return nil, fmt.Errorf("system user not found: %v", err)
	}
	
	return sum.entityToSecurityUser(systemEntity)
}

// VerifySystemUser validates that the system user exists and has correct properties
func (sum *SystemUserManager) VerifySystemUser() error {
	systemUser, err := sum.GetSystemUser()
	if err != nil {
		return fmt.Errorf("system user verification failed: %v", err)
	}
	
	// Verify UUID
	if systemUser.ID != SystemUserID {
		return fmt.Errorf("system user has incorrect UUID: expected %s, got %s", SystemUserID, systemUser.ID)
	}
	
	// Verify username (skip for recovered entities)
	if systemUser.Username != SystemUserUsername {
		// For recovered entities, allow missing username if UUID is correct
		if systemUser.ID == SystemUserID && systemUser.Username == "" {
			logger.Warn("System user has empty username (likely recovered entity) - allowing for bootstrap")
		} else {
			return fmt.Errorf("system user has incorrect username: expected %s, got %s", SystemUserUsername, systemUser.Username)
		}
	}
	
	// Check if this is a recovered entity placeholder
	cleanTags := systemUser.Entity.GetTagsWithoutTimestamp()
	isRecoveredPlaceholder := false
	for _, tag := range cleanTags {
		if tag == "status:recovered" && systemUser.ID == SystemUserID {
			isRecoveredPlaceholder = true
			logger.Warn("System user is a recovery placeholder - skipping validation during bootstrap")
			break
		}
	}
	
	// For non-recovered entities, verify system tags
	if !isRecoveredPlaceholder {
		hasSystemTag := false
		for _, tag := range cleanTags {
			if tag == "system:true" || tag == "bootstrap:root" {
				hasSystemTag = true
				break
			}
		}
		
		if !hasSystemTag {
			return fmt.Errorf("system user is not marked as system entity")
		}
	}
	
	// Verify mandatory tags (skip for recovered entities during bootstrap)
	if !isRecoveredPlaceholder {
		if err := systemUser.Entity.ValidateEntity(); err != nil {
			return fmt.Errorf("system user failed entity validation: %v", err)
		}
	} else {
		logger.Warn("Skipping mandatory tag validation for recovered system user placeholder")
	}
	
	logger.Info("System user verification successful: %s", SystemUserID)
	return nil
}

// entityToSecurityUser converts an Entity to SecurityUser
func (sum *SystemUserManager) entityToSecurityUser(entity *Entity) (*SecurityUser, error) {
	cleanTags := entity.GetTagsWithoutTimestamp()
	
	var username, email string
	for _, tag := range cleanTags {
		if tag == "identity:username:" + SystemUserUsername {
			username = SystemUserUsername
		} else if len(tag) > len("profile:email:") && tag[:len("profile:email:")] == "profile:email:" {
			email = tag[len("profile:email:"):]
		}
	}
	
	if username == "" {
		username = SystemUserUsername // Default fallback
	}
	
	return &SecurityUser{
		ID:       entity.ID,
		Username: username,
		Email:    email,
		Status:   "active",
		Entity:   entity,
	}, nil
}

// CreateAdminUser creates the admin user owned by the system user
// This replaces the old self-owning admin user pattern
func (sum *SystemUserManager) CreateAdminUser(username, password, email string) (*SecurityUser, error) {
	logger.Info("Creating admin user owned by system user")
	
	// Check if admin user already exists
	existingAdmins, err := sum.entityRepo.ListByTag("identity:username:" + username)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing admin user: %v", err)
	}
	
	if len(existingAdmins) > 0 {
		logger.Info("Admin user already exists: %s", existingAdmins[0].ID)
		// Convert to SecurityUser for return
		cleanTags := existingAdmins[0].GetTagsWithoutTimestamp()
		var userEmail string
		for _, tag := range cleanTags {
			if len(tag) > len("profile:email:") && tag[:len("profile:email:")] == "profile:email:" {
				userEmail = tag[len("profile:email:"):]
				break
			}
		}
		
		return &SecurityUser{
			ID:       existingAdmins[0].ID,
			Username: username,
			Email:    userEmail,
			Status:   "active",
			Entity:   existingAdmins[0],
		}, nil
	}
	
	// Generate new UUID for admin user
	adminUUID, err := NewEntityUUID("user")
	if err != nil {
		return nil, fmt.Errorf("failed to generate admin UUID: %v", err)
	}
	
	// Generate password hash and salt for admin
	salt := generateSalt()
	hashedPassword, err := HashPassword(password + salt)
	if err != nil {
		return nil, fmt.Errorf("failed to hash admin password: %v", err)
	}
	
	// Create admin user entity owned by system user
	adminUserEntity, err := NewEntityWithMandatoryTags(
		"user",           // entityType
		"system",         // dataset
		SystemUserID,     // createdBy (owned by system user)
		[]string{
			"identity:username:" + username,
			"identity:uuid:" + adminUUID.Value,
			"name:Administrator",
			"status:active",
			"profile:email:" + email,
			"has:credentials",               // Has embedded credentials
			"rbac:role:admin",              // Admin role
			"rbac:perm:*",                  // All permissions wildcard
			"created_by_system:true",       // Marked as system-created
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin user entity: %v", err)
	}
	
	// Set the generated UUID
	adminUserEntity.ID = adminUUID.Value
	
	// Store credentials in content as salt|hash format
	credentialContent := fmt.Sprintf("%s|%s", salt, string(hashedPassword))
	adminUserEntity.Content = []byte(credentialContent)
	
	// Create the admin user entity
	if err := sum.entityRepo.Create(adminUserEntity); err != nil {
		return nil, fmt.Errorf("failed to create admin user in repository: %v", err)
	}
	
	logger.Info("Successfully created admin user: %s (owned by system user: %s)", adminUUID.Value, SystemUserID)
	
	return &SecurityUser{
		ID:       adminUUID.Value,
		Username: username,
		Email:    email,
		Status:   "active",
		Entity:   adminUserEntity,
	}, nil
}

// HashPassword hashes a password using bcrypt with configurable cost
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
}

// SetBcryptCost sets the bcrypt cost for password hashing
func SetBcryptCost(cost int) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		logger.Warn("Invalid bcrypt cost %d, using default %d", cost, bcrypt.DefaultCost)
		BcryptCost = bcrypt.DefaultCost
		return
	}
	BcryptCost = cost
	logger.Info("Bcrypt cost set to %d", cost)
}