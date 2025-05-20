# EntityDB Test Suite

This directory contains testing scripts for the EntityDB system.

## Overview

The test suite includes tools for:
- API endpoint testing
- Authentication testing
- Entity creation and manipulation
- Content format testing
- System validation

## Quick Start

To verify all API endpoints are working:

```bash
./test_all_endpoints.sh
```

## Available Test Scripts

### Core Tests

- **test_all_endpoints.sh**: Tests all documented API endpoints
- **test_all.sh**: Runs all available tests

### Authentication Tests

- **create_test_admin.sh**: Creates a test admin user
- **debug_login_500.sh**: Diagnoses login issues that result in 500 errors
- **try_multiple_passwords.sh**: Tests multiple password combinations
- **reset_admin_password.sh**: Resets admin password

### Entity Tests

- **test_simple_entity.sh**: Tests basic entity creation and retrieval
- **test_entity_json_content.sh**: Tests JSON content in entities
- **test_entity_relationships.sh**: Tests entity relationships

### Advanced Tests

- **test_temporal_features.sh**: Tests temporal features (as-of, history, diff)
- **test_high_performance_correct.sh**: Tests high-performance mode correctness
- **test_ssl.sh**: Tests SSL functionality

## Content Format Tests

- **create_simple_format_admin.sh**: Creates an admin user with simplified content format
- **create_unified_admin.sh**: Creates an admin user with unified entity model

## Usage Notes

- Most scripts require the EntityDB server to be running
- Some scripts require admin credentials
- For SSL tests, ensure SSL is properly configured
- For relationship tests, ensure entity relationships are enabled

## Troubleshooting

If tests fail:

1. Check server status (`/opt/entitydb/bin/entitydbd.sh status`)
2. Verify database exists and is accessible
3. Check logs at `/opt/entitydb/var/entitydb.log`
4. Try rebuilding the server (`cd /opt/entitydb/src && make`)

## Documentation

For detailed information on the testing framework, see:
- [API_TESTING_FRAMEWORK.md](/opt/entitydb/docs/API_TESTING_FRAMEWORK.md)
- [CONTENT_FORMAT_TROUBLESHOOTING.md](/opt/entitydb/docs/CONTENT_FORMAT_TROUBLESHOOTING.md)