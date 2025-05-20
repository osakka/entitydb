# EntityDB Authentication Integration

This document describes the implementation of the integrated authentication system for EntityDB.

## Overview

The authentication system provides standard REST endpoints for user authentication that work with the entity-based architecture while maintaining backward compatibility with existing clients.

## Authentication Endpoints

The following endpoints are implemented:

- `/api/v1/auth/login` - Authenticates users and issues tokens
- `/api/v1/auth/status` - Checks token validity and returns user info
- `/api/v1/auth/logout` - Invalidates tokens
- `/api/v1/auth/refresh` - Issues new tokens

## Integration Details

The authentication system is directly integrated into the main server code (`server_db.go`), avoiding any cross-origin issues. It uses the existing user storage and token management functions.

## Testing

Test pages are available at:
- `/final_test.html` - Comprehensive test for all authentication features
- `/simple_debug.html` - Simple debugging tool for auth endpoints

## Default Credentials

- Username: `admin`
- Password: `password`

Other available users:
- `osakka` / `mypassword` (admin)
- `regular_user` / `password123` (regular user)
- `readonly_user` / `password123` (read-only user)

## Implementation Notes

The implementation follows these principles:
1. Keeps the entity-based architecture intact
2. Provides standard authentication endpoints
3. Maintains backward compatibility
4. Integrates seamlessly with the UI

## Maintenance

For future maintenance, all authentication logic is centralized in the `handleAuth` function and its helper methods in `server_db.go`.
