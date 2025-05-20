# EntityDB Technical Specifications

This document provides a comprehensive technical specification of the EntityDB platform, detailing the architecture, data formats, APIs, and performance characteristics.

## Data Model

### Entity Model

The core data unit in EntityDB is the Entity, defined as:

```go
type Entity struct {
    ID      string   // Unique identifier (36-byte UUID)
    Tags    []string // Temporal tags with timestamps
    Content []byte   // Binary content (autochunked)
}
```

Each entity has three fundamental components:
1. **ID**: Unique 36-byte UUID that identifies the entity
2. **Tags**: Collection of timestamped strings for classification and querying
3. **Content**: Binary data of any size (automatically chunked if >4MB)

### Tag Format

All tags are stored with nanosecond-precision timestamps in the format:

```
TIMESTAMP|tag
```

Example: `2025-05-20T12:34:56.123456789Z|type:document`

Tags follow a hierarchical namespace pattern:
- `type:` - Entity classification (user, document, config, etc.)
- `id:` - Unique identifiers within a type (username, email, etc.)
- `status:` - Entity state (active, pending, deleted, etc.)
- `rbac:` - Role-based access control tags
- `content:` - Content metadata
- `conf:` - Configuration settings
- `feat:` - Feature flags
- `rel:` - Relationship information
- `meta:` - General metadata
- `app:` - Application-specific tags

### Relationship Model

Relationships between entities are stored as first-class entities with special properties:

```go
type EntityRelationship struct {
    ID         string            // Unique identifier
    SourceID   string            // ID of source entity
    TargetID   string            // ID of target entity
    Type       string            // Relationship type
    Properties map[string]string // Additional properties
    CreatedAt  time.Time         // Creation timestamp
}
```

## Storage Format

### Binary Format (EBF)

EntityDB uses a custom binary format (EntityDB Binary Format - EBF) with the following characteristics:

#### File Structure
- 8-byte magic header: "ENTITYDB"
- 4-byte version marker
- 4-byte flags field
- Entity records (variable length)
- Index section (variable length)
- 8-byte footer with checksum

#### Entity Record Format
- 36-byte entity ID
- 4-byte length marker
- 4-byte tag count
- Tag entries (each with 8-byte timestamp + variable length tag)
- 4-byte content length
- Content data (variable length)

#### Temporal Index Format
- B-tree structure for efficient timeline queries
- Each node contains:
  - Timestamp range
  - Pointers to child nodes
  - Entity ID references within range
  - 4-byte node count

### Write-Ahead Log (WAL)

The WAL provides durability and crash recovery:

- 8-byte WAL magic header: "ENTITYWAL"
- 4-byte WAL version marker
- 8-byte sequence number
- Transaction entries:
  - 8-byte transaction ID
  - 4-byte operation type (CREATE, UPDATE, DELETE)
  - Entity record (same format as main storage)
  - 8-byte transaction checksum

### Autochunking

Large content (>4MB by default) is automatically split into chunks:

1. Parent entity contains metadata and no content
2. Child chunk entities contain content segments
3. Chunks are related to parent via `parent:` tag
4. Indexed access allows for efficient streaming

## API Specifications

### Entity API

#### Create Entity
- **Endpoint**: `POST /api/v1/entities/create`
- **Authentication**: Required (`rbac:perm:entity:create`)
- **Request Body**:
  ```json
  {
    "tags": ["type:document", "status:active"],
    "content": "Base64EncodedContent or Plain Text"
  }
  ```
- **Response**:
  ```json
  {
    "status": "ok",
    "entity": {
      "id": "entity_uuid",
      "tags": ["type:document", "status:active"],
      "created_at": "2025-05-20T12:34:56Z"
    }
  }
  ```

#### Get Entity
- **Endpoint**: `GET /api/v1/entities/get?id=entity_uuid`
- **Authentication**: Required (`rbac:perm:entity:view`)
- **Response**:
  ```json
  {
    "status": "ok",
    "entity": {
      "id": "entity_uuid",
      "tags": ["type:document", "status:active"],
      "content": "Base64EncodedContent",
      "created_at": "2025-05-20T12:34:56Z",
      "updated_at": "2025-05-20T12:34:56Z"
    }
  }
  ```

#### List Entities
- **Endpoint**: `GET /api/v1/entities/list?tag=type:document&limit=10&offset=0`
- **Authentication**: Required (`rbac:perm:entity:view`)
- **Parameters**:
  - `tag`: Filter by specific tag
  - `tags`: Multiple tags (comma separated)
  - `limit`: Maximum number of results (default: 100)
  - `offset`: Pagination offset (default: 0)
  - `sort`: Sort field (created_at, updated_at, id)
  - `order`: Sort order (asc, desc)
- **Response**:
  ```json
  {
    "status": "ok",
    "entities": [
      {
        "id": "entity_uuid",
        "tags": ["type:document", "status:active"],
        "created_at": "2025-05-20T12:34:56Z"
      }
    ],
    "count": 1,
    "total": 10
  }
  ```

#### Update Entity
- **Endpoint**: `PUT /api/v1/entities/update`
- **Authentication**: Required (`rbac:perm:entity:update`)
- **Request Body**:
  ```json
  {
    "id": "entity_uuid",
    "tags": ["type:document", "status:archived"],
    "content": "Base64EncodedContent"
  }
  ```
- **Response**:
  ```json
  {
    "status": "ok",
    "entity": {
      "id": "entity_uuid",
      "tags": ["type:document", "status:archived"],
      "updated_at": "2025-05-20T12:34:56Z"
    }
  }
  ```

### Temporal API

