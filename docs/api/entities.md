# EntityDB API Reference

This document provides a comprehensive reference for the EntityDB REST API.

## Authentication

All API endpoints (except `/api/v1/auth/login`) require JWT authentication.

### Headers
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

## Base URL
```
http://localhost:8085
```

## Endpoints

### Authentication

#### Login
```http
POST /api/v1/auth/login
```

Request:
```json
{
  "username": "admin",
  "password": "password"
}
```

Response:
```json
{
  "token": "tk_admin_1234567890",
  "user": {
    "id": "usr_admin",
    "username": "admin",
    "roles": ["admin"]
  }
}
```

#### Logout
```http
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

Response:
```json
{
  "status": "ok",
  "message": "Logged out successfully"
}
```

#### Status
```http
GET /api/v1/auth/status
Authorization: Bearer <token>
```

Response:
```json
{
  "status": "ok",
  "user": {
    "id": "usr_admin",
    "username": "admin",
    "roles": ["admin"]
  }
}
```

### Entities

#### List Entities
```http
GET /api/v1/entities/list?type=<type>&tags=<tags>&status=<status>
Authorization: Bearer <token>
```

Query Parameters:
- `type` (optional): Filter by entity type
- `tags` (optional): Comma-separated list of tags
- `status` (optional): Filter by status

Response:
```json
{
  "status": "ok",
  "data": [
    {
      "id": "entity_123",
      "type": "issue",
      "title": "Sample Issue",
      "description": "Issue description",
      "status": "active",
      "tags": ["type:issue", "priority:high"],
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

#### Create Entity
```http
POST /api/v1/entities
Authorization: Bearer <token>
```

Request:
```json
{
  "type": "issue",
  "title": "New Issue",
  "description": "Detailed description",
  "tags": ["priority:high", "status:pending"],
  "properties": {
    "custom_field": "value"
  }
}
```

Response:
```json
{
  "status": "ok",
  "message": "Entity created successfully",
  "data": {
    "id": "entity_1234567890",
    "type": "issue",
    "title": "New Issue",
    "description": "Detailed description",
    "tags": ["priority:high", "status:pending"],
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Get Entity
```http
GET /api/v1/entities/get?id=<entity_id>
Authorization: Bearer <token>
```

Response:
```json
{
  "status": "ok",
  "data": {
    "id": "entity_123",
    "type": "issue",
    "title": "Sample Issue",
    "description": "Issue description",
    "status": "active",
    "tags": ["type:issue", "priority:high"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Update Entity
```http
PUT /api/v1/entities/update
Authorization: Bearer <token>
```

Request:
```json
{
  "id": "entity_123",
  "title": "Updated Title",
  "tags": ["priority:low", "status:completed"]
}
```

Response:
```json
{
  "status": "ok",
  "message": "Entity updated successfully",
  "data": {
    "id": "entity_123",
    "title": "Updated Title",
    "tags": ["priority:low", "status:completed"],
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Entity Relationships

#### Create Relationship
```http
POST /api/v1/entity-relationships
Authorization: Bearer <token>
```

Request:
```json
{
  "source_id": "entity_123",
  "target_id": "entity_456",
  "type": "depends_on",
  "properties": {
    "weight": 1
  }
}
```

Response:
```json
{
  "status": "ok",
  "message": "Relationship created successfully",
  "data": {
    "id": "rel_789",
    "source_id": "entity_123",
    "target_id": "entity_456",
    "type": "depends_on",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### List Relationships by Source
```http
GET /api/v1/entity-relationships/source?source=<entity_id>
Authorization: Bearer <token>
```

Response:
```json
{
  "status": "ok",
  "data": [
    {
      "id": "rel_789",
      "source_id": "entity_123",
      "target_id": "entity_456",
      "type": "depends_on",
      "properties": {
        "weight": 1
      }
    }
  ]
}
```

#### List Relationships by Target
```http
GET /api/v1/entity-relationships/target?target=<entity_id>
Authorization: Bearer <token>
```

#### Delete Relationship
```http
DELETE /api/v1/entity-relationships?id=<relationship_id>
Authorization: Bearer <token>
```

## Tag System

### Tag Format
```
namespace:category:subcategory:value
```

### Core Namespaces

1. **type:** - Entity classification
   - `type:user`
   - `type:agent`
   - `type:issue`
   - `type:workspace`

2. **rbac:** - Access control
   - `rbac:role:admin`
   - `rbac:role:user`
   - `rbac:perm:entity:create`
   - `rbac:perm:issue:*`
   - `rbac:perm:*`

3. **status:** - Entity state
   - `status:active`
   - `status:pending`
   - `status:completed`
   - `status:archived`

4. **id:** - Unique identifiers
   - `id:username:admin`
   - `id:agent:claude-2`
   - `id:issue:issue_123`

5. **priority:** - Priority levels
   - `priority:critical`
   - `priority:high`
   - `priority:medium`
   - `priority:low`

## Permission System

### Required Permissions

| Endpoint | Required Permission |
|----------|-------------------|
| GET /api/v1/entities/list | rbac:perm:entity:read |
| POST /api/v1/entities | rbac:perm:entity:create |
| PUT /api/v1/entities/update | rbac:perm:entity:update |
| DELETE /api/v1/entities | rbac:perm:entity:delete |
| GET /api/v1/entity-relationships/* | rbac:perm:relationship:read |
| POST /api/v1/entity-relationships | rbac:perm:relationship:create |
| DELETE /api/v1/entity-relationships | rbac:perm:relationship:delete |

### Wildcard Permissions

- `rbac:perm:*` - All permissions
- `rbac:perm:entity:*` - All entity operations
- `rbac:perm:issue:*` - All issue operations

## Error Responses

### 400 Bad Request
```json
{
  "status": "error",
  "message": "Invalid request format"
}
```

### 401 Unauthorized
```json
{
  "status": "error",
  "message": "Authentication required"
}
```

### 403 Forbidden
```json
{
  "status": "error",
  "message": "Permission denied"
}
```

### 404 Not Found
```json
{
  "status": "error",
  "message": "Entity not found"
}
```

### 500 Internal Server Error
```json
{
  "status": "error",
  "message": "Internal server error"
}
```

## Rate Limiting

Currently no rate limiting is implemented.

## Pagination

Pagination is not yet implemented. All list endpoints return complete results.

## Examples

### Create Issue with Authentication
```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r .token)

# Create issue
curl -X POST http://localhost:8085/api/v1/entities \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "issue",
    "title": "Fix login bug",
    "description": "Users cannot login with special characters",
    "tags": ["priority:high", "status:pending", "component:auth"]
  }'

# List issues
curl -X GET "http://localhost:8085/api/v1/entities/list?type=issue" \
  -H "Authorization: Bearer $TOKEN"
```

### Create Entity Relationship
```bash
# Create dependency between issues
curl -X POST http://localhost:8085/api/v1/entity-relationships \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_id": "entity_123",
    "target_id": "entity_456",
    "type": "depends_on"
  }'
```