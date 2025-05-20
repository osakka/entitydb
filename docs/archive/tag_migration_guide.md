# Tag-Based Architecture Migration Guide

## Overview

This document outlines the strategy for migrating the EntityDB system from explicit Type and Status fields to a flexible tag-based architecture. The migration is designed to minimize disruption while enabling the enhanced capabilities of the tag system.

## Migration Strategy

### Phase 1: Parallel Implementation (Completed)

1. **Code Updates**
   - ✅ Added Tags field to Issue struct
   - ✅ Created tag helper methods (GetType, GetStatus, AddTag, RemoveTag)
   - ✅ Implemented tag repository and handlers
   - ✅ Created test endpoints for validation

2. **Testing Infrastructure**
   - ✅ Created fully mocked endpoints that bypass authentication
   - ✅ Added test version of client script (entitydbc.sh.tag_based)
   - ✅ Verified tag operations with various entity types

### Phase 2: Schema Migration (In Progress)

1. **Create Migration Scripts**
   - Create issue_tags table
   - Migrate existing Type and Status values to tags
   - Update foreign key constraints

2. **Data Migration**
   ```sql
   -- Create the issue_tags table
   CREATE TABLE issue_tags (
       issue_id TEXT NOT NULL,
       tag TEXT NOT NULL,
       PRIMARY KEY (issue_id, tag),
       FOREIGN KEY (issue_id) REFERENCES issues(id) ON DELETE CASCADE
   );

   -- Migrate Type values to tags
   INSERT INTO issue_tags (issue_id, tag)
   SELECT id, 'type:' || LOWER(type) FROM issues;

   -- Migrate Status values to tags
   INSERT INTO issue_tags (issue_id, tag)
   SELECT id, 'status:' || LOWER(status) FROM issues;

   -- Create a temporary table for the new structure
   CREATE TABLE issues_new (
       id TEXT PRIMARY KEY,
       title TEXT NOT NULL,
       description TEXT,
       priority TEXT NOT NULL CHECK (priority IN ('high', 'medium', 'low')),
       estimated_effort REAL DEFAULT 0,
       due_date TIMESTAMP,
       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
       created_by TEXT NOT NULL,
       workspace_id TEXT,
       parent_id TEXT,
       child_count INTEGER DEFAULT 0,
       child_completed INTEGER DEFAULT 0,
       progress INTEGER DEFAULT 0 CHECK (progress BETWEEN 0 AND 100),
       FOREIGN KEY (workspace_id) REFERENCES issues(id) ON DELETE CASCADE,
       FOREIGN KEY (parent_id) REFERENCES issues(id) ON DELETE CASCADE
   );

   -- Copy data to the new structure without Type and Status
   INSERT INTO issues_new (
       id, title, description, priority, estimated_effort, 
       due_date, created_at, created_by, workspace_id, 
       parent_id, child_count, child_completed, progress
   )
   SELECT 
       id, title, description, priority, estimated_effort, 
       due_date, created_at, created_by, workspace_id, 
       parent_id, child_count, child_completed, progress
   FROM issues;

   -- Replace the old table with the new one
   DROP TABLE issues;
   ALTER TABLE issues_new RENAME TO issues;
   ```

### Phase 3: API Transition (Planned)

1. **Update API Endpoints**
   - Modify endpoints to use tag operations exclusively
   - Update CRUD operations to handle tags properly
   - Implement backwards compatibility for older clients

2. **Authentication and Permissions**
   - Update RBAC to work with tag-based permissions
   - Modify authentication middleware to handle tag-based access control

3. **Tag-Based Queries**
   - Implement tag expression search capability
   - Update filtering to use tag-based queries

### Phase 4: Client Update (Planned)

1. **Client Script Updates**
   - Replace entitydbc.sh with tag-based version
   - Update all commands to use tag operations
   - Add new tag management commands

