# EntityDB Test Suite

This directory contains the testing framework and test cases for EntityDB.

## Test Framework

The EntityDB test framework is a simple, shell-based system for testing the API. It's designed to be:

- **Simple**: Pure shell implementation with no external dependencies
- **Unified**: Single test file per test case containing both request and validation
- **Maintainable**: Clean structure for better readability and maintenance
- **Extensible**: Easy to add new test cases and custom validation logic
- **Performance-Aware**: Detailed timing metrics for identifying bottlenecks

## Directory Structure

- `test_framework.sh` - The core framework implementation
- `run_tests.sh` - The test runner (wrapper for test_framework.sh)
- `test_cases/` - Individual test case definitions
  - `*.test` - Unified test files (containing both request and validation)

## Quick Start

```bash
# Run all tests with timing metrics
cd /opt/entitydb/share/tests
./run_tests.sh --clean --login --all --timing

# Run a specific test
./run_tests.sh --login create_entity

# Create a new test
./run_tests.sh --new my_test POST api/endpoint "Test description"

# Run a sequence of dependent tests
./run_tests.sh --sequence
```

## Command-line Options

```
Usage: ./run_tests.sh [options] [test_name]

Options:
  -h, --help        Show help message
  -c, --clean       Clean database before testing
  -a, --all         Run all tests
  -d, --dir DIR     Specify test directory
  -l, --login       Perform login before tests
  -n, --new NAME    Create a new test file
  -t, --timing      Show timing information
  -s, --sequence    Run a test sequence
```

## Creating Tests

Each test is defined in a single `.test` file:

```bash
#!/bin/bash
# Test case: Test user login

# Test description
DESCRIPTION="Test user login with valid credentials"

# Request definition
METHOD="POST"
ENDPOINT="auth/login"
HEADERS="-H \"Content-Type: application/json\""
DATA="{\"username\":\"admin\",\"password\":\"admin\"}"

# Response validation
validate_response() {
  local resp="$1"
  
  # Check for token in response
  if [[ "$resp" == *"\"token\":"* ]]; then
    return 0
  fi
  
  return 1
}
```

## Performance Metrics

The framework provides detailed timing information when run with the `--timing` flag:

- Total execution time for all tests
- Average test execution time
- Fastest and slowest tests
- Individual test execution times (sorted)

This helps identify bottlenecks and track performance improvements over time.