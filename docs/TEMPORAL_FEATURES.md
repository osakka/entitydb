# EntityDB Temporal Features

## Overview

EntityDB implements a revolutionary temporal data model where every tag is timestamped, providing complete time-travel capabilities and event sourcing out of the box.

## Tag Format

Each tag in EntityDB has two representations:
1. **Temporal Format**: `2025-05-17T18:00:58.001351096.type=user`
2. **Simple Format**: `type:user`

The temporal format captures the exact nanosecond when the tag was added, making EntityDB a truly temporal database.

## Key Benefits

### 1. Time-Travel Queries
```bash
# Get entity state at a specific time
entitydb-cli entity get --id=ent_123 --at="2025-01-15T10:00:00"

# Find all changes in a time range
entitydb-cli entity history --id=ent_123 --from="2025-01-01" --to="2025-01-31"
```

### 2. Natural Audit Trail
Every tag change is automatically audited:
- Who made the change (from JWT token)
- When it was made (nanosecond precision)
- What changed (tag name and value)
- No separate audit tables needed

### 3. Event Sourcing
- Each tag is an immutable event
- Entity state is the sum of all events
- Can replay history to any point
- Perfect for debugging and compliance

### 4. Temporal Relationships
```
# Track relationship changes over time
2025-01-01T10:00:00.000000000.rel:assigned_to=user_123
2025-01-15T14:30:00.000000000.rel:assigned_to=user_456
```

### 5. Version Control Built-In
- Natural versioning through timestamps
- Can diff entities between timestamps
- Rollback to previous states
- Branch and merge entity states

## Implementation Examples

### Temporal Tag Structure
```go
// When adding a tag
timestamp := GenerateTimestamp() // "2025-05-17T18:00:58.001351096"
temporalTag := fmt.Sprintf("%s.%s=%s", timestamp, namespace, value)
```

### Query Examples

#### Find Entity State at Timestamp
```go
func GetEntityAtTime(entityID string, timestamp time.Time) *Entity {
    entity, _ := repo.GetByID(entityID)
    
    // Filter tags to those before timestamp
    activeTags := []string{}
    for _, tag := range entity.Tags {
        tagTime := extractTimestamp(tag)
        if tagTime.Before(timestamp) {
            activeTags = append(activeTags, tag)
        }
    }
    
    // Reconstruct entity at that time
    entity.Tags = activeTags
    return entity
}
```

#### Track Tag Changes
```go
func GetTagHistory(entityID, tagName string) []TagChange {
    entity, _ := repo.GetByID(entityID)
    
    changes := []TagChange{}
    for _, tag := range entity.Tags {
        if strings.Contains(tag, tagName) {
            changes = append(changes, parseTagChange(tag))
        }
    }
    
    // Sort by timestamp
    sort.Slice(changes, func(i, j int) bool {
        return changes[i].Timestamp.Before(changes[j].Timestamp)
    })
    
    return changes
}
```

## Advanced Features

### 1. Temporal Indexes
The binary format can maintain temporal indexes:
- Time-based partitioning
- Efficient range queries
- Quick historical lookups

### 2. Point-in-Time Recovery
```bash
# Restore entity to specific time
entitydb-cli entity restore --id=ent_123 --to="2025-01-15T10:00:00"
```

### 3. Temporal Constraints
```go
// Ensure tags are unique per timestamp
func (e *Entity) AddTemporalTag(namespace, value string) {
    timestamp := GenerateTimestamp()
    
    // Check if tag exists at this timestamp
    for _, tag := range e.Tags {
        if extractTimestamp(tag) == timestamp &&
           extractNamespace(tag) == namespace {
            // Wait 1 nanosecond and retry
            time.Sleep(1)
            return e.AddTemporalTag(namespace, value)
        }
    }
    
    temporalTag := fmt.Sprintf("%s.%s=%s", timestamp, namespace, value)
    e.Tags = append(e.Tags, temporalTag)
}
```

## Performance Considerations

### Binary Format Optimizations
1. **Temporal Compression**: Group tags by time ranges
2. **Time-based Indexing**: Partition by day/hour
3. **Lazy Loading**: Load only relevant time ranges
4. **Incremental Updates**: Append-only for new timestamps

### Query Optimizations
1. **Time Range Filters**: Skip irrelevant data blocks
2. **Temporal Caching**: Cache frequently accessed time periods
3. **Parallel Processing**: Process time ranges concurrently

## Use Cases

### 1. Compliance & Auditing
- Complete audit trail for all changes
- Prove compliance at any point in time
- Forensic analysis of data changes

### 2. Debugging & Troubleshooting
- See exactly when issues started
- Track configuration changes
- Replay scenarios for debugging

### 3. Analytics & Reporting
- Historical trend analysis
- Point-in-time reports
- Change frequency analysis

### 4. Data Recovery
- Restore accidental deletions
- Rollback bad changes
- Maintain data integrity

## Future Enhancements

1. **Temporal Query Language**
   ```sql
   SELECT * FROM entities 
   WHERE type = 'user' 
   AS OF '2025-01-15 10:00:00'
   ```

2. **Time-based Triggers**
   - Execute actions at specific times
   - Schedule future tag changes
   - Temporal workflows

3. **Historical Diffs**
   - Visual diff between time periods
   - Change notifications
   - Automated change summaries

4. **Temporal Joins**
   - Join entities as they were at specific times
   - Historical relationship analysis
   - Time-based aggregations

## Conclusion

EntityDB's temporal features make it unique among databases:
- Every change is captured with nanosecond precision
- Complete history is maintained automatically
- Time-travel queries are natural and efficient
- No additional infrastructure needed for auditing

This makes EntityDB perfect for:
- Financial systems requiring audit trails
- Healthcare systems with compliance needs
- Configuration management with rollback
- Any system where history matters