#### Get Entity As-Of
- **Endpoint**: `GET /api/v1/entities/as-of`
- **Authentication**: Required (`rbac:perm:entity:view`)
- **Request Body**:
  ```json
  {
    "id": "entity_uuid",
    "timestamp": "2025-05-01T00:00:00Z"
  }
  ```
- **Response**: Entity as it existed at specified time

#### Get Entity History
- **Endpoint**: `GET /api/v1/entities/history?id=entity_uuid&limit=10`
- **Authentication**: Required (`rbac:perm:entity:view`)
- **Response**: List of entity changes with timestamps

#### Get Entity Diff
- **Endpoint**: `GET /api/v1/entities/diff`
- **Authentication**: Required (`rbac:perm:entity:view`)
- **Request Body**:
  ```json
  {
    "id": "entity_uuid",
    "start_time": "2025-05-01T00:00:00Z",
    "end_time": "2025-05-20T00:00:00Z"
  }
  ```
- **Response**: Differences between entity states at the specified times

### Relationship API

#### Create Relationship
- **Endpoint**: `POST /api/v1/entity-relationships`
- **Authentication**: Required (`rbac:perm:relation:create`)
- **Request Body**:
  ```json
  {
    "source_id": "entity_uuid_1",
    "target_id": "entity_uuid_2",
    "type": "parent"
  }
  ```
- **Response**: Created relationship

#### Get Relationships
- **Endpoint**: `GET /api/v1/entity-relationships?source=entity_uuid`
- **Authentication**: Required (`rbac:perm:relation:view`)
- **Parameters**:
  - `source`: Filter by source entity ID
  - `target`: Filter by target entity ID
  - `type`: Filter by relationship type
- **Response**: List of matching relationships

## Authentication & Authorization

### Authentication

EntityDB uses JWT-based authentication:

- **Login Endpoint**: `POST /api/v1/auth/login`
- **Token Format**: JSON Web Token (JWT)
- **Token Expiration**: 24 hours by default
- **Token Refresh**: Supported via `/api/v1/auth/refresh`

### RBAC System

Role-Based Access Control is implemented via entity tags:

- **Roles**: Assigned via `rbac:role:*` tags (admin, user, etc.)
- **Permissions**: Granted via `rbac:perm:*` tags
- **Permission Hierarchy**: Supports wildcards and hierarchical permissions
  - `rbac:perm:*` - All permissions
  - `rbac:perm:entity:*` - All entity operations
  - `rbac:perm:entity:view` - View entities

## Performance Specifications

### Query Performance

| Operation Type              | Dataset Size | Average Response Time | Throughput (QPS) |
|-----------------------------|--------------|------------------------|------------------|
| Simple entity retrieval     | 1M entities  | 0.5ms                  | 2000             |
| Tag-filtered list           | 1M entities  | 1.2ms                  | 830              |
| Temporal query (as-of)      | 1M entities  | 0.8ms                  | 1250             |
| Full-text content search    | 1M entities  | 5.0ms                  | 200              |
| Relationship query          | 1M entities  | 1.0ms                  | 1000             |

### Storage Efficiency

| Entity Count | Total Size | Index Overhead | Content Storage | Tag Storage |
|--------------|------------|----------------|-----------------|-------------|
| 1M (text)    | ~500MB     | ~50MB (10%)    | ~400MB (80%)    | ~50MB (10%) |
| 1M (binary)  | ~1.5GB     | ~150MB (10%)   | ~1.2GB (80%)    | ~150MB (10%)|

### Scaling Characteristics

EntityDB scales nearly linearly with entity count:

- **Query Time**: O(log n) with B-tree indexing
- **Storage**: O(n) linear scaling with entity count
- **Memory Usage**: Configurable with memory-mapped files
- **Concurrency**: Multi-reader, single-writer architecture

## Configuration System

Configuration uses a hierarchical system:

1. Command Line Flags
2. Environment Variables
3. Instance Config File
4. Default Config File
5. Hardcoded Defaults

### Environment Variables

All configuration options can be set via environment variables:

| Variable                 | Default       | Description                        |
|--------------------------|---------------|------------------------------------|
| ENTITYDB_PORT            | 8085          | HTTP port                          |
| ENTITYDB_SSL_PORT        | 8443          | HTTPS port                         |
| ENTITYDB_USE_SSL         | true          | Enable SSL                         |
| ENTITYDB_SSL_CERT        | cert.pem      | SSL certificate path               |
| ENTITYDB_SSL_KEY         | key.pem       | SSL key path                       |
| ENTITYDB_DATA_PATH       | var/          | Data directory                     |
| ENTITYDB_LOG_LEVEL       | info          | Log level                          |
| ENTITYDB_CHUNK_SIZE      | 4194304       | Chunk size in bytes (4MB)          |
| ENTITYDB_CHUNK_THRESHOLD | 4194304       | Auto-chunking threshold (4MB)      |
| ENTITYDB_INDEX_CACHE     | 1000          | Index cache size (entities)        |
| ENTITYDB_ENABLE_CACHE    | true          | Enable in-memory caching           |

## System Limits

| Resource                    | Limit                             | Notes                                   |
|-----------------------------|-----------------------------------|----------------------------------------|
| Maximum entity size         | Limited by available memory       | Large entities auto-chunked            |
| Maximum number of entities  | No hard limit                     | Performance degrades beyond 100M       |
| Maximum tags per entity     | No hard limit                     | Performance optimal <1000 tags         |
| Maximum query result size   | Configurable (default: 1000)      | Pagination recommended for large sets  |
| Maximum content chunk size  | 4MB default (configurable)        | Larger files split into multiple chunks|
| Concurrent connections      | Limited by system resources       | Typically thousands per instance       |
| Query complexity            | Limited for performance reasons   | Complex queries may timeout            |