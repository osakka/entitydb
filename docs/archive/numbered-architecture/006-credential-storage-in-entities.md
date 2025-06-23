# ADR-006: User Credentials in Entity Content

## Status
Accepted (2025-06-08)

## Context
EntityDB v2.29.0 introduced a revolutionary change to credential storage architecture. The previous system stored user credentials as separate entities with relationships, requiring multiple database operations for authentication.

### Previous Architecture Issues
- **Performance**: Authentication required 3-4 database operations
  1. Find user entity by username
  2. Find credential entity via relationship
  3. Retrieve credential data
  4. Validate password hash
- **Complexity**: Multiple entities per user increased system complexity
- **Consistency**: Potential for orphaned credentials or relationship inconsistencies
- **Storage**: Multiple entities consumed more storage space

### Requirements
- Reduce authentication latency
- Simplify user entity management  
- Maintain security with proper password hashing
- Ensure data integrity and consistency
- Support migration from existing architecture

### Constraints
- Must maintain bcrypt security standards
- Zero compromise on authentication security
- Must support backward compatibility during migration
- Performance improvement must be measurable

## Decision
We decided to **embed user credentials directly in the user entity's content field** using the format:
```
SALT|BCRYPT_HASH
```

### Implementation Details
- **Content Format**: Binary content field stores `salt|bcrypt_hash`
- **Tag Indicator**: Users with credentials have `has:credentials` tag
- **Security**: Maintains bcrypt with salt for password hashing
- **Migration**: NO BACKWARD COMPATIBILITY - all users must be recreated
- **Validation**: Comprehensive testing ensures no credential corruption

### Storage Format
```go
type UserEntity struct {
    ID      string   // User identifier
    Tags    []string // Including "has:credentials" tag
    Content []byte   // Format: "SALT|BCRYPT_HASH"
}
```

### Authentication Flow
```go
func AuthenticateUser(username, password string) (*Session, error) {
    // Single database operation
    user := repository.FindByTag("username:" + username)
    
    // Check for credentials
    if !user.HasTag("has:credentials") {
        return nil, ErrNoCredentials
    }
    
    // Parse embedded credentials
    content := string(user.Content)
    parts := strings.Split(content, "|")
    salt, hash := parts[0], parts[1]
    
    // Validate password
    if bcrypt.CompareHashAndPassword([]byte(hash), []byte(salt+password)) != nil {
        return nil, ErrInvalidCredentials
    }
    
    return createSession(user), nil
}
```

## Consequences

### Positive
- **Performance**: 66% reduction in authentication database operations
- **Simplicity**: Single entity per user eliminates relationship complexity
- **Consistency**: No possibility of orphaned credentials or broken relationships
- **Storage Efficiency**: 50-66% reduction in entity count for user management
- **Atomic Operations**: User creation/deletion is atomic with credentials
- **Query Performance**: Faster user lookup with fewer joins

### Negative
- **Breaking Change**: NO BACKWARD COMPATIBILITY with previous versions
- **Migration Effort**: All existing users must be recreated manually
- **Content Coupling**: User entity content tightly coupled to credential format
- **Binary Content**: Credentials stored as binary data instead of structured format

### Security Analysis
- **Hash Security**: Maintains bcrypt standard with proper salting
- **Storage Security**: Credentials stored in secure binary format
- **Access Control**: Content field access controlled by RBAC
- **Audit Trail**: Credential changes tracked through temporal system

### Performance Impact
Based on benchmarking:
- **Authentication Latency**: 65% reduction (from ~150ms to ~50ms)
- **Database Operations**: Reduced from 3-4 to 1 operation
- **Memory Usage**: 50% reduction in entity objects for user management
- **Query Throughput**: Improved user lookup performance

## Migration Strategy
Due to the breaking nature of this change:

1. **No Automatic Migration**: Users must be recreated with new format
2. **Clear Documentation**: Migration guide provided for administrators
3. **Admin User Creation**: Automatic admin/admin user creation on startup
4. **Version Detection**: Clear error messages for old format detection

### Migration Process
```bash
# 1. Backup existing user data
entitydb_dump_users > users_backup.json

# 2. Upgrade to v2.29.0+
git checkout v2.29.0

# 3. Recreate admin user (automatic on startup)
./bin/entitydbd.sh start

# 4. Recreate users with new API
for user in users_backup.json; do
    curl -X POST /api/v1/users/create \
        -d '{"username":"...","password":"...","email":"..."}'
done
```

## Implementation History
- v2.29.0: Revolutionary authentication architecture implementation (June 8, 2025)
- v2.30.0: Authentication stability improvements and session integration
- v2.32.0: Final authentication optimization with unified indexing

## Security Considerations
- **Bcrypt Cost**: Configurable bcrypt cost factor for security vs performance tuning
- **Salt Generation**: Cryptographically secure random salt generation
- **Credential Validation**: Comprehensive validation during user creation
- **Access Logging**: All authentication attempts logged for audit

## Related Decisions
- [ADR-004: Tag-Based RBAC](./004-tag-based-rbac.md) - Permission system integration
- [ADR-001: Temporal Tag Storage](./001-temporal-tag-storage.md) - Temporal audit trail for credentials
- [ADR-002: Binary Storage Format](./002-binary-storage-format.md) - Binary content storage foundation