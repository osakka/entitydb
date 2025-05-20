# EntityDB API Testing Framework

This document describes the API testing framework for EntityDB that ensures all endpoints are functioning correctly.

## Overview

The testing framework provides a comprehensive set of scripts to validate all API endpoints documented in [CLAUDE.md](/opt/entitydb/CLAUDE.md). The framework allows for:

- Verifying endpoint existence and accessibility
- Testing authentication and authorization
- Creating and manipulating test entities
- Testing temporal features
- Testing relationship features
- Verifying administrative functions

## Main Test Script

The primary test script is located at `/opt/entitydb/share/tests/test_all_endpoints.sh`. This script:

1. Tests all documented API endpoints
2. Reports success/failure for each endpoint
3. Provides detailed output for debugging
4. Works with authentication or in a no-auth mode

### Usage

```bash
cd /opt/entitydb/share/tests
./test_all_endpoints.sh
```

### Output Format

The script outputs detailed information about each endpoint test:

```
=== Entity Operations ===

Testing: Create entity
✓ ENDPOINT EXISTS Create entity (Status: 200)

Testing: List entities
✓ ENDPOINT EXISTS List entities (Status: 200)
```

## Tested Endpoints

The framework tests the following categories of endpoints:

### Authentication
- `POST /api/v1/auth/login` - User login
- Authentication token handling

### Entity Operations
- `POST /api/v1/entities/create` - Create entity
- `GET /api/v1/entities/list` - List entities
- `GET /api/v1/entities/get` - Get entity by ID
- `PUT /api/v1/entities/update` - Update entity
- `GET /api/v1/entities/query` - Query entities

### Temporal Operations
- `GET /api/v1/entities/as-of` - Get entity as of timestamp
- `GET /api/v1/entities/history` - Get entity history
- `GET /api/v1/entities/changes` - Get recent changes
- `GET /api/v1/entities/diff` - Get entity diff

### Relationship Operations
- `POST /api/v1/entity-relationships` - Create relationship
- `GET /api/v1/entity-relationships` - Get relationships

### Admin Operations
- `POST /api/v1/users/create` - Create user
- `GET /api/v1/dashboard/stats` - Get dashboard stats
- `GET /api/v1/config` - Get config
- `POST /api/v1/feature-flags/set` - Set feature flag

### API Documentation
- `GET /swagger/` - Access Swagger UI
- `GET /swagger/doc.json` - Get OpenAPI specification

## Supplementary Test Scripts

Additional scripts are available for specific testing scenarios:

### `create_test_admin.sh`
Creates a test admin user with a known password.

### `debug_login_500.sh`
Diagnoses 500 errors during login attempts.

### `try_multiple_passwords.sh`
Tests multiple password combinations to find working credentials.

### `reset_admin_password.sh`
Attempts to reset the admin password directly in the database.

### `create_simple_format_admin.sh` and `create_unified_admin.sh`
Creates admin users with different content formats to test compatibility.

## Testing Methodology

The framework follows these testing principles:

1. **Independence**: Each endpoint is tested independently
2. **Error Handling**: Expected errors are appropriately handled
3. **Authentication Awareness**: Tests work with or without authentication
4. **Fallbacks**: Tests use fallback approaches if primary tests fail
5. **Comprehensive Coverage**: All documented endpoints are tested

## Extending the Framework

To add tests for new endpoints:

1. Identify the endpoint to test
2. Determine appropriate test data
3. Add a `make_request` call in the appropriate section
4. Handle the response and report success/failure

## Troubleshooting

If tests fail, consider these common issues:

- **Authentication**: Ensure admin credentials are correct
- **Server Status**: Verify server is running (`/opt/entitydb/bin/entitydbd.sh status`)
- **Port Configuration**: Check that the server is using the expected port (default: 8085)
- **Content Format**: Ensure entity content follows the expected format (see CONTENT_FORMAT_TROUBLESHOOTING.md)

## Best Practices

1. Run tests after any system changes
2. Run tests after database resets
3. Review failing tests carefully - the nature of the failure can provide valuable diagnostic information
4. Use the supplementary scripts for deeper diagnosis when needed