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
	"strconv"
	"strings"
	"time"
)

// EntityRepository defines the contract for entity storage implementations.
// All storage backends must implement this interface to provide entity
// persistence with temporal support.
type EntityRepository interface {
	// Core CRUD operations
	Create(entity *Entity) error
	GetByID(id string) (*Entity, error)
	Update(entity *Entity) error
	Delete(id string) error
	List() ([]*Entity, error)
	
	// Tag-based queries
	ListByTag(tag string) ([]*Entity, error)
	ListByTags(tags []string, matchAll bool) ([]*Entity, error)
	ListByTagSQL(tag string) ([]*Entity, error)
	ListByTagWildcard(pattern string) ([]*Entity, error)
	ListByNamespace(namespace string) ([]*Entity, error)
	
	// Content queries
	SearchContent(query string) ([]*Entity, error)
	SearchContentByType(contentType string) ([]*Entity, error)
	
	// Advanced query
	QueryAdvanced(params map[string]interface{}) ([]*Entity, error)
	
	// Transaction support
	Transaction(fn func(tx interface{}) error) error
	Commit(tx interface{}) error
	Rollback(tx interface{}) error
	
	// Tag operations
	AddTag(id string, tag string) error
	RemoveTag(id string, tag string) error
	
	// Temporal operations
	GetEntityAsOf(id string, timestamp time.Time) (*Entity, error)
	GetEntityHistory(id string, limit int) ([]*EntityChange, error)
	GetRecentChanges(limit int) ([]*EntityChange, error)
	GetEntityDiff(id string, startTime, endTime time.Time) (*Entity, *Entity, error)
	
	// Relationship operations
	CreateRelationship(rel interface{}) error
	GetRelationshipByID(id string) (interface{}, error)
	GetRelationshipsBySource(sourceID string) ([]interface{}, error)
	GetRelationshipsByTarget(targetID string) ([]interface{}, error)
	DeleteRelationship(id string) error
	
	// Query builder
	Query() *EntityQuery
	
	// Maintenance operations
	ReindexTags() error
	VerifyIndexHealth() error
}

// Entity - The ONE entity type we need
type Entity struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	Content []byte   `json:"content,omitempty"`
	
	// Timestamps as nanosecond epoch for efficiency
	CreatedAt int64 `json:"created_at,omitempty"`
	UpdatedAt int64 `json:"updated_at,omitempty"`
}

// ContentItem for backward compatibility - will be removed
// Deprecated: Use Content []byte directly
type ContentItem struct {
	Timestamp int64  `json:"timestamp"` // Now using nanosecond epoch
	Type      string `json:"type"`
	Value     string `json:"value"`
}

// EntityChange represents a change to an entity
type EntityChange struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"` // Nanosecond epoch for consistency
	OldValue  string `json:"old_value,omitempty"`
	NewValue  string `json:"new_value,omitempty"`
	EntityID  string `json:"entity_id,omitempty"`  // Reference to the entity this change belongs to
}

// ChunkConfig configures autochunking behavior
type ChunkConfig struct {
	DefaultChunkSize   int64 `json:"default_chunk_size"`   // Default: 4MB
	AutoChunkThreshold int64 `json:"auto_chunk_threshold"` // Files > this get chunked
}

// DefaultChunkConfig returns sensible defaults
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		DefaultChunkSize:   4 * 1024 * 1024, // 4MB
		AutoChunkThreshold: 4 * 1024 * 1024, // Auto-chunk files > 4MB
	}
}

// NewEntity creates a new entity with auto-generated UUID
func NewEntity() *Entity {
	timestamp := Now()
	return &Entity{
		ID:        GenerateUUID(),
		Tags:      make([]string, 0, 8), // Pre-allocate common size
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}
}

// GenerateUUID generates a unique identifier for entities
func GenerateUUID() string {
	// Simple implementation - in production use crypto/rand
	timestamp := fmt.Sprintf("%d", Now())
	hash := sha256.Sum256([]byte(timestamp))
	return hex.EncodeToString(hash[:16])
}

// AddTag adds a tag with automatic timestamping and interning
func (e *Entity) AddTag(tag string) {
	// Use our centralized temporal tag formatting and intern the result
	temporalTag := FormatTemporalTag(tag)
	e.Tags = append(e.Tags, Intern(temporalTag))
}

// AddTagWithValue adds a key:value tag with automatic timestamping  
func (e *Entity) AddTagWithValue(key, value string) {
	tag := fmt.Sprintf("%s:%s", key, value)
	e.AddTag(tag)
}

// GetTagsWithoutTimestamp returns tags without their timestamp prefix
func (e *Entity) GetTagsWithoutTimestamp() []string {
	result := []string{}
	for _, tag := range e.Tags {
		// Handle multiple timestamp formats:
		// 1. ISO|tag (standard format)
		// 2. ISO|NANO|tag (double timestamp format from temporal repository)
		// 3. NANO|tag (numeric only format)
		
		parts := strings.Split(tag, "|")
		if len(parts) >= 2 {
			// Take the last part as the tag value
			actualTag := parts[len(parts)-1]
			result = append(result, actualTag)
		} else {
			// No timestamp delimiter found, return as is
			result = append(result, tag)
		}
	}
	return result
}

// HasTag checks if entity has a tag (checks without timestamp)
func (e *Entity) HasTag(tag string) bool {
	cleanTags := e.GetTagsWithoutTimestamp()
	for _, cleanTag := range cleanTags {
		if cleanTag == tag {
			return true
		}
	}
	return false
}

// GetTagValue returns the most recent value for a tag key (gets latest value with timestamp)
func (e *Entity) GetTagValue(key string) string {
	// Look for the most recent tag with the given key prefix
	var latestValue string
	var latestTimestamp int64
	
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
			
			// Check if this tag matches our key
			actualTag := parts[len(parts)-1]
			if strings.HasPrefix(actualTag, key+":") {
				tagParts := strings.SplitN(actualTag, ":", 2)
				if len(tagParts) == 2 && tagParts[0] == key {
					timestampNanos := timestamp.UnixNano()
					if timestampNanos > latestTimestamp {
						latestTimestamp = timestampNanos
						latestValue = tagParts[1]
					}
				}
			}
		}
	}
	return latestValue
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

