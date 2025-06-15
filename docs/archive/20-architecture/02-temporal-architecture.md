# EntityDB Temporal Architecture

> **Status**: Core feature since v2.8.0  
> **Last Updated**: June 7, 2025

## Overview

EntityDB implements a groundbreaking temporal data architecture where every tag is timestamped with nanosecond precision, enabling complete time-travel capabilities and event sourcing without additional infrastructure.

## Core Temporal Design

### Tag Structure
Every tag in EntityDB has two forms:
1. **Temporal Tag**: `1749303910369730667|namespace:value`
2. **Simple Tag**: `namespace:value` (timestamps hidden in API responses)

The temporal tag captures the exact moment of creation using nanosecond epoch timestamps, making every change trackable through time.

### Event Sourcing by Default
- Each tag is an immutable event
- Entity state = sum of all temporal events up to a point in time
- No updates, only appends (new temporal tags)
- Complete audit trail built-in
- Zero additional configuration required

## Temporal Capabilities

### 1. Time-Travel Queries
```bash
# Get entity state at any point in time
curl "https://localhost:8085/api/v1/entities/as-of?id=ENTITY_ID&timestamp=2025-01-01T00:00:00Z"

# Get complete entity history
curl "https://localhost:8085/api/v1/entities/history?id=ENTITY_ID"

# Get changes between two points in time
curl "https://localhost:8085/api/v1/entities/diff?id=ENTITY_ID&from=T1&to=T2"
```

### 2. Temporal Tag Management
```go
// Add timestamped tag (automatic)
entity.AddTag("status:completed")
// Internally stored as: "1749303910369730667|status:completed"

// Query at specific time
entities := repo.ListByTagAsOf("status:active", timestamp)

// Get tag timeline
timeline := repo.GetTagTimeline(entityID, "status")
```

### 3. Change Detection
```bash
# Get all changes for an entity
curl "https://localhost:8085/api/v1/entities/changes?id=ENTITY_ID"

# Response shows temporal evolution
{
  "changes": [
    {
      "timestamp": 1749303910369730667,
      "tags_added": ["status:draft"],
      "tags_removed": []
    },
    {
      "timestamp": 1749303920479834123,
      "tags_added": ["status:completed"],
      "tags_removed": ["status:draft"]
    }
  ]
}
```

## Storage Implementation

### Temporal Repository
**File**: `src/storage/binary/temporal_repository.go`

Combines high-performance storage with temporal capabilities:

```go
type TemporalRepository struct {
    *HighPerformanceRepository
    timelineIndex  *TemporalIndex
    temporalCache  *LRUCache
}

func (tr *TemporalRepository) AddTag(entityID, tag string) error {
    timestamp := time.Now().UnixNano()
    temporalTag := fmt.Sprintf("%d|%s", timestamp, tag)
    
    // Store temporal tag
    if err := tr.addTemporalTag(entityID, temporalTag); err != nil {
        return err
    }
    
    // Update timeline index
    return tr.timelineIndex.Add(entityID, timestamp, tag)
}
```

### B-tree Timeline Indexes
**File**: `src/storage/binary/temporal_btree.go`

Optimized data structure for temporal queries:

```go
type TemporalBTree struct {
    root        *BTNode
    order       int
    timeIndex   map[int64]*BTNode  // Timestamp -> Node mapping
    entityIndex map[string]*BTNode // Entity -> Timeline mapping
}

func (bt *TemporalBTree) SearchAsOf(entityID string, timestamp int64) []*TemporalTag {
    timeline := bt.entityIndex[entityID]
    if timeline == nil {
        return nil
    }
    
    // Binary search for closest timestamp <= target
    node := bt.findClosestTime(timeline, timestamp)
    return bt.collectTagsAsOf(node, timestamp)
}
```

### Temporal Index Persistence
**File**: `src/storage/binary/tag_index_persistence_v2.go`

Persistent temporal indexes for fast startup:

