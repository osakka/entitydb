# EntityDB Simple Test Framework

A lightweight, shell-based test framework for EntityDB API testing.

## Overview

This framework provides a simple way to test the EntityDB API using request/response pairs. Each test consists of two files:
- `test_name_request` - Defines the API request details
- `test_name_response` - Defines the response validation criteria

The framework automatically handles authentication, test execution, and result reporting.

## Features

- Simple shell-based implementation with no external dependencies
- Clear separation of request and response validation
- Support for test sequences and dependencies
- Automatic authentication handling
- Color-coded test results
- Extensible validation functions

## Directory Structure

```
/opt/entitydb/share/tests/
├── new_framework/            # Framework files
│   ├── test_framework.sh     # Core framework functions
│   ├── run_tests.sh          # Main test runner
│   └── run_test_sequence.sh  # Example of test sequence
│
└── test_cases/               # Test case definitions
    ├── login_admin_request   # Request definition
    ├── login_admin_response  # Response validation
    ├── create_entity_request
    ├── create_entity_response
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
./run_tests.sh --login list_entities
```

## Creating Tests

### 1. Create a Request File

Create a file named `your_test_name_request` in the test_cases directory:

```bash
# Description: Test entity creation
METHOD="POST"
ENDPOINT="entities/create"
HEADERS="-H \"Content-Type: application/json\""
DATA="{\"tags\":[\"type:test\"],\"content\":{\"test\":\"value\"}}"
```

### 2. Create a Response File

Create a file named `your_test_name_response` in the test_cases directory:

```bash
# Simple validation using markers
SUCCESS_MARKER="\"id\":"

# Optional: Custom validation function
validate_response() {
  local resp="$1"
  
  if [[ "$resp" == *"\"id\":"* && "$resp" == *"\"tags\":"* ]]; then
    return 0  # Success
  fi
  
  return 1  # Failure
}
```

## Validation Options

You can validate responses in multiple ways:

1. **Success Marker**: Define `SUCCESS_MARKER` to check for the presence of a string
2. **Error Marker**: Define `ERROR_MARKER` to ensure a string is NOT present
3. **Custom Validation**: Define a `validate_response()` function for complex logic

## Test Sequences

For tests that depend on each other, use the `run_test_sequence.sh` script as a template:

```bash
#!/bin/bash
source "$(dirname "$0")/test_framework.sh"

# Login
login

# Create entity and get ID
response=$(create_entity "[\"type:test\"]" "{}")
entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | sed 's/"id":"//')

# Run a test that uses that ID
# [...]
```

## Help and Options

To see all available options:

```bash
./run_tests.sh --help
```

## Troubleshooting

If tests are failing:

1. Check the server logs: `tail -f /opt/entitydb/var/entitydb.log`
2. Run with clean database: `./run_tests.sh --clean --login --all`
3. Verify the API endpoint details in the request files