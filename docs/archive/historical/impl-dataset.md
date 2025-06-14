# Multi-Dataset Architecture Implementation Guide

> **Status**: Completed in v2.27.0  
> **Component**: Multi-tenant dataset system

## Overview

EntityDB's multi-dataset architecture provides complete tenant isolation through tagged entities, enabling secure multi-tenant applications on a single EntityDB instance.

## Architecture Design

### Core Concepts

**Dataset**: A logical namespace that provides complete data isolation between tenants.

**Tag Structure**:
```
dataset:worca                          # Dataset membership (required)
worca:self:type:task                     # Entity's own properties  
worca:self:status:todo                   # Self namespace
worca:trait:org:TechCorp                 # Inherited context
worca:trait:project:Mobile               # Trait namespace
```

### Entity Model

```json
{
  "id": "entity_id",
  "dataset": "worca",
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

### 1. Dataset Middleware

**File**: `src/api/dataset_middleware.go`

Validates dataset access on every request:
- Checks user has required dataset permissions
- Enforces strict dataset isolation
- Allows global admin override

```go
func DatasetMiddleware(entityRepo *binary.EntityRepository) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract dataset from URL or request
            dataset := extractDataset(r)
            
            // Validate user access to dataset
            if !hasDatasetAccess(r.Context(), dataset) {
                http.Error(w, "Dataset access denied", http.StatusForbidden)
                return
            }
            
            // Add dataset to request context
            ctx := context.WithValue(r.Context(), "dataset", dataset)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 2. RBAC Integration

**Permission Format**:
```
rbac:perm:entity:*:dataset:worca      # Full dataset access
rbac:perm:dataset:create              # Can create datasets
rbac:perm:dataset:manage:worca        # Can manage specific dataset
```

**Permission Hierarchy**:
- `rbac:perm:*` - Global admin (all datasets)
- `rbac:perm:entity:*:dataset:X` - Full access to dataset X
- `rbac:perm:entity:read:dataset:X` - Read-only access to dataset X

### 3. API Endpoints

```
POST   /api/v1/datasets/create              # Create new dataset
GET    /api/v1/datasets/list                # List accessible datasets
DELETE /api/v1/datasets/delete              # Delete dataset
POST   /api/v1/datasets/{dataset}/entities/create     # Create entity in dataset
GET    /api/v1/datasets/{dataset}/entities/query      # Query dataset entities
```

### 4. Query Isolation

All entity queries automatically filter by dataset:

```go
func (r *EntityRepository) ListByDataset(dataset string) ([]*Entity, error) {
    return r.ListByTag(fmt.Sprintf("dataset:%s", dataset))
}
```

## Security Model

### Strict Isolation

- Users can only access datasets with explicit permissions
- No cross-dataset data leakage
- Global admin can override for management

### Permission Evaluation

```go
func hasDatasetAccess(ctx context.Context, dataset string) bool {
    user := getUserFromContext(ctx)
    
    // Check global admin
    if user.HasPermission("rbac:perm:*") {
        return true
    }
    
    // Check dataset-specific permissions
    return user.HasPermission(fmt.Sprintf("rbac:perm:entity:*:dataset:%s", dataset))
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
curl -X POST https://localhost:8085/api/v1/datasets/worca/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "self": {"type": "task", "assignee": "john"},
    "traits": {"org": "TechCorp", "project": "Mobile"},
    "content": "Task description"
  }'

# Query by self properties
curl "https://localhost:8085/api/v1/datasets/worca/entities/query?self=type:task"

# Query by trait properties  
curl "https://localhost:8085/api/v1/datasets/worca/entities/query?traits=org:TechCorp"
```

## Session Management

### Dataset-Scoped Sessions

Sessions can be scoped to specific datasets:

```json
{
  "type": "session",
  "dataset": "worca",
  "status": "active",
  "expires_at": "2025-06-08T12:00:00Z",
  "user_id": "user_123"
}
```

### Multi-Dataset Users

Users with multiple dataset access:
- `rbac:perm:entity:*:dataset:worca`
- `rbac:perm:entity:*:dataset:acme`

## Implementation Status

### ✅ Completed Features

- **Core dataset infrastructure** - Middleware and validation
- **RBAC integration** - Permission-based dataset access
- **API endpoints** - Complete dataset management API
- **Self/trait namespaces** - Structured entity properties
- **Query isolation** - Automatic dataset filtering
- **Session scoping** - Dataset-aware authentication

### 🔧 Configuration

```bash
# Enable dataset isolation (default: true)
ENTITYDB_DATASET_ISOLATION=true

# Default dataset for legacy entities
ENTITYDB_DEFAULT_DATASET=default

# Strict mode (reject entities without dataset tag)
ENTITYDB_DATASET_STRICT_MODE=true
```

## Migration Guide

### From Single-Tenant to Multi-Tenant

1. **Assign default dataset** to existing entities:
```bash
# Tag all existing entities with default dataset
curl -X POST https://localhost:8085/api/v1/admin/migrate-dataset \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"default_dataset": "main"}'
```

2. **Update user permissions**:
```bash
# Grant dataset access to existing users
curl -X POST https://localhost:8085/api/v1/users/update \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "user_id": "existing_user",
    "add_tags": ["rbac:perm:entity:*:dataset:main"]
  }'
```

3. **Update application code** to use dataset-aware APIs.

## Testing

### Unit Tests

```bash
cd src && go test ./... -run TestDataset
```

### Integration Tests

```bash
./tests/test_dataset.sh
./tests/test_dataset_isolation.sh
```

## Performance Considerations

- **Index optimization**: Dataset tags are indexed for fast filtering
- **Memory usage**: Minimal overhead (~1% for dataset tags)
- **Query performance**: Dataset filtering adds <1ms to queries

## Troubleshooting

### Common Issues

1. **Permission denied**: Verify user has dataset permissions
2. **Cross-dataset access**: Check for global admin permissions
3. **Legacy entities**: Migrate entities to include dataset tags

### Debug Commands

```bash
# Check user dataset permissions
curl "https://localhost:8085/api/v1/rbac/user-permissions?user_id=USER_ID"

# List user's accessible datasets
curl "https://localhost:8085/api/v1/datasets/list" \
  -H "Authorization: Bearer $TOKEN"
```

---

**Next Steps**: See [Session Management Implementation](impl-sessions.md) for dataset-scoped authentication details.