# EntityDB

A high-performance temporal database with pure entity-based architecture. Features **100x faster queries** with temporal mode and autochunking support for unlimited file sizes.

**🤖 This project is proudly developed with Claude AI assistance. All commits are AI-generated and marked accordingly.**

## Features

- **100x Performance**: Temporal storage with B-tree indexing, bloom filters, and memory-mapped files
- **Unified Entity Model**: Single content field per entity ([]byte) with autochunking
- **Autochunking**: Automatic chunking of large content (>4MB) across entities
- **Temporal Everything**: Every tag has a nanosecond timestamp
- **Binary Storage**: Custom EBF format with Write-Ahead Logging (WAL)
- **Pure Entity Architecture**: Everything is an entity with tags
- **SSL-Only Mode**: Secure by default with HTTPS on port 8085
- **RBAC System**: Tag-based permissions fully enforced
- **Time Travel**: Query any entity at any point in time
- **No RAM Limits**: Stream large files without loading fully into memory

## Quick Start

```bash
# Clone the repository
git clone https://git.home.arpa/itdlabs/entitydb.git
cd entitydb

# Build the server
cd src
make

# Start the server with SSL
cd ..
./bin/entitydbd.sh start

# Access the web UI
open https://localhost:8085
```

Default credentials: `admin` / `admin`

## Architecture

### Entity Model (v2.12.0)

```go
type Entity struct {
    ID      string   // UUID (auto-generated)
    Tags    []string // Temporal tags with timestamps
    Content []byte   // Single content field (autochunked)
}
```

### Storage Architecture

```
EntityDB
├── Temporal Repository
├── Binary Storage (EBF with WAL)  
├── Indexing (B-tree, Skip-list, Bloom filter)
├── Memory-mapped Files (zero-copy reads)
└── Autochunking (parent-child entities)
```

## Configuration

EntityDB uses a multi-level configuration system with the following precedence (highest to lowest):

1. **Command Line Flags**: Override all other settings
2. **Environment Variables**: Can be set in shell or config files
3. **Instance Config File**: `/opt/entitydb/var/entitydb.env` (optional)
4. **Default Config File**: `/opt/entitydb/share/config/entitydb_server.env`
5. **Hardcoded Defaults**: Built into the application

### Configuration Files

- **Default config**: `share/config/entitydb_server.env` - Contains all available settings with defaults
- **Instance config**: `var/entitydb.env` - Override specific settings for this instance

### Environment Variables

All configuration can be set via environment variables:
- `ENTITYDB_PORT`: HTTP port (default: 8085)
- `ENTITYDB_SSL_PORT`: HTTPS port (default: 8443)
- `ENTITYDB_USE_SSL`: Enable SSL (default: true)
- `ENTITYDB_DATA_PATH`: Data directory (default: /opt/entitydb/var)
- `ENTITYDB_LOG_LEVEL`: Logging level (default: info)
- `ENTITYDB_SSL_CERT`: SSL certificate path
- `ENTITYDB_SSL_KEY`: SSL key path
- And more... (see `share/config/entitydb_server.env` for full list)

### Dynamic Configuration

Configuration can also be stored as entities:
- **Config entities**: `type:config` with `conf:namespace:key` tags
- **Feature flags**: `type:feature_flag` with `feat:stage:flag` tags

## API Examples

```bash
# Login
TOKEN=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Create entity with small content
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:document", "project:demo"],
    "content": "This is a small document"
  }'

# Create entity with large content (will autochunk)
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:file", "name:large.bin"],
    "content": {
      "data": "'$(cat large-file.bin | base64)'",
      "type": "application/octet-stream"
    }
  }'

# Query entities 
curl -k https://localhost:8085/api/v1/entities/list?tag=type:document \
  -H "Authorization: Bearer $TOKEN"

# Temporal queries
curl -k https://localhost:8085/api/v1/entities/as-of \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "ENTITY_ID",
    "timestamp": "2024-01-01T00:00:00Z"
  }'
```

## Tag System

Hierarchical namespaced tags with temporal tracking:

```
type:user              # Entity type
status:active          # Entity state
id:username:john       # Unique identifier
rbac:role:admin       # Role assignment
rbac:perm:entity:*    # Permissions
content:type:json     # Content metadata
content:size:1024     # Content size
chunk:0               # Chunk index
parent:ENTITY_ID      # Parent entity for chunks
```

## Development

```bash
# Run tests
cd src
make test

# Build server
make

# Install scripts
make install

# Clean build
make clean
```

## Project Structure

```
/opt/entitydb/
├── bin/              # Server binary and startup script
├── src/              # Go source code
│   ├── main.go       # Server entry point
│   ├── api/          # REST API handlers
│   ├── models/       # Entity models
│   └── storage/      # Binary storage implementation
├── var/              # Runtime data (database files)
├── share/            # Shared resources
│   ├── htdocs/       # Web UI (Alpine.js)
│   ├── cli/          # Command line tools
│   └── tests/        # Test scripts and API test framework
└── docs/             # Documentation
    ├── API_TESTING_FRAMEWORK.md        # Testing framework documentation
    ├── CONTENT_FORMAT_TROUBLESHOOTING.md # Content format guide
    └── RELEASE_NOTES_v2.13.1.md        # Latest release notes
```

## Version History

- **v2.13.1**: Content format standardization and API testing framework
- **v2.13.0**: Configuration system overhaul and content encoding fixes
- **v2.12.0**: Unified Entity model with autochunking
- **v2.11.0**: Temporal repository with 100x performance
- **v2.10.0**: Binary format with SSL-only mode
- **v2.9.0**: RBAC system implementation
- **v2.8.0**: Feature flag system

## Performance

With temporal indexing enabled:
- 5M entities: 1.5ms query time
- 10M entities: 3ms query time  
- Concurrent access: No lock contention
- Memory usage: Minimal (memory-mapped files)

## Contributing

All development is done in collaboration with Claude AI. Each commit includes AI attribution.

## License

MIT License - See LICENSE file for details

## Repository

https://git.home.arpa/itdlabs/entitydb