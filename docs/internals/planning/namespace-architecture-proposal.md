# Namespace Architecture Proposal

## Executive Summary

Transform EntityDB from a dataset-centric model to a namespace-isolated architecture, where each namespace is a bounded context with its own index file, providing massive performance improvements and true multi-tenant isolation.

## Conceptual Shift: Dataset → Namespace

### What Changes Fundamentally

**Dataset Model** (Current):
- Entities "connect to" datasets via tags
- Single global index for all data
- Datasets are just filtered views of global data
- Mental model: "One database with access points"

**Namespace Model** (Proposed):
- Entities "exist within" namespaces
- Each namespace has isolated storage/indexes
- Namespaces are bounded contexts
- Mental model: "Federation of mini-databases"

## Architecture Design

### File Structure
```
/var/entitydb/
├── entities.ebf              # Global entity storage (shared)
├── namespaces/
│   ├── default.idx          # Default namespace index
│   ├── worca.idx            # Worca namespace index
│   ├── worca.wal            # Worca write-ahead log
│   ├── metrics.idx          # Metrics namespace index
│   ├── metrics.wal          # Metrics write-ahead log
│   └── system.idx           # System namespace index
```

### Data Model

```go
// Entity tags would use namespace prefix
type Entity struct {
    ID        string
    Namespace string   // Explicit namespace field
    Tags      []string // No namespace prefix needed within namespace
    Content   []byte
}

// Examples:
// Old: tags = ["dataset:worca", "worca:self:type:task", "worca:trait:status:open"]
// New: namespace = "worca", tags = ["type:task", "status:open"]
```

### Index Structure

```go
type NamespaceIndex struct {
    Name      string
    FilePath  string                    // /var/entitydb/namespaces/worca.idx
    Entities  map[string]bool          // Entity IDs in this namespace
    TagIndex  map[string][]string      // tag -> entity IDs (namespace-local)
    TypeIndex map[string][]string      // type:value -> entity IDs
    MetaIndex map[string]map[string][]string // Complex queries
    
    // Performance optimizations
    BloomFilter *BloomFilter          // Quick existence checks
    SkipList    *SkipList            // Ordered access
    Cache       *LRUCache            // Hot data
}
```

## Performance Benefits

### 1. Query Performance
- **Current**: O(n) where n = all entities in system
- **Namespaced**: O(m) where m = entities in namespace
- **Typical improvement**: 10-100x for namespace queries

### 2. Write Performance
- **Current**: Update global index on every write
- **Namespaced**: Update only namespace index
- **Benefit**: Parallel writes to different namespaces

### 3. Memory Usage
- **Current**: All indexes in memory
- **Namespaced**: Load only active namespace indexes
- **Benefit**: 90% memory reduction for inactive namespaces

### 4. Startup Time
- **Current**: Load and index all entities
- **Namespaced**: Load only requested namespaces on-demand
- **Benefit**: Near-instant startup

## Implementation Plan

### Phase 1: Add Namespace Support (Week 1)
1. Add namespace field to Entity model
2. Create NamespaceIndex type
3. Implement namespace-aware repository
4. Keep backward compatibility with dataset tags

### Phase 2: Per-Namespace Files (Week 2)
1. Implement .idx file format for namespaces
2. Create namespace index loader/saver
3. Add namespace-specific WAL files
4. Implement lazy loading of namespaces

### Phase 3: API Migration (Week 3)
1. Add `/api/v1/namespaces/` endpoints
2. Update entity APIs to be namespace-aware
3. Create migration tool for dataset→namespace
4. Update documentation

### Phase 4: Optimization (Week 4)
1. Implement namespace-specific caching
2. Add namespace metrics/monitoring
3. Optimize cross-namespace queries
4. Performance testing and tuning

## API Design

### Namespace-Aware Endpoints

```bash
# Namespace management
GET    /api/v1/namespaces                    # List namespaces
POST   /api/v1/namespaces                    # Create namespace
DELETE /api/v1/namespaces/{namespace}        # Delete namespace

# Entity operations within namespace
GET    /api/v1/namespaces/{namespace}/entities
POST   /api/v1/namespaces/{namespace}/entities
GET    /api/v1/namespaces/{namespace}/entities/{id}
PUT    /api/v1/namespaces/{namespace}/entities/{id}
DELETE /api/v1/namespaces/{namespace}/entities/{id}

# Queries within namespace
GET    /api/v1/namespaces/{namespace}/query
GET    /api/v1/namespaces/{namespace}/entities/by-tag/{tag}

# Cross-namespace queries (requires special permission)
POST   /api/v1/query/cross-namespace
```

### Backward Compatibility

```bash
# Old dataset endpoints redirect to namespace endpoints
GET /api/v1/datasets/{dataset}/entities → GET /api/v1/namespaces/{dataset}/entities
```

## Benefits Summary

1. **Performance**: 10-100x improvement for namespace-scoped queries
2. **Scalability**: Linear scaling with number of namespaces
3. **Isolation**: True multi-tenant data isolation
4. **Memory**: Only active namespaces in memory
5. **Simplicity**: Cleaner mental model and API
6. **Migration**: Can run parallel to existing system

## Example Use Cases

### 1. Worca (Workforce Management)
- Namespace: `worca`
- Contains: tasks, projects, teams
- Benefits: Fast project queries, isolated from other systems

### 2. Metrics Collection
- Namespace: `metrics`
- Contains: performance data, system metrics
- Benefits: Time-series optimized indexes, separate retention

### 3. Configuration
- Namespace: `config`
- Contains: system settings, feature flags
- Benefits: Small, fast, frequently accessed

### 4. User Data
- Namespace: `users`
- Contains: user profiles, preferences
- Benefits: GDPR compliance, easy user deletion

## Conclusion

The shift from datasets to namespaces isn't just a rename - it's a fundamental architectural improvement that:
- Provides massive performance gains
- Enables true multi-tenancy
- Simplifies the mental model
- Prepares for distributed deployment

This positions EntityDB as a modern, scalable data platform rather than just a tag-based database.