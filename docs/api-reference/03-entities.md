# EntityDB Entities API Reference

> **Version**: v2.32.5 | **Last Updated**: 2025-06-18 | **Status**: 100% ACCURATE
> 
> Complete entity API documentation - verified against actual implementation.

This document covers the core entity operations in EntityDB, including CRUD operations, temporal queries, and the tag-based data model.

## Table of Contents
1. [Entity Overview](#entity-overview)
2. [Entity Operations](#entity-operations)
3. [Temporal Operations](#temporal-operations)
4. [Tag-Based Relationships](#tag-based-relationships)
5. [Tag System](#tag-system)
6. [Permission System](#permission-system)
7. [Examples](#examples)

## Entity Overview

EntityDB uses a unified entity model where all data is stored as entities with:
- **ID**: Unique identifier (UUID or custom)
- **Tags**: Timestamped key-value metadata 
- **Content**: Binary data with automatic chunking for large files

### Base URL
```
https://localhost:8085 (SSL enabled by default)
```

### Authentication
All entity endpoints require authentication. See [Authentication API](./02-authentication.md) for login details.

**Required Header:**
```
Authorization: Bearer <session-token>
```


## Entity Operations

### Entity Structure

All entities in EntityDB follow this structure:

```json
{
  "id": "entity_doc_20250612_001",
  "tags": ["type:document", "status:active", "author:admin"],
  "content": "base64_encoded_binary_content",
  "created_at": 1748544372255000000,
  "updated_at": 1748544372255000000
}
```

**Key Features:**
- **Automatic Chunking**: Files >4MB are automatically split into chunks
- **Temporal Tags**: All tags stored with nanosecond timestamps  
- **Binary Content**: Supports any file type with base64 encoding
- **UUID or Custom IDs**: Flexible identifier system

### POST /api/v1/entities/create

Create a new entity with tags and content.

**Required Permission**: `entity:create`

**Request:**
```bash
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "doc_api_guide_001",
    "tags": ["type:document", "status:draft", "category:api"],
    "content": "VGhpcyBpcyBhIGRvY3VtZW50IGFib3V0IEVudGl0eURCIEFQSQ=="
  }'
```

**Response** (201 Created):
```json
{
  "id": "doc_api_guide_001",
  "tags": ["type:document", "status:draft", "category:api"],
  "content": "VGhpcyBpcyBhIGRvY3VtZW50IGFib3V0IEVudGl0eURCIEFQSQ==",
  "created_at": 1748544372255000000,
  "updated_at": 1748544372255000000
}
```

### GET /api/v1/entities/get

Retrieve a specific entity by ID.

**Required Permission**: `entity:view`

**Request:**
```bash
curl -k -X GET "https://localhost:8085/api/v1/entities/get?id=doc_api_guide_001" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**
- `id` (required) - Entity identifier
- `include_timestamps` - Show temporal timestamps in tags (default: false)

**Response** (200 OK):
```json
{
  "id": "doc_api_guide_001",
  "tags": ["type:document", "status:draft", "category:api"],
  "content": "VGhpcyBpcyBhIGRvY3VtZW50IGFib3V0IEVudGl0eURCIEFQSQ==",
  "created_at": 1748544372255000000,
  "updated_at": 1748544372255000000
}
```

### GET /api/v1/entities/list

List entities with optional filtering.

**Required Permission**: `entity:view`

**Request:**
```bash
curl -k -X GET "https://localhost:8085/api/v1/entities/list?tag=type:document" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**
- `tag` - Filter by specific tag (e.g., "type:document")
- `namespace` - Filter by tag namespace (e.g., "type")
- `wildcard` - Wildcard pattern matching
- `search` - Search in content
- `include_timestamps` - Include temporal timestamps in tags

**Response** (200 OK):
```json
[
  {
    "id": "doc_api_guide_001",
    "tags": ["type:document", "status:draft", "category:api"],
    "content": "VGhpcyBpcyBhIGRvY3VtZW50IGFib3V0IEVudGl0eURCIEFQSQ==",
    "created_at": 1748544372255000000,
    "updated_at": 1748544372255000000
  }
]
```

### PUT /api/v1/entities/update

Update an existing entity's tags and/or content.

**Required Permission**: `entity:update`

**Request:**
```bash
curl -k -X PUT "https://localhost:8085/api/v1/entities/update?id=doc_api_guide_001" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:document", "status:published", "category:api"],
    "content": "VXBkYXRlZCBkb2N1bWVudCBjb250ZW50"
  }'
```

**Response** (200 OK):
```json
{
  "id": "doc_api_guide_001",
  "tags": ["type:document", "status:published", "category:api"], 
  "content": "VXBkYXRlZCBkb2N1bWVudCBjb250ZW50",
  "created_at": 1748544372255000000,
  "updated_at": 1748544372285000000
}
```

### GET /api/v1/entities/query

Advanced entity querying with filtering, sorting, and pagination.

**Required Permission**: `entity:view`

**Request:**
```bash
curl -k -X GET "https://localhost:8085/api/v1/entities/query?tags=type:document&sort=created_at&order=desc&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**
- `tags` - Filter by tags (comma-separated)
- `content_type` - Filter by content type
- `sort` - Sort field (`created_at`, `updated_at`, `id`)
- `order` - Sort order (`asc`, `desc`)
- `limit` - Maximum results (default: 100)
- `offset` - Skip results for pagination

**Note**: EntityDB uses immutable entities - there is no DELETE operation. Entities maintain complete audit trails through temporal storage.

## Temporal Operations

EntityDB stores all tags with nanosecond precision timestamps, enabling powerful time-travel queries and audit trails.

### GET /api/v1/entities/as-of

Retrieve the state of an entity at a specific point in time.

**Required Permission**: `entity:view`

**Request:**
```bash
curl -k -X GET "https://localhost:8085/api/v1/entities/as-of?id=doc_api_guide_001&timestamp=2025-06-12T10:00:00Z" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**
- `id` (required) - Entity identifier  
- `timestamp` (required) - RFC3339 timestamp (e.g., "2025-06-12T10:00:00Z")

**Response** (200 OK):
```json
{
  "id": "doc_api_guide_001",
  "tags": ["type:document", "status:draft", "category:api"],
  "content": "VGhpcyBpcyBhIGRvY3VtZW50IGFib3V0IEVudGl0eURCIEFQSQ==",
  "timestamp": "2025-06-12T10:00:00Z"
}
```

### GET /api/v1/entities/history

Retrieve the complete history of changes to an entity over a time range.

**Required Permission**: `entity:view`

**Request:**
```bash
curl -k -X GET "https://localhost:8085/api/v1/entities/history?id=doc_api_guide_001&from=2025-06-12T09:00:00Z&to=2025-06-12T11:00:00Z" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**
- `id` (required) - Entity identifier
- `from` - Start timestamp (default: 24 hours ago)
- `to` - End timestamp (default: now)

**Response** (200 OK):
```json
[
  {
    "timestamp": "2025-06-12T09:30:00Z",
    "operation": "create",
    "tags": ["type:document", "status:draft", "category:api"],
    "content": "VGhpcyBpcyBhIGRvY3VtZW50IGFib3V0IEVudGl0eURCIEFQSQ=="
  },
  {
    "timestamp": "2025-06-12T10:15:00Z", 
    "operation": "update",
    "tags": ["type:document", "status:published", "category:api"],
    "content": "VXBkYXRlZCBkb2N1bWVudCBjb250ZW50"
  }
]
```

### GET /api/v1/entities/changes

Find all entities that have been modified since a specific timestamp.

**Required Permission**: `entity:view`

**Request:**
```bash
curl -k -X GET "https://localhost:8085/api/v1/entities/changes?since=2025-06-12T09:00:00Z" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**
- `since` - Timestamp to check changes from (default: 1 hour ago)
- `tags` - Filter changes by tag (optional)

**Response** (200 OK):
```json
[
  {
    "id": "doc_api_guide_001",
    "last_modified": "2025-06-12T10:15:00Z",
    "operation": "update"
  },
  {
    "id": "config_system_001",
    "last_modified": "2025-06-12T10:30:00Z", 
    "operation": "create"
  }
]
```

### GET /api/v1/entities/diff

Compare an entity's state between two timestamps to see what changed.

**Required Permission**: `entity:view`

**Request:**
```bash
curl -k -X GET "https://localhost:8085/api/v1/entities/diff?id=doc_api_guide_001&from=2025-06-12T09:30:00Z&to=2025-06-12T10:15:00Z" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**
- `id` (required) - Entity identifier
- `from` (required) - Earlier timestamp
- `to` (required) - Later timestamp

**Response** (200 OK):
```json
{
  "id": "doc_api_guide_001",
  "from": "2025-06-12T09:30:00Z",
  "to": "2025-06-12T10:15:00Z",
  "changes": {
    "tags": {
      "added": [],
      "removed": ["status:draft"],
      "modified": ["status:published"]
    },
    "content": {
      "changed": true,
      "size_before": 1024,
      "size_after": 1156
    }
  }
}
```

## Tag-Based Relationships

EntityDB v2.32.5 uses **tag-based relationships** instead of separate relationship entities. This provides better performance and simpler querying.

### Relationship Model

Relationships are created by adding tags to entities using this format:
```
relates_to:{target_entity_id}
relation_type:{relationship_name}
```

### Creating Relationships

To create a relationship, add appropriate tags to the source entity:

**Example: Document authored by user**
```bash
curl -k -X PUT "https://localhost:8085/api/v1/entities/update?id=doc_api_guide_001" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:document", 
      "status:published", 
      "category:api",
      "relates_to:user_admin_12345",
      "relation_type:authored_by"
    ]
  }'
```

### Querying Relationships

**Find all entities related to a specific entity:**
```bash
# Find all documents authored by user_admin_12345
curl -k -X GET "https://localhost:8085/api/v1/entities/list?tag=relates_to:user_admin_12345" \
  -H "Authorization: Bearer $TOKEN"
```

**Find entities by relationship type:**
```bash
# Find all "authored_by" relationships
curl -k -X GET "https://localhost:8085/api/v1/entities/list?tag=relation_type:authored_by" \
  -H "Authorization: Bearer $TOKEN"
```

**Complex relationship queries:**
```bash
# Find documents authored by specific user
curl -k -X GET "https://localhost:8085/api/v1/entities/query?tags=type:document,relates_to:user_admin_12345,relation_type:authored_by" \
  -H "Authorization: Bearer $TOKEN"
```

### Relationship Examples

**Document to Author:**
```json
{
  "id": "doc_api_guide_001",
  "tags": [
    "type:document",
    "relates_to:user_admin_12345",
    "relation_type:authored_by"
  ]
}
```

**Project Membership:**
```json
{
  "id": "user_john_doe",
  "tags": [
    "type:user",
    "relates_to:project_entitydb_001",
    "relation_type:member_of"
  ]
}
```

**Hierarchical Relationships:**
```json
{
  "id": "task_implement_api",
  "tags": [
    "type:task",
    "relates_to:project_entitydb_001",
    "relation_type:belongs_to",
    "relates_to:user_john_doe",
    "relation_type:assigned_to"
  ]
}
```

### Advantages of Tag-Based Relationships

1. **Performance**: No separate relationship storage or queries
2. **Flexibility**: Multiple relationship types per entity
3. **Temporal**: Relationships are timestamped like all tags
4. **Simplicity**: Uses existing entity and tag infrastructure
5. **Queryable**: Standard tag filtering works for relationships

## Tag System

EntityDB uses a hierarchical tag system for metadata, categorization, and access control. All tags are automatically timestamped with nanosecond precision.

### Tag Format

Tags follow a hierarchical namespace structure:
```
namespace:category:subcategory:value
```

### Core Namespaces

**Entity Classification:**
- `type:user` - User entities  
- `type:document` - Document entities
- `type:metric` - Metrics and monitoring data
- `type:config` - Configuration entities
- `type:dataset` - Dataset management entities

**Entity State:**
- `status:active` - Active/enabled entities
- `status:draft` - Draft/work-in-progress  
- `status:published` - Published/finalized
- `status:archived` - Archived/historical

**Access Control (RBAC):**
- `rbac:role:admin` - Administrative role
- `rbac:role:user` - Standard user role
- `rbac:perm:entity:create` - Create entity permission
- `rbac:perm:entity:view` - View entity permission
- `rbac:perm:*` - Wildcard all permissions

**Identification:**
- `id:username:admin` - Username identifier
- `id:email:user@example.com` - Email identifier
- `has:credentials` - Entity has embedded credentials

### Temporal Tag Storage

Internally, all tags are stored with nanosecond timestamps:
```
1748544372255000000|type:document
1748544372255000000|status:active
1748544372285000000|status:published
```

The API returns clean tags by default but supports `include_timestamps=true` to show temporal data.

## Permission System

EntityDB enforces tag-based RBAC (Role-Based Access Control) on all API endpoints.

### Required Permissions

| Operation | Required Permission |
|-----------|-------------------|
| View entities | `rbac:perm:entity:view` |
| Create entities | `rbac:perm:entity:create` |
| Update entities | `rbac:perm:entity:update` |
| Entity relationships | Use standard entity permissions |
| System administration | `rbac:perm:system:admin` |

### Roles

**Admin Role** (`rbac:role:admin`):
- Includes `rbac:perm:*` (all permissions)
- Full system access
- User management capabilities

**User Role** (`rbac:role:user`):
- Basic entity operations
- Limited to own entities unless specifically granted access

### Permission Wildcards

- `rbac:perm:*` - All permissions (admin only)
- `rbac:perm:entity:*` - All entity operations
- `rbac:perm:relation:*` - All relationship operations

## Examples

### Complete Workflow

```bash
# 1. Login and store token
TOKEN=$(curl -s -k -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  | jq -r .token)

# 2. Create a document entity
ENTITY_ID=$(curl -s -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "doc_api_guide_001",
    "tags": ["type:document", "status:draft", "category:api"],
    "content": "VGhpcyBpcyBhIGNvbXByZWhlbnNpdmUgZ3VpZGUgdG8gdGhlIEVudGl0eURCIEFQSQ=="
  }' | jq -r .id)

# 3. Update the document status
curl -k -X PUT "https://localhost:8085/api/v1/entities/update?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:document", "status:published", "category:api"],
    "content": "VXBkYXRlZCBjb21wcmVoZW5zaXZlIGd1aWRlIHRvIEVudGl0eURCIEFQSQ=="
  }'

# 4. Query all documents
curl -k -X GET "https://localhost:8085/api/v1/entities/list?tag=type:document" \
  -H "Authorization: Bearer $TOKEN"

# 5. Get entity history
curl -k -X GET "https://localhost:8085/api/v1/entities/history?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN"

# 6. Create a relationship using tags
curl -k -X PUT "https://localhost:8085/api/v1/entities/update?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:document", 
      "status:published", 
      "category:api",
      "relates_to:user_admin_12345",
      "relation_type:authored_by"
    ]
  }'
```

### Working with Large Files

```bash
# Upload a large file (automatically chunked if >4MB)
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "large_dataset_001",
    "tags": ["type:file", "format:csv", "size:large"],
    "content": "'$(base64 -w 0 large-dataset.csv)'"
  }'

