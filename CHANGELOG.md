# EntityDB Changelog

All notable changes to the EntityDB Platform will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.29.0] - 2025-06-11

### Added
- **Complete UI/UX Overhaul**: Professional 5-phase implementation
  - **Phase 1 - Foundation Components**:
    - Centralized API client with unified error handling
    - Structured logging system with session tracking
    - Toast notification system with queue management
    - Base component framework for reusability
  - **Phase 2 - Design System**:
    - Comprehensive CSS variables and theming
    - Dark mode support with persistent preferences
    - Responsive grid system and utilities
    - Enhanced entity browser with card layout
    - Widget system with drag-and-drop support
  - **Phase 3 - State Management**:
    - Vuex-inspired centralized state management
    - Component lazy loading with Intersection Observer
    - Virtual scrolling for large datasets
    - Reactive updates and subscriptions
  - **Phase 4 - Advanced Features**:
    - Multi-tier cache management (memory, localStorage, IndexedDB)
    - Enhanced chart components with real-time updates
    - Performance optimizations and monitoring
  - **Phase 5 - Testing & Documentation**:
    - Component testing framework
    - Interactive UI documentation
    - Comprehensive implementation guide

### Changed
- **Major Terminology Update**: Renamed "dataspace" to "dataset" throughout entire codebase
  - Updated all API endpoints from `/dataspace` to `/dataset`
  - Renamed environment variables from `ENTITYDB_DATASPACE_*` to `ENTITYDB_DATASET_*`
  - Changed all Go types and functions to use Dataset naming
  - Updated UI components and JavaScript files to reflect new terminology
  - Modified documentation to use consistent "dataset" terminology
  - Maintains backward compatibility through compatibility layer

### Fixed
- **Vue.js Integration**: Resolved compiler errors and component registration issues
  - Fixed Vue 3 template compilation error #30
  - Corrected nested component dependencies
  - Simplified tab navigation structure
  - Fixed entity browser API client integration
- **Code Quality**: Comprehensive repository cleanup and maintenance
  - Removed duplicate index.html files (index_new.html, index_simple.html)
  - Cleaned up old debug and fix binaries from bin directory
  - Moved obsolete files to trash following single source of truth principle
  - Ensured clean build with zero warnings

### Documentation
- Updated all references from "dataspace" to "dataset" in documentation
- Maintained consistency across READMEs, API docs, and inline comments
- Added comprehensive UI/UX implementation documentation
- Created documentation audit and migration plans

## [2.28.0] - 2025-06-07

### Added
- **Enhanced Metrics System**: Comprehensive metrics collection and management
  - Configurable retention policies for raw and aggregated metrics data
  - Metric types system (Counter, Gauge, Histogram) leveraging temporal storage
  - Retention manager with automatic data lifecycle management (raw, 1min, 1hour, daily)
  - Histogram bucket configuration for latency and distribution tracking
  - Conditional metrics collection with separate flags for request/storage tracking
  - Standalone metrics dashboard with auto-refresh and time range selection
  - Enhanced Chart.js integration with multiple chart types and real-time updates
- **Connection Stability Improvements**: Fixed browser connection hangs
  - TE header middleware to handle Transfer-Encoding header conflicts
  - Connection close middleware for proper connection termination
  - Comprehensive request tracing for debugging connection issues
  - Disabled HTTP/2 in TLS configuration to fix ERR_HTTP2_PROTOCOL_ERROR
- **Logging System Enhancements**: Professional logging with trace subsystems
  - Log bridge to redirect standard library logs through structured logger
  - Trace subsystem support for targeted debugging
  - Lock operation tracing for deadlock detection
  - HTTP request tracing with goroutine IDs
- **Professional Documentation Library**: Complete transformation of EntityDB documentation
  - Evolved existing documentation structure using professional taxonomy
  - Consolidated ~150 scattered files into well-organized, accurate documentation  
  - Created comprehensive master index with clear navigation paths
  - Established cross-reference system for easy navigation between related topics
  - Added documentation maintenance guidelines with quarterly review process
  - Enhanced architecture documentation (temporal, RBAC, performance)
