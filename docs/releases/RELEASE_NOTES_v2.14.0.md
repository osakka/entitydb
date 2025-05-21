# EntityDB Release Notes - v2.14.0

## Overview

Version 2.14.0 focuses on stability, performance, and large file handling. This release introduces a high-performance mode with memory-mapped file support and enhanced autochunking capabilities for handling large files without memory constraints.

## Key Features

### High-Performance Mode

- **Memory-Mapped Files**: Zero-copy reads with OS-managed caching
- **Reduced Memory Footprint**: Stream large files without loading fully into RAM
- **Performance Option**: Enable with `--high-performance` flag or `ENTITYDB_HIGH_PERFORMANCE=true`
- **Optimized Indexing**: Faster lookups with improved B-tree and skip-list implementation
- **Parallel Query Support**: Concurrent query execution for improved throughput

### Enhanced Autochunking

- **Improved Content Handling**: Better detection and handling of content types
- **Streaming Upload/Download**: Process large files in chunks without full memory loading
- **Configurable Chunk Size**: Default 4MB chunk size, configurable via environment
- **Chunk Validation**: Integrity checks with SHA-256 hashing
- **Transparent API**: API remains unchanged despite underlying chunking

### Temporal Tag Fixes

- **Temporal Tag Indexing**: Fixed issue with temporal tag indexing for improved search reliability
- **Tag Lookups**: Enhanced handling of timestamp prefixes in tag queries
- **Performance Optimizations**: Faster tag-based entity lookup
- **Debug Logging**: Added extensive debug logging for troubleshooting

## Additional Improvements

- **Content Type Handling**: Improved detection and handling of content types
- **Base64 Support**: Better support for base64-encoded content
- **API Error Reporting**: Enhanced error reporting and debugging
- **Documentation**: Expanded architectural documentation
- **Test Suite**: Added comprehensive tests for large files and chunking capabilities

## Breaking Changes

None. All changes maintain backward compatibility with existing data and APIs.

## API Changes

No API-level changes. All improvements are transparent to API users.

## Configuration Changes

- **New Environment Variables**:
  - `ENTITYDB_HIGH_PERFORMANCE`: Set to "true" to enable memory-mapped files (default: "false")
  - `ENTITYDB_CHUNK_SIZE`: Set the chunk size in bytes for large files (default: 4194304 [4MB])

- **New Command Line Flags**:
  - `--high-performance`: Enable high-performance mode
  - `--chunk-size`: Set the chunk size in bytes

## Known Issues

- High-performance mode requires sufficient system memory for optimal operation
- The SSL implementation still needs improvements for automatic certificate renewal

## Upgrade Instructions

1. Stop the EntityDB server:
   ```bash
   /opt/entitydb/bin/entitydbd.sh stop
   ```

2. Backup your data:
   ```bash
   cp -r /opt/entitydb/var /opt/entitydb/var_backup
   ```

3. Update to the new version:
   ```bash
   git pull
   cd /opt/entitydb/src
   make clean
   make
   make install
   ```

4. Start the server:
   ```bash
   /opt/entitydb/bin/entitydbd.sh start
   ```

5. Verify the upgrade:
   ```bash
   curl -k https://localhost:8443/api/v1/status
   ```

## Contributors

- Development Team at ITD Labs