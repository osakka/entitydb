# ADR-029: Entity Deletion Index Tracking

## Status
Accepted (2025-06-22)

## Context
EntityDB's unified temporal deletion architecture was designed to track deleted entities through a deletion index. However, the implementation had a critical gap: while the deletion index infrastructure existed, it wasn't being properly utilized during entity deletion or queries. This resulted in "deleted" entities still being returned by API queries, causing confusion and potential data integrity issues.

The problem manifested in three areas:
1. The `Delete()` method only removed entities from in-memory indexes but didn't add them to the deletion index
2. `GetByID()` didn't check if an entity was marked as deleted
3. `List()` and `ListByTag()` operations returned deleted entities

## Decision
Implement proper deletion index tracking by:
1. Adding deletion entries to the deletion index when entities are deleted
2. Checking the deletion index in all query operations
3. Filtering out deleted entities from list operations
4. Properly cleaning up all indexes when entities are deleted

## Implementation Details

### Delete Method Enhancement
```go
// Create deletion entry
deletionEntry := &DeletionEntry{
    DeletionTimestamp: time.Now().UnixNano(),
    LifecycleState:    3, // Purged state
    Flags:             0,
}

// Add to deletion index
if r.deletionIndex != nil {
    r.deletionIndex.AddEntry(deletionEntry)
}

// Remove from all indexes
r.shardedTagIndex.RemoveTag(id, actualTag)
r.tagVariantCache.RemoveEntityFromVariant(id)
r.temporalIndex.RemoveEntity(id)
```

### Query Method Updates
```go
// GetByID checks deletion index first
if r.deletionIndex != nil {
    if _, deleted := r.deletionIndex.GetEntry(id); deleted {
        return nil, fmt.Errorf("entity %s not found", id)
    }
}

// List operations filter deleted entities
if r.deletionIndex != nil && entities != nil {
    filtered := make([]*models.Entity, 0, len(entities))
    for _, entity := range entities {
        if _, deleted := r.deletionIndex.GetEntry(entity.ID); !deleted {
            filtered = append(filtered, entity)
        }
    }
    entities = filtered
}
```

## Consequences

### Positive
- Deleted entities are now properly hidden from all API queries
- Deletion state is consistently tracked across the system
- No physical deletion required - maintains audit trail
- Clean separation between logical and physical deletion
- All deletion lifecycle states (soft_deleted, archived, purged) work correctly

### Negative
- Small performance overhead for checking deletion index on queries
- Deletion index grows over time (mitigated by retention policies)

### Neutral
- Deletion index was already initialized but unused - now properly utilized
- No changes to binary format or storage layer required

## Testing
Comprehensive test suite (`test_deletion_apis.go`) validates:
- Soft deletion with lifecycle states
- Deletion status queries
- Deleted entity filtering in list operations
- Entity restoration
- Permanent purging with confirmation
- RBAC enforcement for deletion operations

## References
- ADR-027: Database File Unification
- ADR-028: WAL Corruption Prevention
- Unified Temporal Deletion Architecture design