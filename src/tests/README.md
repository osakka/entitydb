# EntityDB Tests

This directory contains all tests for the EntityDB platform. Tests are now located here in the `src` directory to:

1. Ensure tests are versioned alongside source code
2. Maintain a consistent development workflow
3. Simplify the build and test process
4. Keep all maintenance-related files in one central location

## Test Organization

- `test_framework.sh`: The main test framework for API testing
- `test_temporal_api.sh`: Tests for the temporal API functionality
- `cases/`: Contains individual test case files
- `temporal/`: Tests for temporal features and timestamp handling 
- `chunking/`: Tests for large file handling and autochunking
- `performance/`: Performance and stress tests
- `integrity/`: Data integrity and basic functionality tests
- `run_all_tests.sh`: Comprehensive test suite runner
- `run_extended_tests.sh`: Extended tests for specific features

## Running Tests

Tests can be run using the Makefile in the parent directory:

```bash
cd /opt/entitydb/src
make test          # Run all tests
make unit-tests    # Run Go unit tests only
make api-tests     # Run API tests
```

Or run a specific test directly:

```bash
cd /opt/entitydb/src/tests
./test_temporal_api.sh
./run_all_tests.sh 
./run_extended_tests.sh
```

## Test Categories

### Core API Tests

The tests in `cases/` directory verify the basic API functionality:
- Entity creation, retrieval, update, and deletion
- Authentication and authorization
- Relationship management
- Configuration management

### Temporal Tests

Tests in the `temporal/` directory focus on:
- Temporal tag handling
- Time-based queries (as-of, history, diff)
- Timestamp storage and retrieval

### Chunking Tests

Tests in the `chunking/` directory focus on:
- Large file handling
- Automatic content chunking
- Streaming uploads and downloads

### Performance Tests

Tests in the `performance/` directory focus on:
- Stress testing
- Concurrency handling
- Large dataset performance

### Integrity Tests

Tests in the `integrity/` directory focus on:
- Data consistency
- Error handling and recovery
- Edge cases and boundary conditions

## Temporal API Testing

The `test_temporal_api.sh` script tests the core temporal features of EntityDB:

1. Entity history retrieval
2. Entity as-of queries (retrieving entity state at a specific point in time)
3. Entity changes (listing changes between timestamps)
4. Entity diff (comparing entity states between timestamps)

The temporal API is a key feature that enables:
- Point-in-time data querying
- Auditing
- Change tracking
- Rollback capabilities