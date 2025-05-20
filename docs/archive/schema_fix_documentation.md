# Schema Fixes for Authentication and Agent Creation

## Issue Overview

The system had several schema mismatches that were causing the following issues:

1. Authentication tokens were failing due to a missing `token_type` column in the `auth_tokens` table
2. Agent creation was failing due to several issues:
   - Column name mismatch: `last_active_at` vs `last_active` 
   - Missing columns: `worker_pool_id`, `expertise`, and `capability_score`
   - Missing agent-related tables: `agent_capabilities` and `agent_performance`

## Solution Implemented

### Authentication Token Fixes

Created a schema update file that drops and recreates the `auth_tokens` table with all needed columns:

```sql
-- Drop and recreate the auth_tokens table to match the code implementation
DROP TABLE IF EXISTS auth_tokens;

CREATE TABLE auth_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    agent_id TEXT,
    token_type TEXT NOT NULL DEFAULT 'access',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_revoked INTEGER DEFAULT 0,
    last_used_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_auth_tokens_user_id ON auth_tokens(user_id);
CREATE INDEX idx_auth_tokens_token_type ON auth_tokens(token_type);
```

### Agent Table Fixes

Added the following schema changes to fix the agent implementation:

```sql
-- Fix column names to match what the code expects
ALTER TABLE agents RENAME COLUMN last_active_at TO last_active;

-- Add missing agent columns
ALTER TABLE agents ADD COLUMN worker_pool_id TEXT;
ALTER TABLE agents ADD COLUMN expertise TEXT DEFAULT '';
ALTER TABLE agents ADD COLUMN capability_score INTEGER DEFAULT 0;

-- Create agent capabilities table if not exists
CREATE TABLE IF NOT EXISTS agent_capabilities (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    capability_type TEXT NOT NULL,
    capability_name TEXT NOT NULL,
    proficiency_level TEXT NOT NULL,
    last_assessment TIMESTAMP NOT NULL,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- Create agent performance table if not exists
CREATE TABLE IF NOT EXISTS agent_performance (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    metric_type TEXT NOT NULL,
    metric_value REAL NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    notes TEXT,
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_agent_capabilities_agent_id ON agent_capabilities(agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_performance_agent_id ON agent_performance(agent_id);
```

## RBAC Permissions

Added necessary permissions to users for agent creation and viewing:

1. Added `agent.create` permission
2. Added `agent.view` permission
3. Added admin role to ensure proper authorization

## Implementation Process

1. First identified schema issues by examining the source code and comparing with database schema
2. Created schema update files in the `/opt/entitydb/src/models/sqlite/` directory
3. Restarted the server to apply schema changes
4. Verified token creation and authentication
5. Added necessary permissions to users
6. Successfully created and viewed agents

## Working Features

The following features are now working correctly:

1. User registration
2. User login with JWT tokens
3. Token refresh and validation
4. Agent creation
5. Agent retrieval

## Future Recommendations

1. Ensure schema migration scripts are created whenever the database schema is changed
2. Add validation in the code to gracefully handle missing columns
3. Create a more robust migration system that can handle incremental changes
4. Implement database schema version tracking to prevent database recreation
5. Add automated tests for authentication and agent operations

## Testing

You can test these features using the following commands:

### User Registration
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"username":"admin5","password":"password","display_name":"Admin User"}' \
  http://localhost:8085/api/v1/auth/register
```

### User Login
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"username":"admin5","password":"password"}' \
  http://localhost:8085/api/v1/auth/login
```

### Create Agent
```bash
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{"handle":"test-agent", "display_name":"Test Agent", "specialization":"Testing", "personality_profile":"Helpful tester"}' \
  http://localhost:8085/api/v1/agents/create
```

### View Agent
```bash
curl -X GET -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  http://localhost:8085/api/v1/agents/get?agent_id=AGENT_ID_HERE
```

### List Agents
```bash
curl -X GET -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  http://localhost:8085/api/v1/agents/list
```