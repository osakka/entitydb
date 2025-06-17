# EntityDB

> A high-performance temporal database where every tag is timestamped with nanosecond precision

[![Version](https://img.shields.io/badge/version-v2.32.0%20🚀%20Battle%20Tested-blue)](./CLAUDE.md)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)
[![Documentation](https://img.shields.io/badge/docs-comprehensive-brightgreen)](./docs)
[![API Coverage](https://img.shields.io/badge/API%20docs-100%25%20accurate-success)](./docs/api-reference)

## What is EntityDB?

EntityDB is a revolutionary temporal database platform that stores all data as entities with timestamped tags. Built with a custom binary format (EBF) and Write-Ahead Logging, it provides ACID compliance, time-travel queries, and enterprise-grade RBAC.

> **🚀 NEW in v2.32.0**: Production battle-tested across 5 comprehensive real-world scenarios. Critical security vulnerability in multi-tag queries fixed (OR→AND logic). Performance optimizations achieving 60%+ improvement in complex queries. Complete temporal database functionality with nanosecond precision, comprehensive RBAC integration, and enterprise-grade security.

> **⚠️ BREAKING CHANGE in v2.29.0**: Authentication architecture has changed. User credentials are now stored directly in the user entity's content field. This change has **NO BACKWARD COMPATIBILITY** - all users must be recreated. See [Authentication Guide](./docs/api-reference/02-authentication.md) for details.

### Key Features

- ⏱️ **Temporal Storage**: Every tag timestamped with nanosecond precision
- 🏢 **Dataset Isolation**: Complete multi-tenancy with isolated namespaces
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
# Web UI: https://localhost:8085 (SSL enabled by default)
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

### Dataset Isolation

Complete multi-tenancy through isolated datasets:
```bash
# Create a dataset
curl -k -X POST https://localhost:8085/api/v1/datasets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"myapp","description":"My Application"}'

# Create entity in dataset
curl -k -X POST https://localhost:8085/api/v1/datasets/myapp/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags":["type:task"],"content":"Task data"}'
```

## API Overview

### Authentication
```bash
# Login
TOKEN=$(curl -s -k -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')
```

### Entity Operations
```bash
# Create entity
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
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

- [Quick Start Guide](./docs/getting-started/03-quick-start.md)
- [API Reference](./docs/api-reference/03-entities.md)
- [Architecture Overview](./docs/architecture/01-system-overview.md)
- [RBAC & Security](./docs/admin-guide/01-security-configuration.md)
- [Development Guide](./docs/developer-guide/01-contributing.md)
- [Performance Tuning](./docs/reference/performance/performance_optimization_report.md)

## Project Structure

```
entitydb/
├── bin/                    # Server binaries and scripts
├── docs/                   # Comprehensive documentation
├── share/                  # Web assets and configuration
│   ├── config/            # Default configuration
│   └── htdocs/            # Web UI and dashboard
├── src/                   # Source code
│   ├── api/               # HTTP API handlers
│   ├── models/            # Core data models
│   ├── storage/binary/    # Binary storage engine
│   └── tests/             # Test suites
└── var/                   # Runtime data (database, logs)
```

## Configuration

EntityDB uses a comprehensive three-tier configuration system:

1. **Database Configuration Entities** (highest priority)
2. **CLI Flags**
3. **Environment Variables** (lowest priority)

### Environment Variables

```bash
# Server Configuration
ENTITYDB_PORT=8085                    # HTTP server port (when SSL disabled)
ENTITYDB_SSL_PORT=8085               # HTTPS server port (when SSL enabled)
ENTITYDB_USE_SSL=true                # Enable SSL/TLS (true by default)

# Paths
ENTITYDB_DATA_PATH=/opt/entitydb/var # Database storage path
ENTITYDB_STATIC_DIR=/opt/entitydb/share/htdocs # Web files path

# Timeouts
ENTITYDB_HTTP_READ_TIMEOUT=15        # HTTP read timeout (seconds)
ENTITYDB_METRICS_INTERVAL=30         # Metrics collection interval

# See docs/configuration-management.md for complete reference
```

### Configuration Files

- **Default**: `/opt/entitydb/share/config/entitydb.env`
- **Instance**: `/opt/entitydb/var/entitydb.env` (overrides defaults)

For complete configuration documentation, see [Configuration Management Guide](./docs/60-developer-guides/04-configuration-management.md).

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

See [Contributing Guide](./docs/60-developer-guides/01-contributing.md) for development guidelines.

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Links

- **Repository**: https://git.home.arpa/itdlabs/entitydb
- **Issues**: https://git.home.arpa/itdlabs/entitydb/issues
- **Documentation**: [./docs](./docs)
- **Changelog**: [CHANGELOG.md](./CHANGELOG.md)

---

Built with ❤️ by ITDLabs