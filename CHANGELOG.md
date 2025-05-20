# EntityDB Changelog

All notable changes to the EntityDB Platform will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v2.13.0] - 2025-05-19

### Changed
- **Configuration System Overhaul**: Implemented comprehensive environment-based configuration
  - All hardcoded values moved to environment variables
  - New configuration hierarchy: CLI flags > env vars > instance config > default config
  - Default configuration in `share/config/entitydb_server.env`
  - Instance-specific overrides in `var/entitydb.env`
  - All configuration variables prefixed with `ENTITYDB_`
- Project structure cleanup:
  - Moved temporary scripts to `tmp/` directory
  - Reorganized configuration files to `share/config/`
  - Updated all documentation to reflect new structure
- Removed unused `--config` flag that was never implemented
- Updated startup script to source configuration files
- Changed default SSL setting to false for development
- Updated default ports to 8085 (HTTP) and 8443 (HTTPS)
- **Fixed content encoding issues in entity API**:
  - Resolved critical double-encoding problem for entity content
  - Added proper content type handling for text content
  - Fixed serialization to store content directly without wrapping in JSON
  - Ensured content is properly decoded with a single base64 decode

### Added
- **Full SSL/TLS Support**:
  - Configurable via environment variables (`ENTITYDB_USE_SSL`, `ENTITYDB_SSL_PORT`, etc.)
  - SSL certificate verification and validation
  - Runtime certificate configuration
  - Single port SSL configuration (all on same port)
  - Certificate information display
- `share/config/entitydb_server.env` - Default configuration file with all available settings
- `docs/CONFIG_SYSTEM.md` - Comprehensive configuration documentation
- Environment variable support for all configuration options
- Automatic configuration file sourcing in startup script
- Configuration precedence hierarchy documentation
- Better debugging and error reporting for entity operations
- **In-memory entity cache** for faster and more reliable entity retrieval
- **Strong durability guarantees** for entity storage with multiple sync points
- **MIME Type Detection**:
  - Auto-detection of content types (string vs JSON)
  - Content type tagging with `content:type:*` tags
  - Proper base64 encoding with content preservation

### Fixed
- Entity persistence issues - created entities can now be immediately retrieved
- Content deserialization to avoid double encoding
- Entity retrieval reliability through improved caching
- Entity file format to properly handle binary content
- Prevented duplicate content type tags
- Improved synchronization with disk for database operations
- **Critical Content Encoding Issues**:
  - Fixed JSON content double-encoding problem
  - Corrected content storage format issues
  - Resolved inconsistencies in base64 encoding/decoding
  - Fixed data corruption risks with binary content
  - Ensured backward compatibility with existing entities
- SSL certificate verification only run when SSL is enabled
- Proper port handling for both HTTP and HTTPS modes

### Removed
- Unused `--config` command line flag
- Hardcoded configuration values in source code
- Unnecessary base64 encoding of entity content
- Legacy JSON wrapping of binary content

## [v2.11.1] - 2025-05-19

### Changed
- Modified `entitydbd.sh` to enable SSL by default
- Server runs in SSL-only mode (no HTTP listener)
- Uses specified certificates: `/etc/ssl/certs/server.pem` and `/etc/ssl/private/server.key`
- All URLs in daemon script updated to use HTTPS
- Added SSL certificate validation on startup
- Added SSL certificate information display

### Security
- Removed HTTP listener for enhanced security
- All connections now encrypted by default
- HTTPS-only mode prevents accidental unencrypted connections

## [v2.11.0] - 2025-05-19

### Added
- **SSL/TLS Support**: Full HTTPS support with configurable certificates
  - Automatic HTTP to HTTPS redirect
  - Configurable SSL port (default: 8443)
  - Support for self-signed and CA certificates
  - SSL setup utility script
  - Comprehensive SSL testing tools
- SSL configuration flags:
  - `--use-ssl`: Enable SSL/TLS
  - `--ssl-cert`: Path to certificate file
  - `--ssl-key`: Path to private key file
  - `--ssl-port`: HTTPS port number
