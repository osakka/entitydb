# Entity Relationship Implementation

This document describes the implementation of entity relationships in the EntityDB platform.

## Overview

Entity relationships provide a flexible way to represent connections between entities in the system. Unlike the traditional issue dependency model, entity relationships are:

1. **Typed**: Each relationship has a specific type (e.g., `depends_on`, `blocks`, `parent_of`, etc.)
2. **Bidirectional**: Relationships can be queried from both source and target perspectives
3. **Metadata-rich**: Additional information can be attached to relationships via metadata
4. **Generic**: Can connect any two entities, not just issues

This implementation extends the entity-issue integration to use entity relationships for representing dependencies between issues.

## Data Model

The core of the implementation is the `EntityRelationship` struct:

```go
// EntityRelationship represents a typed relationship between two entities
type EntityRelationship struct {
    SourceID         string    `json:"source_id"`
    RelationshipType string    `json:"relationship_type"`
    TargetID         string    `json:"target_id"`
    CreatedAt        time.Time `json:"created_at"`
    CreatedBy        string    `json:"created_by,omitempty"`
    Metadata         string    `json:"metadata,omitempty"` // JSON string for additional data
}
```

The three key fields (SourceID, RelationshipType, and TargetID) form the composite primary key, ensuring uniqueness of relationships.

## Common Relationship Types

The system defines several standard relationship types:

```go
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
)
```

## Repository Interface

The `EntityRelationshipRepository` interface provides methods for working with relationships:

```go
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
```

## Database Schema

Entity relationships are stored in the `entity_relationships` table with the following schema:

```sql
CREATE TABLE IF NOT EXISTS entity_relationships (
    source_id TEXT NOT NULL,
    relationship_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    created_by TEXT,
    metadata TEXT,
    PRIMARY KEY (source_id, relationship_type, target_id)
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_entity_relationships_source ON entity_relationships (source_id);
CREATE INDEX IF NOT EXISTS idx_entity_relationships_target ON entity_relationships (target_id);
CREATE INDEX IF NOT EXISTS idx_entity_relationships_type ON entity_relationships (relationship_type);
```

## Integration with Issues

The entity-issue integration includes bidirectional conversion between issue dependencies and entity relationships:

```go
// ConvertIssueDependencyToEntityRelationship converts an issue dependency to an entity relationship
func ConvertIssueDependencyToEntityRelationship(dependency *models.IssueDependency) *models.EntityRelationship

// ConvertEntityRelationshipToIssueDependency converts an entity relationship to an issue dependency
func ConvertEntityRelationshipToIssueDependency(relationship *models.EntityRelationship) *models.IssueDependency
```

## API Endpoints

The following API endpoints are available for working with entity relationships:

### Create Relationship
- **Endpoint**: `POST /api/entity/relationship`
- **Request Body**:
  ```json
  {
      "source_id": "entity-id-1",
      "relationship_type": "depends_on",
      "target_id": "entity-id-2",
      "metadata": {
          "key1": "value1",
          "key2": "value2"
      }
  }
  ```

### Delete Relationship
- **Endpoint**: `DELETE /api/entity/relationship?source_id=entity-id-1&relationship_type=depends_on&target_id=entity-id-2`

### Get Relationship
- **Endpoint**: `GET /api/entity/relationship?source_id=entity-id-1&relationship_type=depends_on&target_id=entity-id-2`

### List Relationships by Source
- **Endpoint**: `GET /api/entity/relationship/source?source_id=entity-id-1`
- **Optional**: `relationship_type=depends_on` (filter by type)

### List Relationships by Target
- **Endpoint**: `GET /api/entity/relationship/target?target_id=entity-id-2`
- **Optional**: `relationship_type=depends_on` (filter by type)

### List Relationships by Type
- **Endpoint**: `GET /api/entity/relationship/type?relationship_type=depends_on`

## Dual-Write Mode

When operating in dual-write mode, the `EntityIssueHandler` will:

1. Write to both the traditional issue dependency system and the entity relationship system
2. For reads, it will attempt to fetch from the entity relationship system first
3. If that fails or returns no results, it will fall back to the traditional system

This ensures backward compatibility while gradually migrating to the new entity-based architecture.

## Testing

The implementation includes comprehensive tests:

1. Unit tests for the `EntityRelationship` model and methods
2. Repository tests for both SQLite and in-memory implementations
3. Conversion tests between issue dependencies and entity relationships
4. Integration tests with the `EntityIssueHandler`

## Example Usage

```go
// Create a relationship
relationship := models.NewEntityRelationship(
    "entity-id-1",              // Source
    models.RelationshipTypeDependsOn, // Type
    "entity-id-2",              // Target
)

// Add metadata
metadata := map[string]interface{}{
    "dependency_type": "blocker",
    "description":     "Entity 1 depends on Entity 2",
}
relationship.AddMetadata(metadata)

// Set creator
relationship.SetCreatedBy("user1")

// Save the relationship
relationshipRepo.Create(relationship)

// Query relationships
relationships, _ := relationshipRepo.GetBySource("entity-id-1")
```

## Future Enhancements

1. **Relationship Validation**: Add validation based on entity types
2. **Circular Dependency Detection**: Detect and prevent circular dependencies
3. **Relationship Events**: Emit events when relationships are created/deleted
4. **Relationship Permissions**: Add permission checks for relationship operations
5. **Bidirectional Relationships**: Automatically create reverse relationships for certain types