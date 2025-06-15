# EntityDB Dataset Management API

> **Version**: v2.32.0-dev | **Last Updated**: 2025-06-15 | **Status**: 100% ACCURATE
> 
> Complete dataset management API documentation - verified against actual implementation.

Dataset management in EntityDB provides multi-tenant data organization, allowing logical separation of entities while maintaining unified access patterns.

## Table of Contents
1. [Dataset Overview](#dataset-overview)
2. [Dataset Operations](#dataset-operations)
3. [Dataset Entity Operations](#dataset-entity-operations)
4. [Permission System](#permission-system)
5. [Examples](#examples)

## Dataset Overview

Datasets in EntityDB provide:
- **Logical Separation**: Organize entities by project, tenant, or environment
- **Scoped Queries**: Filter operations to specific datasets
- **Access Control**: Dataset-level RBAC permissions
- **Metadata Management**: Dataset-level configuration and tagging

### Base URL
```
https://localhost:8085/api/v1/datasets
```

### Authentication
All dataset endpoints require authentication and appropriate permissions.

**Required Header:**
```
Authorization: Bearer <session-token>
```

## Dataset Operations

### GET /api/v1/datasets

List all datasets accessible to the current user.

**Required Permission**: `dataset:view`

**Request:**
```bash
curl -k -X GET "https://localhost:8085/api/v1/datasets" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**
- `include_stats` - Include entity counts and size statistics (default: false)
- `sort` - Sort by (`name`, `created_at`, `updated_at`)
- `order` - Sort order (`asc`, `desc`)

**Response** (200 OK):
```json
[
  {
    "id": "prod_entitydb",
    "name": "EntityDB Production Dataset",
    "description": "Production data for EntityDB platform",
    "tags": ["env:production", "type:primary"],
    "entity_count": 15420,
    "created_at": 1748544372255000000,
    "updated_at": 1748544372255000000
  },
  {
    "id": "dev_testing",
    "name": "Development Testing",
    "description": "Development and testing environment",
    "tags": ["env:development", "type:testing"],
    "entity_count": 847,
    "created_at": 1748544372255000000,
    "updated_at": 1748544372255000000
  }
]
```

### POST /api/v1/datasets

Create a new dataset.

**Required Permission**: `dataset:create`

**Request:**
```bash
curl -k -X POST https://localhost:8085/api/v1/datasets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "project_alpha",
    "name": "Project Alpha Dataset",
    "description": "Dataset for Project Alpha development",
    "tags": ["env:development", "project:alpha", "team:backend"]
  }'
```

**Request Body:**
```json
{
  "id": "project_alpha",
  "name": "Project Alpha Dataset", 
  "description": "Dataset for Project Alpha development",
  "tags": ["env:development", "project:alpha", "team:backend"]
}
```

**Response** (201 Created):
```json
{
  "id": "project_alpha",
  "name": "Project Alpha Dataset",
  "description": "Dataset for Project Alpha development", 
  "tags": ["env:development", "project:alpha", "team:backend"],
  "entity_count": 0,
  "created_at": 1748544372255000000,
  "updated_at": 1748544372255000000
}
```

### GET /api/v1/datasets/{id}

Retrieve a specific dataset by ID.

**Required Permission**: `dataset:view`

**Request:**
```bash
curl -k -X GET "https://localhost:8085/api/v1/datasets/project_alpha" \
  -H "Authorization: Bearer $TOKEN"
```

**Path Parameters:**
- `id` (required) - Dataset identifier

**Query Parameters:**
- `include_entities` - Include recent entity list (default: false)
- `entity_limit` - Limit entity results (default: 10, max: 100)

**Response** (200 OK):
```json
{
  "id": "project_alpha",
  "name": "Project Alpha Dataset",
  "description": "Dataset for Project Alpha development",
  "tags": ["env:development", "project:alpha", "team:backend"],
  "entity_count": 156,
  "storage_size": 2048576,
  "created_at": 1748544372255000000,
  "updated_at": 1748544372255000000,
  "recent_entities": [
    {
      "id": "user_alice_001",
      "tags": ["type:user", "team:backend"],
      "created_at": 1748544372255000000
    }
  ]
}
```

### PUT /api/v1/datasets/{id}

Update an existing dataset's metadata.

**Required Permission**: `dataset:update`

**Request:**
```bash
curl -k -X PUT "https://localhost:8085/api/v1/datasets/project_alpha" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Project Alpha Production Dataset",
    "description": "Promoted to production environment",
    "tags": ["env:production", "project:alpha", "team:backend"]
  }'
```

**Request Body:**
```json
{
  "name": "Project Alpha Production Dataset",
  "description": "Promoted to production environment", 
  "tags": ["env:production", "project:alpha", "team:backend"]
}
```

**Response** (200 OK):
```json
{
  "id": "project_alpha",
  "name": "Project Alpha Production Dataset",
  "description": "Promoted to production environment",
  "tags": ["env:production", "project:alpha", "team:backend"],
  "entity_count": 156,
  "created_at": 1748544372255000000,
  "updated_at": 1748544372300000000
}
```

### DELETE /api/v1/datasets/{id}

Delete a dataset and all its entities.

**Required Permission**: `dataset:delete`

**⚠️ WARNING**: This operation permanently deletes all entities in the dataset.

**Request:**
```bash
curl -k -X DELETE "https://localhost:8085/api/v1/datasets/project_alpha" \
  -H "Authorization: Bearer $TOKEN"
```

**Query Parameters:**
- `confirm` (required) - Must be "true" to confirm deletion
- `backup` - Create backup before deletion (default: false)

**Request with confirmation:**
```bash
curl -k -X DELETE "https://localhost:8085/api/v1/datasets/project_alpha?confirm=true" \
  -H "Authorization: Bearer $TOKEN"
```

**Response** (200 OK):
```json
{
  "message": "Dataset deleted successfully",
  "id": "project_alpha",
  "entities_deleted": 156,
  "backup_created": false
}
```

## Dataset Entity Operations

### POST /api/v1/datasets/{dataset}/entities/create

Create an entity within a specific dataset.

**Required Permission**: `entity:create`

**Request:**
```bash
curl -k -X POST https://localhost:8085/api/v1/datasets/project_alpha/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user_alice_001",
    "tags": ["type:user", "role:developer", "team:backend"],
    "content": "eyJ1c2VybmFtZSI6ICJhbGljZSIsICJlbWFpbCI6ICJhbGljZUBleGFtcGxlLmNvbSJ9"
  }'
```

**Request Body:**
```json
{
  "id": "user_alice_001",
  "tags": ["type:user", "role:developer", "team:backend"],
  "content": "eyJ1c2VybmFtZSI6ICJhbGljZSIsICJlbWFpbCI6ICJhbGljZUBleGFtcGxlLmNvbSJ9"
}
```

**Response** (201 Created):
```json
{
  "id": "user_alice_001", 
  "dataset": "project_alpha",
  "tags": ["type:user", "role:developer", "team:backend"],
  "content": "eyJ1c2VybmFtZSI6ICJhbGljZSIsICJlbWFpbCI6ICJhbGljZUBleGFtcGxlLmNvbSJ9",
  "created_at": 1748544372255000000,
  "updated_at": 1748544372255000000
}
```

### GET /api/v1/datasets/{dataset}/entities/query

Query entities within a specific dataset.

**Required Permission**: `entity:view`

**Request:**
```bash
curl -k -X GET "https://localhost:8085/api/v1/datasets/project_alpha/entities/query?tag=type:user&sort=created_at&order=desc" \
  -H "Authorization: Bearer $TOKEN"
```

**Path Parameters:**
- `dataset` (required) - Dataset identifier

**Query Parameters:**
- `tag` - Filter by specific tag
- `tags` - Filter by multiple tags (comma-separated)
- `namespace` - Filter by tag namespace
- `search` - Search in content
- `sort` - Sort field (`created_at`, `updated_at`, `id`)
- `order` - Sort order (`asc`, `desc`)
- `limit` - Maximum results (default: 100, max: 1000)
- `offset` - Skip results for pagination
- `include_timestamps` - Include temporal timestamps in tags

**Response** (200 OK):
```json
{
  "dataset": "project_alpha",
  "total": 24,
  "offset": 0,
  "limit": 100,
  "entities": [
    {
      "id": "user_alice_001",
      "tags": ["type:user", "role:developer", "team:backend"],
      "content": "eyJ1c2VybmFtZSI6ICJhbGljZSIsICJlbWFpbCI6ICJhbGljZUBleGFtcGxlLmNvbSJ9",
      "created_at": 1748544372255000000,
      "updated_at": 1748544372255000000
    },
    {
      "id": "user_bob_002", 
      "tags": ["type:user", "role:designer", "team:frontend"],
      "content": "eyJ1c2VybmFtZSI6ICJib2IiLCAiZW1haWwiOiAiYm9iQGV4YW1wbGUuY29tIn0=",
      "created_at": 1748544372240000000,
      "updated_at": 1748544372240000000
    }
  ]
}
```

## Permission System

Dataset operations use hierarchical RBAC permissions:

### Dataset Permissions

| Operation | Required Permission |
|-----------|-------------------|
| List datasets | `dataset:view` |
| View dataset details | `dataset:view` |
| Create datasets | `dataset:create` |
| Update datasets | `dataset:update` |
| Delete datasets | `dataset:delete` |
| Dataset administration | `dataset:admin` |

### Entity Permissions within Datasets

Entity operations within datasets require both dataset access and entity permissions:

| Operation | Required Permissions |
|-----------|-------------------|
| Create entities in dataset | `dataset:view` + `entity:create` |
| Query entities in dataset | `dataset:view` + `entity:view` |
| Update entities in dataset | `dataset:view` + `entity:update` |

### Permission Wildcards

- `dataset:*` - All dataset operations
- `rbac:perm:*` - All permissions (admin only)

### Role Assignments

**Admin Role** (`rbac:role:admin`):
- Full access to all datasets
- Can create, update, and delete any dataset
- Can manage dataset permissions

**Dataset Manager Role** (`rbac:role:dataset_manager`):
- Create and manage datasets  
- Assign dataset permissions to users
- View dataset analytics

**User Role** (`rbac:role:user`):
- View assigned datasets
- Create entities in permitted datasets
- Limited to specific dataset access

## Examples

### Complete Dataset Workflow

```bash
# 1. Login and store token
TOKEN=$(curl -s -k -X POST https://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  | jq -r .token)

# 2. Create a new dataset
curl -k -X POST https://localhost:8085/api/v1/datasets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "customer_data",
    "name": "Customer Data Platform",
    "description": "Customer relationship management dataset",
    "tags": ["env:production", "type:crm", "compliance:gdpr"]
  }'

# 3. Create entities in the dataset
curl -k -X POST https://localhost:8085/api/v1/datasets/customer_data/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "customer_acme_corp",
    "tags": ["type:customer", "tier:enterprise", "region:us_west"],
    "content": "eyJuYW1lIjogIkFjbWUgQ29ycCIsICJlbWFpbCI6ICJjb250YWN0QGFjbWUuY29tIn0="
  }'

# 4. Query entities in the dataset
curl -k -X GET "https://localhost:8085/api/v1/datasets/customer_data/entities/query?tag=type:customer&sort=created_at&order=desc" \
  -H "Authorization: Bearer $TOKEN"

# 5. Get dataset statistics
curl -k -X GET "https://localhost:8085/api/v1/datasets/customer_data?include_entities=true&entity_limit=5" \
  -H "Authorization: Bearer $TOKEN"

# 6. Update dataset metadata
curl -k -X PUT "https://localhost:8085/api/v1/datasets/customer_data" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Customer Data Platform v2",
    "description": "Enhanced CRM with GDPR compliance",
    "tags": ["env:production", "type:crm", "compliance:gdpr", "version:2.0"]
  }'

