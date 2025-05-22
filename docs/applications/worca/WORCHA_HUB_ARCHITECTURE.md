# Worca Hub Architecture Design

## Overview
Worca (Workforce Orchestrator) serves as the reference implementation for EntityDB's Multi-Hub Platform, demonstrating sophisticated tag-based inheritance and enterprise RBAC patterns.

## Core Design Principle

**Hub**: `worca` - The workforce orchestrator application space
**Self Properties**: What the entity IS (type, status, own attributes)  
**Trait Properties**: What the entity BELONGS TO (organizational context)

## Entity Mapping

### 1. Organizations
```javascript
{
  "hub": "worca",
  "self": {
    "type": "organization",
    "name": "TechCorp",
    "status": "active"
  },
  "traits": {
    "industry": "technology",
    "size": "large",
    "region": "north-america"
  },
  "content": {
    "description": "Leading technology company",
    "founded": "2010",
    "employees": 500,
    "headquarters": "San Francisco, CA"
  }
}
```

**Tags Generated:**
- `hub:worca`
- `worca:self:type:organization`
- `worca:self:name:TechCorp`
- `worca:self:status:active`
- `worca:trait:industry:technology`
- `worca:trait:size:large`
- `worca:trait:region:north-america`

### 2. Projects
```javascript
{
  "hub": "worca",
  "self": {
    "type": "project",
    "name": "MobileApp",
    "status": "active",
    "priority": "high",
    "progress": "75"
  },
  "traits": {
    "org": "TechCorp",
    "team": "mobile-development",
    "technology": "react-native",
    "phase": "beta"
  },
  "content": {
    "description": "Next-generation mobile application",
    "budget": 250000,
    "deadline": "2025-12-31",
    "lead": "john.doe"
  }
}
```

**Tags Generated:**
- `hub:worca`
- `worca:self:type:project`
- `worca:self:name:MobileApp`
- `worca:self:status:active`
- `worca:self:priority:high`
- `worca:self:progress:75`
- `worca:trait:org:TechCorp`
- `worca:trait:team:mobile-development`
- `worca:trait:technology:react-native`
- `worca:trait:phase:beta`

### 3. Epics
```javascript
{
  "hub": "worca",
  "self": {
    "type": "epic",
    "name": "UserAuthentication",
    "status": "in-progress",
    "complexity": "high",
    "story_points": "34"
  },
  "traits": {
    "org": "TechCorp",
    "project": "MobileApp",
    "feature_area": "security",
    "priority_tier": "p1"
  },
  "content": {
    "description": "Complete user authentication system",
    "acceptance_criteria": ["OAuth2 integration", "2FA support", "Password reset"],
    "dependencies": ["SecurityFramework", "DatabaseSchema"]
  }
}
```

### 4. Stories
```javascript
{
  "hub": "worca",
  "self": {
    "type": "story",
    "name": "LoginFlow",
    "status": "ready",
    "story_points": "8",
    "complexity": "medium"
  },
  "traits": {
    "org": "TechCorp",
    "project": "MobileApp",
    "epic": "UserAuthentication",
    "feature": "login",
    "persona": "end-user"
  },
  "content": {
    "description": "As a user, I want to log in securely",
    "acceptance_criteria": ["Email/password login", "Remember me option", "Error handling"],
    "mockups": ["login-screen.png", "error-states.png"]
  }
}
```

### 5. Tasks
```javascript
{
  "hub": "worca",
  "self": {
    "type": "task",
    "title": "Implement OAuth2 callback handler",
    "status": "doing",
    "assignee": "john.doe",
    "priority": "high",
    "estimated_hours": "4"
  },
  "traits": {
    "org": "TechCorp",
    "project": "MobileApp", 
    "epic": "UserAuthentication",
    "story": "LoginFlow",
    "sprint": "Sprint-24",
    "component": "backend-api"
  },
  "content": {
    "description": "Implement secure OAuth2 callback handling with token validation",
    "technical_notes": "Use passport.js strategy, validate state parameter",
    "definition_of_done": ["Unit tests written", "Integration tests pass", "Code reviewed"]
  }
}
```

### 6. Users (Team Members)
```javascript
{
  "hub": "worca",
  "self": {
    "type": "user",
    "username": "john.doe",
    "display_name": "John Doe",
    "status": "active",
    "role": "senior-developer"
  },
  "traits": {
    "org": "TechCorp",
    "team": "mobile-development",
    "department": "engineering",
    "location": "san-francisco",
    "skill_level": "senior"
  },
  "content": {
    "email": "john.doe@techcorp.com",
    "skills": ["React Native", "Node.js", "OAuth2", "Testing"],
    "capacity": "40",
    "timezone": "PST"
  }
}
```

### 7. Sprints
```javascript
{
  "hub": "worca",
  "self": {
    "type": "sprint",
    "name": "Sprint-24",
    "status": "active",
    "velocity": "42",
    "progress": "65"
  },
  "traits": {
    "org": "TechCorp",
    "project": "MobileApp",
    "team": "mobile-development",
    "quarter": "Q2-2025"
  },
  "content": {
    "start_date": "2025-05-15",
    "end_date": "2025-05-29",
    "goal": "Complete user authentication MVP",
    "committed_points": "42",
    "completed_points": "27"
  }
}
```

## Query Patterns

