# Authentication Architecture v2.29.0

> [!CRITICAL]
> This document describes the NEW authentication architecture introduced in v2.29.0.
> There is NO backward compatibility with previous versions.

## Overview

Starting with v2.29.0, EntityDB uses a simplified authentication architecture where user credentials are stored directly within the user entity's content field. This eliminates the need for separate credential entities and relationships.

## Architecture

### User Entity Structure

```go
Entity {
    ID: "user_123abc...",
    Tags: [
        "type:user",
        "dataspace:_system",
        "identity:username:john",
        "identity:uuid:user_123abc...",
        "status:active",
        "profile:email:john@example.com",
        "has:credentials",  // Indicates embedded credentials
        "rbac:role:admin",  // Direct role assignment
        "created:1749380951675261627"
    ],
    Content: []byte("salt|bcrypt_hash"), // Credentials stored here
}
```

### Key Changes

1. **No Credential Entities**: Credentials are no longer separate entities
2. **No Credential Relationships**: No `has_credential` relationships needed
3. **Direct Storage**: Password hash and salt stored in user's content field
4. **Tag Indicator**: `has:credentials` tag indicates user has embedded credentials
5. **Simple Format**: Content format is `salt|bcrypt_hash`

## Authentication Flow

### User Creation

```go
func CreateUser(username, password, email string) (*SecurityUser, error) {
    // 1. Generate user ID
    userID := "user_" + generateSecureUUID()
    
    // 2. Hash password with salt
    salt := generateSalt()
    hashedPassword := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
    
    // 3. Create user entity with embedded credentials
    userEntity := &Entity{
        ID: userID,
        Tags: []string{
            "type:user",
            "dataspace:_system",
            "identity:username:" + username,
            "identity:uuid:" + userID,
            "status:active",
            "profile:email:" + email,
            "has:credentials",  // Mark as having credentials
            "created:" + NowString(),
        },
        Content: []byte(fmt.Sprintf("%s|%s", salt, string(hashedPassword))),
    }
    
    // 4. Single entity creation
    return entityRepo.Create(userEntity)
}
```

### User Authentication

```go
func AuthenticateUser(username, password string) (*SecurityUser, error) {
    // 1. Find user by username (single query)
    users, err := entityRepo.ListByTag("identity:username:" + username)
    
    // 2. Check user status and credentials tag
    user := users[0]
    if !hasTag(user, "status:active") {
        return nil, fmt.Errorf("user account is not active")
    }
    if !hasTag(user, "has:credentials") {
        return nil, fmt.Errorf("no credentials found for user")
    }
    
    // 3. Extract credentials from content
    parts := strings.SplitN(string(user.Content), "|", 2)
    salt := parts[0]
    hashedPassword := []byte(parts[1])
    
    // 4. Verify password
    err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password+salt))
    if err != nil {
        return nil, fmt.Errorf("invalid password")
    }
    
    return user, nil
}
```

## Performance Benefits

| Operation | Old Architecture | New Architecture | Improvement |
|-----------|-----------------|------------------|-------------|
| User Creation | 2 entities + 1 relationship | 1 entity | 66% fewer operations |
| Authentication | 2-3 entity reads | 1 entity read | 50-66% fewer reads |
| Storage per User | ~1KB (user + credential + relationship) | ~500B | 50% less storage |
| Index Entries | 3-4 per user | 1 per user | 66-75% fewer indexes |

## Security Considerations

1. **Content Privacy**: Credentials are stored in the content field, which is never indexed or searchable
2. **Tag Security**: The `has:credentials` tag is public metadata but reveals no sensitive information
3. **Bcrypt Security**: Standard bcrypt hashing with configurable cost factor
4. **Salt Storage**: Each user has a unique salt stored with their hash

## Migration

> [!WARNING]
> There is NO automatic migration from the old system. All users must be recreated.

### Manual Migration Steps

1. Export user data from old system (usernames, emails, roles)
2. Delete the database files
3. Restart the server (admin/admin user will be auto-created)
4. Recreate users with new passwords

## Related Documentation

- [Entity Architecture](entities.md)
- [RBAC System](rbac.md)
- [Security Overview](../security.md)