# RBAC Metrics Authentication Issue

## Problem Description

The RBAC metrics endpoint (`/api/v1/rbac/metrics`) returns a 401 Unauthorized error even when authenticated as admin. This is due to the admin user entity not having the proper `rbac:role:admin` tag assigned.

## Root Cause

1. The admin user is created during initialization but the RBAC tags are not properly assigned
2. The RBAC middleware checks for `admin:view` permission which requires either:
   - `rbac:role:admin` tag on the user
   - `rbac:perm:admin:view` tag on the user
   - `rbac:perm:*` tag on the user

## Current State

- Admin user can login successfully
- Admin user receives a valid session token
- Session token is accepted for other endpoints
- RBAC metrics endpoint specifically requires admin permissions which the user doesn't have

## Attempted Fixes

1. Created `fix_admin_rbac.go` tool to add the tag manually
2. Tool couldn't find admin user due to data structure differences

## Workaround

For now, the RBAC metrics page shows an appropriate error message to users. The functionality will work once the admin user has proper permissions assigned.

## Permanent Fix Required

1. Update the security initialization code to properly assign `rbac:role:admin` tag to admin user
2. Ensure the tag is stored in the temporal format used by the system
3. Update the authentication handler to properly load user roles from tags

## Testing

Once fixed, test with:
```bash
# Login as admin
TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r .token)

# Access RBAC metrics
curl -s -X GET http://localhost:8085/api/v1/rbac/metrics \
  -H "Authorization: Bearer $TOKEN" | jq
```

The endpoint should return metrics instead of a 401 error.