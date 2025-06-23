# ADR-001: Temporal Tag Storage with Nanosecond Precision

## Status
Accepted (2025-05-08)

## Context
EntityDB needed a way to store temporal data with high precision for time-travel queries and historical analysis. We evaluated several approaches:

1. **Separate timestamp fields**: Store timestamps in separate entity fields
2. **Tag prefixes**: Store temporal information as tag prefixes
3. **Database-level timestamping**: Rely on database-level temporal features
4. **Tag-embedded timestamps**: Embed timestamps directly in tag values

### Requirements
- Nanosecond precision for temporal queries
- Efficient storage and indexing
- Backward compatibility during migration
- Simple API for temporal operations

### Constraints
- Must work with binary storage format
- Performance critical for large datasets
- Need to support multiple timestamp formats during transition

## Decision
We decided to implement **tag-embedded timestamps** using the format:
```
TIMESTAMP|tag_value
```

Where:
- `TIMESTAMP` is nanoseconds since Unix epoch
- `|` is the delimiter
- `tag_value` is the original tag content

### Implementation Details
- All tags are stored with timestamps automatically
- API returns tags without timestamps by default
- `include_timestamps=true` parameter exposes full temporal format
- Temporal queries use specialized repository methods
- Supports both RFC3339 and epoch nanosecond formats during transition

## Consequences

### Positive
- **High precision**: Nanosecond-level temporal resolution
- **Unified storage**: All data naturally temporal without separate fields
- **Efficient queries**: Direct tag-based temporal indexing
- **Simple migration**: Gradual transition from non-temporal system
- **API transparency**: Users see clean tags unless requesting timestamps

### Negative
- **Storage overhead**: Each tag carries timestamp data
- **Migration complexity**: Required updating all existing tags
- **Format coupling**: Tag format tightly coupled to temporal implementation
- **Breaking changes**: No backward compatibility for pre-v2.8.0 data

### Risks Mitigated
- **Performance**: Optimized indexing prevents temporal query slowdown
- **Complexity**: Transparent API hides implementation details from users
- **Data loss**: Comprehensive migration tools ensure data preservation

## Implementation History
- v2.7.0: Initial temporal tag implementation
- v2.8.0: Full temporal-only system deployment
- v2.30.0: Fixed temporal tag search and indexing issues
- v2.32.2: Complete temporal functionality with all endpoints operational

## Related Decisions
- [ADR-002: Binary Storage Format](./002-binary-storage-format.md) - Storage layer foundation
- [ADR-003: Unified Sharded Indexing](./003-unified-sharded-indexing.md) - Indexing optimization