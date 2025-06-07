# EntityDB API Reference

> **Current Version**: 2.27.0  
> **Base URL**: `https://localhost:8443` (HTTPS) or `http://localhost:8085` (HTTP)  
> **Content-Type**: `application/json`

## Authentication

EntityDB uses session-based authentication. Most endpoints require a valid session token.

### Login

```http
POST /api/v1/auth/login
```

**Request Body:**
```json
{
  "username": "admin",
  "password": "admin"
}
```

**Response:**
```json
{
  "token": "session_token_here",
  "user": {
    "id": "user_entity_id",
    "username": "admin",
    "role": "admin"
  }
}
```

## Entity Operations

### Create Entity

```http
POST /api/v1/entities/create
Authorization: Bearer {session_token}
```

**Request Body:**
```json
{
  "id": "optional_custom_id",
  "tags": ["type:document", "status:draft"],
  "content": "base64_encoded_content_or_raw_data"
}
```

**Response:**
```json
{
  "id": "generated_or_custom_entity_id",
  "tags": ["type:document", "status:draft"],
  "content": "base64_content",
  "created_at": 1749303910369730667,
  "updated_at": 1749303910369730667
}
```

### Get Entity

```http
GET /api/v1/entities/get?id={entity_id}
Authorization: Bearer {session_token}
```

**Query Parameters:**
- `id` (required): Entity identifier
- `include_timestamps` (optional): Include temporal tag timestamps

### Update Entity

```http
PUT /api/v1/entities/update
Authorization: Bearer {session_token}
```

**Request Body:**
```json
{
  "id": "entity_id",
  "tags": ["type:document", "status:published"],
  "content": "updated_content"
}
```

### List Entities

```http
GET /api/v1/entities/list
Authorization: Bearer {session_token}
```

**Query Parameters:**
- `limit` (optional): Maximum number of entities to return
- `offset` (optional): Number of entities to skip
- `include_timestamps` (optional): Include temporal tag timestamps

### Query Entities

```http
GET /api/v1/entities/query
Authorization: Bearer {session_token}
```

**Query Parameters:**
- `tags` (optional): Comma-separated list of tags to filter by
- `content_type` (optional): Filter by content type
- `sort_by` (optional): Sort field (timestamp, id, tag_count)
- `sort_order` (optional): asc or desc
- `limit` (optional): Maximum results
- `offset` (optional): Pagination offset

### List By Tag

```http
GET /api/v1/entities/listbytag?tag={tag_name}
Authorization: Bearer {session_token}
```

**Query Parameters:**
- `tag` (required): Tag to search for
- `include_timestamps` (optional): Include temporal tag timestamps

## Chunked Content Operations

EntityDB automatically chunks large files (>4MB). These endpoints handle chunked content:

### Get Chunk

```http
GET /api/v1/entities/get-chunk?id={entity_id}&chunk={chunk_index}
Authorization: Bearer {session_token}
```

### Stream Content

```http
GET /api/v1/entities/stream-content?id={entity_id}
Authorization: Bearer {session_token}
```

## Temporal Operations

### Entity As-Of (Time Travel)

```http
GET /api/v1/entities/as-of?id={entity_id}&timestamp={timestamp}
Authorization: Bearer {session_token}
```

**Query Parameters:**
- `id` (required): Entity identifier
- `timestamp` (required): Unix nanosecond timestamp or ISO 8601 format

### Entity History

```http
GET /api/v1/entities/history?id={entity_id}
Authorization: Bearer {session_token}
```

**Query Parameters:**
- `id` (required): Entity identifier
- `limit` (optional): Maximum history entries

### Entity Changes

```http
GET /api/v1/entities/changes?id={entity_id}
Authorization: Bearer {session_token}
```

### Entity Diff

```http
GET /api/v1/entities/diff?id={entity_id}&start_time={start}&end_time={end}
Authorization: Bearer {session_token}
```

## Relationship Operations

### Create Relationship

```http
POST /api/v1/entity-relationships
Authorization: Bearer {session_token}
```

**Request Body:**
```json
{
  "source_id": "source_entity_id",
  "target_id": "target_entity_id",
  "relationship_type": "contains"
}
```

### Get Relationships

