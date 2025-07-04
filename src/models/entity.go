// Package models defines the core data structures and interfaces for EntityDB.
//
// The models package provides:
//   - Entity definition with temporal tag support
//   - Repository interfaces for storage backends
//   - Session management structures
//   - Security and RBAC models
//   - Utility functions for entity manipulation
//
// All entities in EntityDB are represented as collections of timestamped tags,
// enabling powerful temporal queries and maintaining complete history.
package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"
)

// EntityRepository defines the contract for entity storage implementations.
// All storage backends must implement this interface to provide entity
// persistence with temporal support.
//
// Implementations include:
//   - BinaryEntityRepository: High-performance binary format storage
//   - CachedRepository: In-memory caching layer
//   - TemporalRepository: Temporal query optimizations
//   - WALOnlyRepository: Write-ahead log mode for extreme performance
type EntityRepository interface {
	// Core CRUD Operations
	
	// Create stores a new entity with auto-generated ID if not provided.
	// Returns an error if the entity already exists.
	Create(entity *Entity) error
	
	// GetByID retrieves an entity by its unique identifier.
	// Returns nil, nil if the entity doesn't exist.
	GetByID(id string) (*Entity, error)
	
	// Update modifies an existing entity, preserving temporal history.
	// Returns an error if the entity doesn't exist.
	Update(entity *Entity) error
	
	// Delete removes an entity from storage.
	// Note: This is a hard delete, not a soft delete with tags.
	Delete(id string) error
	
	// List returns all entities in the repository.
	// Warning: Can be memory-intensive for large datasets.
	List() ([]*Entity, error)
	
	// Tag-Based Queries
	
	// ListByTag returns entities with the specified tag.
	// Handles temporal tags transparently (strips timestamps).
	ListByTag(tag string) ([]*Entity, error)
	
	// ListByTags returns entities matching multiple tags.
	// If matchAll is true, entities must have ALL tags; otherwise ANY tag.
	ListByTags(tags []string, matchAll bool) ([]*Entity, error)
	
	// ListByTagSQL returns entities using SQL-like tag patterns.
	// Supports % wildcard for pattern matching.
	ListByTagSQL(tag string) ([]*Entity, error)
	
	// ListByTagWildcard returns entities matching a glob pattern.
	// Supports * and ? wildcards (e.g., "status:*", "user:?123").
	ListByTagWildcard(pattern string) ([]*Entity, error)
	
	// ListByNamespace returns entities with tags in the specified namespace.
	// Example: namespace "status" matches "status:active", "status:draft", etc.
	ListByNamespace(namespace string) ([]*Entity, error)
	
	// GetUniqueTagValues returns unique values for a given tag namespace.
	// Example: namespace "dataset" returns ["default", "production", "staging"].
	GetUniqueTagValues(namespace string) ([]string, error)
	
	// Content Queries
	
	// SearchContent performs full-text search on entity content.
	// Returns entities where content contains the query string.
	SearchContent(query string) ([]*Entity, error)
	
	
	// Advanced Query
	
	// QueryAdvanced executes complex queries with multiple criteria.
	// Supported parameters: tags, namespace, content, limit, offset, sort.
	QueryAdvanced(params map[string]interface{}) ([]*Entity, error)
	
	// Transaction Support
	
	// Transaction executes a function within a database transaction.
	// Automatically rolls back on error, commits on success.
	Transaction(fn func(tx interface{}) error) error
	
	// Commit explicitly commits a transaction.
	// Usually not needed when using Transaction().
	Commit(tx interface{}) error
	
	// Rollback explicitly rolls back a transaction.
	// Usually not needed when using Transaction().
	Rollback(tx interface{}) error
	
	// Tag Operations
	
	// AddTag appends a new tag to an entity.
	// The tag is automatically timestamped by the storage layer.
	AddTag(id string, tag string) error
	
	// RemoveTag removes all instances of a tag from an entity.
	// Handles temporal tags by matching the tag content.
	RemoveTag(id string, tag string) error
	
	// Temporal Operations
	
	// GetEntityAsOf returns the entity state at a specific point in time.
	// Reconstructs the entity by selecting appropriate temporal tag values.
	GetEntityAsOf(id string, timestamp time.Time) (*Entity, error)
	
	// GetEntityHistory returns the change history for an entity.
	// Limited to the specified number of most recent changes.
	GetEntityHistory(id string, limit int) ([]*EntityChange, error)
	
	// GetRecentChanges returns recent changes across all entities.
	// Useful for activity feeds and audit logs.
	GetRecentChanges(limit int) ([]*EntityChange, error)
	
	// GetEntityDiff compares entity states between two timestamps.
	// Returns (before, after) snapshots for the specified time range.
	GetEntityDiff(id string, startTime, endTime time.Time) (*Entity, *Entity, error)
	
	// Relationship operations removed - use pure tag-based relationships instead
	// Example: To relate entity A to entity B, add tag "relates_to:entity_B_id" to entity A
	
	// Query Builder
	
	// Query returns a new query builder for fluent query construction.
	// Example: repo.Query().WithTag("status:active").Limit(10).Execute()
	Query() *EntityQuery
	
	// Maintenance Operations
	
	// ReindexTags rebuilds the tag index from entity data.
	// Use after corruption or to optimize performance.
	ReindexTags() error
	
	// VerifyIndexHealth checks index integrity and consistency.
	// Returns an error if corruption is detected.
	VerifyIndexHealth() error
	
	// Lifecycle Operations
	
	// ListActive returns all entities in active state.
	// Equivalent to filtering by "status:active" tag.
	ListActive() ([]*Entity, error)
	
	// ListSoftDeleted returns all entities in soft deleted state.
	// Equivalent to filtering by "status:soft_deleted" tag.
	ListSoftDeleted() ([]*Entity, error)
	
	// ListArchived returns all entities in archived state.
	// Equivalent to filtering by "status:archived" tag.
	ListArchived() ([]*Entity, error)
	
	// ListByLifecycleState returns entities in a specific lifecycle state.
	// Provides unified access to lifecycle filtering.
	ListByLifecycleState(state EntityLifecycleState) ([]*Entity, error)
}

