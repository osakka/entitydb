# EntityDB

> A high-performance temporal database where every tag is timestamped with nanosecond precision

[![Version](https://img.shields.io/badge/version-v2.20.0-blue)](./CHANGELOG.md)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)
[![Documentation](https://img.shields.io/badge/docs-comprehensive-brightgreen)](./docs)

## What is EntityDB?

EntityDB is a revolutionary temporal database platform that stores all data as entities with timestamped tags. Built with a custom binary format (EBF) and Write-Ahead Logging, it provides ACID compliance, time-travel queries, and enterprise-grade RBAC.

### Key Features

- ⏱️ **Temporal Storage**: Every tag timestamped with nanosecond precision
- 🏢 **Dataspace Isolation**: Complete multi-tenancy with isolated namespaces
- 🔒 **Enterprise RBAC**: Tag-based permissions with fine-grained access control
- 📦 **Autochunking**: Handle files of any size without memory limits
- 🚀 **High Performance**: Memory-mapped files, B-tree indexes, bloom filters
- 🔍 **Time Travel**: Query any entity state at any point in history
- 📊 **Real-time Metrics**: Comprehensive monitoring with temporal storage
- 🧩 **Entity Relationships**: Native support for complex data relationships

## Quick Start

```bash
# Clone the repository
git clone https://git.home.arpa/itdlabs/entitydb.git
cd entitydb

# Build the server
cd src && make && cd ..

# Start the server (creates admin/admin user automatically)
./bin/entitydbd.sh start

# Access the dashboard
# Web UI: https://localhost:8085
# Default credentials: admin/admin
```

## Core Concepts

### Everything is an Entity

In EntityDB, all data is stored as entities with:
- **ID**: Unique identifier
- **Tags**: Timestamped key-value pairs
- **Content**: Binary data (automatically chunked if >4MB)

### Temporal Tags

Every tag is stored with a nanosecond timestamp:
```
1748544372255000000|type:user
1748544372255000000|status:active
1748544372285000000|status:inactive
```

### Dataspace Isolation

Complete multi-tenancy through isolated dataspaces:
```bash
# Create a dataspace
curl -k -X POST https://localhost:8085/api/v1/dataspaces/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"myapp","description":"My Application"}'

# Create entity in dataspace
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"tags":["dataspace:myapp","type:task"],"content":"Task data"}'
```

## API Overview

### Authentication
```bash
# Login
TOKEN=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')
```

### Entity Operations
```bash
# Create entity
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"tags":["type:document","status:draft"],"content":"My document"}'

# Query entities
curl -k -X GET "https://localhost:8085/api/v1/entities/query?tags=type:document" \
  -H "Authorization: Bearer $TOKEN"

# Get entity history
curl -k -X GET "https://localhost:8085/api/v1/entities/history?id=ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN"
```

### Temporal Queries
```bash
# Get entity state at specific time
curl -k -X GET "https://localhost:8085/api/v1/entities/as-of?id=ID&timestamp=2024-01-01T00:00:00Z" \
  -H "Authorization: Bearer $TOKEN"

# Get changes between times
curl -k -X GET "https://localhost:8085/api/v1/entities/diff?id=ID&from=T1&to=T2" \
  -H "Authorization: Bearer $TOKEN"
```

## Architecture

EntityDB uses a layered architecture optimized for temporal operations:

```
┌─────────────────────────────────────────┐
│         API Layer (REST/JSON)           │
├─────────────────────────────────────────┤
│     RBAC Middleware & Security          │
├─────────────────────────────────────────┤
│       Temporal Repository Layer         │
├─────────────────────────────────────────┤
│     Binary Storage Engine (EBF)         │
├─────────────────────────────────────────┤
│   WAL │ Indexes │ Memory-Mapped Files   │
└─────────────────────────────────────────┘
```

## Documentation

- [Quick Start Guide](./docs/guides/quick-start.md)
- [API Reference](./docs/api/README.md)
- [Architecture Overview](./docs/architecture/overview.md)
- [RBAC & Security](./docs/guides/security.md)
- [Development Guide](./docs/development/contributing.md)
- [Performance Tuning](./docs/performance/README.md)

## Project Structure

```
entitydb/
├── bin/                    # Server binaries and scripts
├── docs/                   # Comprehensive documentation
├── share/                  # Web assets and configuration
│   ├── config/            # Default configuration
│   └── htdocs/            # Web UI and applications
│       └── worca/         # Workforce orchestrator demo
├── src/                   # Source code
│   ├── api/               # HTTP API handlers
│   ├── models/            # Core data models
│   ├── storage/binary/    # Binary storage engine
│   └── tests/             # Test suites
└── var/                   # Runtime data (database, logs)
```

## Applications

### Worca - Workforce Orchestrator

A complete project management application demonstrating EntityDB capabilities:
- Hierarchical task management
- Real-time Kanban boards
- Team collaboration features
- Temporal audit trails

Access at: https://localhost:8085/worca/

## Performance

EntityDB achieves exceptional performance through:
- Memory-mapped file access
- B-tree temporal indexes
- Bloom filters for tag queries
- WAL with automatic checkpointing
- Query result caching

Benchmarks show:
- 100,000+ entities/second write throughput
- Sub-millisecond temporal queries
- Linear scaling with proper indexing

## Contributing

See [CONTRIBUTING.md](./docs/core/contributing/CONTRIBUTING.md) for development guidelines.

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Links

- **Repository**: https://git.home.arpa/itdlabs/entitydb
- **Issues**: https://git.home.arpa/itdlabs/entitydb/issues
- **Documentation**: [./docs](./docs)
- **Changelog**: [CHANGELOG.md](./CHANGELOG.md)

---

Built with ❤️ by ITDLabs