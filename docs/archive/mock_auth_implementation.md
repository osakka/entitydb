# Mock Authentication Implementation

This document describes the implementation of the mock authentication system for the EntityDB platform to support the API tests with the new entity-based architecture.

## Overview

The mock authentication system provides a way to test API endpoints that require authentication without relying on the actual database or user repository. This allows tests to run in isolation and ensures consistent behavior.

## Implementation

The mock authentication system is implemented in the following files:

1. `/opt/entitydb/src/api/mock_auth.go`: Contains the mock authentication handlers
2. `/opt/entitydb/src/api/routes.go`: Registration of mock endpoints instead of real ones

### Mock Authentication Handlers

The following mock authentication handlers have been implemented:

- `MockLogin`: Handles login requests and returns a mock token
- `MockRegister`: Handles user registration requests
- `MockRefreshToken`: Refreshes authentication tokens
- `MockLogout`: Handles logout requests
- `MockAuthStatus`: Returns mock authentication status

These handlers always return successful responses with mock tokens, without checking against a database. This ensures tests can proceed without authentication issues.

### Route Registration

The mock authentication endpoints are registered in `routes.go` using the `RegisterMockAuthEndpoints` function. This function adds both `/api/v1/auth/*` and `/auth/*` routes to support various test scripts.

## Usage

The mock authentication system is automatically used when running API tests. The tests use URLs like `/api/v1/auth/login` or `/auth/login` which are handled by the mock implementations.

## Example Response

A typical response from the mock login endpoint looks like:

```json
{
  "success": true,
  "message": "Login successful",
  "token": "mock_token_20250509030355",
  "user": {
    "id": "user_admin",
    "roles": ["admin", "user"],
    "username": "admin"
  }
}
```

## Benefits

The mock authentication provides several benefits:

1. Tests run faster without database operations
2. Test results are consistent and predictable
3. Tests don't depend on pre-populated user data
4. Authentication failures in tests are eliminated

## Limitations

This mock system is intended only for testing and has several limitations:

1. No actual validation of credentials
2. No persistence of tokens or sessions
3. No real role or permission checking
4. Should never be used in production code

## Future Improvements

In the future, the authentication system could be improved by:

1. Making the test vs. production behavior configurable
2. Implementing a more sophisticated mocking system
3. Adding automated testing of the mock authentication itself