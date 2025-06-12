---
title: Core Concepts
category: Getting Started
tags: [concepts, entities, tags, temporal]
last_updated: 2025-06-11
version: v2.29.0
---

# EntityDB Core Concepts

Understanding these fundamental concepts will help you effectively use EntityDB.

## Entities

**Everything in EntityDB is an entity.** This is the core design principle.

An entity consists of:
- **ID**: Unique identifier (string, up to 64 characters)
- **Tags**: Array of key-value descriptors
- **Content**: Binary data (JSON, files, images, etc.)
- **Relationships**: Connections to other entities

```json
{
  "id": "user_john_doe",
  "tags": ["type:user", "name:John Doe", "department:engineering"],
  "content": "{\"profile\": \"Senior Engineer\", \"skills\": [\"Go\", \"React\"]}",
  "created_at": "2025-06-11T10:00:00Z",
  "updated_at": "2025-06-11T15:30:00Z"
}
```

### Entity Types
Common entity patterns:
- **Users**: `type:user`
- **Documents**: `type:document`
- **Configuration**: `type:config`
- **Relationships**: `type:relationship`

## Tags

Tags are the primary way to describe and query entities. They're stored as `key:value` pairs.

### Tag Structure
```
namespace:category:subcategory:value
```

Examples:
```
type:user
department:engineering:backend
status:active
priority:high
location:us:california:san_francisco
```

### Temporal Tags

> **Key Feature**: Every tag is automatically timestamped with nanosecond precision.

Internal format: `TIMESTAMP|tag`
```
1673510400123456789|type:user
1673510400987654321|status:active
```

The API returns tags without timestamps by default:
```json
{
  "tags": ["type:user", "status:active"]
}
```

To see timestamps, use `include_timestamps=true`:
```json
{
  "tags": [
    "2025-06-11T10:00:00.123456789Z|type:user",
    "2025-06-11T10:00:00.987654321Z|status:active"
  ]
}
```

### Tag Namespaces

| Namespace | Purpose | Examples |
|-----------|---------|----------|
| `type:` | Entity classification | `type:user`, `type:document` |
| `id:` | Additional identifiers | `id:employee:12345` |
| `status:` | State information | `status:active`, `status:archived` |
| `rbac:` | Security permissions | `rbac:role:admin`, `rbac:perm:entity:view` |
| `conf:` | Configuration | `conf:feature:enabled` |
| `meta:` | Metadata | `meta:version:1.0`, `meta:author:john` |

## Temporal Storage

EntityDB maintains complete history of all changes.

### Timeline Storage
- Every tag addition/removal is timestamped
- Content changes are versioned
- No data is ever deleted (only marked as archived)

### Temporal Queries

**As-of queries**: View data as it existed at a specific time
```bash
curl "http://localhost:8085/api/v1/entities/as-of?timestamp=2025-06-11T10:00:00Z&tags=type:user"
```

**History queries**: See all changes to an entity
```bash
curl "http://localhost:8085/api/v1/entities/history?id=user_john_doe"
```

**Diff queries**: Compare entity state between two points in time
```bash
curl "http://localhost:8085/api/v1/entities/diff?id=user_john_doe&from=2025-06-11T09:00:00Z&to=2025-06-11T15:00:00Z"
```

## Datasets

Datasets provide logical grouping and access control.

### Purpose
- **Organization**: Group related entities
- **Security**: Dataset-level RBAC
- **Performance**: Isolated indexing
- **Multi-tenancy**: Separate customer data

### Default Dataset
- All entities belong to the `default` dataset unless specified
- Admin users have access to all datasets

### Dataset Operations
```bash
# Create a dataset
curl -X POST http://localhost:8085/api/v1/datasets \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "hr_data", "description": "Human Resources entities"}'

# Create entity in specific dataset
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"id": "emp_123", "dataset": "hr_data", "tags": ["type:employee"]}'
```

## Relationships

Entities can be connected through typed relationships.

### Relationship Structure
```json
{
  "from_entity_id": "user_john_doe",
  "to_entity_id": "project_web_app",
  "relationship_type": "assigned_to",
  "created_at": "2025-06-11T10:00:00Z"
}
```

