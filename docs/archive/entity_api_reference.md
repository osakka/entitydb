# Entity API Reference

This document provides a comprehensive reference for all Entity API endpoints in the EntityDB system.

## Overview

The Entity API provides a flexible, tag-based approach to storing and retrieving data in EntityDB. Entities can represent any object in the system (issues, workspaces, agents, etc.) and are distinguished by their tags rather than separate schemas. This architecture allows for greater flexibility, extensibility, and data integration across the platform.

## Core Concepts

### Entity

An entity is the fundamental data structure in the system with the following properties:

- **ID**: Unique identifier for the entity (format: `entity_[uuid]`)
- **Tags**: List of key-value pairs in the format `key:value` for classification and filtering
- **Content**: Array of content items, each with a key and value for storing arbitrary data
- **Timestamps**: Creation and update timestamps

Example entity structure:
```json
{
  "id": "entity_a1b2c3d4",
  "tags": [
    "type:issue",
    "status:pending",
    "priority:high"
  ],
  "content": [
    {
      "key": "title",
      "value": "Implement Entity API Documentation"
    },
    {
      "key": "description",
      "value": "Create comprehensive API documentation for the entity system"
    }
  ],
  "created_at": "2025-05-10T12:34:56Z",
  "updated_at": "2025-05-10T12:34:56Z"
}
```

### Entity Relationship

Entity relationships connect entities with typed connections:

- **SourceID**: ID of the source entity
- **TargetID**: ID of the target entity
- **RelationshipType**: Type of relationship (e.g., `parent_of`, `depends_on`, `assigned_to`)
- **Metadata**: Optional JSON string for additional relationship data
- **Timestamps**: Creation timestamp and creator information

Example relationship structure:
```json
{
  "source_id": "entity_a1b2c3d4",
  "target_id": "entity_e5f6g7h8",
  "relationship_type": "depends_on",
  "metadata": "{\"priority\": \"high\", \"reason\": \"Blocking issue\"}",
  "created_at": "2025-05-10T12:34:56Z",
  "created_by": "agent_claude"
}
```

## Authentication

All Entity API endpoints require authentication using JWT tokens. Include the token in the `Authorization` header:

```
Authorization: Bearer <your_token>
```

Most endpoints also require specific permissions, as detailed in each endpoint documentation.

## Permissions

The Entity API uses a Role-Based Access Control (RBAC) system. The following permissions are relevant:

- `entity.create` - Create new entities
- `entity.view` - View existing entities
- `entity.update` - Update existing entities
- `entity.delete` - Delete existing entities
- `entity.relationship.create` - Create entity relationships
- `entity.relationship.view` - View entity relationships
- `entity.relationship.delete` - Delete entity relationships

## Endpoints

### Entity Management

#### Create Entity

**POST** `/api/v1/entities`

Create a new entity with specified tags and content.

**Required Permission**: `entity.create`

**Request Body:**
```json
{
  "tags": ["type:issue", "priority:high", "status:pending"],
  "content": [
    {
      "key": "title",
      "value": "Sample Issue Title"
    },
    {
      "key": "description",
      "value": "This is a sample issue description"
    }
  ]
}
```

**Response:**
```json
{
  "id": "entity_12345",
  "tags": ["type:issue", "priority:high", "status:pending"],
  "content": [
    {
      "key": "title",
      "value": "Sample Issue Title"
    },
    {
      "key": "description",
      "value": "This is a sample issue description"
    }
  ],
  "created_at": "2025-05-10T15:32:10Z",
  "updated_at": "2025-05-10T15:32:10Z"
}
```

**Status Codes:**
- `201 Created`: Entity successfully created
- `400 Bad Request`: Invalid request format
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions

#### List Entities

**GET** `/api/v1/entities`

Retrieve a list of entities, optionally filtered by tags.

**Required Permission**: `entity.view`

**Query Parameters:**
- `tag` (optional): Filter by tag (can be repeated for multiple tags, e.g., `?tag=type:issue&tag=priority:high`)
- `page` (optional): Page number for pagination (defaults to 1)
- `limit` (optional): Number of entities per page (defaults to 20, max 100)

