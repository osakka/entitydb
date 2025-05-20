# Deprecated Scripts

This directory contains deprecated scripts and files that were used for temporary fixes and workarounds. These files are kept for historical reference only and should not be used in production.

## Deprecated Fix Scripts

1. **Schema Fix Scripts**:
   - `scripts/fix_agent_fk.sql` - Fixed agent table foreign keys (integrated into migrations)
   - `scripts/fix_issue_fk.sql` - Fixed issue table foreign keys (integrated into migrations)
   - `scripts/workspace_fix.sh` - Workspace ID prefix handling (integrated into client)

2. **Workaround Scripts**:
   - `scripts/check_agent.sh` - Direct agent information from database (use client commands instead)
   - `scripts/check_session.sh` - Direct session management (use client commands instead)
   - `scripts/rbac_wrapper.sh` - RBAC test wrapper (integrated into main API)
   - `scripts/test_issue_create` - Test script for issue creation (use official API tests)

## Integrated Functionality

All functionality from these scripts has been properly integrated into the main system:

1. **Schema Fixes**:
   - Consolidated schema migrations in `/opt/entitydb/src/models/sqlite/schema_update_consolidated.sql`
   - Automated schema updates via the migration system

2. **Client Commands**:
   - For agent management: `./bin/entitydbc.sh agent list`, `./bin/entitydbc.sh agent view <id>`
   - For session management: `./bin/entitydbc.sh session list`, `./bin/entitydbc.sh session view <id>`

3. **Testing**:
   - Official test suite in `/opt/entitydb/share/tests/api/`
   - RBAC tests in `/opt/entitydb/share/tests/api/rbac/`

## Documentation

For current documentation on these functionalities, see:

1. `/opt/entitydb/docs/rbac_implementation.md` - RBAC implementation details
2. `/opt/entitydb/docs/schema_fixes_summary.md` - Summary of all schema fixes
3. `/opt/entitydb/README.md` - Main project documentation