- **Entity Model Enhancements**: Temporal tag utility methods
  - `HasTag()`: Check for tag existence without timestamp concerns
  - `GetTagValue()`: Retrieve most recent value for a given tag key
  - Both methods properly handle RFC3339 and epoch nanosecond formats

### Changed  
- **Documentation Accuracy**: Fixed critical documentation issues
  - Corrected architecture docs claiming SQLite when using binary format
  - Fixed RBAC permission format documentation
  - Updated API reference to include all v2.28.0 endpoints
  - Aligned all technical documentation with actual implementation
- **Build System**: Enhanced swagger documentation generation
  - Integrated swagger generation into standard build process
  - Added tab structure validation for UI stability
  - Maintained clean build with zero warnings

### Fixed
- **Single Source of Truth**: Repository maintenance
  - Removed all obsolete test scripts and debug utilities
  - Integrated all patches and fixes into main codebase
  - Eliminated parallel implementations and redundant code
  - Enforced clean workspace guidelines
- **Version Consistency**: Updated all version references to v2.28.0
  - Configuration files, documentation, and code all use consistent version
  - Regenerated Swagger documentation with correct version

## [2.27.0] - 2025-06-07

### Added
- **Engineering Excellence Infrastructure**: Complete CI/CD and development tooling
  - GitHub Actions workflows for automated testing, security scanning, and releases
  - Production-ready Dockerfile with multi-stage builds and security hardening
  - One-command developer setup script with all dependencies
  - Enhanced Makefile with CI/CD targets (test-ci, security-scan, lint, docker)
  - Hot reload development environment with Air
  - Pre-commit hooks for code quality enforcement
- **Documentation Taxonomy**: Professional documentation organization system
  - Comprehensive documentation audit and reorganization plan
  - Technical accuracy verification against actual codebase
  - Industry-standard naming schema and cross-referencing system
  - API documentation accuracy report with discrepancy analysis

### Changed
- **Configuration Management Enhancement**: Fine-tuned 3-tier configuration system
  - Improved database configuration caching with proper expiry handling
  - Enhanced runtime log level adjustment via multiple interfaces
  - Consolidated all hardcoded values into configurable parameters

### Fixed
- **Test Framework Compatibility**: Updated test files to match current Entity model
  - Fixed temporal tag parsing in test files to use nanosecond epoch format
  - Added missing Entity model methods (HasTag, GetTagValue) for backward compatibility
  - Corrected timestamp parsing from RFC3339Nano to nanosecond epoch in repository code

## [2.26.0] - 2025-06-07

### Added
- **HTTP Connection Hang Fixes**: Comprehensive authentication and connection stability improvements
  - `ConnectionCloseMiddleware`: Forces connection closure to prevent hanging connections 
  - `TEHeaderMiddleware`: Strips problematic TE header that causes server hangs
  - `TraceMiddleware`: Request tracing for debugging connection and authentication issues
  - `LogBridge`: Redirects standard library HTTP error logs to EntityDB logger with proper categorization
- **Advanced Locking System**: Enhanced concurrency control for high-performance scenarios
  - `ShardedLockManager`: Distributed locking across multiple shards to reduce contention
  - `TracedLocks`: Lock tracing and deadlock detection for debugging concurrent access issues
  - `LockTracer`: Comprehensive lock operation tracking and timing analysis
- **Implementation Documentation**: Detailed records of authentication and performance fixes
  - Authentication hang fix analysis and implementation plan
  - ListByTag deadlock detection and prevention strategy
  - TE header hang root cause analysis and solution
  - Performance optimization implementation guide

### Fixed
- **Authentication System Stability**: Resolved login hangs and timeouts
  - Fixed browser-specific hanging when TE: trailers header is present
  - Prevented connection pooling issues that caused authentication delays
  - Enhanced error logging for TLS and authentication failures
- **Concurrent Access Issues**: Eliminated deadlocks in high-concurrency scenarios
  - Fixed ListByTag method deadlock when acquiring multiple entity locks
  - Improved lock ordering to prevent circular wait conditions
  - Enhanced lock timeout handling and error recovery

