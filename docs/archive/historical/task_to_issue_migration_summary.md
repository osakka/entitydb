# Task to Issue Migration Summary

## Overview

This document outlines the migration of the EntityDB codebase from using "task" terminology to "issue" terminology, providing backward compatibility while also allowing for a gradual transition to the new API.

## Completed Changes

1. **Model Layer**:
   - Created `/opt/entitydb/src/models/issue.go` with equivalent functionality to `task.go`
   - Renamed all types: Task → Issue, TaskType → IssueType, etc.
   - Updated constants and function names appropriately
   - Created a new IssueRepository interface

2. **API Handler Layer**:
   - Created `/opt/entitydb/src/api/issue.go` by adapting `task.go`
   - Created `/opt/entitydb/src/api/issue_hierarchy_handler.go` to handle epic/story relationships
   - Created `/opt/entitydb/src/api/issue_metrics_handler.go` for tracking issue metrics
   - Created `/opt/entitydb/src/api/issue_pool_assignment.go` for agent pool assignment

3. **Router Configuration**:
   - Created `/opt/entitydb/src/api/updated_routes.go` that maintains all legacy task endpoints while adding new issue endpoints
   - Used identical permission model for both task and issue operations
   - Added dedicated workspace endpoints via the issue API

4. **Dashboard Integration**:
   - Created `/opt/entitydb/src/api/updated_dashboard_handler.go` with support for both task and issue repositories
   - Implemented fallback mechanisms to use task data if issue data isn't available
   - Added helper function to convert tasks to issues for compatibility

## Next Steps

1. **Repository Implementation**:
   - Implement a concrete SQLite-backed `IssueRepository` class
   - Create database migrations to support both tables during transition
   - Consider implementing synchronization between task and issue tables initially

2. **Client-Side Updates**:
   - Update client commands to use issue API endpoints
   - Keep support for the old task commands but mark them as deprecated

3. **Test Coverage**:
   - Add unit and integration tests for the new issue APIs
   - Test both task and issue paths to ensure backward compatibility works

4. **Documentation**:
   - Update system documentation with new issue terminology
   - Mark task API as deprecated and provide migration guides

5. **Full Migration**:
   - After a stable period, consider removing the legacy task API
   - When ready, remove all task-related code and simplify the codebase

## Migration Strategy

The recommended migration approach is:

1. **Parallel Operation**: Initially operate both task and issue APIs side by side
2. **New Feature Preference**: Implement all new features using the issue API
3. **Client Migration**: Update clients to prefer issue API endpoints
4. **Legacy Support**: Maintain task API compatibility for several release cycles
5. **Eventual Removal**: Once all clients have migrated, remove task API entirely

## Benefits

This migration to "issue" terminology:
- Better aligns with industry standard terminology
- Creates a cleaner API for new features without breaking existing clients
- Allows for a gradual, controlled transition
- Provides an opportunity to improve API design and interfaces