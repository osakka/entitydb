# EntityDB Documentation

EntityDB is a temporal database system with a pure entity-based architecture where all data is represented as entities with hierarchical tags, backed by a high-performance binary storage format.

## Current State (v2.3.0)

### What Works
- **Binary Persistence**: Custom binary format (EBF) with Write-Ahead Logging
- **Entity Operations**: Full CRUD for entities
- **Temporal Queries**: Time-travel queries to any point in history
- **Authentication**: Simple token-based auth
- **Dashboard UI**: Web interface for entity management
- **User Management**: Admin can create users
- **Configuration**: Config and feature flags as entities

### Architecture
- All data is entities with tags and content
- Binary storage format with nanosecond timestamps
- REST API for all operations
- No SQL, no schemas, just entities

## API Endpoints

### Entity Operations
```
GET    /api/v1/entities/list      # List entities
GET    /api/v1/entities/get       # Get by ID  
POST   /api/v1/entities/create    # Create entity
PUT    /api/v1/entities/update    # Update entity
```

### Temporal Operations
```
GET    /api/v1/entities/as-of     # Entity at timestamp
GET    /api/v1/entities/history   # Entity history
GET    /api/v1/entities/changes   # Recent changes
GET    /api/v1/entities/diff      # Compare versions
```

### Other Operations
```
POST   /api/v1/auth/login         # Login
POST   /api/v1/auth/logout        # Logout
GET    /api/v1/auth/status        # Auth status
POST   /api/v1/users/create       # Create user (admin)
GET    /api/v1/dashboard/stats    # Dashboard stats
GET    /api/v1/config             # Get config
POST   /api/v1/config/set         # Set config
GET    /api/v1/feature-flags      # Get flags
POST   /api/v1/feature-flags/set  # Set flags
```

## Tag Namespaces

- `type:` - Entity type (user, issue, workspace)
- `id:` - Unique identifiers  
- `rbac:` - Roles and permissions
- `status:` - Entity state
- `meta:` - Metadata
- `rel:` - Relationships
- `conf:` - Configuration
- `feat:` - Feature flags

## Quick Start

```bash
# Start server
cd /opt/entitydb
./bin/entitydbd.sh start

# Login (default: admin/admin - auto-created on first start)
./share/cli/entitydb-cli login admin admin

# Create entity
./share/cli/entitydb-cli entity create \
  --type=issue \
  --title="My Issue" \
  --tags="priority:high,status:pending"

# List entities
./share/cli/entitydb-cli entity list --tag="type:issue"
```

## Files & Structure

```
/opt/entitydb/
├── bin/                # Core executables
│   ├── entitydb        # Server binary
│   └── entitydbd.sh    # Daemon script (handles admin init)
├── src/                # Source code
│   ├── main.go         # Server implementation
│   ├── api/            # API handlers
│   ├── models/         # Entity models
│   └── storage/        # Binary storage
├── var/                # Runtime data
│   ├── entities.ebf    # Entity database
│   └── entitydb.wal    # Write-ahead log
└── share/              # Shared resources
    ├── cli/            # CLI tools
    ├── tests/          # Test scripts
    ├── utilities/      # Utility programs
    └── htdocs/         # Web UI
```

## Implementation Notes

- Authentication is handled directly in main.go
- No middleware or permission enforcement yet
- Binary format supports concurrent reads/writes
- All data stored in /opt/entitydb/var/db/binary/