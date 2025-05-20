# EntityDB Relationship Test Suite

This directory contains comprehensive tests for the EntityDB relationship functionality.

## Test Files

### 1. `test_relationships_working.sh`
Basic smoke test that verifies:
- Test endpoints are accessible
- Basic relationship creation
- Query by source and target
- Entity listing includes relationships

### 2. `test_entity_relationships_comprehensive.sh`
Comprehensive test coverage including:
- Various entity types (projects, tasks, users)
- Multiple relationship types (contains, blocks, assigned_to)
- Complex queries (by source, target, type)
- Bidirectional relationships
- Performance testing (50+ relationships)
- Edge cases and error handling

### 3. `test_relationships_rbac.sh`
RBAC enforcement tests:
- Different user roles (admin, editor, user)
- Permission checking for create/view operations
- Wildcard permissions (relation:*)
- Anonymous access denial
- Special permission combinations

### 4. `test_relationships_persistence.sh`
Binary persistence and recovery:
- Data persistence across server restarts
- WAL recovery from crashes
- Temporal data preservation
- Concurrent access handling
- File structure verification

### 5. `run_all_relationship_tests.sh`
Master test runner that executes all relationship tests in sequence.

## Running Tests

### Individual Tests
```bash
cd /opt/entitydb/share/tests
./test_relationships_working.sh
```

### All Tests
```bash
cd /opt/entitydb/share/tests
./run_all_relationship_tests.sh
```

## Test Requirements

1. EntityDB server must be running on port 8085
2. Test endpoints must be enabled (they are in v2.5.0+)
3. Binary database path must be writable
4. curl and jq should be installed for output formatting

## Test Coverage

The test suite covers:
- ✅ Basic CRUD operations for relationships
- ✅ Query operations (by source, target, type)
- ✅ RBAC permission enforcement
- ✅ Binary format persistence
- ✅ WAL recovery mechanisms
- ✅ Temporal data integrity
- ✅ Concurrent access patterns
- ✅ Edge cases and error conditions
- ✅ Performance benchmarks

## Test Data

Tests create various entities and relationships:
- User entities with different roles
- Project and task entities
- Multiple relationship types (contains, blocks, references, etc.)
- Bidirectional relationships
- Chain dependencies

All test data uses unique IDs and can be safely run multiple times.

## Troubleshooting

### Tests Failing with 404
- Ensure server is running with v2.5.0+ which includes the gorilla/mux router fix
- Check that test endpoints are properly registered in main.go
- Verify the server started successfully: `/opt/entitydb/bin/entitydbd.sh status`

### Permission Errors
- Ensure the binary database directory is writable
- Check that RBAC middleware is properly configured
- Verify test tokens are being accepted

### Persistence Issues
- Check WAL files in `/opt/entitydb/var/db/binary/`
- Ensure proper file locking is working
- Verify disk space is available

## Contributing

When adding new relationship features:
1. Add tests to the appropriate test file
2. Update this documentation
3. Ensure all tests pass before committing
4. Consider performance implications for large datasets