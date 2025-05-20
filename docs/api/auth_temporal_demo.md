# Authentication and Temporal System Demo

This demonstrates how authentication works transparently with the temporal system in EntityDB v2.8.0.

## Key Points

1. **No Temporal Handling Required for Authentication**
   - Login requests remain the same
   - No timestamps needed in auth requests
   - Session tokens work as before

2. **Temporal Data is Internal**
   - User entities have timestamped tags internally
   - Authentication logic uses current state
   - Password changes are tracked temporally

## Authentication Flow

```json
// 1. Login Request (No temporal handling needed)
POST /api/v1/auth/login
{
  "username": "admin",
  "password": "yourpassword"
}

// Response (No timestamps shown)
{
  "session_token": "abc123...",
  "user_id": "3c116c56-a2f3-4fb2-9607-d70cf85a",
  "expires_at": "2025-05-19T22:45:00Z"
}
```

## Behind the Scenes

When you login, the system:

1. Finds the user entity with tag `id:username:admin`
2. Reads the current password_hash field
3. Verifies the password using bcrypt
4. Creates a session with timestamps

All temporal tracking happens automatically:

```
# User entity internally has timestamped tags:
2025-05-18T22:38:48.807532034+01:00|type:user
2025-05-18T22:38:48.807532034+01:00|id:username:admin
2025-05-18T22:38:48.807532034+01:00|rbac:role:admin
2025-05-18T22:38:48.807532034+01:00|rbac:perm:*
2025-05-18T22:38:48.807532034+01:00|status:active
```

## Using Authenticated Endpoints

```json
// Make requests with session token (No temporal handling)
GET /api/v1/entities/list
Authorization: Bearer abc123...

// Response has clean tags by default
{
  "entities": [{
    "id": "uuid-here",
    "tags": ["type:issue", "status:open"],
    "fields": {...}
  }]
}

// Optionally request with timestamps
GET /api/v1/entities/list?include_timestamps=true
Authorization: Bearer abc123...

// Response includes temporal data
{
  "entities": [{
    "id": "uuid-here", 
    "tags": [
      "2025-05-18T22:45:00.000000000+00:00|type:issue",
      "2025-05-18T22:45:00.000000000+00:00|status:open"
    ],
    "fields": {...}
  }]
}
```

## Password Changes Track History

When an admin password is changed:

```
# Old state
password_hash: $2a$10$old...

# After change - both states exist temporally
2025-05-18T22:00:00.000000000+00:00|password_hash:$2a$10$old...
2025-05-18T22:45:00.000000000+00:00|password_hash:$2a$10$new...
```

Authentication always uses the most recent password.

## Summary

✅ **Temporal is Transparent for Auth**
- No changes to login requests
- No changes to session handling  
- No timestamps in auth payloads

✅ **Benefits of Temporal Auth**
- Full audit trail of password changes
- Track role/permission changes over time
- Query historical user states
- But none of this complexity exposed to API users

The temporal system operates behind the scenes, maintaining history while presenting a clean, simple API interface.