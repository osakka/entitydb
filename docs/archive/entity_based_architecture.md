# Entity-Based Architecture

## Overview

The entity-based architecture is a flexible, extensible approach to managing data in the EntityDB platform. It introduces a generic Entity model that can represent various domain objects (issues, workspaces, agents, etc.) in a unified way, allowing for more flexible relationships, better extensibility, and improved data querying.

## Core Components

### Entity Model

The Entity model is a generic, flexible data container with the following structure:

- **ID**: Unique identifier for the entity
- **Tags**: A collection of key-value pairs stored as strings for quick classification and lookup
- **Content**: A collection of typed content items with timestamps

This structure allows entities to store any type of data in a flexible way, with the ability to query by tags.

### Entity Relationships

The Entity Relationship model represents typed connections between entities:

- **SourceID**: The ID of the source entity
- **RelationshipType**: The type of relationship (e.g., "depends_on", "blocks", "parent_of")
- **TargetID**: The ID of the target entity
- **Metadata**: Additional JSON data about the relationship

This allows for flexible modeling of complex relationships between entities.

## Integration with Existing Issue Model

The entity-based architecture is designed to work alongside the existing Issue model, with the ability to gradually migrate from one to the other. The key integration points are:

1. **Dual-Write Mode**: In this mode, operations on Issues also create/update corresponding Entities.
2. **Entity-Issue Adapter**: Provides a compatibility layer that presents Entities through the Issue API.
3. **Entity Migration**: Tools to migrate existing Issues to Entities.

## Entity API

The entity API provides direct access to the entity model:

- `POST /api/v1/entities`: Create a new entity
- `GET /api/v1/entities`: List entities, with optional tag filtering
- `GET /api/v1/entities/get`: Get a specific entity by ID

For relationships:

- `POST /api/v1/entity-relationships`: Create a new relationship between entities
- `GET /api/v1/entity-relationships/source/:id`: Get relationships where an entity is the source
- `GET /api/v1/entity-relationships/target/:id`: Get relationships where an entity is the target
- `DELETE /api/v1/entity-relationships/:source/:type/:target`: Delete a specific relationship

## Schema

The entity model is stored in two main tables:

1. **entities**: Stores entity data with JSON for tags and content
2. **entity_relationships**: Stores relationship data

Additional tables include:

- **entity_migrations**: Tracks migration of legacy data to entities

## Configuration

The entity system can be configured through the system config:

- `entity_api_enabled`: Enable/disable the entity API (default: true)
- `entity_migration_enabled`: Enable/disable entity migration (default: false)
- `entity_migration_status`: Status of entity migration (pending, in_progress, completed)

## Permissions

The entity system defines several permissions:

- `entity.view`: Permission to view entities
- `entity.create`: Permission to create entities
- `entity.edit`: Permission to edit entities
- `entity.delete`: Permission to delete entities
- `entity.relationship.view`: Permission to view relationships
- `entity.relationship.create`: Permission to create relationships
- `entity.relationship.delete`: Permission to delete relationships

## Usage Examples

### Creating an Entity

```bash
curl -X POST http://localhost:8085/api/v1/entities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "tags": ["type=issue", "status=pending", "priority=high"],
    "content": [
      {"type": "title", "value": "Fix critical bug"},
      {"type": "description", "value": "Fix the critical issue with data loss"}
    ]
  }'
```

### Creating a Relationship

```bash
curl -X POST http://localhost:8085/api/v1/entity-relationships \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "source_id": "ent_123",
    "relationship_type": "depends_on",
    "target_id": "ent_456",
    "metadata": {
      "dependency_type": "blocker",
      "description": "Must be completed before work can start"
    }
  }'
```

## Migration Workflow

The planned migration workflow from Issues to Entities:

1. Enable dual-write mode (updates to Issues also create/update Entities)
2. Run migration script to create Entities for existing Issues
3. Switch APIs to use Entity-Issue adapter
4. Test thoroughly in this hybrid mode
5. Gradually transition to direct Entity API usage
6. Eventually retire the old Issue model

## Implementation Status

The entity-based architecture has been fully integrated into the codebase with the following components implemented:

- ✅ Entity and EntityRelationship models
- ✅ Entity and EntityRelationship repositories
- ✅ Entity API endpoints
- ✅ Entity-Issue adapter layer
- ✅ Schema migrations
- ✅ Integration with the main server

The system is configured to maintain backward compatibility with the current issue-based API.