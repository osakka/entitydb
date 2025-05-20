#!/bin/bash
#
# ========================================================
# EntityDB Initialization Script
# ========================================================
#
# This script initializes the EntityDB database with the following:
#
# USERS:
# - osakka (admin password: osakka123)
#
# AGENT POOLS:
# - administrator: Administrative agents with full system access
# - user: Regular user agents with basic access
# - software engineer: Software engineering agents
# - quality assurance: Quality assurance and testing agents
# - technical writer: Documentation and technical writing agents
# - software architect: System architecture and design agents
#
# AGENTS:
# - osakka-agent: Human agent for Omar Sakka
# - claude-agent: AI agent for code analysis and debugging
# 
# WORKSPACES:
# - entitydb: Default workspace for the AI Workforce Orchestration platform
# - tcc_build: Workspace for TCC build project
#
# RBAC ROLES:
# - Administrator: Full system access
# - User: Basic user role with limited permissions
#
# SAMPLE ISSUES:
# - Creates a hierarchy of epics, stories, and issues for both workspaces
# - Assigns several issues to agents
#
# REQUIREMENTS:
# - The EntityDB server must be running
# - The database must exist with default admin/password credentials
#
# USAGE:
#   ./initialize_entitydb.sh
#
# NOTES:
# - If components already exist, the script will attempt to continue
# - The script uses the EntityDB client commands where possible and direct API calls otherwise
#
# ========================================================

set -e  # Exit immediately if a command exits with a non-zero status
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")"
CLIENT="${BASE_DIR}/bin/entitydbc.sh"
SERVER_URL="http://localhost:8085"

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

check_server() {
    log_info "Checking if EntityDB server is running..."
    if ! curl -s "$SERVER_URL/api/v1/status" > /dev/null; then
        log_error "EntityDB server is not running. Please start the server using ./bin/entitydbd.sh start"
    fi
    log_success "EntityDB server is running"
}

login_admin() {
    log_info "Logging in as admin..."
    if ! "$CLIENT" login --username=admin --password=password > /dev/null; then
        log_error "Failed to login as admin. Please check if the server is running and has the default admin account."
    fi
    log_success "Logged in as admin"
}

# Main execution starts here
echo "=========================================================="
echo "EntityDB Initialization Script"
echo "=========================================================="
echo "This script will initialize the EntityDB database with:"
echo "- Users: osakka (administrator)"
echo "- Agent pools: administrator, user, software engineer, "
echo "               quality assurance, technical writer, software architect"
echo "- Workspaces: entitydb, tcc_build"
echo "- Sample issues with parent-child relationships"
echo "=========================================================="
echo ""

# Check if the server is running
check_server

# Login as admin
login_admin

create_user() {
    local username=$1
    local password=$2
    local full_name=$3
    local email=$4
    
    log_info "Creating user: $username..."
    
    # User the API directly since there's no client command for user creation
    local result=$(curl -s -X POST "$SERVER_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\",\"full_name\":\"$full_name\",\"email\":\"$email\"}")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to create user $username: $result"
        log_info "User might already exist, continuing..."
    else
        log_success "User $username created successfully!"
    fi
}

