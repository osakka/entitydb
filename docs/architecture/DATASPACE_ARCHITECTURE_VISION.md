# Dataspace Architecture: A New Vision for EntityDB

## The Revelation: Dataspace → Namespace → DATASPACE

Each evolution reveals deeper truth:
- **Dataspace**: Connection points (too limiting)
- **Namespace**: Naming isolation (still thinking hierarchically)  
- **Dataspace**: Complete data universes (TRUE insight!)

## What is a Dataspace?

A dataspace is not just a namespace or folder - it's a **complete, self-contained data universe** with:
- Its own index laws (how data is organized)
- Its own query physics (how data is accessed)
- Its own temporal rules (how time works)
- Its own performance characteristics
- Its own access patterns

## The Mental Model Shift

### Old: Single Database with Filters
```
EntityDB
  └── All entities (global soup)
       └── Filtered by dataspace tags
            └── Same rules everywhere
```

### New: Federation of Data Universes
```
EntityDB Federation
  ├── dataspace: worca/
  │   ├── Optimized for: Kanban workflows
  │   ├── Index strategy: By status/assignee
  │   └── Temporal: Task history tracking
  │
  ├── dataspace: metrics/
  │   ├── Optimized for: Time-series data
  │   ├── Index strategy: By timestamp/metric
  │   └── Temporal: Rolling windows
  │
  └── dataspace: knowledge/
      ├── Optimized for: Graph relationships
      ├── Index strategy: By connections
      └── Temporal: Version history
```

## Why "Dataspace" Changes Everything

### 1. Each Dataspace Can Have Different Physics
```go
type DataspaceConfig struct {
    Name            string
    IndexStrategy   IndexType    // BTREE, HASH, TIMESERIES, GRAPH
    TemporalMode    TemporalType // FULL_HISTORY, ROLLING_WINDOW, LATEST_ONLY
    Compression     bool         // Different per space
    CacheStrategy   CacheType    // LRU, LFU, ADAPTIVE
    RetentionDays   int         // Data lifecycle per space
}
```

### 2. Specialized Storage Per Dataspace
```
/var/entitydb/dataspaces/
├── worca/
│   ├── data.ebf         # Entities
│   ├── status.idx       # Status-optimized index
│   ├── assignee.idx     # Person-optimized index
│   └── temporal.idx     # Change tracking
│
├── metrics/
│   ├── data.tsdb        # Time-series format!
│   ├── metric.idx       # Metric name index
│   └── window.idx       # Time window index
│
└── knowledge/
    ├── data.ebf         # Entities
    ├── graph.idx        # Relationship graph
    └── vector.idx       # Semantic search
```

### 3. Query Optimization Per Dataspace

```go
// Worca dataspace - optimized for status queries
worca.Query("status:open AND assignee:john")  // Uses status.idx

// Metrics dataspace - optimized for time ranges  
metrics.Query("metric:cpu.usage AND time:[now-1h TO now]")  // Uses window.idx

// Knowledge dataspace - optimized for relationships
knowledge.Query("related_to:quantum-physics DEPTH:3")  // Uses graph.idx
```

## The Architecture

```go
type Dataspace interface {
    // Core operations
    Create(entity *Entity) error
    Get(id string) (*Entity, error)
    Query(query Query) ([]*Entity, error)
    
    // Dataspace-specific operations
    GetConfig() DataspaceConfig
    Optimize() error  // Each space optimizes differently
    Export() (io.Reader, error)
    Import(io.Reader) error
}

type EntityDB struct {
    dataspaces map[string]Dataspace
    
    // Federation operations
    CreateDataspace(config DataspaceConfig) error
    GetDataspace(name string) (Dataspace, error)
    ListDataspaces() []string
    
    // Cross-space operations (when needed)
    FederatedQuery(spaces []string, query Query) ([]*Entity, error)
}
```

## Dataspace Types (Built-in)

### 1. Workflow Dataspace (like Worca)
- Optimized for: State machines, task tracking
- Indexes: Status, assignee, priority, deadlines
- Special features: State transition tracking

### 2. Metrics Dataspace  
- Optimized for: Time-series data
- Indexes: Metric names, time windows
- Special features: Downsampling, aggregations

### 3. Document Dataspace
- Optimized for: Large content, full-text search
- Indexes: Content chunks, full-text, metadata
- Special features: Compression, deduplication

### 4. Graph Dataspace
- Optimized for: Relationships, networks
- Indexes: Edges, paths, clusters
- Special features: Graph algorithms built-in

### 5. Config Dataspace
- Optimized for: Small, frequently accessed
- Indexes: Key-value, version history
- Special features: In-memory, instant access

## Implementation Benefits

### 1. True Performance Isolation
- Each dataspace has its own locks
- No cross-space contention
- Parallel operations by default

### 2. Semantic Clarity
- Data lives in its natural space
- No prefix pollution
- Clear boundaries

### 3. Future-Proof Architecture
- Each space can evolve independently
- New storage engines per space
- Easy to distribute/shard

### 4. Developer Experience
```go
// Clear, intuitive API
worca := db.GetDataspace("worca")
tasks := worca.Query("status:open")

// Not this mess:
tasks := db.Query("dataspace:worca AND worca:self:status:open")
```

## Migration Path

1. Start calling dataspaces "dataspaces" internally
2. Create dataspace abstraction over current dataspace system
3. Implement per-dataspace index files
4. Gradually optimize each dataspace
5. Full federation when ready

## Vision

EntityDB becomes not just a database, but a **data federation platform** where each dataspace is optimized for its specific use case. Like having Redis for cache, PostgreSQL for transactions, and Elasticsearch for search - but all in one cohesive system with unified API.

This is the future: **Dataspaces - where data lives in its natural habitat.**