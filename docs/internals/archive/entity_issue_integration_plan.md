# Entity-Issue Integration Plan

This document outlines the plan for integrating the Entity model with the current Issue handling code. The goal is to maintain backward compatibility while transitioning toward the new Entity architecture.

## Current Architecture

1. **Issue Model**: Defined in `/opt/entitydb/src/models/issue.go`, it represents issues with explicit fields and a tag-based approach for types and status.
2. **Issue Repository**: Interface in `/opt/entitydb/src/models/issue.go` and implementation in `/opt/entitydb/src/models/sqlite/issue_repository.go`.
3. **Issue Handler**: API endpoints in `/opt/entitydb/src/api/issue.go` for CRUD operations on issues.
4. **Entity Model**: Defined in `/opt/entitydb/src/models/entity.go`, it represents a generic entity with tags and content.
5. **Entity Repository**: Interface in `/opt/entitydb/src/models/entity.go` and implementation in `/opt/entitydb/src/models/sqlite/entity_repository.go`.
6. **Entity Handler**: API endpoints in `/opt/entitydb/src/api/entity_handler.go` for CRUD operations on entities.
7. **Adapter**: Tag-issue adapter in `/opt/entitydb/src/api/tag_issue_adapter.go` for converting between approaches.

## Integration Strategy

The integration approach will be based on an adapter pattern with these components:

1. **Dual Storage**: Initially, store data in both repositories to minimize risk.
2. **Adapter Layer**: Enhance the existing adapter to convert between models.
3. **Repository Wrapper**: Create a wrapper repository implementing IssueRepository backed by EntityRepository.
4. **API Compatibility**: Support both issue and entity endpoints for the same resources.
5. **Gradual Migration**: Shift from issue endpoints to entity endpoints in phases.

## Implementation Plan

### Phase 1: Enhance Adapter Layer

1. **Implement Entity to Issue Conversion**
   - Create functions to convert Entity to Issue model
   - Ensure all Issue fields can be populated from Entity tags and content
   - Handle special cases like assignments and dependencies

2. **Implement Issue to Entity Conversion**
   - Enhance existing adapter for full conversion of Issue to Entity
   - Map Issue fields to appropriate Entity tags and content
   - Ensure backward compatibility for tag formats

### Phase 2: Create EntityIssueRepository Wrapper

1. **Define Repository Wrapper**
   - Create a new repository implementation that wraps EntityRepository
   - Implement IssueRepository interface using EntityRepository operations
   - Use adapter functions for model conversions

2. **Update Repository Factory**
   - Modify factory to optionally create EntityIssueRepository
   - Add configuration option to enable entity-based repository

### Phase 3: Update API Handlers

1. **Update Issue Handler**
   - Modify issue handlers to work with the repository wrapper
   - Handle entity-specific errors and edge cases
   - Maintain backward compatibility in responses

2. **Enhance Entity Handler**
   - Add issue-specific convenience methods to EntityHandler
   - Implement entity endpoints that mirror issue endpoints

### Phase 4: Test Compatibility

1. **Expand Test Suite**
   - Enhance entity_issue_compatibility.sh tests
   - Test both storage approaches with the same operations
   - Verify data consistency between the two repositories

2. **Performance Testing**
   - Benchmark both approaches
   - Optimize entity repository for issue-specific operations

### Phase 5: Enable Dual Write Mode

1. **Implement Write Operations to Both Repositories**
   - Add configuration option for dual write mode
   - Log any inconsistencies between the two approaches
   - Implement automatic reconciliation for differences

2. **Read from Primary Repository with Fallback**
   - Configure primary repository for reads
   - Fall back to secondary repository on errors or misses

### Phase 6: Gradual Migration

1. **Shift Traffic to Entity Endpoints**
   - Add client-side support for entity endpoints
   - Monitor usage and errors
   - Increase traffic to entity endpoints gradually

2. **Deprecate Issue-specific Code**
   - Mark issue endpoints as deprecated
   - Prepare for removal of duplicate code
   - Document migration path for clients

## Code Structure

### EntityIssueRepository (New)

```go
// EntityIssueRepository implements IssueRepository using EntityRepository
type EntityIssueRepository struct {
    entityRepo models.EntityRepository
}

// Implement all IssueRepository methods using the EntityRepository
func (r *EntityIssueRepository) Create(issue *models.Issue) error {
    // Convert Issue to Entity
    entity := ConvertIssueToEntity(issue)
    
    // Create entity
    return r.entityRepo.Create(entity)
}

// Other method implementations...
```

### Model Conversion Functions (Enhanced)

```go
// ConvertIssueToEntity converts an Issue to an Entity
func ConvertIssueToEntity(issue *models.Issue) *models.Entity {
    entity := models.NewEntity(issue.ID)
    
    // Add type, status, and priority tags
    entity.AddTag("type", issue.GetType())
    entity.AddTag("status", issue.GetStatus())
    entity.AddTag("priority", issue.Priority)
    
    // Add workspace and parent relationships
    if issue.WorkspaceID != "" {
        entity.AddTag("workspace", issue.WorkspaceID)
    }
    if issue.ParentID != "" {
        entity.AddTag("parent", issue.ParentID)
    }
    
    // Add content items
    entity.AddContent("title", issue.Title)
    entity.AddContent("description", issue.Description)
    
    // Additional fields as needed...
    
    return entity
}

// ConvertEntityToIssue converts an Entity to an Issue
func ConvertEntityToIssue(entity *models.Entity) *models.Issue {
    // Extract basic data
    id := entity.ID
    title := entity.GetLatestContent("title")
    description := entity.GetLatestContent("description")
    issueType := entity.GetTagValue("type")
    status := entity.GetTagValue("status")
    priority := entity.GetTagValue("priority")
    workspaceID := entity.GetTagValue("workspace")
    parentID := entity.GetTagValue("parent")
    
    // Create issue with extracted data
    issue := &models.Issue{
        ID:          id,
        Title:       title,
        Description: description,
        Priority:    priority,
        WorkspaceID: workspaceID,
        ParentID:    parentID,
        // Other fields...
    }
    
    // Set tags
    issue.SetType(issueType)
    issue.SetStatus(status)
    
    return issue
}
```

## Timeline and Milestones

1. **Weeks 1-2**: Enhance adapter layer and create repository wrapper
2. **Week 3-4**: Update API handlers and expand test suite
3. **Week 5-6**: Enable dual write mode and test extensively
4. **Week 7-8**: Begin gradual traffic shift to entity endpoints
5. **Week 9-10**: Monitor, fix issues, and document migration path

## Risks and Mitigation

1. **Data Inconsistency**: Use dual writes and automatic reconciliation
2. **Performance Impact**: Optimize entity queries and consider caching
3. **API Compatibility**: Thorough testing of both repositories with the same operations
4. **Migration Complexity**: Provide clear migration guides and helper utilities

## Conclusion

This integration plan provides a staged approach to transitioning from the Issue model to the Entity model while maintaining backward compatibility. By using an adapter pattern and a repository wrapper, we can minimize risk and gradually shift to the new architecture.