# 7. List all datasets with statistics
curl -k -X GET "https://localhost:8085/api/v1/datasets?include_stats=true&sort=name&order=asc" \
  -H "Authorization: Bearer $TOKEN"
```

### Multi-Dataset Entity Management

```bash
# Create entities in different datasets
curl -k -X POST https://localhost:8085/api/v1/datasets/development/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test_user_001",
    "tags": ["type:user", "env:test", "role:tester"],
    "content": "eyJ1c2VybmFtZSI6ICJ0ZXN0ZXIxIiwgImVtYWlsIjogInRlc3RlcjFAZXhhbXBsZS5jb20ifQ=="
  }'

curl -k -X POST https://localhost:8085/api/v1/datasets/production/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "prod_user_001",
    "tags": ["type:user", "env:production", "role:customer"],
    "content": "eyJ1c2VybmFtZSI6ICJqb2huZG9lIiwgImVtYWlsIjogImpvaG5AZXhhbXBsZS5jb20ifQ=="
  }'

# Query across datasets (requires appropriate permissions)
curl -k -X GET "https://localhost:8085/api/v1/entities/list?tag=type:user" \
  -H "Authorization: Bearer $TOKEN"
```

### Dataset Migration

```bash
# 1. Create target dataset
curl -k -X POST https://localhost:8085/api/v1/datasets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "archive_2024",
    "name": "2024 Archive Dataset",
    "description": "Archived data from 2024",
    "tags": ["env:archive", "year:2024", "type:historical"]
  }'