### Common Relationship Types
- `assigned_to`: User assigned to project
- `reports_to`: Organizational hierarchy
- `depends_on`: Project dependencies
- `belongs_to`: Group membership
- `created_by`: Authorship

### Querying Relationships
```bash
# Get all relationships for an entity
curl "http://localhost:8085/api/v1/entity-relationships?entity_id=user_john_doe"

# Get specific relationship type
curl "http://localhost:8085/api/v1/entity-relationships?from_entity_id=user_john_doe&relationship_type=assigned_to"
```

## RBAC (Role-Based Access Control)

EntityDB uses tag-based RBAC for security.

### Permission Tags
```
rbac:role:admin           # Administrative role
rbac:role:user            # Standard user role
rbac:perm:entity:view     # Can view entities
rbac:perm:entity:create   # Can create entities
rbac:perm:entity:update   # Can update entities
rbac:perm:dataset:manage  # Can manage datasets
```

### Permission Hierarchy
```
rbac:perm:*                    # All permissions
rbac:perm:entity:*             # All entity permissions
rbac:perm:entity:view          # Specific permission
```

### Enforcement
- All API endpoints check permissions
- Middleware validates token and permissions
- Dataset-level access control

## Binary Storage Format (EBF)

EntityDB uses a custom binary format for high performance.

### Features
- **Memory-mapped**: Zero-copy reads
- **Compressed**: Automatic compression for content > 1KB
- **Indexed**: B-tree timelines, skip-lists
- **Concurrent**: Safe for multiple readers/writers
- **WAL**: Write-Ahead Logging for durability

### Performance Benefits
- 100x faster than traditional databases
- Nanosecond timestamp precision
- Concurrent access without locks
- Streaming support for large files

## Autochunking

Large content is automatically split into chunks.

### Chunking Rules
- Files > 4MB are automatically chunked
- Chunks are 4MB each (configurable)
- Transparent reassembly on read
- Efficient streaming for large files

### Benefits
- No RAM limits for large files
- Efficient bandwidth usage
- Parallel chunk processing

## Configuration System

EntityDB uses a 3-tier configuration hierarchy:

1. **Database configuration** (highest priority)
2. **Command-line flags**
3. **Environment variables** (lowest priority)

### Configuration as Entities
- Configuration stored as entities with `type:config` tags
- Runtime updates via API
- Version history maintained

## Metrics and Observability

EntityDB provides comprehensive metrics:

### Metric Types
- **Counters**: Incrementing values (requests, errors)
- **Gauges**: Point-in-time values (memory usage, connections)
- **Histograms**: Distribution data (latencies, sizes)

### Collection
- Real-time metrics via temporal tags
- Configurable retention policies
- Automatic aggregation (1min, 1hour, daily)

### Endpoints
- `/health`: Health check with system metrics
- `/metrics`: Prometheus format
- `/api/v1/system/metrics`: Comprehensive EntityDB metrics

## Next Steps

Now that you understand the core concepts:

1. **[API Reference](../30-api-reference/)** - Learn the complete API
2. **[User Guides](../40-user-guides/)** - Common tasks and workflows
3. **[Architecture](../20-architecture/)** - Deep dive into system design
4. **[Security](../50-admin-guides/01-security-configuration.md)** - Secure your installation

## Common Patterns

### Hierarchical Data
```json
{
  "id": "org_acme",
  "tags": ["type:organization", "name:ACME Corp"]
},
{
  "id": "dept_engineering", 
  "tags": ["type:department", "name:Engineering", "parent:org_acme"]
},
{
  "id": "user_john",
  "tags": ["type:user", "name:John Doe", "department:dept_engineering"]
}
```

### Time-Series Data
```json
{
  "id": "metric_cpu_usage",
  "tags": ["type:metric", "host:server1", "value:85.2"],
  "content": "{\"timestamp\": \"2025-06-11T10:00:00Z\", \"cpu_percent\": 85.2}"
}
```

### Document Management
```json
{
  "id": "doc_user_manual",
  "tags": ["type:document", "category:manual", "version:2.1", "format:pdf"],
  "content": "<binary PDF data>"
}
```