# Tag-Based RBAC Implementation Guide

## Quick Start

### Running the Tag-Based Auth Server

```bash
# Start the tag-based auth server
/opt/entitydb/bin/entitydbd-tag.sh start

# Stop the server
/opt/entitydb/bin/entitydbd-tag.sh stop

# Check status
/opt/entitydb/bin/entitydbd-tag.sh status

# Restart the server
/opt/entitydb/bin/entitydbd-tag.sh restart
```

### Default Login Credentials

After a fresh start, these users are available:

| Username | Password | Permissions |
|----------|----------|-------------|
| admin | admin | Full access (`permission:*`) |
| osakka | osakka | Create, read, update all |
| regular_user | password123 | Read all, create entities |
| readonly_user | readonly123 | Read only |

### Example API Calls

**1. Login as Admin**
```bash
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}'
```

**2. Use the Token**
```bash
# Store the token from login response
TOKEN="your_jwt_token_here"

# List all entities
curl http://localhost:8085/api/v1/entities/list \
  -H "Authorization: Bearer $TOKEN"

# Create a new entity
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "type": "issue",
    "title": "New Issue",
    "description": "This is a test issue",
    "tags": ["type:issue", "priority:high", "status:open"]
  }'
```

## Implementation Details

### Server Architecture

The tag-based RBAC server (`server_tag_simple.go`) includes:

1. **In-Memory Storage**
   - Entities stored in memory (not persistent across restarts)
   - Fast lookups and queries

2. **JWT Authentication**
   - Tokens generated on login
   - Tokens include user permissions
   - Configurable expiration time

3. **Middleware Stack**
   - Authentication middleware validates tokens
   - Permission checking on each request
   - Automatic user context injection

### Permission System

Permissions follow this format: `permission:action:resource`

Common permissions:
- `permission:*` - All permissions
- `permission:create:*` - Create any resource
- `permission:read:*` - Read any resource
- `permission:update:*` - Update any resource
- `permission:delete:*` - Delete any resource
- `permission:read:entity` - Read entities only
- `permission:create:issue` - Create issues only

### Web UI Integration

The web interface (`/opt/entitydb/share/htdocs/`) works with tag-based auth:

1. **Login Flow**
   - Login form sends credentials to `/api/v1/auth/login`
   - Stores JWT token in localStorage
   - Includes token in all API requests

2. **Entity Management**
   - Double-click to edit entities
   - Tags displayed for each entity
   - Type shown as a tag (e.g., `type:issue`)

3. **Permission-Based UI**
   - UI elements enabled/disabled based on permissions
   - Error messages shown for unauthorized actions

## Extending the System

### Adding New Permissions

1. Define the permission pattern:
   ```
   permission:action:resource
   ```

2. Add to user entity tags:
   ```json
   {
     "tags": [
       "type:user",
       "username:newuser",
       "permission:custom:action"
     ]
   }
   ```

3. Check in middleware:
   ```go
   if !CheckTagPermission(permissions, "custom:action") {
       RespondError(w, "Forbidden", http.StatusForbidden)
       return
   }
   ```

### Creating Custom Roles

Roles are just collections of permissions as tags:

```json
// Developer role
{
  "tags": [
    "type:user",
    "username:developer",
    "permission:read:*",
    "permission:create:issue",
    "permission:update:issue",
    "permission:create:comment",
    "role:developer"
  ]
}

// Manager role
{
  "tags": [
    "type:user",
    "username:manager",
    "permission:read:*",
    "permission:create:*",
    "permission:update:*",
    "permission:assign:issue",
    "role:manager"
  ]
}
```

### Integration with Main Server

To integrate tag-based auth with the main EntityDB server:

1. Replace authentication handlers with tag-based versions
2. Update middleware to use tag permission checking
3. Ensure user entities are created with proper permission tags
4. Update JWT token generation to include permissions

## Troubleshooting

### Common Issues

1. **"Unauthorized" Error**
   - Check if token has expired
   - Verify token is included in Authorization header
   - Ensure user has required permissions

2. **Cannot Edit Entities**
   - Verify user has `permission:update:entity`
   - Check if entity exists and is accessible

3. **Login Fails**
   - Verify username exists (check for `username:` tag)
   - Ensure password is correct
   - Check server logs for errors

### Debug Mode

Enable debug logging in the server:
```go
// In server_tag_simple.go
if debug {
    log.Printf("Checking permission: %s for user: %s", permission, userID)
}
```

### Viewing User Permissions

To see a user's permissions:
```bash
# Get user entity
curl http://localhost:8085/api/v1/entities/entity_user_admin \
  -H "Authorization: Bearer $TOKEN" | jq .data.tags
```

This will show all tags including permissions.

## Migration from Traditional RBAC

To migrate from role-based to tag-based RBAC:

1. Map existing roles to permission tags
2. Create user entities with appropriate tags
3. Update authentication handlers
4. Replace permission checks with tag-based checks
5. Update UI to handle JWT tokens

## Security Best Practices

1. **Strong Passwords**
   - Enforce minimum password requirements
   - Hash passwords with bcrypt (cost factor 10+)

2. **Token Management**
   - Short token expiration times
   - Refresh tokens for extended sessions
   - Revoke tokens on logout

3. **Permission Principles**
   - Principle of least privilege
   - Regular permission audits
   - No default admin accounts in production

4. **API Security**
   - HTTPS in production
   - Rate limiting
   - Request validation
   - Audit logging

## Future Enhancements

1. **Enhanced Storage**
   - Advanced entity relationship queries
   - Optimized temporal permission caching

2. **Advanced Permissions**
   - Conditional permissions
   - Time-based permissions
   - Resource-specific permissions

3. **UI Improvements**
   - Permission management interface
   - User creation wizard
   - Role templates

4. **Monitoring**
   - Permission usage analytics
   - Failed authentication tracking
   - Audit trails