### Changed
- **Development Workflow**: Major cleanup of debug and temporary tools
  - Moved 40+ debug tools, test scripts, and analysis utilities to trash
  - Retained only essential repair tools for maintenance operations
  - Consolidated all fixes into main codebase following single source of truth principle
  - Clean repository structure with clear separation of production vs debug code

## [2.25.0] - 2025-06-05

### Fixed
- **Complete Metrics System Overhaul**: Fixed all performance metrics showing 0 values
  - Fixed WAL persistence to save current in-memory entity state instead of initial WAL entry state
  - Re-enabled auth event tracking that was disabled for performance
  - Added error tracking with `TrackHTTPError` throughout entity and auth handlers
  - Fixed query metrics tracking in `ListEntities` function
  - Fixed temporal tag parsing in public RBAC metrics handler
  - Changed metrics aggregation window from 30 minutes to 24 hours for better coverage
  - All metrics now show real values: HTTP duration, query time, storage operations, error counts

### Added
- **Metrics History API**: Registered previously unconnected endpoints
  - `/api/v1/metrics/history` - Get historical values for specific metrics
  - `/api/v1/metrics/available` - List all available metrics
  - Enables UI charts to display historical data properly

### Changed
- **Code Quality**: Comprehensive audit and cleanup
  - Removed all temporary debug tools and test scripts
  - Consolidated duplicate implementations into single source of truth
  - Moved obsolete files to trash directory
  - Updated placeholder comments to indicate unimplemented features
  - Clean build with zero warnings

## [2.24.0] - 2025-06-03

### Fixed
- **Critical WAL Persistence Bug**: Fixed issue where temporal metric values were lost during checkpoints
  - Added `persistWALEntries()` method to write WAL entries to binary files before truncation
  - Ensures all `AddTag()` operations for temporal metrics are durably persisted
  - Prevents data loss when WAL is truncated during checkpoint operations
- **Metrics Aggregation**: Fixed aggregator to properly collect and sum labeled metrics
  - Re-fetches entities with full temporal tags for accurate timestamp parsing
  - Correctly identifies recent values within the 30-minute aggregation window
  - UI graphs now display actual metric data instead of zeros

### Added
- **Metrics Aggregator Service**: New background service for UI metric aggregation
  - Aggregates labeled metrics (with dimensions) into simple metrics for UI consumption
  - Runs every 30 seconds by default (configurable via `ENTITYDB_METRICS_AGGREGATION_INTERVAL`)
  - Supports sum, avg, max, min, and last aggregation methods

### Changed
- **Code Organization**: Major cleanup for maintainability
  - Consolidated duplicate cleanup tools (kept most comprehensive versions)
  - Removed duplicate admin creation tools
  - Moved redundant implementations to trash directory
  - Cleaned up compiled binaries from source directories

## [2.23.0] - 2025-06-02

### Changed
- **Application-Agnostic Platform**: Removed all application-specific code from core server
  - Replaced worca-specific `/api/v1/worca/metrics` endpoint with generic `/api/v1/application/metrics`
  - Applications can now filter metrics by `namespace` or `app` query parameter
  - Moved example applications (worca, methub) out of core distribution to trash directory
  - EntityDB is now a pure database platform without embedded applications
  - Updated all documentation to reflect the application-agnostic design

### Added
- **Generic Application Metrics API**: New endpoint for application-specific metrics
  - `/api/v1/application/metrics` accepts namespace/app parameter for filtering
  - Returns metrics in a format suitable for any application
  - Maintains RBAC enforcement (requires `metrics:read` permission)

### Removed
- Worca application files from `/share/htdocs/worca/`
- Methub application files from `/share/htdocs/methub/`
- Application-specific handlers (`worca_metrics_handler.go`)

## [2.22.0] - 2025-06-02

### Added
- **Comprehensive Metrics System**: Phase 1 implementation of advanced observability
  - Query performance metrics with complexity scoring and slow query detection
  - Storage operation metrics tracking read/write latencies, WAL operations, and compression
  - Error tracking system with categorization, pattern detection, and recovery metrics
  - Request/response metrics middleware for HTTP performance monitoring
  - Configurable metrics collection interval via `ENTITYDB_METRICS_INTERVAL` environment variable
  - Enhanced Performance tab in UI with new metric cards and charts
  - All metrics stored using temporal tags with configurable retention policies