# 2. Query entities to migrate from source dataset
ENTITIES=$(curl -s -k -X GET "https://localhost:8085/api/v1/datasets/old_dataset/entities/query?tag=status:archived" \
  -H "Authorization: Bearer $TOKEN")

# 3. Recreate entities in target dataset
# (Process entities JSON and create in archive_2024 dataset)

# 4. Verify migration completed
curl -k -X GET "https://localhost:8085/api/v1/datasets/archive_2024?include_stats=true" \
  -H "Authorization: Bearer $TOKEN"
```

## Error Handling

Dataset API endpoints return consistent error responses:

**Common HTTP Status Codes:**
- `200` - Success
- `201` - Created
- `400` - Bad Request (invalid dataset ID, missing parameters)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient dataset permissions)
- `404` - Not Found (dataset not found)
- `409` - Conflict (dataset ID already exists)
- `500` - Internal Server Error

**Error Examples:**

```json
{
  "error": "Dataset not found: invalid_dataset_id"
}
```

```json
{
  "error": "Permission denied: dataset:create required"
}
```

```json
{
  "error": "Dataset ID already exists: customer_data"
}
```

```json
{
  "error": "Cannot delete dataset: confirmation required"
}
```

## Related Documentation

- [Entity API](./03-entities.md) - Core entity operations
- [Authentication API](./02-authentication.md) - Login and session management
- [RBAC System](../architecture/03-rbac-architecture.md) - Access control details
- [Getting Started](../getting-started/02-quick-start.md) - Quick start guide

## Version History

- **v2.32.0-dev**: Current dataset API with unified sharded indexing
- **v2.29.0**: Dataset terminology migration from "dataspace"
- **v2.28.0**: Enhanced dataset management with comprehensive metrics