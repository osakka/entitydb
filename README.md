<p align="center">
  <img src="share/resources/logo_white.svg" alt="EntityDB Logo" width="400">
</p>

<p align="center">High-Performance Temporal Entity Database Architecture</p>

<p align="center">
  <strong>100x Faster Queries</strong> • 
  <strong>Nanosecond Precision</strong> • 
  <strong>Unlimited Content Size</strong> • 
  <strong>Time Travel Queries</strong>
</p>

## What is EntityDB?

EntityDB is a high-performance temporal database where every tag is timestamped with nanosecond precision. It features a pure entity-based architecture with everything represented as entities with tags.

- **Temporal Database:** Every change is tracked with nanosecond-precision timestamps
- **Binary Storage Format:** Custom binary format (EBF) with Write-Ahead Logging 
- **Autochunking:** Unlimited file sizes with automatic splitting across entities
- **Memory-Mapped Files:** Zero-copy reads with OS-managed caching
- **Advanced Indexing:** B-tree timeline, skip-lists, bloom filters

## Key Features

- ⚡ **100x Performance:** Temporal storage with optimized binary format
- 🧩 **Unified Entity Model:** Everything is an entity with tags
- 📝 **Content Streaming:** No RAM limits with automatic chunking
- 🔒 **RBAC Enforcement:** Tag-based permission system
- ⏱️ **Time Travel:** Query any entity at any point in time
- 🔄 **Entity Relationships:** Native relationship support

## Quick Start

```bash
# Clone repository
git clone https://git.home.arpa/itdlabs/entitydb.git
cd entitydb

# Build the server
cd src
make
cd ..

# Start the server
./bin/entitydbd.sh start

# Access web UI
# Default: https://localhost:8085 (credentials: admin/admin)
```

## API Examples

```bash
# Login and get token
TOKEN=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Create entity
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:document", "project:demo"],
    "content": "This is a test document"
  }'

# Query entities by tag
curl -k -X GET "https://localhost:8085/api/v1/entities/list?tag=type:document" \
  -H "Authorization: Bearer $TOKEN"

# Time travel query (as-of)
curl -k -X GET https://localhost:8085/api/v1/entities/as-of \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "entity_123",
    "timestamp": "2023-01-01T00:00:00Z"
  }'
```

## Architecture

EntityDB is built on a pure entity-based architecture:

```
┌─────────────────────────────────────┐
│            EntityDB API             │
└─────────────────────────────────────┘
               │
┌─────────────────────────────────────┐
│         RBAC Authorization          │
└─────────────────────────────────────┘
               │
┌─────────────────────────────────────┐
│        Temporal Repository          │
├─────────────────────────────────────┤
│  B-tree │ Skip-list │ Bloom Filter  │
└─────────────────────────────────────┘
               │
┌─────────────────────────────────────┐
│      Binary Storage Format (EBF)    │
├─────────────────────────────────────┤
│       Write-Ahead Log (WAL)         │
└─────────────────────────────────────┘
               │
┌─────────────────────────────────────┐
│      Memory-Mapped File Access      │
└─────────────────────────────────────┘
```

## Building & Development

```bash
# Build server
cd /opt/entitydb/src
make

# Run all tests
make test

# Start server in development mode
cd ..
./bin/entitydbd.sh start

# Stop server
./bin/entitydbd.sh stop
```

## Configuration

EntityDB uses a hierarchical configuration system:

1. Command Line Flags
2. Environment Variables
3. Instance Config (`/opt/entitydb/var/entitydb.env`)
4. Default Config (`/opt/entitydb/share/config/entitydb_server.env`)
5. Hardcoded Defaults

Key configuration options:
- `ENTITYDB_PORT`: HTTP port (default: 8085)
- `ENTITYDB_USE_SSL`: Enable SSL (default: true)
- `ENTITYDB_DATA_PATH`: Data directory (default: /opt/entitydb/var)
- `ENTITYDB_LOG_LEVEL`: Logging level (default: info)

## Performance

With the temporal storage engine, EntityDB achieves breakthrough performance:

| Dataset Size | Query Time | Throughput    |
|--------------|------------|--------------|
| 1M entities  | 0.5ms      | 2000 op/sec  |
| 5M entities  | 1.5ms      | 670 op/sec   |
| 10M entities | 3ms        | 333 op/sec   |

Memory usage remains minimal due to memory-mapped files with OS-level caching.

## Documentation

Detailed documentation is available in the [docs](./docs) directory:

- [API Guide](./docs/api)
- [Architecture](./docs/architecture)
- [Development Guide](./docs/development)
- [Release Notes](./docs/releases)

## Version History

- **v2.13.1** - Content format standardization and API testing framework
- **v2.13.0** - Configuration system overhaul and content encoding fixes
- **v2.12.0** - Unified Entity model with autochunking
- **v2.11.0** - Temporal repository with 100x performance
- **v2.10.0** - Binary format with SSL-only mode
- **v2.9.0** - RBAC system implementation
- **v2.8.0** - Feature flag system

## Project Structure

```
/opt/entitydb/
├── bin/         # Executable binaries and scripts
├── docs/        # Documentation
├── share/       # Shared resources and tools
├── src/         # Source code
├── trash/       # Retired code (keep for reference)
└── var/         # Variable data (database, logs)
```

> **Development Convention:** Always move unused, outdated, or deprecated code to the `/trash` directory instead of deleting it. This preserves reference implementations while keeping the main codebase clean.

## Repository

https://git.home.arpa/itdlabs/entitydb

## License

MIT