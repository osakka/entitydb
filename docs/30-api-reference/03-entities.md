# EntityDB Core API Reference

**Version**: 2.19.0  
**Last Updated**: 2025-05-30

> [!NOTE]
> This document covers the core entity operations. For the complete API reference including all 85+ endpoints, see [API_REFERENCE_COMPLETE.md](./API_REFERENCE_COMPLETE.md).

## Table of Contents
1. [Authentication](#authentication)
2. [Entity Operations](#entity-operations)
3. [Temporal Operations](#temporal-operations)
4. [Entity Relationships](#entity-relationships)
5. [Tag System](#tag-system)
6. [Permission System](#permission-system)
7. [Examples](#examples)

## Authentication

All API endpoints (except `/api/v1/auth/login`) require authentication via Bearer token.

### Headers
```
Authorization: Bearer <token>
Content-Type: application/json
```

### Base URL
```
http://localhost:8085
https://localhost:8443 (when SSL enabled)
```

### Login
```http
POST /api/v1/auth/login
```

**Request:**
```json
{
  "username": "admin",
  "password": "admin"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "roles": ["admin", "user"]
  },
  "expires_at": "2025-01-02T10:00:00Z"
}
```

### Logout
```http
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

### Check Status
```http
GET /api/v1/auth/status
Authorization: Bearer <token>
```

### Who Am I
```http
GET /api/v1/auth/whoami
Authorization: Bearer <token>
```

### Refresh Token
```http
POST /api/v1/auth/refresh
Authorization: Bearer <token>
```

## Entity Operations

EntityDB uses a unified entity model where everything is stored as entities with tags and content.

### Entity Structure
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tags": ["type:document", "status:active"],
  "content": "base64_encoded_or_json_content",
  "created_at": 1737564000000000000,
  "updated_at": 1737564000000000000
}
```

### List Entities
```http
GET /api/v1/entities/list
Authorization: Bearer <token>
```

**Query Parameters:**
- `tag` - Filter by specific tag (e.g., "type:user")
- `wildcard` - Filter by wildcard pattern
- `search` - Search in content
- `contentType` - Content type for search
- `namespace` - Filter by tag namespace
- `include_timestamps` - Include temporal timestamps in tags (default: false)

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "tags": ["type:document", "status:active", "author:admin"],
    "content": "...",
    "created_at": 1737564000000000000,
    "updated_at": 1737564000000000000
  }
]
```

### Get Entity
```http
GET /api/v1/entities/get?id=<entity_id>
Authorization: Bearer <token>
```

**Query Parameters:**
- `id` (required) - Entity ID
- `include_timestamps` - Include temporal timestamps in tags

### Create Entity
```http
POST /api/v1/entities/create
Authorization: Bearer <token>
```

**Request:**
```json
{
  "id": "optional_custom_id",
  "tags": ["type:document", "status:draft", "priority:high"],
  "content": "string, base64, or JSON object"
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tags": ["type:document", "status:draft", "priority:high"],
  "content": "...",
  "created_at": 1737564000000000000,
  "updated_at": 1737564000000000000
}
```

### Update Entity
```http
PUT /api/v1/entities/update?id=<entity_id>
Authorization: Bearer <token>
```

**Request:**
```json
{
  "tags": ["type:document", "status:published", "priority:low"],
  "content": "updated content"
}
```

### Query Entities (Advanced)
```http
GET /api/v1/entities/query
Authorization: Bearer <token>
```

**Query Parameters:**
- `filter` - Field to filter on (created_at, updated_at, tag:*)
- `operator` - Comparison operator (eq, ne, gt, lt, gte, lte, like, in)
- `value` - Value to compare
- `sort` - Sort field (created_at, updated_at, id, tag_count)
- `order` - Sort order (asc, desc)
- `limit` - Maximum results
- `offset` - Skip results

### Stream Entity
For large files, use streaming:
```http
GET /api/v1/entities/stream?id=<entity_id>
Authorization: Bearer <token>
```

### Download Entity
Download entity as file:
```http
GET /api/v1/entities/download?id=<entity_id>
Authorization: Bearer <token>
```

## Temporal Operations

EntityDB stores all tags with nanosecond precision timestamps, enabling powerful temporal queries.

### Get Entity As-Of
Retrieve entity state at a specific time:
```http
GET /api/v1/entities/as-of?id=<entity_id>&as_of=<timestamp>
Authorization: Bearer <token>
```

**Parameters:**
- `id` - Entity ID
- `as_of` - RFC3339 timestamp

### Get Entity History
Retrieve entity changes over time:
```http
GET /api/v1/entities/history?id=<entity_id>&from=<timestamp>&to=<timestamp>
Authorization: Bearer <token>
```

**Parameters:**
- `id` - Entity ID
- `from` - Start timestamp (default: 24 hours ago)
- `to` - End timestamp (default: now)

### Get Recent Changes
Find entities modified recently:
```http
GET /api/v1/entities/changes?since=<timestamp>
Authorization: Bearer <token>
```

**Parameters:**
- `since` - Timestamp (default: 1 hour ago)

### Get Entity Diff
Compare entity at two times:
```http
GET /api/v1/entities/diff?id=<entity_id>&t1=<timestamp>&t2=<timestamp>
Authorization: Bearer <token>
```

## Entity Relationships

Create and manage relationships between entities.

### Create Relationship
```http
POST /api/v1/entity-relationships
Authorization: Bearer <token>
```

**Request:**
```json
{
  "source_id": "550e8400-e29b-41d4-a716-446655440000",
  "target_id": "660e8400-e29b-41d4-a716-446655440001",
  "relationship_type": "contains"
}
```

### Get Relationship
```http
GET /api/v1/entity-relationships?source_id=<id>&relationship_type=<type>&target_id=<id>
Authorization: Bearer <token>
```

### List by Source
```http
GET /api/v1/entity-relationships/by-source?source_id=<entity_id>
Authorization: Bearer <token>
```

### List by Target
```http
GET /api/v1/entity-relationships/by-target?target_id=<entity_id>
Authorization: Bearer <token>
```

### Delete Relationship
```http
DELETE /api/v1/entity-relationships?source_id=<id>&relationship_type=<type>&target_id=<id>
Authorization: Bearer <token>
```

## Tag System

EntityDB uses a hierarchical tag system for metadata and permissions.

### Tag Format
```
namespace:category:subcategory:value
```

### Core Namespaces

1. **type:** - Entity classification
   - `type:user`
   - `type:document`
   - `type:metric`
   - `type:config`

2. **rbac:** - Access control
   - `rbac:role:admin`
   - `rbac:role:user`
   - `rbac:perm:entity:create`
   - `rbac:perm:*`

3. **status:** - Entity state
   - `status:active`
   - `status:draft`
   - `status:published`
   - `status:archived`

4. **id:** - Unique identifiers
   - `id:username:admin`
   - `id:email:user@example.com`

### Temporal Tags
All tags are stored with timestamps internally:
```
1737564000000000000|type:document
```

The API handles this transparently unless you specify `include_timestamps=true`.

## Permission System

### Required Permissions

| Endpoint | Required Permission |
|----------|-------------------|
| GET /api/v1/entities/list | `entity:view` |
| POST /api/v1/entities/create | `entity:create` |
| PUT /api/v1/entities/update | `entity:update` |
| DELETE /api/v1/entities | `entity:delete` |
| GET /api/v1/entity-relationships/* | `relation:view` |
| POST /api/v1/entity-relationships | `relation:create` |
| DELETE /api/v1/entity-relationships | `relation:delete` |

### Roles
- `rbac:role:admin` - Full access (includes `rbac:perm:*`)
- `rbac:role:user` - Basic user access

### Wildcard Permissions
- `rbac:perm:*` - All permissions
- `rbac:perm:entity:*` - All entity operations
- `rbac:perm:user:*` - All user operations

## Examples

### Complete Workflow
```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  | jq -r .token)

# 2. Create a document
ENTITY_ID=$(curl -s -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:document", "status:draft", "title:API Guide"],
    "content": {
      "title": "EntityDB API Guide",
      "body": "Complete guide to using the EntityDB API"
    }
  }' | jq -r .id)

# 3. Update the document
curl -X PUT "http://localhost:8085/api/v1/entities/update?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:document", "status:published", "title:API Guide"],
    "content": {
      "title": "EntityDB API Guide",
      "body": "Updated guide with more examples"
    }
  }'

# 4. Query documents
curl -X GET "http://localhost:8085/api/v1/entities/list?tag=type:document" \
  -H "Authorization: Bearer $TOKEN"

# 5. Get temporal history
curl -X GET "http://localhost:8085/api/v1/entities/history?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN"

# 6. Create a relationship
curl -X POST http://localhost:8085/api/v1/entity-relationships \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_id": "'$ENTITY_ID'",
    "target_id": "another_entity_id",
    "relationship_type": "references"
  }'
```

### Working with Large Files
```bash
# Upload a large file (>4MB will auto-chunk)
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:file", "filename:data.csv", "size:10485760"],
    "content": "'$(base64 -w 0 large-file.csv)'"
  }'

