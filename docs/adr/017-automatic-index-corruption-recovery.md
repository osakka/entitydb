# ADR-017: Automatic Index Corruption Recovery

## Status
**ACCEPTED** - Implemented in v2.32.0

## Context

EntityDB experienced critical production stability issues where index corruption caused:
- 36-second query response times (normal: ~70ms)
- Authentication timeouts leading to user session failures
- "Kicked out directly" user experience during sluggish server performance
- Manual intervention required for index recovery

The root cause was identified as stale/corrupted index files that were significantly older than the primary database file, containing invalid offset data that caused expensive linear scans through large datasets.

## Decision

Implement a comprehensive automatic index corruption recovery system that:

1. **Detects corruption automatically** during repository initialization
2. **Recovers without external intervention** - no CLI tools required
3. **Maintains system transparency** through detailed logging
4. **Preserves data integrity** with automatic backup creation
5. **Follows single source of truth** architecture principles

## Architecture

### Core Components

```go
// IndexCorruptionRecovery manages detection and recovery
type IndexCorruptionRecovery struct {
    dataPath string
}

// Main recovery workflow
func (icr *IndexCorruptionRecovery) DiagnoseAndRecover() error {
    // 1. Diagnose corruption
    // 2. Create backup
    // 3. Rebuild index
    // 4. Validate recovery
}
```

### Integration Points

1. **Repository Initialization**: `EntityRepository.performAutomaticRecovery()`
2. **Corruption Detection**: Enhanced `reader.go` validation
3. **Logging Integration**: Comprehensive recovery operation logging
4. **Startup Flow**: Seamless integration with normal database startup

### Recovery Process

```
1. Repository Initialization
   ↓
2. Automatic Corruption Detection
   - File timestamp comparison
   - Magic number validation
   - Offset range validation
   ↓
3. Backup Creation (if corruption detected)
   - Create `.corrupt-YYYYMMDD-HHMMSS` backup
   ↓
4. Index Rebuilding
   - Parse primary database file
   - Rebuild sharded tag indexes
   - Reconstruct entity timelines
   ↓
5. Validation & Continuation
   - Verify recovered index integrity
   - Continue normal startup
```

## Implementation Details

### Detection Criteria
- Index file significantly older than database file (>5 minutes)
- Invalid magic numbers in index headers
- Corrupted offset data causing range errors
- High corruption rates during index loading

### Recovery Operations
- **Backup**: Timestamped backup files preserve corrupted state for analysis
- **Rebuilding**: Complete index reconstruction from primary database
- **Validation**: Post-recovery integrity checks ensure successful recovery
- **Logging**: Detailed operation logging for transparency and debugging

### Performance Characteristics
- **Detection**: O(1) file timestamp and header checks
- **Recovery**: O(n) where n = number of entities in database
- **Impact**: ~9 minutes for 1.2GB database with 831 entities
- **Result**: 500x performance improvement (36s → 70ms queries)

## Benefits

### Operational Excellence
- **Zero Manual Intervention**: Fully automatic recovery operation
- **Production Stability**: Eliminates authentication timeouts and user disruption
- **Performance Restoration**: Returns query times to normal levels
- **Data Integrity**: No data loss during recovery operations

### Architectural Alignment
- **Single Source of Truth**: No external tools or duplicate implementations
- **Self-Healing Database**: Automatic maintenance following EntityDB principles
- **Transparent Operations**: Full logging maintains operational visibility
- **Zen Architecture**: Simple, automatic, reliable operation

### User Experience
- **Seamless Recovery**: Users unaware of corruption/recovery process
- **Reliable Authentication**: No more "kicked out directly" issues
- **Consistent Performance**: Stable query response times
- **Production Ready**: Handles large databases efficiently

## Testing Results

### Recovery Validation
- Successfully recovered from 5+ hour index staleness
- Processed 831 entities with 28,814 tag index entries
- Recovery completed in ~9 minutes for 1.2GB database
- Zero data loss during recovery process

### Performance Verification
- **Pre-Recovery**: 36-second query times, authentication timeouts
- **Post-Recovery**: 70ms query times, clean authentication flow
- **Performance Gain**: 500x improvement in query response time
- **Stability**: No recurring corruption issues after recovery

### Production Impact
- **Authentication Success**: 100% login success rate post-recovery
- **User Experience**: Eliminated "kicked out directly" issues
- **System Stability**: Consistent performance under load
- **Operational Confidence**: Self-healing system reduces maintenance burden

## Alternatives Considered

### Manual CLI Tools
- **Rejected**: Requires external intervention, violates single source of truth
- **Issues**: Operational complexity, human error potential, availability requirements

### Scheduled Maintenance
- **Rejected**: Reactive approach, doesn't prevent user impact
- **Issues**: Downtime requirements, timing coordination complexity

### Index Validation Only
- **Rejected**: Detection without automatic recovery still requires manual intervention
- **Issues**: Partial solution, doesn't eliminate operational burden

## Consequences

### Positive Impact
- **Automatic Operation**: Zero operational overhead for index corruption
- **Performance Reliability**: Consistent query performance regardless of corruption
- **User Experience**: Seamless authentication and API interaction
- **Production Confidence**: Self-healing database reduces operational risk

### Implementation Requirements
- **Startup Time**: Additional ~1-2 seconds for corruption detection checks
- **Storage Overhead**: Backup files require additional disk space
- **Recovery Time**: Extended startup time during actual recovery operations
- **Logging Volume**: Increased log output during recovery operations

### Monitoring Considerations
- **Recovery Events**: Monitor for index corruption frequency patterns
- **Performance Metrics**: Track recovery time vs database size
- **Storage Usage**: Monitor backup file accumulation
- **Success Rates**: Validate recovery effectiveness over time

## Compliance

### EntityDB Principles
✅ **Single Source of Truth**: No external tools or duplicate implementations  
✅ **Temporal Architecture**: Maintains nanosecond timestamp precision  
✅ **Binary Storage**: Compatible with EBF format and WAL operations  
✅ **RBAC Integration**: Preserves authentication and authorization functionality  
✅ **High Performance**: Optimized recovery with concurrent index building  

### Version Compatibility
- **Introduced**: v2.32.0
- **Backward Compatibility**: Fully compatible with existing databases
- **Migration**: Automatic - no database schema changes required
- **Rollback**: Safe - recovery creates backups before modification

## Related ADRs
- [ADR-002: Binary Storage Format](002-binary-storage-format.md) - Foundation for recovery operations
- [ADR-003: Unified Sharded Indexing](003-unified-sharded-indexing.md) - Index structure being recovered
- [ADR-015: WAL Management and Checkpointing](015-wal-management-and-checkpointing.md) - Complementary durability mechanisms
- [ADR-016: Error Recovery and Resilience](016-error-recovery-and-resilience.md) - Overall recovery strategy

---
*This ADR documents the implementation of automatic index corruption recovery, ensuring EntityDB maintains its self-healing architecture principles while providing production-grade reliability and performance.*