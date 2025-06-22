# ADR-029: Memory Optimization Strategy

**Status**: Implemented  
**Date**: 2025-06-22  
**Version**: 2.34.0

## Context

EntityDB experienced server crashes due to unbounded memory growth. Root cause analysis revealed:

1. Unbounded string interning cached every unique tag forever
2. Entity cache had no eviction policy
3. Metrics collection created recursive loops
4. Temporal tags grew at 86,400/day per metric
5. No memory pressure relief mechanisms

The system needed comprehensive memory management while maintaining performance.

## Decision

Implement a multi-layered memory optimization strategy:

1. **Bounded Caches with LRU Eviction**
   - String interning limited to 100k entries / 100MB
   - Entity cache limited to 10k entries / 1GB
   - LRU ensures hot data stays cached

2. **Metrics Recursion Prevention**
   - Global atomic counters track operation depth
   - Entity type detection prevents metric entities creating metrics
   - Emergency kill switch for critical situations

3. **Temporal Data Self-Cleaning**
   - Retention policies by entity type
   - Memory-pressure aware cleanup
   - No separate background processes

4. **Active Memory Monitoring**
   - 30-second monitoring interval
   - Graduated pressure levels (Low/Medium/High/Critical)
   - Component callbacks for coordinated relief

5. **Pressure Relief Actions**
   - High (80%): Evict 30-40% of cached data
   - Critical (90%): Disable metrics, force GC, reduce limits

## Rationale

### Why LRU Over Other Algorithms?

**Considered**: FIFO, LFU, ARC, W-TinyLFU

**Chose LRU** because:
- Simple, proven algorithm with O(1) operations
- Well-suited for temporal access patterns
- Low memory overhead
- Easy to implement correctly

### Why Global Recursion Prevention?

**Considered**: Thread-local storage, context propagation, goroutine IDs

**Chose Global Atomics** because:
- Go's goroutine model makes thread-local unreliable
- Context propagation too invasive
- Global state simple and effective
- Atomic operations ensure thread safety

### Why Self-Cleaning vs Background Processes?

**Considered**: Separate GC process, time-based cleanup, external retention service

**Chose Self-Cleaning** because:
- Eliminates separate process complexity
- Naturally rate-limited by operations
- Memory pressure triggers immediate action
- Follows "single source of truth" principle

### Why 80%/90% Pressure Thresholds?

**Considered**: 70%/85%, 85%/95%, dynamic thresholds

**Chose 80%/90%** because:
- 80% allows time for graceful cleanup
- 90% leaves emergency buffer
- Industry standard for memory pressure
- Proven in production systems

## Consequences

### Positive

1. **Predictable Memory Usage**: Growth bounded by configuration
2. **Automatic Recovery**: System self-heals under pressure
3. **Maintained Performance**: Hot data stays cached
4. **No Manual Intervention**: Fully automated management
5. **Graceful Degradation**: Features disable vs crash

### Negative

1. **Cache Misses**: Eviction causes some performance loss
2. **Configuration Complexity**: More tuning parameters
3. **Memory Overhead**: LRU tracking adds ~100 bytes/entry
4. **GC Pressure**: Cleanup can trigger GC pauses

### Neutral

1. **Monitoring Required**: Must watch pressure metrics
2. **Workload Dependent**: Tuning varies by usage
3. **Feature Coupling**: Components must support callbacks

## Implementation

### Phase 1: Bounded Caches (Complete)
- String interning with LRU
- Entity cache with memory limits
- Configuration via env/flags

### Phase 2: Recursion Prevention (Complete)
- Global operation tracking
- Entity type detection
- Emergency disable

### Phase 3: Retention Management (Complete)
- Policy-based cleanup
- Memory-aware retention
- Self-cleaning operations

### Phase 4: Memory Monitoring (Complete)
- Pressure detection
- Callback system
- Automatic relief

## Validation

### Test Results
- Memory growth: 6MB/min → 1MB/min (stabilizes)
- Stress test: Stable at 300MB (vs 2GB+ crash)
- Concurrent safety: All tests pass
- No memory leaks detected

### Production Metrics
- Uptime: Continuous operation achieved
- Memory usage: Stable within configured limits
- Performance: <5% degradation under pressure
- Recovery: Automatic from pressure events

## Related Decisions

- [ADR-007: Temporal Storage Architecture](./ADR-007-temporal-storage-architecture.md)
- [ADR-028: WAL Corruption Prevention](./ADR-028-wal-corruption-prevention.md)
- [ADR-027: Database File Unification](./ADR-027-database-file-unification.md)

## References

- [Go Memory Management](https://go.dev/doc/gc-guide)
- [LRU Implementation](https://github.com/hashicorp/golang-lru)
- [Memory Pressure Handling](https://www.kernel.org/doc/html/latest/admin-guide/mm/concepts.html)