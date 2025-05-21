<p align="center">
  <img src="share/resources/logo_white.svg" alt="EntityDB Logo" width="400">
</p>

<p align="center">High-Performance Temporal Entity Database Architecture</p>

<p align="center">
  <strong>RESTful API</strong> • 
  <strong>Temporal Database</strong> • 
  <strong>Entity Relationship Model</strong> • 
  <strong>Chunked Content Handler</strong> • 
  <strong>Transactional Operations</strong>
</p>

## What is EntityDB?

EntityDB is a high-performance temporal database where every tag is timestamped with nanosecond precision. It features a pure entity-based architecture with everything represented as entities with tags.

- **Temporal Database:** Every change is tracked with nanosecond-precision timestamps
- **Binary Storage Format:** Custom binary format (EBF) with Write-Ahead Logging 
- **Autochunking:** Unlimited file sizes with automatic splitting across entities
- **Memory-Mapped Files:** Zero-copy reads with OS-managed caching
- **Advanced Indexing:** B-tree timeline, skip-lists, bloom filters

## Key Features

- 🔄 **RESTful API:** Complete HTTP API with JSON request/response format
- ⏱️ **Temporal Storage:** Nanosecond precision timestamps on all entity tags
- 🧩 **Entity Relationship Model:** Pure entity architecture with native relationship support
- 📝 **Chunked Content Handling:** Unlimited content size with automatic chunking
- 💾 **Transactional Operations:** ACID compliance via Write-Ahead Logging
- 🔒 **RBAC Enforcement:** Tag-based permission system with fine-grained access control
- 🔍 **Time Travel Queries:** View entity state at any point in history

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
# Default: https://localhost:8443 (credentials: admin/admin)
```

## API Examples

```bash
# Login and get token
TOKEN=$(curl -k -s -X POST https://localhost:8443/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Create entity
curl -k -X POST https://localhost:8443/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:document", "project:demo"],
    "content": "This is a test document"
  }'

# Query entities by tag
curl -k -X GET "https://localhost:8443/api/v1/entities/list?tag=type:document" \
  -H "Authorization: Bearer $TOKEN"

# Time travel query (as-of)
curl -k -X GET https://localhost:8443/api/v1/entities/as-of \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "entity_123",
    "timestamp": "2023-01-01T00:00:00Z"
  }'
```

## Architecture

EntityDB is built on a pure entity-based architecture with layered components:

<p align="center">
  <img src="share/resources/architecture.svg" alt="EntityDB Architecture" width="500">
</p>

## Building & Development

```bash
# Build server
cd /opt/entitydb/src
make

# Run all tests with timing metrics
cd ../src/tests
./run_tests.sh --clean --login --all --timing

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

EntityDB is designed for efficient operation with large datasets:

| Dataset Size | Average Query Time | Estimated Throughput |
|--------------|-------------------|---------------------|
| 10K entities | 5-15ms            | 150-300 op/sec     |
| 100K entities | 15-30ms          | 75-150 op/sec      |
| 1M entities   | 30-100ms         | 30-75 op/sec       |

Actual performance varies based on hardware, query complexity, and entity relationships.
The memory-mapped file architecture helps maintain reasonable memory usage even with
large datasets.

## Testing

EntityDB uses a simple shell-based test framework for API testing with performance metrics:

```bash
# Run all tests with timing
cd /opt/entitydb/src/tests
./run_tests.sh --clean --login --all --timing

# Create a new test
./run_tests.sh --new my_test POST endpoint "Description"

# See test creation guide
less cases/README.md

# Run a specific test
./run_tests.sh --login create_entity

# Run temporal API tests
./test_temporal_api.sh
```

The test framework is based on request/response pairs:
- Each test has a `*_request` file defining the API call 
- Each test has a `*_response` file defining validation criteria
- Tests can be chained together for complex scenarios
- No external dependencies required - pure shell implementation

For more details, see the [Testing Framework Documentation](/src/tests/README.md).

## Documentation

Detailed documentation is available in the [docs](./docs) directory:

- [API Guide](./docs/api)
- [Architecture](./docs/architecture)
- [Development Guide](./docs/development)
- [Testing Framework](/src/tests/README.md)
- [Release Notes](./docs/releases)

## Version History

- **v2.14.0** - High-performance mode and enhanced autochunking capabilities
- **v2.13.1** - Content format standardization and API testing framework
- **v2.13.0** - Configuration system overhaul and content encoding fixes
- **v2.12.0** - Unified Entity model with autochunking
- **v2.11.0** - Temporal repository implementation
- **v2.10.0** - Binary format with SSL-only mode
- **v2.9.0** - RBAC system implementation
- **v2.8.0** - Feature flag system

## Project Structure

```
/opt/entitydb/
├── bin/         # Executable binaries and scripts
├── docs/        # Documentation
├── share/       # Shared resources (configs, web assets)
├── src/         # Source code
│   ├── tools/   # Command-line tools
│   │   ├── users/     # User management tools
│   │   ├── entities/  # Entity management tools
│   │   ├── maintenance/ # System maintenance tools
│   │   ├── diagnostics/ # Debugging and performance tools
│   │   ├── content/     # Content handling tools
│   │   └── temporal/    # Temporal data management tools
│   └── tests/   # Test framework and test cases
├── trash/       # Retired code (keep for reference)
└── var/         # Variable data (database, logs)
```

> **Development Conventions:** 
> 1. Always move unused or outdated code to the `/trash` directory instead of deleting it.
> 2. Add new command-line tools to the appropriate category in `/src/tools/`.
> 3. All command-line tools follow the `entitydb_` naming convention.

## Repository

https://git.home.arpa/itdlabs/entitydb

## License

MIT