// Entity represents the universal data structure in EntityDB.
// Everything is an entity - users, documents, configurations, relationships.
//
// Entities consist of:
//   - A unique identifier (UUID)
//   - A collection of temporal tags with nanosecond timestamps
//   - Binary content (optional)
//   - Creation and update timestamps
//
// Example:
//
//	entity := &Entity{
//	    ID: "user-123",
//	    Tags: []string{
//	        "2024-01-15T10:30:45.123456789.type:user",
//	        "2024-01-15T10:30:45.123456789.status:active",
//	        "2024-01-15T10:30:45.123456789.rbac:role:admin",
//	    },
//	    Content: []byte("user profile data"),
//	}
type Entity struct {
	// ID is the unique identifier for the entity (typically a UUID)
	ID      string   `json:"id"`
	
	// Tags are temporal tags with nanosecond timestamps
	// Format: "TIMESTAMP|tag" or "TIMESTAMP.tag"
	Tags    []string `json:"tags"`
	
	// Content stores binary data (files, JSON, credentials, etc.)
	// Supports autochunking for large files
	Content []byte   `json:"content,omitempty"`
	
	// CreatedAt is the creation timestamp in nanoseconds since Unix epoch
	CreatedAt int64 `json:"created_at,omitempty"`
	
	// UpdatedAt is the last modification timestamp in nanoseconds since Unix epoch
	UpdatedAt int64 `json:"updated_at,omitempty"`
	
	// Performance optimization: cache for tag values
	// Maps tag key to latest value, built lazily on first access
	tagValueCache map[string]string `json:"-"`
	cacheValid    bool              `json:"-"`
	
	// Cache for cleaned tags (without timestamps)
	cleanTagsCache []string `json:"-"`
	cleanCacheValid bool    `json:"-"`
}


// EntityChange represents a single change event in an entity's history.
// Used for audit trails, history queries, and change tracking.
type EntityChange struct {
	// Type of change: "added", "modified", "removed"
	Type      string `json:"type"`
	
	// Timestamp when the change occurred (nanosecond epoch)
	Timestamp int64  `json:"timestamp"`
	
	// OldValue contains the previous value (for modifications and removals)
	OldValue  string `json:"old_value,omitempty"`
	
	// NewValue contains the new value (for additions and modifications)
	NewValue  string `json:"new_value,omitempty"`
	
	// EntityID references the entity this change belongs to
	EntityID  string `json:"entity_id,omitempty"`
}

