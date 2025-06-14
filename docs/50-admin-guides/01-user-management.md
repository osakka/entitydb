# User Management Guide

> **Version**: v2.31.0 | **Last Updated**: 2025-06-14 | **Status**: AUTHORITATIVE

This guide covers comprehensive user management in EntityDB, including user creation, RBAC configuration, password management, and security best practices.

## Overview

EntityDB uses an embedded authentication system (v2.29.0+) where user credentials are stored directly in the user entity's content field as `salt|bcrypt_hash`. This eliminates the need for separate credential entities and simplifies user management.

## Default Administrator Account

### Initial Setup
EntityDB automatically creates a default admin user on first startup:
- **Username**: `admin`
- **Password**: `admin`
- **Permissions**: Full system access (`rbac:perm:*`)

### Security Warning
⚠️ **CRITICAL**: Change the default admin password immediately after installation:

```bash
# Login as admin
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# Extract token from response
export TOKEN="your-jwt-token-here"

# Change password (update entity content)
curl -X PUT http://localhost:8085/api/v1/entities/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "admin-user-id",
    "content": "new-salt|new-bcrypt-hash",
    "tags": ["type:user", "rbac:role:admin", "rbac:perm:*", "has:credentials"]
  }'
```

## Creating Users

### 1. Standard User Creation

```bash
# Create a new user entity
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:user",
      "id:username:john.doe",
      "rbac:role:user",
      "rbac:perm:entity:view",
      "rbac:perm:entity:create",
      "status:active",
      "has:credentials"
    ],
    "content": "generated-salt|bcrypt-hash-of-password"
  }'
```

### 2. Administrative User Creation

```bash
# Create an admin user
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:user",
      "id:username:jane.admin",
      "rbac:role:admin",
      "rbac:perm:*",
      "status:active",
      "has:credentials"
    ],
    "content": "generated-salt|bcrypt-hash-of-password"
  }'
```

### 3. Read-Only User Creation

```bash
# Create a read-only user
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:user",
      "id:username:viewer",
      "rbac:role:readonly",
      "rbac:perm:entity:view",
      "status:active",
      "has:credentials"
    ],
    "content": "generated-salt|bcrypt-hash-of-password"
  }'
```

## RBAC Permission System

### Core Permission Structure

EntityDB uses tag-based permissions with hierarchical inheritance:

```
rbac:perm:*                    # All permissions (admin)
rbac:perm:entity:*             # All entity permissions
rbac:perm:entity:view          # View entities
rbac:perm:entity:create        # Create entities
rbac:perm:entity:update        # Update entities
rbac:perm:entity:delete        # Delete entities
rbac:perm:relation:*           # All relationship permissions
rbac:perm:relation:create      # Create relationships
rbac:perm:relation:view        # View relationships
rbac:perm:user:*               # All user management permissions
rbac:perm:user:create          # Create users
rbac:perm:user:update          # Update users
rbac:perm:config:*             # All configuration permissions
rbac:perm:config:view          # View configuration
rbac:perm:config:update        # Update configuration
rbac:perm:system:*             # All system permissions
rbac:perm:system:view          # View system metrics
rbac:perm:metrics:read         # Read application metrics
```

### Standard Role Definitions

#### Admin Role
```json
{
  "tags": [
    "rbac:role:admin",
    "rbac:perm:*"
  ]
}
```

#### Power User Role
```json
{
  "tags": [
    "rbac:role:poweruser",
    "rbac:perm:entity:*",
    "rbac:perm:relation:*",
    "rbac:perm:system:view"
  ]
}
```

#### Standard User Role
```json
{
  "tags": [
    "rbac:role:user",
    "rbac:perm:entity:view",
    "rbac:perm:entity:create",
    "rbac:perm:entity:update",
    "rbac:perm:relation:view"
  ]
}
```

#### Read-Only Role
```json
{
  "tags": [
    "rbac:role:readonly",
    "rbac:perm:entity:view",
    "rbac:perm:relation:view"
  ]
}
```

## Password Management

### Password Hashing Process

EntityDB uses bcrypt with unique salts for each user:

1. **Generate Salt**: Create cryptographically secure random salt
2. **Hash Password**: Use bcrypt with salt and high cost factor
3. **Store Format**: `salt|bcrypt_hash` in entity content field

### Password Change Process

