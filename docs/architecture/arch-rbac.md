# EntityDB RBAC Architecture

> **Status**: Core security feature since v2.0  
> **Last Updated**: June 7, 2025

## Overview

EntityDB implements a comprehensive Role-Based Access Control (RBAC) system using its native tag-based architecture. Every permission, role, and user is represented as entities with temporal tags, providing fine-grained security with complete audit trails.

## Core Concepts

### Everything is an Entity

In EntityDB, all security objects are entities with tags:
- **Users** - Stored as entities with authentication data
- **Roles** - Defined through permission tags
- **Sessions** - Managed as temporal entities
- **Permissions** - Expressed as hierarchical tags

### Permission Format

EntityDB uses a hierarchical permission format:

```
rbac:perm:resource:action:scope
```

**Examples**:
- `rbac:perm:entity:view` - View entities
- `rbac:perm:entity:create:dataset:worca` - Create entities in worca dataset
- `rbac:perm:system:admin` - System administration
- `rbac:perm:*` - All permissions (global admin)

### Role Format

Roles are assigned through tags:

```
rbac:role:role_name
```

**Examples**:
- `rbac:role:admin` - Administrator role
- `rbac:role:user` - Standard user role
- `rbac:role:viewer` - Read-only role

## User Management

### User Entities

Users are stored as entities with authentication credentials:

```json
{
  "id": "user_admin_123",
  "tags": [
    "type:user",
    "username:admin", 
    "rbac:role:admin",
    "rbac:perm:*",
    "status:active"
  ],
  "content": {
    "password_hash": "bcrypt_hash_here",
    "created_at": "2025-06-07T12:00:00Z",
    "last_login": "2025-06-07T14:30:00Z"
  }
}
```

### Permission Hierarchy

EntityDB supports hierarchical permissions with inheritance:

1. **Global Permissions**: `rbac:perm:*` (admin access)
2. **Resource Permissions**: `rbac:perm:entity:*` (all entity operations)
3. **Action Permissions**: `rbac:perm:entity:view` (specific operations)
4. **Scoped Permissions**: `rbac:perm:entity:view:dataset:worca` (dataset-specific)

### Default Roles

#### Administrator
```json
{
  "tags": [
    "rbac:role:admin",
    "rbac:perm:*"
  ]
}
```

#### Standard User  
```json
{
  "tags": [
    "rbac:role:user",
    "rbac:perm:entity:view",
    "rbac:perm:entity:create", 
    "rbac:perm:entity:update"
  ]
}
```

#### Read-Only User
```json
{
  "tags": [
    "rbac:role:viewer",
    "rbac:perm:entity:view"
  ]
}
```

## Implementation

### RBAC Middleware

**File**: `src/api/rbac_middleware.go`

Enforces permissions on every API request:

```go
func RBACMiddleware(entityRepo *binary.EntityRepository) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract user from session
            user, err := getUserFromSession(r, entityRepo)
            if err != nil {
                http.Error(w, "Authentication required", http.StatusUnauthorized)
                return
            }
            
            // Check required permission
            requiredPerm := getRequiredPermission(r)
            if !user.HasPermission(requiredPerm) {
                http.Error(w, "Insufficient permissions", http.StatusForbidden)
                return
            }
            
            // Add user context to request
            ctx := context.WithValue(r.Context(), "user", user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Permission Checking

**File**: `src/models/security.go`

Hierarchical permission evaluation:

```go
func (u *User) HasPermission(required string) bool {
    // Check for global admin
    if u.HasTag("rbac:perm:*") {
        return true
    }
    
    // Check exact permission match
    if u.HasTag(required) {
        return true
    }
    
    // Check hierarchical permissions
    parts := strings.Split(required, ":")
    for i := len(parts) - 1; i > 2; i-- {
        wildcard := strings.Join(parts[:i], ":") + ":*"
        if u.HasTag(wildcard) {
            return true
        }
    }
    
    return false
}
```

### Session Management

**File**: `src/models/session.go`

Sessions stored as temporal entities:

```go
type Session struct {
    ID       string    `json:"id"`
    UserID   string    `json:"user_id"`
    Token    string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
    Status   string    `json:"status"`
}

