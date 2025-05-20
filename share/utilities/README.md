# EntityDB Tools

This directory contains utility tools and test implementations for the EntityDB platform.

## Entity Management Tools

These tools manage the entity-based data model, which provides a flexible, tag-based approach to storing and relating different types of objects in the system.

### Entity Management

- **add_entity.go**: Creates a new entity with specified type, title, and tags
  ```
  go run add_entity.go <type> <title> [tag1:value1 tag2:value2 ...]
  ```
  Example: `go run add_entity.go workspace "New Workspace" status:active owner:admin`

- **list_entities.go**: Lists entities with various filtering options
  ```
  go run list_entities.go [--tag=<tag>] [--type=<type>] [--id=<id>] [--no-content] [--no-tags]
  ```
  Example: `go run list_entities.go --type=workspace --tag=status:active`

### Entity Relationships

- **add_entity_relationship.go**: Creates a relationship between two entities
  ```
  go run add_entity_relationship.go <source_id> <relationship_type> <target_id> [metadata_key1=value1 metadata_key2=value2 ...]
  ```
  Example: `go run add_entity_relationship.go entity_123 depends_on entity_456 dependency_type=blocker description="Entity 123 depends on Entity 456"`

- **list_entity_relationships.go**: Lists relationships with various filtering options
  ```
  go run list_entity_relationships.go [--source=<id>] [--target=<id>] [--type=<type>] [--no-metadata]
  ```
  Example: `go run list_entity_relationships.go --source=entity_123 --type=depends_on`

### Migration Tools

- **migrate_issues_to_entities.go**: Migrates legacy issues to the entity-based model
  ```
  go run migrate_issues_to_entities.go [--dry-run]
  ```
  The `--dry-run` flag shows what would be migrated without making changes.

## User Management Tools

- **add_user.go** - Command-line tool to add a single user to the system
  ```
  go run add_user.go <username> <password> <email> <full_name>
  ```

- **create_users.go** - Batch user creation utility for setting up test users
  ```
  go run create_users.go
  ```

## Test Utilities

- **run_api_tests.sh** - API test runner with support for entity, RBAC, and other component tests
  ```
  ./run_api_tests.sh
  ```

- **update_api_tests.sh** - Updates the API test script with the improved version
  ```
  ./update_api_tests.sh
  ```

## Entity Test Implementations

The `tests/` directory contains test implementations for the entity-based architecture:

- `entity_repository_test_wrapper.go` - Mock implementation of entity repository for testing
- `entity_relationships_test.go` - Tests for entity relationship functionality
- `test_entity_repository_impl.go` - In-memory entity repository for testing
- `test_entity_issue_handler.go` - Tests for entity-based issue handlers
- `test_entity_issue_integration.go` - Integration tests for entity-based issues
- `test_entity_issue_repo.go` - Tests for entity-based issue repository
- `test_entity_relationships.go` - Tests for entity relationship functionality
- `test_entity_repository_impl.go` - Tests for entity repository implementation

These tests ensure proper functionality of the entity model which is the foundation for the next generation of EntityDB's data architecture.

## Building Tools

Tools can be built into binaries using the Makefile in the src directory:

```bash
cd /opt/entitydb/src
make tools
```

This will compile all the Go tools and place the binaries in the /opt/entitydb/bin directory.

## Integration with Main System

These tools integrate with the main EntityDB system and require a running EntityDB server and database to function properly.

Prerequisites:
1. Ensure the EntityDB server is running (`/opt/entitydb/bin/entitydbd.sh status`)
2. Verify the database is initialized (`ls -l /opt/entitydb/var/db/entitydb.db`)
3. Check that entity tables are created in the database schema

For more information about the entity-based architecture, see the following documentation:
- `/opt/entitydb/docs/entity_model_architecture.md`: Overview of the entity model architecture
- `/opt/entitydb/docs/entity_migration_guide.md`: Guide for migrating from legacy models to entities