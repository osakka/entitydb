# EntityDB Architecture Overview

> **Version**: v2.32.2 | **Last Updated**: 2025-06-13 | **Status**: AUTHORITATIVE

## System Architecture

EntityDB is a high-performance temporal database built around a pure entity model with nanosecond-precision timestamps. The architecture is designed for speed, durability, and temporal query capabilities.

## Core Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   REST API      │    │   Web Dashboard │    │   Admin Tools   │
│   (main.go)     │    │   (htdocs/)     │    │   (tools/)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                        API Layer                                │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Auth Middleware │  │ RBAC Middleware │  │ Trace Middleware│ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Entity Handler  │  │ Metrics Handler │  │ Admin Handler   │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                      Storage Layer                              │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Entity Repo     │  │ Temporal Repo   │  │ Relationship    │ │
│  │ (binary/)       │  │ (temporal_*)    │  │ Repo            │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Sharded Locks   │  │ Tag Index       │  │ WAL Manager     │ │
│  │ (locks_*)       │  │ (tag_index_*)   │  │ (wal.go)        │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                 │
┌─────────────────────────────────────────────────────────────────┐
│                      Binary Storage                             │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Memory-Mapped   │  │ Write-Ahead     │  │ Tag Indexes     │ │
│  │ Files (.ebf)    │  │ Log (.wal)      │  │ (.idx)          │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Entity Model

### Pure Entity Architecture

EntityDB uses a unified entity model where everything is an entity:

```go
type Entity struct {
    ID        string   `json:"id"`        // 64-byte UUID
    Tags      []string `json:"tags"`      // Temporal tags with timestamps
    Content   []byte   `json:"content"`   // Binary content (auto-chunked)
    CreatedAt int64    `json:"created_at"` // Nanosecond epoch
    UpdatedAt int64    `json:"updated_at"` // Nanosecond epoch
}
```

### Temporal Tag System

All tags are stored with nanosecond timestamps using the format:

```
TIMESTAMP|tag
```

Examples:
- `1749303910369730667|type:user`
- `1749303910369732135|rbac:role:admin`
- `1749303910369733452|status:active`

## Storage Layer

### Binary Format (.ebf)

EntityDB uses a custom binary format for maximum performance:

- **Memory-mapped files** for zero-copy reads
- **Compression** for content > 1KB using gzip
- **Checksums** for data integrity verification
- **Variable-length encoding** for space efficiency

### Write-Ahead Logging (WAL)

- **Durability guarantee** through WAL persistence
- **Automatic checkpointing** every 1000 operations, 5 minutes, or 100MB
- **Concurrent writes** with proper ordering
- **Recovery** from WAL on startup

### Indexing System

#### Tag Indexing
- **Sharded tag index** for high concurrency (configurable shards)
- **Bloom filters** for fast existence checks
- **Skip-list indexes** for O(log n) lookups
- **Persistent index files** (.idx) for fast startup

#### Temporal Indexing
- **B-tree timeline indexes** for temporal queries
- **Time-bucketed indexes** for range queries
- **Per-entity temporal timelines**
- **Temporal query caching** with LRU eviction

## Performance Optimizations

### Memory Management
- **String interning** for tag storage (up to 70% memory reduction)
- **Safe buffer pools** with size-based allocation (small/medium/large)
- **Memory-mapped files** with OS-managed caching
- **Zero-copy operations** where possible

### Concurrency Control
- **Sharded locking system** to distribute contention
- **Fair queue implementation** for read/write access
- **Deadlock detection** and prevention
- **Lock tracing** for debugging concurrent access

### Auto-chunking
- **Automatic chunking** for files > 4MB (configurable)
- **Streaming support** for large files without RAM limits
- **Chunk retrieval** on demand
- **Content checksums** for integrity verification

## Security Architecture

### RBAC System

EntityDB implements tag-based Role-Based Access Control:

#### Permission Format
```
rbac:perm:resource:action
```

Examples:
- `rbac:perm:entity:view`
- `rbac:perm:entity:create`
- `rbac:perm:system:admin`

