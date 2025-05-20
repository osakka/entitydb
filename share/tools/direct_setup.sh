#!/bin/bash
#
# Direct database setup script using SQLite commands
# This bypasses the API entirely for reliability

# Database path
DB_PATH="/opt/entitydb/var/db/entitydb.db"
LOG_FILE="/opt/entitydb/var/log/direct_setup.log"

# Ensure log directory exists
mkdir -p /opt/entitydb/var/log

# Logging function
log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Remove existing database
log "Removing existing database..."
rm -f "$DB_PATH"

# Restart server to create fresh database
log "Restarting server to create fresh database..."
/opt/entitydb/bin/entitydbd.sh restart
sleep 5

# Direct database setup with SQLite commands
log "Setting up database directly with SQLite..."

NOW=$(date "+%Y-%m-%d %H:%M:%S")

# Create osakka user with admin role
sqlite3 "$DB_PATH" <<EOF
-- Create osakka user
INSERT OR REPLACE INTO users (id, username, password_hash, email, full_name, roles, created_at, agent_id, active, status)
VALUES ('usr_osakka', 'osakka', '\$2a\$10\$zLdwilHER7xoOYOH2iSUZ.ENwU7ijjzdgJyZAKxshaCBXGn3Kzz5.',
        'osakka@example.com', 'Omar Sakka', 'admin', '$NOW', NULL, 1, 'active');

-- Grant all permissions to osakka
INSERT OR IGNORE INTO user_permissions
SELECT 'usr_osakka', id FROM permissions;

-- Handle existing agents to avoid conflicts
DELETE FROM agents WHERE handle IN ('claude-1', 'claude-2', 'claude-3', 'osakka');

-- Create Claude agents
INSERT INTO agents (id, handle, name, display_name, type, specialization, personality_profile, status, created_at)
VALUES ('ag_claude1', 'claude-1', 'Claude 1', 'Claude 1', 'ai', 'code analysis, debugging, issue management', 'Helpful AI assistant', 'active', '$NOW');

INSERT INTO agents (id, handle, name, display_name, type, specialization, personality_profile, status, created_at)
VALUES ('ag_claude2', 'claude-2', 'Claude 2', 'Claude 2', 'ai', 'code analysis, debugging, issue management', 'Helpful AI assistant', 'active', '$NOW');

INSERT INTO agents (id, handle, name, display_name, type, specialization, personality_profile, status, created_at)
VALUES ('ag_claude3', 'claude-3', 'Claude 3', 'Claude 3', 'ai', 'code analysis, debugging, issue management', 'Helpful AI assistant', 'active', '$NOW');

-- Link osakka user with an agent
INSERT INTO agents (id, handle, name, display_name, type, status, created_at)
VALUES ('ag_osakka', 'osakka', 'Omar Agent', 'Omar Sakka', 'human', 'active', '$NOW');

UPDATE users SET agent_id = 'ag_osakka' WHERE id = 'usr_osakka';

-- Delete any existing pools to avoid conflicts
DELETE FROM agent_pools WHERE id = 'pool_claude_team';

-- Create agent pool for Claude team
INSERT INTO agent_pools (id, name, description, created_at, created_by, status)
VALUES ('pool_claude_team', 'Claude Team', 'Team of Claude agents for code analysis and debugging', '$NOW', 'usr_osakka', 'active');

-- Remove existing pool agent assignments to avoid conflicts
DELETE FROM pool_agents WHERE pool_id = 'pool_claude_team';
DELETE FROM pool_agents WHERE agent_id IN ('ag_claude1', 'ag_claude2', 'ag_claude3');

-- Add agents to pool
INSERT INTO pool_agents (pool_id, agent_id, added_at, added_by)
VALUES ('pool_claude_team', 'ag_claude1', '$NOW', 'usr_osakka');

INSERT INTO pool_agents (pool_id, agent_id, added_at, added_by)
VALUES ('pool_claude_team', 'ag_claude2', '$NOW', 'usr_osakka');

INSERT INTO pool_agents (pool_id, agent_id, added_at, added_by)
VALUES ('pool_claude_team', 'ag_claude3', '$NOW', 'usr_osakka');

-- Delete any existing project workspace and issues
DELETE FROM issues WHERE id = 'workspace_project';
DELETE FROM issues WHERE id IN ('issue_oauth', 'issue_performance', 'issue_ui');

-- Create project workspace
INSERT INTO issues (id, title, description, type, priority, created_at, created_by, workspace_id, parent_id, status)
VALUES ('workspace_project', 'Project Workspace', 'Main project workspace for development', 'workspace', 'high', '$NOW', 'ag_osakka', NULL, NULL, 'active');

-- Create issues
INSERT INTO issues (id, title, description, type, priority, created_at, created_by, workspace_id, parent_id, status)
VALUES ('issue_oauth', 'Implement OAuth integration', 'Add OAuth2 support for third-party authentication', 'issue', 'high', '$NOW', 'ag_osakka', 'workspace_project', NULL, 'pending');

INSERT INTO issues (id, title, description, type, priority, created_at, created_by, workspace_id, parent_id, status)
VALUES ('issue_performance', 'Performance Optimization', 'Optimize database queries and API response times', 'issue', 'medium', '$NOW', 'ag_osakka', 'workspace_project', NULL, 'pending');

INSERT INTO issues (id, title, description, type, priority, created_at, created_by, workspace_id, parent_id, status)
VALUES ('issue_ui', 'UI Enhancements', 'Improve user interface and experience', 'issue', 'low', '$NOW', 'ag_osakka', 'workspace_project', NULL, 'pending');

-- Delete existing issue assignments to avoid conflicts
DELETE FROM issue_assignments WHERE issue_id IN ('issue_oauth', 'issue_performance');
DELETE FROM issue_assignments WHERE agent_id IN ('ag_claude1', 'ag_claude3');

-- Assign issues to agents
INSERT INTO issue_assignments (issue_id, agent_id, assigned_at, assigned_by)
VALUES ('issue_oauth', 'ag_claude1', '$NOW', 'ag_osakka');

INSERT INTO issue_assignments (issue_id, agent_id, assigned_at, assigned_by)
VALUES ('issue_performance', 'ag_claude3', '$NOW', 'ag_osakka');

-- Update issue status to reflect assignments
UPDATE issues SET status = 'assigned' WHERE id IN ('issue_oauth', 'issue_performance');

-- Delete existing pool-workspace associations to avoid conflicts
DELETE FROM pool_workspaces WHERE pool_id = 'pool_claude_team';
DELETE FROM pool_workspaces WHERE workspace_id = 'workspace_project';

-- Associate Claude team pool with project workspace
INSERT INTO pool_workspaces (pool_id, workspace_id, added_at, added_by)
VALUES ('pool_claude_team', 'workspace_project', '$NOW', 'usr_osakka');
EOF

log "Direct database setup complete."

# Restart server again with the fully setup database
log "Restarting server with completed database..."
/opt/entitydb/bin/entitydbd.sh restart

exit 0