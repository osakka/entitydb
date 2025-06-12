# Entity API Usage Examples

This document provides practical examples of using the new entity-based API endpoints. These examples can help you migrate from the legacy API to the new entity-based architecture.

## Prerequisites

- EntityDB server running on localhost port 8085
- Admin user credentials (e.g., `osakka`/`mypassword`)

## Authentication

```bash
# Login to get a JWT token
TOKEN=$(curl -s -X POST "http://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "osakka",
    "password": "mypassword"
  }' | grep -o '"token":"[^"]*' | sed 's/"token":"//')

# Save token for use in subsequent requests
echo "Authorization: Bearer $TOKEN" > auth_header.txt
```

## Workspace Operations

### Create a Workspace

```bash
# Create a workspace
curl -X POST "http://localhost:8085/api/v1/direct/workspace/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "New Workspace",
    "description": "A new workspace for development",
    "priority": "high",
    "tags": ["area:backend", "team:engineering"]
  }'
```

### List Workspaces

```bash
# List all workspaces
curl -X GET "http://localhost:8085/api/v1/direct/workspace/list" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

### Get Workspace Details

```bash
# Get workspace by ID
curl -X GET "http://localhost:8085/api/v1/direct/workspace/get?workspace_id=workspace_new_workspace" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

## Issue Operations

### Create an Issue

```bash
# Create an issue in a workspace
curl -X POST "http://localhost:8085/api/v1/direct/issue/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Implement Feature X",
    "description": "Add a new feature to the system",
    "priority": "medium",
    "type": "issue",
    "workspace_id": "workspace_new_workspace",
    "tags": ["component:api", "difficulty:medium"]
  }'
```

### List Issues

```bash
# List all issues
curl -X GET "http://localhost:8085/api/v1/direct/issue/list" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"

# List issues filtered by workspace
curl -X GET "http://localhost:8085/api/v1/direct/issue/list?workspace_id=workspace_new_workspace" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"

# List issues filtered by status
curl -X GET "http://localhost:8085/api/v1/direct/issue/list?status=pending" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"

# List issues filtered by priority
curl -X GET "http://localhost:8085/api/v1/direct/issue/list?priority=high" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

### Get Issue Details

```bash
# Get issue by ID
curl -X GET "http://localhost:8085/api/v1/direct/issue/get?issue_id=issue_implement_feature_x" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

### Assign an Issue

```bash
# Assign issue to agent
curl -X POST "http://localhost:8085/api/v1/direct/issue/assign" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "issue_id": "issue_implement_feature_x",
    "agent_id": "agent_claude_1",
    "assigned_by": "osakka"
  }'
```

### Update Issue Status

```bash
# Update issue status to in_progress
curl -X POST "http://localhost:8085/api/v1/direct/issue/status" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "issue_id": "issue_implement_feature_x",
    "status": "in_progress",
    "updated_by": "agent_claude_1"
  }'

# Update issue status to completed
curl -X POST "http://localhost:8085/api/v1/direct/issue/status" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "issue_id": "issue_implement_feature_x",
    "status": "completed",
    "updated_by": "agent_claude_1"
  }'

# Update issue status to blocked
curl -X POST "http://localhost:8085/api/v1/direct/issue/status" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "issue_id": "issue_implement_feature_x",
    "status": "blocked",
    "updated_by": "agent_claude_1"
  }'
```

## Advanced Entity Operations

### Create a Custom Entity

```bash
# Create a custom entity
curl -X POST "http://localhost:8085/api/v1/entities" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "id": "custom_entity_1",
    "tags": [
      "type=document", 
      "status=draft", 
      "author=claude"
    ],
    "content": [
      {
        "type": "title",
        "value": "Design Document"
      },
      {
        "type": "description",
        "value": "Design document for the new feature"
      },
      {
        "type": "content",
        "value": "This is the content of the design document..."
      }
    ]
  }'
```

### Create an Entity Relationship

```bash
# Create a relationship between entities
curl -X POST "http://localhost:8085/api/v1/entity-relationships" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "source_id": "issue_implement_feature_x",
    "relationship_type": "has_document",
    "target_id": "custom_entity_1",
    "metadata": {
      "created_by": "osakka",
      "created_at": "2025-05-11T12:00:00Z"
    }
  }'
```

### List Entity Relationships

```bash
# List relationships by source
curl -X GET "http://localhost:8085/api/v1/entity-relationships/source?source_id=issue_implement_feature_x" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"

# List relationships by target
curl -X GET "http://localhost:8085/api/v1/entity-relationships/target?target_id=custom_entity_1" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN"
```

## Shell Script Examples

Here are some shell script examples for common operations:

### Create and Assign an Issue

```bash
#!/bin/bash

# Login
TOKEN=$(curl -s -X POST "http://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "osakka",
    "password": "mypassword"
  }' | grep -o '"token":"[^"]*' | sed 's/"token":"//')

# Create issue
ISSUE_ID=$(curl -s -X POST "http://localhost:8085/api/v1/direct/issue/create" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "New Task",
    "description": "This is a new task",
    "priority": "high",
    "type": "issue",
    "workspace_id": "workspace_entitydb_development",
    "tags": ["area:backend"]
  }' | grep -o '"id":"[^"]*' | sed 's/"id":"//')

echo "Created issue: $ISSUE_ID"

# Assign to agent
curl -s -X POST "http://localhost:8085/api/v1/direct/issue/assign" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"issue_id\": \"$ISSUE_ID\",
    \"agent_id\": \"agent_claude_1\",
    \"assigned_by\": \"osakka\"
  }"

echo "Issue assigned to agent_claude_1"
```

### List Issues for a Workspace

```bash
#!/bin/bash

# Login
TOKEN=$(curl -s -X POST "http://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "osakka",
    "password": "mypassword"
  }' | grep -o '"token":"[^"]*' | sed 's/"token":"//')

# Get workspace ID
WORKSPACE_ID="workspace_entitydb_development"

# List issues for workspace
curl -s -X GET "http://localhost:8085/api/v1/direct/issue/list?workspace_id=$WORKSPACE_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

These examples should help you get started with the new entity-based API. For more details, see the full API documentation and architecture guide.