# API Documentation Audit Report

## Summary
This report compares the actual API endpoints implemented in the EntityDB codebase with the documentation in `docs/api/entities.md`.

## Implemented Endpoints (from src/main.go)

### Authentication Endpoints
- `POST /api/v1/auth/login` ✓ (Documented)
- `POST /api/v1/auth/logout` ✓ (Documented)
- `GET /api/v1/auth/status` ✓ (Documented)
- `GET /api/v1/auth/whoami` ✗ (Not documented)
- `POST /api/v1/auth/refresh` ✗ (Not documented)

### Entity Endpoints
- `GET /api/v1/entities` ✗ (Not documented - legacy endpoint)
- `POST /api/v1/entities` ✗ (Documented differently)
- `GET /api/v1/entities/list` ✓ (Documented)
- `GET /api/v1/entities/query` ✗ (Not documented)
- `GET /api/v1/entities/get` ✓ (Documented)
- `POST /api/v1/entities/create` ✓ (Documented as POST /entities)
- `PUT /api/v1/entities/update` ✓ (Documented)
- `GET /api/v1/entities/stream` ✗ (Not documented)
- `GET /api/v1/entities/download` ✗ (Not documented)

### Temporal Endpoints
- `GET /api/v1/entities/as-of` ✗ (Not documented)
- `GET /api/v1/entities/history` ✗ (Not documented)
- `GET /api/v1/entities/changes` ✗ (Not documented)
- `GET /api/v1/entities/diff` ✗ (Not documented)
- `GET /api/v1/entities/as-of-fixed` ✗ (Not documented - backward compatibility)
- `GET /api/v1/entities/history-fixed` ✗ (Not documented - backward compatibility)
- `GET /api/v1/entities/changes-fixed` ✗ (Not documented - backward compatibility)
- `GET /api/v1/entities/diff-fixed` ✗ (Not documented - backward compatibility)

### Dataspace Endpoints (Not documented at all)
- `POST /api/v1/dataspaces/entities/create`
- `GET /api/v1/dataspaces/entities/query`
- `GET /api/v1/dataspaces`
- `POST /api/v1/dataspaces`
- `GET /api/v1/dataspaces/{id}`
- `PUT /api/v1/dataspaces/{id}`
- `DELETE /api/v1/dataspaces/{id}`

### Entity Relationship Endpoints
- `GET /api/v1/entity-relationships` ✓ (Partially documented)
- `POST /api/v1/entity-relationships` ✓ (Documented)

### User Management Endpoints
- `POST /api/v1/users/create` ✗ (Not documented)
- `POST /api/v1/users/change-password` ✗ (Not documented)
- `POST /api/v1/users/reset-password` ✗ (Not documented)

### Dashboard/Config Endpoints (Not documented)
- `GET /api/v1/dashboard/stats`
- `GET /api/v1/config`
- `POST /api/v1/config/set`
- `GET /api/v1/feature-flags`
- `POST /api/v1/feature-flags/set`

### Admin Endpoints (Not documented)
- `POST /api/v1/admin/reindex`
- `GET /api/v1/admin/health`

### Metrics Endpoints (Not documented)
- `GET /health`
- `GET /metrics`
- `POST /api/v1/metrics/collect`
- `GET /api/v1/metrics/current`
- `GET /api/v1/metrics/history`
- `GET /api/v1/metrics/available`
- `GET /api/v1/worca/metrics`
- `GET /api/v1/system/metrics`
- `GET /api/v1/rbac/metrics/public`
- `GET /api/v1/rbac/metrics`
- `GET /api/v1/integrity/metrics`

### Other Endpoints (Not documented)
- `GET /api/v1/spec` (Swagger spec)
- `POST /api/v1/patches/reindex-tags`
- `GET /api/v1/patches/status`
- `GET /api/v1/status`
- `GET /debug/ping`

### Test Endpoints (Not documented - intentionally)
- Various `/api/v1/test/*` endpoints

## Key Discrepancies

### 1. Missing Documentation for Major Features
- **Temporal API**: All temporal endpoints (as-of, history, changes, diff) are implemented but not documented
- **Dataspace Management**: Complete dataspace API is implemented but not documented
- **Metrics & Monitoring**: Extensive metrics endpoints are not documented
- **User Management**: Password management endpoints are not documented
- **Configuration Management**: Config and feature flag endpoints are not documented

### 2. Incorrect Examples in Documentation
- The login response example shows incorrect token format (`tk_admin_1234567890` instead of actual JWT)
- The login response includes user object with incorrect ID format (`usr_admin` instead of UUID)

### 3. Authentication Differences
- Documentation mentions JWT but doesn't explain the actual token format
- The `/api/v1/auth/whoami` and `/api/v1/auth/refresh` endpoints are not documented
- Session management details are missing

### 4. RBAC Documentation
- The documentation mentions RBAC permissions but doesn't explain:
  - How permissions are enforced
  - What permissions are required for each endpoint
  - How to check user permissions

### 5. Missing Request/Response Examples
- Query parameters for `/api/v1/entities/query` endpoint
- Chunking/streaming endpoints documentation
- Proper error response formats with actual error codes

### 6. Outdated Information
- Documentation refers to old entity structure (type, title, description fields)
- Actual implementation uses unified entity model with tags and content field
- Tag examples don't reflect temporal storage format (TIMESTAMP|tag)

## Recommendations

1. **Update Core API Documentation**
   - Document all temporal endpoints with examples
   - Add dataspace management API documentation
   - Document user management endpoints
   - Add configuration management endpoints

2. **Add Missing Sections**
   - Metrics and monitoring endpoints
   - Admin operations
   - Chunking and streaming for large files
   - Session management and token refresh

3. **Fix Examples**
   - Update login response to show actual token format
   - Show real entity structure with content field
   - Include temporal tag format in examples

4. **Add API Categories**
   - Group endpoints by functionality
   - Mark which endpoints require authentication
   - Specify required RBAC permissions for each endpoint

5. **Document Query Parameters**
   - Add complete query parameter documentation for list/query endpoints
   - Document pagination parameters (if supported)
   - Document filtering and sorting options