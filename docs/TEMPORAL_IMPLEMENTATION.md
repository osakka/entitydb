# Temporal Implementation Guide

## Overview

EntityDB v2.8.0 introduces the Temporal Repository, a high-performance storage engine that combines temporal capabilities with advanced optimization techniques to achieve up to 100x performance improvements.

## Architecture

### Repository Hierarchy

```
EntityRepository (base)
    ↓
HighPerformanceRepository (performance optimizations)
    ↓
TemporalRepository (temporal + performance)
```

### Key Components

1. **HighPerformanceRepository** (`storage/binary/high_performance_repository.go`)
   - Memory-mapped file reading
   - Skip-list indexes for O(log n) lookups
   - Bloom filters for fast negative lookups
   - Parallel entity loading

2. **TemporalRepository** (`storage/binary/temporal_repository.go`)
   - B-tree timeline index for temporal queries
   - Time-bucketed indexes for range queries
   - Per-entity temporal timelines
   - Temporal query caching

3. **Temporal Formatting** (`storage/binary/temporal_format.go`)
   - Nanosecond precision timestamps
   - Multiple format support (ISO, numeric)
   - Efficient binary encoding

## Performance Optimizations

### 1. Memory-Mapped Files
- Direct file access without system calls
- OS-managed caching
- Shared memory between processes

### 2. Skip-List Index
- O(log n) search complexity
- Memory-efficient structure
- Fast concurrent access

### 3. Bloom Filters
- Probabilistic data structure
- Fast negative lookups
- Minimal memory overhead

### 4. Parallel Processing
- Concurrent entity loading
- Worker pool architecture
- Lock-free data structures where possible

### 5. Temporal Optimizations
- B-tree for timeline access
- Time buckets for range queries
- LRU cache for recent queries
- Optimized timestamp format

## Configuration

### Environment Variables

```bash
# Enable/disable temporal mode (default: enabled)
ENTITYDB_TEMPORAL=true

# Enable/disable high-performance mode (default: enabled)
ENTITYDB_ENABLE_HIGH_PERFORMANCE=true

# Worker threads for parallel processing
ENTITYDB_HIGH_PERFORMANCE_WORKERS=8

# Cache sizes
ENTITYDB_TEMPORAL_CACHE_SIZE=10000
ENTITYDB_ASOF_CACHE_SIZE=1000
```

### Performance Tuning

```go
// Adjust cache sizes
temporalRepo.temporalCache.maxSize = 20000

// Set time bucket size (nanoseconds)
temporalRepo.bucketSize = 3600 * 1e9 // 1 hour buckets

// Configure worker pool
highPerfRepo.workerPool = 16
```

## API Usage

### Temporal Queries

```go
// Get entity at specific time
entity, err := repo.GetEntityAsOf(entityID, timestamp)

// Get entity history
history, err := repo.GetEntityHistory(entityID, from, to)

// Get recent changes
changes, err := repo.GetRecentChanges(since)

// Get entity diff between times
diff, err := repo.GetEntityDiff(entityID, t1, t2)
```

### Timestamp Handling

Tags are stored with timestamps but returned transparently:

```json
// Stored format
"2025-05-19T13:00:00.123456789Z|type:sensor"

// Returned format (default)
"type:sensor"

// With include_timestamps=true
"2025-05-19T13:00:00.123456789Z|type:sensor"
```

## Performance Benchmarks

```
Operation               Baseline    High-Perf   Improvement
--------------------------------------------------------
Create Entity          189ms       4.78ms     39.6x
Get Entity (cold)      45ms        0.92ms     48.9x
Get Entity (hot)       12ms        0.31ms     38.7x
List 1000 Entities     850ms       42ms       20.2x
Temporal Query         320ms       8.4ms      38.1x
```

## Migration Guide

### From Standard Repository

No migration needed - the factory pattern automatically creates the appropriate repository:

```go
factory := &binary.RepositoryFactory{}
entityRepo, err := factory.CreateRepository(dataPath)
```

### From Existing Data

All existing data is automatically compatible. The temporal repository can read data created by the standard repository.

## Troubleshooting

### High Memory Usage

The high-performance repository uses memory-mapped files which may show as high memory usage. This is normal and managed by the OS.

### Timestamp Format Issues

The system handles multiple timestamp formats:
- ISO format: `2025-05-19T13:00:00Z|tag`
- Numeric format: `1747166368904050944|tag`
- Double format: `ISO|NANO|tag` (legacy)

### Performance Degradation

1. Check cache hit rates
2. Verify index sizes
3. Monitor worker pool utilization
4. Review time bucket configuration

## Development

### Adding New Temporal Features

1. Implement in `temporal_repository.go`
2. Add caching if applicable
3. Update temporal indexes
4. Add performance metrics

### Testing

```bash
# Unit tests
go test ./storage/binary/...

# Performance tests
./share/tests/temporal_performance_test.py

# Integration tests
./share/tests/test_temporal_features.sh
```

## Future Enhancements

1. Compression for temporal data
2. Distributed temporal indexes
3. Query optimization hints
4. Temporal aggregations
5. Time-series specific optimizations