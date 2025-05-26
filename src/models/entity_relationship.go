package models

import (
	"encoding/json"
	"strings"
)

// EntityRelationship represents a typed relationship between two entities
type EntityRelationship struct {
	ID               string            `json:"id,omitempty"`
	SourceID         string            `json:"source_id"`
	RelationshipType string            `json:"relationship_type"`
	Type             string            `json:"type"` // Alias for RelationshipType for consistency
	TargetID         string            `json:"target_id"`
	Properties       map[string]string `json:"properties,omitempty"` // Simple key-value properties
	CreatedAt        int64             `json:"created_at"`           // Nanosecond epoch timestamp
	UpdatedAt        int64             `json:"updated_at,omitempty"`
	CreatedBy        string            `json:"created_by,omitempty"`
	UpdatedBy        string            `json:"updated_by,omitempty"`
	Metadata         string            `json:"metadata,omitempty"` // JSON string for additional data
}

// Common relationship types
const (
	RelationshipTypeDependsOn     = "depends_on"
	RelationshipTypeBlocks        = "blocks"
	RelationshipTypeParentOf      = "parent_of"
	RelationshipTypeChildOf       = "child_of"
	RelationshipTypeRelatedTo     = "related_to"
	RelationshipTypeDuplicateOf   = "duplicate_of"
	RelationshipTypeAssignedTo    = "assigned_to"
	RelationshipTypeBelongsTo     = "belongs_to"
	RelationshipTypeCreatedBy     = "created_by"
	RelationshipTypeUpdatedBy     = "updated_by"
	RelationshipTypeLinkedTo      = "linked_to"
	
	// Security relationship types
	RelationshipTypeHasCredential   = "has_credential"
	RelationshipTypeAuthenticatedAs = "authenticated_as"
	RelationshipTypeMemberOf        = "member_of"
	RelationshipTypeHasRole         = "has_role"
	RelationshipTypeGrants          = "grants"
	RelationshipTypeOwns            = "owns"
	RelationshipTypeCanAccess       = "can_access"
)

// NewEntityRelationship creates a new entity relationship
func NewEntityRelationship(sourceID, relationshipType, targetID string) *EntityRelationship {
	now := Now()
	return &EntityRelationship{
		ID:               sourceID + "_" + relationshipType + "_" + targetID,
		SourceID:         sourceID,
		RelationshipType: relationshipType,
		Type:             relationshipType, // Set both for compatibility
		TargetID:         targetID,
		Properties:       make(map[string]string),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// SetCreatedBy sets the creator of the relationship
func (r *EntityRelationship) SetCreatedBy(userID string) {
	r.CreatedBy = userID
}

// AddMetadata sets metadata for the relationship from a map
func (r *EntityRelationship) AddMetadata(metadata map[string]interface{}) error {
	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	r.Metadata = string(data)
	return nil
}

// GetMetadata retrieves the metadata as a map
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

// ParseRelationshipID parses a relationship ID into its component parts
// Assumes the format: sourceID_relationshipType_targetID
func ParseRelationshipID(id string) []string {
	return strings.Split(id, "_")
}

// EntityRelationshipRepository defines the interface for entity relationship persistence
type EntityRelationshipRepository interface {
	// Create creates a new entity relationship
	Create(relationship *EntityRelationship) error
	
	// Delete removes an entity relationship
	Delete(sourceID, relationshipType, targetID string) error
	
	// GetBySource gets all relationships where entity is the source
	GetBySource(sourceID string) ([]*EntityRelationship, error)
	
	// GetBySourceAndType gets all relationships of a given type where entity is the source
	GetBySourceAndType(sourceID, relationshipType string) ([]*EntityRelationship, error)
	
	// GetByTarget gets all relationships where entity is the target
	GetByTarget(targetID string) ([]*EntityRelationship, error)
	
	// GetByTargetAndType gets all relationships of a given type where entity is the target
	GetByTargetAndType(targetID, relationshipType string) ([]*EntityRelationship, error)
	
	// GetByType gets all relationships of a specific type
	GetByType(relationshipType string) ([]*EntityRelationship, error)
	
	// GetRelationship gets a specific relationship
	GetRelationship(sourceID, relationshipType, targetID string) (*EntityRelationship, error)
	
	// Exists checks if a relationship exists
	Exists(sourceID, relationshipType, targetID string) (bool, error)
}