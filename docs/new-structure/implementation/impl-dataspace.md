# Multi-Dataspace Architecture Implementation Guide

> **Status**: Completed in v2.27.0  
> **Component**: Multi-tenant dataspace system

## Overview

EntityDB's multi-dataspace architecture provides complete tenant isolation through tagged entities, enabling secure multi-tenant applications on a single EntityDB instance.

## Architecture Design

### Core Concepts

**Dataspace**: A logical namespace that provides complete data isolation between tenants.

**Tag Structure**:
```
dataspace:worca                          # Dataspace membership (required)
worca:self:type:task                     # Entity's own properties  
worca:self:status:todo                   # Self namespace
worca:trait:org:TechCorp                 # Inherited context
worca:trait:project:Mobile               # Trait namespace
```

### Entity Model

```json
{
  "id": "entity_id",
  "dataspace": "worca",
  "self": {
    "type": "task", 
    "status": "todo", 
    "title": "Test Task"
  },
  "traits": {
    "org": "TechCorp", 
    "project": "Mobile"
  },
  "content": "Task description"
}
```

## Implementation Components

### 1. Dataspace Middleware

**File**: `src/api/dataspace_middleware.go`

Validates dataspace access on every request:
- Checks user has required dataspace permissions
- Enforces strict dataspace isolation
- Allows global admin override

```go
func DataspaceMiddleware(entityRepo *binary.EntityRepository) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract dataspace from URL or request
            dataspace := extractDataspace(r)
            
            // Validate user access to dataspace
            if !hasDataspaceAccess(r.Context(), dataspace) {
                http.Error(w, "Dataspace access denied", http.StatusForbidden)
                return
            }
            
            // Add dataspace to request context
            ctx := context.WithValue(r.Context(), "dataspace", dataspace)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 2. RBAC Integration

**Permission Format**:
```
rbac:perm:entity:*:dataspace:worca      # Full dataspace access
rbac:perm:dataspace:create              # Can create dataspaces
rbac:perm:dataspace:manage:worca        # Can manage specific dataspace
```

**Permission Hierarchy**:
- `rbac:perm:*` - Global admin (all dataspaces)
- `rbac:perm:entity:*:dataspace:X` - Full access to dataspace X
- `rbac:perm:entity:read:dataspace:X` - Read-only access to dataspace X

### 3. API Endpoints

```
POST   /api/v1/dataspaces/create              # Create new dataspace
GET    /api/v1/dataspaces/list                # List accessible dataspaces
DELETE /api/v1/dataspaces/delete              # Delete dataspace
POST   /api/v1/dataspaces/{dataspace}/entities/create     # Create entity in dataspace
GET    /api/v1/dataspaces/{dataspace}/entities/query      # Query dataspace entities
```

### 4. Query Isolation

All entity queries automatically filter by dataspace:

```go
func (r *EntityRepository) ListByDataspace(dataspace string) ([]*Entity, error) {
    return r.ListByTag(fmt.Sprintf("dataspace:%s", dataspace))
}
```

## Security Model

### Strict Isolation

- Users can only access dataspaces with explicit permissions
- No cross-dataspace data leakage
- Global admin can override for management

### Permission Evaluation

```go
func hasDataspaceAccess(ctx context.Context, dataspace string) bool {
    user := getUserFromContext(ctx)
    
    // Check global admin
    if user.HasPermission("rbac:perm:*") {
        return true
    }
    
    // Check dataspace-specific permissions
    return user.HasPermission(fmt.Sprintf("rbac:perm:entity:*:dataspace:%s", dataspace))
}
```

## Self/Trait Namespaces

### Self Namespace

Represents the entity's own properties:
- `worca:self:type:task` - What type of entity this is
- `worca:self:assignee:john` - Who owns this entity
- `worca:self:status:completed` - Current state

### Trait Namespace

Represents inherited or contextual properties:
- `worca:trait:org:TechCorp` - Organization context
- `worca:trait:project:MobileApp` - Project context
- `worca:trait:department:Engineering` - Department context

### API Usage

```bash
# Create entity with self and trait properties
curl -X POST https://localhost:8085/api/v1/dataspaces/worca/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "self": {"type": "task", "assignee": "john"},
    "traits": {"org": "TechCorp", "project": "Mobile"},
    "content": "Task description"
  }'

# Query by self properties
curl "https://localhost:8085/api/v1/dataspaces/worca/entities/query?self=type:task"

# Query by trait properties  
curl "https://localhost:8085/api/v1/dataspaces/worca/entities/query?traits=org:TechCorp"
```

## Session Management

### Dataspace-Scoped Sessions

Sessions can be scoped to specific dataspaces:

```json
{
  "type": "session",
  "dataspace": "worca",
  "status": "active",
  "expires_at": "2025-06-08T12:00:00Z",
  "user_id": "user_123"
}
```

### Multi-Dataspace Users

Users with multiple dataspace access:
- `rbac:perm:entity:*:dataspace:worca`
- `rbac:perm:entity:*:dataspace:acme`

## Implementation Status

### ✅ Completed Features

- **Core dataspace infrastructure** - Middleware and validation
- **RBAC integration** - Permission-based dataspace access
- **API endpoints** - Complete dataspace management API
- **Self/trait namespaces** - Structured entity properties
- **Query isolation** - Automatic dataspace filtering
- **Session scoping** - Dataspace-aware authentication

### 🔧 Configuration

```bash
# Enable dataspace isolation (default: true)
ENTITYDB_DATASPACE_ISOLATION=true

# Default dataspace for legacy entities
ENTITYDB_DEFAULT_DATASPACE=default

# Strict mode (reject entities without dataspace tag)
ENTITYDB_DATASPACE_STRICT_MODE=true
```

## Migration Guide

### From Single-Tenant to Multi-Tenant

1. **Assign default dataspace** to existing entities:
```bash
# Tag all existing entities with default dataspace
curl -X POST https://localhost:8085/api/v1/admin/migrate-dataspace \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"default_dataspace": "main"}'
```

2. **Update user permissions**:
```bash
# Grant dataspace access to existing users
curl -X POST https://localhost:8085/api/v1/users/update \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "user_id": "existing_user",
    "add_tags": ["rbac:perm:entity:*:dataspace:main"]
  }'
```

3. **Update application code** to use dataspace-aware APIs.

## Testing

### Unit Tests

```bash
cd src && go test ./... -run TestDataspace
```

### Integration Tests

```bash
./tests/test_dataspace.sh
./tests/test_dataspace_isolation.sh
```

## Performance Considerations

- **Index optimization**: Dataspace tags are indexed for fast filtering
- **Memory usage**: Minimal overhead (~1% for dataspace tags)
- **Query performance**: Dataspace filtering adds <1ms to queries

## Troubleshooting

### Common Issues

1. **Permission denied**: Verify user has dataspace permissions
2. **Cross-dataspace access**: Check for global admin permissions
3. **Legacy entities**: Migrate entities to include dataspace tags

### Debug Commands

```bash
# Check user dataspace permissions
curl "https://localhost:8085/api/v1/rbac/user-permissions?user_id=USER_ID"

# List user's accessible dataspaces
curl "https://localhost:8085/api/v1/dataspaces/list" \
  -H "Authorization: Bearer $TOKEN"
```

---

**Next Steps**: See [Session Management Implementation](impl-sessions.md) for dataspace-scoped authentication details.