### Fixed
- **Compilation Error**: Fixed unused `startTime` variable in entity creation
  - Added missing storage metrics tracking for Create operation
  - Ensures consistent metrics collection across all storage operations
- **Code Duplication**: Removed duplicate tool files
  - Consolidated `clean_corrupted_entries.go`, `scan_entity_data.go`, and `test_chunking.go`
  - Moved redundant admin tools to trash directory
  - Maintained single source of truth principle

### Changed
- **Metrics Collection**: Made background metrics collection interval configurable
  - Default 30 seconds, supports any Go duration format
  - Reduces overhead in production environments
- **Documentation**: Updated metrics documentation
  - Created METRICS_AUDIT_FINDINGS.md with comprehensive gap analysis
  - Created METRICS_ACTION_PLAN.md with phased implementation roadmap
  - Created METRICS_IMPLEMENTATION_SUMMARY.md documenting Phase 1 completion

### Documentation
- Comprehensive metrics implementation documentation
- Updated action plan showing Phase 1 complete
- Detailed configuration and usage examples

## [2.21.0] - 2025-06-01

### Added
- **Tab Structure Validation System**: Comprehensive validation to prevent UI tab rendering issues
  - Runtime validation with `/js/tab-validator.js` that checks tab structure on page load
  - Build-time validation script `/scripts/validate_tab_structure.sh`
  - Git pre-commit hook to prevent committing broken tab structures
  - Detailed documentation in `/docs/development/TAB_STRUCTURE_GUIDELINES.md`
- **Request/Response Metrics**: New middleware for HTTP request tracking
  - Tracks request duration, size, status codes, and errors
  - Stores metrics using temporal tags for historical analysis
  - Integrated into main server initialization
- **Enhanced UI Charts**: Improved monitoring dashboards
  - Added legends to all charts with proper positioning
  - Implemented tooltips with formatted values and units
  - Added proper axis labels and scaling

### Fixed
- **Critical Tab Rendering Issue**: Fixed dashboard tabs not displaying
  - Root cause: Using `x-show` with flex layouts caused tabs after Storage to be invisible
  - Solution: Converted all 10 tabs from `x-show` to `x-if` with template tags
  - This ensures only active tab is in DOM, preventing flex layout conflicts
- **WAL Checkpoint Metrics**: Added proper metrics collection for checkpoint operations
  - Tracks checkpoint success/failure, duration, and size reduction
  - Provides visibility into storage health and performance

### Changed
- **Tab Implementation Pattern**: Migrated from x-show to x-if templates
  - All tabs now use `<template x-if="activeTab === 'name'">` pattern
  - Improves performance by removing inactive tabs from DOM
  - Prevents layout calculation issues with hidden flex children
- **Build Process**: Added tab validation to Makefile
  - Server build now validates tab structure before compilation
  - Ensures UI consistency is maintained across builds

### Documentation
- Created comprehensive tab structure guidelines
- Updated build documentation with validation steps
- Added troubleshooting guide for tab-related issues

## [2.20.0] - 2025-05-30

### Added
- **Advanced Memory Optimization**: Comprehensive memory management improvements
  - String interning for tag storage reducing memory by up to 70% for duplicate tags
  - Sharded lock system for high-concurrency scenarios
  - Safe buffer pool implementation with size-based pools (small, medium, large)
  - Compression support for entity content with 1KB threshold
  - Memory pool integration throughout storage layer

### Fixed
- **Authentication System**: Resolved credential storage and retrieval issues
  - Fixed compression handling for credential entities
  - Corrected reader implementation to properly handle both compressed and uncompressed content
  - Ensured bcrypt hashes are stored and retrieved without corruption
  - Fixed binary format reader to correctly parse both original and compressed sizes

