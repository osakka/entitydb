# Release Notes v2.13.1

## Overview

EntityDB v2.13.1 is a maintenance release that addresses critical authentication and content format issues. This release also introduces comprehensive API testing tools.

## Critical Fixes

### Authentication Issues
- Fixed critical authentication issues related to content format in user entities
- Resolved 500 errors during login caused by incompatible content encoding
- Improved error handling in authentication system

### Content Format Standardization
- Standardized content format for user entities
- Added documentation on content format requirements
- Provided tools for diagnosing and fixing content format issues

## New Features

### API Testing Framework
- Added comprehensive API testing framework (`test_all_endpoints.sh`)
- Added diagnostic tools for authentication issues
- Created documentation for API testing methodology
- Improved test coverage for all endpoints

### Documentation
- Added `CONTENT_FORMAT_TROUBLESHOOTING.md` with detailed guidance on content format issues
- Added `API_TESTING_FRAMEWORK.md` documenting the testing tools and methodology
- Updated system documentation to reference content format requirements

## System Requirements

No changes to system requirements from v2.13.0:
- Go 1.18+
- 4GB RAM minimum
- 1GB disk space for database

## Upgrade Notes

This is a safe upgrade from v2.13.0 that does not introduce database format changes. However, if you're experiencing authentication issues, we recommend:

1. Backing up your database files (`/opt/entitydb/var/entities.ebf`, `/opt/entitydb/var/entitydb.wal`)
2. Stopping the service (`/opt/entitydb/bin/entitydbd.sh stop`)
3. Removing the database files
4. Restarting the service (`/opt/entitydb/bin/entitydbd.sh start`)
5. Recreating your users

## Known Issues

- Token format in authorization header requires no prefix (not "Bearer")
- Temporal query parameters require RFC3339 format and may return errors with other formats
- Relationship creation requires both entities to exist

## Documentation

Full documentation is available in the `/opt/entitydb/docs/` directory:
- [CONTENT_FORMAT_TROUBLESHOOTING.md](/opt/entitydb/docs/CONTENT_FORMAT_TROUBLESHOOTING.md)
- [API_TESTING_FRAMEWORK.md](/opt/entitydb/docs/API_TESTING_FRAMEWORK.md)
- [CLAUDE.md](/opt/entitydb/CLAUDE.md)