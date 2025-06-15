# EntityDB Query API Documentation

## Overview

EntityDB provides powerful query capabilities for searching and filtering entities based on tags, content, and other criteria. All queries are performed via the entity list endpoint with various query parameters.

## Query Types

### 1. Basic Tag Filtering

Filter entities by exact tag match:

```bash
# Find all issue entities
entitydb-cli entity list --tag="type:issue"

# Find entities with high priority
entitydb-cli entity list --tag="priority:high"

# API endpoint
GET /api/v1/entities/list?tag=type:issue
```

### 2. Wildcard Tag Matching

Use wildcards to match tag patterns:

```bash
# Find all entities with any type
entitydb-cli entity list --wildcard="type:*"

# Find all RBAC permission entities
entitydb-cli entity list --wildcard="rbac:perm:*"

# Find all entities in the meta namespace
entitydb-cli entity list --wildcard="meta:*"

# API endpoint
GET /api/v1/entities/list?wildcard=rbac:perm:*
```

### 3. Content Search

Search for text within entity content:

```bash
# Search for "login" in any content
entitydb-cli entity list --search="login"

# Search for "authentication" in descriptions
entitydb-cli entity list --search="authentication" --content-type="description"

# API endpoint
GET /api/v1/entities/list?search=login
GET /api/v1/entities/list?search=authentication&contentType=description
```

### 4. Namespace Filtering

Filter entities by tag namespace:

```bash
# Find all entities with tags in the "rbac" namespace
entitydb-cli entity list --namespace="rbac"

# Find all entities with tags in the "type" namespace
entitydb-cli entity list --namespace="type"

# API endpoint
GET /api/v1/entities/list?namespace=rbac
```

## Query Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `tag` | Exact tag match | `type:issue` |
| `wildcard` | Tag pattern with wildcards | `rbac:perm:*` |
| `search` | Search text in content | `authentication` |
| `contentType` | Content type for search | `description` |
| `namespace` | Tag namespace filter | `rbac` |
| `limit` | Number of results | `20` |
| `offset` | Pagination offset | `0` |

## Binary Format Implementation

The queries use the custom binary format (EBF) with sharded indexing for efficient searching:

- Tag filtering: Sharded index lookups with O(1) performance within shards
- Content search: Direct binary format scanning with temporal tag support
- Namespace filtering: Efficient prefix pattern matching with bloom filters

## Examples

### Complex Queries

Find all user entities with admin role:
```bash
entitydb-cli entity list --wildcard="type:user" --tag="rbac:role:admin"
```

Search for issues containing "bug":
```bash
entitydb-cli entity list --wildcard="type:issue" --search="bug"
```

Find all permissions in entity namespace:
```bash
entitydb-cli entity list --wildcard="rbac:perm:entity:*"
```

### API Usage

```bash
# Using curl
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?wildcard=type:*&search=login"

# Using the Python client
from entitydb_client import EntityDBClient

client = EntityDBClient("http://localhost:8085", token)
entities = client.list_entities(wildcard="type:*", search="login")
```

## Performance Considerations

1. **Binary format queries**: All queries use the custom binary format with sharded indexing
2. **Index usage**: Sharded tag indexes provide O(1) lookups within shards  
3. **Wildcard patterns**: Use specific patterns when possible (e.g., `type:issue` instead of `*:issue`)
4. **Content search**: Direct binary scanning with bloom filter optimization

## Advanced Query API (Future)

The `QueryAdvanced` method supports complex conditions:

```go
conditions := map[string]interface{}{
    "tags": []string{"type:issue", "status:active"},
    "content": "authentication",
    "contentType": "description",
}
entities, err := repo.QueryAdvanced(conditions)
```

This is not yet exposed via the API but provides a foundation for future complex query capabilities.