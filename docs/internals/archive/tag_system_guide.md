# EntityDB Tag-Based Architecture Guide

## Overview

The EntityDB system now uses a flexible tag-based architecture for classifying and organizing all entities. This document provides an overview of the tag system, how it works, and best practices for using it.

## Core Concepts

### Tags

Tags are lightweight labels attached to entities in the system. Each tag has a format of `namespace:value` where:

- **namespace**: A category or type of metadata (e.g., `type`, `status`, `area`)
- **value**: The specific value for that namespace (e.g., `issue`, `pending`, `frontend`)

Examples of complete tags:
- `type:issue`
- `status:in_progress`
- `area:security`
- `priority:high`
- `assignee:claude-2`

### Tag Namespaces

Common namespaces include:

 < /dev/null |  Namespace | Description | Example Values |
|-----------|-------------|----------------|
| `type` | Entity type | `workspace`, `epic`, `story`, `issue`, `subissue` |
| `status` | Current status | `pending`, `in_progress`, `blocked`, `completed` |
| `area` | Functional area | `frontend`, `backend`, `security`, `ui`, `database` |
| `priority` | Issue priority | `high`, `medium`, `low` |
| `assignee` | Assigned agent | `claude-2`, `agent_123` |
| `milestone` | Target milestone | `v1.0`, `v2.0`, `q3_release` |
| `customer` | Associated customer | `acme`, `globex` |
| `team` | Responsible team | `frontend`, `backend`, `support` |
| `parent` | Parent issue ID | `issue_12345` |
| `complexity` | Implementation complexity | `easy`, `medium`, `hard` |
| `sprint` | Sprint assignment | `current`, `next`, `backlog` |
| `component` | System component | `api`, `ui`, `auth`, `database` |

### Benefits of Tag-Based Architecture

1. **Flexibility**: Add new classifications without schema changes
2. **Extensibility**: Custom tags can be created for specific needs
3. **Searchability**: Complex queries using tag expressions
4. **Consistency**: Unified approach across all entity types
5. **Simplicity**: Single model for all entities with different characteristics

## Implementation Details

### Entity Model

All entities (including workspaces, issues, epics, etc.) now use a unified data model with tags:

```go
type Issue struct {
    ID             string
    Title          string
    Description    string
    Priority       string
    EstimatedEffort float64
    DueDate        time.Time
    CreatedAt      time.Time
    CreatedBy      string
    WorkspaceID    string
    ParentID       string
    ChildCount     int
    ChildCompleted int
    Progress       int
    Tags           []string
}
```

Helper methods are provided to work with tags:

```go
// Get the type of the issue
func (i *Issue) GetType() string {
    return i.GetTagWithPrefix("type:")
}

// Get the status of the issue
func (i *Issue) GetStatus() string {
    return i.GetTagWithPrefix("status:")
}

// Add a tag to the issue
func (i *Issue) AddTag(tag string) {
    // Implementation details
}

// Remove a tag from the issue
func (i *Issue) RemoveTag(tag string) {
    // Implementation details
}

// Replace a tag with a new one in the same namespace
func (i *Issue) ReplaceTagWithPrefix(prefix string, newTag string) {
    // Implementation details
}
```

### Storage

Tags are stored in a junction table:

```sql
CREATE TABLE issue_tags (
    issue_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (issue_id, tag),
    FOREIGN KEY (issue_id) REFERENCES issues(id) ON DELETE CASCADE
);
```

### API Support

The API supports tag operations:

- GET `/api/v1/tag/:entityType/:entityID` - Get all tags for an entity
- POST `/api/v1/tag/add` - Add a tag to an entity
- DELETE `/api/v1/tag/remove` - Remove a tag from an entity
- GET `/api/v1/tag/list/:entityType` - List entities with specific tags
- POST `/api/v1/tag/search` - Search entities using tag expressions

## Tag Expressions

Tag expressions allow for complex queries:

- `type:issue AND status:pending` - Find all pending issues
- `area:frontend OR area:ui` - Find all frontend or UI-related entities
- `type:story AND NOT status:completed` - Find incomplete stories
- `assignee:claude-2 AND priority:high` - Find high-priority tasks assigned to claude-2

## Best Practices

1. **Consistent Namespaces**: Use consistent namespace names across the system
2. **Namespace Conventions**: Use singular nouns for namespaces
3. **Value Conventions**: Use lowercase with underscores for multi-word values
4. **Minimal Tags**: Don't overload entities with unnecessary tags
5. **Comprehensive Tags**: Ensure entities have all required classification tags
6. **Type and Status**: Always include type and status tags for every entity

## Example Usage

### Creating a New Issue

```json
{
  "title": "Implement Login UI",
  "description": "Create the login screen according to designs",
  "priority": "medium",
  "tags": [
    "type:issue",
    "status:pending",
    "area:frontend",
    "component:auth",
    "assignee:claude-2"
  ]
}
```

### Updating Issue Status

```json
{
  "tags": ["status:in_progress"]
}
```

## Relationship with RBAC

Tag-based architecture integrates with Role-Based Access Control:

- Permission rules can reference tags: "Can view issues with tag `area:security`"
- Role assignments can be based on tags: "Assign security role to users working on `area:security` issues"
- Access control can use tag expressions for complex rules

## Migration Guide

When migrating from the previous type/status fields:
1. Ensure all entities have appropriate `type:X` and `status:Y` tags
2. Use helper methods to access type and status values
3. Update UI to display and modify tags instead of fixed fields

## Conclusion

The tag-based architecture provides a flexible, powerful way to classify and organize entities in the EntityDB system. By leveraging tags, the system becomes more adaptable to changing requirements while maintaining a consistent data model.