// ChunkConfig configures the autochunking behavior for large content.
// When content exceeds the threshold, it's automatically split into chunks
// for efficient storage and retrieval.
type ChunkConfig struct {
	// DefaultChunkSize is the size of each chunk in bytes (default: 4MB)
	DefaultChunkSize   int64 `json:"default_chunk_size"`
	
	// AutoChunkThreshold is the minimum file size to trigger chunking (default: 4MB)
	AutoChunkThreshold int64 `json:"auto_chunk_threshold"`
}

// DefaultChunkConfig returns the default chunking configuration.
// Default values are optimized for typical file storage scenarios.
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		DefaultChunkSize:   4 * 1024 * 1024, // 4MB chunks
		AutoChunkThreshold: 4 * 1024 * 1024, // Chunk files > 4MB
	}
}

// NewEntity creates a new entity with an auto-generated UUID and initialized timestamps.
// The entity is ready for immediate use with pre-allocated tag storage.
//
// DEPRECATED: Use NewEntityWithMandatoryTags instead for new UUID-based architecture.
//
// Example:
//
//	entity := NewEntity()
//	entity.AddTag("type:document")
//	entity.Content = []byte("Hello, World!")
func NewEntity() *Entity {
	timestamp := Now()
	return &Entity{
		ID:        GenerateUUID(),
		Tags:      make([]string, 0, 8), // Pre-allocate for ~8 tags
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}
}

// NewEntityWithMandatoryTags creates a new entity with enforced UUID and mandatory tags.
// This is the preferred method for creating entities in the new architecture.
//
// Parameters:
//   - entityType: The type of entity (user, session, metric, etc.)
//   - dataset: The dataset namespace (system, default, etc.)
//   - createdByUUID: UUID of the user/system creating this entity
//   - additionalTags: Optional additional tags to include
//
// Returns a properly structured entity with:
//   - Pure UUID as ID (no prefixes)
//   - All mandatory tags (type, dataset, created_at, created_by, uuid)
//   - Proper timestamps
//
// Example:
//
//	entity, err := NewEntityWithMandatoryTags("user", "system", SystemUserID, 
//	    []string{"identity:username:john", "status:active"})
func NewEntityWithMandatoryTags(entityType, dataset, createdByUUID string, additionalTags []string) (*Entity, error) {
	// Generate mandatory tags and UUID
	mandatoryTags, entityUUID, err := GenerateMandatoryTags(entityType, dataset, createdByUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate mandatory tags: %v", err)
	}
	
	// Combine mandatory and additional tags
	allTags := make([]string, 0, len(mandatoryTags)+len(additionalTags))
	allTags = append(allTags, mandatoryTags...)
	allTags = append(allTags, additionalTags...)
	
	// Create entity with pure UUID
	timestamp := Now()
	entity := &Entity{
		ID:        entityUUID.Value, // Pure UUID, no prefix
		Tags:      allTags,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}
	
	return entity, nil
}

// GenerateUUID generates a unique identifier for entities.
// Uses SHA256 hash of current nanosecond timestamp for uniqueness.
//
// Note: For production use, consider using a proper UUID library
// like github.com/google/uuid for guaranteed uniqueness.
func GenerateUUID() string {
	timestamp := fmt.Sprintf("%d", Now())
	hash := sha256.Sum256([]byte(timestamp))
	return hex.EncodeToString(hash[:16])
}

// AddTag appends a tag to the entity with automatic timestamping.
// The tag is formatted as a temporal tag and interned for memory efficiency.
//
// Example:
//
//	entity.AddTag("status:active")
//	// Stored as: "2024-01-15T10:30:45.123456789|status:active"
func (e *Entity) AddTag(tag string) {
	// Format as temporal tag and intern for memory efficiency
	temporalTag := FormatTemporalTag(tag)
	e.Tags = append(e.Tags, Intern(temporalTag))
	// Invalidate cache since tags were modified
	e.invalidateTagValueCache()
}

// AddTagWithValue is a convenience method for adding namespace:value tags.
// Equivalent to AddTag(fmt.Sprintf("%s:%s", key, value)).
//
// Example:
//
//	entity.AddTagWithValue("status", "active")
//	// Same as: entity.AddTag("status:active")
func (e *Entity) AddTagWithValue(key, value string) {
	tag := fmt.Sprintf("%s:%s", key, value)
	e.AddTag(tag)
	// Cache invalidation is handled by AddTag
}

