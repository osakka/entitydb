# RBAC Tag Format Standards

The EntityDB system uses a standardized tag format for Role-Based Access Control (RBAC). All permission-related tags use the `rbac:` prefix to clearly distinguish them from other tag types.

## Permission Tag Format

### Basic Permissions
```
rbac:entity.view
rbac:entity.create
rbac:entity.update
rbac:entity.delete
rbac:issue.view
rbac:issue.create
rbac:issue.update
rbac:issue.assign
rbac:workspace.view
rbac:workspace.create
rbac:workspace.update
rbac:session.view
rbac:session.create
rbac:agent.view
rbac:agent.create
rbac:agent.update
rbac:system.view
```

### Wildcard Permissions
```
rbac:*           # All permissions
rbac:entity.*    # All entity permissions
rbac:issue.*     # All issue permissions
```

### Role Tags
```
rbac:role:admin
rbac:role:user
rbac:role:readonly
```

## User Entity Structure

Users are entities with specific tags that define their permissions:

```json
{
  "id": "entity_user_admin",
  "type": "user",
  "title": "admin",
  "tags": [
    "type:user",
    "username:admin",
    "rbac:role:admin",
    "rbac:*"
  ]
}
```

## Examples

### Admin User
```json
{
  "tags": [
    "type:user",
    "username:admin",
    "rbac:role:admin",
    "rbac:*"
  ]
}
```

### Regular User
```json
{
  "tags": [
    "type:user",
    "username:regular_user",
    "rbac:role:user",
    "rbac:entity.view",
    "rbac:entity.create",
    "rbac:entity.update"
  ]
}
```

### Read-Only User
```json
{
  "tags": [
    "type:user",
    "username:readonly_user",
    "rbac:role:readonly",
    "rbac:entity.view"
  ]
}
```

## Benefits

1. **Clear Namespace**: The `rbac:` prefix clearly identifies permission-related tags
2. **No Conflicts**: Prevents confusion with other tag types
3. **Easy Filtering**: Can quickly filter for all RBAC tags
4. **Consistent Pattern**: All permissions follow the same format
5. **Extensible**: Easy to add new permissions following the pattern