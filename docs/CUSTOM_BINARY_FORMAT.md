# EntityDB Custom Binary Format Specification

## Overview

EntityDB Binary Format (EBF) is designed specifically for EntityDB's entity-tag-content model. It provides:
- Fast sequential and random access
- Efficient tag-based indexing
- Simple append-only operations
- Built-in compression for tags
- Minimal overhead

## File Structure

```
[Header (64 bytes)]
[Tag Dictionary]
[Entity Index]
[Entity Data Block 1]
[Entity Data Block 2]
...
[Entity Data Block N]
```

## Detailed Format

### 1. Header (64 bytes)

```
Offset  Size  Description
0       4     Magic number (0x45424446) "EBDF"
4       4     Format version (1)
8       8     Total file size
16      8     Tag dictionary offset
24      8     Tag dictionary size
32      8     Entity index offset
40      8     Entity index size
48      8     Number of entities
56      8     Last modified timestamp
```

### 2. Tag Dictionary

Compressed dictionary of all unique tag strings to reduce storage:

```
[Count: 4 bytes]
[Tag Entry 1]
[Tag Entry 2]
...

Tag Entry:
[ID: 4 bytes]
[Length: 2 bytes]
[String: N bytes]
```

### 3. Entity Index

Fixed-size index for O(1) entity lookups:

```
[Entity Entry 1: 32 bytes]
[Entity Entry 2: 32 bytes]
...

Entity Entry:
[Entity ID: 16 bytes]
[Offset: 8 bytes]
[Size: 4 bytes]
[Flags: 4 bytes]
```

### 4. Entity Data Block

```
[Entity Header: 16 bytes]
  - Modified timestamp: 8 bytes
  - Tag count: 2 bytes
  - Content count: 2 bytes
  - Reserved: 4 bytes

[Tags Section]
  - Tag ID: 4 bytes (references dictionary)
  - Tag ID: 4 bytes
  ...

[Content Section]
  - Content Entry 1
  - Content Entry 2
  ...

Content Entry:
  - Type length: 2 bytes
  - Type string: N bytes
  - Value length: 4 bytes
  - Value: N bytes
  - Timestamp: 8 bytes
```

## Operations

### 1. Writing

```go
type EntityWriter struct {
    file      *os.File
    tagDict   *TagDictionary
    index     *EntityIndex
    buffer    *bytes.Buffer
}

func (w *EntityWriter) WriteEntity(entity *Entity) error {
    // Compress tags using dictionary
    tagIDs := w.tagDict.GetOrCreateIDs(entity.Tags)
    
    // Write to buffer
    w.buffer.Reset()
    writeEntityHeader(w.buffer, entity)
    writeTagIDs(w.buffer, tagIDs)
    writeContent(w.buffer, entity.Content)
    
    // Append to file
    offset := w.file.Seek(0, io.SeekEnd)
    w.file.Write(w.buffer.Bytes())
    
    // Update index
    w.index.Add(entity.ID, offset, w.buffer.Len())
    
    return nil
}
```

### 2. Reading

```go
type EntityReader struct {
    file    *os.File
    tagDict *TagDictionary
    index   *EntityIndex
}

func (r *EntityReader) GetEntity(id string) (*Entity, error) {
    // Lookup in index
    entry, ok := r.index.Get(id)
    if !ok {
        return nil, ErrNotFound
    }
    
    // Seek and read
    r.file.Seek(entry.Offset, io.SeekStart)
    data := make([]byte, entry.Size)
    r.file.Read(data)
    
    // Parse entity
    return parseEntity(data, r.tagDict)
}
```

### 3. Querying

```go
type QueryEngine struct {
    reader    *EntityReader
    tagIndex  *TagIndex  // In-memory inverse index
}

func (q *QueryEngine) QueryByTag(tag string) ([]*Entity, error) {
    // Get entities from tag index
    entityIDs := q.tagIndex.GetEntities(tag)
    
    // Batch read entities
    entities := make([]*Entity, 0, len(entityIDs))
    for _, id := range entityIDs {
        entity, _ := q.reader.GetEntity(id)
        entities = append(entities, entity)
    }
    
    return entities, nil
}
```

## Advantages Over SQLite

1. **Simplicity**: No SQL parsing, query planning, or B-trees
2. **Performance**: Direct memory mapping, zero-copy reads
3. **Size**: ~40% smaller due to tag compression
4. **Speed**: 10x faster entity lookups (no SQL overhead)
5. **Control**: Complete control over storage format
6. **Append-only**: Natural fit for event sourcing

## Migration Plan

### Phase 1: Implement Core
1. Binary format reader/writer
2. Tag dictionary compression
3. Entity index management

### Phase 2: Query Engine
1. Tag-based indexing
2. Wildcard support
3. Content search

### Phase 3: Integration
1. Repository implementation
2. API compatibility layer
3. Migration tool from SQLite

### Phase 4: Optimization
1. Memory-mapped files
2. Concurrent access
3. Write-ahead logging
4. Incremental indexing

## Performance Targets

- Entity write: < 100μs
- Entity read by ID: < 10μs  
- Tag query (1000 results): < 1ms
- Startup time: < 50ms
- File size: ~60% of SQLite

## File Management

### Data Files
```
/opt/entitydb/var/data/
  entities.ebf      # Main data file
  entities.idx      # Tag indexes
  entities.wal      # Write-ahead log
```

### Archival
```
/opt/entitydb/var/archive/
  entities-2025-01.ebf
  entities-2025-02.ebf
```

## Conclusion

This custom format is optimized for EntityDB's specific use case:
- Entities are immutable (append-only)
- Tag-based queries are primary access pattern
- Relationships are just tags
- No complex joins or aggregations needed

The simplicity allows for better performance and smaller storage footprint while maintaining all required functionality.