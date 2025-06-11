# EntityDB API Reference (Complete)

**Version**: 2.29.0  
**Last Updated**: 2025-06-08

## Table of Contents
1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Core Entity Operations](#core-entity-operations)
4. [Temporal Operations](#temporal-operations)
5. [Dataset Management](#dataset-management)
6. [Entity Relationships](#entity-relationships)
7. [User Management](#user-management)
8. [Configuration & Feature Flags](#configuration--feature-flags)
9. [Metrics & Monitoring](#metrics--monitoring)
10. [Health & Status](#health--status)
11. [RBAC & Security](#rbac--security)
12. [Error Handling](#error-handling)
13. [Examples](#examples)

## Overview

### Base URL
```
http://localhost:8085
https://localhost:8443 (when SSL enabled)
```

### API Version
All endpoints are prefixed with `/api/v1/`

### Content Types
- Request: `application/json`
- Response: `application/json` (except for Prometheus metrics endpoint)

### Authentication
Most endpoints require authentication via Bearer token in the Authorization header:
```
Authorization: Bearer <token>
```

## Authentication

> **Authentication Architecture v2.29.0+**: EntityDB now uses embedded credentials stored directly in user entity content. No separate credential entities or relationships are needed.

### Login
Authenticate and receive a session token.

```http
POST /api/v1/auth/login
```

**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "roles": ["admin", "user"]
  },
  "expires_at": "2025-01-01T12:00:00Z"
}
```

### Logout
Invalidate the current session.

```http
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

**Response:**
```json
{
  "status": "ok",
  "message": "Logged out successfully"
}
```

### Who Am I
Get information about the currently authenticated user.

```http
GET /api/v1/auth/whoami
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "admin",
  "roles": ["admin", "user"]
}
```

### Refresh Token
Refresh the session token to extend expiration.

```http
POST /api/v1/auth/refresh
Authorization: Bearer <token>
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2025-01-01T14:00:00Z"
}
```

### Check Status
Check authentication status and session validity.

```http
GET /api/v1/auth/status
Authorization: Bearer <token>
```

**Response:**
```json
{
  "authenticated": true,
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "roles": ["admin", "user"]
  },
  "expires_at": "2025-01-01T12:00:00Z"
}
```

## Core Entity Operations

### List Entities
List all entities with optional filtering.

```http
GET /api/v1/entities/list
Authorization: Bearer <token>
```

**Query Parameters:**
- `tag` (string, optional): Filter by tag (e.g., "type:user")
- `wildcard` (string, optional): Filter by wildcard pattern
- `search` (string, optional): Search content
- `contentType` (string, optional): Content type for search
- `namespace` (string, optional): Filter by namespace
- `include_timestamps` (boolean, optional): Include temporal timestamps in tags

**Response:**
```json
{
  "status": "ok",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "tags": ["type:user", "status:active", "rbac:role:admin"],
      "content": "base64_encoded_content",
      "created_at": 1737564000000000000,
      "updated_at": 1737564000000000000
    }
  ],
  "count": 1
}
```

### Get Entity
Retrieve a single entity by ID.

```http
GET /api/v1/entities/get?id=<entity_id>
Authorization: Bearer <token>
```

**Query Parameters:**
- `id` (string, required): Entity ID
- `include_timestamps` (boolean, optional): Include temporal timestamps in tags

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tags": ["type:user", "status:active"],
  "content": "base64_encoded_content",
  "created_at": 1737564000000000000,
  "updated_at": 1737564000000000000
}
```

### Create Entity
Create a new entity.

```http
POST /api/v1/entities/create
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "id": "optional_custom_id",
  "tags": ["type:document", "status:draft"],
  "content": "string_or_base64_or_json_object"
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tags": ["type:document", "status:draft"],
  "content": "...",
  "created_at": 1737564000000000000,
  "updated_at": 1737564000000000000
}
```

### Update Entity
Update an existing entity.

```http
PUT /api/v1/entities/update?id=<entity_id>
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "tags": ["type:document", "status:published"],
  "content": "updated_content"
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tags": ["type:document", "status:published"],
  "content": "updated_content",
  "created_at": 1737564000000000000,
  "updated_at": 1737565000000000000
}
```

### Query Entities (Advanced)
Query entities with advanced filtering, sorting, and pagination.

```http
GET /api/v1/entities/query
Authorization: Bearer <token>
```

**Query Parameters:**
- `filter` (string): Filter field (e.g., "created_at", "tag:type")
- `operator` (string): Filter operator (eq, ne, gt, lt, gte, lte, like, in)
- `value` (string): Filter value
- `sort` (string): Sort field (created_at, updated_at, id, tag_count)
- `order` (string): Sort order (asc, desc)
- `limit` (integer): Limit results
- `offset` (integer): Offset results

**Response:**
```json
{
  "entities": [...],
  "total": 100,
  "limit": 10,
  "offset": 0
}
```

### Stream Entity Content
Stream large entity content.

```http
GET /api/v1/entities/stream?id=<entity_id>
Authorization: Bearer <token>
```

**Response:** Binary stream of entity content

### Download Entity
Download entity content as a file.

```http
GET /api/v1/entities/download?id=<entity_id>
Authorization: Bearer <token>
```

**Response:** File download with appropriate Content-Type and Content-Disposition headers

## Temporal Operations

### Get Entity As-Of Timestamp
Retrieve an entity as it existed at a specific point in time.

```http
GET /api/v1/entities/as-of?id=<entity_id>&as_of=<timestamp>
Authorization: Bearer <token>
```

**Query Parameters:**
- `id` (string, required): Entity ID
- `as_of` (string, required): Timestamp in RFC3339 format

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tags": ["type:document", "status:draft"],
  "content": "...",
  "created_at": 1737564000000000000,
  "updated_at": 1737564000000000000
}
```

### Get Entity History
Retrieve the history of an entity within a time range.

```http
GET /api/v1/entities/history?id=<entity_id>&from=<timestamp>&to=<timestamp>
Authorization: Bearer <token>
```

**Query Parameters:**
- `id` (string, required): Entity ID
- `from` (string, optional): Start timestamp in RFC3339 format (default: 24 hours ago)
- `to` (string, optional): End timestamp in RFC3339 format (default: now)

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "tags": ["type:document", "status:draft"],
    "content": "...",
    "created_at": 1737564000000000000,
    "updated_at": 1737564000000000000
  }
]
```

### Get Recent Changes
Retrieve entities that have changed since a given timestamp.

```http
GET /api/v1/entities/changes?since=<timestamp>
Authorization: Bearer <token>
```

**Query Parameters:**
- `since` (string, optional): Timestamp in RFC3339 format (default: 1 hour ago)

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "tags": ["type:document", "status:published"],
    "content": "...",
    "created_at": 1737564000000000000,
    "updated_at": 1737565000000000000
  }
]
```

### Get Entity Diff
Compare an entity at two different points in time.

```http
GET /api/v1/entities/diff?id=<entity_id>&t1=<timestamp>&t2=<timestamp>
Authorization: Bearer <token>
```

**Query Parameters:**
- `id` (string, required): Entity ID
- `t1` (string, required): First timestamp in RFC3339 format
- `t2` (string, required): Second timestamp in RFC3339 format

**Response:**
```json
{
  "t1": {
    "tags": ["type:document", "status:draft"],
    "content": "original_content"
  },
  "t2": {
    "tags": ["type:document", "status:published"],
    "content": "updated_content"
  },
  "changes": {
    "tags": {
      "added": ["status:published"],
      "removed": ["status:draft"]
    },
    "content_changed": true
  }
}
```

## Dataset Management

### List Datasets
List all datasets.

```http
GET /api/v1/datasets
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "default",
    "name": "Default Dataset",
    "description": "Main dataset",
    "created_at": "2025-01-01T00:00:00Z"
  }
]
```

### Create Dataset
Create a new dataset.

```http
POST /api/v1/datasets
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "id": "metrics",
  "name": "Metrics Dataset",
  "description": "Dataset for system metrics"
}
```

### Get Dataset
Get details of a specific dataset.

```http
GET /api/v1/datasets/{id}
Authorization: Bearer <token>
```

### Update Dataset
Update dataset details.

```http
PUT /api/v1/datasets/{id}
Authorization: Bearer <token>
```

### Delete Dataset
Delete a dataset.

```http
DELETE /api/v1/datasets/{id}
Authorization: Bearer <token>
```

### Create Entity in Dataset
Create an entity within a specific dataset.

```http
POST /api/v1/datasets/entities/create
Authorization: Bearer <token>
X-Dataset: <dataset_id>
```

### Query Entities in Dataset
Query entities within a specific dataset.

```http
GET /api/v1/datasets/entities/query
Authorization: Bearer <token>
X-Dataset: <dataset_id>
```

## Entity Relationships

### Create Relationship
Create a relationship between two entities.

```http
POST /api/v1/entity-relationships
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "source_id": "550e8400-e29b-41d4-a716-446655440000",
  "target_id": "660e8400-e29b-41d4-a716-446655440001",
  "relationship_type": "contains"
}
```

**Response:**
```json
{
  "id": "rel_770e8400-e29b-41d4-a716-446655440002",
  "source_id": "550e8400-e29b-41d4-a716-446655440000",
  "target_id": "660e8400-e29b-41d4-a716-446655440001",
  "relationship_type": "contains",
  "created_at": "2025-01-01T10:00:00Z"
}
```

### Get Relationship
Get a specific relationship.

```http
GET /api/v1/entity-relationships?source_id=<id>&relationship_type=<type>&target_id=<id>
Authorization: Bearer <token>
```

### List Relationships by Source
List all relationships where the entity is the source.

```http
GET /api/v1/entity-relationships/by-source?source_id=<entity_id>&relationship_type=<type>
Authorization: Bearer <token>
```

### List Relationships by Target
List all relationships where the entity is the target.

```http
GET /api/v1/entity-relationships/by-target?target_id=<entity_id>&relationship_type=<type>
Authorization: Bearer <token>
```

### List Relationships by Type
List all relationships of a specific type.

```http
GET /api/v1/entity-relationships/by-type?relationship_type=<type>
Authorization: Bearer <token>
```

### Delete Relationship
Delete a relationship.

```http
DELETE /api/v1/entity-relationships?source_id=<id>&relationship_type=<type>&target_id=<id>
Authorization: Bearer <token>
```

## User Management

### Create User
Create a new user with embedded credentials (requires admin permission). Creates a single user entity with credentials stored in the content field.

```http
POST /api/v1/users/create
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "username": "newuser",
  "password": "secure_password",
  "email": "user@example.com",
  "full_name": "New User",
  "role": "user"
}
```

**Response:**
```json
{
  "id": "880e8400-e29b-41d4-a716-446655440003",
  "tags": ["type:user", "rbac:role:user", "status:active"],
  "created_at": 1737564000000000000
}
```

### Change Password
Change the current user's password.

```http
POST /api/v1/users/change-password
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "username": "currentuser",
  "current_password": "old_password",
  "new_password": "new_secure_password"
}
```

**Response:**
```json
{
  "status": "ok"
}
```

### Reset Password
Reset a user's password (requires admin permission).

```http
POST /api/v1/users/reset-password
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "username": "targetuser",
  "password": "new_password"
}
```

## Configuration & Feature Flags

### Get Configuration
Retrieve system configuration.

```http
GET /api/v1/config?namespace=<namespace>&key=<key>
Authorization: Bearer <token>
```

**Query Parameters:**
- `namespace` (string, optional): Configuration namespace
- `key` (string, optional): Configuration key

**Response:**
```json
[
  {
    "id": "config_123",
    "tags": ["type:config", "conf:system:max_connections"],
    "content": "100"
  }
]
```

### Set Configuration
Update configuration values.

```http
POST /api/v1/config/set
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "namespace": "system",
  "key": "max_connections",
  "value": "200"
}
```

### Get Feature Flags
Retrieve feature flags.

```http
GET /api/v1/feature-flags?stage=<stage>
Authorization: Bearer <token>
```

**Query Parameters:**
- `stage` (string, optional): Filter by stage (alpha, beta, stable)

### Set Feature Flag
Update a feature flag.

```http
POST /api/v1/feature-flags/set
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "flag": "new_ui",
  "enabled": true
}
```

## Metrics & Monitoring

### Health Check
Basic health check endpoint (no authentication required).

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "version": "2.19.0",
  "uptime": "2h15m30s",
  "timestamp": "2025-01-01T10:00:00Z",
  "checks": {
    "database": "ok",
    "storage": "ok",
    "memory": "ok"
  },
  "metrics": {
    "entity_count": 1000,
    "user_count": 10,
    "database_size_bytes": 1048576,
    "memory_usage": {
      "alloc_bytes": 10485760,
      "total_alloc_bytes": 20971520,
      "sys_bytes": 73400320,
      "num_gc": 5
    },
    "goroutines": 25
  }
}
```

### Prometheus Metrics
Prometheus-format metrics endpoint (no authentication required).

```http
GET /metrics
```

**Response:** Text format metrics following Prometheus exposition format.

### System Metrics
Comprehensive EntityDB system metrics (no authentication required).

```http
GET /api/v1/system/metrics
```

**Response:**
```json
{
  "system": {
    "version": "2.19.0",
    "uptime": 132.651279965,
    "go_version": "go1.24.2",
    "num_cpu": 8,
    "num_goroutines": 24
  },
  "memory": {
    "alloc_bytes": 10031960,
    "heap_alloc_bytes": 10031960,
    "heap_sys_bytes": 16121856,
    "total_alloc_bytes": 10685192
  },
  "database": {
    "total_entities": 1000,
    "entities_by_type": {
      "user": 10,
      "document": 500,
      "metric": 490
    },
    "tags_total": 5000,
    "tags_unique": 250,
    "avg_tags_per_entity": 5.0
  },
  "storage": {
    "database_size_bytes": 52428800,
    "wal_size_bytes": 1048576,
    "index_size_bytes": 524288,
    "compression_ratio": 0.75
  },
  "performance": {
    "query_cache_hits": 150,
    "query_cache_miss": 50,
    "index_lookups": 1000,
    "gc_runs": 3,
    "last_gc_pause_ns": 140707
  },
  "temporal": {
    "temporal_tags_count": 4500,
    "non_temporal_tags_count": 500,
    "temporal_tags_ratio": 0.9,
    "time_range_start": "2025-01-01T00:00:00Z",
    "time_range_end": "2025-01-01T10:00:00Z"
  }
}
```

### Metric History
Get historical values for a specific metric.

```http
GET /api/v1/metrics/history?metric_name=<name>&hours=<hours>&limit=<limit>
Authorization: Bearer <token>
```

**Query Parameters:**
- `metric_name` (string, required): Metric name (e.g., "memory_alloc", "entity_count_total")
- `hours` (integer, optional): Number of hours to look back (default: 24)
- `limit` (integer, optional): Maximum data points (default: 100)

**Response:**
```json
{
  "metric_name": "memory_alloc",
  "unit": "bytes",
  "start_time": "2025-01-01T00:00:00Z",
  "end_time": "2025-01-02T00:00:00Z",
  "count": 96,
  "data_points": [
    {
      "timestamp": "2025-01-01T00:00:00Z",
      "value": 10485760
    }
  ]
}
```

### Available Metrics
List all available metrics being collected.

```http
GET /api/v1/metrics/available
Authorization: Bearer <token>
```

**Response:**
```json
[
  "memory_alloc",
  "memory_heap_alloc",
  "memory_sys",
  "entity_count_total",
  "entity_count_by_type",
  "database_size",
  "wal_size",
  "goroutines",
  "gc_runs"
]
```

### RBAC Metrics
Get RBAC and session metrics (requires admin).

```http
GET /api/v1/rbac/metrics
Authorization: Bearer <token>
```

**Response:**
```json
{
  "timestamp": "2025-01-01T10:00:00Z",
  "users": {
    "total_users": 10,
    "admin_count": 2
  },
  "sessions": {
    "active_count": 5,
    "total_today": 25,
    "avg_duration_ms": 3600000
  },
  "auth": {
    "successful_logins": 23,
    "failed_logins": 2,
    "success_rate": 0.92
  },
  "permissions": {
    "total_checks": 10000,
    "checks_per_second": 2.5,
    "cache_hit_rate": 0.85
  },
  "security_events": [
    {
      "id": "event_123",
      "type": "failed_login",
      "username": "unknown",
      "timestamp": "2025-01-01T09:00:00Z",
      "status": "blocked",
      "details": "Invalid credentials"
    }
  ]
}
```

### Public RBAC Metrics
Basic RBAC metrics (no authentication required).

```http
GET /api/v1/rbac/metrics/public
```

**Response:**
```json
{
  "timestamp": "2025-01-01T10:00:00Z",
  "sessions": {
    "active_count": 5
  },
  "auth": {
    "successful_logins": 23,
    "failed_logins": 2,
    "success_rate": 0.92
  }
}
```

### Worca Metrics
Workforce orchestrator specific metrics.

```http
GET /api/v1/worca/metrics
Authorization: Bearer <token>
```

### Dashboard Stats
Get dashboard statistics and recent activity.

```http
GET /api/v1/dashboard/stats
Authorization: Bearer <token>
```

**Response:**
```json
{
  "user_count": 10,
  "workspace_count": 5,
  "issue_stats": {
    "total": 100,
    "by_status": {
      "open": 45,
      "in_progress": 30,
      "closed": 25
    },
    "by_priority": {
      "critical": 5,
      "high": 20,
      "medium": 50,
      "low": 25
    }
  },
  "recent_activity": [
    {
      "type": "entity_created",
      "timestamp": "2025-01-01T09:55:00Z",
      "description": "User 'john' created document 'API Guide'"
    }
  ]
}
```

## Health & Status

### API Status
Simple API status check.

```http
GET /api/v1/status
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2025-01-01T10:00:00Z",
  "api_status": "connected"
}
```

### Admin Health Check
Detailed health check (requires admin).

```http
GET /api/v1/admin/health
Authorization: Bearer <token>
```

### Admin Reindex
Trigger index rebuild (requires admin).

```http
POST /api/v1/admin/reindex
Authorization: Bearer <token>
```

## RBAC & Security

### Permission Model
EntityDB uses tag-based RBAC with the following permission format:
```
rbac:perm:<resource>:<action>
```

### Core Permissions
- `rbac:perm:*` - All permissions
- `rbac:perm:entity:*` - All entity operations
- `rbac:perm:entity:view` - View entities
- `rbac:perm:entity:create` - Create entities
- `rbac:perm:entity:update` - Update entities
- `rbac:perm:entity:delete` - Delete entities
- `rbac:perm:user:*` - All user operations
- `rbac:perm:user:create` - Create users
- `rbac:perm:user:update` - Update users
- `rbac:perm:user:delete` - Delete users
- `rbac:perm:relation:*` - All relationship operations
- `rbac:perm:system:*` - All system operations
- `rbac:perm:config:*` - All configuration operations

### Roles
- `rbac:role:admin` - Administrator role (includes `rbac:perm:*`)
- `rbac:role:user` - Regular user role

## Error Handling

### Error Response Format
All errors follow this format:
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional context"
  }
}
```

### Common Error Codes
- `400 Bad Request` - Invalid request format or parameters
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate ID)
- `500 Internal Server Error` - Server error

## Examples

### Complete Authentication Flow
```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  | jq -r .token)

# 2. Create an entity
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:document", "status:draft", "author:admin"],
    "content": {
      "title": "API Documentation",
      "body": "Complete API reference for EntityDB"
    }
  }'

# 3. Query entities by tag
curl -X GET "http://localhost:8085/api/v1/entities/list?tag=type:document" \
  -H "Authorization: Bearer $TOKEN"

# 4. Get temporal history
curl -X GET "http://localhost:8085/api/v1/entities/history?id=<entity_id>&from=2025-01-01T00:00:00Z" \
  -H "Authorization: Bearer $TOKEN"

# 5. Create a relationship
curl -X POST http://localhost:8085/api/v1/entity-relationships \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_id": "<parent_id>",
    "target_id": "<child_id>",
    "relationship_type": "contains"
  }'

# 6. Logout
curl -X POST http://localhost:8085/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

### Working with Large Files
```bash
# Upload a large file (auto-chunking enabled)
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:file", "filename:large-dataset.csv"],
    "content": "'$(base64 -w 0 large-dataset.csv)'"
  }'

