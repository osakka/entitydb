# EntityDB Schema Fixes

## Overview

This repository contains schema fixes for the EntityDB (EntityDB) platform to address authentication and agent-related issues. The fixes focus on database schema structure, adding missing columns and tables, and ensuring proper relationships between entities.

## Quick Start

To apply the schema fixes, run the following command:

```bash
/opt/entitydb/bin/fix_database.sh
```

This script will:
1. Back up the current database
2. Apply all necessary schema updates
3. Verify the schema changes were applied correctly
4. Restart the server with the updated schema

## Files Included

- `/opt/entitydb/bin/fix_database.sh` - Main script to apply schema fixes
- `/opt/entitydb/src/models/sqlite/schema_update_consolidated.sql` - Consolidated SQL schema updates
- `/opt/entitydb/bin/check_agent.sh` - Workaround script to check agent information
- `/opt/entitydb/bin/check_session.sh` - Workaround script to manage sessions
- `/opt/entitydb/docs/schema_fix_report.md` - Detailed report of schema fixes
- `/opt/entitydb/docs/schema_fix_summary.md` - Executive summary of changes
- `/opt/entitydb/docs/workaround_scripts.md` - Documentation for workaround scripts

## What Was Fixed

### Authentication Issues

The authentication system was fixed by addressing the following issues:

1. Added missing `token_type` column to the `auth_tokens` table
2. Created proper indexes for token lookups
3. Ensured foreign key constraints to maintain data integrity
4. Created an admin2 user (password: admin) for easier access

### Agent Issues

The agent system was fixed by addressing the following issues:

1. Renamed `last_active_at` column to `last_active` to match code expectations
2. Added missing columns:
   - `worker_pool_id` (TEXT)
   - `expertise` (TEXT)
   - `capability_score` (INTEGER)
3. Created missing agent-related tables:
   - `agent_capabilities`
   - `agent_performance`

## Remaining Issues

While the schema fixes resolved the core authentication and agent creation issues, the following API endpoints still have problems:

1. Agent listing endpoint (`/api/v1/agents/list`)
2. Agent profile endpoint (`/api/v1/agents/get`)
3. Session creation endpoint (`/api/v1/sessions/create`)

To work around these issues, use the provided scripts:
- `/opt/entitydb/bin/check_agent.sh` for agent operations
- `/opt/entitydb/bin/check_session.sh` for session operations

## Usage Examples

### Using the Fix Script

```bash
# Apply schema fixes and restart server
/opt/entitydb/bin/fix_database.sh
```

### Using the Agent Script

```bash
# List all agents
/opt/entitydb/bin/check_agent.sh list

# View details for a specific agent
/opt/entitydb/bin/check_agent.sh view claude-2
```

### Using the Session Script

```bash
# List all sessions
/opt/entitydb/bin/check_session.sh list

# Create a new session
/opt/entitydb/bin/check_session.sh create claude-2 workspace_entitydb "Test Session" "Description"

# View session details
/opt/entitydb/bin/check_session.sh view sess_1234567890_1234
```

## Future Work

The following areas need further development:

1. Fix the agent listing API endpoint
2. Fix the agent profile API endpoint
3. Fix the session creation API endpoint
4. Add more comprehensive tests for API endpoints
5. Implement automated schema validation on application startup

## Contributors

- Claude AI Assistant