### Changed
- **Storage Layer Optimizations**: 
  - Enhanced writer with compression support using gzip for content > 1KB
  - Improved reader with proper decompression handling
  - Added trace logging for compression operations
  - Integrated buffer pools for reduced GC pressure

### Removed
- **Development Tools Cleanup**: Moved 30+ debug/fix tools to trash
  - Removed temporary authentication debugging tools
  - Cleaned up credential fix utilities
  - Removed duplicate reader implementations
  - Maintained single source of truth principle

## [2.19.0] - 2025-05-30

### Fixed
- **Critical WAL Management Issue**: Prevented unbounded WAL growth that caused 8GB disk space exhaustion
  - Implemented automatic WAL checkpointing: every 1000 operations, 5 minutes, or 100MB size
  - Added checkpoint triggers to Create(), Update(), and AddTag() operations
  - Fixed root cause: WAL was only truncated at startup, never during runtime
  - Added WAL monitoring metrics: wal_size, wal_size_mb, wal_warning (>50MB), wal_critical (>100MB)
- **Temporal Timeline Indexing**: Fixed "entity timeline not found" errors for metrics
  - Added AddTag() method to TemporalRepository that maintains timeline indexes
  - Fixed metrics history API that was failing due to missing timeline entries
  - Ensured all temporal tag additions update the entity timeline index

### Added
- **Real-Time Temporal Metrics System**: Complete metrics collection and visualization
  - Background collector runs every 1 second with change-only detection
  - Temporal storage using AddTag() for time-series data
  - Retention management tags: retention:count:100, retention:period:3600
  - Fixed time periods for charts: 1h, 24h, 7d, 30d with appropriate grid sizing
  - Zero-fill for missing data points, no fallback to mock data
- **Enhanced Metrics Collection**: Comprehensive system monitoring
  - Memory metrics: alloc, total_alloc, sys, heap_alloc, heap_inuse
  - GC metrics: runs, pause duration
  - Database metrics: size, WAL size, index size
  - Entity metrics: counts by type, creation statistics
  - All metrics stored as temporal tags for historical analysis
- **Code Consolidation**: Major cleanup maintaining single source of truth
  - Moved 28+ debug/fix tools from src/tools to trash/tools_debug
  - Removed redundant API handlers to trash/api_redundant
  - Cleaned up temporal fix scripts to trash/temporal_fixes
  - Maintained production code integrity while removing development artifacts

### Changed
- **Metrics Background Collector**: Enhanced with thread-safety and efficiency
  - Added lastValues map for change detection
  - Thread-safe implementation with sync.RWMutex
  - Only writes metrics when values actually change
  - Proper mutex protection for concurrent access

## [2.18.0] - 2025-05-29

### Added
- **Logging Standards Implementation**: Professional logging system with consistent formatting
  - Removed redundant manual prefixes (`[Transaction]`, `[WAL]`, `[Writer]`, `[Reader]`) since logger provides file/function/line automatically
  - Enhanced API error messages with contextual information (entity IDs, query parameters, operation details)
  - Fixed inappropriate log levels (error conditions moved from DEBUG to WARN/ERROR, detailed operations moved from INFO to TRACE)
  - Reduced excessive INFO logging in storage layer (reader.go and writer.go operations now at TRACE level)
  - Created comprehensive logging audit and standards documentation  
  - Established pattern for replacing direct print statements with structured logger calls
- **Public RBAC Metrics Endpoint**: New unauthenticated endpoint for basic metrics
  - `/api/v1/rbac/metrics/public` provides basic authentication and session counts without requiring admin access
  - Complements existing authenticated `/api/v1/rbac/metrics` endpoint
- **RBAC Tag Manager**: Enhanced RBAC management component for user tag operations
- **Repository Cleanup Tools**: Maintenance utilities for duplicate user cleanup and system health
- **Data Integrity System**: Comprehensive operation tracking and logging infrastructure
  - Operation ID generation for all data operations (READ, WRITE, DELETE, INDEX, WAL)
  - Enhanced logging in Writer with SHA256 checksums and write verification
  - Enhanced logging in Reader with better bounds checking and EOF handling
  - WAL operation tracking with detailed replay logging
  - Created `/opt/entitydb/src/models/operation_tracking.go` for centralized tracking
