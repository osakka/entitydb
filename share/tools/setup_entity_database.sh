#!/bin/bash
# setup_entity_database.sh
# Sets up a clean EntityDB database with entity-based architecture
# This script uses the new direct entity-based APIs
# Recommended to use with a clean database

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$SCRIPT_DIR")"
DB_PATH="$BASE_DIR/var/db/entitydb.db"
SERVER_URL="http://localhost:8085"

# Check if server is running
echo "Checking if EntityDB server is running..."
if ! curl -s "$SERVER_URL/api/v1/status" | grep -q "ok"; then
  echo "Error: EntityDB server is not running. Please start the server with './bin/entitydbd.sh start'"
  exit 1
fi

echo "EntityDB server is running. Proceeding with database setup..."

# Step 1: Create admin user
echo "Creating admin user 'osakka'..."
curl -s -X POST "$SERVER_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "osakka",
    "password": "mypassword",
    "email": "osakka@example.com",
    "full_name": "Admin User"
  }'

# Step 2: Login as admin to get token
echo "Logging in as admin..."
TOKEN=$(curl -s -X POST "$SERVER_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "osakka",
    "password": "mypassword"
  }' | grep -o '"token":"[^"]*' | sed 's/"token":"//')

if [ -z "$TOKEN" ]; then
  echo "Error: Failed to get authentication token"
  exit 1
fi

echo "Successfully authenticated with token"

# Step 3: Make the user an admin
echo "Granting admin role to user..."
curl -s -X POST "$SERVER_URL/api/v1/direct/rbac/admin_assign" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "username": "osakka"
  }'

# Step 4: Create Claude agents
echo "Creating claude-1 agent..."
CLAUDE1_ID=$(curl -s -X POST "$SERVER_URL/api/v1/agents/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "handle": "claude-1",
    "name": "Claude 1",
    "description": "AI assistant",
    "specialization": "general tasks, code review, writing"
  }' | grep -o '"id":"[^"]*' | sed 's/"id":"//')

echo "Creating claude-2 agent..."
CLAUDE2_ID=$(curl -s -X POST "$SERVER_URL/api/v1/agents/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "handle": "claude-2",
    "name": "Claude 2",
    "description": "AI assistant",
    "specialization": "system design, software development, problem-solving"
  }' | grep -o '"id":"[^"]*' | sed 's/"id":"//')

echo "Creating claude-3 agent..."
CLAUDE3_ID=$(curl -s -X POST "$SERVER_URL/api/v1/agents/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "handle": "claude-3",
    "name": "Claude 3",
    "description": "AI assistant",
    "specialization": "debugging, testing, documentation"
  }' | grep -o '"id":"[^"]*' | sed 's/"id":"//')

# Step 5: Create an agent pool
echo "Creating 'claude_team' agent pool..."
POOL_ID=$(curl -s -X POST "$SERVER_URL/api/v1/direct/pool/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "claude_team",
    "description": "Team of Claude agents",
    "specialization": "AI assistance and software development"
  }' | grep -o '"id":"[^"]*' | sed 's/"id":"//')

# Step 6: Add agents to the pool
echo "Adding agents to the pool..."
curl -s -X POST "$SERVER_URL/api/v1/direct/pool/add_agent" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"pool_id\": \"$POOL_ID\",
    \"agent_id\": \"$CLAUDE1_ID\",
    \"added_by\": \"osakka\"
  }"

curl -s -X POST "$SERVER_URL/api/v1/direct/pool/add_agent" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"pool_id\": \"$POOL_ID\",
    \"agent_id\": \"$CLAUDE2_ID\",
    \"added_by\": \"osakka\"
  }"

curl -s -X POST "$SERVER_URL/api/v1/direct/pool/add_agent" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"pool_id\": \"$POOL_ID\",
    \"agent_id\": \"$CLAUDE3_ID\",
    \"added_by\": \"osakka\"
  }"

