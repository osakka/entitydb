package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
	
	"entitydb/logger"
)

// UUID validation and generation for entities
const (
	// UUIDLength is the required length for entity UUIDs (32 hex characters)
	UUIDLength = 32
)

var (
	// SystemUserID is the configurable UUID for the system user
	// Default: "00000000000000000000000000000001" (configurable via ENTITYDB_SYSTEM_USER_ID)
	SystemUserID = "00000000000000000000000000000001"
	
	// SystemUserUsername is the configurable username for the system user
	// Default: "system" (configurable via ENTITYDB_SYSTEM_USERNAME)
	SystemUserUsername = "system"
)

var (
	// UUIDPattern regex for validating UUIDs (32 hex characters)
	UUIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)
	
	// Reserved system UUIDs that cannot be used for regular entities
	ReservedUUIDs = map[string]bool{
		SystemUserID: true,
		"00000000000000000000000000000000": true, // Null UUID
	}
)

// InitializeSystemUserConfiguration sets the system user configuration from provided values.
// This should be called once during application startup to configure system user parameters.
func InitializeSystemUserConfiguration(systemUserID, systemUsername string) {
	// Validate the provided system user ID
	if err := ValidateEntityUUID(systemUserID); err != nil {
		logger.Warn("Invalid system user ID provided: %s, using default", systemUserID)
		return
	}
	
	// Remove old system user ID from reserved UUIDs
	delete(ReservedUUIDs, SystemUserID)
	
	// Update the system user configuration
	SystemUserID = systemUserID
	SystemUserUsername = systemUsername
	
	// Add new system user ID to reserved UUIDs
	ReservedUUIDs[SystemUserID] = true
	
	logger.Info("System user configuration updated: ID=%s, Username=%s", SystemUserID, SystemUserUsername)
}

// EntityUUID represents a validated entity UUID
type EntityUUID struct {
	Value string
	Type  string // user, session, metric, config, etc.
}

// NewEntityUUID creates a new validated UUID for an entity type
func NewEntityUUID(entityType string) (*EntityUUID, error) {
	if entityType == "" {
		return nil, fmt.Errorf("entity type cannot be empty")
	}
	
	uuid, err := generateUUID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %v", err)
	}
	
	return &EntityUUID{
		Value: uuid,
		Type:  entityType,
	}, nil
}

// NewSystemUserUUID creates the system user UUID (immutable)
func NewSystemUserUUID() *EntityUUID {
	return &EntityUUID{
		Value: SystemUserID,
		Type:  "user",
	}
}

// ValidateEntityUUID validates that a string is a proper entity UUID
func ValidateEntityUUID(uuid string) error {
	if uuid == "" {
		return fmt.Errorf("UUID cannot be empty")
	}
	
	if len(uuid) != UUIDLength {
		return fmt.Errorf("UUID must be exactly %d characters, got %d", UUIDLength, len(uuid))
	}
	
	if !UUIDPattern.MatchString(uuid) {
		return fmt.Errorf("UUID must contain only lowercase hexadecimal characters")
	}
	
	return nil
}

// IsReservedUUID checks if a UUID is reserved for system use
func IsReservedUUID(uuid string) bool {
	return ReservedUUIDs[uuid]
}

// IsSystemUser checks if a UUID belongs to the system user
func IsSystemUser(uuid string) bool {
	return uuid == SystemUserID
}

// generateUUID creates a cryptographically secure 32-character hex UUID
func generateUUID() (string, error) {
	bytes := make([]byte, 16) // 16 bytes = 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}
	
	uuid := hex.EncodeToString(bytes)
	
	// Ensure we don't generate a reserved UUID (extremely unlikely but check anyway)
	if IsReservedUUID(uuid) {
		logger.Warn("Generated reserved UUID %s, regenerating", uuid)
		return generateUUID() // Recursive retry
	}
	
	return uuid, nil
}

// MandatoryTags represents the required tags for all entities
type MandatoryTags struct {
	Type      string    // Entity type (user, session, metric, etc.)
	Dataset   string    // Dataset namespace (system, default, etc.)
	CreatedAt time.Time // Creation timestamp
	CreatedBy string    // UUID of the user/system that created this entity
	UUID      string    // Entity's UUID (immutable)
}

