# EntityDB Source Code

This directory contains the source code for the EntityDB platform. This document provides an overview of the source code structure and development guidelines.

## Directory Structure

```
/src/
├── api/                      # API handlers and routing
│   ├── entity_handler.go     # Entity API implementation
│   ├── entity_handler_rbac.go # RBAC wrapper for entity API
│   ├── entity_relationship_handler.go # Relationship API
│   ├── relationship_handler_rbac.go # RBAC wrapper for relationships
│   ├── user_handler.go       # User management API
│   ├── user_handler_rbac.go  # RBAC wrapper for user API
│   ├── auth_middleware.go    # Authentication middleware
│   ├── rbac_middleware.go    # RBAC enforcement middleware
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
│       └── mmap_reader.go            # Memory-mapped access
├── cache/                    # Caching implementations
│   └── query_cache.go        # Query result caching
├── logger/                   # Logging system
│   └── logger.go             # Structured logging
├── tools/                    # Development tools
│   ├── fix_index.go          # Index repair utility
│   └── check_admin_user.go   # Admin user verification
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
```

## Architecture Policy

1. **Pure Entity Model**: All data must be stored as entities with tags
2. **No Direct Database Access**: All operations through the API
3. **Proper Authentication**: All endpoints must use auth middleware
4. **RBAC Enforcement**: All operations must check permissions
5. **Clean Architecture**: Clear separation of concerns