# Issue CRUD Functionality Analysis

## Current Status

The Issue CRUD (Create, Read, Update, Delete) functionality in EntityDB currently has several issues that prevent it from working correctly. This document analyzes the problems and outlines what needs to be fixed.

## Identified Problems

### 1. Foreign Key Constraint Failures

When attempting to create an issue, we encountered the following error:
```
{
  "error": "Failed to create issue: failed to insert issue: FOREIGN KEY constraint failed"
}
```

This indicates that the issue creation process is trying to reference foreign keys that don't exist in the database, likely due to:
- Missing workspace records
- Missing user records
- Missing agent records
- Inconsistent ID formats between tables

### 2. Authentication Integration Issues

- The system accepts login credentials and issues a JWT token
- However, this token doesn't appear to be properly validated or utilized in all API endpoints
- The issue creation process isn't correctly using the authenticated user information

### 3. RBAC Permission Problems

- RBAC roles and permissions are defined but not consistently applied
- Some endpoints might not be checking for the required permissions
- Issue creation should validate that the user has the appropriate permissions

## Required Fixes

### Database Schema and Constraints

1. **Analyze Foreign Key Relationships**:
   - Map out all foreign key relationships for the `issues` table
   - Verify that all referenced tables have the necessary records
   - Check if constraints are properly defined in the schema

2. **Ensure Consistent ID Formats**:
   - Review ID generation for all related entities
   - Make sure all ID formats are consistent between referencing tables

### Authentication Integration

1. **Fix User Context Extraction**:
   - Update `/opt/entitydb/src/api/issue.go` to properly extract authenticated user from request context
   - Replace "system" placeholders with actual user information

2. **Ensure Token Validation**:
   - Verify that all API endpoints validate the JWT token
   - Check that authorization headers are properly parsed and validated

### RBAC Implementation

1. **Apply Consistent Permission Checking**:
   - Ensure all issue API endpoints check for appropriate permissions
   - Implement middleware for permission validation across all issue operations
   - Verify role-based access for different issue actions (create, read, update, delete)

### Testing and Validation

1. **Create Comprehensive Test Cases**:
   - Test issue creation with various user roles
   - Test foreign key constraint handling
   - Test authentication failures and successes
   - Test permission validation

2. **Server Restart Testing**:
   - Verify functionality persists after server restart
   - Ensure database state remains consistent

## Implementation Strategy

1. **Database Fixes**:
   - Review and correct schema definitions
   - Add missing reference records if needed
   - Consider temporarily disabling constraints during testing

2. **Authentication Updates**:
   - Fix user context extraction in all issue handlers
   - Ensure consistent authentication across all APIs

3. **RBAC Enhancements**:
   - Apply permission middleware consistently
   - Document required permissions for each operation

4. **Documentation**:
   - Update API documentation
   - Add examples of working issue operations

## Resources

The main files involved in fixing this issue include:

- `/opt/entitydb/src/api/issue.go` - Issue CRUD operations
- `/opt/entitydb/src/api/auth.go` - Authentication handling
- `/opt/entitydb/src/api/rbac_handler.go` - RBAC implementation
- `/opt/entitydb/src/models/sqlite/issue_repository.go` - Database operations for issues
- `/opt/entitydb/src/models/issue.go` - Issue model definition

## Next Steps

1. Begin with a detailed code review of the issue creation workflow
2. Trace the request from API handler through to database operation
3. Identify exact point of failure with foreign key constraints
4. Fix authentication integration issues
5. Implement and test RBAC permission validation
6. Document all changes and fixes