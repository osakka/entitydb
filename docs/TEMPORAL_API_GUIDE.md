# EntityDB Temporal API Guide

EntityDB provides comprehensive temporal query capabilities through its REST API and CLI. This guide covers all temporal endpoints and their usage.

## API Endpoints

### 1. Get Entity As Of Time

Retrieve an entity as it existed at a specific point in time.

**Endpoint**: `GET /api/v1/entities/as-of`

**Parameters**:
- `id` (required): Entity ID
- `as_of` (required): Timestamp in RFC3339 format

**Example**:
```bash
curl "http://localhost:8085/api/v1/entities/as-of?id=ent_123&as_of=2025-01-15T10:00:00Z"
```

**Response**:
```json
{
  "id": "ent_123",
  "tags": ["type:user", "status:active"],
  "content": []
}
```

### 2. Get Entity History

Retrieve the history of changes for an entity within a time range.

**Endpoint**: `GET /api/v1/entities/history`

**Parameters**:
- `id` (required): Entity ID
- `from` (optional): Start timestamp (default: 24 hours ago)
- `to` (optional): End timestamp (default: now)

**Example**:
```bash
curl "http://localhost:8085/api/v1/entities/history?id=ent_123&from=2025-01-01T00:00:00Z&to=2025-01-31T23:59:59Z"
```

**Response**:
```json
[
  {
    "id": "ent_123",
    "tags": ["type:user", "status:pending"],
    "timestamp": "2025-01-01T10:00:00Z"
  },
  {
    "id": "ent_123",
    "tags": ["type:user", "status:active"],
    "timestamp": "2025-01-15T14:30:00Z"
  }
]
```

### 3. Get Recent Changes

Find all entities that have changed since a specific time.

**Endpoint**: `GET /api/v1/entities/changes`

**Parameters**:
- `since` (optional): Timestamp in RFC3339 format (default: 1 hour ago)

**Example**:
```bash
curl "http://localhost:8085/api/v1/entities/changes?since=2025-05-17T10:00:00Z"
```

**Response**:
```json
[
  {
    "id": "ent_123",
    "tags": ["2025-05-17T10:30:00.123456789.status=updated", "status:updated"],
    "content": []
  }
]
```

### 4. Get Entity Diff

Compare an entity between two points in time.

**Endpoint**: `GET /api/v1/entities/diff`

**Parameters**:
- `id` (required): Entity ID
- `t1` (required): First timestamp 
- `t2` (required): Second timestamp

**Example**:
```bash
curl "http://localhost:8085/api/v1/entities/diff?id=ent_123&t1=2025-01-01T00:00:00Z&t2=2025-01-15T00:00:00Z"
```

**Response**:
```json
[
  {
    "entity_id": "ent_123",
    "timestamp": "2025-01-15T00:00:00Z",
    "type": "added",
    "tag": "priority:high",
    "new_value": "high"
  },
  {
    "entity_id": "ent_123",
    "timestamp": "2025-01-15T00:00:00Z",
    "type": "modified",
    "tag": "status:active",
    "old_value": "pending",
    "new_value": "active"
  }
]
```

## CLI Commands

The EntityDB CLI provides convenient access to all temporal features.

### Setup

```bash
# Login first to get authentication token
./entitydb-cli login admin password
```

### Temporal Commands

#### 1. Entity As Of

```bash
# Get entity state at specific time
./entitydb-cli temporal as-of ent_123 2025-01-15T10:00:00Z

# With custom output format
./entitydb-cli --format=json temporal as-of ent_123 2025-01-15T10:00:00Z
```

#### 2. Entity History

```bash
# Get full history (last 24 hours)
./entitydb-cli temporal history ent_123

# Get history for specific range
./entitydb-cli temporal history ent_123 \
  --from=2025-01-01T00:00:00Z \
  --to=2025-01-31T23:59:59Z

# Get history for last week
./entitydb-cli temporal history ent_123 \
  --from=$(date -d '7 days ago' -u +%Y-%m-%dT%H:%M:%SZ)
```

#### 3. Recent Changes

```bash
# Get changes in last hour (default)
./entitydb-cli temporal changes

# Get changes since specific time
./entitydb-cli temporal changes --since=2025-05-17T10:00:00Z

# Get changes for today
./entitydb-cli temporal changes \
  --since=$(date -u +%Y-%m-%dT00:00:00Z)
```

