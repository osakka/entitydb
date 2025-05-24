# Tag Index Persistence Bug

## Issue
When EntityDB restarts, the tag index is not persisted or rebuilt, causing queries that depend on tags to return empty results even though the entities exist.

## Symptoms
- After restart, dataspace queries return 0 entities
- Direct tag queries work for newly created entities but not for entities created before restart
- MetDataspace dashboard shows blank panels because metrics queries return no data

## Root Cause
The `tagIndex` map in `EntityRepository` is built in memory as entities are created but:
1. It's not persisted to disk
2. It's not rebuilt from WAL or data files on startup
3. Only newly created entities after restart get added to the index

## Debug Evidence
```
2025/05/24 10:36:01.199666 [EntityDB] DEBUG: [ListByTags] ListByTags: Looking for tag 'dataspace:metrics' in index
2025/05/24 10:36:01.199707 [EntityDB] DEBUG: [ListByTags] ListByTags: Tag 'dataspace:metrics' not found in index
```

## Workaround
Currently, the only workaround is to recreate all entities after restart, which is not practical.

## Fix Required
The EntityRepository needs to:
1. Either persist the tag index to disk (in a .idx file)
2. Or rebuild the tag index from all entities on startup
3. Or ensure the WAL replay rebuilds the index

## Impact
This affects all tag-based queries including:
- Dataspace queries
- RBAC permission checks
- Entity relationship queries
- Any query using ListByTags