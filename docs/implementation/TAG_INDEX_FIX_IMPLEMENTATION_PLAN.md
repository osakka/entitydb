# Tag Index Fix Implementation Plan

## Problem Summary
The persistent tag index loads correctly but WAL replay interferes with the loaded state, causing entity count mismatches.

## Root Cause
The issue is in the startup sequence:
1. `buildIndexes()` loads persistent index (12 tags)
2. `replayWAL()` processes entities and calls `updateIndexes()`
3. `updateIndexes()` has deduplication but it's not working correctly
4. Result: loaded index gets corrupted during WAL replay

## Implementation Strategy

### Option 1: Skip WAL Replay When Index Loaded (CHOSEN)
**Rationale**: If we successfully loaded a persistent index, the entities should already be indexed correctly.

### Option 2: Clear Index Before WAL Replay
**Rationale**: Always rebuild from WAL for consistency

### Option 3: Smart WAL Replay
**Rationale**: Make WAL replay aware of existing index state

## Step-by-Step Implementation

### Step 1: Fix WAL Replay Conflict
**Goal**: Prevent WAL replay from corrupting loaded persistent index

**Changes**:
1. Add flag to track if persistent index was loaded
2. Skip WAL replay if persistent index is loaded and valid
3. Only use WAL replay as fallback when no persistent index exists

**Files to modify**:
- `src/storage/binary/entity_repository.go`

### Step 2: Improve Index Health Validation
**Goal**: Better detection of index consistency issues

**Changes**:
1. Add more detailed index health checks
2. Validate entity-to-tag mappings
3. Better error reporting

### Step 3: Add Index Metadata
**Goal**: Track index freshness and validity

**Changes**:
1. Add timestamp to index file
2. Compare with entity data file modification time
3. Invalidate stale indexes

### Step 4: Enhanced Deduplication
**Goal**: Ensure deduplication works correctly for temporal tags

**Changes**:
1. Improve deduplication logic in `updateIndexes`
2. Handle temporal tag prefixes correctly
3. Add debug logging for deduplication

## Implementation Details

### Step 1 Implementation

```go
// In NewEntityRepository, track if persistent index was loaded
type EntityRepository struct {
    // ... existing fields ...
    persistentIndexLoaded bool
}

// In buildIndexes()
if tagIndex, err := LoadTagIndexV2(r.getDataFile()); err == nil {
    // ... load index ...
    r.persistentIndexLoaded = true
} else {
    r.persistentIndexLoaded = false
}

// In NewEntityRepository, conditional WAL replay
if r.persistentIndexLoaded {
    logger.Info("Persistent index loaded, skipping WAL replay")
} else {
    logger.Info("No persistent index, replaying WAL...")
    if err := r.replayWAL(); err != nil {
        logger.Error("Failed to replay WAL: %v", err)
    }
}
```

### Step 2 Implementation

```go
// Enhanced health check
func (r *EntityRepository) VerifyIndexHealth() error {
    // Count unique entities in tag index
    indexedEntities := make(map[string]int) // entity -> tag count
    for tag, entityIDs := range r.tagIndex {
        for _, id := range entityIDs {
            indexedEntities[id]++
        }
    }
    
    // Detailed reporting
    entityCount := len(r.entities)
    indexCount := len(indexedEntities)
    
    if entityCount != indexCount {
        logger.Error("Index mismatch details:")
        logger.Error("- Entities in repository: %d", entityCount)
        logger.Error("- Entities in tag index: %d", indexCount)
        
        // Find missing entities
        for entityID := range r.entities {
            if _, exists := indexedEntities[entityID]; !exists {
                logger.Error("- Entity %s not in tag index", entityID)
            }
        }
        
        return fmt.Errorf("index mismatch: %d entities but %d in tag index", entityCount, indexCount)
    }
    
    return nil
}
```

## Testing Plan

### Test 1: Basic Persistence
1. Start server
2. Create entities
3. Stop server (should save index)
4. Start server (should load index)
5. Verify entity count matches

### Test 2: Multiple Cycles
1. Repeat Test 1 multiple times
2. Verify consistency maintained

### Test 3: Mixed Operations
1. Start server
2. Create entities
3. Restart server
4. Create more entities
5. Restart again
6. Verify all entities accessible

## Success Criteria

1. **✅ Index Persistence**: Index file created on shutdown
2. **✅ Index Loading**: Index loaded on startup
3. **❌ Count Consistency**: Entity count matches after restart (CURRENT ISSUE)
4. **✅ Query Correctness**: Queries return same results before/after restart
5. **⚠️ Performance**: Startup faster than full rebuild (needs measurement)

## Risk Mitigation

1. **Fallback**: If persistent index loading fails, fall back to full rebuild
2. **Validation**: Always validate index health after loading
3. **Logging**: Comprehensive logging for debugging
4. **Recovery**: Manual reindex endpoint for emergency recovery

## Implementation Priority

### Phase 1 (Immediate): Fix WAL Replay Conflict
- **Priority**: Critical
- **Effort**: Low
- **Risk**: Low

### Phase 2 (Next): Enhanced Validation
- **Priority**: High  
- **Effort**: Medium
- **Risk**: Low

### Phase 3 (Future): Index Metadata
- **Priority**: Medium
- **Effort**: High
- **Risk**: Medium