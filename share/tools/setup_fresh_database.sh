#!/bin/bash
#
# Script to set up a fresh database with all required elements
# This script fixes issues with standard API endpoints and provides
# a reliable way to initialize the database
#

# Database path
DB_PATH="/opt/entitydb/var/db/entitydb.db"
LOG_FILE="/opt/entitydb/var/log/setup.log"
SERVER_URL="http://localhost:8085"

# Ensure log directory exists
mkdir -p /opt/entitydb/var/log

# Logging function
log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Function to check if server is running
check_server() {
  if curl -s "$SERVER_URL/api/v1/status" > /dev/null; then
    return 0  # Server is running
  else
    return 1  # Server is not running
  fi
}

# Start with a clean database
log "Removing existing database..."
rm -f "$DB_PATH"

# Start the server to create a fresh database
log "Starting server to initialize database..."
if ! check_server; then
  /opt/entitydb/bin/entitydbd.sh start
  # Wait for server to initialize
  sleep 5
else
  log "Server already running"
fi

# Register osakka user
log "Creating user osakka..."
RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"osakka","password":"mypassword","email":"osakka@example.com","full_name":"Omar Sakka"}')

log "User creation response: $RESPONSE"

# Extract token for further operations
TOKEN=$(echo "$RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

# Make osakka an admin
log "Making osakka an admin..."
ADMIN_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/admin/assign" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"username":"osakka"}')

log "Admin assignment response: $ADMIN_RESPONSE"

# Create Claude agents
log "Creating Claude agents..."
for AGENT_NUM in 1 2 3; do
  AGENT_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/agent/register" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"handle\":\"claude-$AGENT_NUM\",\"name\":\"Claude $AGENT_NUM\",\"specialization\":\"code analysis, debugging, issue management\",\"personality_profile\":\"Helpful AI assistant\"}")
  
  log "Agent claude-$AGENT_NUM creation response: $AGENT_RESPONSE"
done

# Create Claude team pool
log "Creating Claude team pool..."
POOL_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/direct/pool/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id":"pool_claude_team","name":"Claude Team","description":"Team of Claude agents for code analysis and debugging","created_by":"usr_osakka"}')

log "Pool creation response: $POOL_RESPONSE"

# Get agent IDs
curl -s -X GET "$SERVER_URL/api/v1/agents/list" \
  -H "Authorization: Bearer $TOKEN" > /tmp/agents.json

CLAUDE1_ID=$(grep -o '"id":"[^"]*","handle":"claude-1"' /tmp/agents.json | cut -d'"' -f4)
CLAUDE2_ID=$(grep -o '"id":"[^"]*","handle":"claude-2"' /tmp/agents.json | cut -d'"' -f4)
CLAUDE3_ID=$(grep -o '"id":"[^"]*","handle":"claude-3"' /tmp/agents.json | cut -d'"' -f4)

log "Agent IDs: claude-1=$CLAUDE1_ID, claude-2=$CLAUDE2_ID, claude-3=$CLAUDE3_ID"

# Add agents to pool
log "Adding agents to pool..."
for AGENT_ID in "$CLAUDE1_ID" "$CLAUDE2_ID" "$CLAUDE3_ID"; do
  if [ -n "$AGENT_ID" ]; then
    ADD_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/direct/pool/add_agent" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "{\"pool_id\":\"pool_claude_team\",\"agent_id\":\"$AGENT_ID\",\"added_by\":\"usr_osakka\"}")
    
    log "Add agent $AGENT_ID to pool response: $ADD_RESPONSE"
  else
    log "Warning: Skipping empty agent ID"
  fi
done

# Create project workspace
log "Creating project workspace..."
WORKSPACE_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/direct/workspace/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Project Workspace","description":"Main project workspace for development","priority":"high","creator_id":"usr_osakka"}')

log "Workspace creation response: $WORKSPACE_RESPONSE"

# Create issues
log "Creating issues..."
ISSUE1_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/direct/issue/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Implement OAuth integration","description":"Add OAuth2 support for third-party authentication","priority":"high","type":"issue","workspace_id":"workspace_project","creator_id":"usr_osakka"}')

log "Issue 1 creation response: $ISSUE1_RESPONSE"

ISSUE2_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/direct/issue/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"Performance Optimization","description":"Optimize database queries and API response times","priority":"medium","type":"issue","workspace_id":"workspace_project","creator_id":"usr_osakka"}')

log "Issue 2 creation response: $ISSUE2_RESPONSE"

ISSUE3_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/direct/issue/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"UI Enhancements","description":"Improve user interface and experience","priority":"low","type":"issue","workspace_id":"workspace_project","creator_id":"usr_osakka"}')

log "Issue 3 creation response: $ISSUE3_RESPONSE"

# Assign issues to agents
if [ -n "$CLAUDE1_ID" ]; then
  log "Assigning Implement OAuth integration to claude-1..."
  ASSIGN1_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/direct/issue/assign" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"issue_id\":\"issue_implement_oauth\",\"agent_id\":\"$CLAUDE1_ID\",\"assigned_by\":\"usr_osakka\"}")
  
  log "Assignment 1 response: $ASSIGN1_RESPONSE"
fi

if [ -n "$CLAUDE3_ID" ]; then
  log "Assigning Performance Optimization to claude-3..."
  ASSIGN2_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/direct/issue/assign" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{\"issue_id\":\"issue_performance_optimi\",\"agent_id\":\"$CLAUDE3_ID\",\"assigned_by\":\"usr_osakka\"}")
  
  log "Assignment 2 response: $ASSIGN2_RESPONSE"
fi

# Associate pool with workspace
log "Associating Claude team pool with project workspace..."
POOL_WS_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/direct/pool/add_workspace" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"pool_id":"pool_claude_team","workspace_id":"workspace_project","added_by":"usr_osakka"}')

log "Pool-workspace association response: $POOL_WS_RESPONSE"

log "Database setup complete."

# Optionally restart the server
if [ "$1" == "--restart" ]; then
  log "Restarting server..."
  /opt/entitydb/bin/entitydbd.sh restart
fi

exit 0