```go
type TemporalIndexPersistence struct {
    indexFile  string
    version    uint32
    entityMap  map[string]*EntityTimeline
}

func (tip *TemporalIndexPersistence) SaveIndex() error {
    file, err := os.Create(tip.indexFile)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Write index header
    binary.Write(file, binary.LittleEndian, tip.version)
    binary.Write(file, binary.LittleEndian, uint32(len(tip.entityMap)))
    
    // Write entity timelines
    for entityID, timeline := range tip.entityMap {
        if err := tip.writeTimeline(file, entityID, timeline); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Performance Optimizations

### Timeline Caching
- **LRU Cache**: Recently accessed timelines cached in memory
- **Pre-loading**: Active entity timelines pre-loaded on startup
- **Change-only Updates**: Only modified timelines written to disk

### Query Optimization
- **Time-bucketed Indexes**: Group events by time periods for range queries
- **Binary Search**: O(log n) temporal lookups
- **Parallel Processing**: Multiple timeline queries executed concurrently

### Memory Management
- **Lazy Loading**: Timelines loaded on first access
- **Memory Mapping**: Large timeline files memory-mapped for efficiency
- **Compression**: Old timeline segments compressed automatically

## API Design

### Transparent Timestamps
By default, API responses hide temporal information:

```json
{
  "id": "entity_123",
  "tags": ["type:task", "status:completed"],
  "content": "Task description"
}
```

With `include_timestamps=true`:

```json
{
  "id": "entity_123", 
  "tags": [
    "1749303910369730667|type:task",
    "1749303920479834123|status:completed"
  ],
  "content": "Task description"
}
```

### Temporal Query Parameters
- `timestamp`: Unix nanoseconds or ISO 8601 format
- `from` / `to`: Range queries
- `include_timestamps`: Show temporal tag format
- `timeline`: Return full timeline instead of point-in-time state

## Use Cases

### 1. Audit Trails
```bash
# Who changed what when?
curl "https://localhost:8085/api/v1/entities/history?id=ENTITY_ID"

# What was the state at specific time?
curl "https://localhost:8085/api/v1/entities/as-of?id=ENTITY_ID&timestamp=2025-01-01T12:00:00Z"
```

### 2. Event Sourcing
```bash
# Replay all events for entity
curl "https://localhost:8085/api/v1/entities/changes?id=ENTITY_ID"

# Rebuild state from events
for change in changes:
    apply_change(entity, change.tags_added, change.tags_removed)
```

### 3. Compliance & Forensics
```bash
# Generate compliance report
curl "https://localhost:8085/api/v1/entities/diff?id=ENTITY_ID&from=2025-01-01&to=2025-12-31"

# Investigate data changes
curl "https://localhost:8085/api/v1/entities/history?id=ENTITY_ID&include_timestamps=true"
```

### 4. Version Control
```bash
# Compare two versions
curl "https://localhost:8085/api/v1/entities/diff?id=ENTITY_ID&from=T1&to=T2"

# Rollback to previous state (create new tags)
curl -X POST "https://localhost:8085/api/v1/entities/update" \
  -d '{"id": "ENTITY_ID", "tags": ["status:draft"]}' # Adds new temporal tag
```

## Configuration

### Temporal Settings
```bash
# Timeline index settings
ENTITYDB_TEMPORAL_INDEX_SHARDS=16
ENTITYDB_TEMPORAL_CACHE_SIZE=10000

# Compression settings
ENTITYDB_TEMPORAL_COMPRESS_AGE_DAYS=30
ENTITYDB_TEMPORAL_COMPRESS_RATIO=0.7

# Performance settings
ENTITYDB_TEMPORAL_PRELOAD_ACTIVE=true
ENTITYDB_TEMPORAL_PARALLEL_QUERIES=10
```

### Retention Policies
```bash
# Set retention via entity tags
curl -X POST "https://localhost:8085/api/v1/entities/create" \
  -d '{
    "tags": [
      "type:temp_data", 
      "retention:period:86400",    # 24 hours
      "retention:count:100"        # Keep max 100 versions
    ],
    "content": "Temporary data"
  }'
```

## Troubleshooting

### Common Issues

1. **Slow temporal queries**: Check timeline index integrity
2. **Memory usage**: Tune timeline cache size
3. **Missing history**: Verify temporal tag storage
4. **Timestamp format**: Ensure nanosecond precision

### Debug Commands

```bash
# Check temporal index health
curl "https://localhost:8085/api/v1/admin/temporal-index-stats"

# Rebuild temporal indexes
curl -X POST "https://localhost:8085/api/v1/admin/rebuild-temporal-index"

# Enable temporal tracing
curl -X POST "https://localhost:8085/api/v1/admin/trace-subsystems" \
  -d '{"subsystems": ["temporal"]}'
```

### Performance Monitoring

Available metrics via `/metrics`:
- `entitydb_temporal_query_duration_seconds`
- `entitydb_temporal_index_size_bytes`
- `entitydb_temporal_cache_hit_ratio`
- `entitydb_timeline_load_duration_seconds`

## Migration Notes

### From Non-Temporal Systems
1. **Existing entities**: Automatically get temporal tags on first modification
2. **Backward compatibility**: Simple tag format still supported
3. **Gradual migration**: Temporal features work alongside existing functionality

### Performance Impact
- **Storage**: ~15% increase for timestamp storage
- **Query performance**: Temporal queries ~2x faster than scanning full history
- **Memory usage**: Timeline indexes use ~5% of total memory

---

EntityDB's temporal architecture provides comprehensive time-travel capabilities with exceptional performance, making it ideal for applications requiring audit trails, compliance, and temporal analytics.