**Response:**
```json
{
  "entities": [
    {
      "id": "entity_12345",
      "tags": ["type:issue", "priority:high", "status:pending"],
      "content": [
        {
          "key": "title",
          "value": "Sample Issue Title"
        },
        {
          "key": "description",
          "value": "This is a sample issue description"
        }
      ],
      "created_at": "2025-05-10T15:32:10Z",
      "updated_at": "2025-05-10T15:32:10Z"
    },
    {
      "id": "entity_67890",
      "tags": ["type:issue", "priority:medium", "status:in_progress"],
      "content": [
        {
          "key": "title",
          "value": "Another Issue"
        },
        {
          "key": "description",
          "value": "Another issue description"
        }
      ],
      "created_at": "2025-05-10T14:30:00Z",
      "updated_at": "2025-05-10T16:45:22Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 2
  }
}
```

**Status Codes:**
- `200 OK`: Entities successfully retrieved
- `400 Bad Request`: Invalid query parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions

#### Get Entity

**GET** `/api/v1/entities/get`

Retrieve a specific entity by ID.

**Required Permission**: `entity.view`

**Query Parameters:**
- `id` (required): Entity ID

**Response:**
```json
{
  "id": "entity_12345",
  "tags": ["type:issue", "priority:high", "status:pending"],
  "content": [
    {
      "key": "title",
      "value": "Sample Issue Title"
    },
    {
      "key": "description",
      "value": "This is a sample issue description"
    }
  ],
  "created_at": "2025-05-10T15:32:10Z",
  "updated_at": "2025-05-10T15:32:10Z"
}
```

**Status Codes:**
- `200 OK`: Entity successfully retrieved
- `400 Bad Request`: Missing ID parameter
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Entity not found

#### Update Entity

**PUT** `/api/v1/entities`

Update an existing entity. This can update tags, content, or both.

**Required Permission**: `entity.update`

**Request Body:**
```json
{
  "id": "entity_12345",
  "tags": ["type:issue", "priority:medium", "status:in_progress"],
  "content": [
    {
      "key": "title",
      "value": "Updated Issue Title"
    },
    {
      "key": "description",
      "value": "Updated issue description"
    },
    {
      "key": "assigned_to",
      "value": "agent_claude"
    }
  ]
}
```

**Response:**
```json
{
  "id": "entity_12345",
  "tags": ["type:issue", "priority:medium", "status:in_progress"],
  "content": [
    {
      "key": "title",
      "value": "Updated Issue Title"
    },
    {
      "key": "description",
      "value": "Updated issue description"
    },
    {
      "key": "assigned_to",
      "value": "agent_claude"
    }
  ],
  "created_at": "2025-05-10T15:32:10Z",
  "updated_at": "2025-05-10T16:45:22Z"
}
```

**Status Codes:**
- `200 OK`: Entity successfully updated
- `400 Bad Request`: Invalid request format or missing ID
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Entity not found

#### Delete Entity

**DELETE** `/api/v1/entities`

Delete an entity by ID.

**Required Permission**: `entity.delete`

**Query Parameters:**
- `id` (required): Entity ID

**Response:**
```json
{
  "success": true,
  "message": "Entity deleted successfully",
  "id": "entity_12345"
}
```

**Status Codes:**
- `200 OK`: Entity successfully deleted
- `400 Bad Request`: Missing ID parameter
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Entity not found

### Entity Relationship Management

#### Create Relationship

**POST** `/api/v1/entity-relationships`

Create a relationship between two entities.

**Required Permission**: `entity.relationship.create`

**Request Body:**
```json
{
  "source_id": "entity_12345",
  "target_id": "entity_67890",
  "relationship_type": "depends_on",
  "metadata": "{\"priority\": \"high\", \"reason\": \"Blocking issue\"}"
}
```

**Response:**
```json
{
  "source_id": "entity_12345",
  "target_id": "entity_67890",
  "relationship_type": "depends_on",
  "metadata": "{\"priority\": \"high\", \"reason\": \"Blocking issue\"}",
  "created_at": "2025-05-10T17:00:00Z",
  "created_by": "agent_claude"
}
```

**Status Codes:**
- `201 Created`: Relationship successfully created
- `400 Bad Request`: Invalid request format
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Source or target entity not found
- `409 Conflict`: Relationship already exists

#### List Relationships by Source

**GET** `/api/v1/entity-relationships/source`

Retrieve all relationships where the specified entity is the source.

**Required Permission**: `entity.relationship.view`

**Query Parameters:**
- `id` (required): Source entity ID
- `type` (optional): Filter by relationship type

