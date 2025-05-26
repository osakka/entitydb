# EntityDB Data Integrity and Observability Plan

## Executive Summary

This document outlines a comprehensive plan to ensure all read/write operations in EntityDB are fully observable, traceable, and verifiable. The goal is to make it impossible for data corruption or inconsistencies to occur without detection.

## Current Issues Identified

1. **Index Corruption**: EOF errors when reading index entries
2. **Missing Entities**: Entities created but not findable in index
3. **Index Mismatch**: Repository has 71 entities, index only tracks 57
4. **Silent Failures**: Operations fail without adequate logging
5. **State Inconsistency**: WAL, index, and data file can become out of sync

## Core Principles

1. **Every Operation Logged**: No data operation should occur without a log entry
2. **Checksums Everywhere**: All data should be checksummed at write and verified at read
3. **Atomic Operations**: All multi-step operations must be atomic
4. **Fail-Fast**: Any inconsistency should immediately halt operations
5. **Self-Healing**: System should detect and repair inconsistencies when possible

## Detailed Implementation Plan

### Phase 1: Comprehensive Logging Infrastructure

#### 1.1 Operation Tracking
- Add unique operation IDs to track operations across components
- Log entry and exit of every data operation
- Include timing information for performance tracking
- Log all parameters and return values

#### 1.2 Data Flow Logging
- Log when data enters the system (API layer)
- Log when data is transformed (business logic)
- Log when data is persisted (storage layer)
- Log when data is retrieved
- Log when data leaves the system (API response)

#### 1.3 State Change Logging
- Before state: log current state before changes
- Operation: log what change is being made
- After state: log resulting state
- Verification: log that change was successful

### Phase 2: Write Path Integrity

#### 2.1 Pre-Write Validation
- Validate entity structure
- Verify all required fields
- Check data size limits
- Validate tag formats
- Log validation results

#### 2.2 Write Operation Tracking
- Generate write transaction ID
- Log intent to write
- Calculate checksums
- Log data size and offset
- Log index updates
- Log WAL entries
- Verify write completion

#### 2.3 Post-Write Verification
- Read back written data
- Verify checksum matches
- Verify index entry created
- Verify WAL entry written
- Log verification results

### Phase 3: Read Path Integrity

#### 3.1 Pre-Read Validation
- Verify index entry exists
- Check offset and size validity
- Verify file size sufficient
- Log validation results

#### 3.2 Read Operation Tracking
- Generate read transaction ID
- Log read intent
- Log actual bytes read
- Calculate and verify checksum
- Log any discrepancies

#### 3.3 Index Integrity
- Verify index completeness on startup
- Track index operations separately
- Log all index modifications
- Implement index rebuild capability

### Phase 4: Transaction and Consistency

#### 4.1 WAL Integration
- Log all WAL operations
- Implement WAL replay verification
- Track WAL-to-data synchronization
- Log any replay failures

#### 4.2 Multi-File Consistency
- Implement cross-file transaction IDs
- Ensure atomic updates across files
- Log coordination between components
- Implement rollback on partial failure

#### 4.3 Checkpoint and Recovery
- Regular consistency checkpoints
- Log checkpoint operations
- Implement recovery from checkpoints
- Verify recovery success

### Phase 5: Monitoring and Alerting

#### 5.1 Health Checks
- Continuous index verification
- Data file integrity checks
- WAL consistency verification
- Tag index validation

#### 5.2 Metrics Collection
- Operation success/failure rates
- Read/write latencies
- Data corruption detection rate
- Index hit/miss rates

#### 5.3 Alert Conditions
- Any checksum mismatch
- Any EOF during read
- Any index inconsistency
- Any WAL replay failure

## Implementation Components

### 1. Storage Layer Changes

#### Writer (writer.go)
- Add pre-write validation
- Add checksum calculation
- Add post-write verification
- Add comprehensive logging

#### Reader (reader.go)
- Add bounds checking
- Add checksum verification
- Add read retry logic
- Add detailed error logging

#### Index Management
- Add index transaction log
- Add index verification
- Add index rebuild capability
- Add index checksum

### 2. Repository Layer Changes

#### Entity Repository
- Add operation tracking
- Add transaction support
- Add consistency verification
- Add detailed logging

#### WAL Management
- Add WAL verification
- Add replay tracking
- Add corruption detection
- Add recovery logging

### 3. Monitoring Infrastructure

#### Health Endpoint
- Real-time integrity status
- Recent operation history
- Error rate tracking
- Performance metrics

#### Debug Endpoints
- Force integrity check
- Dump internal state
- Replay specific operations
- Manual recovery triggers

## Success Criteria

1. **Zero Silent Failures**: Every failure produces a log entry
2. **Complete Traceability**: Can trace any data from entry to storage
3. **Instant Detection**: Corruption detected immediately
4. **Full Recovery**: Can recover from any single-point failure
5. **Performance Impact**: Less than 5% overhead from integrity checks

## Risk Mitigation

1. **Performance**: Use async logging where possible
2. **Storage**: Implement log rotation and cleanup
3. **Complexity**: Phase implementation to maintain stability
4. **Compatibility**: Maintain backward compatibility

## Timeline

- Phase 1: 2 days - Logging infrastructure
- Phase 2: 3 days - Write path integrity
- Phase 3: 3 days - Read path integrity
- Phase 4: 4 days - Transaction consistency
- Phase 5: 2 days - Monitoring
- Testing: 2 days - Comprehensive testing
- Total: 16 days

## Next Steps

1. Review and approve plan
2. Create detailed action items
3. Begin Phase 1 implementation
4. Regular progress updates
5. Continuous testing throughout