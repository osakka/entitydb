# Chunked Content API Guide

## Overview

EntityDB provides support for large files (>4MB) through its auto-chunking feature. This document explains how to work with chunked content through the API.

## How Chunking Works

When a file exceeds the default chunk size (4MB), EntityDB automatically:

1. Breaks the file into multiple chunks (4MB each by default)
2. Stores each chunk as a separate entity with IDs like `{parent-id}-chunk-{index}`
3. Adds metadata tags to the main entity (e.g., `content:chunks:3`, `content:chunk-size:4194304`)
4. Sets the main entity's `Content` field to empty

## API Endpoints for Chunked Content

### 1. Standard Entity Retrieval

```
GET /api/v1/entities/get?id={entity_id}&include_content=true
```

The standard entity retrieval endpoint now automatically reassembles chunked content. When `include_content=true` is specified, the system:
- Detects if the entity is chunked
- Fetches all chunks in order
- Reassembles them into the original content
- Returns the complete content in the response

### 2. Direct Streaming (Recommended for Large Files)

```
GET /api/v1/entities/stream?id={entity_id}&stream=true
```

For improved performance with large files, use the streaming endpoint. This endpoint:
- Streams chunks directly to the client without buffering the entire content
- Sets appropriate Content-Type and Content-Disposition headers
- Provides better memory efficiency for very large files

### 3. Download Endpoint

```
GET /api/v1/entities/download?id={entity_id}
```

This endpoint is an alias for the streaming endpoint, designed for direct downloads.

## Identifying Chunked Entities

An entity is chunked if it has the following tags:
- `content:chunks:{n}` - Indicates the entity has n chunks
- `content:chunk-size:{size}` - Indicates the size of each chunk in bytes
- `content:size:{total_size}` - Indicates the total size of the content

## Best Practices

1. **Use the streaming endpoint for large files** - The `/api/v1/entities/stream` endpoint is more efficient for large files as it streams chunks directly to the client.

2. **Consider headers for binary content** - When retrieving binary content, set appropriate headers in your client:
   ```
   Accept: application/octet-stream
   ```

3. **Check content size before retrieval** - Use the metadata endpoint to check the content size before deciding which retrieval method to use:
   ```
   GET /api/v1/entities/get?id={entity_id}
   ```
   Check the `content:size` tag to determine the total size.

## Implementation Details

- Each chunk entity has an ID that follows the pattern `{parent-id}-chunk-{index}`
- Chunks are stored as separate entities with the tag `type:chunk`
- Chunk entities include their own checksums for integrity validation

This implementation ensures that large files can be efficiently stored and retrieved while maintaining backward compatibility with existing API clients.