**Response:**
```json
{
  "relationships": [
    {
      "source_id": "entity_12345",
      "target_id": "entity_67890",
      "relationship_type": "depends_on",
      "metadata": "{\"priority\": \"high\", \"reason\": \"Blocking issue\"}",
      "created_at": "2025-05-10T17:00:00Z",
      "created_by": "agent_claude"
    },
    {
      "source_id": "entity_12345",
      "target_id": "entity_abcdef",
      "relationship_type": "parent_of",
      "metadata": "{}",
      "created_at": "2025-05-10T17:05:00Z",
      "created_by": "agent_claude"
    }
  ]
}
```

**Status Codes:**
- `200 OK`: Relationships successfully retrieved
- `400 Bad Request`: Missing ID parameter
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Source entity not found

#### List Relationships by Target

**GET** `/api/v1/entity-relationships/target`

Retrieve all relationships where the specified entity is the target.

**Required Permission**: `entity.relationship.view`

**Query Parameters:**
- `id` (required): Target entity ID
- `type` (optional): Filter by relationship type

**Response:**
```json
{
  "relationships": [
    {
      "source_id": "entity_abcdef",
      "target_id": "entity_12345",
      "relationship_type": "assigned_to",
      "metadata": "{}",
      "created_at": "2025-05-10T17:10:00Z",
      "created_by": "agent_claude"
    }
  ]
}
```

**Status Codes:**
- `200 OK`: Relationships successfully retrieved
- `400 Bad Request`: Missing ID parameter
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Target entity not found

#### Delete Relationship

**DELETE** `/api/v1/entity-relationships`

Delete a relationship between two entities.

**Required Permission**: `entity.relationship.delete`

**Query Parameters:**
- `source_id` (required): Source entity ID
- `target_id` (required): Target entity ID
- `type` (required): Relationship type

**Response:**
```json
{
  "success": true,
  "message": "Relationship deleted successfully",
  "source_id": "entity_12345",
  "target_id": "entity_67890",
  "relationship_type": "depends_on"
}
```

**Status Codes:**
- `200 OK`: Relationship successfully deleted
- `400 Bad Request`: Missing parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Relationship not found

## Legacy API Redirection

All legacy API endpoints are automatically redirected to the unified entity API with deprecation notices. For example, `/api/v1/issues/list` would be redirected to `/api/v1/entities?tag=type:issue` and a deprecation notice would be included in the response.

```json
{
  "status": "ok",
  "data": [...],
  "count": 2,
  "deprecation_notice": "WARNING: This legacy endpoint is deprecated and will be removed soon. Please use /api/v1/entities?tag=type:issue instead."
}
```

## Error Handling

All API endpoints return consistent error formats:

```json
{
  "error": "Error message describing the issue",
  "code": "ERROR_CODE",
  "details": {
    "field": "Description of the issue with this specific field"
  }
}
```

Common error codes:
- `INVALID_REQUEST`: The request format is invalid
- `ENTITY_NOT_FOUND`: The requested entity does not exist
- `RELATIONSHIP_NOT_FOUND`: The requested relationship does not exist
- `PERMISSION_DENIED`: The user does not have required permissions
- `AUTHENTICATION_REQUIRED`: Authentication is missing or invalid
- `RELATIONSHIP_EXISTS`: The relationship already exists

## Rate Limiting

API requests are rate-limited to 100 requests per minute per user. Rate limit information is included in response headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1589472000
```

When rate limit is exceeded, the API will respond with a 429 Too Many Requests status code.

## Pagination

List endpoints support pagination using the following query parameters:

- `page`: Page number (1-indexed)
- `limit`: Number of items per page (default: 20, max: 100)

Pagination information is included in the response:

```json
{
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 42,
    "pages": 3
  }
}
```

## Command-Line Tools

The Entity API is accessible via the following command-line tools:

- `entitydbc.sh entity create`: Create a new entity
- `entitydbc.sh entity list`: List and filter entities
- `entitydbc.sh entity get`: Get a specific entity
- `entitydbc.sh entity update`: Update an existing entity
- `entitydbc.sh entity relationship create`: Create a relationship between entities
- `entitydbc.sh entity relationship list`: List and filter entity relationships
- `entitydbc.sh entity relationship delete`: Delete a relationship between entities

## Common Patterns and Use Cases

For detailed examples of how to use the Entity API for common tasks, see the [Entity API Usage Guide](entity_api_usage_guide.md) document.