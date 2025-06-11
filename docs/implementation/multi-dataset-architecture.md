# Multi-Dataset Architecture Implementation Plan

## Overview
Implementation of multi-database (dataset) system with tag-based inheritance and enhanced session management for EntityDB.

## Design Decisions Made

### 1. Permission Evaluation Strategy ✅
**Choice**: Simple tag matching (Option A)
- `rbac:perm:entity:write:dataset:worca` = can write entities with `dataset:worca`
- Natural fit with EntityDB's existing tag system
- High performance, minimal changes to RBAC middleware

### 2. Session Management Strategy ✅
**Choice**: Sessions as entities + relationships (Options A + C)
- Session entities: `type:session`, `dataset:worca`, `status:active`
- User-session relationships via EntityDB relationship system
- Multi-session support per user
- Rich session metadata and temporal tracking

### 3. Dataset + Trait/Self Architecture ✅
**Choice**: Dataset namespace with trait/self separation
```
dataset:worca                          # Dataset membership
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

### Dataset System
- **Dataset Tag**: `dataset:worca` (mandatory on all entities)
- **Self Namespace**: `worca:self:namespace:value` (entity's own properties)
- **Trait Namespace**: `worca:trait:namespace:value` (inherited context)

### Session Management
- **Session Entities**: `type:session`, `dataset:worca`, `status:active/expired`
- **Session-User Relationships**: Link users to their active sessions
- **Multi-session Support**: Users can have multiple concurrent sessions
- **Dataset-scoped Sessions**: Sessions belong to specific datasets

### RBAC Enhancement
- **Dataset-scoped Permissions**: `rbac:perm:entity:write:dataset:worca`
- **Trait-based Permissions**: `rbac:perm:entity:read:worca:trait:org:TechCorp`
- **Self-based Permissions**: `rbac:perm:entity:update:worca:self:assignee:self`

## Implementation Steps

### Phase 1: Core Dataset Infrastructure
1. **Dataset Validation Middleware** - Ensure all entities have `dataset:` tag
2. **Dataset-scoped Queries** - Modify query handlers to filter by dataset
3. **API Endpoints** - Add dataset parameter to create/update operations
4. **RBAC Integration** - Update permission checking for dataset-scoped access

### Phase 2: Trait/Self System
1. **Tag Namespace Handling** - Parse and validate `datasetname:self/trait:namespace:value`
2. **API Enhancement** - Support dataset, self, traits in request/response
3. **Query Filters** - Add trait/self filtering capabilities
4. **Documentation** - API docs and examples

### Phase 3: Session Management
1. **Session Entities** - Create session entity type with dataset association
2. **Session-User Relationships** - Link sessions to users
3. **Session Lifecycle** - Login/logout session state management
4. **Multi-session Support** - Allow multiple concurrent sessions per user

### Phase 4: Testing & Documentation
1. **Unit Tests** - Dataset, trait/self, session functionality
2. **Integration Tests** - End-to-end multi-dataset scenarios
3. **Performance Tests** - Dataset-scoped query performance
4. **Documentation** - Complete API documentation and examples

## Design Decisions Continued

### 4. RBAC Granularity Level ✅
**Choice**: All levels supported (Option D)
- **Dataset-level**: `rbac:perm:entity:write:dataset:worca`
- **Trait-level**: `rbac:perm:entity:write:worca:trait:org:TechCorp` 
- **Self-level**: `rbac:perm:entity:write:worca:self:type:task`
- **Maximum flexibility**: Users can have any combination

### 5. Cross-Dataset Access ✅
**Choice**: Global admin override + Ultra strict isolation (Options C + A)
- **Global Admin**: `rbac:perm:*` can access all datasets (natural part of tagging system)
- **Regular Users**: Strict dataset isolation - can only access their assigned dataset
- **Security**: No accidental cross-dataset data leaks
- **Administration**: Global admin can manage all datasets

### 6. Dataset Creation Permissions ✅
**Choice**: Explicit dataset permissions via RBAC tags
- **Dataset Management Permissions**:
  - `rbac:perm:dataset:create` - Can create new datasets
  - `rbac:perm:dataset:delete` - Can delete datasets
  - `rbac:perm:dataset:manage:worca` - Can manage specific dataset
  - `rbac:perm:dataset:assign-admin` - Can assign dataset admins
  - `rbac:perm:dataset:*` - All dataset management permissions
- **Natural Permission Flow**: Global admin has all, dataset creators get create, dataset admins get manage
- **Granular Control**: Dataset-specific management permissions

### 7. Session Cleanup Strategy ✅
**Choice**: Hybrid approach (Option C)
- **Auto-expiration**: Sessions have `expires_at:timestamp` tag for security
- **Activity Extension**: Active sessions get `expires_at` updated on API calls
- **Manual Logout**: Users can explicitly set `status:expired` 
- **Grace Period**: Configurable session timeout (default: 24h, extendable to 7 days max)
- **Cleanup Process**: Background task to mark expired sessions as `status:expired`

### 8. User Identity Scope ✅
**Choice**: Users are global, dataset access via RBAC (Option A)
- **Global Identity**: `username:john` unique system-wide
- **Dataset Access**: Controlled entirely by RBAC permissions
- **Multi-dataset Users**: Users get dataset access via `rbac:perm:*:dataset:*` tags
- **Session Flexibility**: Sessions can be dataset-specific or multi-dataset based on permissions
- **Natural Scaling**: Add more dataset permissions to expand user access

## Complete Design Summary

All design decisions finalized! Ready for implementation:
1. ✅ Simple tag matching for permissions
2. ✅ Sessions as entities + relationships  
3. ✅ Dataset + trait/self architecture
4. ✅ All RBAC granularity levels supported
5. ✅ Global admin + strict dataset isolation
6. ✅ Explicit dataset management permissions
7. ✅ Hybrid session cleanup strategy
8. ✅ Global users with RBAC-driven dataset access

## Implementation Status

### ✅ **COMPLETED - Phase 1: Core Dataset Infrastructure**
- **Dataset Validation Middleware** - `dataset_middleware.go` with dataset access control ✅
- **Dataset-scoped RBAC** - Permission checking for dataset isolation ✅  
- **API Enhancement** - Dataset-aware entity creation and queries ✅
- **Route Integration** - New dataset endpoints added to main.go ✅

### ✅ **COMPLETED - Phase 2: Trait/Self System**
- **Tag Namespace Handling** - `datasetname:self/trait:namespace:value` parsing ✅
- **API Enhancement** - Dataset, self, traits in request/response ✅
- **Query Filters** - Trait/self filtering capabilities ✅
- **Dataset Entity Creation** - Working dataset-aware entity creation ✅

### 🔄 **IN PROGRESS - Phase 2 Debugging**
- **Dataset Listing Bug Fix** - Minor issue with dataset entity parsing (nearly complete)

### ⏳ **PENDING - Phase 3: Session Management**
- Session entities with dataset association
- Session-user relationships  
- Multi-session support per user
- Session lifecycle management

### ⏳ **PENDING - Phase 4: Testing & Documentation**
- Comprehensive testing
- Performance validation
- API documentation updates

## Implemented Features

### **Multi-Dataset API Endpoints** 🚀
```
POST   /api/v1/datasets/create              # Create new dataset
GET    /api/v1/datasets/list                # List accessible datasets  
DELETE /api/v1/datasets/delete              # Delete dataset (if empty)
POST   /api/v1/datasets/entities/create     # Create dataset-aware entities
GET    /api/v1/datasets/entities/query      # Query dataset entities
```

### **Dataset-Aware Entity Format** ✨
```json
{
  "dataset": "worca",
  "self": {"type": "task", "status": "todo", "title": "Test Task"},
  "traits": {"org": "TechCorp", "project": "Mobile"},
  "content": "Task description"
}
```

### **Tag Structure Implementation** 🏷️
```
dataset:worca                          # Dataset membership
worca:self:type:task              # Entity's own properties  
worca:self:status:todo
worca:trait:org:TechCorp          # Inherited context
worca:trait:project:Mobile
```

### **RBAC Permissions** 🔐
```
rbac:perm:entity:*:dataset:worca      # Full dataset access
rbac:perm:dataset:create               # Can create datasets
rbac:perm:dataset:manage:worca        # Can manage specific dataset
```

## Testing Results

### ✅ **Successful Tests**
- **Dataset Creation**: `POST /api/v1/datasets/create` ✅
- **Dataset-Aware Entity Creation**: Entities with dataset/self/traits ✅  
- **Dataset Entity Queries**: Filtering by dataset, self, traits ✅
- **RBAC Integration**: Permission checks working ✅
- **Build & Deploy**: Clean compilation and server restart ✅

### 🔧 **Minor Issue**
- **Dataset Listing**: Returns inconsistent results (debugging in progress)

## Next Steps
1. ✅ Complete dataset listing bug fix
2. 📝 Update Worca to use new dataset-aware API
3. 🔄 Implement session management with dataset scoping
4. 📋 Create comprehensive test suite
5. 📚 Update API documentation
6. 🚀 Git commit and deployment