- **RBAC Metrics Dashboard**: Comprehensive real-time monitoring system for authentication, sessions, and access control
  - `/api/v1/rbac/metrics` endpoint with detailed session analytics
  - Authentication success/failure timeline with visual charts
  - Active session monitoring with user details and duration tracking
  - Role distribution analysis and permission usage statistics
  - Security scoring and health indicators
  - Professional charts using Chart.js with dark/light theme support
  - Zero mock data - 100% real session and authentication metrics
- **Enhanced Admin Interface**: New RBAC Metrics tab in EntityDB admin dashboard
  - Real-time session table with username, role, duration, and status
  - Authentication activity log with timestamps and details
  - Interactive charts for authentication timeline and session activity
  - Role distribution doughnut chart with live data
  - Summary cards showing key security metrics and statistics

### Changed
- **Time Format Standardization**: All timestamps now use int64 nanoseconds since Unix epoch
  - Created `/opt/entitydb/src/models/time_utils.go` with standardized time functions
  - Removed duplicate temporal tag implementations (maintaining single source of truth)
  - Deprecated `temporal_format.go` and consolidated to `time_utils.go`
- **Fixed Index Corruption**: Binary writer now writes index entries in sorted order
  - Fixed map iteration causing random index order
  - Added verification that written entries match header count
  - Auto-correction of header count if mismatch detected
- **System Metrics Enhancement**: Added environment variables to `/api/v1/system/metrics` response
- **Fixed Index Health Metrics**: Updated Storage Engine page to show real index metrics instead of placeholders
- **Storage Components Display**: Replaced non-functional "File System Analysis" with real storage component breakdown
- **API Documentation**: Updated Swagger specifications to include new RBAC metrics structures

### Fixed
- **UUID Storage Format**: Fixed critical authentication bug by increasing EntityID storage from 36 to 64 bytes
  - Resolved login failures caused by truncated UUID values in binary format
  - Updated all entity operations to handle full UUID strings correctly
  - Fixed user authentication and session management issues
- **Index Write Operations**: Fixed critical bug where index entries were written in random order
- **Authentication Failures**: Fixed admin user creation in startup script
- **Build Errors**: Moved debug_auth.go to tools directory to avoid duplicate main
- **Single Source of Truth**: Removed duplicate temporal tag parsing implementations
- **Tag Indexing**: Fixed critical bug where non-timestamped versions of temporal tags weren't indexed
  - Authentication lookups now work correctly with temporal tags
  - Tag index properly handles both timestamped and non-timestamped queries
- **Relationship Storage**: Fixed EntityRelationship to set both Type and RelationshipType fields
- **Password Hashing**: Ensured consistent bcrypt+salt hashing across all authentication paths

## [v2.14.0] - 2025-05-20

### Changed
- **Directory Structure Reorganization**:
  - Removed obsolete directories (share/cli, share/utilities, share/scripts, share/tools)
  - Updated documentation to use API calls instead of CLI tools
  - Fixed references to removed tools in scripts
  - Improved separation of concerns in project structure
- **Testing Improvements**:
  - Added test case files to git with updated .gitignore
  - Ensured all tests are properly tracked in version control
  - Fixed test framework path references
- **Documentation Updates**:
  - Improved README with more accurate feature descriptions
  - Updated performance estimates to more realistic values
  - Removed exaggerated performance claims

## [v2.13.1] - 2025-05-20

### Added
- **API Testing Framework**: Comprehensive testing tools for all API endpoints
  - Added `test_all_endpoints.sh` for complete API validation
  - Created supplementary diagnostic tools for authentication issues
  - Added detailed documentation in `API_TESTING_FRAMEWORK.md`
- **Content Format Documentation**:
  - Added `CONTENT_FORMAT_TROUBLESHOOTING.md` for diagnosing content issues
  - Updated system documentation to reference content format requirements

### Fixed
- **Critical Authentication Issues**:
  - Fixed 500 errors during login caused by incompatible content encoding
  - Resolved user entity content format inconsistencies
  - Improved error handling in authentication system
