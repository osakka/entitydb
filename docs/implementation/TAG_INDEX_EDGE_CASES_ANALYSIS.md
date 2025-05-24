# Tag Index Edge Cases Analysis

## Current Problem

The persistent tag index test shows:
- **Before restart**: 11 entities with tag 'test:persistent'
- **After restart**: 1 entity with tag 'test:persistent'
- **Index file**: Created successfully (922 bytes)
- **Loading**: 12 tags loaded from persistent index
- **Health check**: Index mismatch: 11 entities but only 1 in tag index

## Root Cause Analysis

### Issue 1: Entity Count Mismatch After WAL Replay
**Symptom**: Index health check fails with mismatch between entity count and indexed entities
**Root Cause**: WAL replay is affecting the loaded tag index, but deduplication isn't working correctly

### Issue 2: Temporal vs Non-Temporal Tag Confusion
**Symptom**: Tags are stored with temporal prefixes but queries expect non-temporal tags
**Root Cause**: Mixed temporal/non-temporal indexing creates inconsistencies

### Issue 3: Index Loading vs Entity Loading Order
**Symptom**: Entities loaded from data file but tag index from persistence file may be out of sync
**Root Cause**: Time gap between index save and entity persistence to disk

## Detailed Analysis

### Data Flow Problem
1. **Entity Creation**: Entity created with temporal tags (e.g., "2025-05-24T12:00:00|test:persistent")
2. **Index Storage**: Tag index saves both temporal and non-temporal versions
3. **Index Loading**: Loads tag -> entity mappings
4. **WAL Replay**: Replays entities and tries to update already-loaded index
5. **Deduplication Failure**: Entity is already in index but deduplication doesn't detect it properly

### Logging Evidence
```
2025/05/24 12:06:41.812570 [EntityDB] INFO: [buildIndexes] Loading tag index from persistent storage...
2025/05/24 12:06:41.812582 [EntityDB] INFO: [buildIndexes] Loaded 12 tags from persistent index in 175.467µs
2025/05/24 12:06:41.813062 [EntityDB] INFO: [buildIndexes] Loaded 11 entities and built supplementary indexes
2025/05/24 12:06:41.813262 [EntityDB] INFO: [VerifyIndexHealth] Index health check: 11 entities in repository, 1 entities in tag index
```

## Edge Cases Identified

### Edge Case 1: WAL Replay Interference
- **Problem**: WAL replay runs after index loading and corrupts the loaded index
- **Impact**: Index becomes inconsistent with actual entities
- **Severity**: High

### Edge Case 2: Temporal Tag Duplication
- **Problem**: Same entity indexed multiple times under different temporal versions of same tag
- **Impact**: Query results may include duplicates
- **Severity**: Medium

### Edge Case 3: Index-Entity Sync Gap
- **Problem**: Index file may be newer/older than entity data file
- **Impact**: Index points to non-existent entities or misses existing entities
- **Severity**: Medium

### Edge Case 4: Partial Index Corruption
- **Problem**: If index loading fails partially, system falls back to rebuild but may have partial state
- **Impact**: Inconsistent system state
- **Severity**: Low

## Solution Strategy

### Phase 1: Fix WAL Replay Logic
1. Skip WAL replay if persistent index was loaded successfully
2. Or: Clear index before WAL replay and rebuild completely
3. Or: Make WAL replay aware of existing index state

### Phase 2: Improve Deduplication
1. Fix deduplication logic to handle temporal tags properly
2. Add entity ID tracking to prevent double-indexing
3. Validate index consistency after all operations

### Phase 3: Add Index Versioning
1. Add timestamp to index file to compare with entity data
2. Invalidate index if data is newer than index
3. Add checksum validation

### Phase 4: Enhanced Error Handling
1. Graceful fallback to full rebuild if index is corrupted
2. Better logging and diagnostics
3. Index repair tools

## Implementation Plan

### Step 1: Immediate Fix (High Priority)
- Fix the WAL replay vs loaded index conflict
- Ensure consistent entity counting

### Step 2: Robustness (Medium Priority)  
- Add index validation and repair
- Improve temporal tag handling

### Step 3: Optimization (Low Priority)
- Add incremental updates
- Reduce memory usage during rebuild

## Success Criteria

1. **Consistency**: Entity count matches index count after restart
2. **Reliability**: Index survives multiple restart cycles
3. **Performance**: Startup time improved vs full rebuild
4. **Correctness**: Query results identical before and after restart