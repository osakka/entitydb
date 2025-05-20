# EntityDB Test Suite

This directory contains the testing framework and test cases for EntityDB.

## Test Framework

The EntityDB test framework is a simple, shell-based system for testing the API. It's designed to be:

- **Simple**: Pure shell implementation with no external dependencies
- **Maintainable**: Clear separation of request and response validation
- **Extensible**: Easy to add new test cases and custom validation logic

## Directory Structure

- `new_framework/` - The core test framework implementation
  - `test_framework.sh` - The main framework library
  - `run_tests.sh` - The primary test runner
  - `run_test_sequence.sh` - Example of dependent test execution
  - `migrate_legacy_tests.sh` - Tool to convert old tests to the new format
  
- `test_cases/` - Individual test case definitions
  - `*_request` - API request definition files
  - `*_response` - Response validation files

## Quick Start

```bash
# Run all tests
cd /opt/entitydb/share/tests/new_framework
./run_tests.sh --clean --login --all

# Run a specific test
./run_tests.sh --login create_entity
```

## Creating Tests

1. Create a request file (`test_name_request`) defining the API call
2. Create a response file (`test_name_response`) defining validation criteria
3. Run the test using the test runner

For detailed documentation, see the [Test Framework README](/opt/entitydb/share/tests/new_framework/README.md).