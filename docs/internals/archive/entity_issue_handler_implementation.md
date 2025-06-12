# Entity-Issue Handler Implementation

This document describes the implementation of the `EntityIssueHandler` that enables dual-write mode and support for the entity-based architecture while maintaining compatibility with existing issue-based APIs.

## Overview

The `EntityIssueHandler` extends the base `IssueHandler` to transparently route API requests to both the traditional issue repository and the new entity repository. This allows for a gradual transition from the issue-based model to the entity-based architecture.

## Implementation Details

The implementation consists of:

1. **Entity-Issue Adapter**: Converts between Issue and Entity models
2. **Entity-Issue Repository**: Implements IssueRepository using EntityRepository
3. **Entity-Issue Handler**: Extends IssueHandler to support dual-write mode

### Entity-Issue Handler

The `EntityIssueHandler` embeds the original `IssueHandler` and extends it with:

1. A reference to the entity repository
2. Dual-write mode flag
3. Override methods for all issue operations

```go
type EntityIssueHandler struct {
    *IssueHandler           // Embed the original IssueHandler
    entityRepo models.EntityRepository // Reference to the entity repository
    dualWrite  bool         // Flag to enable dual write mode
}
```

### Dual-Write Mode

When dual-write mode is enabled:

1. Operations write to both repositories
2. Reads check both repositories and reconcile data
3. Discrepancies are logged for debugging

This ensures data consistency during the transition period and provides a safety net against data loss.

### Issue Operations

The handler overrides all issue operations:

- CreateIssue, GetIssue, UpdateIssue, ListIssues
- AssignIssue, UnassignIssue
- StartIssue, UpdateIssueProgress, CompleteIssue
- BlockIssue, UnblockIssue
- ListDependencies, AddDependency, RemoveDependency
- ListWorkspaces, GetWorkspace, CreateWorkspace, UpdateWorkspace

Each operation:

1. Performs parameter validation
2. Checks if dual-write mode is enabled
3. If so, performs the operation on both repositories
4. If not, delegates to the original IssueHandler
5. Logs any discrepancies between repositories

### Integration with Router

To use the `EntityIssueHandler`, register it in place of the `IssueHandler` in the router:

```go
// Create handlers
issueHandler := NewIssueHandler(issueRepo, auth)
entityIssueHandler := NewEntityIssueHandler(issueRepo, entityRepo, auth, true)

// Register routes using entityIssueHandler instead of issueHandler
router.GET("/api/v1/issues/list", entityIssueHandler.ListIssues)
router.GET("/api/v1/issues/get", entityIssueHandler.GetIssue)
// ... and so on for all issue and workspace routes
```

## Testing

A comprehensive test suite has been created to verify:

1. Basic operations (create, retrieve, update)
2. Status changes (start, complete, block, unblock)
3. Assignment operations (assign, unassign)
4. Workspace operations (list, get, create, update)
5. Dual-write consistency

The tests use in-memory repositories to ensure fast and reliable testing.

## Rollout Strategy

The recommended rollout strategy is:

1. Deploy the handler with dual-write mode enabled
2. Monitor for discrepancies and fix any issues
3. Gradually transition to entity-only operations
4. Eventually disable dual-write mode when confident in entity repository

## Performance Considerations

While dual-write mode adds some overhead, it provides important safety guarantees during the transition period. The performance impact is expected to be minimal for most operations.

Operations that may have higher impact:

1. CreateIssue - Writes to both repositories
2. ListIssues - Needs to reconcile data from both repositories
3. UpdateIssue - Requires careful synchronization

## Conclusion

The `EntityIssueHandler` provides a robust solution for transitioning from the issue-based model to the entity-based architecture while maintaining backward compatibility. It enables a gradual migration path that minimizes risk and ensures data consistency.