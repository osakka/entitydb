# EntityDB Implementation Guide

This directory contains detailed implementation documentation for EntityDB's core systems.

## 🏗️ Core Implementation Areas

### Storage & Indexing
- **[Sharded Indexing Implementation](./sharded-indexing-implementation.md)** - 256-shard concurrent indexing system details
- **[Temporal Implementation](./temporal-implementation-guide.md)** - Time-travel query implementation
- **[WAL Implementation](./wal-implementation-guide.md)** - Write-Ahead Logging system

### Performance & Optimization
- **[Memory Management](./memory-management-implementation.md)** - Cache, interning, and buffer pool implementation
- **[Query Optimization](./query-optimization-guide.md)** - Multi-tag query performance optimization
- **[Batch Operations](./batch-operations-guide.md)** - Batch writer implementation

### Security & RBAC
- **[RBAC Implementation](./rbac-implementation-guide.md)** - Tag-based permission system
- **[Authentication Flow](./authentication-implementation.md)** - JWT and session management
- **[Security Hardening](./security-hardening-guide.md)** - Input validation and sanitization

## 📋 Implementation Patterns

### Code Organization
```
src/
├── storage/binary/     # Core storage implementation
├── api/               # HTTP handlers and middleware
├── models/            # Entity and data models
└── logger/            # Logging subsystem
```

### Key Interfaces
```go
// Repository interface - core data access
type Repository interface {
    Create(entity *Entity) error
    GetByID(id string) (*Entity, error)
    Update(entity *Entity) error
    Delete(id string) error
    ListByTag(tag string) ([]*Entity, error)
}

// TemporalRepository - time-travel queries
type TemporalRepository interface {
    Repository
    GetAsOf(id string, timestamp time.Time) (*Entity, error)
    GetHistory(id string) ([]*Entity, error)
    GetChanges(since time.Time) ([]*Entity, error)
}
```

### Performance Considerations

**Indexing Strategy**:
- 256 shards for concurrent access
- Read-write locks per shard
- Tag variant caching for O(1) lookups

**Memory Management**:
- Entity cache with LRU eviction
- String interning for tags
- Buffer pools for allocations

**Query Optimization**:
- Smart ordering for multi-tag queries
- Early termination on empty results
- Intersection-based AND logic

## 🔧 Implementation Guidelines

### Error Handling
```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create entity %s: %w", entity.ID, err)
}

// Use appropriate log levels
logger.Error("Critical operation failed", "entity", entity.ID, "error", err)
logger.Warn("Retrying operation", "attempt", attempt)
logger.Debug("Cache hit", "entity", entity.ID)
```

### Testing Requirements
- Unit tests for all public functions
- Integration tests for API endpoints
- Performance benchmarks for critical paths
- Concurrent operation testing

### Code Standards
- Follow Go idioms and best practices
- Document all exported functions
- Use meaningful variable names
- Keep functions focused and small

## 📚 Related Documentation

### Technical Specifications
- [Binary Format Spec](../../reference/technical-specs/binary-format-specification.md)
- [File Format Spec](../../reference/technical-specs/unified-file-format-specification.md)
- [Memory Architecture](../../reference/technical-specs/memory-optimization-architecture.md)

### Testing Guides
- [Production Battle Testing](../testing/production-battle-testing-guide.md)
- [E2E Testing Guide](../testing/)
- [Performance Testing](../testing/)

### Architecture Decisions
- [ADR Index](../../architecture/adr/)
- [Architecture Overview](../../architecture/)

---

**Last Updated**: 2025-06-23  
**Maintainers**: EntityDB Core Team