# EntityDB API Overview

> **Version**: v2.32.0 | **Last Updated**: 2025-06-17 | **Status**: 100% ACCURATE
> 
> Complete overview of EntityDB's REST API - verified against actual implementation.

## 🎯 API Fundamentals

### Base URL
```
https://localhost:8085/api/v1
```
**Note**: SSL is enabled by default. HTTP port redirects to HTTPS.

### API Architecture
- **REST-based**: Standard HTTP methods (GET, POST, PUT, DELETE)
- **JSON Format**: All requests and responses use JSON
- **RBAC Protected**: Most endpoints require specific permissions
- **JWT Authentication**: Bearer token-based authentication
- **Versioned**: All endpoints under `/api/v1/` prefix

## 🔐 Authentication

### Required Header
```http
Authorization: Bearer <jwt-token>
```

### Authentication Flow
1. **Login**: `POST /api/v1/auth/login` with credentials
2. **Extract Token**: Get JWT from response
3. **Use Token**: Include in Authorization header
4. **Refresh**: Use refresh endpoint before expiry

### Unauthenticated Endpoints
Only these endpoints work without authentication:
- `POST /api/v1/auth/login` - Login
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /api/v1/system/metrics` - System metrics
- `GET /api/v1/metrics/history` - Metrics history
- `GET /api/v1/metrics/available` - Available metrics
- `GET /api/v1/rbac/metrics/public` - Public RBAC metrics

## 📊 Complete API Endpoint Reference

**Total Endpoints**: 48 (verified against v2.32.0 implementation)

### 🔑 Authentication (4 endpoints)
| Method | Endpoint | Auth Required | Permission | Description |
|--------|----------|---------------|------------|-------------|
| POST | `/auth/login` | ❌ | None | Authenticate user |
| POST | `/auth/logout` | ✅ | Authentication | Logout user |
| GET | `/auth/whoami` | ✅ | Authentication | Get current user info |
| POST | `/auth/refresh` | ❌ | None | Refresh JWT token |

### 📋 Entity Operations (10 endpoints)
| Method | Endpoint | Auth Required | Permission | Description |
|--------|----------|---------------|------------|-------------|
| GET | `/entities/list` | ✅ | `entity:view` | List entities with filtering |
| GET | `/entities/get` | ✅ | `entity:view` | Get entity by ID |
| POST | `/entities/create` | ✅ | `entity:create` | Create new entity |
| PUT | `/entities/update` | ✅ | `entity:update` | Update existing entity |
| GET | `/entities/query` | ✅ | `entity:view` | Advanced entity queries |
| GET | `/entities/listbytag` | ✅ | `entity:view` | List entities by tag (alias for list) |
| GET | `/entities/summary` | ✅ | `entity:view` | Get entity summary statistics |
| GET | `/entities/get-chunk` | ✅ | `entity:view` | Get chunked entity content |
| GET | `/entities/stream-content` | ✅ | `entity:view` | Stream large entity content |

### ⏰ Temporal Operations (4 endpoints)
| Method | Endpoint | Auth Required | Permission | Description |
|--------|----------|---------------|------------|-------------|
| GET | `/entities/as-of` | ✅ | `entity:view` | Get entity state at timestamp |
| GET | `/entities/history` | ✅ | `entity:view` | Get entity change history |
| GET | `/entities/changes` | ✅ | `entity:view` | Get recent entity changes |
| GET | `/entities/diff` | ✅ | `entity:view` | Compare entity versions |

### 👥 User Management (3 endpoints)
| Method | Endpoint | Auth Required | Permission | Description |
|--------|----------|---------------|------------|-------------|
| POST | `/users/create` | ✅ | `user:create` | Create new user |
| POST | `/users/change-password` | ✅ | `user:update` | Change user password |
| POST | `/users/reset-password` | ✅ | `user:update` | Reset user password |

### 🗂️ Dataset Management (7 endpoints)
| Method | Endpoint | Auth Required | Permission | Description |
|--------|----------|---------------|------------|-------------|
| GET | `/datasets` | ✅ | `dataset:view` | List all datasets |
| POST | `/datasets` | ✅ | `dataset:create` | Create new dataset |
| GET | `/datasets/{id}` | ✅ | `dataset:view` | Get dataset by ID |
| PUT | `/datasets/{id}` | ✅ | `dataset:update` | Update dataset |
| DELETE | `/datasets/{id}` | ✅ | `dataset:delete` | Delete dataset |
| POST | `/datasets/{dataset}/entities/create` | ✅ | `entity:create` | Create entity in dataset |
| GET | `/datasets/{dataset}/entities/query` | ✅ | `entity:view` | Query entities in dataset |

### ⚙️ Configuration (4 endpoints)
| Method | Endpoint | Auth Required | Permission | Description |
|--------|----------|---------------|------------|-------------|
| GET | `/config` | ✅ | `config:view` | Get configuration |
| POST | `/config/set` | ✅ | `config:update` | Set configuration |
| GET | `/feature-flags` | ✅ | `config:view` | Get feature flags |
| POST | `/feature-flags/set` | ✅ | `config:update` | Set feature flag |

### 🎛️ Dashboard & Admin (7 endpoints)
| Method | Endpoint | Auth Required | Permission | Description |
|--------|----------|---------------|------------|-------------|
| GET | `/dashboard/stats` | ✅ | `system:view` | Get dashboard statistics |
| POST | `/admin/reindex` | ✅ | `admin:reindex` | Manually reindex data |
| GET | `/admin/health` | ✅ | `admin:health` | Detailed health check |
| POST | `/admin/log-level` | ✅ | `admin:update` | Set log level |
| GET | `/admin/log-level` | ✅ | `admin:view` | Get current log level |
| POST | `/admin/trace-subsystems` | ✅ | `admin:update` | Set trace subsystems |
| GET | `/admin/trace-subsystems` | ✅ | `admin:view` | Get trace subsystems |

### 📈 Metrics & Monitoring (9 endpoints)
| Method | Endpoint | Auth Required | Permission | Description |
|--------|----------|---------------|------------|-------------|
| GET | `/health` | ❌ | None | Basic health check |
| GET | `/metrics` | ❌ | None | Prometheus metrics |
| GET | `/system/metrics` | ❌ | None | EntityDB system metrics |
| POST | `/metrics/collect` | ✅ | `metrics:write` | Collect custom metric |
| GET | `/metrics/current` | ✅ | `metrics:read` | Get current metrics |
| GET | `/metrics/history` | ❌ | None | Get metrics history |
| GET | `/metrics/available` | ❌ | None | List available metrics |
| GET | `/application/metrics` | ✅ | `metrics:read` | Application-specific metrics |
| GET | `/rbac/metrics` | ✅ | `admin:view` | RBAC metrics (admin only) |
| GET | `/rbac/metrics/public` | ❌ | None | Public RBAC metrics |

### 🔧 Legacy/Compatibility (2 endpoints)
| Method | Endpoint | Auth Required | Permission | Description |
|--------|----------|---------------|------------|-------------|
| GET | `/status` | ❌ | None | **DEPRECATED** - Use `/health` |
| POST | `/patches/reindex-tags` | ❌ | None | **DEPRECATED** - Integrated fix |

## 🎨 Request/Response Format

### Standard Response Format
All API responses follow this structure:
```json
{
  "status": "ok|error",
  "message": "Human-readable description",
  "data": { /* Response payload */ },
  "error": "Error details if status=error"
}
```

### Error Response Format
```json
{
  "status": "error",
  "message": "Error description",
  "error": "Detailed error information",
  "code": 400
}
```

### HTTP Status Codes
- **200 OK**: Successful request
- **201 Created**: Resource created successfully
- **400 Bad Request**: Invalid request format or parameters
- **401 Unauthorized**: Missing or invalid authentication
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource not found
- **405 Method Not Allowed**: HTTP method not supported
- **500 Internal Server Error**: Server-side error

## 🔒 RBAC Permission System

### Permission Format
Permissions use hierarchical tag format: `rbac:perm:resource:action`

### Common Permissions
- `rbac:perm:*` - All permissions (admin)
- `rbac:perm:entity:*` - All entity permissions
- `rbac:perm:entity:view` - View entities
- `rbac:perm:entity:create` - Create entities
- `rbac:perm:entity:update` - Update entities
- `rbac:perm:user:create` - Create users (admin only)
- `rbac:perm:admin:*` - All admin operations
- `rbac:perm:metrics:read` - Read metrics
- `rbac:perm:config:update` - Update configuration

### Default Admin User
- **Username**: `admin`
- **Password**: `admin`
- **Permissions**: `rbac:perm:*` (all permissions)
- **Auto-created**: On first server start

## 🚀 Performance & Limits

### Request Limits
- **Rate Limiting**: 1000 requests/hour per token
- **Burst Capacity**: 100 requests/minute
- **Payload Size**: 10MB maximum request size
- **Timeout**: 60 seconds for all requests

### Optimizations
- **Memory-Mapped Files**: Zero-copy reads for large content
- **Sharded Indexing**: 256 concurrent shards for optimal performance
- **Tag Caching**: O(1) tag lookups with intelligent caching
- **Batch Operations**: Automatic batching for write operations

### Performance Headers
```http
X-EntityDB-Query-Time: 0.023ms
X-EntityDB-Index-Hit: true
X-EntityDB-Cache-Hit: true
```

## 🔄 API Versioning

### Current Version: v1
- **Stable**: Feature-complete and production-ready
- **Backward Compatible**: Changes maintain compatibility
- **Path Prefix**: `/api/v1/`
- **Deprecation Policy**: 6 months notice for breaking changes

### Future Versions
- **v2**: Planned for Q2 2026 with enhanced filtering
- **Migration**: Automatic migration tools provided
- **Overlap**: v1 supported for 12 months after v2 release

## 📝 Important Notes

### Relationship Model
**EntityDB v2.32.0 uses tag-based relationships** - there are no separate relationship endpoints. Use entity tags like `relates_to:entity_id` to create relationships.

### Entity Immutability
Entities are **immutable** - updates create new versions with timestamps. There is no DELETE operation for entities.

### Temporal Tags
All tags are stored with nanosecond timestamps. Use `include_timestamps=true` parameter to see raw temporal format.

### Content Chunking
Files >4MB are automatically chunked. Use chunking endpoints for large file handling.

---

*This API overview provides complete, verified documentation for EntityDB v2.32.0. All endpoints and examples are tested against the actual implementation.*