### 1. Hierarchical Queries
```javascript
// All tasks in TechCorp organization
GET /api/v1/hubs/entities/query?hub=worca&self=type:task&traits=org:TechCorp

// All items in MobileApp project
GET /api/v1/hubs/entities/query?hub=worca&traits=project:MobileApp

// User authentication work items
GET /api/v1/hubs/entities/query?hub=worca&traits=epic:UserAuthentication

// Current sprint tasks
GET /api/v1/hubs/entities/query?hub=worca&self=type:task&traits=sprint:Sprint-24

// John's active tasks
GET /api/v1/hubs/entities/query?hub=worca&self=type:task,status:doing,assignee:john.doe
```

### 2. Cross-Cutting Queries  
```javascript
// All high-priority work across organization
GET /api/v1/hubs/entities/query?hub=worca&self=priority:high&traits=org:TechCorp

// Mobile team capacity
GET /api/v1/hubs/entities/query?hub=worca&self=type:user&traits=team:mobile-development

// Security-related work items
GET /api/v1/hubs/entities/query?hub=worca&traits=feature_area:security

// Backend component tasks
GET /api/v1/hubs/entities/query?hub=worca&self=type:task&traits=component:backend-api
```

### 3. Analytics Queries
```javascript
// Project velocity over time
GET /api/v1/hubs/entities/query?hub=worca&self=type:sprint&traits=project:MobileApp

// Team workload distribution  
GET /api/v1/hubs/entities/query?hub=worca&self=type:task,status:doing&traits=team:mobile-development

// Epic progress tracking
GET /api/v1/hubs/entities/query?hub=worca&traits=epic:UserAuthentication

// Organization-wide metrics
GET /api/v1/hubs/entities/query?hub=worca&traits=org:TechCorp
```

## RBAC Permission Mapping

### 1. Organization-Level Permissions
```bash
# Full organization access
rbac:perm:entity:*:worca:trait:org:TechCorp

# Project-specific access  
rbac:perm:entity:*:worca:trait:project:MobileApp

# Team-specific access
rbac:perm:entity:read:worca:trait:team:mobile-development
```

### 2. Role-Based Permissions
```bash
# Project Manager
rbac:perm:entity:*:worca:trait:project:MobileApp
rbac:perm:entity:create:worca:self:type:epic
rbac:perm:entity:update:worca:self:type:sprint

# Team Lead  
rbac:perm:entity:*:worca:trait:team:mobile-development
rbac:perm:entity:assign:worca:self:assignee:*

# Developer
rbac:perm:entity:read:worca:trait:team:mobile-development  
rbac:perm:entity:update:worca:self:assignee:self
rbac:perm:entity:update:worca:self:status:*

# Product Owner
rbac:perm:entity:*:worca:trait:project:MobileApp
rbac:perm:entity:create:worca:self:type:story
rbac:perm:entity:update:worca:self:priority:*
```

### 3. Granular Task Permissions
```bash
# Can only update own task status
rbac:perm:entity:update:worca:self:status:*:assignee:self

# Can assign tasks within team
rbac:perm:entity:update:worca:self:assignee:*:trait:team:mobile-development

# Can view all project tasks
rbac:perm:entity:read:worca:self:type:task:trait:project:MobileApp
```

## Data Migration Strategy

### 1. Current Worca Data → Hub Architecture
```javascript
// OLD FORMAT (current)
{
  "id": "task123",
  "tags": ["type:task", "status:doing", "assignee:john", "project:MobileApp"],
  "title": "Implement OAuth2",
  "content": "Task description"
}

// NEW FORMAT (hub architecture)  
{
  "hub": "worca",
  "self": {
    "type": "task", 
    "title": "Implement OAuth2",
    "status": "doing",
    "assignee": "john"
  },
  "traits": {
    "project": "MobileApp",
    "org": "TechCorp",
    "epic": "UserAuthentication"
  },
  "content": "Task description"
}
```

### 2. Backward Compatibility
- Keep existing API endpoints working
- Gradually migrate frontend to use hub endpoints
- Support both tag formats during transition
- Provide migration utilities for data transformation

## Benefits of Hub Architecture for Worca

### 1. **Cleaner Data Model**
- Clear separation: what entities ARE vs what they BELONG TO
- Natural hierarchy without complex tag parsing
- Intuitive querying with inheritance

### 2. **Enhanced Performance**  
- Hub-scoped queries are faster
- Trait-based filtering is optimized
- Better indexing on self vs trait properties

### 3. **Superior RBAC**
- Project-level access control
- Team-based permissions  
- Granular task assignment rules
- Organization-wide security policies

### 4. **Scalability**
- Multiple organizations in single Worca instance
- Clear tenant isolation
- Cross-project analytics
- Enterprise-ready architecture

### 5. **Developer Experience**
- Intuitive API patterns
- Self-documenting data structure  
- Reusable permission templates
- Clear inheritance model

## Implementation Phases

### Phase 1: API Wrapper Update
- Modify `worca-api.js` to use hub endpoints
- Transform data between old and new formats
- Maintain UI compatibility

### Phase 2: Frontend Enhancement  
- Update components to leverage trait-based queries
- Implement advanced filtering by traits
- Add cross-project views

### Phase 3: Advanced Features
- Multi-organization support
- Advanced RBAC implementation  
- Performance optimizations
- Analytics enhancements

**Worca will become the definitive reference for building sophisticated applications on EntityDB's Multi-Hub Platform!** 🚀