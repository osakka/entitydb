# EntityDB Temporal Tag Fix

This directory contains several scripts and files to fix an issue with temporal tags in EntityDB.

## Problem Description

EntityDB stores tags with timestamps in the format `TIMESTAMP|tag` (e.g., `2025-05-21T12:22:00.123456789Z|type:test`). However, the `ListByTag` function in `entity_repository.go` does not correctly handle this format when searching for tags, resulting in recently added entities not being found when searching by tag.

## Solution Files

1. **temporal_tag_fix.sh** - Basic test script that demonstrates the issue
2. **temporal_tag_fix_v2.sh** - Enhanced test script with better error reporting
3. **improved_temporal_fix.sh** - Full test script with proper error handling and temporal feature testing
4. **direct_temporal_fix.sh** - Script that directly modifies the source code and recompiles the server

## Fix Implementation

The fix involves modifying the `ListByTag` function in `entity_repository.go` to properly handle temporal tags. The key changes are:

1. When searching for tags, also check for tags with the format `TIMESTAMP|tag`
2. Extract the actual tag part after the timestamp separator
3. Compare with the requested tag
4. Include matching entities in the result

## How to Apply the Fix

### Option 1: Direct Source Code Fix (Recommended)

```bash
cd /opt/entitydb/src
./direct_temporal_fix.sh
```

This script will:
1. Backup the original file
2. Replace the `ListByTag` function with the fixed version
3. Recompile the server
4. Restart the server with the fixed code
5. Test the fix

### Option 2: Manual Fix

1. Stop the EntityDB server:
   ```bash
   /opt/entitydb/bin/entitydbd.sh stop
   ```

2. Edit the file `/opt/entitydb/src/storage/binary/entity_repository.go`

3. Replace the `ListByTag` function with the fixed implementation from `temporal_tag_fix_v2.sh`

4. Recompile the server:
   ```bash
   cd /opt/entitydb/src
   go build -o /opt/entitydb/bin/entitydb
   ```

5. Restart the server:
   ```bash
   /opt/entitydb/bin/entitydbd.sh start
   ```

## Testing the Fix

After applying the fix, run the improved test script:

```bash
cd /opt/entitydb
./improved_temporal_fix.sh
```

The test script will:
1. Create a new entity with the tag `type:test`
2. Wait for indexing to complete
3. Search for entities with that tag
4. Verify that the newly created entity is found

## Additional Fixes

The following additional files were created to handle more complex aspects of temporal tag handling:

- `/opt/entitydb/src/storage/binary/improved_temporal_fix.go` - Enhanced implementations of temporal functions
- `/opt/entitydb/src/storage/binary/patch_temporal_tags.go` - Runtime patching mechanism
- `/opt/entitydb/src/apply_temporal_fix.go` - Helper for applying fixes at server startup

## Temporal API Fixes

This fix also addresses related issues in the temporal API endpoints:
- `GetEntityAsOf`
- `GetEntityHistory`
- `GetRecentChanges`
- `GetEntityDiff`

These endpoints now properly handle temporal tags in all circumstances.