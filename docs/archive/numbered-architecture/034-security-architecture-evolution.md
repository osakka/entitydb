# ADR-034: Security Architecture Evolution from Component-Based to Unified Model

**Status**: Accepted  
**Date**: 2025-06-08  
**Authors**: EntityDB Architecture Team  
**Related**: ADR-006 (Credential Storage), ADR-004 (Tag-Based RBAC)

## Context

This ADR documents the evolution of EntityDB's security architecture from a complex component-based system to the current unified entity-based security model. The old repository reveals a sophisticated but complex security system that was simplified and unified with the overall entity architecture.

## Historical Security Architecture

### Original Component-Based Security System
The old repository's `deprecated/security_components_backup/` shows the original architecture:

```go
// Original security components
type SecurityManager struct {
    validator   *InputValidator
    auditLogger *AuditLogger  
    server      interface{}
}

type InputValidator struct {
    maxLength      int
    allowedChars   map[rune]bool
    blockedPatterns []string
}

type AuditLogger struct {
    logDir    string
    entities  map[string]map[string]interface{}
    logFile   *os.File
}
```

### Separate Security Components
The old implementation featured distinct security layers:

#### 1. Input Validation System
- Separate `InputValidator` component
- Complex character validation rules
- Pattern-based attack detection
- Length and format validation

#### 2. Audit Logging System  
- Dedicated `AuditLogger` component
- File-based audit trails
- Separate audit database
- Complex log rotation and management

#### 3. Authentication Manager
- Separate authentication component
- In-memory session management
- No password hashing (documented security gap)
- Hardcoded admin credentials

#### 4. Authorization Bridge
- Complex permission bridging system
- Separate RBAC rule engine
- Multiple authorization layers
- Performance overhead from component interaction

### Security Gaps in Original System
From the old repository's `IMPLEMENTATION_STATUS.md`:

```
Authentication ⚠️
- Token-based auth (in-memory)
- No password hashing
- Admin user hardcoded  
- No session management

What's NOT Implemented ❌
- Permission enforcement (RBAC defined but not used)
- User password hashing
- Middleware/interceptors
```

## Problems with Component-Based Security

### 1. Complexity and Fragmentation
- Multiple separate security components
- Complex initialization and dependency management
- Difficult to maintain consistency across components
- Performance overhead from component interactions

### 2. Security Gaps
Evidence from old repository:
- No password hashing implementation
- Hardcoded admin credentials
- In-memory session management without persistence
- RBAC defined but not enforced

### 3. Maintenance Burden
- Separate codebases for each security component
- Complex configuration management
- Difficult to ensure security consistency
- Multiple potential failure points

## Decision Rationale

### Unified Security Vision
Integrate security into the core entity model rather than maintaining separate components:

1. **User Credentials as Entity Content**: Store hashed credentials directly in user entity content
2. **Sessions as Entities**: Represent sessions as entities with tags for state management
3. **Permissions as Tags**: Use entity tags for RBAC permissions
4. **Unified Storage**: All security data in the same binary format

### Benefits of Entity-Based Security

#### 1. Simplification
- Single storage mechanism for all security data
- Unified CRUD operations for security entities
- Consistent backup and recovery for security data
- Simplified configuration management

#### 2. Performance
- Binary format optimization for security queries
- Tag-based permission lookups
- Memory-mapped access for session data
- Reduced component interaction overhead

#### 3. Consistency
- All security data follows same entity model
- Temporal tracking of security events
- Unified relationship modeling for permissions
- Consistent error handling and logging

## Implementation Strategy

### Migration from Component-Based System

#### Phase 1: Credential Storage Integration
From ADR-006, credentials moved to entity content:
```go
// Old approach: separate credential component
type Credential struct {
    UserID   string
    Password string  // No hashing!
    Salt     string
}

// New approach: embedded in user entity
type User struct {
    ID      string
    Tags    []string
    Content []byte   // "salt|bcrypt_hash"
}
```

#### Phase 2: Session Management as Entities
Sessions became entities with tag-based state:
```go
// Session entity with tags
entity := &Entity{
    ID: sessionID,
    Tags: []string{
        "type:session",
        "user_id:" + userID,
        "expires_at:" + expirationTime,
        "status:active",
    },
    Content: []byte(sessionToken),
}
```

#### Phase 3: RBAC Tag Integration
Permissions stored as entity tags:
```go
// User with RBAC permissions
userEntity := &Entity{
    ID: userID,
    Tags: []string{
        "type:user",
        "username:" + username,
        "rbac:role:admin",
        "rbac:perm:entity:*",
        "rbac:perm:config:view",
        "has:credentials",
    },
}
```

#### Phase 4: Component Deprecation
Old security components moved to `deprecated/security_components_backup/`:
- SecurityManager removed
- InputValidator removed
- AuditLogger removed
- Security bridge components removed

## Technical Implementation

### Unified Security Model

#### 1. User Authentication
```go
// Secure credential storage in entity content
func (s *SecurityManager) AuthenticateUser(username, password string) (*Entity, error) {
    user := s.getUserByUsername(username)
    if user == nil {
        return nil, ErrUserNotFound
    }
    
    // Extract salt|hash from entity content
    parts := strings.Split(string(user.Content), "|")
    salt, hash := parts[0], parts[1]
    
    // Verify password with bcrypt
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(salt+password))
}
```

