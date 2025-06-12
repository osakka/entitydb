// Package models provides core data structures and business logic for EntityDB.
// This file implements entity relationships for modeling connections between entities.
package models

import (
	"encoding/json"
	"strings"
)

// EntityRelationship represents a typed, directed relationship between two entities.
// Relationships enable complex data modeling by linking entities with semantic meaning.
// Each relationship has a source entity, target entity, and a type that describes
// the nature of the connection.
//
// Relationships support:
//   - Directional connections (source → target)
//   - Typed relationships with semantic meaning
//   - Simple key-value properties
//   - Rich metadata as JSON
//   - Audit trails (created/updated by and timestamps)
//
// Example usage:
//
//	// Create a dependency relationship
//	rel := NewEntityRelationship("task-123", "depends_on", "task-456")
//	rel.Properties["priority"] = "high"
//	rel.AddMetadata(map[string]interface{}{
//	    "notes": "Must complete before starting",
//	    "deadline": "2023-12-31",
//	})
//
// Common patterns:
//   - Task dependencies: task A depends_on task B
//   - Hierarchies: entity A parent_of entity B
//   - Assignments: task assigned_to user
//   - Security: user has_credential credential
type EntityRelationship struct {
	// ID is the unique identifier for this relationship.
	// Format: "sourceID_relationshipType_targetID"
	ID string `json:"id,omitempty"`
	
	// SourceID is the ID of the entity that originates the relationship.
	// In a dependency "A depends_on B", A is the source.
	SourceID string `json:"source_id"`
	
	// RelationshipType describes the nature of the relationship.
	// Use predefined constants or custom types as needed.
	RelationshipType string `json:"relationship_type"`
	
	// Type is an alias for RelationshipType for API consistency.
	// Both fields contain the same value.
	Type string `json:"type"`
	
	// TargetID is the ID of the entity that receives the relationship.
	// In a dependency "A depends_on B", B is the target.
	TargetID string `json:"target_id"`
	
	// Properties contains simple key-value attributes for the relationship.
	// Use this for lightweight metadata like priority, status, or labels.
	Properties map[string]string `json:"properties,omitempty"`
	
	// CreatedAt is the nanosecond epoch timestamp when the relationship was created
	CreatedAt int64 `json:"created_at"`
	
	// UpdatedAt is the nanosecond epoch timestamp when the relationship was last modified
	UpdatedAt int64 `json:"updated_at,omitempty"`
	
	// CreatedBy is the ID of the user who created this relationship
	CreatedBy string `json:"created_by,omitempty"`
	
	// UpdatedBy is the ID of the user who last modified this relationship
	UpdatedBy string `json:"updated_by,omitempty"`
	
	// Metadata contains rich, structured data as a JSON string.
	// Use this for complex attributes that don't fit in simple Properties.
	Metadata string `json:"metadata,omitempty"`
}

// Common relationship types for typical use cases.
// These constants provide standardized relationship semantics.
const (
	// Dependency and workflow relationships
	RelationshipTypeDependsOn   = "depends_on"   // A depends_on B (A cannot proceed without B)
	RelationshipTypeBlocks      = "blocks"       // A blocks B (A prevents B from proceeding)
	RelationshipTypeParentOf    = "parent_of"    // A parent_of B (hierarchical relationship)
	RelationshipTypeChildOf     = "child_of"     // A child_of B (inverse of parent_of)
	RelationshipTypeRelatedTo   = "related_to"   // A related_to B (general association)
	RelationshipTypeDuplicateOf = "duplicate_of" // A duplicate_of B (A is a copy of B)
	
	// Assignment and ownership relationships
	RelationshipTypeAssignedTo = "assigned_to" // A assigned_to B (A is assigned to person/entity B)
	RelationshipTypeBelongsTo  = "belongs_to"  // A belongs_to B (A is part of group/category B)
	RelationshipTypeCreatedBy  = "created_by"  // A created_by B (B created A)
	RelationshipTypeUpdatedBy  = "updated_by"  // A updated_by B (B last modified A)
	RelationshipTypeLinkedTo   = "linked_to"   // A linked_to B (general bidirectional link)
	
	// Security and authentication relationships
	RelationshipTypeHasCredential   = "has_credential"   // User has_credential credential_entity
	RelationshipTypeAuthenticatedAs = "authenticated_as" // Session authenticated_as user
	RelationshipTypeMemberOf        = "member_of"        // User member_of group/role
	RelationshipTypeHasRole         = "has_role"         // User has_role role_entity
	RelationshipTypeGrants          = "grants"           // Role grants permission
	RelationshipTypeOwns            = "owns"             // User owns resource
	RelationshipTypeCanAccess       = "can_access"       // User can_access resource
)

