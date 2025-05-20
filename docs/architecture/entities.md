# EntityDB Entity-Based Architecture Guide

This guide documents the pure entity-based architecture implemented in EntityDB, representing a complete consolidation of all functionality under a unified entity API.

## Overview

The EntityDB system now operates exclusively with an entity-based architecture, replacing the previous structured issue/task/workspace model. This migration has been completed, with all legacy code removed and all functionality reimplemented using the entity model. The system has been fully consolidated to use a unified entity API for all operations, eliminating specialized endpoints for different object types.

## Core Architecture Principles

The new entity-based architecture follows these core principles:

1. **Everything is an Entity**: All objects in the system (issues, workspaces, agents) are represented as entities with appropriate tags
2. **Relationships, Not References**: Connections between entities are defined as explicit relationships with typed semantics
3. **Tags, Not Tables**: Classification is done via tags, not separate tables or schemas
4. **Content, Not Columns**: Flexible content items store data rather than fixed columns
5. **API-First Design**: All operations must go through the API, with zero direct database access
6. **JWT-Based Authentication**: Secure token-based authentication for all operations
7. **Role-Based Access Control**: Permission management through user roles

## Entity Model

### Entity Structure

An entity consists of the following core attributes:

```json
{
  "id": "entity_1234567890",
  "type": "issue",
  "title": "Example Entity",
  "description": "Detailed description of the entity",
  "status": "in_progress",
  "tags": ["api", "backend", "high-priority"],
  "properties": {
    "priority": "high",
    "estimate": "2h",
    "complexity": "medium"
  },
  "created_at": "2023-08-15T10:30:00Z",
  "updated_at": "2023-08-16T08:45:00Z",
  "created_by": "usr_admin",
  "assigned_to": "claude-2"
}
```

### Entity Relationships

Relationships between entities are explicitly stored as relationship objects:

```json
{
  "id": "rel_1234567890",
  "source_id": "entity_workspace_123",
  "target_id": "entity_issue_456",
  "type": "parent",
  "properties": {
    "order": 1
  },
  "created_at": "2023-08-15T10:35:00Z",
  "created_by": "usr_admin"
}
```

## Configuration Settings

The system has been permanently configured with the following settings:

1. `entity.handler_enabled`: Set to `true`
   - The EntityIssueHandler is used for all issue-related API endpoints
   - Implements issue functionality using the entity model

2. `entity.based_repository_enabled`: Set to `true`
   - Uses the entity-based repository for all data storage
   - Leverages the flexible entity model

3. `entity.relationships_enabled`: Set to `true`
   - Uses entity relationships for all hierarchical and dependency relationships
   - Provides flexible typed connections between entities

4. `entity_api_enabled`: Set to `true`
   - Direct entity API endpoints are always available
   - Provides full access to the entity model capabilities

5. `entity.dual_write_enabled`: Set to `false`
   - Legacy code and dual write capabilities have been removed
   - All operations use the entity model exclusively

6. `entity.pure_api_enabled`: Set to `true`
   - The system now uses pure entity API for all operations
   - Legacy endpoints are redirected to the entity API with deprecation notices

## Removed Components

The following components have been permanently removed from the codebase:

1. `/opt/entitydb/src/deprecated/` - Legacy models and API handlers
2. `/opt/entitydb/src/models/sqlite/deprecated/` - Legacy SQLite implementations
3. All dual write functionality from EntityIssueHandler
4. All specialized issue/workspace/task handler implementations
5. All direct database access code

## API Structure

### Authentication

All API operations require authentication using JWT tokens:

```bash
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

The response includes a token to be used in subsequent requests:

```json
{
  "status": "ok",
  "message": "Login successful",
  "token": "tk_admin_1234567890",
  "user": {
    "id": "usr_admin",
    "username": "admin",
    "roles": ["admin"]
  }
}
```

All authenticated requests must include the token in the Authorization header:

```bash
curl -X GET http://localhost:8085/api/v1/entities/list \
  -H "Authorization: Bearer tk_admin_1234567890"
