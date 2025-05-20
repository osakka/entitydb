# Entity Migration Guide

This guide provides instructions for migrating from the traditional issue-based model to the new entity-based architecture in EntityDB.

## Overview

The migration from issues to entities involves several steps:

1. Enable the entity system
2. Configure dual-write mode
3. Run data migration scripts
4. Test with the entity-issue adapter
5. Switch to direct entity API usage

This guide will walk you through each step of the process.

## Prerequisites

- EntityDB server version with entity support (0.2.0+)
- Admin access to the EntityDB system
- Database backup (critical before migration)

## Step 1: Backup Your Data

Before beginning any migration, create a complete backup of your database:

```bash
# Stop the server
./bin/entitydbd.sh stop

# Backup the database
cp /opt/entitydb/var/db/entitydb.db /opt/entitydb/var/db/entitydb.db.pre_entity_migration.bak

# Restart the server
./bin/entitydbd.sh start
```

## Step 2: Enable the Entity System

Set the required configuration values:

```bash
# Enable entity API
./bin/entitydbc.sh config update --key=entity_api_enabled --value=true

# Disable entity migration until ready
./bin/entitydbc.sh config update --key=entity_migration_enabled --value=false

# Set initial migration status
./bin/entitydbc.sh config update --key=entity_migration_status --value=pending
```

## Step 3: Configure Dual-Write Mode

Dual-write mode ensures that all operations on the issue model also update the corresponding entities:

```bash
# Enable dual-write mode for entity-issue adapter
./bin/entitydbc.sh config update --key=entity_issue_dual_write --value=true

# Enable using entity handler
./bin/entitydbc.sh config update --key=entity_issue_handler_enabled --value=true
```

## Step 4: Restart the Server

Restart the server to apply the new configuration:

```bash
./bin/entitydbd.sh restart
```

## Step 5: Verify Entity Tables

Verify that the entity tables were created correctly:

```bash
sqlite3 /opt/entitydb/var/db/entitydb.db "SELECT name FROM sqlite_master WHERE type='table' AND (name='entities' OR name='entity_relationships')"
```

You should see both `entities` and `entity_relationships` tables listed.

## Step 6: Run the Migration Script

Now, run the migration script to convert existing issues to entities:

```bash
# Update migration status
./bin/entitydbc.sh config update --key=entity_migration_status --value=in_progress

# Run the migration tool
cd /opt/entitydb/src/tools
go run migrate_to_entities.go

# Check migration results
sqlite3 /opt/entitydb/var/db/entitydb.db "SELECT COUNT(*) FROM entities; SELECT COUNT(*) FROM issues;"
```

The counts should match if all issues were successfully migrated.

## Step 7: Test the Migration

Test that the system works correctly with both APIs:

1. Try creating an issue using the traditional API:

```bash
./bin/entitydbc.sh issue create --title="Test dual-write" --description="Testing dual-write mode"
```

2. Verify it exists in both tables:

```bash
./bin/entitydbc.sh issue list | grep "Test dual-write"
sqlite3 /opt/entitydb/var/db/entitydb.db "SELECT id FROM entities WHERE tags LIKE '%title=Test dual-write%'"
```

## Step 8: Update Applications

Update any client applications to start using the entity API:

- Replace `/api/v1/issues/...` with `/api/v1/entities/...`
- Update data models to use tags and content instead of direct fields

## Step 9: Finalize Migration

Once everything is working correctly and all testing is complete:

```bash
# Mark migration as complete
./bin/entitydbc.sh config update --key=entity_migration_status --value=completed
```

## Rollback Procedure

If issues occur during migration, you can roll back using your backup:

```bash
# Stop the server
./bin/entitydbd.sh stop

# Restore the database
cp /opt/entitydb/var/db/entitydb.db.pre_entity_migration.bak /opt/entitydb/var/db/entitydb.db

# Restart the server
./bin/entitydbd.sh start

# Disable entity features
./bin/entitydbc.sh config update --key=entity_api_enabled --value=false
./bin/entitydbc.sh config update --key=entity_issue_handler_enabled --value=false
./bin/entitydbc.sh config update --key=entity_migration_status --value=pending
```

## Data Mapping Reference

| Issue Field    | Entity Representation                   |
|----------------|----------------------------------------|
| id             | Entity ID                               |
| title          | Content item with type="title"          |
| description    | Content item with type="description"    |
| type           | Tag "type=value"                        |
| priority       | Tag "priority=value"                    |
| status         | Tag "status=value"                      |
| created_by     | Tag "created_by=value"                  |
| created_at     | Tag "created_at=timestamp"              |
| workspace_id   | Tag "workspace=value"                   |
| parent_id      | Relationship with type="parent_of"      |
| dependencies   | Relationships with type="depends_on"    |
| assignments    | Relationships with type="assigned_to"   |

## Support

If you encounter any issues during migration, contact EntityDB support or file an issue in the repository.