```http
GET /api/v1/entity-relationships?source_id={source_id}
Authorization: Bearer {session_token}
```

**Query Parameters:**
- `source_id` (optional): Filter by source entity
- `target_id` (optional): Filter by target entity
- `relationship_type` (optional): Filter by relationship type

## Dataspace Operations

Multi-tenant dataspace operations:

### Create Entity in Dataspace

```http
POST /api/v1/dataspaces/{dataspace}/entities/create
Authorization: Bearer {session_token}
```

### Query Dataspace Entities

```http
GET /api/v1/dataspaces/{dataspace}/entities/query
Authorization: Bearer {session_token}
```

## User Management

### Create User

```http
POST /api/v1/users/create
Authorization: Bearer {admin_session_token}
```

**Request Body:**
```json
{
  "username": "newuser",
  "password": "secure_password",
  "role": "user"
}
```

**Required Permission:** `rbac:perm:user:create`

## Admin Operations

### Log Level Control

```http
GET /api/v1/admin/log-level
POST /api/v1/admin/log-level
Authorization: Bearer {admin_session_token}
```

**POST Request Body:**
```json
{
  "level": "debug"
}
```

### Trace Subsystem Control

```http
GET /api/v1/admin/trace-subsystems
POST /api/v1/admin/trace-subsystems
Authorization: Bearer {admin_session_token}
```

**POST Request Body:**
```json
{
  "subsystems": ["auth", "storage", "temporal"]
}
```

## Metrics & Monitoring

### Health Check

```http
GET /health
```

**No authentication required.** Returns system health status.

### System Metrics

```http
GET /api/v1/system/metrics
```

**No authentication required.** Returns comprehensive system metrics.

### Prometheus Metrics

```http
GET /metrics
```

**No authentication required.** Returns metrics in Prometheus format.

### RBAC Metrics

```http
GET /api/v1/rbac/metrics
Authorization: Bearer {admin_session_token}
```

**Required Permission:** `rbac:perm:system:view`

### Public RBAC Metrics

```http
GET /api/v1/rbac/metrics/public
```

**No authentication required.** Returns basic authentication statistics.

### Application Metrics

```http
GET /api/v1/application/metrics?namespace={app_name}
Authorization: Bearer {session_token}
```

**Required Permission:** `rbac:perm:metrics:read`

### Metrics History

```http
GET /api/v1/metrics/history?metric={metric_name}&period={time_period}
```

**No authentication required.** Returns historical metric data.

### Available Metrics

```http
GET /api/v1/metrics/available
```

**No authentication required.** Lists all available metrics.

## Configuration

### Get Configuration

```http
GET /api/v1/config
Authorization: Bearer {session_token}
```

### Set Feature Flag

```http
POST /api/v1/feature-flags/set
Authorization: Bearer {admin_session_token}
```

**Request Body:**
```json
{
  "flag": "feature_name",
  "enabled": true
}
```

## Error Responses

All endpoints return errors in this format:

```json
{
  "error": "Error description",
  "code": "ERROR_CODE",
  "timestamp": 1749303910369730667
}
```

**Common HTTP Status Codes:**
- `200` - Success
- `400` - Bad Request (invalid parameters)
- `401` - Unauthorized (missing or invalid session)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (entity or resource doesn't exist)
- `500` - Internal Server Error

## RBAC Permissions

EntityDB uses hierarchical permissions with the format `rbac:perm:resource:action`:

### Entity Permissions
- `rbac:perm:entity:view` - View entities
- `rbac:perm:entity:create` - Create entities
- `rbac:perm:entity:update` - Update entities
- `rbac:perm:entity:delete` - Delete entities

### System Permissions
- `rbac:perm:system:view` - View system information
- `rbac:perm:system:admin` - Administrative access

### User Permissions
- `rbac:perm:user:create` - Create users
- `rbac:perm:user:view` - View user information

### Metrics Permissions
- `rbac:perm:metrics:read` - Read application metrics

### Admin Roles
- `rbac:role:admin` - Full administrative access
- `rbac:role:user` - Standard user access

---

**Note:** All timestamps in EntityDB are stored as nanoseconds since Unix epoch for maximum precision. The API accepts both nanosecond timestamps and ISO 8601 formatted dates.