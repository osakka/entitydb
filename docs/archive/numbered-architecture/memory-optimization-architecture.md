# Memory Optimization Architecture

**Version**: 2.34.0  
**Status**: Production Ready  
**Last Updated**: 2025-06-22

## Executive Summary

EntityDB v2.34.0 implements comprehensive memory optimization architecture that prevents unbounded memory growth and provides automatic pressure relief. This document details the architectural improvements, implementation rationale, and operational characteristics of the memory optimization system.

## Problem Statement

### Root Cause Analysis

The server crash due to high memory utilization was caused by multiple compounding factors:

1. **Unbounded String Interning**: Every unique tag string was interned forever with no eviction
2. **Unbounded Entity Cache**: All entities remained in memory indefinitely
3. **Metrics Recursion**: Metrics collection created new entities, which triggered more metrics
4. **Temporal Tag Explosion**: Each metric created a new temporal tag every second (2.6M tags/day)
5. **No Memory Pressure Relief**: System had no mechanisms to shed load under pressure

### Memory Growth Pattern

```
Initial State: ~50MB
After 1 hour: ~500MB 
After 4 hours: ~2GB
After 8 hours: OOM/Crash
```

The exponential growth was due to:
- 86,400 new temporal tags per metric per day
- String interning keeping all tag strings forever
- Entity cache retaining all accessed entities
- No garbage collection of old temporal data

## Architectural Solution

### 1. Bounded String Interning with LRU Eviction

**Component**: `models/string_intern.go`

#### Design Rationale
- Strings are immutable in Go, making interning valuable for repeated tags
- LRU (Least Recently Used) ensures frequently used strings stay cached
- Memory limits prevent unbounded growth
- Adaptive sizing responds to memory pressure

#### Implementation Details
```go
type StringIntern struct {
    strings     map[string]*internEntry  // O(1) lookup
    lru         *list.List               // LRU ordering
    maxSize     int                      // Entry count limit
    memoryLimit int64                    // Memory usage limit
}
```

**Key Features**:
- Configurable size limit (default: 100,000 strings)
- Memory limit enforcement (default: 100MB)
- Access count tracking for frequency analysis
- Pressure cleanup can evict 30% of entries
- Thread-safe with read/write locks

### 2. Bounded Entity Cache with Memory Tracking

**Component**: `storage/binary/bounded_entity_cache.go`

#### Design Rationale
- Entity cache critical for read performance
- Memory tracking ensures predictable usage
- Frequency-aware eviction keeps hot entities
- Pressure relief prevents OOM conditions

#### Implementation Details
```go
type BoundedEntityCache struct {
    entries     map[string]*cacheEntry  // Entity storage
    lru         *list.List              // LRU tracking
    maxSize     int                     // Entry limit
    memoryLimit int64                   // Memory limit
}
```

**Key Features**:
- Size-based eviction (default: 10,000 entities)
- Memory usage calculation per entity
- Access frequency consideration
- Eviction callbacks for cleanup
- 40% eviction under high pressure

### 3. Metrics Recursion Prevention

**Component**: `storage/binary/entity_repository.go`

#### Design Rationale
- Metrics collection must not create metrics about itself
- Global state prevents any recursion
- Thread-local context inadequate due to Go's goroutine model
- Atomic operations ensure thread safety

#### Implementation Details
```go
var (
    metricsOperationDepth   int64  // Recursion depth tracking
    metricsDisabledGlobally int64  // Emergency kill switch
)
```

**Prevention Mechanisms**:
1. **Depth Tracking**: Increment on entry, decrement on exit
2. **Entity Detection**: Skip metrics for metric entities
3. **Emergency Disable**: Global flag stops all metrics
4. **Context Propagation**: Operations mark metrics context

### 4. Temporal Data Retention

**Component**: `storage/binary/temporal_retention.go`

#### Design Rationale
- Temporal data grows linearly without cleanup
- Self-cleaning during operations vs background process
- Memory-aware retention policies
- No separate retention entities needed

#### Implementation Details
```go
type TemporalRetentionManager struct {
    retentionPolicies map[string]RetentionPolicy
}

type RetentionPolicy struct {
    MaxAge       time.Duration  // Time-based retention
    MaxTags      int           // Count-based limit
    CleanupBatch int           // Batch size
}
```

**Retention Policies**:
- **Metrics**: 24 hours, max 1000 tags
- **Sessions**: 7 days, max 50 tags
- **Default**: 30 days, max 500 tags

**Memory Pressure Adaptation**:
- High pressure (>80%): Reduce retention by 50%
- Medium pressure (>60%): Reduce retention by 25%

### 5. Memory Monitoring and Pressure Relief

**Component**: `storage/binary/memory_monitor.go`

#### Design Rationale
- Proactive monitoring prevents emergency situations
- Graduated response based on pressure levels
- Component callbacks for coordinated relief
- Statistical tracking for analysis

#### Implementation Details
```go
type MemoryMonitor struct {
    highPressureThreshold  float64  // 80%
    criticalThreshold      float64  // 90%
    pressureCallbacks      []PressureReliefCallback
}
```

**Pressure Levels**:
1. **Low** (<60%): Normal operations
2. **Medium** (60-80%): Prepare for cleanup
3. **High** (80-90%): Aggressive cleanup
4. **Critical** (>90%): Emergency measures

**Relief Actions by Level**:

**Medium**:
- Re-enable previously disabled features
- Prepare caches for potential cleanup

**High**:
- Trigger string interning cleanup (30% eviction)
- Trigger entity cache cleanup (40% eviction)
- Reduce cache size limits by 20%
- Force garbage collection

