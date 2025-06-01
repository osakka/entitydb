# Tag Index Persistence Fix Plan

## Problem Statement
When EntityDB restarts, the tag index is lost because:
1. Tag index is only built from the binary data file (.ebf)
2. WAL entries are not replayed to rebuild the index
3. No health check verifies index integrity

## Root Cause Analysis
1. `NewEntityRepository` calls `buildIndexes()` which only reads from the .ebf file
2. WAL replay is not triggered during startup
3. If entities were only in WAL (not yet persisted to .ebf), they're lost from index
4. No mechanism to detect or repair corrupted/missing indexes

## Fix Implementation Plan

### Phase 1: WAL Replay Integration
1. Modify `NewEntityRepository` to replay WAL after building indexes
2. During WAL replay, rebuild tag indexes for each entity operation
3. Ensure CREATE and UPDATE operations add to tag index
4. Ensure DELETE operations remove from tag index

### Phase 2: Index Health Check
1. Add index verification on startup
2. Compare entity count with tag index entries
3. Log warnings if mismatches are found
4. Auto-trigger rebuild if index is unhealthy

### Phase 3: Manual Recovery Tools
1. Add `/api/v1/admin/reindex` endpoint for manual index rebuild
2. Add CLI tool for offline index repair
3. Add index statistics to health endpoint

### Phase 4: Index Persistence (Future)
1. Consider persisting tag index to .idx file
2. Load index from file on startup for faster recovery
3. Keep index file in sync with operations

## Implementation Steps

### Step 1: Add WAL Replay to Constructor
```go
// In NewEntityRepository after buildIndexes()
if err := repo.replayWAL(); err != nil {
    logger.Warn("Failed to replay WAL: %v", err)
}
```

### Step 2: Implement replayWAL Method
```go
func (r *EntityRepository) replayWAL() error {
    return r.wal.Replay(func(entry WALEntry) error {
        switch entry.OpType {
        case WALOpCreate, WALOpUpdate:
            // Rebuild tag index for this entity
            r.updateTagIndex(entry.EntityID, entry.Entity)
        case WALOpDelete:
            // Remove from tag index
            r.removeFromTagIndex(entry.EntityID)
        }
        return nil
    })
}
```

### Step 3: Add Index Health Check
```go
func (r *EntityRepository) verifyIndexHealth() error {
    // Count entities in repository
    // Count unique entities in tag index
    // Log mismatches
    // Return error if severely corrupted
}
```

### Step 4: Add Reindex Endpoint
```go
func (h *AdminHandler) ReindexHandler(w http.ResponseWriter, r *http.Request) {
    // Require admin permission
    // Trigger full reindex
    // Return statistics
}
```

## Testing Plan
1. Create entities
2. Restart EntityDB
3. Verify all entities are queryable by tags
4. Test MetDataspace dashboard shows all metrics
5. Test manual reindex endpoint

## Rollout Strategy
1. Implement and test locally
2. Add comprehensive logging
3. Deploy with monitoring
4. Document recovery procedures