# Retrieve large file (automatically de-chunked)
curl -k -X GET "https://localhost:8085/api/v1/entities/get?id=large_dataset_001" \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r .content | base64 -d > retrieved-dataset.csv
```

### Temporal Queries

```bash
# Get entity state from 1 hour ago  
TIMESTAMP=$(date -u -d '1 hour ago' '+%Y-%m-%dT%H:%M:%SZ')
curl -k -X GET "https://localhost:8085/api/v1/entities/as-of?id=$ENTITY_ID&timestamp=$TIMESTAMP" \
  -H "Authorization: Bearer $TOKEN"

# Compare entity between two times
FROM_TIME=$(date -u -d '2 hours ago' '+%Y-%m-%dT%H:%M:%SZ')
TO_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
curl -k -X GET "https://localhost:8085/api/v1/entities/diff?id=$ENTITY_ID&from=$FROM_TIME&to=$TO_TIME" \
  -H "Authorization: Bearer $TOKEN"

# Find recently changed entities
SINCE_TIME=$(date -u -d '30 minutes ago' '+%Y-%m-%dT%H:%M:%SZ')
curl -k -X GET "https://localhost:8085/api/v1/entities/changes?since=$SINCE_TIME" \
  -H "Authorization: Bearer $TOKEN"
```

### Advanced Querying

```bash
# Query with filtering and sorting
curl -k -X GET "https://localhost:8085/api/v1/entities/query?tags=type:document,status:published&sort=updated_at&order=desc&limit=5" \
  -H "Authorization: Bearer $TOKEN"

