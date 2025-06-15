# EntityDB Autochunking

EntityDB automatically chunks large content to handle files of any size efficiently.

## How It Works

1. **Small Files (≤ 4MB)**: Stored directly in the entity
2. **Large Files (> 4MB)**: Automatically chunked into 4MB pieces

## Configuration

```go
config := DefaultChunkConfig()
// or customize:
config := ChunkConfig{
    DefaultChunkSize:   10 * 1024 * 1024, // 10MB chunks
    AutoChunkThreshold: 1 * 1024 * 1024,  // Chunk files > 1MB
}
```

Environment variables:
```bash
ENTITYDB_CHUNK_SIZE=10485760      # 10MB chunks
ENTITYDB_AUTO_CHUNK=1048576       # Auto-chunk files > 1MB
```

## Examples

### Small File (Direct Storage)
```go
entity := NewEntitySimple()
entity.Tags = []string{"type:document", "name:readme.txt"}

file, _ := os.Open("readme.txt")
defer file.Close()

chunkIDs, _ := entity.SetContent(file, "text/plain", DefaultChunkConfig())
// chunkIDs is nil for small files

// Entity looks like:
{
  "id": "abc123",
  "tags": [
    "type:document",
    "name:readme.txt",
    "content:type:text/plain",
    "content:size:1024",
    "content:checksum:sha256:..."
  ],
  "content": "SGVsbG8gd29ybGQ..."  // Base64 encoded
}
```

### Large File (Auto-chunked)
```go
entity := NewEntitySimple()
entity.Tags = []string{"type:video", "name:movie.mp4"}

file, _ := os.Open("movie.mp4") // 5GB file
defer file.Close()

chunkIDs, _ := entity.SetContent(file, "video/mp4", DefaultChunkConfig())
// chunkIDs contains IDs of all chunk entities to create

// Master entity:
{
  "id": "video123",
  "tags": [
    "type:video",
    "name:movie.mp4",
    "content:type:video/mp4",
    "content:size:5368709120",
    "content:chunks:1280",
    "content:chunk-size:4194304",
    "content:checksum:sha256:..."
  ],
  "content": null
}

// Chunk entities (created separately):
{
  "id": "video123-chunk-0",
  "tags": [
    "type:chunk",
    "parent:video123",
    "content:chunk:0",
    "content:size:4194304",
    "content:checksum:sha256:..."
  ],
  "content": "..."  // 4MB of data
}
```

## Reading Chunked Content

```go
func StreamContent(repo Repository, entityID string, writer io.Writer) error {
    entity, _ := repo.GetByID(entityID)
    
    if !entity.IsChunked() {
        // Direct write for small files
        writer.Write(entity.Content)
        return nil
    }
    
    // Stream chunks for large files
    metadata := entity.GetContentMetadata()
    chunks, _ := strconv.Atoi(metadata["chunks"])
    
    for i := 0; i < chunks; i++ {
        chunkID := fmt.Sprintf("%s-chunk-%d", entityID, i)
        chunk, _ := repo.GetByID(chunkID)
        writer.Write(chunk.Content)
    }
    
    return nil
}
```

## Benefits

1. **Unlimited File Size**: Only limited by storage
2. **Memory Efficient**: Never loads full file in RAM
3. **Streaming Support**: Read/write in chunks
4. **Parallel Processing**: Upload/download chunks concurrently
5. **Resume Support**: Track completed chunks
6. **Transparent**: Small files work exactly as before

## Performance

- Default 4MB chunks balance memory usage and I/O efficiency
- Chunks can be processed in parallel
- Deduplication possible via chunk checksums
- Progressive loading for media streaming