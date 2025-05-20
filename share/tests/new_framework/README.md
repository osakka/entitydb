# EntityDB Simple Test Framework

A lightweight, shell-based test framework for EntityDB API testing.

## Overview

This framework provides a simple way to test the EntityDB API using single unified test files. Each test consists of a single `.test` file that contains both the request definition and response validation logic.

The framework automatically handles authentication, test execution, and result reporting.

## Features

- Simple shell-based implementation with no external dependencies
- Clean, unified test files for better organization
- Support for test sequences and dependencies
- Automatic authentication handling
- Color-coded test results
- Extensible validation functions
- Backward compatibility with legacy test formats

## Directory Structure

```
/opt/entitydb/share/tests/
├── new_framework/            # Framework files
│   ├── test_framework.sh     # Core framework functions
│   ├── run_tests.sh          # Main test runner
│   └── run_test_sequence.sh  # Example of test sequence
│
└── test_cases/               # Test case definitions
    ├── login_admin.test      # Unified test file
    ├── create_entity.test    # Unified test file
    └── ...
```

## Quick Start

To run all the tests:

```bash
cd /opt/entitydb/share/tests/new_framework
./run_tests.sh --clean --login --all
```

To run a specific test:

```bash
./run_tests.sh --login create_entity
```

To create a new test:

```bash
./run_tests.sh --new user_create POST users/create "Create new user test"
```

## Creating Tests

A test file is a shell script with a `.test` extension that defines both the request and response validation:

```bash
#!/bin/bash
# Test case: Test user creation

# Test description
DESCRIPTION="Test creating a new user"

# Request definition
METHOD="POST"
ENDPOINT="users/create"
HEADERS="-H \"Content-Type: application/json\""
DATA="{\"username\":\"testuser\",\"password\":\"testpassword\",\"roles\":[\"user\"]}"

# Response validation
validate_response() {
  local resp="$1"
  
  # Check if response contains expected fields
  if [[ "$resp" == *"\"id\":"* && "$resp" == *"\"username\":\"testuser\""* ]]; then
    return 0
  fi
  
  return 1
}
```

### Required Elements

Each test file must include:

1. **DESCRIPTION**: A human-readable description of the test
2. **METHOD**: The HTTP method (GET, POST, PUT, DELETE)
3. **ENDPOINT**: The API endpoint path (without the base URL)
4. **validate_response()**: A function that takes the response as input and returns 0 for success, 1 for failure

### Optional Elements

The following elements are optional:

- **HEADERS**: HTTP headers for the request
- **DATA**: The request body (for POST/PUT requests)
- **QUERY**: Query string parameters

## Test Sequences

For tests that depend on each other, use `run_test_sequence.sh` as a template:

```bash
#!/bin/bash
source "$(dirname "$0")/test_framework.sh"

# Login
login

# Create entity and get ID
response=$(create_entity "[\"type:test\"]" "{}")
entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | sed 's/"id":"//')

# Update a test file to use the entity ID
sed -i "s/ENTITY_ID/$entity_id/" "$TEST_DIR/entity_history.test"

# Run tests that depend on that ID
run_test "entity_history"
```

## Command Line Options

```
Usage: ./run_tests.sh [options] [test_name]

Options:
  -h, --help        Show this help message
  -c, --clean       Clean database before testing
  -a, --all         Run all tests
  -d, --dir DIR     Specify test directory
  -l, --login       Perform login before tests
  -n, --new NAME    Create a new test file
```

## Helper Functions

The framework provides several helper functions for common operations:

- **login()**: Authenticates with the API and stores the session token
- **get_entity()**: Retrieves an entity by ID
- **create_entity()**: Creates a new entity

## Troubleshooting

If tests are failing:

1. Check the server logs: `tail -f /opt/entitydb/var/entitydb.log`
2. Run with clean database: `./run_tests.sh --clean --login --all`
3. Verify the API endpoint details in the test files

## Legacy Format Support

For backward compatibility, the framework also supports the old split file format with separate `*_request` and `*_response` files. However, the unified `.test` format is recommended for all new tests.