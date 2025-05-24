# Multi-Hub Architecture Implementation Plan

## Overview
Implementation of multi-database (hub) system with tag-based inheritance and enhanced session management for EntityDB.

## Design Decisions Made

### 1. Permission Evaluation Strategy ✅
**Choice**: Simple tag matching (Option A)
- `rbac:perm:entity:write:hub:worca` = can write entities with `hub:worca`
- Natural fit with EntityDB's existing tag system
- High performance, minimal changes to RBAC middleware

### 2. Session Management Strategy ✅
**Choice**: Sessions as entities + relationships (Options A + C)
- Session entities: `type:session`, `hub:worca`, `status:active`
- User-session relationships via EntityDB relationship system
- Multi-session support per user
- Rich session metadata and temporal tracking

### 3. Hub + Trait/Self Architecture ✅
**Choice**: Hub namespace with trait/self separation
```
hub:worca                          # Hub membership
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

### Hub System
- **Hub Tag**: `hub:worca` (mandatory on all entities)
- **Self Namespace**: `worca:self:namespace:value` (entity's own properties)
- **Trait Namespace**: `worca:trait:namespace:value` (inherited context)

### Session Management
- **Session Entities**: `type:session`, `hub:worca`, `status:active/expired`
- **Session-User Relationships**: Link users to their active sessions
- **Multi-session Support**: Users can have multiple concurrent sessions
- **Hub-scoped Sessions**: Sessions belong to specific hubs

### RBAC Enhancement
- **Hub-scoped Permissions**: `rbac:perm:entity:write:hub:worca`
- **Trait-based Permissions**: `rbac:perm:entity:read:worca:trait:org:TechCorp`
- **Self-based Permissions**: `rbac:perm:entity:update:worca:self:assignee:self`

## Implementation Steps

### Phase 1: Core Hub Infrastructure
1. **Hub Validation Middleware** - Ensure all entities have `hub:` tag
2. **Hub-scoped Queries** - Modify query handlers to filter by hub
3. **API Endpoints** - Add hub parameter to create/update operations
4. **RBAC Integration** - Update permission checking for hub-scoped access

### Phase 2: Trait/Self System
1. **Tag Namespace Handling** - Parse and validate `hubname:self/trait:namespace:value`
2. **API Enhancement** - Support hub, self, traits in request/response
3. **Query Filters** - Add trait/self filtering capabilities
4. **Documentation** - API docs and examples

### Phase 3: Session Management
1. **Session Entities** - Create session entity type with hub association
2. **Session-User Relationships** - Link sessions to users
3. **Session Lifecycle** - Login/logout session state management
4. **Multi-session Support** - Allow multiple concurrent sessions per user

### Phase 4: Testing & Documentation
1. **Unit Tests** - Hub, trait/self, session functionality
2. **Integration Tests** - End-to-end multi-hub scenarios
3. **Performance Tests** - Hub-scoped query performance
4. **Documentation** - Complete API documentation and examples

## Design Decisions Continued

### 4. RBAC Granularity Level ✅
**Choice**: All levels supported (Option D)
- **Hub-level**: `rbac:perm:entity:write:hub:worca`
- **Trait-level**: `rbac:perm:entity:write:worca:trait:org:TechCorp` 
- **Self-level**: `rbac:perm:entity:write:worca:self:type:task`
- **Maximum flexibility**: Users can have any combination

### 5. Cross-Hub Access ✅
**Choice**: Global admin override + Ultra strict isolation (Options C + A)
- **Global Admin**: `rbac:perm:*` can access all hubs (natural part of tagging system)
- **Regular Users**: Strict hub isolation - can only access their assigned hub
- **Security**: No accidental cross-hub data leaks
- **Administration**: Global admin can manage all hubs

### 6. Hub Creation Permissions ✅
**Choice**: Explicit hub permissions via RBAC tags
- **Hub Management Permissions**:
  - `rbac:perm:hub:create` - Can create new hubs
  - `rbac:perm:hub:delete` - Can delete hubs
  - `rbac:perm:hub:manage:worca` - Can manage specific hub
  - `rbac:perm:hub:assign-admin` - Can assign hub admins
  - `rbac:perm:hub:*` - All hub management permissions
- **Natural Permission Flow**: Global admin has all, hub creators get create, hub admins get manage
- **Granular Control**: Hub-specific management permissions

### 7. Session Cleanup Strategy ✅
**Choice**: Hybrid approach (Option C)
- **Auto-expiration**: Sessions have `expires_at:timestamp` tag for security
- **Activity Extension**: Active sessions get `expires_at` updated on API calls
- **Manual Logout**: Users can explicitly set `status:expired` 
- **Grace Period**: Configurable session timeout (default: 24h, extendable to 7 days max)
- **Cleanup Process**: Background task to mark expired sessions as `status:expired`

### 8. User Identity Scope ✅
**Choice**: Users are global, hub access via RBAC (Option A)
- **Global Identity**: `username:john` unique system-wide
- **Hub Access**: Controlled entirely by RBAC permissions
- **Multi-hub Users**: Users get hub access via `rbac:perm:*:hub:*` tags
- **Session Flexibility**: Sessions can be hub-specific or multi-hub based on permissions
- **Natural Scaling**: Add more hub permissions to expand user access

## Complete Design Summary

All design decisions finalized! Ready for implementation:
1. ✅ Simple tag matching for permissions
2. ✅ Sessions as entities + relationships  
3. ✅ Hub + trait/self architecture
4. ✅ All RBAC granularity levels supported
5. ✅ Global admin + strict hub isolation
6. ✅ Explicit hub management permissions
7. ✅ Hybrid session cleanup strategy
8. ✅ Global users with RBAC-driven hub access

## Implementation Status

### ✅ **COMPLETED - Phase 1: Core Hub Infrastructure**
- **Hub Validation Middleware** - `hub_middleware.go` with hub access control ✅
- **Hub-scoped RBAC** - Permission checking for hub isolation ✅  
- **API Enhancement** - Hub-aware entity creation and queries ✅
- **Route Integration** - New hub endpoints added to main.go ✅

### ✅ **COMPLETED - Phase 2: Trait/Self System**
- **Tag Namespace Handling** - `hubname:self/trait:namespace:value` parsing ✅
- **API Enhancement** - Hub, self, traits in request/response ✅
- **Query Filters** - Trait/self filtering capabilities ✅
- **Hub Entity Creation** - Working hub-aware entity creation ✅

### 🔄 **IN PROGRESS - Phase 2 Debugging**
- **Hub Listing Bug Fix** - Minor issue with hub entity parsing (nearly complete)

### ⏳ **PENDING - Phase 3: Session Management**
- Session entities with hub association
- Session-user relationships  
- Multi-session support per user
- Session lifecycle management

### ⏳ **PENDING - Phase 4: Testing & Documentation**
- Comprehensive testing
- Performance validation
- API documentation updates

## Implemented Features

### **Multi-Hub API Endpoints** 🚀
```
POST   /api/v1/hubs/create              # Create new hub
GET    /api/v1/hubs/list                # List accessible hubs  
DELETE /api/v1/hubs/delete              # Delete hub (if empty)
POST   /api/v1/hubs/entities/create     # Create hub-aware entities
GET    /api/v1/hubs/entities/query      # Query hub entities
```

### **Hub-Aware Entity Format** ✨
```json
{
  "hub": "worca",
  "self": {"type": "task", "status": "todo", "title": "Test Task"},
  "traits": {"org": "TechCorp", "project": "Mobile"},
  "content": "Task description"
}
```

### **Tag Structure Implementation** 🏷️
```
hub:worca                          # Hub membership
worca:self:type:task              # Entity's own properties  
worca:self:status:todo
worca:trait:org:TechCorp          # Inherited context
worca:trait:project:Mobile
```

### **RBAC Permissions** 🔐
```
rbac:perm:entity:*:hub:worca      # Full hub access
rbac:perm:hub:create               # Can create hubs
rbac:perm:hub:manage:worca        # Can manage specific hub
```

## Testing Results

### ✅ **Successful Tests**
- **Hub Creation**: `POST /api/v1/hubs/create` ✅
- **Hub-Aware Entity Creation**: Entities with hub/self/traits ✅  
- **Hub Entity Queries**: Filtering by hub, self, traits ✅
- **RBAC Integration**: Permission checks working ✅
- **Build & Deploy**: Clean compilation and server restart ✅

### 🔧 **Minor Issue**
- **Hub Listing**: Returns inconsistent results (debugging in progress)

## Next Steps
1. ✅ Complete hub listing bug fix
2. 📝 Update Worca to use new hub-aware API
3. 🔄 Implement session management with hub scoping
4. 📋 Create comprehensive test suite
5. 📚 Update API documentation
6. 🚀 Git commit and deployment