# EntityDB Platform

> [!IMPORTANT]
> EntityDB is a high-performance temporal database where every tag is timestamped with nanosecond precision. All data is stored in a custom binary format (EBF) with Write-Ahead Logging for durability and concurrent access support.

## Current State (v2.14.0)

EntityDB now features a unified Entity model with autochunking:
- **Unified Entity Model**: Single content field ([]byte) per entity
- **Autochunking**: Automatic chunking of large files (>4MB default)
- **No RAM Limits**: Stream large files without loading fully into memory  
- **Temporal Storage**: Up to 100x performance improvement
- **Advanced Indexing**: B-tree timeline, skip-lists, bloom filters
- **Memory-Mapped Files**: Zero-copy reads with OS-managed caching
- **Temporal-Only Storage**: All tags stored with nanosecond timestamps (TIMESTAMP|tag format)
- **Transparent API**: Tags returned without timestamps by default (use include_timestamps=true to see them)
- **Binary Storage**: Custom binary format (EBF) with WAL and concurrent access
- **Pure Entity Model**: Everything is an entity with tags
- **RBAC Enforcement**: Full permission system with tag-based access control
- **Entity Relationships**: Binary format supports relationships between entities
- **Auto-Initialization**: Creates admin/admin user automatically on first start
- **Query Adaptations**: All query functions handle temporal tags transparently
- **Observability**: Comprehensive metrics with health checks, Prometheus format, and system analytics

## What's Implemented

### Core Features
- Entity CRUD operations
- Binary persistence with WAL
- Temporal queries (as-of, history, diff)
- RBAC authentication and authorization (fully enforced)
- Entity relationships with efficient querying
- Dashboard UI (Alpine.js)
- User management with secure password hashing
- Configuration as entities
- Health monitoring and metrics endpoints

### Applications Built on EntityDB
- **Worcha (Workforce Orchestrator)**: Complete workforce management platform with Kanban boards, team management, project hierarchies, and real-time analytics. Features drag-drop functionality, dark/light themes, and mobile-responsive design. Located at `/share/htdocs/worcha/`

### Architecture
```
/opt/entitydb/
├── src/
│   ├── main.go                      # Server implementation
│   ├── api/                         # API handlers
│   │   ├── entity_handler.go
│   │   ├── entity_handler_rbac.go   # RBAC wrapper for entities
│   │   ├── entity_relationship_handler.go
│   │   ├── relationship_handler_rbac.go  # RBAC wrapper for relationships
│   │   ├── dashboard_handler.go
│   │   ├── user_handler.go
│   │   ├── user_handler_rbac.go     # RBAC wrapper for users
│   │   ├── entity_config_handler.go
│   │   ├── config_handler_rbac.go   # RBAC wrapper for config
│   │   ├── health_handler.go        # Health monitoring endpoint
│   │   ├── metrics_handler.go       # Prometheus metrics endpoint
│   │   ├── system_metrics_handler.go # EntityDB system metrics
│   │   └── rbac_middleware.go       # RBAC enforcement middleware
│   ├── models/                      # Entity models
│   └── storage/binary/              # Binary format implementation
│       ├── entity_repository.go
│       └── relationship_repository.go  # Binary relationships
├── bin/
│   ├── entitydb                     # Server binary
│   └── entitydbd.sh                 # Daemon script
├── share/htdocs/worcha/             # Workforce Orchestrator Application
│   ├── index.html                   # Main dashboard
│   ├── cli.html                     # Conversational CLI
│   ├── worcha.js                    # Core application logic
│   └── worcha-api.js                # EntityDB API wrapper
```