```

### Entity API Endpoints

#### List Entities

```
GET /api/v1/entities/list
```

Optional query parameters:
- `type`: Filter by entity type
- `status`: Filter by status
- `tags`: Filter by tags (comma-separated)

Example:
```bash
curl -X GET "http://localhost:8085/api/v1/entities/list?type=issue&status=in_progress" \
  -H "Authorization: Bearer tk_admin_1234567890"
```

#### Get Entity by ID

```
GET /api/v1/entities/{entity_id}
```

Example:
```bash
curl -X GET http://localhost:8085/api/v1/entities/entity_1234567890 \
  -H "Authorization: Bearer tk_admin_1234567890"
```

#### Create Entity

```
POST /api/v1/entities
```

Example:
```bash
curl -X POST http://localhost:8085/api/v1/entities \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer tk_admin_1234567890" \
  -d '{
    "type": "issue",
    "title": "Implement Entity API",
    "description": "Create a comprehensive entity API for the system",
    "status": "pending",
    "tags": ["api", "backend", "high-priority"],
    "properties": {
      "priority": "high",
      "estimate": "8h"
    }
  }'
```

#### Update Entity

```
PUT /api/v1/entities/{entity_id}
```

Example:
```bash
curl -X PUT http://localhost:8085/api/v1/entities/entity_1234567890 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer tk_admin_1234567890" \
  -d '{
    "title": "Updated Entity Title",
    "status": "in_progress",
    "tags": ["api", "backend", "updated"],
    "properties": {
      "priority": "medium",
      "progress": 50
    }
  }'
```

#### Delete Entity

```
DELETE /api/v1/entities/{entity_id}
```

Example:
```bash
curl -X DELETE http://localhost:8085/api/v1/entities/entity_1234567890 \
  -H "Authorization: Bearer tk_admin_1234567890"
```

### Entity Relationship API Endpoints

#### List Relationships

```
GET /api/v1/entity-relationships/list
```

Optional query parameters:
- `source`: Filter by source entity ID
- `target`: Filter by target entity ID
- `type`: Filter by relationship type

Example:
```bash
curl -X GET "http://localhost:8085/api/v1/entity-relationships/list?source=entity_workspace_123" \
  -H "Authorization: Bearer tk_admin_1234567890"
```

#### Get Relationship by ID

```
GET /api/v1/entity-relationships/{relationship_id}
```

Example:
```bash
curl -X GET http://localhost:8085/api/v1/entity-relationships/rel_1234567890 \
  -H "Authorization: Bearer tk_admin_1234567890"
```

#### Create Relationship

```
POST /api/v1/entity-relationships
```

Example:
```bash
curl -X POST http://localhost:8085/api/v1/entity-relationships \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer tk_admin_1234567890" \
  -d '{
    "source_id": "entity_workspace_123",
    "target_id": "entity_issue_456",
    "type": "parent",
    "properties": {
      "order": 1
    }
  }'
```

#### Update Relationship

```
PUT /api/v1/entity-relationships/{relationship_id}
```

Example:
```bash
curl -X PUT http://localhost:8085/api/v1/entity-relationships/rel_1234567890 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer tk_admin_1234567890" \
  -d '{
    "properties": {
      "order": 2,
      "note": "Updated relationship"
    }
  }'
```

#### Delete Relationship

```
DELETE /api/v1/entity-relationships/{relationship_id}
```

Example:
```bash
curl -X DELETE http://localhost:8085/api/v1/entity-relationships/rel_1234567890 \
  -H "Authorization: Bearer tk_admin_1234567890"
```

### Legacy API Redirection

All legacy API endpoints are now redirected to the unified entity API with deprecation notices:

```bash
curl -X GET http://localhost:8085/api/v1/direct/workspace/list \
  -H "Authorization: Bearer tk_admin_1234567890"
