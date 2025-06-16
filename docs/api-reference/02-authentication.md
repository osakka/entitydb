# EntityDB Authentication API

> **Version**: v2.32.20 | **Last Updated**: 2025-06-13 | **Status**: AUTHORITATIVE

## Overview

EntityDB provides a secure authentication system using JWT session tokens with embedded credential storage. As of v2.29.0, user credentials are stored directly in the user entity's content field as `salt|bcrypt_hash`, eliminating the need for separate credential entities.

> **⚠️ BREAKING CHANGE in v2.29.0**: Authentication architecture has fundamentally changed. User credentials are now embedded in user entities. **NO BACKWARD COMPATIBILITY** - all users must be recreated.

## Authentication Flow

1. **Login**: Submit username/password to receive session token
2. **Authorization**: Include `Authorization: Bearer <token>` header in API requests
3. **Session Management**: Tokens expire automatically and can be refreshed
4. **Logout**: Invalidate token to end session

## API Endpoints

### POST /api/v1/auth/login

Authenticate user with username and password.

**Request**:
```bash
curl -k -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin"
  }'
```

**Response** (200 OK):
```json
{
  "token": "session_12345abcdef...",
  "expires_at": "2025-06-12T23:45:00Z",
  "user_id": "user_admin_12345",
  "user": {
    "id": "user_admin_12345",
    "username": "admin",
    "email": "admin@entitydb.local",
    "roles": ["admin", "user"]
  }
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or missing credentials
- `401 Unauthorized`: Invalid username or password
- `500 Internal Server Error`: Failed to create session

### POST /api/v1/auth/logout

Invalidate current session token.

**Request**:
```bash
curl -k -X POST https://localhost:8085/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

**Response** (200 OK):
```json
{
  "message": "Logged out successfully"
}
```

**Error Responses**:
- `401 Unauthorized`: No token provided or invalid token format
- `500 Internal Server Error`: Failed to invalidate session

### POST /api/v1/auth/refresh

Refresh session token to extend expiration.

**Request**:
```bash
curl -k -X POST https://localhost:8085/api/v1/auth/refresh \
  -H "Authorization: Bearer $TOKEN"
```

**Response** (200 OK):
```json
{
  "token": "new_session_67890xyz...",
  "expires_at": "2025-06-13T12:00:00Z",
  "user_id": "user_admin_12345",
  "user": {
    "id": "user_admin_12345",
    "username": "admin",
    "email": "admin@entitydb.local",
    "roles": ["admin", "user"]
  }
}
```

### GET /api/v1/auth/whoami

Get information about currently authenticated user.

**Request**:
```bash
curl -k -X GET https://localhost:8085/api/v1/auth/whoami \
  -H "Authorization: Bearer $TOKEN"
```

**Response** (200 OK):
```json
{
  "id": "user_admin_12345",
  "username": "admin",
  "email": "admin@entitydb.local",
  "roles": ["admin", "user"]
}
```

## Authorization Header Format

All authenticated API requests must include the Authorization header:

```
Authorization: Bearer <session-token>
```

**Example**:
```bash
curl -k -X GET https://localhost:8085/api/v1/entities/list \
  -H "Authorization: Bearer session_12345abcdef..."
```

## Default Admin User

EntityDB automatically creates a default admin user on first startup:

- **Username**: `admin`
- **Password**: `admin`
- **Roles**: `admin`, `user`

> **⚠️ Security Warning**: Change the default admin password immediately in production environments.

## User Roles and Permissions

Users are assigned roles through RBAC tags:

- `rbac:role:admin` - Full administrative access
- `rbac:role:user` - Standard user access

Permissions are checked via tag-based RBAC:
- `rbac:perm:entity:create` - Create entities
- `rbac:perm:entity:view` - View entities
- `rbac:perm:entity:update` - Update entities
- `rbac:perm:system:admin` - System administration

## Security Features

### Password Security
- Passwords hashed using bcrypt with cost 10
- Salt stored with hash in user entity content field
- No plaintext password storage

### Session Security
- Session tokens generated using crypto/rand
- Configurable session TTL (default: 1 hour)
- Automatic session cleanup for expired tokens
- IP address and user agent tracking

### Authentication Tracking
- Failed login attempts are logged
- Session activity monitoring
- Authentication metrics available via `/api/v1/rbac/metrics`

## Error Handling

All authentication endpoints return structured error responses:

```json
{
  "error": "Detailed error message"
}
```

Common error scenarios:
- **Invalid credentials**: Username/password mismatch
- **Missing token**: Authorization header not provided
- **Expired token**: Session has exceeded TTL
- **Invalid token format**: Malformed Authorization header

## Integration Examples

### Login and Store Token
```bash
# Login and extract token
TOKEN=$(curl -s -k -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Use token for authenticated requests
curl -k -X GET https://localhost:8085/api/v1/entities/list \
  -H "Authorization: Bearer $TOKEN"
```

### Session Management
```bash
# Check current user
curl -k -X GET https://localhost:8085/api/v1/auth/whoami \
  -H "Authorization: Bearer $TOKEN"

# Refresh token
curl -k -X POST https://localhost:8085/api/v1/auth/refresh \
  -H "Authorization: Bearer $TOKEN"

# Logout
curl -k -X POST https://localhost:8085/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

## Related Documentation

- [User Management](../50-admin-guides/01-user-management.md)
- [RBAC System](../20-architecture/03-rbac.md)
- [Security Configuration](../50-admin-guides/02-security.md)
- [API Overview](./01-overview.md)

## Version History

- **v2.30.0**: Current authentication system with embedded credentials
- **v2.29.0**: Major authentication architecture change - embedded credentials introduced
- **v2.28.0**: Session-based authentication with JWT tokens
