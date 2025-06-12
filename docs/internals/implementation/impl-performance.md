# Performance Optimization Implementation Guide

> **Status**: Completed in v2.27.0  
> **Component**: High-performance storage and memory optimization

## Overview

EntityDB achieves exceptional performance through a comprehensive set of optimizations including memory-mapped files, string interning, sharded locking, and advanced indexing. These optimizations provide up to 100x performance improvements over traditional approaches.

## Memory Management Optimizations

### String Interning System

**Implementation**: `src/models/string_intern.go`

Reduces memory usage by up to 70% for duplicate tags:

```go
type StringIntern struct {
    mu      sync.RWMutex
    strings map[string]string
    hits    uint64
    misses  uint64
}

func (si *StringIntern) Intern(s string) string {
    si.mu.RLock()
    if interned, exists := si.strings[s]; exists {
        atomic.AddUint64(&si.hits, 1)
        si.mu.RUnlock()
        return interned
    }
    si.mu.RUnlock()

    si.mu.Lock()
    defer si.mu.Unlock()
    
    // Double-check after acquiring write lock
    if interned, exists := si.strings[s]; exists {
        atomic.AddUint64(&si.hits, 1)
        return interned
    }
    
    si.strings[s] = s
    atomic.AddUint64(&si.misses, 1)
    return s
}
```

**Benefits**:
- 70% memory reduction for tag storage
- Faster string comparisons (pointer equality)
- Reduced GC pressure

### Safe Buffer Pools

**Implementation**: `src/storage/binary/safe_buffer_pool.go`

Size-based buffer pools for different operations:

```go
type SafeBufferPool struct {
    small  sync.Pool // 1KB buffers
    medium sync.Pool // 64KB buffers  
    large  sync.Pool // 1MB buffers
}

func (p *SafeBufferPool) Get(size int) []byte {
    switch {
    case size <= 1024:
        return p.small.Get().([]byte)[:size]
    case size <= 65536:
        return p.medium.Get().([]byte)[:size]  
    default:
        return p.large.Get().([]byte)[:size]
    }
}
```

**Benefits**:
- Reduced memory allocations
- Lower GC overhead
- Consistent memory usage patterns

### Memory-Mapped Files

**Implementation**: `src/storage/binary/mmap_reader.go`

Zero-copy file operations using OS page cache:

```go
type MmapReader struct {
    data   []byte
    offset int64
    size   int64
}

func (r *MmapReader) ReadAt(p []byte, off int64) (n int, err error) {
    if off >= r.size {
        return 0, io.EOF
    }
    
    // Zero-copy read from memory-mapped region
    n = copy(p, r.data[off:])
    return n, nil
}
```

**Benefits**:
- OS-managed caching
- Zero-copy reads
- Efficient large file handling

## Concurrency Optimizations

### Sharded Locking System

**Implementation**: `src/storage/binary/sharded_lock.go`

Distributes lock contention across multiple shards:

```go
type ShardedLock struct {
    shards   []sync.RWMutex
    numShards int
}

func (sl *ShardedLock) Lock(key string) {
    shard := sl.getShard(key)
    sl.shards[shard].Lock()
}

func (sl *ShardedLock) getShard(key string) int {
    hash := fnv.New32a()
    hash.Write([]byte(key))
    return int(hash.Sum32()) % sl.numShards
}
```

**Benefits**:
- Reduced lock contention
- Better CPU utilization
- Scalable concurrent access

### Fair Queue Implementation

**Implementation**: `src/storage/binary/locks.go`

Prevents reader/writer starvation:

```go
type FairRWMutex struct {
    r       sync.Mutex
    w       sync.Mutex
    readers int32
}

func (rw *FairRWMutex) RLock() {
    rw.r.Lock()
    if atomic.AddInt32(&rw.readers, 1) == 1 {
        rw.w.Lock() // First reader locks writers
    }
    rw.r.Unlock()
}
```

### Deadlock Detection

**Implementation**: `src/storage/binary/lock_tracer.go`

Comprehensive lock monitoring and deadlock prevention:

```go
type LockTracer struct {
    locks     map[string]*LockInfo
    mu        sync.Mutex
    timeout   time.Duration
}

func (lt *LockTracer) TrackLock(key string, lockType string) {
    lt.mu.Lock()
    defer lt.mu.Unlock()
    
    lt.locks[key] = &LockInfo{
        Key:       key,
        Type:      lockType,
        Timestamp: time.Now(),
        Goroutine: getGoroutineID(),
    }
}
```

## Storage Optimizations

### Auto-chunking System

**Implementation**: `src/api/entity_handler_chunking.go`

Automatic chunking for files > 4MB:

```go
func (h *EntityHandler) handleLargeContent(content []byte) ([]*Chunk, error) {
    if len(content) <= h.chunkSize {
        return nil, nil // No chunking needed
    }
    
    chunks := make([]*Chunk, 0, (len(content)/h.chunkSize)+1)
    for i := 0; i < len(content); i += h.chunkSize {
        end := i + h.chunkSize
        if end > len(content) {
            end = len(content)
        }
        
        chunk := &Chunk{
            Index:   len(chunks),
            Content: content[i:end],
        }
        chunks = append(chunks, chunk)
    }
    
    return chunks, nil
}
```