# Step 7: Create a workspace using the new entity-based API
echo "Creating 'entitydb_development' workspace using entity-based API..."
WORKSPACE_ID=$(curl -s -X POST "$SERVER_URL/api/v1/direct/workspace/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "EntityDB Development",
    "description": "Main development workspace for the AI Workforce Orchestration Platform",
    "priority": "high",
    "creator_id": "osakka",
    "tags": ["area:development", "team:engineering"]
  }' | grep -o '"id":"[^"]*' | sed 's/"id":"//')

# Step 8: Add workspace to pool
echo "Adding workspace to pool..."
curl -s -X POST "$SERVER_URL/api/v1/direct/pool/add_workspace" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"pool_id\": \"$POOL_ID\",
    \"workspace_id\": \"$WORKSPACE_ID\",
    \"added_by\": \"osakka\"
  }"

# Step 9: Create issues using the new entity-based API
echo "Creating issues using entity-based API..."
ISSUE1_ID=$(curl -s -X POST "$SERVER_URL/api/v1/direct/issue/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"title\": \"Implement entity migration\",
    \"description\": \"Migrate existing tables to entity-based architecture\",
    \"priority\": \"high\",
    \"type\": \"issue\",
    \"workspace_id\": \"$WORKSPACE_ID\",
    \"creator_id\": \"osakka\",
    \"tags\": [\"area:database\", \"component:migrations\"]
  }" | grep -o '"id":"[^"]*' | sed 's/"id":"//')

ISSUE2_ID=$(curl -s -X POST "$SERVER_URL/api/v1/direct/issue/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"title\": \"Entity API documentation\",
    \"description\": \"Create comprehensive documentation for the entity-based API\",
    \"priority\": \"medium\",
    \"type\": \"issue\",
    \"workspace_id\": \"$WORKSPACE_ID\",
    \"creator_id\": \"osakka\",
    \"tags\": [\"area:documentation\", \"component:api\"]
  }" | grep -o '"id":"[^"]*' | sed 's/"id":"//')

ISSUE3_ID=$(curl -s -X POST "$SERVER_URL/api/v1/direct/issue/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"title\": \"Tag system implementation\",
    \"description\": \"Implement advanced tag system for entity filtering\",
    \"priority\": \"high\",
    \"type\": \"issue\",
    \"workspace_id\": \"$WORKSPACE_ID\",
    \"creator_id\": \"osakka\",
    \"tags\": [\"area:backend\", \"component:tags\"]
  }" | grep -o '"id":"[^"]*' | sed 's/"id":"//')

# Step 10: Assign issues to agents
echo "Assigning issues to agents..."
curl -s -X POST "$SERVER_URL/api/v1/direct/issue/assign" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"issue_id\": \"$ISSUE1_ID\",
    \"agent_id\": \"$CLAUDE1_ID\",
    \"assigned_by\": \"osakka\"
  }"

curl -s -X POST "$SERVER_URL/api/v1/direct/issue/assign" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"issue_id\": \"$ISSUE2_ID\",
    \"agent_id\": \"$CLAUDE2_ID\",
    \"assigned_by\": \"osakka\"
  }"

curl -s -X POST "$SERVER_URL/api/v1/direct/issue/assign" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"issue_id\": \"$ISSUE3_ID\",
    \"agent_id\": \"$CLAUDE3_ID\",
    \"assigned_by\": \"osakka\"
  }"

# Step 11: Update some issue statuses
echo "Updating issue statuses..."
curl -s -X POST "$SERVER_URL/api/v1/direct/issue/status" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"issue_id\": \"$ISSUE1_ID\",
    \"status\": \"in_progress\",
    \"updated_by\": \"osakka\"
  }"

echo "Setup complete!"
echo "Created workspace: $WORKSPACE_ID"
echo "Created issues: $ISSUE1_ID, $ISSUE2_ID, $ISSUE3_ID"
echo "Created agent pool: $POOL_ID"
echo "Created agents: $CLAUDE1_ID, $CLAUDE2_ID, $CLAUDE3_ID"
echo ""
echo "You can now login with username 'osakka' and password 'mypassword'"