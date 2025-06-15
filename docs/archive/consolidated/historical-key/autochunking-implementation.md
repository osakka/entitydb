# Autochunking Implementation Documentation

## Overview

This document describes the autochunking feature implemented in EntityDB v2.12.0. The feature automatically splits large files into smaller chunks for efficient storage and memory management.

## Architecture

### Entity Model Changes

The Entity model has been simplified to use a single byte array for content:

```go
type Entity struct {
    ID        string   `json:"id"`
    Tags      []string `json:"tags"`
    Content   []byte   `json:"content,omitempty"`
    CreatedAt string   `json:"created_at,omitempty"`
    UpdatedAt string   `json:"updated_at,omitempty"`
}
```

### Autochunking Configuration

Default configuration (4MB chunks):

```go
type ChunkConfig struct {
    DefaultChunkSize   int64 // 4MB
    AutoChunkThreshold int64 // Files > 4MB get chunked
}
```

### How It Works

1. When content > 4MB is uploaded, the system automatically splits it into chunks
2. A parent entity is created with metadata tags but no content
3. Child chunk entities are created, each containing a portion of the data
4. Metadata tags track:
   - Number of chunks: `content:chunks:N`
   - Chunk size: `content:chunk-size:4194304`
   - Content type: `content:type:application/json`
   - Total size: `content:size:BYTES`
   - SHA256 checksum: `content:checksum:sha256:HASH`

### Chunk Entity Structure

Each chunk entity has:
- ID: `{parent-id}-chunk-{index}`
- Tags:
  - `type:chunk`
  - `parent:{parent-id}`
  - `content:chunk:{index}`
  - `content:size:{bytes}`
  - `content:checksum:sha256:{hash}`
- Content: Actual chunk data (up to 4MB)

## API Changes

### Content Handling

The API now accepts flexible content formats:
- String: `{"content": "plain text"}`
- JSON Object: `{"content": {"key": "value"}}`
- JSON Array: `{"content": [1, 2, 3]}`

Large content is automatically chunked when necessary.

### Response Format

Entities with chunked content return:
- Parent entity with chunk metadata tags
- No content in the parent entity
- Chunk references in tags

## Database Storage

All data is stored in `/opt/entitydb/var/`:
- `entities.ebf` - Binary entity file
- `entitydb.wal` - Write-ahead log
- No subdirectories needed

## Performance Benefits

1. **Memory Efficiency**: Never loads entire large files into RAM
2. **Streaming Support**: Can process files of unlimited size
3. **Parallel Processing**: Chunks can be processed independently
4. **Incremental Updates**: Individual chunks can be updated

## Migration Notes

- Clean cut-off approach - no backward compatibility
- Previous multi-content array model completely replaced
- Fresh database required (no migration path)

## Examples

### Small File (< 4MB)
```bash
curl -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "tags": ["type:document"],
    "content": "Small content stored directly"
  }'
```

### Large File (> 4MB)
```bash
# Automatically chunked into multiple entities
curl -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "tags": ["type:document"],
    "content": "...5MB of content..."
  }'

# Response includes chunk metadata:
{
  "id": "abc123",
  "tags": [
    "type:document",
    "content:type:text/plain",
    "content:chunks:2",
    "content:size:5242880",
    "content:checksum:sha256:..."
  ]
}
```

## Future Enhancements

1. Configurable chunk size per entity
2. Compression support for chunks
3. Parallel chunk upload/download
4. Chunk deduplication
5. Incremental chunk updates