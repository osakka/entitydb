# EntityDB Dataspace Management API

EntityDB now includes comprehensive dataspace management capabilities, allowing isolation of data into separate logical universes.

## Overview

Dataspaces in EntityDB provide:
- **Data Isolation**: Each dataspace maintains its own index file
- **Performance**: Per-dataspace indexes prevent global bottlenecks
- **Multi-tenancy**: Support for multiple applications/projects
- **RBAC Integration**: Full permission-based access control

## API Endpoints

### List Dataspaces
```
GET /api/v1/dataspaces
Authorization: Bearer <token>
```

Returns all configured dataspaces the user has permission to view.

### Get Dataspace
```
GET /api/v1/dataspaces/{id}
Authorization: Bearer <token>
```

Returns details for a specific dataspace.

### Create Dataspace
```
POST /api/v1/dataspaces
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "worca",
  "description": "Workforce Orchestrator Application",
  "settings": {
    "theme": "oceanic",
    "features": "kanban,projects,teams"
  }
}
```

Creates a new dataspace. Requires `dataspace:create` permission (admin only).

### Update Dataspace
```
PUT /api/v1/dataspaces/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "worca",
  "description": "Updated description",
  "settings": {
    "theme": "oceanic",
    "features": "kanban,projects,teams,analytics"
  }
}
```

Updates an existing dataspace. Requires `dataspace:update` permission (admin only).

### Delete Dataspace
```
DELETE /api/v1/dataspaces/{id}
Authorization: Bearer <token>
```

Deletes a dataspace. Cannot delete if entities exist in the dataspace.
Requires `dataspace:delete` permission (admin only).

## Permissions

- `dataspace:view` - View dataspace information
- `dataspace:create` - Create new dataspaces (admin only)
- `dataspace:update` - Update dataspace configuration (admin only)
- `dataspace:delete` - Delete dataspaces (admin only)

## Example Usage

```bash
# Login
TOKEN=$(curl -k -s -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# List dataspaces
curl -k -s -X GET https://localhost:8085/api/v1/dataspaces \
  -H "Authorization: Bearer $TOKEN" | jq .

# Create a new dataspace
curl -k -s -X POST https://localhost:8085/api/v1/dataspaces \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production",
    "description": "Production Environment",
    "settings": {
      "environment": "prod",
      "backup": "enabled"
    }
  }' | jq .
```

## Implementation Details

- Dataspaces are stored as entities with `type:dataspace` tag
- Each dataspace gets its own index file in `/dataspaces/` directory
- Dataspace isolation is enforced at the repository level
- WAL-only mode provides O(1) write performance per dataspace

## Current Status

✅ Dataspace CRUD operations implemented
✅ RBAC integration complete
✅ Per-dataspace index isolation
✅ API routes configured and tested

## Next Steps

- Add cross-dataspace query capabilities
- Implement dataspace migration tools
- Add dataspace-specific configuration options
- Create UI for dataspace management in dashboard