```bash
# 1. Generate new salt and hash (outside EntityDB)
NEW_SALT=$(openssl rand -hex 16)
NEW_HASH=$(python3 -c "
import bcrypt
salt = '$NEW_SALT'.encode()
password = 'new_password_here'.encode()
hash = bcrypt.hashpw(password, bcrypt.gensalt())
print(hash.decode())
")

# 2. Update user entity
curl -X PUT http://localhost:8085/api/v1/entities/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user-entity-id",
    "content": "'$NEW_SALT'|'$NEW_HASH'"
  }'
```

## User Management Operations

### Listing All Users

```bash
# Get all user entities
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?tags=type:user"

# Get users with credentials
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?tags=type:user,has:credentials"
```

### Finding Specific User

```bash
# Find user by username
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?tags=id:username:john.doe"
```

### Updating User Permissions

```bash
# Add permission to user
curl -X PUT http://localhost:8085/api/v1/entities/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user-entity-id",
    "tags": [
      "type:user",
      "id:username:john.doe",
      "rbac:role:user",
      "rbac:perm:entity:view",
      "rbac:perm:entity:create",
      "rbac:perm:entity:update",
      "rbac:perm:relation:view",
      "status:active",
      "has:credentials"
    ]
  }'
```

### Deactivating Users

```bash
# Deactivate user (change status)
curl -X PUT http://localhost:8085/api/v1/entities/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user-entity-id",
    "tags": [
      "type:user",
      "id:username:john.doe",
      "rbac:role:user",
      "status:inactive",
      "has:credentials"
    ]
  }'
```

### Removing Users

```bash
# Delete user entity (permanent)
curl -X DELETE http://localhost:8085/api/v1/entities/delete \
  -H "Authorization: Bearer $TOKEN" \
  -d "id=user-entity-id"
```

## Security Best Practices

### 1. Password Policy
- **Minimum Length**: 8 characters
- **Complexity**: Mix of uppercase, lowercase, numbers, symbols
- **Rotation**: Change passwords every 90 days for admin accounts
- **No Reuse**: Prevent password reuse

### 2. Permission Principle of Least Privilege
- Grant minimum permissions required for user's role
- Use specific permissions rather than wildcards when possible
- Regularly audit user permissions
- Remove unused permissions promptly

### 3. Account Security
- Change default admin password immediately
- Use strong, unique passwords for all accounts
- Disable or remove unused accounts
- Monitor authentication logs for suspicious activity

### 4. Session Management
- Configure appropriate JWT token expiry (default 24 hours)
- Implement token refresh for long-running sessions
- Monitor active sessions via RBAC metrics endpoint

## Monitoring and Auditing

### User Activity Monitoring

```bash
# Check RBAC metrics for authentication activity
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/rbac/metrics"

# View system metrics for overall activity
curl "http://localhost:8085/api/v1/system/metrics"
```

### Authentication Logs

Monitor authentication events through system logs:
```bash
# Check application logs for auth events
tail -f /opt/entitydb/var/log/entitydb.log | grep "auth"
```

## Troubleshooting

### Common Issues

#### Authentication Failures
```bash
# Check if user exists
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?tags=id:username:problematic-user"

# Verify user has credentials tag
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?tags=type:user,has:credentials"
```

#### Permission Denied Errors
```bash
# Check user permissions
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/get?id=user-entity-id"

# Verify required permissions for operation
# See API documentation for endpoint permission requirements
```

#### Token Expiry Issues
```bash
# Check token validity
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?limit=1"

# If expired, re-authenticate
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"your-username","password":"your-password"}'
```

## Migration from v2.28.0

⚠️ **BREAKING CHANGE**: Authentication architecture changed in v2.29.0

### Migration Requirements
- **NO BACKWARD COMPATIBILITY**: All existing users must be recreated
- Credentials now embedded in user entity content
- Separate credential entities no longer used

### Migration Process
1. **Export user list** from v2.28.0 system
2. **Upgrade to v2.31.0**
3. **Recreate all users** using new authentication format
4. **Notify users** of password reset requirement
5. **Update integrations** to use new authentication flow

## Quick Reference

### Essential Commands
```bash
# List all users
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?tags=type:user"

# Create user (requires proper salt|hash generation)
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags":["type:user","id:username:NAME","rbac:role:user","has:credentials"],"content":"SALT|HASH"}'

# Update permissions
curl -X PUT http://localhost:8085/api/v1/entities/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"id":"USER-ID","tags":["NEW-PERMISSION-TAGS"]}'
```

---

*This guide provides comprehensive user management capabilities for EntityDB v2.31.0. For additional security configuration, see [Security Configuration](./01-security-configuration.md).*