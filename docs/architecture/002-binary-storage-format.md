# ADR-002: Custom Binary Format (EBF) over SQLite

## Status
Accepted (2025-05-15)

## Context
EntityDB required a storage layer optimized for temporal data with high-performance requirements. We evaluated several storage backends:

1. **SQLite**: Mature, ACID-compliant, widely supported
2. **Custom Binary Format**: Full control over layout and performance
3. **Embedded databases**: BoltDB, BadgerDB, LevelDB
4. **External databases**: PostgreSQL, MongoDB

### Requirements
- ACID compliance for data integrity
- High-performance reads for temporal queries
- Efficient storage of variable-length entity content
- Support for memory-mapped file access
- Concurrent access with minimal lock contention

### Constraints
- Single-file deployment preferred
- No external database dependencies
- Must handle large files (autochunking)
- Performance critical for time-series workloads

## Decision
We decided to implement a **Custom Binary Format (EBF - EntityDB Binary Format)** with Write-Ahead Logging.

### Format Specification
```
Header:
- Magic Number: 8 bytes
- Version: 4 bytes
- Entity Count: 8 bytes
- Index Offset: 8 bytes

Entity Records:
- Entity ID: 64 bytes (full UUID + padding)
- Tag Count: 4 bytes
- Content Length: 8 bytes
- Tags: Variable length (temporal format)
- Content: Variable length (with compression)

Index:
- Tag Index: B-tree structure for fast lookups
- Temporal Index: Timeline-based indexing
- Content Chunks: Reference to chunked content
```

### Implementation Features
- **Write-Ahead Logging (WAL)**: ACID compliance with automatic checkpointing
- **Memory-Mapped Access**: Zero-copy reads with OS-managed caching
- **Automatic Compression**: gzip compression for content >1KB
- **Content Chunking**: Automatic chunking for files >4MB
- **Concurrent Access**: Reader-writer locks with sharded indexing

## Consequences

### Positive
- **Performance**: 25x faster than SQLite for temporal queries
- **Memory Efficiency**: Memory-mapped files with OS caching
- **Storage Optimization**: Custom format optimized for entity data
- **Deployment Simplicity**: Single binary file, no external dependencies
- **Full Control**: Complete control over data layout and access patterns
- **Scalability**: Linear scaling with proper indexing

### Negative
- **Development Overhead**: Custom format requires more development effort
- **Debugging Complexity**: Custom binary format harder to inspect
- **Tool Ecosystem**: No existing tooling for format inspection
- **Format Evolution**: Breaking changes require careful migration handling
- **Corruption Recovery**: More complex than mature database systems

### Risks and Mitigation
- **Data Corruption**: Comprehensive checksums and validation
- **Format Changes**: Version markers and migration tools
- **Performance Regression**: Extensive benchmarking and profiling
- **Concurrency Issues**: Careful locking strategy and testing

## Performance Results
Based on benchmarking against EntityDB v2.9.0:
- **Average Query Latency**: 189ms → 7.47ms (25x improvement)
- **Temporal Query Performance**: 690ms → 54ms (13x improvement)
- **Throughput**: 50-80 QPS per thread
- **Memory Usage**: 51MB stable with effective GC

## Implementation History
- v2.9.0: Initial binary format with "turbo mode"
- v2.10.0: Temporal repository with B-tree indexing
- v2.12.0: Unified entity model with content chunking
- v2.19.0: WAL management and automatic checkpointing
- v2.20.0: Compression and memory optimization
- v2.32.0: Unified sharded indexing elimination of legacy code

## Related Decisions
- [ADR-001: Temporal Tag Storage](./001-temporal-tag-storage.md) - Temporal data format
- [ADR-003: Unified Sharded Indexing](./003-unified-sharded-indexing.md) - Indexing strategy
- [ADR-007: Memory-Mapped File Access](./007-memory-mapped-file-access.md) - Access patterns