2. **User Training**
   - Provide documentation on tag-based commands
   - Explain benefits and use cases for tags
   - Offer examples for common workflows

## Migration Checklist

- [x] Add Tags field to Issue struct
- [x] Create tag helper methods
- [x] Implement tag repositories
- [x] Create test endpoints
- [x] Test tag operations
- [x] Create tag-based client script
- [ ] Create migration scripts
- [ ] Migrate existing data
- [ ] Update all API endpoints
- [ ] Update RBAC for tag-based permissions
- [ ] Implement tag expression search
- [ ] Replace production client script
- [ ] Update documentation

## Data Migration Examples

### Before Migration:

```json
{
  "id": "issue_12345",
  "title": "Add login form",
  "description": "Create a new login form for the application",
  "type": "Issue",
  "status": "Pending",
  "priority": "high",
  "created_by": "agent_claude"
}
```

### After Migration:

```json
{
  "id": "issue_12345",
  "title": "Add login form",
  "description": "Create a new login form for the application",
  "priority": "high",
  "created_by": "agent_claude",
  "tags": [
    "type:issue",
    "status:pending",
    "area:frontend",
    "component:auth"
  ]
}
```

## API Migration Examples

### Before Migration:

```bash
# Create issue
curl -X POST -H "Content-Type: application/json" -d '{
  "title": "New Feature",
  "description": "Add new feature",
  "type": "Story",
  "status": "Pending",
  "priority": "high"
}' http://localhost:8085/api/v1/issues/create

# Update status
curl -X PUT -H "Content-Type: application/json" -d '{
  "status": "In Progress"
}' http://localhost:8085/api/v1/issues/update
```

### After Migration:

```bash
# Create issue
curl -X POST -H "Content-Type: application/json" -d '{
  "title": "New Feature",
  "description": "Add new feature",
  "priority": "high",
  "tags": ["type:story", "status:pending"]
}' http://localhost:8085/api/v1/issues/create

# Update status (add tag)
curl -X POST -H "Content-Type: application/json" -d '{
  "entity_id": "issue_12345",
  "tag": "status:in_progress"
}' http://localhost:8085/api/v1/tag/add

# Update status (remove old tag)
curl -X DELETE -H "Content-Type: application/json" -d '{
  "entity_id": "issue_12345",
  "tag": "status:pending"
}' http://localhost:8085/api/v1/tag/remove
```

## Client Command Migration

### Before Migration:

```bash
# Create issue
./bin/entitydbc.sh issue create --title="New Feature" --description="Add new feature" --type=story --priority=high

# Start issue (update status)
./bin/entitydbc.sh issue start issue_12345
```

### After Migration:

```bash
# Create issue
./bin/entitydbc.sh issue create --title="New Feature" --description="Add new feature" --tags="type:story,status:pending" --priority=high

# Start issue (update status tag)
./bin/entitydbc.sh issue start issue_12345
```

## Rollback Plan

In case of migration issues, we have prepared a rollback strategy:

1. **Revert Code Changes**
   - Restore Type and Status fields to Issue struct
   - Keep Tags field for backward compatibility
   - Update helper methods to work with both approaches

2. **Database Rollback**
   - Restore Type and Status columns in issues table
   - Extract values from tags and populate the columns
   - Keep issue_tags table for future migration attempts

3. **API Rollback**
   - Revert API endpoints to use Type and Status fields
   - Keep tag endpoints for future migration

## Timeline

1. **Phase 1: Parallel Implementation** - Completed
2. **Phase 2: Schema Migration** - In Progress (1 week)
3. **Phase 3: API Transition** - Planned (2 weeks)
4. **Phase 4: Client Update** - Planned (1 week)

Total migration time: 4 weeks

## Conclusion

This migration will enhance the flexibility and extensibility of the EntityDB system while maintaining compatibility with existing functionality. The tag-based architecture enables more dynamic classification and organization of entities, providing a foundation for future enhancements.