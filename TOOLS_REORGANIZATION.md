# EntityDB Tool and Testing Reorganization

## Summary of Changes

The EntityDB project has been reorganized to improve consistency, maintainability, and developer experience:

1. **Tests moved to src/tests**
   - All test files are now located in `/opt/entitydb/src/tests`
   - The previous location (`share/tests`) has been removed
   - Test files are versioned alongside the source code
   - Clearer documentation and organization

2. **Tools consolidated in src/tools**
   - All command-line tools moved to categorized directories:
     - `src/tools/users/`: User management tools
     - `src/tools/entities/`: Entity management tools
     - `src/tools/maintenance/`: System maintenance tools
   - Non-Go tools have been removed in favor of Go implementations

3. **Standardized tool naming convention**
   - All compiled tools are now prefixed with `entitydb_`
   - Examples: `entitydb_add_user`, `entitydb_dump`, `entitydb_fix_index`
   - Improves discoverability and avoids naming conflicts

## Using the New Structure

### Building Tools

```bash
cd /opt/entitydb/src

# Build all tools
make tools

# Build specific tool categories
make user-tools
make entity-tools
make maintenance-tools
```

### Running Tests

```bash
cd /opt/entitydb/src

# Run all tests
make test 

# Run specific test categories
make unit-tests
make api-tests

# Or use the test runner directly
cd /opt/entitydb/src/tests
./run_tests.sh --all
./run_tests.sh --api
./run_tests.sh --temporal
```

## Tool Documentation

For a full list of available tools and usage examples:

```bash
cd /opt/entitydb/src
make test-utils
```

All compiled tools are placed in `/opt/entitydb/bin/` with the `entitydb_` prefix.

## Example Tool Usage

```bash
# Add a new user
/opt/entitydb/bin/entitydb_add_user -username admin -password securepass

# List entities
/opt/entitydb/bin/entitydb_list_entities -type user

# Dump entity data
/opt/entitydb/bin/entitydb_dump -id abc123 -format pretty
```