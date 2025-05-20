# EntityDB Temporal Architecture

## Overview

EntityDB implements a groundbreaking temporal data architecture where every tag is timestamped with nanosecond precision, enabling complete time-travel capabilities and event sourcing without additional infrastructure.

## Core Temporal Design

### Tag Structure
Every tag in EntityDB has two forms:
1. **Temporal Tag**: `2025-05-17T18:00:58.001351096.namespace:value`
2. **Simple Tag**: `namespace:value`

The temporal tag captures the exact moment of creation, making every change trackable through time.

### Event Sourcing by Default
- Each tag is an immutable event
- Entity state = sum of all temporal events
- No updates, only appends
- Complete audit trail built-in

## Temporal Capabilities

### 1. Time-Travel Queries
```go
// Get entity state at any point in time
func GetEntityAsOf(id string, timestamp time.Time) *Entity
func GetEntityHistory(id string, from, to time.Time) []EntityState
```

### 2. Change Detection
```go
// Find what changed between timestamps
func GetEntityDiff(id string, t1, t2 time.Time) []Change
func GetRecentChanges(since time.Time) []EntityChange
```

### 3. Temporal Relationships
- Track relationship evolution
- See when assignments changed
- Understand dependency timelines

### 4. Compliance & Auditing
- Who changed what when
- Complete audit trail
- Regulatory compliance
- Forensic analysis

## Binary Format Temporal Optimizations

### Temporal Indexes
```
TimeIndex
├── Year
│   ├── Month
│   │   ├── Day
│   │   │   ├── Hour
│   │   │   │   └── Entities changed
```

### Time-Based Partitioning
- Partition data by time ranges
- Quick temporal queries
- Efficient historical access
- Archive old partitions

### Temporal Compression
- Group tags by time proximity
- Delta encoding for timestamps
- Efficient storage of temporal data

## Query Patterns

### Point-in-Time Queries
```bash
# Entity state at specific time
entitydb-cli entity get --id=ent_123 --as-of="2025-01-15T10:00:00"

# All entities of type at time
entitydb-cli entity list --type=user --as-of="2025-01-15T10:00:00"
```

### Time Range Queries
```bash
# Changes in time range
entitydb-cli entity changes --from="2025-01-01" --to="2025-01-31"

# Entity history
entitydb-cli entity history --id=ent_123 --last="7d"
```

### Temporal Aggregations
```bash
# Count changes per day
entitydb-cli stats changes --group-by=day --last="30d"

# Most active periods
entitydb-cli stats activity --entity=ent_123 --resolution=hour
```

## Implementation Details

### Tag Timestamp Extraction
```go
func extractTimestamp(tag string) (time.Time, error) {
    // Format: YYYY-MM-DDTHH:MM:SS.nnnnnnnnn.namespace:value
    parts := strings.SplitN(tag, ".", 2)
    if len(parts) < 2 {
        return time.Time{}, fmt.Errorf("invalid temporal tag")
    }
    return time.Parse(time.RFC3339Nano, parts[0])
}
```

### Temporal Entity Reconstruction
```go
func (r *Repository) GetEntityAsOf(id string, asOf time.Time) (*Entity, error) {
    current, err := r.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    // Filter tags up to timestamp
    entity := &Entity{ID: id}
    for _, tag := range current.Tags {
        tagTime, _ := extractTimestamp(tag)
        if tagTime.Before(asOf) || tagTime.Equal(asOf) {
            entity.Tags = append(entity.Tags, tag)
        }
    }
    
    // Build latest state map
    stateMap := make(map[string]string)
    for _, tag := range entity.Tags {
        namespace, value := parseTag(tag)
        stateMap[namespace] = value
    }
    
    // Reconstruct simplified tags
    entity.Tags = []string{}
    for ns, val := range stateMap {
        entity.Tags = append(entity.Tags, fmt.Sprintf("%s:%s", ns, val))
    }
    
    return entity, nil
}
```

### Temporal Diff Algorithm
```go
func (r *Repository) GetEntityDiff(id string, t1, t2 time.Time) ([]Change, error) {
    state1, _ := r.GetEntityAsOf(id, t1)
    state2, _ := r.GetEntityAsOf(id, t2)
    
    changes := []Change{}
    
    // Find added tags
    for _, tag := range state2.Tags {
        if !contains(state1.Tags, tag) {
            changes = append(changes, Change{
                Type: "added",
                Tag:  tag,
                Time: t2,
            })
        }
    }
    
    // Find removed tags
    for _, tag := range state1.Tags {
        if !contains(state2.Tags, tag) {
            changes = append(changes, Change{
                Type: "removed",
                Tag:  tag,
                Time: t2,
            })
        }
    }
    
    return changes, nil
}
```

## Use Cases

### 1. Configuration Management
- Track config changes over time
- Rollback to known good states
- Understand when issues started
- A/B testing with temporal flags

### 2. Workflow Evolution
- See assignment history
- Track status changes
- Understand process bottlenecks
- Measure cycle times

### 3. Security & Compliance
- Complete audit trail
- Access history tracking
- Permission change monitoring
- Regulatory reporting

### 4. Debugging & Analysis
- Reproduce past states
- Understand cascading changes
- Root cause analysis
- Performance regression tracking

## Performance Considerations

### Query Optimization
1. **Temporal Indexes**: Pre-built time-based indexes
2. **Lazy Loading**: Load only required time ranges
3. **Caching**: Cache frequently accessed periods
4. **Parallel Processing**: Process time ranges concurrently

### Storage Optimization
1. **Compression**: Temporal tag compression
2. **Partitioning**: Time-based data partitioning
3. **Archival**: Move old data to cold storage
4. **Deduplication**: Remove redundant temporal data

## Future Enhancements

### 1. Temporal Query Language
```sql
SELECT * FROM entities
WHERE type = 'user'
AS OF SYSTEM TIME '2025-01-15 10:00:00'

SELECT * FROM entities
FOR SYSTEM TIME BETWEEN '2025-01-01' AND '2025-01-31'
WHERE status = 'active'
```

### 2. Temporal Triggers
- Execute actions at specific times
- Schedule future changes
- Time-based workflows
- Temporal constraints

### 3. Advanced Analytics
- Time-series analysis
- Trend detection
- Anomaly detection
- Predictive modeling

### 4. Temporal Visualization
- Timeline views
- Change heatmaps
- Evolution graphs
- Interactive history browser

## Benefits Over Traditional Approaches

1. **No Separate History Tables**: History is built into the data model
2. **No Trigger Complexity**: Changes are captured naturally
3. **Immutable by Design**: No accidental history loss
4. **Query Simplicity**: Time travel is a first-class feature
5. **Performance**: Optimized for temporal queries
6. **Audit Compliance**: Complete trail without additional work

## Conclusion

EntityDB's temporal architecture provides:
- Complete history of all changes
- Natural audit trail
- Time-travel capabilities
- Event sourcing by default
- No additional infrastructure needed

This makes EntityDB ideal for:
- Systems requiring audit trails
- Applications needing rollback capabilities
- Compliance-heavy industries
- Debugging complex issues
- Understanding system evolution