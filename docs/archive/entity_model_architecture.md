# Entity Model Architecture

## Overview

The EntityDB system has been redesigned with a minimalist entity model that consists of just three core components:

1. **ID**: A unique identifier for the entity
2. **Tags**: A collection of timestamped key-value pairs
3. **Content**: A collection of timestamped typed content items

This document provides an overview of the new architecture, its benefits, and how to use it.

## Entity Structure

### ID

Each entity has a unique identifier, which can be automatically generated or explicitly provided. IDs are prefixed with "ent_" by default when auto-generated.

### Tags

Tags represent metadata about the entity. Each tag consists of:

- **Timestamp**: When the tag was added, in ISO format with nanosecond precision
- **Key**: The tag name (can be nested using dot notation)
- **Value**: The tag value

Tags are stored as strings in the format: `timestamp.key=value`

Example tags:
- `2025-05-09T00:00:00.000000000.type=workspace`
- `2025-05-09T00:00:00.000000000.status=active`
- `2025-05-09T00:00:00.000000000.customer.name=acme`

### Content

Content represents the actual data of the entity. Each content item consists of:

- **Timestamp**: When the content was added
- **Type**: The type of content (e.g., text, json, binary)
- **Value**: The content value

## Database Schema

The database schema has been simplified to a single `entities` table:

```sql
CREATE TABLE entities (
    id TEXT PRIMARY KEY,
    tags TEXT NOT NULL,  -- JSON array of tag strings
    content TEXT NOT NULL -- JSON array of content items
);
```

## API

The API has been updated to work with the new entity model:

### Entity Creation

```
POST /api/v1/test/entity/create
```

Request body:
```json
{
  "id": "optional-entity-id",
  "tags": ["type=workspace", "status=active", "name=entitydb"],
  "content": [
    {
      "type": "text",
      "value": "Entity description"
    }
  ]
}
```

### Entity Retrieval

```
GET /api/v1/test/entity/get?id={entity-id}
```

### Entity Listing

```
GET /api/v1/test/entity/list
GET /api/v1/test/entity/list?tag={tag-name}
```

## Command Line Client

A command-line client (`entity-client.sh`) has been created to interact with the entity-based system:

```bash
# Create a new entity
./bin/entity-client.sh create --tags="type=issue,status=pending" --content-value="This is an issue"

# Get an entity by ID
./bin/entity-client.sh get --id=ent_12345

# List entities
./bin/entity-client.sh list
./bin/entity-client.sh list --tag=type
```

## Benefits of the Entity Model

1. **Simplicity**: The model is extremely simple, with just three core components
2. **Flexibility**: Any type of data can be modeled using the entity structure
3. **Historical Tracking**: All changes are automatically timestamped
4. **Schema-less**: No need to modify the database schema for new entity types
5. **Unified Model**: All domain objects use the same structure
6. **Extensibility**: New attributes can be added without code changes

## Data Migration

The system has been completely rebuilt with the entity model, replacing the previous issue-based architecture. The migration process includes:

1. Creating a new `entities` table
2. Dropping all previous tables (issues, issue_tags, etc.)
3. Creating default entities for workspaces and other core objects

## Usage Examples

### Representing a Workspace

```json
{
  "id": "entity_workspace_main",
  "tags": [
    "2025-05-09T00:00:00.000000000.type=workspace",
    "2025-05-09T00:00:00.000000000.status=active",
    "2025-05-09T00:00:00.000000000.name=main"
  ],
  "content": [
    {
      "timestamp": "2025-05-09T00:00:00.000000000",
      "type": "text",
      "value": "Main workspace for EntityDB system"
    }
  ]
}
```

### Representing an Issue

```json
{
  "id": "entity_issue_123",
  "tags": [
    "2025-05-09T00:00:00.000000000.type=issue",
    "2025-05-09T00:00:00.000000000.status=pending",
    "2025-05-09T00:00:00.000000000.priority=high",
    "2025-05-09T00:00:00.000000000.parent=entity_workspace_main"
  ],
  "content": [
    {
      "timestamp": "2025-05-09T00:00:00.000000000",
      "type": "text",
      "value": "Fix bug in authentication system"
    },
    {
      "timestamp": "2025-05-09T00:05:00.000000000",
      "type": "comment",
      "value": "This is a critical issue that needs to be fixed ASAP"
    }
  ]
}
```

## Conclusion

The new entity-based architecture provides a flexible, extensible foundation for the EntityDB system. By simplifying the data model to just ID, tags, and content, we've created a system that can easily adapt to changing requirements without requiring schema changes or complex code updates.