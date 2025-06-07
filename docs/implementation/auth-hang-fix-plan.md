# Authentication Hang Fix Implementation Plan

## Problem Statement

Authentication requests hang indefinitely when the system is processing metric updates. This occurs because:

1. `ListByTag` holds a global read lock while scanning the entire tag index
2. `AddTag` (metrics) needs a write lock to update the tag index
3. No fairness mechanism - writers can be starved by readers
4. The tag index scan is O(n) with the number of tags in the system

## Root Cause

The storage layer uses a single global `sync.RWMutex` for the tag index. When metrics are being written continuously:
- Each metric update calls `AddTag` which needs a write lock
- Authentication calls `ListByTag` which holds a read lock while scanning ALL tags
- With thousands of metric tags, this scan takes significant time
- New metric writes queue up waiting for the write lock
- Authentication queries get stuck behind the queued writes

## Solution Design

### 1. Sharded Tag Index with Lock Striping

Replace the single global lock with a sharded approach:

```go
type ShardedTagIndex struct {
    shards    [256]*TagIndexShard
    shardMask uint32
}

type TagIndexShard struct {
    mu    sync.RWMutex
    tags  map[string][]string
    queue *FairQueue  // Fair request queue
}
```

Benefits:
- Reduces lock contention by 256x
- Different tags can be accessed concurrently
- Metrics and auth queries likely hit different shards

### 2. Fair Request Queue

Implement a fair queuing system to prevent starvation:

```go
type FairQueue struct {
    mu         sync.Mutex
    readers    []chan struct{}
    writers    []chan struct{}
    activeReaders int32
    writerWaiting bool
}
```

Features:
- FIFO ordering for same lock type
- Writers get priority after N reads
- Prevents reader/writer starvation

### 3. Optimized Tag Lookups

Instead of scanning all tags, use prefix-based indexing:

```go
type OptimizedTagIndex struct {
    exact     map[string][]string    // Exact tag matches
    prefixes  map[string]*Trie       // Prefix searches
    temporal  map[string]*BTree      // Temporal tags by timestamp
}
```

This reduces lookup time from O(n) to O(log n) or O(1).

### 4. Background Index Maintenance

Move expensive operations out of the critical path:

```go
type IndexMaintainer struct {
    updateChan chan IndexUpdate
    mergeChan  chan MergeRequest
}
```

- Batch index updates
- Async index merging
- Periodic cleanup

## Implementation Steps

### Phase 1: Sharded Lock Implementation

1. Create `sharded_lock.go` with the sharded index structure
2. Implement fair queuing mechanism
3. Add metrics for lock contention monitoring

### Phase 2: Integrate with EntityRepository

1. Replace `tagIndex map[string][]string` with `ShardedTagIndex`
2. Update `ListByTag` to use sharded lookups
3. Update `AddTag` to use sharded updates
4. Ensure backward compatibility

### Phase 3: Optimization

1. Implement prefix-based indexing for common queries
2. Add background index maintenance
3. Add caching for hot paths

### Phase 4: Testing

1. Unit tests for sharded index
2. Concurrency tests with high metric load
3. Benchmark authentication under load
4. Verify no regressions

## Success Criteria

1. Authentication completes in <100ms under any load
2. Fair access - no request waits >1 second
3. No performance regression for normal operations
4. System handles 1TB+ of data without degradation

## Rollback Plan

The implementation will be behind a feature flag:
```go
if UseShardedIndex {
    // New implementation
} else {
    // Existing implementation
}
```

This allows quick rollback if issues are found.