// GetTagsWithoutTimestamp returns all tags with their timestamps stripped.
// This is useful for tag comparison and display purposes.
//
// Handles multiple timestamp formats:
//   - "ISO|tag" (standard format)
//   - "ISO|NANO|tag" (double timestamp format)
//   - "NANO|tag" (numeric only format)
//   - "ISO.tag" (dot separator format)
//
// Example:
//
//	entity.Tags = []string{
//	    "2024-01-15T10:30:45.123456789|status:active",
//	    "2024-01-15T10:30:45.123456789|type:user",
//	}
//	cleanTags := entity.GetTagsWithoutTimestamp()
//	// cleanTags = ["status:active", "type:user"]
func (e *Entity) GetTagsWithoutTimestamp() []string {
	e.buildCleanTagsCache()
	// Return a copy to prevent external modification
	result := make([]string, len(e.cleanTagsCache))
	copy(result, e.cleanTagsCache)
	return result
}

// GetCurrentTags returns only the most recent value for each tag namespace
// This provides the current state of the entity by deduplicating tags
func (e *Entity) GetCurrentTags() []string {
	if len(e.Tags) == 0 {
		return []string{}
	}
	
	// Map to store the latest value for each tag namespace
	latestTags := make(map[string]tagWithTimestamp)
	
	for _, tag := range e.Tags {
		// Parse temporal tag format: "TIMESTAMP|tag:value"
		lastPipe := strings.LastIndex(tag, "|")
		if lastPipe >= 0 {
			timestampStr := tag[:lastPipe]
			actualTag := tag[lastPipe+1:]
			
			// Parse timestamp
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				// If timestamp parsing fails, treat as non-temporal tag
				latestTags[actualTag] = tagWithTimestamp{tag: actualTag, timestamp: 0}
				continue
			}
			
			// Extract tag namespace (everything before first colon)
			colonIndex := strings.Index(actualTag, ":")
			var namespace string
			if colonIndex >= 0 {
				namespace = actualTag[:colonIndex]
			} else {
				namespace = actualTag
			}
			
			// Keep only the latest timestamp for each namespace
			if existing, exists := latestTags[namespace]; !exists || timestamp > existing.timestamp {
				latestTags[namespace] = tagWithTimestamp{tag: actualTag, timestamp: timestamp}
			}
		} else {
			// Non-temporal tag - use as-is
			latestTags[tag] = tagWithTimestamp{tag: tag, timestamp: 0}
		}
	}
	
	// Extract just the tag values, sorted for consistency
	result := make([]string, 0, len(latestTags))
	for _, tagInfo := range latestTags {
		result = append(result, tagInfo.tag)
	}
	
	// Sort for consistent output
	sort.Strings(result)
	return result
}

// tagWithTimestamp is a helper struct for tag deduplication
type tagWithTimestamp struct {
	tag       string
	timestamp int64
}

// buildCleanTagsCache builds or rebuilds the clean tags cache
func (e *Entity) buildCleanTagsCache() {
	if e.cleanCacheValid && e.cleanTagsCache != nil {
		return // Cache is already valid
	}
	
	// Pre-allocate slice with known capacity
	e.cleanTagsCache = make([]string, 0, len(e.Tags))
	
	for _, tag := range e.Tags {
		// Fast path: find last pipe character
		lastPipe := strings.LastIndex(tag, "|")
		if lastPipe >= 0 {
			// Extract tag after timestamp
			actualTag := tag[lastPipe+1:]
			e.cleanTagsCache = append(e.cleanTagsCache, actualTag)
		} else {
			// No timestamp delimiter found, return as is
			e.cleanTagsCache = append(e.cleanTagsCache, tag)
		}
	}
	
	e.cleanCacheValid = true
}

// HasTag checks if the entity has a specific tag, ignoring timestamps.
// This method strips timestamps before comparison.
//
// Example:
//
//	if entity.HasTag("status:active") {
//	    // Entity is active
//	}
func (e *Entity) HasTag(tag string) bool {
	e.buildCleanTagsCache()
	for _, cleanTag := range e.cleanTagsCache {
		if cleanTag == tag {
			return true
		}
	}
	return false
}

