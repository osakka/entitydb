# Tag Namespace Specification

The EntityDB system uses a hierarchical tag namespace structure for all tags. This provides clear organization, prevents conflicts, and allows for complex filtering and querying.

## Tag Format

```
namespace:category:subcategory:...:value
```

## Core Namespaces

### 1. Type Namespace
Defines the entity type:
```
type:user
type:issue
type:workspace
type:agent
type:session
```

### 2. Identity Namespace
User and system identifiers:
```
id:username:osakka
id:email:user@example.com
id:system:entitydb-core
```

### 3. RBAC Namespace
Role-based access control permissions:
```
rbac:role:admin
rbac:role:user
rbac:role:readonly

rbac:perm:entity:create
rbac:perm:entity:read
rbac:perm:entity:update
rbac:perm:entity:delete

rbac:perm:issue:create
rbac:perm:issue:assign
rbac:perm:issue:close

rbac:perm:workspace:create
rbac:perm:workspace:manage

rbac:perm:system:configure
rbac:perm:system:monitor

rbac:perm:*  # Wildcard for all permissions
rbac:perm:entity:*  # All entity permissions
rbac:perm:issue:*   # All issue permissions
```

### 4. Status Namespace
Entity states and lifecycle:
```
status:active
status:inactive
status:pending
status:archived
status:deleted

status:issue:open
status:issue:in_progress
status:issue:blocked
status:issue:resolved
status:issue:closed
```

### 5. Meta Namespace
Metadata and system information:
```
meta:created:2024-01-01T00:00:00Z
meta:updated:2024-01-02T00:00:00Z
meta:version:1.0.0
meta:source:api
meta:source:import
```

### 6. Relationship Namespace
Entity relationships and hierarchies:
```
rel:parent:entity_123
rel:child:entity_456
rel:depends:entity_789
rel:blocks:entity_012
rel:assigned:user_345
```

### 7. Workspace Namespace
Workspace-specific tags:
```
ws:name:production
ws:env:prod
ws:region:us-east-1
ws:team:backend
```

### 8. Priority Namespace
Priority and importance levels:
```
priority:critical
priority:high
priority:medium
priority:low
priority:trivial
```

### 9. Label Namespace
User-defined labels and categories:
```
label:bug
label:feature
label:documentation
label:tech-debt
label:ui-improvement
```

### 10. Audit Namespace
Audit and compliance tags:
```
audit:action:create
audit:action:update
audit:action:delete
audit:user:admin
audit:timestamp:2024-01-01T00:00:00Z
audit:ip:192.168.1.1
```

## Benefits of Hierarchical Namespaces

1. **Clear Organization**: Each namespace has a specific purpose
2. **No Conflicts**: Different namespaces can use the same values
3. **Flexible Querying**: Can filter by namespace, category, or specific values
4. **Extensible**: Easy to add new namespaces or subcategories
5. **Self-Documenting**: Tag structure explains its purpose
6. **Wildcard Support**: Can use wildcards at any level

## Examples

### Admin User Entity
```json
{
  "id": "entity_user_admin",
  "tags": [
    "type:user",
    "id:username:admin",
    "id:email:admin@entitydb.system",
    "rbac:role:admin",
    "rbac:perm:*",
    "status:active",
    "meta:created:2024-01-01T00:00:00Z"
  ]
}
```

### Issue Entity
```json
{
  "id": "entity_issue_123",
  "tags": [
    "type:issue",
    "status:issue:in_progress",
    "priority:high",
    "label:bug",
    "ws:name:production",
    "rel:parent:entity_epic_456",
    "rel:assigned:entity_user_789",
    "rbac:perm:issue:view",
    "meta:created:2024-01-15T10:30:00Z"
  ]
}
```

### Querying Examples

```javascript
// Find all admin users
entities.filter(e => e.tags.includes("rbac:role:admin"))

// Find all high-priority bugs in production
entities.filter(e => 
  e.tags.includes("type:issue") &&
  e.tags.includes("priority:high") &&
  e.tags.includes("label:bug") &&
  e.tags.includes("ws:name:production")
)

// Find all entities with any entity permissions
entities.filter(e => 
  e.tags.some(tag => tag.startsWith("rbac:perm:entity:"))
)

// Find all inactive users
entities.filter(e => 
  e.tags.includes("type:user") &&
  e.tags.includes("status:inactive")
)
```

## Migration Strategy

To migrate existing tags to the new namespace structure:

1. `user` → `type:user`
2. `admin` → `rbac:role:admin`
3. `entity.create` → `rbac:perm:entity:create`
4. `username:osakka` → `id:username:osakka`
5. `high-priority` → `priority:high`
6. `bug` → `label:bug`