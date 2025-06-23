# ADR-004: Tag-Based RBAC System

## Status
Accepted (2025-04-15)

## Context
EntityDB needed a permission system that integrated naturally with the entity-tag architecture. Traditional role-based access control (RBAC) systems typically use separate permission tables and role assignments.

### Requirements
- Integrate seamlessly with entity-tag system
- Support hierarchical permissions with wildcards
- Fine-grained access control at entity and operation level
- Maintain high performance for permission checks
- Support role inheritance and permission aggregation

### Constraints
- Must work with temporal tag system
- Performance critical for every API request
- Need to support complex permission hierarchies
- Must be intuitive for administrators to manage

### Alternative Approaches
1. **Separate permission tables**: Traditional RBAC with dedicated permission storage
2. **Embedded permission objects**: Store permissions as JSON in user entities
3. **Tag-based permissions**: Use entity tags for permission storage
4. **External authorization service**: Delegate to external auth service

## Decision
We decided to implement **tag-based RBAC** using the hierarchical tag namespace format:

```
rbac:role:admin
rbac:role:user
rbac:perm:*                    # All permissions
rbac:perm:entity:*             # All entity permissions
rbac:perm:entity:view          # View entities
rbac:perm:entity:create        # Create entities
rbac:perm:entity:update        # Update entities
rbac:perm:user:*               # All user management
rbac:perm:config:view          # View configuration
```

### Implementation Details
- **Permission Hierarchy**: Dot-separated namespaces for logical grouping
- **Wildcard Support**: `*` wildcard for permission inheritance
- **Tag Storage**: Permissions stored as regular entity tags
- **Middleware Enforcement**: RBAC middleware checks permissions on every request
- **Session Integration**: Permissions cached in session for performance

### Permission Categories
- `entity:*` - Entity operations (view, create, update, delete)
- `user:*` - User management operations
- `config:*` - Configuration management
- `relation:*` - Entity relationship operations
- `metrics:*` - Metrics and monitoring access
- `admin:*` - Administrative operations

## Consequences

### Positive
- **Natural Integration**: Permissions are entities with tags like everything else
- **Hierarchical Structure**: Clear permission inheritance with wildcards
- **Performance**: Fast tag-based permission lookups
- **Flexibility**: Easy to add new permission types
- **Temporal Compatibility**: Works seamlessly with temporal tag system
- **Audit Trail**: Permission changes tracked automatically
- **Simple Administration**: Intuitive tag-based permission management

### Negative
- **Tag Namespace Conflicts**: Potential conflicts with user-defined tags
- **Complex Wildcard Logic**: Wildcard resolution requires careful implementation
- **Permission Sprawl**: Easy to create complex permission hierarchies
- **Migration Complexity**: Existing permissions need tag conversion

### Security Implications
- **Principle of Least Privilege**: Default deny with explicit permission grants
- **Permission Caching**: Session-based caching reduces database load
- **Validation**: Comprehensive permission validation on all endpoints
- **Audit**: All permission changes logged through temporal system

## Implementation History
- v2.5.0: Initial tag-based RBAC implementation with hierarchical namespaces
- v2.8.0: Integration with temporal tag system
- v2.18.0: Enhanced permission middleware with detailed logging
- v2.29.0: Integration with embedded credential system
- v2.32.0: Unified with sharded indexing for improved performance

## Examples

### User with Admin Role
```json
{
  "id": "user-admin-001",
  "tags": [
    "type:user",
    "rbac:role:admin",
    "rbac:perm:*"
  ]
}
```

### User with Limited Permissions
```json
{
  "id": "user-viewer-001", 
  "tags": [
    "type:user",
    "rbac:role:viewer", 
    "rbac:perm:entity:view",
    "rbac:perm:metrics:read"
  ]
}
```

### Permission Check Logic
```go
func HasPermission(user *Entity, permission string) bool {
    // Check for wildcard permissions first
    if user.HasTag("rbac:perm:*") {
        return true
    }
    
    // Check for specific permission
    if user.HasTag("rbac:perm:" + permission) {
        return true
    }
    
    // Check for namespace wildcards
    parts := strings.Split(permission, ":")
    for i := len(parts) - 1; i > 0; i-- {
        wildcard := strings.Join(parts[:i], ":") + ":*"
        if user.HasTag("rbac:perm:" + wildcard) {
            return true
        }
    }
    
    return false
}
```

## Related Decisions
- [ADR-001: Temporal Tag Storage](./001-temporal-tag-storage.md) - Foundation for tag-based permissions
- [ADR-006: Credential Storage in Entities](./006-credential-storage-in-entities.md) - User authentication integration