# Stream the file content
curl -X GET "http://localhost:8085/api/v1/entities/stream?id=<entity_id>" \
  -H "Authorization: Bearer $TOKEN" \
  --output retrieved-file.csv

# Download with proper headers
curl -X GET "http://localhost:8085/api/v1/entities/download?id=<entity_id>" \
  -H "Authorization: Bearer $TOKEN" \
  -O -J
```

### Temporal Queries
```bash
# Get entity state from yesterday
curl -X GET "http://localhost:8085/api/v1/entities/as-of?id=<entity_id>&as_of=2025-01-01T00:00:00Z" \
  -H "Authorization: Bearer $TOKEN"

# Get all changes in the last hour
curl -X GET "http://localhost:8085/api/v1/entities/changes?since=2025-01-01T09:00:00Z" \
  -H "Authorization: Bearer $TOKEN"

# Compare entity at two points
curl -X GET "http://localhost:8085/api/v1/entities/diff?id=<entity_id>&t1=2025-01-01T00:00:00Z&t2=2025-01-01T12:00:00Z" \
  -H "Authorization: Bearer $TOKEN"
```

### Advanced Queries
```bash
# Query with filtering and sorting
curl -X GET "http://localhost:8085/api/v1/entities/query?filter=created_at&operator=gte&value=2025-01-01T00:00:00Z&sort=created_at&order=desc&limit=10" \
  -H "Authorization: Bearer $TOKEN"

# Query by tag value
curl -X GET "http://localhost:8085/api/v1/entities/query?filter=tag:type&operator=eq&value=document&sort=updated_at&order=desc" \
  -H "Authorization: Bearer $TOKEN"
```

## Notes

1. **Temporal Storage**: All tags are stored with nanosecond precision timestamps in the format `TIMESTAMP|tag`. The API transparently handles this unless `include_timestamps=true` is specified.

2. **Autochunking**: Files larger than 4MB are automatically chunked for efficient storage and retrieval.

3. **Binary Format**: EntityDB uses a custom binary format (EBF) for storage with Write-Ahead Logging for durability.

4. **Performance**: The system is optimized for high-performance temporal queries with memory-mapped files and various indexing strategies.

5. **Security**: All endpoints (except health/metrics) require authentication. RBAC is enforced at the tag level.

## Version History

- **v2.19.0** (2025-05-30): Current version with complete API documentation
- **v2.18.0**: Added logging standards and code cleanup
- **v2.17.0**: Fixed metrics endpoints and background collection
- **v2.16.0**: UUID storage fix for authentication
- **v2.13.0**: Configuration system overhaul

---

For more information, see the [EntityDB Documentation](/docs/README.md).