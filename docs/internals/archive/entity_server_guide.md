# EntityDB Entity Server Implementation Guide

## Overview

The EntityDB Entity Server is a secure API-based server that implements the entity-based architecture without allowing direct database access. This guide explains how to use and interact with the server.

## Key Features

1. **Pure Entity-Based Architecture**: All objects are stored as entities with flexible schemas
2. **Zero Direct Database Access**: All data access is through secure API endpoints
3. **Authentication Required**: All entity API operations require proper authentication
4. **Role-Based Authorization**: Administrative operations require admin role
5. **User-Friendly Dashboard**: Web interface for easy server status monitoring

## Server Architecture

The server implements a pure entity-based approach:

- **Entities**: Generic object containers with IDs, tags, and content
- **Relationships**: Connections between entities (parent/child, assignment, etc.)
- **Tags**: Metadata for flexible filtering and categorization
- **API First**: All operations through HTTP API, no direct database access

## Starting the Server

To start the server in entity-based mode:

```bash
/opt/entitydb/bin/entitydbd.sh use-db start
```

This builds and runs the specialized entity-based server implementation located at `/opt/entitydb/src/server_db.go`.

## Authentication

All entity API operations require authentication:

```bash
# Login to get a token
curl -X POST "http://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"osakka", "password":"mypassword"}'

# Response includes token
{
  "message": "Login successful",
  "status": "ok",
  "token": "tk_osakka_1746951212618116280",
  "user": {
    "id": "usr_osakka",
    "roles": ["admin"],
    "username": "osakka"
  }
}
```

## Authorization Header

Use the token in Authorization header for all API requests:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8085/api/v1/direct/workspace/list
```

## Entity API Endpoints

### Status and Health

- `GET /api/v1/status` - Server status information (no auth required)
- `GET /api/v1/health` - Server health check (no auth required)

### Authentication

- `POST /api/v1/auth/login` - Authenticate and get token

### Workspaces

- `GET /api/v1/direct/workspace/list` - List all workspaces
- `GET /api/v1/direct/workspace/get` - Get workspace by ID
- `POST /api/v1/direct/workspace/create` - Create a new workspace

### Issues

- `GET /api/v1/direct/issue/list` - List all issues
- `GET /api/v1/direct/issue/get` - Get issue by ID
- `POST /api/v1/direct/issue/create` - Create a new issue
- `POST /api/v1/direct/issue/assign` - Assign issue to agent
- `POST /api/v1/direct/issue/status` - Update issue status

### Raw Entity Operations

- `GET /api/v1/entities` - List entities with tag filtering
- `POST /api/v1/entities` - Create a new entity
- `GET /api/v1/entity-relationships/source` - Get relationships by source
- `POST /api/v1/entity-relationships` - Create relationship

## Usage Examples

### Creating a Workspace

```bash
curl -X POST "http://localhost:8085/api/v1/direct/workspace/create" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Workspace",
    "description": "A new workspace for development",
    "priority": "high"
  }'
```

### Creating an Issue

```bash
curl -X POST "http://localhost:8085/api/v1/direct/issue/create" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Issue",
    "description": "This is a new issue",
    "priority": "medium",
    "workspace_id": "entity_sample_2",
    "type": "issue"
  }'
```

### Listing Issues

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8085/api/v1/direct/issue/list"
```

## Testing the Server

A test script is provided to verify server functionality:

```bash
/opt/entitydb/bin/test_entity_server.sh
```

This script:
1. Checks if the server is running
2. Tests authentication
3. Tests workspace and issue operations
4. Ensures proper authorization is enforced

## Dashboard

The server provides a simple web dashboard at http://localhost:8085/ showing:
- Server status and uptime
- API mode (entity-based)
- Available API endpoints

## Server Implementation Details

The server code is implemented in `/opt/entitydb/src/server_db.go` with the following components:

1. **User Management**: In-memory user store with roles
2. **Token Authentication**: JWT-style token generation and validation
3. **API Routing**: Request routing based on URL paths
4. **Security Enforcement**: Authorization checks for all entity operations
5. **Entity Handlers**: Logic for various entity operations

## Security Considerations

The server implements several security measures:
- Authentication required for all entity operations
- Role-based access control for administrative operations
- Token expiration (24 hours)
- No direct database access
- Input validation

## Troubleshooting

If you encounter issues:

1. Check server logs: `/opt/entitydb/var/log/entitydb.log`
2. Verify the server is running: `/opt/entitydb/bin/entitydbd.sh status`
3. Check authentication token is valid
4. Ensure proper authorization headers