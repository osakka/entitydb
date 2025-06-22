# Tag-Based RBAC System Documentation

## Overview

The EntityDB system uses a pure entity-based architecture where everything, including users and permissions, are represented as entities with tags. This document explains how the tag-based Role-Based Access Control (RBAC) system works.

## Core Concepts

### Everything is an Entity

In EntityDB, all objects are entities with tags, including:
- Users
- Workspaces  
- Issues
- Sessions
- Permissions

### Users as Entities

Users are stored as entities with the following structure:

```json
{
  "id": "entity_user_admin",
  "type": "user", 
  "title": "admin",
  "tags": [
    "type:user",
    "username:admin",
    "permission:*",
    "role:admin"
  ],
  "content": [
    {
      "type": "password",
      "value": "hashed_password_here"
    }
  ]
}
```

## Permission Model

### Permission Tags

Permissions are granted through tags that follow this pattern:
- `permission:action:resource`
- `permission:*` (wildcard for all permissions)

Examples:
- `permission:create:entity` - Can create entities
- `permission:read:*` - Can read all resources
- `permission:write:entity` - Can write/update entities
- `permission:delete:issue` - Can delete issues

### Hierarchical Permissions

Permissions support wildcards at different levels:
- `permission:*` - All permissions (super admin)
- `permission:read:*` - Can read everything
- `permission:*:entity` - All actions on entities

## Default Users

When the server starts with a fresh database, it creates four default user entities:

1. **Admin User**
   - ID: `entity_user_admin`
   - Username: admin
   - Password: admin
   - Tags: `type:user`, `username:admin`, `permission:*`, `role:admin`
   - Has all permissions

2. **Osakka User**
   - ID: `entity_user_osakka`
   - Username: osakka
   - Password: osakka
   - Tags: `type:user`, `username:osakka`, `permission:create:*`, `permission:read:*`, `permission:update:*`, `role:creator`
   - Can create, read, and update all resources

3. **Regular User**
   - ID: `entity_user_regular_user`
   - Username: regular_user
   - Password: password123
   - Tags: `type:user`, `username:regular_user`, `permission:read:*`, `permission:create:entity`, `permission:update:entity:self`, `role:user`
   - Can read everything, create entities, and update own entities

4. **Read-Only User**
   - ID: `entity_user_readonly_user`
   - Username: readonly_user
   - Password: readonly123
   - Tags: `type:user`, `username:readonly_user`, `permission:read:*`, `role:viewer`
   - Can only read resources

## Authentication Flow

### Login Process

1. User sends POST request to `/api/v1/auth/login` with username and password
2. Server finds user entity by `username:` tag
3. Server validates password against content item
4. If valid, generates JWT token containing:
   - User ID
   - Username
   - All permission tags
5. Returns token to user

### JWT Token Structure

```json
{
  "user_id": "entity_user_admin",
  "username": "admin",
  "permissions": [
    "permission:*"
  ],
  "exp": 1234567890
}
```

### Authorization Middleware

For each protected API request:
1. Extract JWT token from Authorization header
2. Validate token signature and expiration
3. Extract permissions from token
4. Check if user has required permission for the action
5. Allow/deny access based on permission check

## Permission Checking

The tag-based permission system uses pattern matching:

```go
// CheckTagPermission checks if user has a specific permission
func CheckTagPermission(userPermissions []string, requiredPermission string) bool {
    for _, perm := range userPermissions {
        if strings.HasPrefix(perm, "permission:") {
            // Extract permission part after "permission:"
            permPart := strings.TrimPrefix(perm, "permission:")
            
            // Check for exact match or wildcard match
            if permPart == "*" || permPart == requiredPermission {
                return true
            }
            
            // Check for partial wildcard match (e.g., "read:*")
            parts := strings.Split(permPart, ":")
            reqParts := strings.Split(requiredPermission, ":")
            
            match := true
            for i := 0; i < len(parts) && i < len(reqParts); i++ {
                if parts[i] != "*" && parts[i] != reqParts[i] {
                    match = false
                    break
                }
            }
            
            if match {
                return true
            }
        }
    }
    return false
}
```

## File Structure

The tag-based RBAC system consists of these files:

### `/opt/entitydb/src/api/tag_permissions.go`
Contains the permission checking logic and utilities.

### `/opt/entitydb/src/api/tag_auth_handler.go`
Handles authentication endpoints (login/logout) and JWT generation.

### `/opt/entitydb/src/api/tag_auth_middleware.go`
Middleware that validates JWT tokens and checks permissions for each request.

### `/opt/entitydb/src/server_tag_simple.go`
Complete server implementation with tag-based authentication.

## API Endpoints

### Authentication Endpoints

**POST /api/v1/auth/login**
```json
Request:
{
  "username": "admin",
  "password": "admin"
}

Response:
{
  "token": "jwt_token_here",
  "expires_at": "2024-01-01T00:00:00Z"
}
```

**POST /api/v1/auth/logout**
Simply clears the token on the client side.

### Protected Endpoints

All other endpoints require a valid JWT token with appropriate permissions:
- `GET /api/v1/entities/list` - Requires `permission:read:entity`
- `POST /api/v1/entities/create` - Requires `permission:create:entity`
- `PUT /api/v1/entities/:id` - Requires `permission:update:entity`
- `DELETE /api/v1/entities/:id` - Requires `permission:delete:entity`

## Benefits of Tag-Based RBAC

1. **Flexibility** - Permissions can be composed dynamically by adding/removing tags
2. **Granularity** - Fine-grained permissions at any level
3. **Simplicity** - No complex role hierarchies, just tags
4. **Consistency** - Users are entities just like everything else
5. **Extensibility** - Easy to add new permissions or resources

## Security Considerations

1. Passwords are hashed using bcrypt before storage
2. JWT tokens expire after a configurable duration
3. Tokens are signed with a secret key
4. All API requests require valid authentication
5. Permissions are checked on every request

## Example Usage

### Creating a Custom User

```bash
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "type": "user",
    "title": "custom_user",
    "tags": [
      "type:user",
      "username:custom_user",
      "permission:read:*",
      "permission:create:issue",
      "role:contributor"
    ]
  }'
```

### Checking User Permissions

To see what permissions a user has, simply look at their tags:

```bash
curl http://localhost:8085/api/v1/entities/entity_user_admin \
  -H "Authorization: Bearer $TOKEN" | jq .data.tags
```

This will show all tags including permissions for that user entity.