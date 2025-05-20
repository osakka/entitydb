# EntityDB Test Suite

This directory contains the testing framework and test cases for EntityDB.

## Test Framework

The EntityDB test framework is a simple, shell-based system for testing the API. It's designed to be:

- **Simple**: Pure shell implementation with no external dependencies
- **Unified**: Single test file per test case containing both request and validation
- **Maintainable**: Clean structure for better readability and maintenance
- **Extensible**: Easy to add new test cases and custom validation logic

## Directory Structure

- `new_framework/` - The core test framework implementation
  - `test_framework.sh` - The main framework library
  - `run_tests.sh` - The primary test runner
  - `run_test_sequence.sh` - Example of dependent test execution
  - `convert_legacy_tests.sh` - Tool to convert legacy tests to the unified format
  
- `test_cases/` - Individual test case definitions
  - `*.test` - Unified test files (containing both request and validation)

## Quick Start

```bash
# Run all tests
cd /opt/entitydb/share/tests/new_framework
./run_tests.sh --clean --login --all

# Run a specific test
./run_tests.sh --login create_entity

# Create a new test
./run_tests.sh --new my_test POST api/endpoint "Test description"
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

For detailed documentation, see the [Test Framework README](/opt/entitydb/share/tests/new_framework/README.md).