#### 4. Entity Diff

```bash
# Compare entity between two times
./entitydb-cli temporal diff ent_123 \
  2025-01-01T00:00:00Z \
  2025-01-15T00:00:00Z

# Compare before and after a change
BEFORE="2025-01-15T09:59:59Z"
AFTER="2025-01-15T10:00:01Z"
./entitydb-cli temporal diff ent_123 $BEFORE $AFTER
```

## Use Cases

### 1. Debugging Issues

Find when a problem started:

```bash
# See recent changes to system config
./entitydb-cli temporal changes --since=1h | grep config

# Get config state before issue
./entitydb-cli temporal as-of config_prod 2025-05-17T09:00:00Z
```

### 2. Audit Trail

Track who changed what when:

```bash
# Get complete history for an entity
./entitydb-cli temporal history sensitive_doc_123

# See all changes by looking at diffs
./entitydb-cli temporal diff sensitive_doc_123 \
  2025-01-01T00:00:00Z \
  $(date -u +%Y-%m-%dT%H:%M:%SZ)
```

### 3. Rollback Changes

Restore previous state:

```bash
# Get entity state from before bad change
OLD_STATE=$(./entitydb-cli temporal as-of config_123 2025-01-15T10:00:00Z)

# Could implement restore (future feature)
# ./entitydb-cli entity restore config_123 --to=2025-01-15T10:00:00Z
```

### 4. Compliance Reporting

Generate audit reports:

```bash
# Get all changes for Q1
./entitydb-cli temporal history financial_entity \
  --from=2025-01-01T00:00:00Z \
  --to=2025-03-31T23:59:59Z > q1_audit.json

# Find all entities changed in time period
./entitydb-cli temporal changes \
  --since=2025-01-01T00:00:00Z > q1_changes.json
```

## Performance Tips

1. **Use Specific Time Ranges**: Always provide `from` and `to` parameters when possible to limit data retrieval.

2. **Cache Results**: Temporal queries return immutable data, so results can be cached safely.

3. **Batch Requests**: When checking multiple entities, batch requests when possible.

4. **Index by Time**: The binary format indexes by timestamp for fast temporal queries.

## Error Handling

Common errors and solutions:

1. **Invalid Timestamp Format**: Use RFC3339 format (e.g., `2025-01-15T10:00:00Z`)

2. **Entity Not Found**: Check if entity existed at the requested time

3. **No Changes in Range**: Normal if entity wasn't modified in the time range

4. **Permission Denied**: Ensure you have read permissions for the entity

## Future Enhancements

Planned temporal features:

1. **Restore Command**: Restore entity to previous state
2. **Temporal Aggregations**: Count changes over time
3. **Change Notifications**: Real-time alerts on changes
4. **Temporal Constraints**: Enforce rules based on history
5. **Time-based Triggers**: Execute actions at specific times

## Technical Details

### Timestamp Format

All timestamps use RFC3339 format with nanosecond precision:
- Format: `YYYY-MM-DDTHH:MM:SS.nnnnnnnnnZ`
- Example: `2025-05-17T18:30:45.123456789Z`

### Tag Format

Temporal tags include timestamp prefix:
- Format: `timestamp.namespace=value`
- Example: `2025-05-17T18:30:45.123456789.status=active`

### Performance

- Temporal queries use in-memory indexes for O(log n) performance
- Binary format optimized for time-based access patterns
- Tag compression reduces storage overhead

## Troubleshooting

### Debug Mode

Enable debug output for troubleshooting:

```bash
./entitydb-cli --debug temporal as-of ent_123 2025-01-15T10:00:00Z
```

### Check Server Logs

Temporal query errors are logged:

```bash
tail -f /opt/entitydb/var/log/entitydb.log | grep temporal
```

### Verify Entity Exists

Check if entity exists before temporal queries:

```bash
./entitydb-cli entity get --id=ent_123
```

## Summary

EntityDB's temporal features provide powerful time-travel capabilities:

- Query any entity at any point in time
- Track complete change history
- Compare states between timestamps
- Find recent modifications
- Built-in audit trail with nanosecond precision

These features make EntityDB ideal for applications requiring:
- Regulatory compliance
- Audit trails
- Debugging capabilities
- Change tracking
- Historical analysis