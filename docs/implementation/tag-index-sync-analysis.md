# Tag Index Synchronization Analysis

## Systematic Issue Overview

The tag index synchronization issue is a **systematic problem** that affects the EntityDB storage layer when entities exist in the Write-Ahead Log (WAL) but haven't been properly indexed in memory. This creates a situation where operations like `AddTag` fail because they can't find entities that actually exist.

## Root Cause Analysis

### 1. Index Building Process (buildIndexes)

The `buildIndexes` function (lines 197-293) rebuilds the tag index from the binary data file (`entities.ebf`) but:
- Only reads entities from the persisted binary file
- Does NOT include entities that only exist in the WAL
- Sets `persistentIndexLoaded = false` to trigger WAL replay

### 2. WAL Replay Issues

The `replayWAL` function (lines 2438-2522) has several problems:
- Updates the in-memory `entities` map and `tagIndex`
- But the temporal index and namespace index updates are commented out (TODO)
- No handling for sharded index updates during replay
- Can result in partial index state

### 3. Metric Entity Creation Pattern

The issue frequently occurs with metric entities because:
1. `storeCheckpointMetric` creates new metric entities or updates existing ones
2. These operations use `AddTag` which requires `GetByID` to succeed
3. If the entity was created in WAL but not yet persisted/indexed, `GetByID` fails
4. This triggers "Entity X not found for AddTag" errors

### 4. Checkpoint Race Condition

The checkpoint process has a timing issue:
1. WAL entries are created for new entities
2. Before checkpoint completes, other operations try to access these entities
3. The entities exist in WAL but not in the indexed memory structures
4. Operations fail with "entity not found" errors

## Impact Scope

This is a **pervasive issue** that affects:
- All metric collection operations
- Any entity created during high write load
- Operations that occur between WAL write and checkpoint
- Tag-based queries that rely on the index

## Evidence of Systematic Nature

1. **Repeated Pattern**: The same error occurs for multiple metric entities:
   - `metric_wal_checkpoint_success_total`
   - `metric_wal_checkpoint_duration_ms`
   - `metric_wal_checkpoint_size_reduction_bytes`

2. **Timing Dependent**: The issue is more likely to occur:
   - During checkpoint operations
   - Under high write load
   - When creating entities that are immediately accessed

3. **Index Inconsistency**: The tag index can become permanently out of sync until:
   - A full server restart (triggers `buildIndexes`)
   - Manual reindexing
   - Successful checkpoint that persists WAL entries

## Recommended Fix Strategy

### Short-term Fix
1. Modify `GetByID` to check WAL if entity not found in memory
2. Update `replayWAL` to properly update all indexes (temporal, namespace, sharded)
3. Add synchronization to ensure WAL replay completes before serving requests

### Long-term Fix
1. Implement a unified index that includes both persisted and WAL entities
2. Add index consistency checks during checkpoint operations
3. Implement atomic index updates that can't leave partial state
4. Add metrics to track index health and divergence

## Code Locations Requiring Changes

1. **entity_repository.go:GetByID** (line ~550)
   - Add WAL lookup fallback

2. **entity_repository.go:replayWAL** (line ~2438)
   - Complete temporal and namespace index updates
   - Add sharded index support

3. **entity_repository.go:buildIndexes** (line ~197)
   - Consider including WAL entries in initial build

4. **entity_repository.go:persistWALEntries** (line ~2524)
   - Ensure indexes are updated atomically with persistence

## Testing Approach

1. Create rapid entity creation/access patterns
2. Trigger checkpoints during entity operations  
3. Verify all entities are accessible immediately after creation
4. Monitor index consistency metrics

## Conclusion

This is not an isolated issue but a fundamental synchronization problem in the storage layer. It requires a comprehensive fix that ensures the tag index always reflects the true state of all entities, whether they're in the binary file or only in the WAL.