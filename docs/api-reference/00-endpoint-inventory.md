# EntityDB API Endpoint Inventory

> **Version**: v2.32.0 | **Last Verified**: 2025-06-16 | **Status**: AUTHORITATIVE
>
> **Complete inventory of all API endpoints in EntityDB v2.32.0, verified against actual codebase implementation.**

## Endpoint Summary

**Total Endpoints**: 40 (verified against `src/main.go`)
**Authentication Model**: JWT Bearer tokens
**Base URL**: `https://localhost:8085` (SSL enabled by default)

## Authentication Endpoints (4)

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| `POST` | `/api/v1/auth/login` | None | User login with username/password |
| `POST` | `/api/v1/auth/logout` | Authenticated | Invalidate current session |
| `GET` | `/api/v1/auth/whoami` | Authenticated | Get current user information |
| `POST` | `/api/v1/auth/refresh` | None | Refresh session token |

## Entity Operations (13)

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| `GET` | `/api/v1/entities/list` | `entity:view` | List entities with filtering |
| `GET` | `/api/v1/entities/get` | `entity:view` | Retrieve specific entity by ID |
| `POST` | `/api/v1/entities/create` | `entity:create` | Create new entity |
| `PUT` | `/api/v1/entities/update` | `entity:update` | Update existing entity |
| `GET` | `/api/v1/entities/query` | `entity:view` | Advanced query with filtering |
| `GET` | `/api/v1/entities/listbytag` | `entity:view` | List entities by tag (alias) |
| `GET` | `/api/v1/entities/summary` | `entity:view` | Get entity count and stats |
| `GET` | `/api/v1/entities/as-of` | `entity:view` | Get entity state at timestamp |
| `GET` | `/api/v1/entities/history` | `entity:view` | Get entity change history |
| `GET` | `/api/v1/entities/changes` | `entity:view` | Get recent entity changes |
| `GET` | `/api/v1/entities/diff` | `entity:view` | Compare entity states |
| `GET` | `/api/v1/entities/get-chunk` | `entity:view` | Get chunked content |
| `GET` | `/api/v1/entities/stream-content` | `entity:view` | Stream large entity content |

## Dataset Operations (7)

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| `GET` | `/api/v1/datasets` | `dataset:view` | List available datasets |
| `POST` | `/api/v1/datasets` | `dataset:create` | Create new dataset |
| `GET` | `/api/v1/datasets/{name}` | `dataset:view` | Get dataset information |
| `PUT` | `/api/v1/datasets/{name}` | `dataset:update` | Update dataset properties |
| `DELETE` | `/api/v1/datasets/{name}` | `dataset:delete` | Delete dataset |
| `GET` | `/api/v1/datasets/{name}/entities/list` | `entity:view` | List entities in dataset |
| `POST` | `/api/v1/datasets/{name}/entities/create` | `entity:create` | Create entity in dataset |

## User Management (3)

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| `POST` | `/api/v1/users/create` | `user:create` | Create new user (admin only) |
| `POST` | `/api/v1/users/change-password` | `user:update` | Change user password |
| `POST` | `/api/v1/users/reset-password` | `user:update` | Reset user password |

## Metrics & Monitoring (7)

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| `GET` | `/health` | None | System health check |
| `GET` | `/metrics` | None | Prometheus metrics |
| `POST` | `/api/v1/metrics/collect` | `metrics:write` | Collect custom metric |
| `GET` | `/api/v1/metrics/history` | `metrics:read` | Get metrics history |
| `GET` | `/api/v1/metrics/current` | `metrics:read` | Get current metrics |
| `GET` | `/api/v1/metrics/available` | `metrics:read` | List available metrics |
| `GET` | `/api/v1/metrics/comprehensive` | `metrics:read` | Comprehensive metrics data |

## Administrative Operations (6)

| Method | Endpoint | Permission | Description |
|--------|----------|------------|-------------|
| `POST` | `/api/v1/admin/reindex` | `admin:reindex` | Rebuild system indexes |
| `GET` | `/api/v1/admin/health` | `admin:health` | Detailed system health |
| `POST` | `/api/v1/admin/log-level` | `admin:config` | Change runtime log level |
| `POST` | `/api/v1/admin/trace-subsystems` | `admin:config` | Configure trace subsystems |
| `GET` | `/api/v1/status` | None | Server status check |
| `GET` | `/swagger/doc.json` | None | OpenAPI specification |

## Permission Requirements

### Entity Operations
- **View**: `rbac:perm:entity:view` or `rbac:perm:entity:*`
- **Create**: `rbac:perm:entity:create` or `rbac:perm:entity:*`
- **Update**: `rbac:perm:entity:update` or `rbac:perm:entity:*`

### Dataset Operations
- **View**: `rbac:perm:dataset:view` or `rbac:perm:dataset:*`
- **Create**: `rbac:perm:dataset:create` or `rbac:perm:dataset:*`
- **Update**: `rbac:perm:dataset:update` or `rbac:perm:dataset:*`
- **Delete**: `rbac:perm:dataset:delete` or `rbac:perm:dataset:*`

### User Management
- **Create**: `rbac:perm:user:create` or `rbac:perm:user:*` (admin only)
- **Update**: `rbac:perm:user:update` or `rbac:perm:user:*`

### Metrics
- **Read**: `rbac:perm:metrics:read` or `rbac:perm:metrics:*`
- **Write**: `rbac:perm:metrics:write` or `rbac:perm:metrics:*`

### Administrative
- **All admin operations**: `rbac:perm:admin:*` or `rbac:role:admin`

## Response Format Standards

All API responses follow this consistent structure:

```json
{
  "status": "ok|error",
  "message": "Human-readable description",
  "data": { /* Response payload */ },
  "error": "Error details if status=error"
}
```

## Error Codes

- **400**: Bad Request - Invalid input or malformed request
- **401**: Unauthorized - Missing or invalid authentication
- **403**: Forbidden - Insufficient permissions for operation
- **404**: Not Found - Requested resource doesn't exist
- **500**: Internal Server Error - Server-side processing error

## Content Types

- **Request**: `application/json` for all POST/PUT operations
- **Response**: `application/json` for all structured responses
- **Binary**: `application/octet-stream` for raw entity content

---

**Verification Method**: This inventory was generated by parsing `src/main.go` and cross-referencing with actual handler implementations. All endpoints have been verified to exist in the codebase.

**Last Updated**: 2025-06-16 by Technical Documentation Team
**Source**: `/opt/entitydb/src/main.go` lines 150-300 (route registrations)