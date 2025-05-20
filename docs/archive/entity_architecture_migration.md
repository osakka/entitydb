# Migration to Pure Entity-Based Architecture

## Introduction

The EntityDB platform is transitioning to a fully entity-based architecture, moving away from specific hardcoded tables for different types of objects (issues, workspaces, agents) towards a flexible entity system with tags for classification.

This document describes the migration plan, implementation details, and guidelines for developers to adapt to the new architecture.

## Why Entity-Based Architecture?

The entity-based architecture provides several benefits:

1. **Flexibility**: New entity types can be added without schema changes
2. **Unified API**: Common operations work across all entity types
3. **Rich Metadata**: Tags allow for powerful filtering and categorization
4. **Simpler Code**: Reduced duplication across different entity handlers
5. **Extensibility**: New attributes can be added without schema changes

## Implementation Status

The migration to entity-based architecture is currently in progress:

- ✅ Entity core models implemented
- ✅ Entity repository and relationship repository interfaces defined
- ✅ SQLite implementation of entity repositories
- ✅ Entity API endpoints for CRUD operations
- ✅ Entity relationship API endpoints
- ✅ Tag-based filtering system
- ✅ Direct entity-based API endpoints for workspaces and issues
- ✅ Entity-to-issue adapter for backward compatibility
- ⚠️ Deprecation of legacy issue handlers (in progress)
- ⚠️ Database migration scripts (in progress)

## New Entity-Based API Endpoints

The following new API endpoints have been implemented for direct entity operations:

### Workspace API

- `POST /api/v1/direct/workspace/create` - Create a new workspace
- `GET /api/v1/direct/workspace/list` - List all workspaces
- `GET /api/v1/direct/workspace/get` - Get workspace by ID

### Issue API

- `POST /api/v1/direct/issue/create` - Create a new issue
- `POST /api/v1/direct/issue/assign` - Assign issue to agent
- `POST /api/v1/direct/issue/status` - Update issue status
- `GET /api/v1/direct/issue/list` - List issues with optional filters
- `GET /api/v1/direct/issue/get` - Get issue by ID

### Entity API

- `POST /api/v1/entities` - Create a new entity
- `GET /api/v1/entities` - List entities with optional tag filter
- `GET /api/v1/entities/get` - Get entity by ID

### Entity Relationship API

- `POST /api/v1/entity-relationships` - Create a new relationship
- `GET /api/v1/entity-relationships/source` - List relationships by source
- `GET /api/v1/entity-relationships/target` - List relationships by target
- `DELETE /api/v1/entity-relationships` - Delete a relationship

## Database Schema

The entity-based architecture uses two main tables:

### entities

```sql
CREATE TABLE entities (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    updated_by TEXT
);

CREATE TABLE entity_tags (
    entity_id TEXT,
    tag TEXT,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    added_by TEXT,
    PRIMARY KEY (entity_id, tag),
    FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);

CREATE TABLE entity_content (
    entity_id TEXT,
    timestamp TEXT,
    type TEXT,
    value TEXT,
    PRIMARY KEY (entity_id, timestamp, type),
    FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);
```

### entity_relationships

```sql
CREATE TABLE entity_relationships (
    id TEXT PRIMARY KEY,
    source_id TEXT,
    relationship_type TEXT,
    target_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    metadata TEXT,
    FOREIGN KEY (source_id) REFERENCES entities(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES entities(id) ON DELETE CASCADE,
    UNIQUE (source_id, relationship_type, target_id)
);
```

## Migration Plan

### Phase 1: Dual Implementation (Completed)

- Implement entity models, repositories, and APIs
- Create adapter layer between issues and entities
- Support both models simultaneously

### Phase 2: Direct Entity API (Current)

- Implement direct entity-based API for workspaces and issues
- Create setup scripts using new entity-based API
- Update documentation

### Phase 3: Full Migration (Upcoming)

- Migrate all data from legacy tables to entity tables
- Deprecate legacy API endpoints
- Remove legacy code and tables

## Developer Guidelines

### Creating New Entities

```go
// Create entity
entity := models.NewEntity(id)

// Add tags
entity.AddTag("type", "issue")
entity.AddTag("status", "pending")
entity.AddTag("priority", "high")

// Add content
entity.AddContent("title", "Fix login bug")
entity.AddContent("description", "Users are unable to login with valid credentials")

// Save entity
entityRepo.Create(entity)
```

### Creating Entity Relationships

```go
// Create relationship
relationship := &models.EntityRelationship{
    SourceID:         "issue_123",
    RelationshipType: "assigned_to",
    TargetID:         "agent_456",
    CreatedAt:        time.Now(),
    Metadata:         map[string]interface{}{
        "assigned_by": "user_789",
        "assigned_at": time.Now().Format(time.RFC3339),
    },
}

// Save relationship
entityRelationshipRepo.Create(relationship)
```

### Querying Entities

```go
// Get entity by ID
entity, err := entityRepo.GetByID("issue_123")

// Find entities by tag
issues, err := entityRepo.FindByTag("type:issue")

// Find entities by multiple tags
highPriorityIssues, err := entityRepo.FindByTags([]string{"type:issue", "priority:high"})
```

## Conclusion

The migration to a pure entity-based architecture represents a significant improvement in the flexibility and maintainability of the EntityDB platform. While the transition is still in progress, the new architecture is already available for use and provides a solid foundation for future development.

All new features and enhancements should use the entity-based architecture moving forward.