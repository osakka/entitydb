# EntityDB Platform

> [!IMPORTANT]
> EntityDB is a high-performance temporal database where every tag is timestamped with nanosecond precision. All data is stored in a custom binary format (EBF) with Write-Ahead Logging for durability and concurrent access support.

## Current State (v2.22.0)

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
- **RBAC Metrics**: Real-time session monitoring, authentication analytics, and security metrics dashboard
- **Modular Widget System**: Customizable dashboards with extensible widget registry and full-screen responsive layout
- **Professional Logging**: Structured logging with contextual error messages, appropriate log levels, and automatic file/function/line information

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
- RBAC & session metrics with real-time monitoring dashboard

### Applications Built on EntityDB
- **Worca (Workforce Orchestrator)**: Complete workforce management platform inspired by Orca intelligence. Features modular widget system, full-screen responsive layout, Kanban boards, team management, project hierarchies, and real-time analytics with oceanic theming. Includes customizable dashboards, comprehensive widget registry, dark/light themes, and mobile-responsive design. Located at `/share/htdocs/worca/`

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
│   │   ├── rbac_metrics_handler.go  # RBAC & session metrics
│   │   └── rbac_middleware.go       # RBAC enforcement middleware
│   ├── models/                      # Entity models
│   └── storage/binary/              # Binary format implementation
│       ├── entity_repository.go
│       └── relationship_repository.go  # Binary relationships
├── bin/
│   ├── entitydb                     # Server binary
│   └── entitydbd.sh                 # Daemon script
├── share/htdocs/worca/              # Workforce Orchestrator Application  
│   ├── index.html                   # Main dashboard with full-screen layout
│   ├── worca.js                     # Core application logic with widget system
│   ├── worca-api.js                 # EntityDB API wrapper
│   ├── worca-widgets.js             # Modular widget system registry
│   ├── worca-logo-dark.svg          # Orca-inspired logo (dark theme)
│   ├── worca-logo-light.svg         # Orca-inspired logo (light theme)
│   ├── metrics.html                 # System metrics dashboard
│   └── demo.md                      # Usage examples and demos
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
GET    /api/v1/rbac/metrics          # RBAC & session metrics (requires admin)

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

## Recent Changes (v2.22.0)

- **Comprehensive Metrics System**: Phase 1 implementation of advanced observability
  - Query performance metrics with complexity scoring and slow query detection
  - Storage operation metrics tracking read/write latencies, WAL operations, and compression
  - Error tracking system with categorization, pattern detection, and recovery metrics
  - Request/response metrics middleware for HTTP performance monitoring
  - Configurable metrics collection interval via ENTITYDB_METRICS_INTERVAL
  - Enhanced Performance tab in UI with new metric cards and charts
  - All metrics stored using temporal tags with configurable retention policies
- **Code Quality Improvements**: Build fixes and deduplication
  - Fixed compilation error in entity creation (unused startTime variable)
  - Added missing storage metrics tracking for Create operation
  - Removed duplicate tool files maintaining single source of truth
  - Clean build with no warnings or errors

## Recent Changes (v2.21.0)

- **Tab Structure Validation System**: Comprehensive protection against UI rendering issues
  - Runtime validation automatically checks tab structure on page load
  - Build-time validation integrated into Makefile prevents broken builds
  - Git pre-commit hook blocks commits with invalid tab structures
  - Converted all 10 dashboard tabs from x-show to x-if templates for proper flex layout compatibility
- **Request/Response Metrics**: New HTTP request tracking middleware
  - Tracks duration, size, status codes, and errors with temporal storage
  - Provides historical analysis capabilities for performance monitoring
- **Enhanced Monitoring UI**: Improved chart visualizations
  - Added legends, tooltips, and proper units to all monitoring charts
  - Better data formatting and user experience
- **WAL Checkpoint Metrics**: Added comprehensive checkpoint tracking
  - Monitors checkpoint operations, success rates, and storage efficiency
  - Provides insights into storage health and performance

