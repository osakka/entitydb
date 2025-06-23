# ADR-013: Pure Tag-Based Session Management

## Status
✅ **ACCEPTED** - 2025-06-15

## Context
EntityDB's session management system evolved from external storage to a hybrid model with both entity-based and external session tracking. This created complexity, potential inconsistencies, and violated the pure entity model principle that "everything is an entity with tags."

## Problem
- Hybrid session storage with both entities and external tracking
- Inconsistent session validation across different code paths
- Complex session lifecycle management with multiple storage locations
- Violation of "everything is an entity" architectural principle
- Potential for session state divergence between storage mechanisms

## Decision
Implement pure tag-based session management where sessions are stored exclusively as entities with tags:

### Pure Entity Session Model
```go
// Session as Entity with Tags
type SessionEntity struct {
    ID: "session_uuid"
    Tags: [
        "type:session",
        "token:hashed_token",
        "user_id:authenticated_user_uuid", 
        "expires:2025-06-15T15:30:00Z",
        "ip:127.0.0.1",
        "user_agent:browser_string",
        "status:active"
    ]
}
```

### Eliminated External Session Storage
- Remove all in-memory session maps
- Remove session databases or external storage
- All session operations through entity CRUD
- Session validation through tag-based queries

### Tag-Based Session Lifecycle
```go
// Session Creation
CreateEntity(sessionEntity, tags: ["type:session", "status:active", ...])

// Session Validation  
ListByTag("type:session AND token:hash AND status:active")

// Session Invalidation
UpdateEntity(sessionID, addTags: ["status:expired"])
```

## Implementation Details

### Core Changes
1. **Eliminated SessionManager**: Removed external session tracking
2. **Pure Entity Operations**: All session CRUD through EntityRepository
3. **Tag-Based Queries**: Session validation through tag filtering
4. **Temporal Session History**: Automatic session audit trail
5. **RBAC Integration**: Sessions subject to same permission model

### Session Security
- **Token Hashing**: Store bcrypt hashes, never plaintext tokens
- **Expiration Enforcement**: Time-based validation through tags
- **Audit Trail**: Complete session history through temporal tags
- **Permission Isolation**: Sessions isolated by dataset tags

### Migration Strategy
1. Convert existing sessions to entity format
2. Update authentication middleware to use tag-based validation
3. Remove legacy session storage code
4. Validate session operations through entity API

## Consequences

### Positive
- ✅ **Architectural Purity**: True "everything is an entity" model
- ✅ **Unified Storage**: Sessions use same durability as all data
- ✅ **Temporal History**: Complete session audit trail automatically
- ✅ **RBAC Integration**: Sessions subject to same security model
- ✅ **Simplified Codebase**: One storage model for all data types
- ✅ **Transactional Safety**: Session operations participate in WAL
- ✅ **Query Consistency**: Same query interface for all data

### Negative
- ⚠️ **Performance Overhead**: Entity operations slightly heavier than maps
- ⚠️ **Complexity**: Tag-based queries more complex than hash lookups
- ⚠️ **Storage Usage**: Sessions consume entity storage space

### Risks Mitigated
- 🔒 **Session Inconsistency**: Single source eliminates divergence
- 🔒 **Data Loss**: Sessions benefit from WAL durability
- 🔒 **Security Gaps**: Sessions follow same security model as entities
- 🔒 **Audit Requirements**: Complete session history automatically available

## Performance Characteristics
- **Session Validation**: ~2-5ms through tag index (vs <1ms hash lookup)
- **Session Creation**: ~10-15ms entity operation (vs ~1ms map insert)
- **Storage Overhead**: ~200 bytes per session (vs ~50 bytes in-memory)
- **Durability**: Full WAL protection (vs memory volatility)

## Alternatives Considered
1. **Hybrid Model**: Continue dual storage - rejected for complexity
2. **External Session DB**: Use separate database - rejected for architecture violation
3. **In-Memory Only**: Fast but loses durability - rejected for reliability

## References
- Implementation: `src/models/security.go` - session validation
- Authentication: `src/api/auth_handler.go` - session creation/validation
- Git Commit: `b91d85a` - "feat: complete migration to pure tag-based session management"
- Related: ADR-001 (Temporal Tag Storage), ADR-004 (Tag-Based RBAC)

## Timeline
- **2025-06-15**: Decision made to eliminate hybrid session storage
- **2025-06-15**: Pure entity session implementation completed
- **2025-06-15**: Legacy session storage code removed
- **2025-06-16**: Comprehensive testing and validation

---
*This ADR documents the architectural decision to implement pure tag-based session management, ensuring all session operations follow the same entity model as other data while maintaining security and performance requirements.*