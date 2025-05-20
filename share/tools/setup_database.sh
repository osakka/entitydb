#!/bin/bash
#
# Script to set up the database with all required elements
# This script uses direct SQLite commands which have been proven to work
#

# Database path
DB_PATH="/opt/entitydb/var/db/entitydb.db"
LOG_FILE="/opt/entitydb/var/log/db_setup.log"

# Ensure log directory exists
mkdir -p /opt/entitydb/var/log

# Logging function
log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Function to run SQL commands
run_sql() {
  sqlite3 "$DB_PATH" "$1"
  if [ $? -eq 0 ]; then
    log "SQL command executed successfully"
  else
    log "Error executing SQL command: $1"
  fi
}

# Start with a clean database
if [ "$1" == "--clean" ]; then
  log "Removing existing database..."
  rm -f "$DB_PATH"
  
  log "Restarting server to create fresh database..."
  /opt/entitydb/bin/entitydbd.sh restart
  sleep 5
fi

log "Setting up database with required elements..."

# Create osakka user with admin role
log "Creating osakka user..."
run_sql "INSERT INTO users (id, username, password_hash, email, full_name, roles, created_at, active, status) VALUES ('usr_osakka', 'osakka', '\$2a\$10\$zLdwilHER7xoOYOH2iSUZ.ENwU7ijjzdgJyZAKxshaCBXGn3Kzz5.', 'osakka@example.com', 'Omar Sakka', 'admin', datetime('now'), 1, 'active');"

# Grant all permissions to osakka
run_sql "INSERT OR IGNORE INTO user_permissions SELECT 'usr_osakka', id FROM permissions;"

# Delete existing agents to avoid conflicts
log "Creating Claude agents..."
run_sql "DELETE FROM agents WHERE handle IN ('claude-1', 'claude-2', 'claude-3');"

# Create Claude agents
run_sql "INSERT INTO agents (id, handle, name, display_name, type, status, created_at) VALUES ('ag_claude1', 'claude-1', 'Claude 1', 'Claude 1', 'ai', 'active', datetime('now'));"
run_sql "INSERT INTO agents (id, handle, name, display_name, type, status, created_at) VALUES ('ag_claude2', 'claude-2', 'Claude 2', 'Claude 2', 'ai', 'active', datetime('now'));"
run_sql "INSERT INTO agents (id, handle, name, display_name, type, status, created_at) VALUES ('ag_claude3', 'claude-3', 'Claude 3', 'Claude 3', 'ai', 'active', datetime('now'));"

# Create agent pool for Claude team
log "Creating Claude team pool..."
run_sql "DELETE FROM agent_pools WHERE id = 'pool_claude_team';"
run_sql "INSERT INTO agent_pools (id, name, description, created_at, created_by, status) VALUES ('pool_claude_team', 'Claude Team', 'Team of Claude agents for code analysis and debugging', datetime('now'), 'usr_osakka', 'active');"

# Clean up existing pool assignments to avoid conflicts
run_sql "DELETE FROM pool_agents WHERE pool_id = 'pool_claude_team';"
run_sql "DELETE FROM pool_agents WHERE agent_id IN ('ag_claude1', 'ag_claude2', 'ag_claude3');"

# Add agents to pool
log "Adding agents to Claude team pool..."
run_sql "INSERT INTO pool_agents (pool_id, agent_id, added_at, added_by) VALUES ('pool_claude_team', 'ag_claude1', datetime('now'), 'usr_osakka');"
run_sql "INSERT INTO pool_agents (pool_id, agent_id, added_at, added_by) VALUES ('pool_claude_team', 'ag_claude2', datetime('now'), 'usr_osakka');"
run_sql "INSERT INTO pool_agents (pool_id, agent_id, added_at, added_by) VALUES ('pool_claude_team', 'ag_claude3', datetime('now'), 'usr_osakka');"

# Create project workspace
log "Creating project workspace..."
run_sql "DELETE FROM issues WHERE id = 'workspace_project';"
run_sql "INSERT INTO issues (id, title, description, type, priority, created_at, created_by, workspace_id, parent_id, status) VALUES ('workspace_project', 'Project Workspace', 'Main project workspace for development', 'workspace', 'high', datetime('now'), 'usr_osakka', NULL, NULL, 'active');"

# Create issues
log "Creating issues in project workspace..."
run_sql "DELETE FROM issues WHERE id IN ('issue_oauth', 'issue_performance', 'issue_ui');"
run_sql "INSERT INTO issues (id, title, description, type, priority, created_at, created_by, workspace_id, parent_id, status) VALUES ('issue_oauth', 'Implement OAuth integration', 'Add OAuth2 support for third-party authentication', 'issue', 'high', datetime('now'), 'usr_osakka', 'workspace_project', NULL, 'pending');"
run_sql "INSERT INTO issues (id, title, description, type, priority, created_at, created_by, workspace_id, parent_id, status) VALUES ('issue_performance', 'Performance Optimization', 'Optimize database queries and API response times', 'issue', 'medium', datetime('now'), 'usr_osakka', 'workspace_project', NULL, 'pending');"
run_sql "INSERT INTO issues (id, title, description, type, priority, created_at, created_by, workspace_id, parent_id, status) VALUES ('issue_ui', 'UI Enhancements', 'Improve user interface and experience', 'issue', 'low', datetime('now'), 'usr_osakka', 'workspace_project', NULL, 'pending');"

# Clean up existing issue assignments to avoid conflicts
log "Assigning issues to agents..."
run_sql "DELETE FROM issue_assignments WHERE issue_id IN ('issue_oauth', 'issue_performance');"
run_sql "DELETE FROM issue_assignments WHERE agent_id IN ('ag_claude1', 'ag_claude3');"

# Assign issues to agents
run_sql "INSERT INTO issue_assignments (issue_id, agent_id, assigned_at, assigned_by) VALUES ('issue_oauth', 'ag_claude1', datetime('now'), 'usr_osakka');"
run_sql "INSERT INTO issue_assignments (issue_id, agent_id, assigned_at, assigned_by) VALUES ('issue_performance', 'ag_claude3', datetime('now'), 'usr_osakka');"

# Update issue status to reflect assignments
run_sql "UPDATE issues SET status = 'assigned' WHERE id IN ('issue_oauth', 'issue_performance');"

# Associate pool with workspace
log "Associating Claude team pool with project workspace..."
run_sql "DELETE FROM pool_workspaces WHERE pool_id = 'pool_claude_team' AND workspace_id = 'workspace_project';"
run_sql "INSERT INTO pool_workspaces (pool_id, workspace_id, added_at, added_by) VALUES ('pool_claude_team', 'workspace_project', datetime('now'), 'usr_osakka');"

log "Database setup completed successfully."

# Restart server to ensure everything is applied
if [ "$1" == "--restart" ] || [ "$2" == "--restart" ]; then
  log "Restarting server..."
  /opt/entitydb/bin/entitydbd.sh restart
fi

exit 0