## Recent Changes (v2.20.0)

- **Advanced Memory Optimization**: Comprehensive memory management improvements
  - String interning for tag storage reducing memory by up to 70% for duplicate tags
  - Sharded lock system for high-concurrency scenarios  
  - Safe buffer pool implementation with size-based pools (small, medium, large)
  - Compression support for entity content with 1KB threshold
  - Memory pool integration throughout storage layer
- **Authentication System Fix**: Resolved credential storage and retrieval issues
  - Fixed compression handling for credential entities
  - Corrected reader implementation to properly handle both compressed and uncompressed content
  - Ensured bcrypt hashes are stored and retrieved without corruption
  - Fixed binary format reader to correctly parse both original and compressed sizes
- **Storage Layer Optimizations**: 
  - Enhanced writer with compression support using gzip for content > 1KB
  - Improved reader with proper decompression handling
  - Added trace logging for compression operations
  - Integrated buffer pools for reduced GC pressure
- **Development Tools Cleanup**: Moved 30+ debug/fix tools to trash
  - Removed temporary authentication debugging tools
  - Cleaned up credential fix utilities
  - Removed duplicate reader implementations
  - Maintained single source of truth principle

## Previous Changes (v2.19.0)

- **Critical WAL Management Fix**: Prevented unbounded WAL growth that caused disk space exhaustion
  - Implemented automatic WAL checkpointing: every 1000 operations, 5 minutes, or 100MB size
  - Added checkpoint triggers to Create(), Update(), and AddTag() operations
  - Fixed temporal timeline indexing in AddTag() method for metrics collection
  - Added WAL monitoring metrics: wal_size, wal_size_mb, wal_warning, wal_critical
- **Temporal Metrics System**: Complete real-time metrics implementation
  - 1-second collection interval with change-only detection
  - Temporal storage of metric values using AddTag() with proper timeline indexing
  - Retention tags for automatic data lifecycle (retention:count:100, retention:period:3600)
  - Fixed "entity timeline not found" errors by maintaining indexes during AddTag operations
- **Background Metrics Collector**: Enhanced system metrics collection
  - Memory, GC, database, entity, and WAL metrics
  - Change detection to prevent redundant writes
  - Thread-safe implementation with proper mutex protection
- **Code Audit and Cleanup**: Major codebase consolidation
  - Moved 28+ debug/fix tools to trash directory
  - Removed redundant handler implementations
  - Cleaned up temporal fix scripts
  - Maintained single source of truth principle

## Previous Changes (v2.18.0)

- **Logging Standards Implementation**: Professional logging system with consistent formatting
  - Removed redundant manual prefixes since logger provides file/function/line automatically  
  - Enhanced API error messages with contextual information (entity IDs, query parameters, operation details)
  - Fixed inappropriate log levels (error conditions moved from DEBUG to WARN/ERROR, detailed operations moved from INFO to TRACE)
  - Reduced excessive INFO logging in storage layer (reader.go and writer.go operations now at TRACE level)
  - Created comprehensive logging audit and standards documentation
- **Public RBAC Metrics Endpoint**: New unauthenticated endpoint for basic metrics
  - `/api/v1/rbac/metrics/public` provides basic authentication and session counts without requiring admin access
- **RBAC Tag Manager**: Enhanced RBAC management component for user tag operations
- **Code Cleanup**: Moved 19 fix files and 15 debug tools to trash, consolidated redundant implementations

## Previous Changes (v2.16.0)

- **UUID Storage Fix**: Fixed critical authentication bug by increasing EntityID from 36 to 64 bytes
  - Resolved login failures due to truncated UUIDs in binary format
  - All entity operations now correctly handle full UUID strings
  - Fixed user authentication and session management

## Previous Changes (v2.13.0)

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
- Latest tag: v2.22.0

## Development Principles

- Never recreate parallel implementations, always integrate and test your fixes directly in the main code
- Always move unused, outdated, or deprecated code to the `/trash` directory instead of deleting it