// NewEntityRelationship creates a new entity relationship with the specified parameters.
// The relationship ID is automatically generated using the format "sourceID_type_targetID".
// Timestamps are set to the current time, and an empty Properties map is initialized.
//
// Parameters:
//   - sourceID: The ID of the source entity (relationship origin)
//   - relationshipType: The type of relationship (use constants for standard types)
//   - targetID: The ID of the target entity (relationship destination)
//
// Returns:
//   - *EntityRelationship: A new relationship ready for use
//
// Example:
//
//	// Create a task dependency
//	rel := NewEntityRelationship("task-123", RelationshipTypeDependsOn, "task-456")
//	rel.Properties["severity"] = "critical"
//	
//	// Create a user-role assignment
//	roleRel := NewEntityRelationship("user-789", RelationshipTypeHasRole, "role-admin")
func NewEntityRelationship(sourceID, relationshipType, targetID string) *EntityRelationship {
	now := Now()
	return &EntityRelationship{
		ID:               sourceID + "_" + relationshipType + "_" + targetID,
		SourceID:         sourceID,
		RelationshipType: relationshipType,
		Type:             relationshipType, // Set both for API compatibility
		TargetID:         targetID,
		Properties:       make(map[string]string),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// SetCreatedBy sets the creator of the relationship for audit purposes.
// This tracks which user created the relationship for accountability and debugging.
//
// Parameters:
//   - userID: The ID of the user who created this relationship
//
// Example:
//
//	rel := NewEntityRelationship("task-1", "assigned_to", "user-123")
//	rel.SetCreatedBy("admin-user")
func (r *EntityRelationship) SetCreatedBy(userID string) {
	r.CreatedBy = userID
}

// AddMetadata sets rich metadata for the relationship from a map structure.
// The metadata is serialized to JSON and stored in the Metadata field.
// This is useful for complex attributes that don't fit in simple Properties.
//
// Parameters:
//   - metadata: A map containing the metadata to attach to the relationship
//
// Returns:
//   - error: Any error encountered during JSON serialization
//
// Example:
//
//	rel := NewEntityRelationship("task-1", "depends_on", "task-2")
//	err := rel.AddMetadata(map[string]interface{}{
//	    "notes": "Critical dependency for sprint completion",
//	    "estimated_delay_hours": 24,
//	    "stakeholders": []string{"product-manager", "tech-lead"},
//	})
//	if err != nil {
//	    log.Printf("Failed to add metadata: %v", err)
//	}
func (r *EntityRelationship) AddMetadata(metadata map[string]interface{}) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	r.Metadata = string(data)
	return nil
}

// GetMetadata retrieves the relationship metadata as a structured map.
// If no metadata is set, returns an empty map. The JSON metadata is
// parsed back into a map[string]interface{} structure.
//
// Returns:
//   - map[string]interface{}: The metadata as a structured map
//   - error: Any error encountered during JSON parsing
//
// Example:
//
//	metadata, err := rel.GetMetadata()
//	if err != nil {
//	    log.Printf("Failed to parse metadata: %v", err)
//	    return
//	}
//	
//	if notes, exists := metadata["notes"]; exists {
//	    fmt.Printf("Relationship notes: %s\n", notes)
//	}
func (r *EntityRelationship) GetMetadata() (map[string]interface{}, error) {
	if r.Metadata == "" {
		return make(map[string]interface{}), nil
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(r.Metadata), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ParseRelationshipID parses a relationship ID into its component parts.
// Relationship IDs follow the format "sourceID_relationshipType_targetID".
// This function splits the ID and returns the components as a slice.
//
// Parameters:
//   - id: The relationship ID to parse
//
// Returns:
//   - []string: A slice containing [sourceID, relationshipType, targetID]
//
// Example:
//
//	id := "task-123_depends_on_task-456"
//	parts := ParseRelationshipID(id)
//	// parts[0] = "task-123" (source)
//	// parts[1] = "depends_on" (type)
//	// parts[2] = "task-456" (target)
//
// Note: This function assumes the relationship ID was generated using the standard format.
// If the ID doesn't follow this format, the results may be unexpected.
func ParseRelationshipID(id string) []string {
	return strings.Split(id, "_")
}

// EntityRelationshipRepository defines the interface for entity relationship persistence.
// This interface abstracts the storage layer for relationships, allowing different
// implementations (binary, in-memory, SQL, etc.) while maintaining a consistent API.
//
// The repository supports:
//   - CRUD operations for relationships
//   - Querying by source, target, or relationship type
//   - Efficient lookups for graph traversal
//   - Existence checks for validation
//
// Implementation considerations:
//   - Index by source ID for outbound relationship queries
//   - Index by target ID for inbound relationship queries
//   - Index by relationship type for type-specific queries
//   - Support for compound queries (source + type, target + type)
type EntityRelationshipRepository interface {
	// Create persists a new entity relationship to storage.
	// The relationship ID should be unique; creating a duplicate may result in an error.
	//
	// Parameters:
	//   - relationship: The relationship to persist
	//
	// Returns:
	//   - error: Any error encountered during persistence
	//
	// Example:
	//
	//	rel := NewEntityRelationship("task-1", "depends_on", "task-2")
	//	if err := repo.Create(rel); err != nil {
	//	    log.Printf("Failed to create relationship: %v", err)
	//	}
	Create(relationship *EntityRelationship) error
	
	// Delete removes an entity relationship from storage.
	// If the relationship doesn't exist, this operation should be idempotent (no error).
	//
	// Parameters:
	//   - sourceID: The source entity ID
	//   - relationshipType: The type of relationship
	//   - targetID: The target entity ID
	//
	// Returns:
	//   - error: Any error encountered during deletion
	//
	// Example:
	//
	//	err := repo.Delete("task-1", "depends_on", "task-2")
	Delete(sourceID, relationshipType, targetID string) error
	
	// GetBySource retrieves all relationships where the specified entity is the source.
	// This is useful for finding all outbound relationships from an entity.
	//
	// Parameters:
	//   - sourceID: The ID of the source entity
	//
	// Returns:
	//   - []*EntityRelationship: All relationships originating from the entity
	//   - error: Any error encountered during retrieval
	//
	// Example:
	//
	//	// Find all tasks that depend on task-1
	//	relationships, err := repo.GetBySource("task-1")
	GetBySource(sourceID string) ([]*EntityRelationship, error)
	
	// GetBySourceAndType retrieves relationships of a specific type from a source entity.
	// This is more efficient than GetBySource when you only need certain relationship types.
	//
	// Parameters:
	//   - sourceID: The ID of the source entity
	//   - relationshipType: The type of relationships to retrieve
	//
	// Returns:
	//   - []*EntityRelationship: Matching relationships
	//   - error: Any error encountered during retrieval
	//
	// Example:
	//
	//	// Find all tasks assigned to users from task-1
	//	assignments, err := repo.GetBySourceAndType("task-1", "assigned_to")
	GetBySourceAndType(sourceID, relationshipType string) ([]*EntityRelationship, error)
	
	// GetByTarget retrieves all relationships where the specified entity is the target.
	// This is useful for finding all inbound relationships to an entity.
	//
	// Parameters:
	//   - targetID: The ID of the target entity
	//
	// Returns:
	//   - []*EntityRelationship: All relationships pointing to the entity
	//   - error: Any error encountered during retrieval
	//
	// Example:
	//
	//	// Find all entities that depend on task-2
	//	dependencies, err := repo.GetByTarget("task-2")
	GetByTarget(targetID string) ([]*EntityRelationship, error)
	
	// GetByTargetAndType retrieves relationships of a specific type to a target entity.
	// This is more efficient than GetByTarget when you only need certain relationship types.
	//
	// Parameters:
	//   - targetID: The ID of the target entity
	//   - relationshipType: The type of relationships to retrieve
	//
	// Returns:
	//   - []*EntityRelationship: Matching relationships
	//   - error: Any error encountered during retrieval
	//
	// Example:
	//
	//	// Find all tasks that depend on task-2
	//	deps, err := repo.GetByTargetAndType("task-2", "depends_on")
	GetByTargetAndType(targetID, relationshipType string) ([]*EntityRelationship, error)
	
	// GetByType retrieves all relationships of a specific type across the entire system.
	// This is useful for analyzing patterns or auditing specific relationship types.
	//
	// Parameters:
	//   - relationshipType: The type of relationships to retrieve
	//
	// Returns:
	//   - []*EntityRelationship: All relationships of the specified type
	//   - error: Any error encountered during retrieval
	//
	// Example:
	//
	//	// Find all dependency relationships in the system
	//	allDeps, err := repo.GetByType("depends_on")
	GetByType(relationshipType string) ([]*EntityRelationship, error)
	
	// GetRelationship retrieves a specific relationship by its components.
	// This is useful for checking relationship details or verifying existence with data.
	//
	// Parameters:
	//   - sourceID: The source entity ID
	//   - relationshipType: The relationship type
	//   - targetID: The target entity ID
	//
	// Returns:
	//   - *EntityRelationship: The relationship if found, nil if not found
	//   - error: Any error encountered during retrieval (not including "not found")
	//
	// Example:
	//
	//	rel, err := repo.GetRelationship("task-1", "depends_on", "task-2")
	//	if err != nil {
	//	    log.Printf("Error retrieving relationship: %v", err)
	//	} else if rel == nil {
	//	    log.Println("Relationship not found")
	//	} else {
	//	    log.Printf("Found relationship created at %d", rel.CreatedAt)
	//	}
	GetRelationship(sourceID, relationshipType, targetID string) (*EntityRelationship, error)
	
	// Exists checks if a specific relationship exists without retrieving its data.
	// This is more efficient than GetRelationship when you only need to verify existence.
	//
	// Parameters:
	//   - sourceID: The source entity ID
	//   - relationshipType: The relationship type
	//   - targetID: The target entity ID
	//
	// Returns:
	//   - bool: true if the relationship exists, false otherwise
	//   - error: Any error encountered during the check
	//
	// Example:
	//
	//	exists, err := repo.Exists("task-1", "depends_on", "task-2")
	//	if err != nil {
	//	    log.Printf("Error checking relationship: %v", err)
	//	} else if exists {
	//	    log.Println("Dependency already exists")
	//	}
	Exists(sourceID, relationshipType, targetID string) (bool, error)
}