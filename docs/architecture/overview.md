# EntityDB Architecture Documentation

## System Overview

EntityDB (EntityDB) is a distributed system designed to manage AI agents, issues, and workspaces through a unified entity-based architecture.

## Architecture Principles

### 1. Entity-Based Design
- All data objects are represented as entities
- Entities are classified and organized using hierarchical tags
- Relationships between entities are explicitly modeled
- No specialized tables or data structures

### 2. Tag-Based Classification
- Hierarchical namespace structure: `namespace:category:subcategory:value`
- Tags provide flexible categorization and permission control
- Wildcard support for permissions (`rbac:perm:*`)

### 3. API-First Architecture
- All operations go through RESTful API
- Embedded credential authentication (v2.29.0+)
- Tag-based permission checks
- No direct database access

### 4. Embedded Authentication (v2.29.0+)
- User credentials stored directly in entity content field
- No separate credential entities or relationships needed
- Simple format: `salt|bcrypt_hash` in user entity content
- Users marked with `has:credentials` tag for identification

## System Components

### Server (server_db.go)

The main server implementation provides:
- HTTP request routing
- JWT authentication
- Entity CRUD operations
- In-memory entity storage
- SQLite database persistence
- Static file serving

Key features:
- Single binary deployment
- Integrated authentication
- Auto-refresh for entities
- WebSocket-ready architecture

### API Layer (/src/api/)

API handlers implement:
- Entity operations (create, read, update, delete)
- Entity relationship management
- Authentication endpoints
- Permission middleware
- Legacy endpoint compatibility

Key modules:
- `entity_handler.go` - Entity CRUD operations
- `entity_relationship_handler.go` - Relationship management
- `auth_permissions.go` - Permission checking
- `router.go` - HTTP routing

### Data Models (/src/models/)

Core data structures:
- **Entity** - Base object with tags and properties
- **EntityRelationship** - Connections between entities
- **TagHierarchy** - Hierarchical tag parsing
- **User** - Authentication and authorization

Repository interfaces:
- `EntityRepository` - Entity persistence
- `EntityRelationshipRepository` - Relationship persistence
- `UserRepository` - User management
- `TokenStore` - JWT token storage

### Web Interface (/share/htdocs/)

Alpine.js-based dashboard:
- Real-time entity updates
- Tag-based filtering
- Inline editing
- Auto-refresh (60 seconds)
- Dark/light themes

Key files:
- `index.html` - Main dashboard
- `js/app.js` - Alpine.js application
- `css/style.css` - Tailwind CSS styles

### CLI Tools (/share/cli/)

Command-line interfaces:
- `entitydb-api.sh` - Bash CLI wrapper
- `entitydb_client.py` - Python client library
- `test_api.sh` - API testing script

## Data Flow

### 1. Request Flow
```
Client → HTTP Server → Router → Middleware → Handler → Repository → Database
                                     ↓
                            Permission Check
```

### 2. Authentication Flow (v2.29.0+)
```
Login → Content-based Credential Check → JWT Generation → Token Storage → Request Headers → Validation
```

### 3. Entity Operations
```
Create → Tag Assignment → Validation → Storage → Response
Read → Permission Check → Filter → Transform → Response
Update → Permission Check → Merge → Storage → Response
Delete → Permission Check → Remove → Response
```

## Tag Namespace Architecture

### Core Namespaces

1. **type:** - Entity classification
   ```
   type:user
   type:agent
   type:issue
   type:workspace
   ```

2. **rbac:** - Role-based access control
   ```
   rbac:role:admin
   rbac:perm:entity:create
   rbac:perm:*
   ```

3. **status:** - Entity state
   ```
   status:active
   status:pending
   status:completed
   ```

4. **id:** - Unique identifiers
   ```
   id:username:admin
   id:agent:claude-2
   id:issue:issue_123
   ```

### Tag Resolution

Tags are resolved hierarchically:
1. Exact match: `rbac:perm:entity:create`
2. Wildcard match: `rbac:perm:entity:*`
3. Global wildcard: `rbac:perm:*`

## Security Architecture

### Authentication (v2.29.0+)
- Embedded credentials in user entity content
- Username/password validation via bcrypt
- JWT tokens with configurable expiry
- Token refresh mechanism
- Secure password hashing with unique salts

### Authorization
- Tag-based permission system
- Hierarchical permission inheritance
- Per-endpoint permission requirements
- Wildcard permission support

### Security Headers
- CORS configuration
- Content-Type validation
- Authorization header checks

## Database Design

### Entity Table
```sql
CREATE TABLE entities (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    tags TEXT NOT NULL,  -- JSON array
    content TEXT,        -- JSON data
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT
);
```

### Entity Relationships Table
```sql
CREATE TABLE entity_relationships (
    id TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    target_id TEXT NOT NULL,
    type TEXT NOT NULL,
    properties TEXT,     -- JSON data
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    FOREIGN KEY (source_id) REFERENCES entities(id),
    FOREIGN KEY (target_id) REFERENCES entities(id)
);
```

## Scalability Considerations

### Current Limitations
- In-memory entity storage in server_db.go
- No connection pooling
- Single-server deployment
- No caching layer

### Future Scalability
- Implement proper database repository pattern
- Add Redis caching layer
- Support horizontal scaling
- Implement connection pooling
- Add message queue for async operations

## API Design Patterns

### RESTful Conventions
- GET for reads
- POST for creates
- PUT for updates
- DELETE for deletes

### Response Format
```json
{
  "status": "ok|error",
  "message": "Human-readable message",
  "data": { },
  "error": "Error details if applicable"
}
```

### Query Parameters
- `type` - Filter by entity type
- `tags` - Comma-separated tag list
- `status` - Filter by status
- `source` - Filter relationships by source
- `target` - Filter relationships by target

## Development Patterns

### Clean Tabletop Policy
- Single source of truth
- No duplicate files
- Immediate deprecation removal
- Frequent commits

### Testing Strategy
- Unit tests for models
- API tests for endpoints
- Integration tests for workflows
- Performance tests for scalability

### Error Handling
- Consistent error responses
- Proper HTTP status codes
- Detailed error messages in development
- Generic messages in production

## Deployment Architecture

### Single Binary
- Compiled Go binary
- Embedded static files
- Configuration via flags/environment
- SQLite database co-located

### Directory Structure
```
/opt/entitydb/
├── bin/entitydb         # Server binary
├── var/
│   ├── db/         # SQLite database
│   └── log/        # Application logs
├── share/
│   ├── htdocs/     # Web UI files
│   └── cli/        # CLI tools
└── docs/           # Documentation
```

### Process Management
- Systemd service (recommended)
- PID file tracking
- Signal handling (SIGTERM, SIGINT)
- Graceful shutdown

## Future Architecture Goals

1. **Microservices Migration**
   - Separate API gateway
   - Entity service
   - Auth service
   - Notification service

2. **Event Streaming**
   - Real-time updates via WebSocket
   - Event sourcing for audit trail
   - Change data capture

3. **Distributed Storage**
   - PostgreSQL for ACID compliance
   - Redis for caching
   - S3 for file storage

4. **Container Orchestration**
   - Docker containers
   - Kubernetes deployment
   - Horizontal pod autoscaling
   - Service mesh integration