```

Response:
```json
{
  "status": "ok",
  "data": [...],
  "count": 2,
  "deprecation_notice": "WARNING: This legacy workspace endpoint is deprecated and will be removed soon. Please use /api/v1/entities/list?type=workspace instead."
}
```

## Working with the Entity API

### Entity Operations (Command-line Tools)

1. **Creating Entities**:
   ```bash
   ./bin/add_entity <type> <title> [tags...]
   ```
   Example: `./bin/add_entity issue "Fix login bug" priority:high status:pending workspace:entitydb`

2. **Listing Entities**:
   ```bash
   ./bin/list_entities --tag <tag>
   ```
   Example: `./bin/list_entities --tag type:issue --tag priority:high`

3. **Getting Entity Details**:
   ```bash
   ./bin/list_entities --id <entity_id>
   ```

### Relationship Operations (Command-line Tools)

1. **Creating Relationships**:
   ```bash
   ./bin/add_entity_relationship <source> <relationship_type> <target>
   ```
   Example: `./bin/add_entity_relationship ent_123 depends_on ent_456`

2. **Listing Relationships**:
   ```bash
   ./bin/list_entity_relationships --source=<id> [--type=<type>]
   ./bin/list_entity_relationships --target=<id> [--type=<type>]
   ```

### EntityDB Client Usage

The EntityDB command-line client (`./bin/entitydbc.sh`) has been updated to use the entity-based API while maintaining the familiar interface:

```bash
# List all workspaces (internally uses entity API with type=workspace)
./bin/entitydbc.sh workspace list

# Create a new issue (internally uses entity API with type=issue)
./bin/entitydbc.sh issue create \
  --title="Issue title" \
  --description="Detailed issue description" \
  --priority=medium
```

## Entity Tags Reference

Common tags used in the system:

### Type Tags
- `type:issue`: Represents an issue or task
- `type:workspace`: Represents a workspace (formerly project)
- `type:epic`: Represents an epic (collection of stories)
- `type:story`: Represents a user story
- `type:agent`: Represents an agent profile

### Status Tags
- `status:pending`: Entity is created but not started
- `status:in_progress`: Entity is actively being worked on
- `status:blocked`: Entity is blocked by dependencies
- `status:completed`: Entity has been completed
- `status:archived`: Entity is archived

### Priority Tags
- `priority:low`: Low priority
- `priority:medium`: Medium priority
- `priority:high`: High priority
- `priority:critical`: Critical priority

## Entity Relationship Types

Common relationship types used in the system:

- `parent`: Hierarchical parent-child relationship (workspace to issue)
- `depends_on`: Dependency relationship (source depends on target)
- `assignment`: Assignment relationship (agent to issue)
- `related_to`: General relationship between entities
- `created_by`: Creator relationship
- `belongs_to`: Membership relationship

## Security Considerations

The entity-based architecture enforces strict security measures:

1. **JWT Authentication**: All requests require valid JWT tokens
2. **Role-Based Access Control**: Operations are restricted based on user roles
3. **API-Only Access**: No direct database access is allowed
4. **Input Validation**: All API inputs are validated to prevent security issues
5. **Audit Logging**: All operations are logged for audit purposes

## Testing

To verify the entity-based server implementation, run the test script:

```bash
./bin/test_entity_server.sh
```

This script performs comprehensive tests of:
1. Entity API operations (create, read, update, delete)
2. Entity relationship operations
3. Legacy API redirection
4. Unauthorized access rejection
5. Token-based authentication

## Benefits of Pure Entity Architecture

1. **Flexibility**: Any object type can be represented as an entity with appropriate tags
2. **Extensibility**: New entity types can be added without schema changes
3. **Unified API**: Common API for all entity types
4. **Rich Relationships**: Flexible relationship types between entities
5. **Advanced Filtering**: Tag-based filtering for powerful queries
6. **Simplified Codebase**: Single data model versus multiple specialized models
7. **Future-Proof**: Easy to adapt to new requirements without schema changes
8. **Consolidated Access**: All operations go through a single API
9. **Zero Direct Database Access**: Enhanced security through API-only access
10. **Legacy Compatibility**: Redirection from legacy endpoints with deprecation notices