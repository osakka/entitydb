# ADR-016: Error Recovery and Resilience Architecture

## Status
✅ **ACCEPTED** - 2025-06-16

## Context
EntityDB required robust error recovery mechanisms to handle corrupted entities, index inconsistencies, and system failures gracefully. The system needed to be resilient to various failure modes while maintaining data integrity and availability.

## Problem
- Entity corruption could render specific entities inaccessible
- Index corruption could cause system-wide query failures
- Missing entities in indexes could break functionality
- No automatic recovery mechanisms for common failure scenarios
- System failures could cascade without proper error boundaries
- Lack of resilience for production deployment requirements

## Decision
Implement comprehensive error recovery and resilience architecture:

### Entity-Level Recovery
```go
type RecoveryManager struct {
    WALReplay       func(entityID string) error
    IndexRebuild    func(entityID string) error  
    PartialRecovery func(entityID string) error
    FileSystemScan  func(entityID string) (*Entity, error)
}
```

### Multi-Tier Recovery Strategy
1. **WAL Replay Recovery**: Attempt to recover from Write-Ahead Log
2. **Index Rebuild**: Reconstruct indexes for specific entities
3. **Partial Recovery**: Create placeholder entities for system continuity
4. **File System Scan**: Direct binary file scanning as last resort

### Automatic Recovery Integration
```go
func (repo *EntityRepository) GetByID(id string) (*Entity, error) {
    entity, err := repo.getFromIndex(id)
    if err != nil {
        // Automatic recovery attempt
        if recovered, recErr := repo.recoverEntity(id); recErr == nil {
            return recovered, nil
        }
        return nil, err
    }
    return entity, nil
}
```

## Implementation Details

### Recovery Mechanisms

#### WAL Replay Recovery
- **Purpose**: Recover entities from recent operations logged in WAL
- **Use Case**: Index corruption but WAL intact
- **Process**: Replay WAL entries to reconstruct entity state
- **Success Rate**: High for recently modified entities

#### Index Rebuild Recovery  
- **Purpose**: Reconstruct missing index entries
- **Use Case**: Entity exists in storage but missing from index
- **Process**: Scan binary storage and rebuild index entries
- **Success Rate**: High for storage-intact scenarios

#### Partial Recovery
- **Purpose**: Maintain system functionality when entity is corrupted
- **Use Case**: Critical system entities (like system user) corrupted
- **Process**: Create minimal placeholder entity with essential tags
- **Success Rate**: 100% for system continuity

#### File System Scan Recovery
- **Purpose**: Last resort recovery from binary files
- **Use Case**: Both index and WAL corrupted
- **Process**: Direct binary file parsing to extract entity data
- **Success Rate**: Moderate, depends on file integrity

### Error Boundaries and Resilience

#### Graceful Degradation
```go
func (h *EntityHandler) handleEntityAccess(entityID string) {
    entity, err := h.repo.GetByID(entityID)
    if err != nil {
        // Log error but continue with degraded functionality
        h.logger.Warn("Entity access failed, using cached data: %v", err)
        return h.getCachedEntity(entityID)
    }
    return entity
}
```

#### Circuit Breaker Pattern
- **Failed Operation Tracking**: Count consecutive failures
- **Automatic Fallback**: Switch to alternative methods after threshold
- **Recovery Detection**: Periodically test if normal operation restored
- **Metrics Integration**: Track circuit breaker states

### Recovery Logging and Monitoring
```go
type RecoveryMetrics struct {
    AttemptedRecoveries int    `json:"attempted_recoveries"`
    SuccessfulRecoveries int   `json:"successful_recoveries"`
    PartialRecoveries   int    `json:"partial_recoveries"`
    FailedRecoveries    int    `json:"failed_recoveries"`
    RecoveryMethods     map[string]int `json:"recovery_methods"`
}
```

## Consequences

### Positive
- ✅ **High Availability**: System continues operating despite entity corruption
- ✅ **Data Integrity**: Recovery mechanisms preserve data where possible
- ✅ **Automatic Healing**: No manual intervention required for common failures
- ✅ **Graceful Degradation**: System functionality preserved during failures
- ✅ **Production Reliability**: Robust error handling for enterprise deployment
- ✅ **Observability**: Comprehensive logging and metrics for recovery operations

### Negative
- ⚠️ **Complexity**: Additional code paths and error handling logic
- ⚠️ **Performance Overhead**: Recovery attempts add latency to failed operations
- ⚠️ **Resource Usage**: Recovery operations consume CPU and I/O

### Risk Mitigation
- 🔒 **Entity Corruption**: Multiple recovery strategies prevent total data loss
- 🔒 **Index Corruption**: Automatic index rebuild maintains query functionality
- 🔒 **Cascade Failures**: Error boundaries prevent failure propagation
- 🔒 **System Instability**: Graceful degradation maintains core functionality

## Recovery Scenarios

### System User Recovery
```go
func (s *SystemUser) InitializeSystemUser() error {
    systemUser, err := s.repo.GetByID(SystemUserUUID)
    if err != nil {
        s.logger.Warn("System user recovery placeholder - will replace")
        return s.createSystemUser() // Create fresh system user
    }
    return nil
}
```

### Index Corruption Recovery
```go
func (repo *EntityRepository) recoverFromIndexCorruption() error {
    // Full index rebuild from storage
    entities, err := repo.scanAllEntities()
    if err != nil {
        return err
    }
    
    return repo.rebuildIndexes(entities)
}
```

### Session Recovery
```go
func (sm *SecurityManager) ValidateSession(token string) (*SecurityUser, error) {
    session, err := sm.getSession(token)
    if err != nil {
        // Attempt session recovery from entity storage
        if recovered := sm.recoverSession(token); recovered != nil {
            return recovered, nil
        }
        return nil, err
    }
    return session, nil
}
```

## Monitoring and Alerting
- **Recovery Rate Tracking**: Monitor success/failure rates
- **Performance Impact**: Track recovery operation duration
- **Failure Pattern Analysis**: Identify recurring failure types
- **Proactive Alerts**: Notify on unusual recovery activity

## Alternatives Considered
1. **Fail Fast**: Rejected for poor user experience
2. **External Recovery Tools**: Rejected for operational complexity
3. **Manual Recovery Only**: Rejected for production requirements
4. **Full System Recovery**: Rejected for performance impact

## References
- Implementation: `src/storage/binary/recovery.go` - recovery mechanisms
- Integration: `src/storage/binary/entity_repository.go` - automatic recovery
- System User: `src/models/system_user.go` - critical entity recovery
- Git Commits: System user recovery and error resilience improvements
- Related: ADR-015 (WAL Management), ADR-002 (Binary Storage Format)

## Timeline
- **2025-06-16**: Error recovery requirements identified
- **2025-06-16**: Multi-tier recovery architecture designed
- **2025-06-16**: Automatic recovery integration implemented
- **2025-06-17**: Production validation and resilience testing

---
*This ADR documents the architectural decision to implement comprehensive error recovery and resilience mechanisms, ensuring EntityDB maintains high availability and data integrity even in the face of various failure scenarios.*