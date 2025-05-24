# Multi-Dataspace Architecture Implementation Plan

## Overview
Implementation of multi-database (dataspace) system with tag-based inheritance and enhanced session management for EntityDB.

## Design Decisions Made

### 1. Permission Evaluation Strategy ✅
**Choice**: Simple tag matching (Option A)
- `rbac:perm:entity:write:dataspace:worca` = can write entities with `dataspace:worca`
- Natural fit with EntityDB's existing tag system
- High performance, minimal changes to RBAC middleware

### 2. Session Management Strategy ✅
**Choice**: Sessions as entities + relationships (Options A + C)
- Session entities: `type:session`, `dataspace:worca`, `status:active`
- User-session relationships via EntityDB relationship system
- Multi-session support per user
- Rich session metadata and temporal tracking

### 3. Dataspace + Trait/Self Architecture ✅
**Choice**: Dataspace namespace with trait/self separation
```
dataspace:worca                          # Dataspace membership
worca:self:type:task              # Entity's own properties
worca:self:assignee:john
worca:trait:org:TechCorp          # Inherited/shared context
worca:trait:project:MobileApp
```

**Logic**: 
- `self:` = What I am (my identity)
- `trait:` = What I belong to/inherit from (my context)
- No auto-inheritance = Explicit on request (controlled)

## Architecture Components

