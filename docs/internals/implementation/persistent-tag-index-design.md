# Persistent Tag Index Design

## Overview
Implement a persistent tag index stored in a `.idx` file alongside the `.ebf` data file. This will dramatically improve startup time by avoiding the need to rebuild the entire tag index from scratch.

## File Format Design

### Index File Structure
```
[Header]
- Magic Number: 4 bytes ("TIDX")
- Version: 2 bytes (uint16)
- Entry Count: 8 bytes (uint64)
- Checksum: 32 bytes (SHA256 of index data)

[Index Entries]
For each tag:
- Tag Length: 4 bytes (uint32)
- Tag: N bytes (UTF-8 string)
- Entity Count: 4 bytes (uint32)
- Entity IDs: 36 bytes * count (UUIDs)

[Footer]
- End Marker: 4 bytes ("ENDT")
```

### Benefits
1. **Fast Startup**: Load pre-built index instead of scanning all entities
2. **Consistency**: Index stays in sync with data file
3. **Verification**: Checksum ensures index integrity
4. **Atomic Updates**: Write new index file and atomically rename

## Implementation Plan

### Phase 1: Index Writer
```go
type TagIndexWriter struct {
    file *os.File
    hasher hash.Hash
}

func (w *TagIndexWriter) WriteHeader(version uint16, entryCount uint64) error
func (w *TagIndexWriter) WriteEntry(tag string, entityIDs []string) error
func (w *TagIndexWriter) WriteFooter() error
func (w *TagIndexWriter) Close() error
```

### Phase 2: Index Reader
```go
type TagIndexReader struct {
    file *os.File
    header IndexHeader
}

func (r *TagIndexReader) ReadHeader() (IndexHeader, error)
func (r *TagIndexReader) ReadAllEntries() (map[string][]string, error)
func (r *TagIndexReader) VerifyChecksum() error
func (r *TagIndexReader) Close() error
```

### Phase 3: Integration
1. On startup:
   - Try to load `.idx` file
   - Verify checksum
   - If valid, use it to populate tag index
   - If invalid/missing, rebuild from data file and save new index

2. On entity operations:
   - Update in-memory index
   - Mark index as dirty
   - Periodically persist to disk

3. On shutdown:
   - If index is dirty, save to disk

### Phase 4: Optimization
1. **Incremental Updates**: Track changes and update only modified sections
2. **Compression**: Use simple compression for entity ID lists
3. **Bloom Filters**: Add bloom filter for quick negative lookups
4. **Parallel Loading**: Load index in parallel with data file

## File Locations
- Data file: `entity_YYYYMMDD_HHMMSS.ebf`
- Index file: `entity_YYYYMMDD_HHMMSS.idx`
- Temp index: `entity_YYYYMMDD_HHMMSS.idx.tmp`

## Performance Targets
- Index load time: < 100ms for 1M tags
- Index save time: < 1s for 1M tags
- Memory overhead: ~50 bytes per tag entry

## Error Handling
1. Corrupted index: Log warning, rebuild from data
2. Version mismatch: Rebuild for new version
3. Missing index: Normal operation, build on first start
4. Write failure: Keep in-memory index, retry later