- **Content Format Standardization**:
  - Standardized content format for user entities
  - Fixed binary content persistence issues
  - Added validation for content format integrity

### Security
- Improved password validation and storage
- Enhanced error reporting without exposing sensitive information
- Fixed potential authentication bypass issues

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

## [v2.12.0] - 2025-05-18

### Added
- **Unified Entity Model**: Single content field ([]byte) per entity
- **Autochunking**: Automatic chunking for large files (>4MB)
- **Content Streaming**: No RAM limits with progressive loading
- **Binary Format Enhancements**:
  - Improved journal format for durability
  - Advanced corruption detection
  - Format version markers for compatibility
  - Automatic content chunking support
- **Entity API Improvements**:
  - Content type detection
  - Base64 content handling
  - Chunked entity retrieval
  - Content size reporting

### Changed
- Simplified entity model to use single binary content field
- Updated serialization format to handle large content efficiently
- Modified entity repository to support chunked storage and retrieval
- Updated entity API to transparently handle chunked content
- Improved error handling for large content operations

### Fixed
- Memory exhaustion with large content uploads
- Content corruption during partial writes
- File size limitations in entity storage
- Inefficient memory usage during entity operations
- Base64 encoding issues with binary content

## [v2.11.1] - 2025-05-17

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

## [v2.11.0] - 2025-05-16

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

## [v2.10.0] - 2025-05-15

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

## [v2.9.0] - 2025-05-14

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

## [v2.8.0] - 2025-05-10

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

## [v2.7.1] - 2025-05-08

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

## [v2.7.0] - 2025-05-05

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

## [v2.6.0] - 2025-05-01

### Added
- Secure session management with TTL
- Session refresh capability
- Automatic session cleanup
- Support for concurrent sessions
- Token generation using crypto/rand
- Session expiration tracking
- Session-based authentication middleware

## [v2.5.0] - 2025-04-15

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

## [0.3.0] - 2025-04-01

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

## [0.2.0] - 2025-03-15

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

## [0.1.0] - 2025-03-01

### Added
- Initial EntityDB platform
- Agent management
- Issue tracking
- Workspace organization
- Basic authentication
- REST API
- Command-line client

[Unreleased]: https://git.home.arpa/itdlabs/entitydb/compare/v2.13.1...HEAD
[v2.13.1]: https://git.home.arpa/itdlabs/entitydb/compare/v2.13.0...v2.13.1
[v2.13.0]: https://git.home.arpa/itdlabs/entitydb/compare/v2.12.0...v2.13.0
[v2.12.0]: https://git.home.arpa/itdlabs/entitydb/compare/v2.11.1...v2.12.0
[v2.11.1]: https://git.home.arpa/itdlabs/entitydb/compare/v2.11.0...v2.11.1
[v2.11.0]: https://git.home.arpa/itdlabs/entitydb/compare/v2.10.0...v2.11.0
[v2.10.0]: https://git.home.arpa/itdlabs/entitydb/compare/v2.9.0...v2.10.0
[v2.9.0]: https://git.home.arpa/itdlabs/entitydb/compare/v2.8.0...v2.9.0
[v2.8.0]: https://git.home.arpa/itdlabs/entitydb/compare/v2.7.1...v2.8.0
[v2.7.1]: https://git.home.arpa/itdlabs/entitydb/compare/v2.7.0...v2.7.1
[v2.7.0]: https://git.home.arpa/itdlabs/entitydb/compare/v2.6.0...v2.7.0
[v2.6.0]: https://git.home.arpa/itdlabs/entitydb/compare/v2.5.0...v2.6.0
[v2.5.0]: https://git.home.arpa/itdlabs/entitydb/compare/v0.3.0...v2.5.0
[0.3.0]: https://git.home.arpa/itdlabs/entitydb/compare/v0.2.0...v0.3.0
[0.2.0]: https://git.home.arpa/itdlabs/entitydb/compare/v0.1.0...v0.2.0
[0.1.0]: https://git.home.arpa/itdlabs/entitydb/releases/tag/v0.1.0