### API Endpoints
```
# Entity operations (RBAC enforced)
GET    /api/v1/entities/list         # Requires entity:view
GET    /api/v1/entities/get          # Requires entity:view
POST   /api/v1/entities/create       # Requires entity:create
PUT    /api/v1/entities/update       # Requires entity:update
GET    /api/v1/entities/query        # Advanced query with sorting/filtering

# Temporal operations (RBAC enforced)
GET    /api/v1/entities/as-of        # Requires entity:view
GET    /api/v1/entities/history      # Requires entity:view
GET    /api/v1/entities/changes      # Requires entity:view
GET    /api/v1/entities/diff         # Requires entity:view

# Relationship operations (RBAC enforced)
POST   /api/v1/entity-relationships  # Requires relation:create
GET    /api/v1/entity-relationships  # Requires relation:view

# Auth & Admin
POST   /api/v1/auth/login            # No auth required
POST   /api/v1/users/create          # Requires user:create (admin only)
GET    /api/v1/dashboard/stats       # Requires system:view
GET    /api/v1/config                # Requires config:view
POST   /api/v1/feature-flags/set     # Requires config:update

# Monitoring & Observability
GET    /health                       # Health check with system metrics (no auth)
GET    /metrics                      # Prometheus metrics format (no auth)
GET    /api/v1/system/metrics        # EntityDB comprehensive metrics (no auth)

# API Documentation
GET    /swagger/                     # Swagger UI
GET    /swagger/doc.json            # OpenAPI spec
```

## Development Guidelines

### Key Principles
1. **Everything is an Entity**: No special tables or structures
2. **Binary Storage**: All data in custom binary format
3. **Tag-Based RBAC**: Permissions are entity tags
4. **Clean Codebase**: Remove unused code immediately

### Building
```bash
cd /opt/entitydb/src
make                # Build server
make install        # Install scripts
```

### Running
```bash
./bin/entitydbd.sh start   # Start server (auto-creates admin/admin if needed)
./bin/entitydbd.sh status  # Check status
./bin/entitydbd.sh stop    # Stop server
```

### Default Admin User
The server automatically creates a default admin user if none exists:
- Username: `admin`
- Password: `admin`

## Tag Namespaces

- `type:` - Entity type (user, issue, workspace, relationship)
- `id:` - Unique identifiers
- `status:` - Entity state
- `rbac:` - Roles/permissions (FULLY ENFORCED)
  - `rbac:role:admin` - Admin role
  - `rbac:role:user` - Regular user role
  - `rbac:perm:*` - All permissions
  - `rbac:perm:entity:*` - All entity permissions
  - `rbac:perm:entity:view` - View entities
  - etc.
- `conf:` - Configuration
- `feat:` - Feature flags

## What's NOT Implemented

- Rate limiting
- Audit logging
- Aggregation queries (beyond sorting/filtering)

## Recent Changes (v2.13.0)

- **Configuration System Overhaul**: Environment-based configuration with no hardcoded values
- **Configuration Hierarchy**: CLI flags > env vars > instance config > default config
- **Configuration Files**: Default in `share/config/`, instance in `var/`
- **Removed --config Flag**: Eliminated unused configuration flag
- **Project Cleanup**: Reorganized directory structure, moved scripts to proper locations
- **SSL Default Changes**: Disabled by default for development
- **Port Standardization**: Consistent use of 8085/8443 across documentation

## Previous Changes (v2.8.0)

- **Temporal-Only System**: All tags now stored with timestamps (TIMESTAMP|tag)
- **Transparent API**: Timestamps hidden by default, optional include_timestamps parameter
- **Fixed Authentication**: Updated ListByTag to handle temporal tags correctly
- **Fixed RBAC**: Updated GetTagsByNamespace and HasPermission for temporal tags
- **Fixed UUID Storage**: Changed from 32 to 36 bytes to store full UUIDs
- **Auto-Initialization**: Admin user (admin/admin) created automatically on first start
- **Query Functions**: All search functions now handle temporal tags transparently

## Previous Changes (v2.6.0)

- Implemented secure session management with TTL
- Added session refresh capability
- Created automatic session cleanup
- Added support for concurrent sessions
- Token generation using crypto/rand
- Session expiration tracking
- Previous v2.5.0 changes: RBAC, relationships, gorilla/mux routing

## Known Issues

- RBAC permission caching could improve performance
- Rate limiting not yet implemented

## Git Workflow

All development follows the standardized Git workflow described in [docs/development/git-workflow.md](./docs/development/git-workflow.md). This document defines:

- Branch strategy (trunk-based development)
- Commit message standards and format
- Pull request protocol
- Git hygiene rules
- State tracking with Git describe
- Tagging conventions

## Repository Information

- URL: https://git.home.arpa/itdlabs/entitydb.git
- Branch: main
- Latest tag: v2.13.0

## Development Principles

- Never recreate parallel implementations, always integrate and test your fixes directly in the main code
- Always move unused, outdated, or deprecated code to the `/trash` directory instead of deleting it
