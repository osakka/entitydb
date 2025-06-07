# EntityDB Source Code

This directory contains the source code for the EntityDB platform. This document provides an overview of the source code structure and development guidelines.

## Directory Structure

```
/src/
├── tests/                    # Test files
│   ├── cases/                # Test case definitions
│   │   ├── entity_as_of.test # Temporal as-of test
│   │   ├── entity_changes.test # Temporal changes test
│   │   ├── entity_diff.test  # Temporal diff test
│   │   └── entity_history.test # Entity history test
│   ├── chunking/             # Chunking test scripts
│   ├── integrity/            # Data integrity tests
│   ├── performance/          # Performance benchmarks
│   ├── temporal/             # Temporal feature tests
│   ├── verification/         # System verification tests
│   ├── test_framework.sh     # Test framework implementation
│   ├── test_temporal_api.sh  # Temporal API test script
│   ├── run_tests.sh          # Test runner
│   └── README.md             # Test documentation
├── api/                      # API handlers and routing
│   ├── entity_handler.go     # Entity API implementation
│   ├── entity_handler_rbac.go # RBAC wrapper for entity API
│   ├── entity_relationship_handler.go # Relationship API
│   ├── relationship_handler_rbac.go # RBAC wrapper for relationships
│   ├── user_handler.go       # User management API
│   ├── user_handler_rbac.go  # RBAC wrapper for user API
│   ├── connection_close_middleware.go # Connection stability middleware
│   ├── te_header_middleware.go # TE header fix middleware
│   ├── trace_middleware.go     # Request tracing middleware
│   ├── trace_context.go        # Trace context management
│   ├── auth_middleware.go    # Authentication middleware
│   ├── rbac_middleware.go    # RBAC enforcement middleware
│   ├── metrics_handler.go    # Prometheus metrics endpoint
│   ├── system_metrics_handler.go # System metrics API
│   ├── rbac_metrics_handler.go # RBAC & session metrics
│   ├── metrics_background_collector.go # Background metrics collection
│   ├── metrics_history_handler.go # Temporal metrics history
│   ├── query_metrics_middleware.go # Query performance tracking
│   ├── error_metrics_collector.go # Error tracking system
│   ├── request_metrics_middleware.go # HTTP request/response metrics
│   ├── router.go             # HTTP router setup
│   └── response_helpers.go   # Response formatting utilities
├── models/                   # Data models
│   ├── entity.go             # Core entity model
│   ├── entity_relationship.go # Relationship model
│   ├── entity_query.go       # Query builder pattern
│   ├── session.go            # Session management
│   ├── tag_namespace.go      # Tag namespace utilities
│   └── errors.go             # Error definitions
├── storage/                  # Storage implementations
│   └── binary/               # Binary format storage
│       ├── entity_repository.go      # Entity storage
│       ├── relationship_repository.go # Relationship storage
│       ├── temporal_repository.go    # Temporal features
│       ├── high_performance_repository.go # Optimized implementation
│       ├── writer.go                 # Binary format writing
│       ├── reader.go                 # Binary format reading
│       ├── format.go                 # Binary format specification
│       ├── wal.go                    # Write-ahead logging
│       ├── mmap_reader.go            # Memory-mapped access
│       └── metrics_instrumentation.go # Storage operation metrics
├── cache/                    # Caching implementations
│   └── query_cache.go        # Query result caching
├── logger/                   # Logging system
│   └── logger.go             # Structured logging
├── tools/                    # Command-line tools
│   ├── users/                # User management tools
│   │   ├── add_user.go       # Add user to the system
│   │   ├── create_users.go   # Create multiple users
│   │   └── generate_hash.go  # Password hash generation
│   ├── entities/             # Entity management tools
│   │   ├── add_entity.go     # Create new entities
│   │   ├── list_entities.go  # List entities with filtering
│   │   ├── dump_entity.go    # Export entity data
│   │   └── add_entity_relationship.go # Create entity relationships
│   ├── maintenance/          # System maintenance tools
│   │   ├── fix_index.go      # Index repair utility
│   │   ├── check_admin_user.go # Admin user verification
│   │   └── scan_entity_data.go # Scan entity data
│   ├── README.md             # Tool documentation
│   └── IMPLEMENTATION.md     # Tool implementation guide
├── main.go                   # Server entry point
└── Makefile                  # Build system
```

