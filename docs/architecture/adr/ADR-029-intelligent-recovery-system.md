# ADR-029: Intelligent Recovery System Architecture

**Status**: Accepted  
**Date**: 2025-06-21  
**Authors**: EntityDB Core Team  
**Relates to**: ADR-007 (Temporal Retention), ADR-027 (Database File Unification)

## Context

EntityDB's recovery system was experiencing catastrophic performance issues due to indiscriminate recovery attempts. The background metrics collector was generating lookups for non-existent metric entities, triggering expensive recovery operations for entities that were never supposed to exist in the first place.

### Problem Details

1. **Infinite Recovery Loops**: Metrics collector looks up 20+ metrics per minute, each generating random entity IDs when metrics don't exist
2. **CPU Explosion**: Recovery system attempting file scans for every failed lookup, causing 100% CPU usage
3. **Log Spam**: Hundreds of recovery messages cluttering operational visibility
4. **Performance Degradation**: System becoming unresponsive during metrics collection cycles

### Previous Recovery Approach
```go
// Old approach: Attempt recovery for ANY missing entity
if err != nil {
    if recoveredEntity, recErr := r.recovery.RecoverCorruptedEntity(r, id); recErr == nil {
        return recoveredEntity, nil
    }
}
```

This indiscriminate approach treated all missing entities as corruption requiring recovery.

## Decision

We implement an **Intelligent Recovery System** that distinguishes between entities that should exist (and warrant expensive recovery) versus artifacts from normal operations (that should fail gracefully).

### Core Principle
**Only attempt recovery for entities that have a reasonable expectation of existing.**

### Entity Classification Strategy

| Entity Type | Pattern | Recovery Decision | Reasoning |
|-------------|---------|-------------------|-----------|
| System User | `00000000000000000000000000000001` | **Always Recover** | Critical for authentication |
| Pure Hex IDs | 32 chars, only `[0-9a-fA-F]` | **Skip Recovery** | Likely metric lookup artifacts |
| Mixed UUIDs | 32 chars, alphanumeric mix | **Attempt Recovery** | Likely real user/entity UUIDs |
| Non-standard | Length ≠ 32 | **Attempt Recovery** | May be legitimate legacy format |

### Implementation Architecture

```go
func shouldAttemptEntityRecovery(entityID string) bool {
    // Critical system entities always warrant recovery
    if entityID == models.SystemUserID {
        return true
    }
    
    // Skip recovery for pure hex entities (metric artifacts)
    if len(entityID) == 32 && isAllHex(entityID) {
        return false
    }
    
    // Mixed alphanumeric suggests real UUID
    if len(entityID) == 32 && !isAllHex(entityID) {
        return true
    }
    
    // Other formats may be legitimate
    return len(entityID) != 32
}
```

## Consequences

### Positive
- **50% CPU Reduction**: Eliminated expensive recovery attempts for metric artifacts
- **Clean Logging**: Recovery messages only for entities that warrant investigation
- **Faster Response**: System remains responsive during metrics collection
- **Intelligent Behavior**: Recovery system focuses on entities likely to exist

### Negative
- **Pattern Dependency**: Recovery decisions based on entity ID patterns
- **Edge Cases**: Unusual but legitimate entity ID formats might be skipped
- **Complexity**: Additional logic in recovery path

### Risk Mitigation
- **Conservative Approach**: When in doubt, attempt recovery (len ≠ 32)
- **System Critical**: Always recover system user (authentication dependency)
- **Monitoring**: Track recovery attempt patterns to identify new issues

## Implementation Details

### Files Modified
- `entity_repository.go`: Core recovery decision logic
- `recovery.go`: Already had hex string detection for filtering

### Performance Impact
- **Before**: 1.12+ CPU load (100% sustained)
- **After**: 0.58 CPU load (50% reduction)
- **Recovery Rate**: ~95% reduction in unnecessary attempts

### Behavioral Changes
```bash
# Before: Expensive recovery for every failed lookup
ERROR: entity not found
INFO: attempting to recover corrupted entity
WARN: attempting partial recovery by file scan

# After: Clean failure for metric artifacts  
ERROR: entity not found
TRACE: skipping recovery for non-critical entity (likely metrics lookup artifact)
```

## Alternatives Considered

### 1. Disable Recovery Entirely
**Rejected**: Would prevent legitimate corruption recovery

### 2. Timeout-Based Recovery
**Rejected**: Still performs expensive operations, just limits duration

### 3. Metrics Collection Fix
**Rejected**: Addresses symptom but not architectural issue

### 4. Recovery Cache Only
**Rejected**: Still attempts initial expensive recovery

## Monitoring and Validation

### Success Metrics
- CPU usage during metrics collection cycles
- Recovery attempt frequency and patterns
- System entity recovery success rate
- Authentication system stability

### Alerting Thresholds
- Sustained recovery attempts (>10/minute) may indicate new issues
- System user recovery failures (critical alert)
- CPU usage spike correlation with metrics collection

## Future Considerations

### Entity ID Standards
Consider formalizing entity ID patterns to improve recovery intelligence:
- System entities: Reserved patterns
- User entities: UUID v4 standard
- Metric entities: Prefixed patterns (`metric_*`)

### Recovery Telemetry
Add metrics for recovery decision outcomes:
- Entities skipped by pattern
- Successful recoveries by entity type
- Recovery attempt success rates

### Pattern Evolution
Monitor for new entity ID patterns that should be classified:
- Application-generated entities
- Import/export temporary entities
- Cross-system federation entities

## Related ADRs

- **ADR-007**: Temporal retention cleanup (reduced recovery noise)
- **ADR-027**: Database file unification (simplified recovery targets)
- **ADR-028**: Configuration management (recovery system configurability)

## Conclusion

The Intelligent Recovery System represents a fundamental shift from "recover everything" to "recover intelligently." This architecture eliminates performance issues while maintaining critical data recovery capabilities.

The pattern-based approach provides immediate benefits while establishing a framework for more sophisticated recovery intelligence as EntityDB evolves.

**Decision**: Accepted - Implementation provides significant performance improvements with acceptable complexity trade-offs.