- Documentation for SSL configuration and best practices

### Security
- Encrypted client-server communication
- Secure defaults for TLS configuration
- Support for modern TLS versions and ciphers

## [v2.10.0] - 2025-05-19

### Added
- **Temporal Repository**: Enhanced high-performance mode with temporal capabilities
  - B-tree timeline index for temporal queries
  - Time-bucketed indexes for efficient range queries
  - Per-entity temporal timelines
  - Temporal query caching with LRU eviction
  - Multiple timestamp format support (ISO, numeric, legacy)
- **Performance Optimization**: Up to 100x improvement for temporal operations
- Comprehensive test suites for temporal features
- Documentation for temporal implementation

### Changed
- Default repository is now TemporalRepository
- Updated `GetTagsWithoutTimestamp()` to handle all timestamp formats
- Repository factory pattern extended for temporal repository
- Renamed all "turbo" terminology to "high-performance" for professional clarity

### Fixed
- Timestamp stripping for mixed format databases
- Admin user creation issue after database deletion
- Entity retrieval with various timestamp formats

## [v2.9.0] - 2025-05-19

### Added
- **High-Performance Mode by Default**: 25x performance improvement with advanced optimizations
  - Memory-mapped files for zero-copy reads
  - Skip-list indexes for O(log n) lookups
  - Bloom filters for fast existence checks
  - Parallel query processing with worker pools
  - Advanced multi-level caching
- Performance benchmarking tools (`share/tests/high_performance_benchmark.py`)
- Configurable high-performance mode (disable with `ENTITYDB_DISABLE_HIGH_PERFORMANCE=true`)
- Repository factory pattern for choosing implementation
- Performance statistics tracking

### Performance Improvements
- Average query latency reduced from 189ms to 7.47ms (25x faster)
- Query throughput increased to 50-80 QPS per thread
- Optimized temporal queries: 54ms (down from 690ms)
- Optimized namespace queries: 78ms (down from 368ms)
- Fixed index corruption and entity data corruption

### Fixed
- Index corruption with invalid seek offsets (5 entries repaired)
- Entity data corruption (5 corrupted entries removed)
- Memory leak in reader pool
- Concurrent access issues in index operations
- Build errors with multiple main functions in tools directory

### Changed
- Turbo mode is now the default behavior
- Standard mode requires explicit `ENTITYDB_DISABLE_TURBO=true`
- Improved error handling in deserialization

## [v2.8.0] - 2025-05-18

### Added
- Temporal-only system - all tags now stored as TIMESTAMP|tag
- Transparent API - timestamps hidden by default, optional include_timestamps
- Auto-initialization - admin/admin user created automatically on first start
- Fixed UUID storage - changed from 32 to 36 bytes to store full UUIDs

### Changed
- ListByTag now searches ignoring timestamps for non-temporal queries
- GetTagsByNamespace, ParseTag, IsNamespace updated for temporal tags
- HasPermission function updated to handle temporal permission tags
- Entity index format updated to support full 36-byte UUIDs
- Authentication and RBAC now work correctly with temporal tags

### Fixed
- Authentication now works with temporal tag system
- RBAC permissions properly checked with temporal tags
- Entity IDs no longer truncated in binary storage
- Query functions transparently handle temporal tags

## [v2.7.1] - 2025-05-18

### Added
- Temporal-only system with automatic timestamps on all tags
- Transparent nanosecond precision timestamps using | delimiter
- Admin user initialization integrated into entitydbd.sh startup script
- Debug logging for static file serving

### Changed
- Entity model now uses temporal-only tags without backward compatibility
- Storage layer enforces temporal format on all tags
- Static file serving uses absolute paths for security checks
- Test fixtures updated for temporal tag format
- Swagger documentation updated to explain temporal behavior

### Fixed
- Static file serving now correctly serves dashboard at root path
- Path security check for static files properly resolves relative paths
- Database admin initialization runs automatically on startup

