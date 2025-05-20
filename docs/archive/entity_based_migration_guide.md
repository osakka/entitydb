# Entity-Based Architecture Migration Guide

This guide provides detailed instructions for migrating from the traditional issue-based architecture to the new entity-based architecture in the EntityDB platform.

## Overview

The EntityDB platform is transitioning from a structured issue-based model to a more flexible entity-based architecture. This migration provides several benefits:

1. **More Flexible Data Model**: Entities use tags and typed content instead of fixed schemas
2. **Improved Extensibility**: New features can be added without schema changes
3. **Better Performance**: Tag-based filtering is more efficient for complex queries
4. **Simplified Codebase**: One model can represent all domain objects

## Migration Process

The migration is designed to be gradual and non-disruptive. The main steps are:

1. **Deploy Entity-Ready Code**: Update to the latest version with entity support
2. **Enable Dual-Write Mode**: All operations write to both repositories
3. **Test Compatibility**: Verify that both APIs function correctly
4. **Migrate Existing Data**: Convert existing issues to entities
5. **Switch to Entity-Based APIs**: Change API handlers to use entity-based storage
6. **Monitor and Validate**: Ensure everything works as expected

## Prerequisites

Before starting the migration, ensure:

1. Your system is updated to the latest version
2. You have a backup of your database
3. You have sufficient disk space (2-3x your current database size)
4. You have administrative access to the system
5. All critical services are in a stable state

## Step 1: Deploy Entity-Ready Code

The entity-based architecture requires specific code components:

1. Entity model and repository
2. Entity-issue adapter for bidirectional conversion
3. Entity-issue repository wrapper
4. Entity-issue handler with dual-write support
5. Entity configuration endpoints

These are all included in the latest version of the EntityDB platform.

## Step 2: Enable Dual-Write Mode

Dual-write mode ensures that all operations are performed on both the issue and entity repositories, maintaining data consistency during migration.

To enable dual-write mode with the migration script:

```bash
cd /opt/entitydb/share/tools
./migrate_to_entity.sh --dual-write
```

This script:
1. Creates a database backup
2. Sets up entity tables
3. Enables dual-write mode
4. Leaves the API using the traditional issue repository

You can also enable dual-write manually through the configuration API:

```bash
curl -X PUT "http://localhost:8085/api/v1/entity/config" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer admin_token" \
  -d '{"dual_write_enabled": true}'
```

## Step 3: Test Compatibility

After enabling dual-write mode, test the system thoroughly to ensure both repositories are being updated correctly:

1. Create test issues
2. Update their status
3. Assign/unassign them
4. Manage dependencies
5. Make other changes to verify dual-write is working

You can verify entity creation by querying the database directly:

```bash
sqlite3 /opt/entitydb/var/db/entitydb.db "SELECT COUNT(*) FROM entities;"
```

## Step 4: Migrate Existing Data

Once you're confident in the dual-write mode, you can migrate existing data to the entity format:

```bash
cd /opt/entitydb/share/tools
./migrate_to_entity.sh
```

This script:
1. Creates a database backup
2. Migrates existing issues to entities
3. Enables entity-based API handlers
4. Runs tests to verify migration

The migration preserves all issue data, including:
- Basic metadata (type, status, priority)
- Relationships (workspace, parent)
- Tags and content
- Assignment information
- Progress values

## Step 5: Switch to Entity-Based APIs

After the migration script completes, the system will be using entity-based API handlers. You can also toggle this manually:

```bash
curl -X PUT "http://localhost:8085/api/v1/entity/config" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer admin_token" \
  -d '{"entity_handler_enabled": true}'
```

The server will need to be restarted for this change to take effect:

```bash
/opt/entitydb/bin/entitydbd.sh restart
```

## Step 6: Monitor and Validate

After the migration, monitor the system closely:

1. Check server logs for any errors
2. Verify API functionality for all operations
3. Compare entity counts with issue counts
4. Test complex operations (filtering, dependencies)

If issues arise, you can temporarily revert to the issue-based API:

```bash
curl -X PUT "http://localhost:8085/api/v1/entity/config" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer admin_token" \
  -d '{"entity_handler_enabled": false}'
```

## Final Steps

After a successful migration and sufficient validation:

1. Consider disabling dual-write mode to improve performance
2. Update client applications to use entity-specific features
3. Clean up any migration artifacts

## Entity Model Overview

The entity model provides a flexible foundation for all domain objects:

**Entity**:
- `id`: Unique identifier
- `tags`: Metadata in tag format (`key:value`)
- `content`: Typed content items with timestamps

**Tag Formats**:
- Simple: `key:value` (e.g., `type:issue`, `status:pending`)
- Timestamped: `YYYY-MM-DDTHH:MM:SS.nanos.key=value`

**Content Items**:
```json
{
  "timestamp": "2025-05-10T12:34:56Z",
  "type": "title",
  "value": "Example Title"
}
```

## Troubleshooting

If you encounter issues during migration:

1. **Database Issues**: Restore from the automatic backup
2. **API Failures**: Check logs for specific errors, revert to issue-based API
3. **Data Inconsistencies**: Run the migration script again with `--force`
4. **Performance Problems**: Tune the database, consider adding indexes

## Support

For assistance with the migration process, contact:

- Technical Support: support@entitydb.example.com
- Documentation: docs.entitydb.example.com/entity-migration
- EntityDB Community: community.entitydb.example.com