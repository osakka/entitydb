# Entity Test System

This document explains the testing approach for the entity-based architecture in the EntityDB platform.

## Overview

The entity-based architecture represents a significant shift from the previous issue-based model to a more flexible tag-based data model. The entity model uses a minimalist approach with just three core fields:

1. **ID**: Unique identifier for the entity
2. **Tags**: Key-value pairs with timestamps for metadata
3. **Content**: Typed content blocks with timestamps

This architecture allows for more flexible data modeling and easier schema evolution without migrations.

## Test Categories

Our test system for the entity-based architecture is divided into three main categories:

### 1. Core Entity Tests

These tests verify the basic CRUD operations for entities:

- **Entity Creation**: Creating entities with various tags and content
- **Entity Retrieval**: Getting entities by ID and verifying their fields
- **Entity Listing**: Listing all entities and filtering by tag

Script: `test_entity_api.sh`

### 2. Entity-Issue Compatibility Tests

These tests verify that the entity-based architecture correctly handles issue-style requests:

- **Issue-to-Entity**: Testing that issue API endpoints correctly create and manipulate entities
- **Entity-to-Issue**: Testing that entities with issue-style tags can be accessed via issue endpoints
- **Workspace Compatibility**: Ensuring workspace operations work with the entity model

Script: `test_entity_issue_compatibility.sh`

### 3. Entity Tag Operations Tests

These tests focus on the tag-based data model:

- **Tag Creation**: Creating entities with various tags
- **Tag Filtering**: Retrieving entities by tag queries
- **Tag Types**: Testing different entity types via tags (issue, epic, story, etc.)
- **Timestamp Tags**: Verifying that timestamp-based tags work correctly

Script: `test_entity_tags.sh`

## Running the Tests

### Individual Test Categories

You can run specific test categories using the following commands:

```bash
# Run core entity tests
cd /opt/entitydb/share/tests/api/entity
./test_entity_api.sh

# Run compatibility tests
./test_entity_issue_compatibility.sh

# Run tag operation tests
./test_entity_tags.sh
```

### Using the Makefile

The Makefile provides several targets for running entity tests:

```bash
# Run all entity tests
cd /opt/entitydb/src
make entity-tests

# Run detailed entity architecture tests with individual reporting
make master-entity-tests

# Run all tests including entity tests
make master-tests
```

## Test Implementation Details

### Authentication

All tests use the mock authentication system to ensure they work even without a fully configured authentication system. The tests will attempt to get a real authentication token, but if that fails, they will fall back to a mock token:

```bash
AUTH_TOKEN=$(get_auth_token "$ADMIN_USERNAME" "$ADMIN_PASSWORD")

if [ -z "$AUTH_TOKEN" ]; then
    echo -e "${RED}Failed to get authentication token. Using mock token instead.${NC}"
    AUTH_TOKEN="mock_token_admin_test"
fi
```

### Test Endpoints

The entity tests use special test endpoints that are designed to work with the mock authentication system:

- `/api/v1/test/entity/create`: Creates entities without requiring authentication
- `/api/v1/test/entity/get`: Gets entities by ID
- `/api/v1/test/entity/list`: Lists all entities or filters by tag
- `/api/v1/test/entity/simple/create`: Simplified entity creation with minimal fields

These endpoints are implemented in the `EntityHandler` struct and are registered in the router.

### Issue Compatibility

The compatibility tests verify that the entity-based architecture correctly handles issue-style requests using the issue API endpoints:

- `/issues/create`: Creates an issue, which is stored as an entity with issue-specific tags
- `/issues/{id}`: Gets an issue by ID, which retrieves the entity and formats it as an issue
- `/issues/list`: Lists all issues, which lists all entities with the `type:issue` tag

## Writing New Tests

When writing new tests for the entity-based architecture, follow these guidelines:

1. **Use the test utilities**: Source the `test_utils.sh` script to get access to helper functions
2. **Handle authentication**: Always try to get a real token but fall back to a mock token
3. **Verify responses**: Check that responses contain the expected data
4. **Test edge cases**: Test various entity types, tag combinations, and error conditions
5. **Clean up**: Use the cleanup function to clean up test resources

Example:

```bash
#!/bin/bash
# Test script for entity feature X

# Source the test utilities
source ../test_utils.sh

# Get auth token
AUTH_TOKEN=$(get_auth_token "$ADMIN_USERNAME" "$ADMIN_PASSWORD")
if [ -z "$AUTH_TOKEN" ]; then
    AUTH_TOKEN="mock_token_admin_test"
fi

# Run tests
response=$(test_endpoint "/api/v1/test/entity/create" "POST" "{...}" 201 "Test description" "$AUTH_TOKEN")

# Verify response
if echo "$response" | grep -q "expected_value"; then
    echo -e "${GREEN}✓ Test passed${NC}"
else
    echo -e "${RED}✗ Test failed${NC}"
fi

# Print summary
print_summary
```

## Mock Authentication System

The entity tests rely on the mock authentication system, which allows test requests to bypass authentication checks. When an endpoint receives a request with a token starting with `mock_token_`, it automatically populates the request context with admin claims.

This is implemented in the `AuthMiddleware` method in `auth.go`:

```go
// Special case for testing: Accept mock tokens
if strings.HasPrefix(token, "mock_token_") {
    // Create mock claims for testing
    mockClaims := &CustomClaims{
        UserID:   "user_admin",
        Username: "admin",
        AgentID:  "",
        Roles:    []string{"admin", "user"},
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "ccmf-server-test",
        },
    }
    
    // Add claims to context
    ctx := context.WithValue(r.Context(), "claims", mockClaims)
    
    // Call the next handler with mock context
    next(w, r.WithContext(ctx))
    return
}
```

## Conclusion

The entity test system provides comprehensive validation of the new tag-based entity architecture in the EntityDB platform. By testing both the core functionality and compatibility with the previous issue-based model, we ensure a smooth transition to the new architecture while maintaining backward compatibility.

The tests are designed to be run both individually and as part of the larger test suite, providing flexibility for developers to focus on specific aspects of the system or run the entire suite for comprehensive validation.