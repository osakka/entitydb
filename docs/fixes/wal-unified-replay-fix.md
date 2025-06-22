# WAL Unified File Replay Fix

**Date**: 2025-06-22  
**Version**: v2.34.2  
**Type**: Critical Bug Fix  
**Component**: WAL (Write-Ahead Log)

## Issue Description

The server crashed during entity CRUD testing when attempting to replay the WAL during checkpoint operations. The crash occurred immediately after logging "Starting WAL replay from /opt/entitydb/var/entities.edb".

## Root Cause Analysis

The root cause was in the `Replay()` method in `/opt/entitydb/src/storage/binary/wal.go`. The method was seeking to position 0 (start of file) for all WAL types:

```go
// Seek to the beginning
if _, err := w.file.Seek(0, io.SeekStart); err != nil {
    // ...
}
```

However, in the unified file format (.edb), the WAL section is embedded within the file at a specific offset (stored in `w.walOffset`). Seeking to position 0 would read the file header instead of WAL entries, causing the replay to fail when trying to interpret header bytes as WAL entry lengths.

## Solution

The fix properly handles unified vs standalone WAL files by seeking to the correct position:

```go
// Seek to the beginning of WAL section
seekPos := int64(0)
if w.isUnified {
    seekPos = int64(w.walOffset)
    logger.Debug("Seeking to WAL section at offset %d in unified file", seekPos)
}

if _, err := w.file.Seek(seekPos, io.SeekStart); err != nil {
    op.Fail(err)
    logger.Error("Failed to seek to WAL position %d: %v", seekPos, err)
    return err
}
```

## Implementation Details

1. **Detection**: Check if the WAL is unified (`w.isUnified`)
2. **Offset Calculation**: Use `w.walOffset` for unified files, 0 for standalone
3. **Logging**: Added debug logging to track seek operations
4. **Error Handling**: Enhanced error messages to include seek position

## Testing and Verification

After applying the fix:

1. Server started successfully
2. Entity CRUD operations completed without crashes
3. Concurrent updates handled properly
4. WAL checkpointing works correctly

## Principles Upheld

1. **Single Source of Truth**: Modified only the necessary code in the WAL replay method
2. **No Parallel Implementations**: Used existing `isUnified` and `walOffset` fields
3. **No Workarounds**: Proper fix addressing the root cause
4. **Bar Raising**: Added debug logging for better observability
5. **No Regressions**: Standalone WAL files continue to work (seekPos = 0)

## Impact

- **Severity**: Critical - Prevented server operation
- **Scope**: Affected all operations that trigger WAL replay (checkpoints)
- **Resolution**: Complete - Server now stable

## Lessons Learned

The unified file format requires careful handling of offsets throughout the codebase. Any operation that seeks within the file must be aware of section boundaries and use appropriate offsets rather than assuming traditional file layouts.