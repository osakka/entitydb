# Dataset Architecture: A New Vision for EntityDB

## The Revelation: Dataset → Namespace → DATASET

Each evolution reveals deeper truth:
- **Dataset**: Connection points (too limiting)
- **Namespace**: Naming isolation (still thinking hierarchically)  
- **Dataset**: Complete data universes (TRUE insight!)

## What is a Dataset?

A dataset is not just a namespace or folder - it's a **complete, self-contained data universe** with:
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
       └── Filtered by dataset tags
            └── Same rules everywhere
```

### New: Federation of Data Universes
```
EntityDB Federation
  ├── dataset: worca/
  │   ├── Optimized for: Kanban workflows
  │   ├── Index strategy: By status/assignee
  │   └── Temporal: Task history tracking
  │
  ├── dataset: metrics/
  │   ├── Optimized for: Time-series data
  │   ├── Index strategy: By timestamp/metric
  │   └── Temporal: Rolling windows
  │
  └── dataset: knowledge/
      ├── Optimized for: Graph relationships
      ├── Index strategy: By connections
      └── Temporal: Version history
```

## Why "Dataset" Changes Everything

### 1. Each Dataset Can Have Different Physics
```go
type DatasetConfig struct {
    Name            string
    IndexStrategy   IndexType    // BTREE, HASH, TIMESERIES, GRAPH
    TemporalMode    TemporalType // FULL_HISTORY, ROLLING_WINDOW, LATEST_ONLY
    Compression     bool         // Different per space
    CacheStrategy   CacheType    // LRU, LFU, ADAPTIVE
    RetentionDays   int         // Data lifecycle per space
}
```

### 2. Specialized Storage Per Dataset
```
/var/entitydb/datasets/
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

### 3. Query Optimization Per Dataset

```go
// Worca dataset - optimized for status queries
worca.Query("status:open AND assignee:john")  // Uses status.idx

// Metrics dataset - optimized for time ranges  
metrics.Query("metric:cpu.usage AND time:[now-1h TO now]")  // Uses window.idx

// Knowledge dataset - optimized for relationships
knowledge.Query("related_to:quantum-physics DEPTH:3")  // Uses graph.idx
```

## The Architecture

```go
type Dataset interface {
    // Core operations
    Create(entity *Entity) error
    Get(id string) (*Entity, error)
    Query(query Query) ([]*Entity, error)
    
    // Dataset-specific operations
    GetConfig() DatasetConfig
    Optimize() error  // Each space optimizes differently
    Export() (io.Reader, error)
    Import(io.Reader) error
}

type EntityDB struct {
    datasets map[string]Dataset
    
    // Federation operations
    CreateDataset(config DatasetConfig) error
    GetDataset(name string) (Dataset, error)
    ListDatasets() []string
    
    // Cross-space operations (when needed)
    FederatedQuery(spaces []string, query Query) ([]*Entity, error)
}
```

## Dataset Types (Built-in)

### 1. Workflow Dataset (like Worca)
- Optimized for: State machines, task tracking
- Indexes: Status, assignee, priority, deadlines
- Special features: State transition tracking

### 2. Metrics Dataset  
- Optimized for: Time-series data
- Indexes: Metric names, time windows
- Special features: Downsampling, aggregations

### 3. Document Dataset
- Optimized for: Large content, full-text search
- Indexes: Content chunks, full-text, metadata
- Special features: Compression, deduplication

### 4. Graph Dataset
- Optimized for: Relationships, networks
- Indexes: Edges, paths, clusters
- Special features: Graph algorithms built-in

### 5. Config Dataset
- Optimized for: Small, frequently accessed
- Indexes: Key-value, version history
- Special features: In-memory, instant access

## Implementation Benefits

### 1. True Performance Isolation
- Each dataset has its own locks
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
worca := db.GetDataset("worca")
tasks := worca.Query("status:open")

// Not this mess:
tasks := db.Query("dataset:worca AND worca:self:status:open")
```

## Migration Path

1. Start calling datasets "datasets" internally
2. Create dataset abstraction over current dataset system
3. Implement per-dataset index files
4. Gradually optimize each dataset
5. Full federation when ready

## Vision

EntityDB becomes not just a database, but a **data federation platform** where each dataset is optimized for its specific use case. Like having Redis for cache, PostgreSQL for transactions, and Elasticsearch for search - but all in one cohesive system with unified API.

This is the future: **Datasets - where data lives in its natural habitat.**