func CreateSession(userID string) (*Session, error) {
    session := &Session{
        ID:        generateSessionID(),
        UserID:    userID,
        Token:     generateSecureToken(),
        ExpiresAt: time.Now().Add(2 * time.Hour),
        Status:    "active",
    }
    
    // Store as entity with temporal tags
    entity := &Entity{
        ID:   session.ID,
        Tags: []string{
            "type:session",
            fmt.Sprintf("user_id:%s", userID),
            "status:active",
        },
        Content: marshalSession(session),
    }
    
    return session, repo.Create(entity)
}
```

## API Security

### Protected Endpoints

All endpoints require appropriate permissions:

```go
// Endpoint permission requirements
var endpointPermissions = map[string]string{
    "GET /api/v1/entities/list":     "rbac:perm:entity:view",
    "POST /api/v1/entities/create":  "rbac:perm:entity:create", 
    "PUT /api/v1/entities/update":   "rbac:perm:entity:update",
    "DELETE /api/v1/entities/delete": "rbac:perm:entity:delete",
    "POST /api/v1/users/create":     "rbac:perm:user:create",
    "GET /api/v1/system/metrics":    "rbac:perm:system:view",
    "POST /api/v1/admin/*":          "rbac:perm:system:admin",
}
```

### Dataset Security

Dataset-scoped permissions provide multi-tenant security:

```go
// Dataset permission check
func checkDatasetPermission(user *User, dataset, action string) bool {
    // Global admin override
    if user.HasPermission("rbac:perm:*") {
        return true
    }
    
    // Dataset-specific permission
    perm := fmt.Sprintf("rbac:perm:entity:%s:dataset:%s", action, dataset)
    return user.HasPermission(perm)
}
```

## Authentication Flow

### Login Process

1. **Credential Validation**: Check username/password against user entity
2. **Session Creation**: Generate secure session token
3. **Entity Storage**: Store session as temporal entity
4. **Token Return**: Provide session token to client

```bash
# Login request
curl -X POST https://localhost:8085/api/v1/auth/login \
  -d '{"username": "admin", "password": "admin"}'

# Response
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "user_admin_123", 
    "username": "admin",
    "role": "admin"
  }
}
```

### Request Authentication

1. **Token Extraction**: Get Bearer token from Authorization header
2. **Session Lookup**: Find active session entity by token
3. **User Loading**: Load user entity from session
4. **Permission Check**: Validate required permissions
5. **Context Addition**: Add user to request context

## Permission Management

### User Creation

```bash
# Create new user with specific permissions
curl -X POST https://localhost:8085/api/v1/users/create \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "username": "developer",
    "password": "secure_password",
    "tags": [
      "rbac:role:user",
      "rbac:perm:entity:view",
      "rbac:perm:entity:create:dataset:development"
    ]
  }'
```

### Role Assignment

```bash
# Add permissions to existing user
curl -X PUT https://localhost:8085/api/v1/users/update \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "user_id": "user_developer_456",
    "add_tags": ["rbac:perm:entity:update:dataset:development"]
  }'
```

### Permission Revocation

```bash
# Remove permissions from user
curl -X PUT https://localhost:8085/api/v1/users/update \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{
    "user_id": "user_developer_456", 
    "remove_tags": ["rbac:perm:entity:delete"]
  }'
```

## Security Features

### Secure Password Storage

- **bcrypt hashing** with configurable cost factor
- **Salt generation** for each password
- **Timing attack protection** using constant-time comparison

### Session Security

- **Secure token generation** using crypto/rand
- **Session expiration** with configurable TTL
- **Automatic cleanup** of expired sessions
- **Concurrent session support** per user

### Audit Logging

All authentication events are logged with temporal tags:

```json
{
  "timestamp": 1749303910369730667,
  "event": "user_login",
  "user_id": "user_admin_123",
  "ip_address": "192.168.1.100",
  "success": true
}
```

## Monitoring & Metrics

### RBAC Metrics

Available via `/api/v1/rbac/metrics` (admin only):

- **Active sessions**: Current session count
- **Authentication success/failure rates**
- **Permission check latency**
- **User activity patterns**

### Public Metrics

Available via `/api/v1/rbac/metrics/public` (no auth):

- **Total registered users**
- **Active sessions** (count only)
- **Authentication attempt rate**

## Configuration

### Security Settings

```bash
# Session configuration
ENTITYDB_SESSION_TTL_HOURS=2
ENTITYDB_TOKEN_SECRET=your-secret-key

# Password security
ENTITYDB_BCRYPT_COST=12
ENTITYDB_PASSWORD_MIN_LENGTH=8

# Authentication
ENTITYDB_MAX_LOGIN_ATTEMPTS=5
ENTITYDB_LOCKOUT_DURATION_MINUTES=15
```

### Auto-Initialization

EntityDB automatically creates an admin user on first startup:

```json
{
  "username": "admin",
  "password": "admin", 
  "tags": ["rbac:role:admin", "rbac:perm:*"]
}
```

## Best Practices

### Permission Design

1. **Principle of Least Privilege**: Grant minimum required permissions
2. **Dataset Isolation**: Use dataset-scoped permissions for multi-tenancy
3. **Role-Based**: Assign permissions through roles, not directly
4. **Hierarchical**: Use permission hierarchy for maintainability

### Security Guidelines

1. **Change Default Passwords**: Update admin password immediately
2. **Regular Token Rotation**: Implement token refresh in applications
3. **Monitor Sessions**: Track and audit authentication activity
4. **Network Security**: Use HTTPS in production environments

## Troubleshooting

### Common Issues

1. **Permission denied**: Check user permissions and hierarchy
2. **Session expired**: Implement token refresh in client applications
3. **Authentication failed**: Verify username/password and account status

### Debug Commands

```bash
# Check user permissions
curl "https://localhost:8085/api/v1/rbac/user-permissions?user_id=USER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# List active sessions
curl "https://localhost:8085/api/v1/rbac/sessions" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# View authentication metrics
curl "https://localhost:8085/api/v1/rbac/metrics" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

---

EntityDB's RBAC system provides enterprise-grade security with the flexibility and auditability of the temporal entity model.