// GetTagValue returns the most recent value for a given tag namespace.
// This method implements sophisticated temporal tag resolution with multiple timestamp formats.
//
// Algorithm:
// 1. Initialize tracking variables for latest value and timestamp
// 2. Iterate through all entity tags
// 3. For each tag:
//    a. Parse temporal format (TIMESTAMP|tag) to extract timestamp and clean tag
//    b. Support multiple timestamp formats: RFC3339Nano and epoch nanoseconds
//    c. Check if clean tag matches the requested namespace (key:*)
//    d. Extract value portion after the colon separator
//    e. Compare timestamp - keep only the most recent value
// 4. Return the value from the most recent timestamp, or empty string if none found
//
// Temporal Format Support:
//   - RFC3339Nano: "2024-01-01T10:00:00.000000000|status:active" 
//   - Epoch nanos: "1704110400000000000|status:active"
//   - Fallback: Skip malformed timestamps but continue processing
//
// Tag Structure: "TIMESTAMP|namespace:value"
//   - namespace: The tag category (e.g., "status", "type", "priority")
//   - value: The actual value (e.g., "active", "user", "high")
//
// Performance: O(n) where n is the number of tags on the entity
// Memory: O(1) additional space (only stores latest value)
//
// Example:
//	// Given temporal tags:
//	// "2024-01-01T10:00:00.000000000|status:draft"
//	// "2024-01-05T14:30:00.000000000|status:published" 
//	// "1704542400000000000|status:archived"
//	value := entity.GetTagValue("status")
//	// Returns: "archived" (most recent by timestamp)
func (e *Entity) GetTagValue(key string) string {
	// Use cached value if available
	e.buildTagValueCache()
	if value, exists := e.tagValueCache[key]; exists {
		return value
	}
	return ""
}

// buildTagValueCache builds or rebuilds the tag value cache for O(1) lookups
func (e *Entity) buildTagValueCache() {
	if e.cacheValid && e.tagValueCache != nil {
		return // Cache is already valid
	}
	
	// Initialize cache
	e.tagValueCache = make(map[string]string)
	
	// Track latest timestamp for each key
	latestTimestamp := make(map[string]int64)
	
	for _, tag := range e.Tags {
		parts := strings.Split(tag, "|")
		if len(parts) >= 2 {
			// Parse timestamp 
			timestampStr := parts[0]
			timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
			if err != nil {
				// Try parsing as epoch nanoseconds
				if nanos, err := strconv.ParseInt(timestampStr, 10, 64); err == nil {
					timestamp = time.Unix(0, nanos)
				} else {
					continue
				}
			}
			
			// Extract the actual tag
			actualTag := parts[len(parts)-1]
			if colonIndex := strings.Index(actualTag, ":"); colonIndex > 0 {
				key := actualTag[:colonIndex]
				value := actualTag[colonIndex+1:]
				
				timestampNanos := timestamp.UnixNano()
				if existingTimestamp, exists := latestTimestamp[key]; !exists || timestampNanos > existingTimestamp {
					latestTimestamp[key] = timestampNanos
					e.tagValueCache[key] = value
				}
			}
		}
	}
	
	e.cacheValid = true
}

// invalidateTagValueCache invalidates all caches when tags are modified
func (e *Entity) invalidateTagValueCache() {
	e.cacheValid = false
	e.tagValueCache = nil
	e.cleanCacheValid = false
	e.cleanTagsCache = nil
}

// SetTags replaces all tags and invalidates the cache (used by storage layer)
func (e *Entity) SetTags(tags []string) {
	e.Tags = tags
	e.invalidateTagValueCache()
}

// AppendTag adds a tag without timestamp formatting (used by storage layer for pre-formatted tags)
func (e *Entity) AppendTag(tag string) {
	e.Tags = append(e.Tags, tag)
	e.invalidateTagValueCache()
}