#### 2. Session Management
```go
// Sessions as entities with temporal tracking
func (s *SecurityManager) CreateSession(userID string) (*Entity, error) {
    sessionID := generateSessionID()
    token := generateSecureToken()
    
    sessionEntity := &Entity{
        ID: sessionID,
        Tags: []string{
            "type:session",
            "user_id:" + userID,
            "expires_at:" + time.Now().Add(24*time.Hour).Format(time.RFC3339),
            "status:active",
        },
        Content: []byte(token),
    }
    
    return s.repository.Create(sessionEntity)
}
```

#### 3. Permission Checking
```go
// Tag-based permission checking
func (s *SecurityManager) HasPermission(userID, permission string) bool {
    user := s.getUserByID(userID)
    if user == nil {
        return false
    }
    
    // Check for specific permission or wildcard
    permissionTag := "rbac:perm:" + permission
    wildcardTag := "rbac:perm:" + strings.Split(permission, ":")[0] + ":*"
    
    return user.HasTag(permissionTag) || user.HasTag(wildcardTag)
}
```

## Security Improvements Achieved

### 1. Password Security
- **Before**: No password hashing (plaintext storage)
- **After**: bcrypt hashing with salt stored in entity content
- **Impact**: Production-grade password security

### 2. Session Management
- **Before**: In-memory sessions (lost on restart)
- **After**: Persistent sessions as entities with temporal tracking
- **Impact**: Sessions survive server restarts, full audit trail

### 3. Permission Enforcement
- **Before**: RBAC defined but not enforced
- **After**: Comprehensive tag-based RBAC with middleware enforcement
- **Impact**: Full security model with fine-grained permissions

### 4. Audit Capabilities
- **Before**: Separate audit logging system
- **After**: Built-in temporal tracking of all security events
- **Impact**: Complete audit trail with nanosecond precision

## Performance Impact

### Component Elimination Benefits
- **Reduced Memory**: Eliminated separate security component overhead
- **Faster Authentication**: Direct entity lookups vs component interactions
- **Improved Caching**: Security data benefits from entity caching
- **Better Scaling**: Security scales with entity system performance

### Security Query Performance
- Tag-based permission checks: O(1) with proper indexing
- Session validation: Direct entity lookup
- User authentication: Single entity read operation
- RBAC evaluation: Tag intersection operations

## Migration Challenges Overcome

### 1. Data Migration
- Migrated existing users to new credential format
- Converted sessions to entity-based storage
- Preserved permission mappings during RBAC transition
- Maintained backward compatibility during migration

### 2. API Compatibility
- Maintained existing authentication endpoints
- Preserved session token format
- Kept permission checking interface consistent
- No breaking changes for clients

### 3. Security Continuity
- No security gaps during migration
- Gradual component deprecation
- Maintained audit trail continuity
- Preserved existing user accounts

## Consequences

### Positive Outcomes

#### 1. Simplified Architecture
- Single security model integrated with entity system
- Eliminated component complexity and dependencies
- Unified storage and backup for security data
- Consistent error handling and logging

#### 2. Enhanced Security
- Production-grade password hashing implemented
- Persistent session management with full audit trail
- Comprehensive RBAC enforcement across all operations
- Temporal tracking of all security events

#### 3. Performance Improvements
- Faster authentication through direct entity operations
- Improved permission checking performance
- Better caching of security data
- Reduced memory footprint

#### 4. Maintainability
- Single codebase for all security functionality
- Unified testing approach for security features
- Consistent configuration management
- Simplified deployment and monitoring

### Security Architecture Benefits
1. **Defense in Depth**: Multiple security layers integrated seamlessly
2. **Audit Completeness**: Full temporal tracking of security events
3. **Scalability**: Security scales with entity system performance
4. **Flexibility**: Easy to add new security features and permissions

## Historical Lessons Learned

### Success Factors
1. **Gradual Migration**: Phased approach maintained security continuity
2. **Integration vs Separation**: Unified model simplified maintenance
3. **Performance Focus**: Entity-based approach improved performance
4. **Security First**: Never compromised security during migration

### Best Practices Established
1. **Security Integration**: Integrate security with core data model
2. **Credential Protection**: Always use proper password hashing
3. **Session Persistence**: Maintain sessions across server restarts
4. **Permission Enforcement**: Implement comprehensive RBAC
5. **Audit Completeness**: Track all security events temporally

## Future Implications

This security architecture evolution established EntityDB's principle of **integrated security** and demonstrates that security can be both simpler and stronger when properly integrated with the core data model. This influences all future security decisions toward unified, entity-based approaches.

## References

- Old repository: `deprecated/security_components_backup/`
- Old repository: `docs/archive/security_implementation_*.md`
- ADR-006: User Credentials in Entity Content
- ADR-004: Tag-Based RBAC System
- Current implementation: `src/models/security.go`

---

**Implementation Status**: Complete  
**Migration Date**: 2025-06-08  
**Security Level**: Production-grade with comprehensive RBAC  
**Component Reduction**: 4 components → unified entity model