// ValidateMandatoryTags ensures all required tags are present and valid
func ValidateMandatoryTags(tags []string) (*MandatoryTags, error) {
	mandatory := &MandatoryTags{}
	found := make(map[string]bool)
	
	// Extract mandatory tags from the tag list
	for _, tag := range tags {
		// Handle temporal tags by extracting the actual tag
		actualTag := tag
		if pipePos := strings.Index(tag, "|"); pipePos != -1 {
			actualTag = tag[pipePos+1:]
		}
		
		switch {
		case strings.HasPrefix(actualTag, "type:"):
			mandatory.Type = strings.TrimPrefix(actualTag, "type:")
			found["type"] = true
			
		case strings.HasPrefix(actualTag, "dataset:"):
			mandatory.Dataset = strings.TrimPrefix(actualTag, "dataset:")
			found["dataset"] = true
			
		case strings.HasPrefix(actualTag, "created_at:"):
			createdAtStr := strings.TrimPrefix(actualTag, "created_at:")
			// Parse nanosecond timestamp
			if ns, err := parseNanosecondTimestamp(createdAtStr); err == nil {
				mandatory.CreatedAt = time.Unix(0, ns)
				found["created_at"] = true
			} else {
				return nil, fmt.Errorf("invalid created_at timestamp: %v", err)
			}
			
		case strings.HasPrefix(actualTag, "created_by:"):
			mandatory.CreatedBy = strings.TrimPrefix(actualTag, "created_by:")
			found["created_by"] = true
			
		case strings.HasPrefix(actualTag, "uuid:"):
			mandatory.UUID = strings.TrimPrefix(actualTag, "uuid:")
			found["uuid"] = true
		}
	}
	
	// Check that all mandatory tags are present
	required := []string{"type", "dataset", "created_at", "created_by", "uuid"}
	for _, req := range required {
		if !found[req] {
			return nil, fmt.Errorf("missing mandatory tag: %s", req)
		}
	}
	
	// Validate the extracted values
	if err := ValidateEntityUUID(mandatory.UUID); err != nil {
		return nil, fmt.Errorf("invalid UUID in uuid tag: %v", err)
	}
	
	if err := ValidateEntityUUID(mandatory.CreatedBy); err != nil {
		return nil, fmt.Errorf("invalid UUID in created_by tag: %v", err)
	}
	
	if mandatory.Type == "" {
		return nil, fmt.Errorf("type tag cannot be empty")
	}
	
	if mandatory.Dataset == "" {
		return nil, fmt.Errorf("dataset tag cannot be empty")
	}
	
	if mandatory.CreatedAt.IsZero() {
		return nil, fmt.Errorf("created_at timestamp is invalid")
	}
	
	return mandatory, nil
}

// GenerateMandatoryTags creates the required tags for a new entity
func GenerateMandatoryTags(entityType, dataset, createdByUUID string) ([]string, *EntityUUID, error) {
	// Validate inputs
	if entityType == "" {
		return nil, nil, fmt.Errorf("entity type cannot be empty")
	}
	
	if dataset == "" {
		return nil, nil, fmt.Errorf("dataset cannot be empty")
	}
	
	if err := ValidateEntityUUID(createdByUUID); err != nil {
		return nil, nil, fmt.Errorf("invalid created_by UUID: %v", err)
	}
	
	// Generate new UUID for this entity
	entityUUID, err := NewEntityUUID(entityType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate entity UUID: %v", err)
	}
	
	// Create mandatory tags
	now := time.Now()
	tags := []string{
		"type:" + entityType,
		"dataset:" + dataset,
		"created_at:" + fmt.Sprintf("%d", now.UnixNano()),
		"created_by:" + createdByUUID,
		"uuid:" + entityUUID.Value,
	}
	
	return tags, entityUUID, nil
}

// parseNanosecondTimestamp parses a nanosecond timestamp string
func parseNanosecondTimestamp(timestampStr string) (int64, error) {
	if timestampStr == "" {
		return 0, fmt.Errorf("timestamp string is empty")
	}
	
	// Try parsing as nanosecond timestamp
	ns, err := parseInt64(timestampStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse timestamp as nanoseconds: %v", err)
	}
	
	// Sanity check: timestamp should be reasonable (after 2020, before 2050)
	year2020 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	year2050 := time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	
	if ns < year2020 || ns > year2050 {
		return 0, fmt.Errorf("timestamp %d is outside reasonable range", ns)
	}
	
	return ns, nil
}

// parseInt64 safely parses a string to int64
func parseInt64(s string) (int64, error) {
	var result int64
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid character in number: %c", r)
		}
		result = result*10 + int64(r-'0')
	}
	return result, nil
}

// EntityIDFromUUID creates a proper entity ID from a UUID and type
func EntityIDFromUUID(uuid, entityType string) string {
	return uuid // Pure UUID - no prefixes
}

// ExtractUUIDFromEntityID extracts UUID from entity ID (for migration compatibility)
func ExtractUUIDFromEntityID(entityID string) (string, error) {
	// If it's already a pure UUID, return it
	if err := ValidateEntityUUID(entityID); err == nil {
		return entityID, nil
	}
	
	// Handle legacy format with prefixes (user_, session_, etc.)
	parts := strings.SplitN(entityID, "_", 2)
	if len(parts) == 2 {
		uuid := parts[1]
		if err := ValidateEntityUUID(uuid); err == nil {
			return uuid, nil
		}
	}
	
	return "", fmt.Errorf("cannot extract valid UUID from entity ID: %s", entityID)
}