# Stream the file back
curl -X GET "http://localhost:8085/api/v1/entities/stream?id=$FILE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  --output retrieved-file.csv
```

### Temporal Queries
```bash
# Get entity state from 1 hour ago
TIMESTAMP=$(date -u -d '1 hour ago' '+%Y-%m-%dT%H:%M:%SZ')
curl -X GET "http://localhost:8085/api/v1/entities/as-of?id=$ENTITY_ID&as_of=$TIMESTAMP" \
  -H "Authorization: Bearer $TOKEN"

# Compare entity between two times
T1=$(date -u -d '2 hours ago' '+%Y-%m-%dT%H:%M:%SZ')
T2=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
curl -X GET "http://localhost:8085/api/v1/entities/diff?id=$ENTITY_ID&t1=$T1&t2=$T2" \
  -H "Authorization: Bearer $TOKEN"
```

## Error Responses

All errors follow this format:
```json
{
  "error": "Error message"
}
```

Common status codes:
- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

## Additional Resources

- [Complete API Reference](./API_REFERENCE_COMPLETE.md) - All 85+ endpoints
- [Query API Guide](./query_api.md) - Advanced query operations
- [Authentication Guide](./auth.md) - Detailed auth documentation
- [Temporal Features](../features/TEMPORAL_FEATURES.md) - Temporal capabilities
- [Examples](./examples.md) - More code examples

---

For the latest updates and complete documentation, visit the [EntityDB Documentation](/docs/README.md).