#### Role Format
```
rbac:role:role_name
```

Examples:
- `rbac:role:admin`
- `rbac:role:user`

### Session Management
- **Session-based authentication** with TTL
- **Secure token generation** using crypto/rand
- **Session refresh** capability
- **Automatic cleanup** of expired sessions
- **Concurrent session** support

### Security Middleware
- **Authentication middleware** for session validation
- **RBAC middleware** for permission enforcement
- **Trace middleware** for request tracking
- **Connection middleware** for stability

## Configuration System

### Three-Tier Hierarchy

1. **Database Configuration** (highest priority)
   - Stored as entities with `type:config` tags
   - Runtime updates via API
   - Cached for 5 minutes

2. **Command-Line Flags** (medium priority)
   - Long format: `--entitydb-xxx`
   - Short flags reserved for `-h` and `-v`

3. **Environment Variables** (lowest priority)
   - All prefixed with `ENTITYDB_`
   - Loaded from `share/config/entitydb.env`

### ConfigManager
- **Centralized configuration management**
- **Database caching** with automatic expiry
- **Runtime configuration updates**
- **Hierarchical precedence** enforcement

## Metrics & Observability

### Metrics Collection
- **Real-time metrics** with 1-second collection interval
- **Temporal storage** using AddTag() for time-series data
- **Change-only detection** to prevent redundant writes
- **Thread-safe implementation** with proper mutex protection

### Metrics Types
- **System metrics**: Memory, GC, database size
- **Performance metrics**: Query latency, storage operations
- **Authentication metrics**: Login success/failure, session activity
- **Error tracking**: Categorized error patterns and recovery

### Health Monitoring
- **Health endpoint** (`/health`) with comprehensive checks
- **Prometheus metrics** (`/metrics`) for monitoring integration
- **WAL monitoring** with size warnings and critical alerts
- **Index health verification** with automatic repair

## Dataset Architecture

### Multi-Tenant Design
- **Logical separation** of data by dataset
- **Shared underlying storage** for efficiency
- **Index isolation** per dataset
- **RBAC enforcement** across dataset boundaries

### Dataset Operations
- **Entity operations** scoped to dataset
- **Query isolation** between datasets
- **Relationship management** within dataset context
- **Metrics collection** per dataset

## API Design

### RESTful Architecture
- **Resource-oriented** URL design
- **HTTP verbs** for operation semantics
- **JSON payload** format
- **Consistent error responses**

### Middleware Stack
1. **Connection handling** middleware
2. **TE header stripping** middleware
3. **Request tracing** middleware
4. **Authentication** middleware
5. **RBAC** middleware
6. **Request metrics** middleware

### Endpoint Organization
- **Entity operations**: `/api/v1/entities/*`
- **Temporal queries**: `/api/v1/entities/as-of`, `/api/v1/entities/history`
- **Relationships**: `/api/v1/entity-relationships`
- **Datasets**: `/api/v1/datasets/{id}/*`
- **Admin operations**: `/api/v1/admin/*`
- **Metrics**: `/api/v1/metrics/*`

## Development Architecture

### Engineering Excellence
- **CI/CD pipelines** with GitHub Actions
- **Docker containerization** with multi-stage builds
- **One-command setup** for development environment
- **Hot reload** development with Air
- **Pre-commit hooks** for code quality

### Code Organization
```
src/
├── main.go              # Server initialization and routing
├── api/                 # API handlers and middleware
├── models/              # Data models and business logic
├── storage/binary/      # Binary storage implementation
├── logger/              # Logging infrastructure
├── config/              # Configuration management
└── tools/               # Administrative utilities
```

### Quality Assurance
- **Comprehensive testing** with multiple test frameworks
- **Security scanning** with gosec
- **Code quality checks** with golangci-lint
- **Documentation accuracy** verified against codebase
- **Performance benchmarking** with automated testing

---

This architecture provides a foundation for high-performance, scalable, and maintainable temporal database operations while ensuring security, durability, and ease of operation.