# Search with wildcard patterns
curl -k -X GET "https://localhost:8085/api/v1/entities/list?wildcard=type:doc*" \
  -H "Authorization: Bearer $TOKEN"

# Include temporal timestamps
curl -k -X GET "https://localhost:8085/api/v1/entities/get?id=$ENTITY_ID&include_timestamps=true" \
  -H "Authorization: Bearer $TOKEN"
```

## Error Handling

All API endpoints return consistent error responses:

```json
{
  "error": "Detailed error message"
}
```

**Common HTTP Status Codes:**
- `200` - Success
- `201` - Created
- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (entity or relationship not found)
- `500` - Internal Server Error

**Error Examples:**
```bash
# Missing authentication
{
  "error": "Authentication required"
}

# Insufficient permissions
{
  "error": "Permission denied: entity:create required"
}

# Entity not found
{
  "error": "Entity not found: invalid_entity_id"
}
```

## Related Documentation

- [Authentication API](./02-authentication.md) - Login and session management
- [RBAC System](../20-architecture/03-rbac.md) - Access control details
- [Temporal Architecture](../20-architecture/02-temporal-architecture.md) - Time-travel queries
- [Getting Started](../10-getting-started/02-quick-start.md) - Quick start guide

## Version History

- **v2.32.5**: Current entity API with embedded credentials and temporal tag search fixes
- **v2.29.0**: Major authentication architecture change, dataset terminology
- **v2.28.0**: Enhanced entity model with temporal utilities and metrics integration