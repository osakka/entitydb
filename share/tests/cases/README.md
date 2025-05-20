# EntityDB Test Cases

This directory contains test cases for the EntityDB API in the unified `.test` format.

## How to Create a Test

Creating a test is simple. Each test is defined in a single file with a `.test` extension.

### Step 1: Create a new test file

```bash
# Option 1: Use the test framework's built-in command
cd /opt/entitydb/share/tests
./run_tests.sh --new my_test_name POST endpoint/path "Description of my test"

# Option 2: Create the file manually
touch cases/my_test_name.test
```

### Step 2: Define the test content

A test file must contain these components:

```bash
#!/bin/bash
# Test case: Short description

# Test description (displayed in test output)
DESCRIPTION="Full description of what this test checks"

# Request definition
METHOD="POST"                                           # HTTP method (GET, POST, PUT, DELETE)
ENDPOINT="endpoint/path"                                # API endpoint (without base URL)
HEADERS="-H \"Content-Type: application/json\""         # HTTP headers
DATA="{\"key\":\"value\",\"another\":\"value\"}"        # Request body (for POST/PUT)
QUERY="param1=value1&param2=value2"                     # Query parameters

# Response validation
validate_response() {
  local resp="$1"   # Response body passed to function
  
  # Custom validation logic - return 0 for success, 1 for failure
  if [[ "$resp" == *"\"success\":true"* ]]; then
    return 0  # Test passed
  fi
  
  return 1  # Test failed
}
```

### Required Elements

The following elements are required in every test file:

1. **DESCRIPTION**: A human-readable description of the test
2. **METHOD**: The HTTP method to use (GET, POST, PUT, DELETE)
3. **ENDPOINT**: The API endpoint path (without the base URL)
4. **validate_response()**: A function that validates the response

### Optional Elements

The following elements are optional:

- **HEADERS**: HTTP headers for the request
- **DATA**: The request body (for POST/PUT requests)
- **QUERY**: Query string parameters

## Examples

### Authentication Test

```bash
#!/bin/bash
# Test case: User authentication

DESCRIPTION="Test user login with valid credentials"

METHOD="POST"
ENDPOINT="auth/login"
HEADERS="-H \"Content-Type: application/json\""
DATA="{\"username\":\"admin\",\"password\":\"admin\"}"

validate_response() {
  local resp="$1"
  
  # Check for valid token in response
  if [[ "$resp" == *"\"token\":"* ]]; then
    return 0
  fi
  
  return 1
}
```

### Entity Creation Test

```bash
#!/bin/bash
# Test case: Entity creation

DESCRIPTION="Create a new entity with tags and content"

METHOD="POST"
ENDPOINT="entities/create"
HEADERS="-H \"Content-Type: application/json\""
DATA="{\"tags\":[\"type:test\",\"status:active\"],\"content\":{\"description\":\"Test entity\"}}"

validate_response() {
  local resp="$1"
  
  # Check if the entity was created with an ID
  if [[ "$resp" == *"\"id\":"* && "$resp" == *"\"tags\":"* ]]; then
    return 0
  fi
  
  return 1
}
```

### Query Test

```bash
#!/bin/bash
# Test case: Entity listing

DESCRIPTION="List entities with tag filter"

METHOD="GET"
ENDPOINT="entities/list"
QUERY="tag=type:test&limit=10"

validate_response() {
  local resp="$1"
  
  # Check if response is a valid array (even if empty)
  if [[ "$resp" == "[]" || ("$resp" == "["* && "$resp" == *"]") ]]; then
    return 0
  fi
  
  return 1
}
```

## Advanced Validation

You can create more complex validation logic:

```bash
validate_response() {
  local resp="$1"
  
  # Check for presence of specific fields
  if [[ "$resp" != *"\"id\":"* ]]; then
    echo "Missing 'id' field"
    return 1
  fi
  
  # Check content structure
  if [[ "$resp" == *"\"content\":"* ]]; then
    # Check content type
    if [[ "$resp" != *"\"content_type\":\"application/json\""* ]]; then
      echo "Incorrect content type"
      return 1
    fi
  fi
  
  # Extract value and check it
  local status=$(echo "$resp" | grep -o '"status":"[^"]*' | sed 's/"status":"//')
  if [[ "$status" != "active" ]]; then
    echo "Status should be 'active', got '$status'"
    return 1
  fi
  
  # All checks passed
  return 0
}
```

## Dependencies Between Tests

For tests that need an entity ID or other data from previous tests, use the `--sequence` option in the framework, which handles creating entities and passing IDs to dependent tests.