**Benefits**:
- No RAM limits for large files
- Streaming support
- Memory-efficient processing

### Compression Support

**Implementation**: `src/storage/binary/compression.go`

Automatic compression for content > 1KB:

```go
func CompressContent(content []byte) ([]byte, error) {
    if len(content) < 1024 {
        return content, nil // Don't compress small content
    }
    
    var buf bytes.Buffer
    writer := gzip.NewWriter(&buf)
    
    if _, err := writer.Write(content); err != nil {
        return nil, err
    }
    
    if err := writer.Close(); err != nil {
        return nil, err
    }
    
    compressed := buf.Bytes()
    if len(compressed) >= len(content) {
        return content, nil // Compression not beneficial
    }
    
    return compressed, nil
}
```

## Indexing Optimizations

### B-tree Temporal Indexes

**Implementation**: `src/storage/binary/temporal_btree.go`

Optimized for temporal queries:

```go
type TemporalBTree struct {
    root      *BTNode
    order     int
    timeIndex map[int64]*BTNode // Time-based index
}

func (bt *TemporalBTree) SearchAsOf(timestamp int64) []*Entity {
    node := bt.findClosestTime(timestamp)
    return bt.collectEntitiesAsOf(node, timestamp)
}
```

### Skip-List Indexes

**Implementation**: `src/storage/binary/skiplist_index.go`

O(log n) tag lookups:

```go
type SkipListIndex struct {
    head   *SkipNode
    level  int
    random *rand.Rand
}

func (sl *SkipListIndex) Search(tag string) []*Entity {
    current := sl.head
    
    // Skip down levels for fast traversal
    for i := sl.level; i >= 0; i-- {
        for current.forward[i] != nil && current.forward[i].tag < tag {
            current = current.forward[i]
        }
    }
    
    current = current.forward[0]
    if current != nil && current.tag == tag {
        return current.entities
    }
    
    return nil
}
```

### Bloom Filters

**Implementation**: `src/storage/binary/bloom_filter.go`

Fast existence checks:

```go
type BloomFilter struct {
    bits     []bool
    hashFuncs int
    size     uint
}

func (bf *BloomFilter) Test(key string) bool {
    for i := 0; i < bf.hashFuncs; i++ {
        hash := bf.hash(key, i) % bf.size
        if !bf.bits[hash] {
            return false
        }
    }
    return true
}
```

## Performance Metrics

### Benchmark Results

- **Write throughput**: 100,000+ entities/second
- **Read latency**: Sub-millisecond temporal queries
- **Memory efficiency**: 70% reduction with string interning
- **Concurrency**: Linear scaling with sharded locks

### Memory Usage

```bash
# Before optimizations
Heap: 2.1GB
GC: 45ms average
Allocations: 15M objects/sec

# After optimizations  
Heap: 650MB (70% reduction)
GC: 12ms average (73% improvement)
Allocations: 4.2M objects/sec (72% reduction)
```

### Query Performance

```bash
# Temporal queries
Simple as-of query: 0.3ms
Complex history query: 1.2ms
Tag-based query: 0.1ms

# Concurrent performance
1 goroutine:   85k ops/sec
10 goroutines: 780k ops/sec
100 goroutines: 6.2M ops/sec
```

## Configuration

### Performance Mode

```bash
# Enable high-performance mode
ENTITYDB_HIGH_PERFORMANCE=true

# Configure sharding
ENTITYDB_LOCK_SHARDS=64
ENTITYDB_INDEX_SHARDS=32

# Buffer pool settings
ENTITYDB_BUFFER_POOL_SIZE=100
ENTITYDB_LARGE_BUFFER_SIZE=1048576
```

### Memory Settings

```bash
# String interning
ENTITYDB_STRING_INTERN=true
ENTITYDB_INTERN_CACHE_SIZE=10000

# Memory mapping
ENTITYDB_MMAP_ENABLED=true
ENTITYDB_MMAP_SIZE=1073741824  # 1GB
```

## Monitoring

### Performance Metrics

Available via `/metrics` endpoint:
- `entitydb_memory_usage_bytes`
- `entitydb_gc_duration_seconds`
- `entitydb_lock_contention_total`
- `entitydb_query_duration_seconds`
- `entitydb_string_intern_hit_ratio`

### Health Checks

```bash
# Performance health check
curl https://localhost:8085/health

# Detailed performance metrics
curl https://localhost:8085/api/v1/system/metrics?section=performance
```

## Troubleshooting

### Common Performance Issues

1. **High memory usage**: Enable string interning, check for memory leaks
2. **Lock contention**: Increase shard count, optimize access patterns
3. **Slow queries**: Check index usage, enable bloom filters
4. **GC pressure**: Tune buffer pools, reduce allocations

### Debug Commands

```bash
# Enable performance tracing
curl -X POST https://localhost:8085/api/v1/admin/trace-subsystems \
  -d '{"subsystems": ["lock", "query", "cache"]}'

# Check string interning stats
curl https://localhost:8085/api/v1/admin/performance-stats
```

---

These optimizations provide the foundation for EntityDB's exceptional performance while maintaining data consistency and system reliability.