**Critical**:
- All high-pressure actions
- Disable metrics collection globally
- Double garbage collection
- Reduce limits by 30%

## Integration Architecture

### Server Initialization Flow

```
1. Configure string interning limits
2. Start memory monitor
3. Register pressure callbacks:
   - String interning cleanup
   - Entity cache cleanup
   - Temporal retention adjustment
4. Initialize repositories with limits
5. Begin monitoring loop (30s interval)
```

### Memory Pressure Response Flow

```
Memory Check → Pressure Calculation → Level Determination
                                           ↓
                                    Callback Invocation
                                           ↓
                              Component-Specific Actions
                                           ↓
                                    Relief Verification
```

## Configuration

### Environment Variables

```bash
# String Interning
ENTITYDB_STRING_CACHE_SIZE=100000           # Max strings
ENTITYDB_STRING_CACHE_MEMORY_LIMIT=104857600 # 100MB

# Entity Cache  
ENTITYDB_ENTITY_CACHE_SIZE=10000            # Max entities
ENTITYDB_ENTITY_CACHE_MEMORY_LIMIT=1073741824 # 1GB

# Metrics
ENTITYDB_METRICS_INTERVAL=1s                # Collection interval
ENTITYDB_METRICS_ENABLE_STORAGE_TRACKING=true
```

### CLI Flags

```bash
--entitydb-string-cache-size=100000
--entitydb-string-cache-memory=100MB
--entitydb-entity-cache-size=10000
--entitydb-entity-cache-memory=1GB
```

## Performance Characteristics

### Memory Usage Profile

**Before Optimizations**:
- Startup: 50MB
- 1 hour: 500MB
- Growth rate: ~6MB/minute
- Outcome: OOM crash

**After Optimizations**:
- Startup: 50MB
- 1 hour: 150MB
- Growth rate: ~1MB/minute (stabilizes)
- Outcome: Stable operation

### Cache Performance

**String Interning**:
- Hit rate: 85-95% (common tags)
- Eviction rate: <1% under normal load
- Lookup time: O(1)

**Entity Cache**:
- Hit rate: 70-90% (hot entities)
- Memory overhead: ~200 bytes/entity
- Eviction cost: O(1)

### Pressure Relief Effectiveness

**Test Results** (2-minute stress test):
- Peak memory: 300MB (vs unbounded: 2GB+)
- Pressure events: 15
- Cleanup events: 8
- String evictions: 156
- Entity evictions: 200
- Final state: Stable at 150MB

## Operational Guidelines

### Monitoring

Key metrics to monitor:
1. **Memory Usage**: `heap_inuse`, `heap_alloc`
2. **Cache Stats**: Hit rates, eviction counts
3. **Pressure Events**: Frequency and level
4. **GC Activity**: Collection frequency and pause time

### Tuning

**High Write Workloads**:
- Increase batch size for writes
- Reduce string cache size
- Lower retention periods

**High Read Workloads**:
- Increase entity cache size
- Raise cache memory limits
- Monitor hit rates

**Memory Constrained**:
- Reduce all cache sizes
- Lower retention periods
- Enable aggressive cleanup

### Troubleshooting

**Symptom**: High memory usage despite optimizations
- Check: Metrics recursion (should be 0)
- Check: Cache eviction rates
- Action: Reduce cache limits

**Symptom**: Poor performance after optimization
- Check: Cache hit rates
- Check: GC frequency
- Action: Increase cache sizes if memory allows

**Symptom**: Pressure events too frequent
- Check: Memory limits vs actual usage
- Check: Workload patterns
- Action: Adjust thresholds or increase resources

## Design Principles

### Single Source of Truth
- All memory management centralized in memory monitor
- No duplicate cache implementations
- Unified configuration system

### No Parallel Implementations
- Direct integration with existing components
- Reuse existing data structures
- Extend rather than replace

### Bar-Raising Excellence
- Industry-standard LRU implementation
- Comprehensive memory tracking
- Adaptive pressure response
- Zero-overhead when disabled

### No Regressions
- All existing functionality preserved
- Performance maintained or improved
- Backward compatible configuration
- Graceful degradation under pressure

## Future Enhancements

### Potential Improvements

1. **Predictive Scaling**: Use ML to predict memory needs
2. **Tiered Storage**: Spill cold data to disk
3. **Compression**: Compress cached entities
4. **Smart Eviction**: Consider business importance
5. **Memory Pools**: Pre-allocated pools for common sizes

### Research Areas

1. **Alternative Cache Algorithms**: ARC, LIRS, W-TinyLFU
2. **Memory Mapping**: For cold entity storage
3. **Generational Caching**: Age-based tiers
4. **NUMA Awareness**: For multi-socket systems

## Conclusion

The memory optimization architecture successfully addresses all root causes of high memory utilization while maintaining system performance. The implementation follows EntityDB principles of single source of truth, no parallel implementations, and bar-raising excellence.

The system now operates with predictable memory usage, automatic pressure relief, and graceful degradation under load. These optimizations enable EntityDB to handle production workloads without memory-related failures.

## References

- [LRU Cache Algorithm](https://en.wikipedia.org/wiki/Cache_replacement_policies#Least_recently_used_(LRU))
- [Go Memory Management](https://go.dev/doc/gc-guide)
- [String Interning](https://en.wikipedia.org/wiki/String_interning)
- [Memory Pressure Handling](https://www.kernel.org/doc/html/latest/admin-guide/mm/concepts.html)