### Removed
- Backward compatibility for old timestamp delimiters (.)
- Legacy tag format support

## [v2.7.0] - 2024-12-XX

### Added
- Full Swagger/OpenAPI documentation for all API endpoints
- Advanced query functionality with sorting, filtering, and pagination
- Query builder pattern for flexible entity searches
- Support for tag filtering, wildcard patterns, and content type filtering
- Sorting by multiple fields (timestamp, ID, tag count, content count)
- Concurrent query operations support
- Comprehensive API documentation available at `/swagger/`
- QueryEntityResponse model for pagination metadata

### Changed
- RBAC middleware updated to work with session authentication
- API authentication now properly integrates with session management
- Entity handler extended with QueryEntities method
- Repository interface extended with Query() method

### Fixed
- Authentication token parsing in RBAC middleware
- CORS headers for Swagger UI access
- Compilation issues with duplicate type definitions
- Token format expectations (now correctly uses "Bearer " prefix)

## [v2.6.0] - 2024-05-XX

### Added
- Secure session management with TTL
- Session refresh capability
- Automatic session cleanup
- Support for concurrent sessions
- Token generation using crypto/rand
- Session expiration tracking
- Session-based authentication middleware

## [v2.5.0] - 2024-05-XX

### Added
- Hierarchical tag namespace system with 10 core namespaces
- Tag-based permission checking with wildcard support
- Permission middleware that actually enforces permissions
- Comprehensive documentation for tag namespaces
- API reference documentation
- Architecture documentation
- Alpine.js web dashboard with auto-refresh
- Entity inline editing in web UI
- Test scripts for RBAC permission validation

### Changed
- Updated all API routes to use hierarchical rbac:perm:* format
- Migrated from flat permission strings to hierarchical tags
- Updated user entities to use new tag namespace format
- Improved CLAUDE.md documentation to reflect current state
- Simplified README.md for better clarity
- Reorganized documentation structure

### Fixed
- Entity update functionality in web UI
- Permission middleware now checks actual permissions, not just authentication
- Tag namespace conflicts between different implementations
- Compilation issues with duplicate type definitions

## [0.3.0] - 2024-05-16

### Added
- Alpine.js web interface with reactive updates
- Entity browser with search and filtering
- Auto-refresh feature (60-second interval)
- Dark/light theme support
- Inline entity editing

### Changed
- Migrated from jQuery to Alpine.js
- Simplified frontend architecture
- Improved UI responsiveness

### Fixed
- Web UI update issues
- CORS headers for API access

## [0.2.0] - 2024-05-15

### Added
- Pure entity-based architecture
- Entity relationship system
- JWT authentication
- Tag-based categorization
- CLI tools (entitydb-cli)
- SQLite persistence

### Changed
- Migrated from task-based to entity-based model
- Consolidated all operations under entity API
- Deprecated specialized endpoints

### Removed
- Legacy task management system
- Direct database access patterns
- Specialized data models

## [0.1.0] - 2024-05-01

### Added
- Initial EntityDB platform
- Agent management
- Issue tracking
- Workspace organization
- Basic authentication
- REST API
- Command-line client

[Unreleased]: https://git.home.arpa/osakka/entitydb/compare/v2.7.0...HEAD
[v2.7.0]: https://git.home.arpa/osakka/entitydb/compare/v2.6.0...v2.7.0
[v2.6.0]: https://git.home.arpa/osakka/entitydb/compare/v2.5.0...v2.6.0
[v2.5.0]: https://git.home.arpa/osakka/entitydb/compare/v0.3.0...v2.5.0
[0.3.0]: https://git.home.arpa/osakka/entitydb/compare/v0.2.0...v0.3.0
[0.2.0]: https://git.home.arpa/osakka/entitydb/compare/v0.1.0...v0.2.0
[0.1.0]: https://git.home.arpa/osakka/entitydb/releases/tag/v0.1.0