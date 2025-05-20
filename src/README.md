# EntityDB Platform (EntityDB)

> [!IMPORTANT]
> The system now uses a pure entity-based architecture with all operations consolidated under a unified entity API. Zero tolerance for specialized endpoints and direct database access. Legacy endpoints are redirected to the entity API with deprecation notices.

## Overview

EntityDB is a comprehensive system for managing AI agents, their issues, and collaboration workflows. It uses a modern client-server architecture with a single command-line client for all agent interactions. The Role-Based Access Control (RBAC) system is fully implemented with JWT-based authentication. User permissions are managed through roles and are effective immediately.

### What the System Does
- **Orchestrates** AI agents across multiple workspaces and issues
- **Tracks** work progress and agent performance 
- **Manages** issue assignments and dependencies
- **Standardizes** agent communication and collaboration
- **Provides** unified command interface for all operations

### System Architecture
- **entitydbd** - Entity-based server with SQLite persistence (runs on port 8085)
- **entitydbc** - Command-line client for all worker interactions
- **web dashboard** - Web interface for system monitoring (available at http://localhost:8085)

### Directory Structure
- **/opt/entitydb/bin/** - System binaries and essential scripts
  - **/opt/entitydb/bin/entitydb** - Main server binary
  - **/opt/entitydb/bin/entitydbc.sh** - Client wrapper script
  - **/opt/entitydb/bin/entitydbd.sh** - Server wrapper script
  - **/opt/entitydb/bin/c.sh** - Shorthand client script
- **/opt/entitydb/src/** - Source code
  - **/opt/entitydb/src/server_db.go** - Pure entity server implementation
  - **/opt/entitydb/src/api/** - API endpoints and handlers
    - `entity_handler.go` - Entity API endpoint implementations
    - `entity_relationship_handler.go` - Relationship API endpoint implementations
    - `entity_issue_handler.go` - Legacy compatibility layer (redirects to entity API)
  - **/opt/entitydb/src/core/** - Core server components
  - **/opt/entitydb/src/models/** - Data models and repositories
    - `entity.go` - Core entity model definition
    - `entity_relationship.go` - Entity relationship model definition
    - `/opt/entitydb/src/models/interfaces/` - Repository interfaces
    - `/opt/entitydb/src/models/sqlite/` - SQLite implementations
    - `/opt/entitydb/src/models/memory/` - In-memory implementations
  - **/opt/entitydb/src/tests/** - Test scripts and utilities
- **/opt/entitydb/var/** - Variable data (database, logs, PID files)
  - **/opt/entitydb/var/db/** - SQLite database storage
  - **/opt/entitydb/var/log/** - Log files
- **/opt/entitydb/share/** - Shared resources
  - **/opt/entitydb/share/htdocs/** - Web server root directory
  - **/opt/entitydb/share/tools/** - Helper scripts and utilities
  - **/opt/entitydb/share/tests/** - Test scripts and testing tools
- **/opt/entitydb/docs/** - System documentation
  - **/opt/entitydb/docs/archive/** - Archived legacy documentation

## Current Project Status

### Pure Entity-Based Architecture

The system now operates exclusively with an entity-based architecture:

1. **Pure Entity Model**:
   - ALL objects must be represented as entities with tags
   - Use entity relationships for connections between objects
   - Never create specialized tables or data structures

2. **API-First Design**:
   - ALL operations must go through the entity API
   - ZERO tolerance for direct database access
   - Always use the proper JWT authentication

3. **Legacy Compatibility**:
   - Redirect legacy endpoints to the entity API
   - Include deprecation notices with all legacy endpoint responses
   - Never add new specialized endpoints

### System Modules

The system consists of the following modules, all implemented with the unified entity-based architecture:

#### Entity Module
Core module providing the underlying architecture for all system objects.

- **Status**: Fully implemented with SQLite backend
- **Description**: Generic entity creation, retrieval, update, and deletion
- **Features**: Flexible tag-based classification, content storage, and entity relationships

#### Entity Relationships
Module for managing connections between entities.

- **Status**: Fully implemented with SQLite backend  
- **Description**: Typed relationships between entities for hierarchies, dependencies, assignments, etc.

#### Agent Module
Implemented as specialized entity types with entity relationships.

- **Status**: Fully implemented on the entity model
- **Description**: Agent registration, profiling, and capability management
- **Features**: Agents are stored as entities with appropriate tags and relationships

#### Session Module
Implemented as specialized entity types with entity relationships.

- **Status**: Fully implemented on the entity model
- **Description**: Session creation, status tracking, and context management
- **Features**: Sessions are stored as entities with agent and workspace relationships

#### Issue Module
Implemented as specialized entity types with entity relationships.

- **Status**: Fully implemented on the entity model
- **Description**: Issue creation, assignment, status tracking, and dependency management
- **Features**: All issue types (workspace, epic, story, issue, subissue) are entities with different tags

### Entity Structure

An entity consists of the following core attributes:

```json
{
  "id": "entity_1234567890",
  "type": "issue",
  "title": "Example Entity",
  "description": "Detailed description of the entity",
  "status": "in_progress",
  "tags": ["api", "backend", "high-priority"],
  "properties": {
    "priority": "high",
    "estimate": "2h",
    "complexity": "medium"
  },
  "created_at": "2023-08-15T10:30:00Z",
  "updated_at": "2023-08-16T08:45:00Z",
  "created_by": "usr_admin",
  "assigned_to": "claude-2"
}
```

### Entity Relationships

Relationships between entities are explicitly stored as relationship objects:

```json
{
  "id": "rel_1234567890",
  "source_id": "entity_workspace_123",
  "target_id": "entity_issue_456",
  "type": "parent",
  "properties": {
    "order": 1
  },
  "created_at": "2023-08-15T10:35:00Z",
  "created_by": "usr_admin"
}
```

## API Structure

The system uses a consolidated entity API with:

- **Entity API**: Create, read, update, delete and list operations for all entities
- **Entity Relationship API**: Create, read, update, delete and list operations for entity relationships
- **Legacy API Redirection**: Redirects specialized endpoints to the entity API with deprecation notices

### Entity API Endpoints

```
GET /api/v1/entities/list
```

Optional query parameters:
- `type`: Filter by entity type
- `status`: Filter by status
- `tags`: Filter by tags (comma-separated)

Example:
```bash
curl -X GET "http://localhost:8085/api/v1/entities/list?type=issue&status=in_progress" \
  -H "Authorization: Bearer tk_admin_1234567890"
```

### Entity Relationship API Endpoints

```
GET /api/v1/entity-relationships/list
```

Optional query parameters:
- `source`: Filter by source entity ID
- `target`: Filter by target entity ID
- `type`: Filter by relationship type

Example:
```bash
curl -X GET "http://localhost:8085/api/v1/entity-relationships/list?source=entity_workspace_123" \
  -H "Authorization: Bearer tk_admin_1234567890"
```

### Legacy API Redirection

All legacy API endpoints are now redirected to the unified entity API with deprecation notices:

```bash
curl -X GET http://localhost:8085/api/v1/direct/workspace/list \
  -H "Authorization: Bearer tk_admin_1234567890"
```

Response:
```json
{
  "status": "ok",
  "data": [...],
  "count": 2,
  "deprecation_notice": "WARNING: This legacy workspace endpoint is deprecated and will be removed soon. Please use /api/v1/entities/list?type=workspace instead."
}
```

## Authentication and Security

- **JWT-based Authentication**: Secure token-based authentication for all operations
- **Role-Based Access Control**: Permission management through user roles
- **API-First Design**: All operations through HTTP API, zero direct database access

## Quick Start Guide

As a worker, you only need to interact with the `./bin/entitydbc.sh` command. This single tool handles all your needs:

```bash
# Check your worker handle
echo $WORKER_ID

# Register with the system (first time only)
./bin/entitydbc.sh agent register \
  --handle=$WORKER_ID \
  --name="Your Name" \
  --specialization="Your specialty areas"

# Verify your registration
./bin/entitydbc.sh agent list | grep $WORKER_ID

# Create a new session
./bin/entitydbc.sh session create \
  --agent=$WORKER_ID \
  --workspace=entitydb \
  --name="Session description" \
  --description="Detailed session information"

# List your active sessions
./bin/entitydbc.sh session list --agent=$WORKER_ID

# Create a new issue
./bin/entitydbc.sh issue create \
  --title="Issue title" \
  --description="Detailed issue description" \
  --priority=medium
```

### Working with Entities Directly

```bash
# Creating Entities
./bin/entitydbc.sh entity create --type=issue --title="Fix login bug" --tags="priority:high,status:pending,workspace:entitydb"

# Listing Entities
./bin/entitydbc.sh entity list --type=issue --tags="priority:high"

# Getting Entity Details
./bin/entitydbc.sh entity get --id=entity_1234

# Creating Relationships
./bin/entitydbc.sh entity relationship create --source=entity_123 --target=entity_456 --type=depends_on

# Listing Relationships
./bin/entitydbc.sh entity relationship list --source=entity_123
```

## Server Management

The EntityDB server can be controlled using the `./bin/entitydbd.sh` command:

```bash
# Start the server
./bin/entitydbd.sh start

# Check server status
./bin/entitydbd.sh status

# Stop the server
./bin/entitydbd.sh stop  

# Restart the server
./bin/entitydbd.sh restart
```

The server stores its PID in `/opt/entitydb/var/entitydb.pid` and logs to `/opt/entitydb/var/log/entitydb.log`. The database is located at `/opt/entitydb/var/db/entitydb.db`.

## Building and Running

### Building the Entity Server

```bash
cd /opt/entitydb/src
go build -o entitydb_server_entity server_db.go  # Build the entity server
```

The entity server can also be built automatically when running the `entitydbd.sh` script.

### Running Tests

```bash
# Run entity API tests
/opt/entitydb/share/tests/entity/test_entity_server.sh

# Run specific tests for entity functionality
/opt/entitydb/share/tests/entity/test_entity_api.sh

# Run general API tests
cd /opt/entitydb/src
make test          # Run all tests (unit, API, and master tests)
make unit-tests    # Run only Go unit tests
make api-tests     # Run only API tests
```

## Implementation Status

The EntityDB system fully implements:

1. **Pure Entity-Based Architecture** with complete consolidation
   - All operations through unified entity API
   - Zero specialized endpoints (legacy endpoints redirect to entity API)
   - No direct database access
   - Complete entity relationship system
   - Tag-based filtering and categorization

2. **Role-Based Access Control (RBAC)** system with permissions and roles
   - JWT-based authentication
   - Role hierarchies and permission inheritance
   - Fine-grained access control with middleware
   - Comprehensive permission validation
   - Role assignment APIs
   - Full test coverage for RBAC functionality

3. **Entity API System**
   - Entity creation, retrieval, update, and deletion
   - Entity relationship management
   - Tag-based querying and filtering
   - Content-based storage for entity data
   - Full implementation of all previous functionality through the entity API

4. **Agent System** (implemented through entity API)
   - Agent registration and profile management
   - Capability tracking with proficiency levels
   - Status updates and activity monitoring
   - Agent linking with user accounts

5. **Issue System** (implemented through entity API)
   - Issue creation, assignment, and lifecycle management
   - Support for workspace, epic, story, issue, and subissue types
   - Detailed progress tracking and status updates
   - Issue dependencies and relationships
   - Hierarchical organization (workspaces > epics > stories > issues > subissues)
   - Time tracking and metrics collection

6. **Session Management** (implemented through entity API)
   - Context storage and retrieval
   - Session activity tracking
   - Workspace-specific sessions
   - Session statistics

7. **Legacy API Compatibility**
   - Redirection of legacy endpoints to entity API
   - Deprecation notices for all legacy API usage
   - Backward compatibility for existing client usage

8. **Custom HTTP Router**
   - Method-based routing (GET, POST, PUT, DELETE)
   - Middleware chain for request processing
   - Static file serving for web dashboard
   - Logging and request tracking
   - CORS support for web clients

9. **Command-Line Client**
   - Comprehensive Bash-based client
   - Token management
   - Full API coverage for both entity API and legacy endpoints
   - User-friendly command structure

## Pending Issues

The following issues are currently pending implementation:

1. **Fix Authenticated User Context in Issue Creation**
   - Location: `/opt/entitydb/src/api/issue.go` (line 81)
   - Description: Extract the authenticated user from request context instead of using the "system" placeholder
   - Acceptance Criteria: Issues are attributed to the actual users who created them

2. **Fix User Context in Issue Pool Assignment**
   - Location: `/opt/entitydb/src/api/issue_pool_assignment.go` (line 61)
   - Description: Extract the authenticated user from request context instead of using the "system" placeholder
   - Acceptance Criteria: Issue pool assignments are attributed to the actual users

3. **Resolve Redundant Database Abstraction**
   - Location: `/opt/entitydb/src/core/database.go`
   - Description: Either implement the TODOs or remove the file if SQLite implementations are used directly
   - Acceptance Criteria: Clear documentation on database access approach with no redundant abstractions

4. **Implement Transaction Support**
   - Location: `/opt/entitydb/src/core/database.go` (lines 128, 159)
   - Description: Implement proper transaction support for operations requiring atomicity
   - Acceptance Criteria: Transactions are properly supported for multi-step operations

5. **Entity API Documentation Improvements**
   - Description: Create comprehensive documentation for entity API usage
   - Acceptance Criteria:
     - Detailed API reference documentation
     - Examples for all common operations
     - Description of entity relationship types
     - Guidance on tag-based filtering
   - Priority: Medium

## Benefits of Pure Entity Architecture

1. **Flexibility**: Any object type can be represented as an entity with appropriate tags
2. **Extensibility**: New entity types can be added without schema changes
3. **Unified API**: Common API for all entity types
4. **Rich Relationships**: Flexible relationship types between entities
5. **Advanced Filtering**: Tag-based filtering for powerful queries
6. **Simplified Codebase**: Single data model versus multiple specialized models
7. **Future-Proof**: Easy to adapt to new requirements without schema changes
8. **Consolidated Access**: All operations go through a single API
9. **Zero Direct Database Access**: Enhanced security through API-only access
10. **Legacy Compatibility**: Redirection from legacy endpoints with deprecation notices

## Development Guidelines

### Architecture Guidelines

1. **Pure Entity Model**:
   - ALL objects must be represented as entities with tags
   - Use entity relationships for connections between objects
   - Never create specialized tables or data structures

2. **API-First Design**:
   - ALL operations must go through the entity API
   - ZERO tolerance for direct database access
   - Always use the proper JWT authentication

3. **Legacy Compatibility**:
   - Redirect legacy endpoints to the entity API
   - Include deprecation notices with all legacy endpoint responses
   - Never add new specialized endpoints

### Clean Tabletop Policy

1. **Single Source of Truth**: We always work on the main code directly.
   - DO NOT create duplicate files or versions
   - DO NOT use temporary extensions like .bak, .new, .old, etc.
   - DO NOT create backup copies or alternative implementations in separate directories
   - ALWAYS DELETE DEPRECATED CODE

2. **File Organization**:
   - Server code belongs in the `/src/` directory
   - Helper scripts belong in `/share/tools/` directory
   - Tests belong in `/share/tests/` directory
   - Common utilities and shared code belongs in `/src/core/`

3. **Version Control**:
   - We use Trunk-based development (main branch)
   - Our version control system is gitea
   - Use the git repository https://git.home.arpa/ your username is claude-2 password claude-password
   - Use git branches for experimental changes
   - Keep the main branch clean and buildable at all times
   - Git commit VERY frequently with clear messages
   - Git push frequently