// SetContent sets content with automatic chunking if needed
func (e *Entity) SetContent(reader io.Reader, mimeType string, config ChunkConfig) ([]string, error) {
	// First, determine the size
	var totalSize int64
	var chunks [][]byte
	
	// Read all data to determine size and chunks
	for {
		chunk := make([]byte, config.DefaultChunkSize)
		n, err := reader.Read(chunk)
		if n > 0 {
			chunks = append(chunks, chunk[:n])
			totalSize += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	
	// Calculate full content hash
	hasher := sha256.New()
	for _, chunk := range chunks {
		hasher.Write(chunk)
	}
	contentHash := hex.EncodeToString(hasher.Sum(nil))
	
	// Add content metadata tags
	e.AddTag(fmt.Sprintf("content:type:%s", mimeType))
	e.AddTag(fmt.Sprintf("content:size:%d", totalSize))
	e.AddTag(fmt.Sprintf("content:checksum:sha256:%s", contentHash))
	
	// Determine if we need to chunk
	if totalSize <= config.AutoChunkThreshold {
		// Small file - store directly
		e.Content = chunks[0] // Only one chunk for small files
		return nil, nil // No chunk entities needed
	}
	
	// Large file - create chunk entities
	e.AddTag(fmt.Sprintf("content:chunks:%d", len(chunks)))
	e.AddTag(fmt.Sprintf("content:chunk-size:%d", config.DefaultChunkSize))
	e.Content = nil // Master entity has no content
	
	// Create chunk entity IDs
	chunkIDs := make([]string, len(chunks))
	for i := range chunks {
		chunkIDs[i] = fmt.Sprintf("%s-chunk-%d", e.ID, i)
	}
	
	return chunkIDs, nil
}

// CreateChunkEntity creates a chunk entity
func CreateChunkEntity(parentID string, chunkIndex int, data []byte) *Entity {
	entity := NewEntity()
	entity.ID = fmt.Sprintf("%s-chunk-%d", parentID, chunkIndex)
	entity.Tags = []string{
		"type:chunk",
		fmt.Sprintf("parent:%s", parentID),
		fmt.Sprintf("content:chunk:%d", chunkIndex),
		fmt.Sprintf("content:size:%d", len(data)),
		fmt.Sprintf("content:checksum:sha256:%s", calculateChecksum(data)),
	}
	entity.Content = data
	return entity
}

// IsChunked returns true if this entity has chunked content
func (e *Entity) IsChunked() bool {
	// First, strip timestamps to get clean tags
	tags := e.GetTagsWithoutTimestamp()
	for _, tag := range tags {
		if startsWith(tag, "content:chunks:") {
			// Additional validation - check if the chunk count is a valid number > 0
			parts := strings.SplitN(tag, ":", 3)
			if len(parts) == 3 {
				chunksStr := parts[2]
				if chunks, err := strconv.Atoi(chunksStr); err == nil && chunks > 0 {
					return true
				}
			}
		}
	}
	return false
}

// GetContentMetadata extracts content metadata from tags
func (e *Entity) GetContentMetadata() map[string]string {
	metadata := make(map[string]string)
	for _, tag := range e.Tags {
		if startsWith(tag, "content:") {
			parts := parseTag(tag)
			if len(parts) >= 3 {
				metadata[parts[1]] = parts[2]
			}
		}
	}
	return metadata
}

// GetContentValue returns the value of a content item by type (for compatibility)
func (e *Entity) GetContentValue(contentType string) string {
	// For new model, this would look in tags
	for _, tag := range e.GetTagsWithoutTimestamp() {
		if startsWith(tag, contentType+":") {
			parts := parseTag(tag)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}

// Helper functions
func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func parseTag(tag string) []string {
	// Simple tag parser: "content:type:text/plain" -> ["content", "type", "text/plain"]
	result := []string{}
	start := 0
	for i := 0; i <= len(tag); i++ {
		if i == len(tag) || tag[i] == ':' {
			result = append(result, tag[start:i])
			start = i + 1
		}
	}
	return result
}

// ValidateEntity performs comprehensive validation on an entity
// Returns error if the entity doesn't meet the new UUID-based architecture requirements
func (e *Entity) ValidateEntity() error {
	// Validate UUID
	if err := ValidateEntityUUID(e.ID); err != nil {
		return fmt.Errorf("invalid entity ID: %v", err)
	}
	
	// Validate mandatory tags
	_, err := ValidateMandatoryTags(e.Tags)
	if err != nil {
		return fmt.Errorf("mandatory tag validation failed: %v", err)
	}
	
	return nil
}

// GetMandatoryTags extracts and validates mandatory tags from the entity
func (e *Entity) GetMandatoryTags() (*MandatoryTags, error) {
	return ValidateMandatoryTags(e.Tags)
}

// IsSystemEntity checks if this entity was created by the system user
func (e *Entity) IsSystemEntity() bool {
	mandatory, err := e.GetMandatoryTags()
	if err != nil {
		return false
	}
	return IsSystemUser(mandatory.CreatedBy)
}

// GetEntityUUID returns the entity's UUID (same as ID in new architecture)
func (e *Entity) GetEntityUUID() string {
	return e.ID
}

// GetEntityType extracts the entity type from mandatory tags
func (e *Entity) GetEntityType() string {
	mandatory, err := e.GetMandatoryTags()
	if err != nil {
		// Fallback to legacy method
		return e.GetTagValue("type")
	}
	return mandatory.Type
}

// GetDataset extracts the dataset from mandatory tags
func (e *Entity) GetDataset() string {
	mandatory, err := e.GetMandatoryTags()
	if err != nil {
		// Fallback to legacy method
		return e.GetTagValue("dataset")
	}
	return mandatory.Dataset
}

// GetCreatedBy returns the UUID of the user/system that created this entity
func (e *Entity) GetCreatedBy() string {
	mandatory, err := e.GetMandatoryTags()
	if err != nil {
		// Fallback to legacy method
		return e.GetTagValue("created_by")
	}
	return mandatory.CreatedBy
}

// ===== ENTITY LIFECYCLE MANAGEMENT =====

// Lifecycle returns a lifecycle manager for this entity
func (e *Entity) Lifecycle() *EntityLifecycle {
	return NewEntityLifecycle(e)
}

// GetLifecycleState returns the current lifecycle state of the entity
func (e *Entity) GetLifecycleState() EntityLifecycleState {
	return e.Lifecycle().GetCurrentState()
}

// IsActive returns true if the entity is in active state
func (e *Entity) IsActive() bool {
	return e.Lifecycle().IsActive()
}

// IsSoftDeleted returns true if the entity is in soft deleted state
func (e *Entity) IsSoftDeleted() bool {
	return e.Lifecycle().IsSoftDeleted()
}

// IsArchived returns true if the entity is in archived state
func (e *Entity) IsArchived() bool {
	return e.Lifecycle().IsArchived()
}

// IsPurged returns true if the entity is in purged state
func (e *Entity) IsPurged() bool {
	return e.Lifecycle().IsPurged()
}

// CanTransitionTo checks if a transition to the target state is valid
func (e *Entity) CanTransitionTo(targetState EntityLifecycleState) bool {
	return e.Lifecycle().CanTransitionTo(targetState)
}

// SoftDelete transitions entity to soft deleted state
func (e *Entity) SoftDelete(userID, reason, policy string) error {
	return e.Lifecycle().SoftDelete(userID, reason, policy)
}

// Undelete transitions entity back to active state
func (e *Entity) Undelete(userID, reason string) error {
	return e.Lifecycle().Undelete(userID, reason)
}

// Archive transitions entity to archived state
func (e *Entity) Archive(userID, reason, policy string) error {
	return e.Lifecycle().Archive(userID, reason, policy)
}

// Purge transitions entity to purged state
func (e *Entity) Purge(userID, reason, policy string) error {
	return e.Lifecycle().Purge(userID, reason, policy)
}

// GetTransitionHistory returns all lifecycle transitions for this entity
func (e *Entity) GetTransitionHistory() []LifecycleTransition {
	return e.Lifecycle().GetTransitionHistory()
}

// GetDeletedAt returns when the entity was deleted (if it has been deleted)
func (e *Entity) GetDeletedAt() *time.Time {
	return e.Lifecycle().GetDeletedAt()
}

// GetArchivedAt returns when the entity was archived (if it has been archived)
func (e *Entity) GetArchivedAt() *time.Time {
	return e.Lifecycle().GetArchivedAt()
}

// GetDeletionPolicy returns the deletion policy applied to this entity
func (e *Entity) GetDeletionPolicy() string {
	return e.Lifecycle().GetDeletionPolicy()
}

// GetDeletedBy returns who deleted the entity
func (e *Entity) GetDeletedBy() string {
	return e.Lifecycle().GetDeletedBy()
}

// GetDeleteReason returns why the entity was deleted
func (e *Entity) GetDeleteReason() string {
	return e.Lifecycle().GetDeleteReason()
}