### Dataspace System
- **Dataspace Tag**: `dataspace:worca` (mandatory on all entities)
- **Self Namespace**: `worca:self:namespace:value` (entity's own properties)
- **Trait Namespace**: `worca:trait:namespace:value` (inherited context)

### Session Management
- **Session Entities**: `type:session`, `dataspace:worca`, `status:active/expired`
- **Session-User Relationships**: Link users to their active sessions
- **Multi-session Support**: Users can have multiple concurrent sessions
- **Dataspace-scoped Sessions**: Sessions belong to specific dataspaces

### RBAC Enhancement
- **Dataspace-scoped Permissions**: `rbac:perm:entity:write:dataspace:worca`
- **Trait-based Permissions**: `rbac:perm:entity:read:worca:trait:org:TechCorp`
- **Self-based Permissions**: `rbac:perm:entity:update:worca:self:assignee:self`

## Implementation Steps

### Phase 1: Core Dataspace Infrastructure
1. **Dataspace Validation Middleware** - Ensure all entities have `dataspace:` tag
2. **Dataspace-scoped Queries** - Modify query handlers to filter by dataspace
3. **API Endpoints** - Add dataspace parameter to create/update operations
4. **RBAC Integration** - Update permission checking for dataspace-scoped access

### Phase 2: Trait/Self System
1. **Tag Namespace Handling** - Parse and validate `dataspacename:self/trait:namespace:value`
2. **API Enhancement** - Support dataspace, self, traits in request/response
3. **Query Filters** - Add trait/self filtering capabilities
4. **Documentation** - API docs and examples

### Phase 3: Session Management
1. **Session Entities** - Create session entity type with dataspace association
2. **Session-User Relationships** - Link sessions to users
3. **Session Lifecycle** - Login/logout session state management
4. **Multi-session Support** - Allow multiple concurrent sessions per user

### Phase 4: Testing & Documentation
1. **Unit Tests** - Dataspace, trait/self, session functionality
2. **Integration Tests** - End-to-end multi-dataspace scenarios
3. **Performance Tests** - Dataspace-scoped query performance
4. **Documentation** - Complete API documentation and examples

## Design Decisions Continued

### 4. RBAC Granularity Level ✅
**Choice**: All levels supported (Option D)
- **Dataspace-level**: `rbac:perm:entity:write:dataspace:worca`
- **Trait-level**: `rbac:perm:entity:write:worca:trait:org:TechCorp` 
- **Self-level**: `rbac:perm:entity:write:worca:self:type:task`
- **Maximum flexibility**: Users can have any combination

### 5. Cross-Dataspace Access ✅
**Choice**: Global admin override + Ultra strict isolation (Options C + A)
- **Global Admin**: `rbac:perm:*` can access all dataspaces (natural part of tagging system)
- **Regular Users**: Strict dataspace isolation - can only access their assigned dataspace
- **Security**: No accidental cross-dataspace data leaks
- **Administration**: Global admin can manage all dataspaces

### 6. Dataspace Creation Permissions ✅
**Choice**: Explicit dataspace permissions via RBAC tags
- **Dataspace Management Permissions**:
  - `rbac:perm:dataspace:create` - Can create new dataspaces
  - `rbac:perm:dataspace:delete` - Can delete dataspaces
  - `rbac:perm:dataspace:manage:worca` - Can manage specific dataspace
  - `rbac:perm:dataspace:assign-admin` - Can assign dataspace admins
  - `rbac:perm:dataspace:*` - All dataspace management permissions
- **Natural Permission Flow**: Global admin has all, dataspace creators get create, dataspace admins get manage
- **Granular Control**: Dataspace-specific management permissions

### 7. Session Cleanup Strategy ✅
**Choice**: Hybrid approach (Option C)
- **Auto-expiration**: Sessions have `expires_at:timestamp` tag for security
- **Activity Extension**: Active sessions get `expires_at` updated on API calls
- **Manual Logout**: Users can explicitly set `status:expired` 
- **Grace Period**: Configurable session timeout (default: 24h, extendable to 7 days max)
- **Cleanup Process**: Background task to mark expired sessions as `status:expired`

### 8. User Identity Scope ✅
**Choice**: Users are global, dataspace access via RBAC (Option A)
- **Global Identity**: `username:john` unique system-wide
- **Dataspace Access**: Controlled entirely by RBAC permissions
- **Multi-dataspace Users**: Users get dataspace access via `rbac:perm:*:dataspace:*` tags
- **Session Flexibility**: Sessions can be dataspace-specific or multi-dataspace based on permissions
- **Natural Scaling**: Add more dataspace permissions to expand user access

## Complete Design Summary

All design decisions finalized! Ready for implementation:
1. ✅ Simple tag matching for permissions
2. ✅ Sessions as entities + relationships  
3. ✅ Dataspace + trait/self architecture
4. ✅ All RBAC granularity levels supported
5. ✅ Global admin + strict dataspace isolation
6. ✅ Explicit dataspace management permissions
7. ✅ Hybrid session cleanup strategy
8. ✅ Global users with RBAC-driven dataspace access

## Implementation Status

### ✅ **COMPLETED - Phase 1: Core Dataspace Infrastructure**
- **Dataspace Validation Middleware** - `dataspace_middleware.go` with dataspace access control ✅
- **Dataspace-scoped RBAC** - Permission checking for dataspace isolation ✅  
- **API Enhancement** - Dataspace-aware entity creation and queries ✅
- **Route Integration** - New dataspace endpoints added to main.go ✅

### ✅ **COMPLETED - Phase 2: Trait/Self System**
- **Tag Namespace Handling** - `dataspacename:self/trait:namespace:value` parsing ✅
- **API Enhancement** - Dataspace, self, traits in request/response ✅
- **Query Filters** - Trait/self filtering capabilities ✅
- **Dataspace Entity Creation** - Working dataspace-aware entity creation ✅

### 🔄 **IN PROGRESS - Phase 2 Debugging**
- **Dataspace Listing Bug Fix** - Minor issue with dataspace entity parsing (nearly complete)

### ⏳ **PENDING - Phase 3: Session Management**
- Session entities with dataspace association
- Session-user relationships  
- Multi-session support per user
- Session lifecycle management

### ⏳ **PENDING - Phase 4: Testing & Documentation**
- Comprehensive testing
- Performance validation
- API documentation updates

## Implemented Features

### **Multi-Dataspace API Endpoints** 🚀
```
POST   /api/v1/dataspaces/create              # Create new dataspace
GET    /api/v1/dataspaces/list                # List accessible dataspaces  
DELETE /api/v1/dataspaces/delete              # Delete dataspace (if empty)
POST   /api/v1/dataspaces/entities/create     # Create dataspace-aware entities
GET    /api/v1/dataspaces/entities/query      # Query dataspace entities
```

### **Dataspace-Aware Entity Format** ✨
```json
{
  "dataspace": "worca",
  "self": {"type": "task", "status": "todo", "title": "Test Task"},
  "traits": {"org": "TechCorp", "project": "Mobile"},
  "content": "Task description"
}
```

### **Tag Structure Implementation** 🏷️
```
dataspace:worca                          # Dataspace membership
worca:self:type:task              # Entity's own properties  
worca:self:status:todo
worca:trait:org:TechCorp          # Inherited context
worca:trait:project:Mobile
```

### **RBAC Permissions** 🔐
```
rbac:perm:entity:*:dataspace:worca      # Full dataspace access
rbac:perm:dataspace:create               # Can create dataspaces
rbac:perm:dataspace:manage:worca        # Can manage specific dataspace
```

## Testing Results

### ✅ **Successful Tests**
- **Dataspace Creation**: `POST /api/v1/dataspaces/create` ✅
- **Dataspace-Aware Entity Creation**: Entities with dataspace/self/traits ✅  
- **Dataspace Entity Queries**: Filtering by dataspace, self, traits ✅
- **RBAC Integration**: Permission checks working ✅
- **Build & Deploy**: Clean compilation and server restart ✅

### 🔧 **Minor Issue**
- **Dataspace Listing**: Returns inconsistent results (debugging in progress)

## Next Steps
1. ✅ Complete dataspace listing bug fix
2. 📝 Update Worca to use new dataspace-aware API
3. 🔄 Implement session management with dataspace scoping
4. 📋 Create comprehensive test suite
5. 📚 Update API documentation
6. 🚀 Git commit and deployment