add_user_to_admin_role() {
    local username=$1
    
    log_info "Adding user $username to admin role..."
    
    # Use the API directly to update user roles
    curl -s -X POST "$SERVER_URL/api/v1/rbac/user/role/assign" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $(cat ~/.entitydb/token)" \
        -d "{\"username\":\"$username\",\"role_id\":\"role_admin\"}"
    
    log_success "User $username added to admin role"
}

# Create the osakka user with admin permissions
create_user "osakka" "osakka123" "Omar Sakka" "osakka@example.com"
add_user_to_admin_role "osakka"

create_agent_pool() {
    local name=$1
    local description=$2
    
    log_info "Creating agent pool: $name..."
    
    # Use the API directly since there's no client command for pool creation
    local result=$(curl -s -X POST "$SERVER_URL/api/v1/pools/create" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $(cat ~/.entitydb/token)" \
        -d "{\"name\":\"$name\",\"description\":\"$description\"}")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to create pool $name: $result"
        log_info "Pool might already exist, continuing..."
    else
        local pool_id=$(echo "$result" | grep -o '"id":"[^"]*"' | sed 's/"id":"//;s/"//')
        log_success "Pool $name created with ID: $pool_id"
        echo "$pool_id"
    fi
}

create_agent() {
    local handle=$1
    local name=$2
    local specialization=$3
    
    log_info "Creating agent: $handle..."
    
    # Use the agent register command since it's available
    local result=$("$CLIENT" agent register \
        --handle="$handle" \
        --name="$name" \
        --specialization="$specialization")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to create agent $handle: $result"
        log_info "Agent might already exist, continuing..."
    else
        log_success "Agent $handle created successfully!"
    fi
}

assign_agent_to_pool() {
    local agent_handle=$1
    local pool_id=$2
    
    log_info "Assigning agent $agent_handle to pool $pool_id..."
    
    # Use the API directly
    local result=$(curl -s -X POST "$SERVER_URL/api/v1/pools/agents/add" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $(cat ~/.entitydb/token)" \
        -d "{\"pool_id\":\"$pool_id\",\"agent_id\":\"$agent_handle\"}")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to assign agent $agent_handle to pool $pool_id: $result"
    else
        log_success "Agent $agent_handle assigned to pool $pool_id"
    fi
}

link_user_to_agent() {
    local username=$1
    local agent_handle=$2
    
    log_info "Linking user $username to agent $agent_handle..."
    
    # Use the API directly
    local result=$(curl -s -X PUT "$SERVER_URL/api/v1/user/update" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $(cat ~/.entitydb/token)" \
        -d "{\"username\":\"$username\",\"agent_id\":\"$agent_handle\"}")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to link user $username to agent $agent_handle: $result"
    else
        log_success "User $username linked to agent $agent_handle"
    fi
}

# Create agent pools
log_info "Creating agent pools..."
POOL_ADMIN=$(create_agent_pool "administrator" "Administrative agents with full system access")
POOL_USER=$(create_agent_pool "user" "Regular user agents with basic access")
POOL_SE=$(create_agent_pool "software engineer" "Software engineering agents")
POOL_QA=$(create_agent_pool "quality assurance" "Quality assurance and testing agents")
POOL_TW=$(create_agent_pool "technical writer" "Documentation and technical writing agents")
POOL_SA=$(create_agent_pool "software architect" "System architecture and design agents")

# Create agents
log_info "Creating agents..."
create_agent "osakka-agent" "Omar Sakka" "System administration, engineering"
create_agent "claude-agent" "Claude" "code analysis, debugging, issue management"

# Assign agents to pools
log_info "Assigning agents to pools..."
assign_agent_to_pool "osakka-agent" "$POOL_ADMIN"
assign_agent_to_pool "claude-agent" "$POOL_SE"

# Link users to agents
log_info "Linking users to agents..."
link_user_to_agent "osakka" "osakka-agent"

create_workspace() {
    local title=$1
    local description=$2
    local priority="${3:-medium}"
    
    log_info "Creating workspace: $title..."
    
    # Use the issue create command with type=workspace
    local result=$("$CLIENT" issue create \
        --title="$title" \
        --description="$description" \
        --type=workspace \
        --priority="$priority")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to create workspace $title: $result"
        log_info "Workspace might already exist, continuing..."
    else
        local workspace_id=$(echo "$result" | grep -o '"id":"[^"]*"' | sed 's/"id":"//;s/"//')
        log_success "Workspace $title created with ID: $workspace_id"
        echo "$workspace_id"
    fi
}

assign_pool_to_workspace() {
    local pool_id=$1
    local workspace_id=$2
    
    log_info "Assigning pool $pool_id to workspace $workspace_id..."
    
    # Use the API directly
    local result=$(curl -s -X POST "$SERVER_URL/api/v1/pools/workspaces/add" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $(cat ~/.entitydb/token)" \
        -d "{\"pool_id\":\"$pool_id\",\"workspace_id\":\"$workspace_id\"}")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to assign pool $pool_id to workspace $workspace_id: $result"
    else
        log_success "Pool $pool_id assigned to workspace $workspace_id"
    fi
}

# Create workspaces if they don't exist already
log_info "Creating workspaces..."
WS_EntityDB=$(create_workspace "EntityDB Default Workspace" "Default workspace for the AI Workforce Orchestration platform")
# If workspace creation failed, try to get the ID of the existing workspace
if [[ -z "$WS_EntityDB" ]]; then
    log_info "Trying to find existing EntityDB workspace..."
    WS_EntityDB="workspace_entitydb"
    log_info "Using default workspace ID: $WS_EntityDB"
fi

WS_TCC=$(create_workspace "TCC Build Workspace" "Workspace for TCC build project")
# If workspace creation failed, try to get the ID of the existing workspace
if [[ -z "$WS_TCC" ]]; then
    log_info "Trying to find existing TCC workspace..."
    WS_TCC="workspace_tcc_build"
    log_info "Using default workspace ID: $WS_TCC"
fi

# Assign pools to workspaces
log_info "Assigning pools to workspaces..."
for pool_var in POOL_ADMIN POOL_SE POOL_QA POOL_TW POOL_SA; do
    pool_id=${!pool_var}
    if [[ -n "$pool_id" ]]; then
        assign_pool_to_workspace "$pool_id" "$WS_EntityDB"
        assign_pool_to_workspace "$pool_id" "$WS_TCC"
    else
        log_warning "Skipping pool $pool_var as it was not created successfully"
    fi
done

create_role() {
    local name=$1
    local description=$2
    
    log_info "Creating role: $name..."
    
    # Use the API directly
    local result=$(curl -s -X POST "$SERVER_URL/api/v1/rbac/role/create" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $(cat ~/.entitydb/token)" \
        -d "{\"name\":\"$name\",\"description\":\"$description\"}")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to create role $name: $result"
        log_info "Role might already exist, continuing..."
        
        # Try to get the role ID by listing roles
        local roles=$(curl -s -X GET "$SERVER_URL/api/v1/rbac/role/list" \
            -H "Authorization: Bearer $(cat ~/.entitydb/token)")
        local role_id=$(echo "$roles" | grep -o "\"id\":\"[^\"]*\",\"name\":\"$name\"" | grep -o "\"id\":\"[^\"]*\"" | sed 's/"id":"//;s/"//')
        
        if [[ -n "$role_id" ]]; then
            log_info "Found existing role $name with ID: $role_id"
            echo "$role_id"
        else
            log_warning "Could not find role $name"
            echo ""
        fi
    else
        local role_id=$(echo "$result" | grep -o '"id":"[^"]*"' | sed 's/"id":"//;s/"//')
        log_success "Role $name created with ID: $role_id"
        echo "$role_id"
    fi
}

add_permission_to_role() {
    local role_id=$1
    local permission=$2
    
    log_info "Adding permission $permission to role $role_id..."
    
    # Use the API directly
    local result=$(curl -s -X POST "$SERVER_URL/api/v1/rbac/role/permission/add/$role_id" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $(cat ~/.entitydb/token)" \
        -d "{\"permission\":\"$permission\",\"role_id\":\"$role_id\"}")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to add permission $permission to role $role_id: $result"
    else
        log_success "Permission $permission added to role $role_id"
    fi
}

# Create roles if they don't exist already
log_info "Creating roles..."
ROLE_ADMIN=$(create_role "Administrator" "Full system access")
ROLE_USER=$(create_role "User" "Basic user role with limited permissions")

# Add permissions to roles
log_info "Adding permissions to roles..."

# Administrator role - full access
log_info "Adding administrator permissions..."
for permission in "system.view" "system.configure" \
                 "agent.view" "agent.create" "agent.update" "agent.delete" \
                 "session.view" "session.create" "session.update" \
                 "issue.view" "issue.create" "issue.update" "issue.assign" "issue.delete"; do
    add_permission_to_role "$ROLE_ADMIN" "$permission"
done

# User role - limited access
log_info "Adding user permissions..."
for permission in "system.view" \
                 "agent.view" \
                 "session.view" "session.create" \
                 "issue.view" "issue.create"; do
    add_permission_to_role "$ROLE_USER" "$permission"
done

create_epic() {
    local title=$1
    local description=$2
    local workspace_id=$3
    local priority="${4:-medium}"
    
    if [[ -z "$workspace_id" ]]; then
        log_warning "Cannot create epic '$title' - workspace ID is empty"
        return
    fi
    
    log_info "Creating epic: $title in workspace $workspace_id..."
    
    local result=$("$CLIENT" issue create \
        --title="$title" \
        --description="$description" \
        --workspace="$workspace_id" \
        --type=epic \
        --priority="$priority")
            
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to create epic using issue create: $result"
        return
    fi
    
    local epic_id=$(echo "$result" | grep -o '"id":"[^"]*"' | head -1 | sed 's/"id":"//;s/"//')
    log_success "Epic $title created with ID: $epic_id"
    echo "$epic_id"
}

create_story() {
    local title=$1
    local description=$2
    local workspace_id=$3
    local epic_id=$4
    local priority="${5:-medium}"
    
    if [[ -z "$workspace_id" ]]; then
        log_warning "Cannot create story '$title' - workspace ID is empty"
        return
    fi
    
    if [[ -z "$epic_id" ]]; then
        log_warning "Cannot create story '$title' - epic ID is empty"
        return
    fi
    
    log_info "Creating story: $title in workspace $workspace_id, epic $epic_id..."
    
    local result=$("$CLIENT" issue create \
        --title="$title" \
        --description="$description" \
        --workspace="$workspace_id" \
        --type=story \
        --parent="$epic_id" \
        --priority="$priority")
            
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to create story using issue create: $result"
        return
    fi
    
    local story_id=$(echo "$result" | grep -o '"id":"[^"]*"' | head -1 | sed 's/"id":"//;s/"//')
    log_success "Story $title created with ID: $story_id"
    echo "$story_id"
}

create_issue() {
    local title=$1
    local description=$2
    local workspace_id=$3
    local parent_id=$4
    local priority="${5:-medium}"
    
    if [[ -z "$workspace_id" ]]; then
        log_warning "Cannot create issue '$title' - workspace ID is empty"
        return
    fi
    
    if [[ -z "$parent_id" ]]; then
        log_warning "Cannot create issue '$title' - parent ID is empty"
        return
    fi
    
    log_info "Creating issue: $title in workspace $workspace_id, parent $parent_id..."
    
    local result=$("$CLIENT" issue create \
        --title="$title" \
        --description="$description" \
        --workspace="$workspace_id" \
        --type=issue \
        --parent="$parent_id" \
        --priority="$priority")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to create issue $title: $result"
        return
    fi
    
    local issue_id=$(echo "$result" | grep -o '"id":"[^"]*"' | head -1 | sed 's/"id":"//;s/"//')
    log_success "Issue $title created with ID: $issue_id"
    echo "$issue_id"
}

assign_issue() {
    local issue_id=$1
    local agent_handle=$2
    
    if [[ -z "$issue_id" ]]; then
        log_warning "Cannot assign issue - issue ID is empty"
        return
    fi
    
    if [[ -z "$agent_handle" ]]; then
        log_warning "Cannot assign issue $issue_id - agent handle is empty"
        return
    fi
    
    log_info "Assigning issue $issue_id to agent $agent_handle..."
    
    local result=$("$CLIENT" issue assign "$issue_id" --agent="$agent_handle")
    
    if [[ "$result" == *"error"* ]]; then
        log_warning "Failed to assign issue $issue_id to agent $agent_handle: $result"
    else
        log_success "Issue $issue_id assigned to agent $agent_handle"
    fi
}

# Create sample issues with parent-child relationships
log_info "Creating sample issues for workspace $WS_EntityDB..."

# Create epics
EPIC1=$(create_epic "System Enhancements" "Core system enhancements for the EntityDB platform" "$WS_EntityDB" "high")
EPIC2=$(create_epic "Documentation Updates" "Documentation updates for the EntityDB platform" "$WS_EntityDB" "medium")

# Create stories under epics
STORY1=$(create_story "Improve User Authentication" "Enhance the user authentication system" "$WS_EntityDB" "$EPIC1" "high")
STORY2=$(create_story "Optimize Database Queries" "Improve database query performance" "$WS_EntityDB" "$EPIC1" "medium")
STORY3=$(create_story "Update User Guide" "Update the user guide with new features" "$WS_EntityDB" "$EPIC2" "medium")
STORY4=$(create_story "Create API Documentation" "Create comprehensive API documentation" "$WS_EntityDB" "$EPIC2" "high")

# Create issues under stories
ISSUE1=$(create_issue "Add 2FA Support" "Implement two-factor authentication support" "$WS_EntityDB" "$STORY1" "high")
ISSUE2=$(create_issue "Refactor Password Reset Flow" "Improve the password reset flow" "$WS_EntityDB" "$STORY1" "medium")
ISSUE3=$(create_issue "Optimize Query Performance" "Improve query performance by adding indexes" "$WS_EntityDB" "$STORY2" "high")
ISSUE4=$(create_issue "Update User Guide Screenshots" "Take and update screenshots in the user guide" "$WS_EntityDB" "$STORY3" "low")
ISSUE5=$(create_issue "Document Authentication API" "Document the authentication API endpoints" "$WS_EntityDB" "$STORY4" "medium")

# Assign issues to agents
assign_issue "$ISSUE1" "osakka-agent"
assign_issue "$ISSUE3" "claude-agent"
assign_issue "$ISSUE5" "claude-agent"

log_info "Creating sample issues for workspace $WS_TCC..."

# Create epics for TCC Build workspace
TCC_EPIC1=$(create_epic "Build System Improvements" "Improvements to the TCC build system" "$WS_TCC" "high")
TCC_EPIC2=$(create_epic "Performance Optimizations" "Performance optimizations for TCC" "$WS_TCC" "medium")

# Create stories under epics
TCC_STORY1=$(create_story "Upgrade Build Tools" "Upgrade build tools to latest versions" "$WS_TCC" "$TCC_EPIC1" "high")
TCC_STORY2=$(create_story "Add CI/CD Integration" "Add continuous integration and deployment" "$WS_TCC" "$TCC_EPIC1" "medium")
TCC_STORY3=$(create_story "Memory Usage Optimization" "Optimize memory usage in TCC" "$WS_TCC" "$TCC_EPIC2" "high")

# Create issues under stories
TCC_ISSUE1=$(create_issue "Upgrade to Webpack 5" "Upgrade the build system to Webpack 5" "$WS_TCC" "$TCC_STORY1" "high")
TCC_ISSUE2=$(create_issue "Set up GitHub Actions" "Configure GitHub Actions for CI/CD" "$WS_TCC" "$TCC_STORY2" "medium")
TCC_ISSUE3=$(create_issue "Identify Memory Leaks" "Identify and fix memory leaks in TCC" "$WS_TCC" "$TCC_STORY3" "high")

# Assign issues to agents
assign_issue "$TCC_ISSUE1" "osakka-agent"
assign_issue "$TCC_ISSUE3" "claude-agent"

echo ""
echo "=========================================================="
log_success "EntityDB initialization completed successfully!"
echo "=========================================================="