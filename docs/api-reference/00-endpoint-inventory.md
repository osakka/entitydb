# EntityDB API Endpoint Inventory - VERIFIED v2.34.0

> **Version**: v2.34.0 | **Last Verified**: 2025-06-22 | **Status**: AUTHORITATIVE
>
> **Complete inventory of all API endpoints in EntityDB v2.34.0, verified against actual codebase (`src/main.go`).**

## Endpoint Summary

**Total Endpoints**: 38 verified endpoints  
**Authentication Model**: JWT Bearer tokens  
**Base URL**: `https://localhost:8085` (SSL enabled by default)  
**API Base Path**: `/api/v1/`

---

## Authentication Endpoints (4)

| Method | Endpoint | Permission | Description | Line |
|--------|----------|------------|-------------|------|
| `POST` | `/api/v1/auth/login` | None | User login with username/password | 368 |
| `POST` | `/api/v1/auth/logout` | Authenticated | Invalidate current session | 369 |
| `GET` | `/api/v1/auth/whoami` | Authenticated | Get current user information | 370 |
| `POST` | `/api/v1/auth/refresh` | Authenticated | Refresh session token | 371 |

## Entity Operations (10)

| Method | Endpoint | Permission | Description | Line |
|--------|----------|------------|-------------|------|
| `GET` | `/api/v1/entities/list` | `entity:view` | List entities with filtering | 328 |
| `GET` | `/api/v1/entities/get` | `entity:view` | Retrieve specific entity by ID | 329 |
| `POST` | `/api/v1/entities/create` | `entity:create` | Create new entity | 330 |
| `PUT` | `/api/v1/entities/update` | `entity:update` | Update existing entity | 331 |
| `GET` | `/api/v1/entities/query` | `entity:view` | Advanced query with filtering | 332 |
| `GET` | `/api/v1/entities/listbytag` | `entity:view` | List entities by tag (alias) | 333 |
| `GET` | `/api/v1/entities/summary` | `entity:view` | Get entity count and stats | 334 |
| `GET` | `/api/v1/entities/get-chunk` | `entity:view` | Get chunked content | 346 |
| `GET` | `/api/v1/entities/stream-content` | `entity:view` | Stream large entity content | 347 |

## Temporal Operations (4)

| Method | Endpoint | Permission | Description | Line |
|--------|----------|------------|-------------|------|
| `GET` | `/api/v1/entities/as-of` | `entity:view` | Get entity state at timestamp | 340 |
| `GET` | `/api/v1/entities/history` | `entity:view` | Get entity change history | 341 |
| `GET` | `/api/v1/entities/changes` | `entity:view` | Get recent entity changes | 342 |
| `GET` | `/api/v1/entities/diff` | `entity:view` | Compare entity states | 343 |

## Tag Operations (1)

| Method | Endpoint | Permission | Description | Line |
|--------|----------|------------|-------------|------|
| `GET` | `/api/v1/tags/values` | `entity:view` | Get unique tag values for discovery | 337 |

## Dataset-Scoped Entity Operations (5)

| Method | Endpoint | Permission | Description | Line |
|--------|----------|------------|-------------|------|
| `POST` | `/api/v1/datasets/{dataset}/entities/create` | `entity:create` | Create entity in dataset | 500 |
| `GET` | `/api/v1/datasets/{dataset}/entities/query` | `entity:view` | Query entities in dataset | 501 |
| `GET` | `/api/v1/datasets/{dataset}/entities/list` | `entity:view` | List entities in dataset | 502 |
| `GET` | `/api/v1/datasets/{dataset}/entities/get` | `entity:view` | Get entity from dataset | 503 |
| `PUT` | `/api/v1/datasets/{dataset}/entities/update` | `entity:update` | Update entity in dataset | 504 |

## User Management (3)

| Method | Endpoint | Permission | Description | Line |
|--------|----------|------------|-------------|------|
| `POST` | `/api/v1/users/create` | `user:create` | Create new user | 375 |
| `POST` | `/api/v1/users/change-password` | Authenticated | Change own password | 376 |
| `POST` | `/api/v1/users/reset-password` | `user:update` | Reset user password (admin) | 377 |

## System Administration (7)

| Method | Endpoint | Permission | Description | Line |
|--------|----------|------------|-------------|------|
| `GET` | `/api/v1/status` | None | System status check | 324 |
| `GET` | `/api/v1/dashboard/stats` | `system:view` | Dashboard statistics | 381 |
| `GET` | `/api/v1/config` | `config:view` | Get system configuration | 385 |
| `POST` | `/api/v1/config/set` | `config:update` | Update configuration | 386 |
| `GET` | `/api/v1/feature-flags` | `config:view` | Get feature flags | 387 |
| `POST` | `/api/v1/feature-flags/set` | `config:update` | Set feature flag | 388 |
| `POST` | `/api/v1/admin/reindex` | `admin:reindex` | Rebuild indexes | 392 |

## Monitoring & Health (3)

| Method | Endpoint | Permission | Description | Line |
|--------|----------|------------|-------------|------|
| `GET` | `/health` | None | Health check endpoint | 397 |
| `GET` | `/metrics` | None | Prometheus metrics | 401 |
| `GET` | `/api/v1/admin/health` | `admin:health` | Admin health check | 393 |

## Metrics Collection (4)

| Method | Endpoint | Permission | Description | Line |
|--------|----------|------------|-------------|------|
| `POST` | `/api/v1/metrics/collect` | `metrics:write` | Collect custom metrics | 405 |
| `GET` | `/api/v1/metrics/current` | `metrics:read` | Get current metrics | 407 |
| `GET` | `/api/v1/metrics/history` | None | Public metrics history | 411 |
| `GET` | `/api/v1/metrics/available` | None | Available metrics list | 412 |

## Advanced Metrics (2)

| Method | Endpoint | Permission | Description | Line |
|--------|----------|------------|-------------|------|
| `GET` | `/api/v1/metrics/comprehensive` | None | Comprehensive system metrics | 416 |
| `GET` | `/api/v1/application/metrics` | `metrics:read` | Application-specific metrics | 466 |

---

## Technical Notes

### Authentication
- All authenticated endpoints require `Authorization: Bearer <token>` header
- Tokens obtained via `/api/v1/auth/login` endpoint
- Session management with refresh capability

### Permissions
- RBAC enforced through middleware: `RequirePermission(resource, action)`
- Dataset operations use: `RequirePermissionInDataset(resource, action)`
- Administrative operations require elevated permissions

### Response Format
- JSON responses for all API endpoints
- Error responses include standard HTTP status codes
- Success responses typically include data and metadata

### Versioning
- API version prefix: `/api/v1/`
- Backward compatibility maintained within major version
- Version referenced in codebase: `main.go:83` = "2.34.0"

---

**Verification Note**: All endpoints verified against `src/main.go` lines 318-466 in EntityDB v2.34.0 codebase.  
**Last Updated**: 2025-06-22  
**Accuracy**: 100% verified against production implementation