# Tag Index Synchronization Fix

## Problem Summary

The tag index in `entity_repository.go` was not properly synchronized with entity data, causing:
- Authentication failures due to "entity not found in index" errors
- Entities created via Create() not immediately findable by tags
- WAL replay not properly rebuilding temporal tag indexes
- Race conditions between index updates and entity writes

## Root Causes

1. **Non-atomic index updates**: Indexes were updated AFTER writing entities, creating a window where entities existed on disk but not in indexes
2. **Incomplete WAL replay**: Temporal tags were not properly indexed during WAL replay (only timestamped versions were indexed, not the actual tag values)
3. **Missing index corruption detection**: No automatic recovery when indexes became out of sync
4. **Inconsistent temporal tag handling**: Some code paths didn't index both timestamped and non-timestamped versions of tags

## Implementation Details

### 1. Atomic Index Updates in Create()

```go
// CRITICAL: Update indexes BEFORE writing to ensure atomicity
r.mu.Lock()
r.updateIndexes(entity)
r.entities[entity.ID] = entity
r.mu.Unlock()

// Write entity using WriterManager
if err := r.writerManager.WriteEntity(entity); err != nil {
    // Rollback index changes on write failure
    r.mu.Lock()
    delete(r.entities, entity.ID)
    // Remove from all indexes
    for _, tag := range entity.Tags {
        r.removeFromTagIndex(tag, entity.ID)
        // Also remove non-timestamped version
        if strings.Contains(tag, "|") {
            parts := strings.SplitN(tag, "|", 2)
            if len(parts) == 2 {
                r.removeFromTagIndex(parts[1], entity.ID)
            }
        }
    }
    r.mu.Unlock()
    return err
}
```

### 2. Enhanced updateIndexes() Method

- Now removes all existing index entries before re-indexing (prevents duplicates)
- Properly handles both timestamped and non-timestamped versions of temporal tags
- Thread-safe with proper mutex usage

### 3. Fixed WAL Replay

```go
// In replayWAL(), now uses updateIndexes() for consistency
case WALOpCreate, WALOpUpdate:
    if entry.Entity != nil {
        r.mu.Lock()
        r.entities[entry.EntityID] = entry.Entity
        r.updateIndexes(entry.Entity)
        r.mu.Unlock()
        entitiesReplayed++
    }
```

### 4. Automatic Index Repair

- Added `RepairIndexes()` method that rebuilds all indexes from in-memory entities
- Called automatically when index health check fails during initialization
- Can be triggered manually for maintenance

### 5. Index Corruption Detection

- Added `detectAndFixIndexCorruption()` method called when entity not found in index
- Automatically re-indexes entities found on disk but missing from indexes
- Logs warnings when corruption is detected and fixed

### 6. Always Replay WAL

- Changed initialization to ALWAYS replay WAL, not just when persistent index is missing
- Ensures any operations in WAL but not yet persisted are properly indexed

## Testing

A comprehensive test script `/opt/entitydb/test-tag-index-sync.sh` verifies:
- Entities are immediately findable after creation
- Tags are properly indexed for queries
- Concurrent operations maintain consistency
- New tags added via update are immediately indexed

## Impact

This fix resolves:
- Authentication failures where valid credentials couldn't be found
- "Entity not found" errors immediately after creation
- Inconsistent query results
- Index corruption after server restarts

## Performance Considerations

- Index updates are now atomic with writes (slight overhead)
- Automatic corruption detection adds minimal overhead to cache misses
- WAL replay ensures consistency at startup
- RepairIndexes() can be expensive but only runs when corruption detected

## Future Improvements

1. Consider persistent index format that includes checksums
2. Add background index verification process
3. Implement index compaction for long-running servers
4. Add metrics for index health and repair operations