## Code Organization Principles

1. **Package Structure**
   - Each package should have a single, well-defined responsibility
   - Avoid circular dependencies between packages
   - Prefer shallow hierarchies over deep nesting

2. **File Naming**
   - Use lowercase with underscores for filenames
   - Group related functionality in the same file
   - Use descriptive, specific names

3. **Interface-Based Design**
   - Define interfaces before implementations
   - Program to interfaces, not concrete types
   - Keep interfaces focused and minimal

4. **Tests and Tools Organization**
   - Tests are stored in `/src/tests` to ensure versioning alongside source code
   - Command-line tools are organized by category in `/src/tools`
   - All compiled tools use the `entitydb_` prefix

5. **Logging Standards**
   - All code uses the structured logger from `entitydb/logger` package
   - Logger automatically provides timestamp, level, file, function, and line information
   - Log levels: TRACE → DEBUG → INFO → WARN → ERROR → FATAL
   - Error messages include contextual information (entity IDs, operation parameters)
   - No manual prefixes or redundant information in log messages
   - See `LOGGING_AUDIT_REPORT.md` for comprehensive standards documentation

## Git Workflow and Protocol

All developers must follow the EntityDB Git workflow guidelines. For detailed information on:

- Branch strategy
- Commit message format and standards
- Pull request protocol
- Git hygiene rules
- State tracking with Git describe
- Tagging conventions
- Daily workflow practices

Please refer to the comprehensive [Git Workflow Guide](/opt/entitydb/docs/development/git-workflow.md).

This document is the centralized source of truth for all Git-related practices in the EntityDB project.

> **Important:** Always move unused or outdated code to the `/trash` directory instead of deleting it (see the Git hygiene rules in the workflow guide).

For the complete list of Git hygiene rules, please refer to the comprehensive [Git Workflow Guide](/opt/entitydb/docs/development/git-workflow.md).

### Git Repository Configuration

- **Repository URL**: https://git.home.arpa/itdlabs/entitydb.git
- **Credentials Management**: Use Git credential helper

## Code Quality Standards

1. **Testing Requirements**
   - All new code must have tests
   - Unit tests for core functionality
   - Integration tests for API endpoints
   - Maintain >80% test coverage

2. **Code Style**
   - Follow Go style conventions (gofmt)
   - Use consistent naming conventions
   - Document all exported functions, types, and constants
   - Keep functions small and focused

3. **Code Review Checklist**
   - Does the code follow our architecture principles?
   - Is the code well-tested?
   - Is there proper error handling?
   - Is the code efficient? Any performance concerns?
   - Is the code secure? Any vulnerability concerns?
   - Is the code maintainable? Clear and readable?

## Build and Test

```bash
# Build the server
make

# Run all tests
make test

# Run only unit tests
make unit-tests

# Run only API tests
make api-tests

# Build and install
make install

# Build all command-line tools
make tools

# Build specific tool categories
make user-tools
make entity-tools
make maintenance-tools

# List available tools and usage examples
make test-utils
```

## Tool Naming Convention

All compiled tools follow the `entitydb_` prefix naming convention, for example:

- `entitydb_add_user`
- `entitydb_list_entities`
- `entitydb_dump`

This convention ensures tools are easily identifiable and prevents naming conflicts. All tools are installed in the `/opt/entitydb/bin` directory.

## Architecture Policy

1. **Pure Entity Model**: All data must be stored as entities with tags
2. **No Direct Database Access**: All operations through the API
3. **Proper Authentication**: All endpoints must use auth middleware
4. **RBAC Enforcement**: All operations must check permissions
5. **Clean Architecture**: Clear separation of concerns