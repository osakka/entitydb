# Dataset and Metrics API Reference

> **Version**: v2.31.0 | **Last Updated**: 2025-06-14 | **Status**: AUTHORITATIVE

This document covers EntityDB's dataset management and metrics API endpoints, providing comprehensive examples and implementation guidance for data organization and monitoring integration.

## Dataset Operations

EntityDB implements a dataset-based architecture for multi-tenancy and data organization. Datasets provide logical separation of data while maintaining unified access patterns.

### Dataset Concepts

#### What is a Dataset?
- **Logical Data Container**: Datasets provide namespace isolation for entities
- **Multi-tenancy Support**: Different applications or environments can use separate datasets
- **Unified API**: All entity operations work transparently across datasets
- **Configuration-Based**: Dataset selection via environment variables or headers

#### Default Dataset Behavior
```bash
# Default dataset configuration
ENTITYDB_DATASET_NAME="production"
ENTITYDB_DATASET_PATH="/opt/entitydb/var/db"

# All entity operations automatically use the configured dataset
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list"
```

### Dataset Management via Entity Operations

Since EntityDB treats everything as entities, dataset management is handled through the standard entity API with specific dataset tags.

#### Creating Dataset Configuration Entities
```bash
# Create dataset configuration entity
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:config",
      "conf:dataset",
      "id:dataset:development",
      "status:active"
    ],
    "content": "{\"name\":\"development\",\"path\":\"/opt/entitydb/var/db-dev\",\"description\":\"Development environment dataset\"}"
  }'
```

#### Listing Dataset Configurations
```bash
# List all dataset configurations
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?tags=type:config,conf:dataset"
```

#### Dataset Switching (Runtime Configuration)
```bash
# Update dataset configuration for runtime switching
curl -X PUT http://localhost:8085/api/v1/entities/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "dataset-config-entity-id",
    "tags": [
      "type:config",
      "conf:dataset",
      "conf:active",
      "id:dataset:current"
    ],
    "content": "{\"name\":\"staging\",\"path\":\"/opt/entitydb/var/db-staging\"}"
  }'
```

### Dataset-Specific Entity Operations

#### Cross-Dataset Entity Queries
```bash
# Query entities with dataset context
curl -H "Authorization: Bearer $TOKEN" \
  -H "X-EntityDB-Dataset: development" \
  "http://localhost:8085/api/v1/entities/list?tags=type:user"

# Alternative: Query dataset configuration entities
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?tags=conf:dataset,status:active"
```

#### Dataset Migration Operations
```bash
# Tag entities for dataset migration
curl -X PUT http://localhost:8085/api/v1/entities/update \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "entity-to-migrate",
    "tags": [
      "type:user",
      "id:username:john.doe",
      "dataset:source:production",
      "dataset:target:development",
      "migration:pending"
    ]
  }'
```

## Metrics API Reference

EntityDB provides comprehensive metrics through multiple endpoints designed for different monitoring needs and access levels.

### 1. Public Health Endpoint

#### GET /health
**Description**: Basic system health check (no authentication required)
**Access**: Public
**Use Case**: Load balancer health checks, basic monitoring

```bash
# Basic health check
curl http://localhost:8085/health
```

**Response Format:**
```json
{
  "status": "healthy",
  "timestamp": "2025-06-14T10:00:00Z",
  "version": "2.31.0",
  "uptime": "24h15m30s",
  "checks": {
    "database": "healthy",
    "memory": "healthy",
    "storage": "healthy",
    "wal": "healthy"
  }
}
```

**Health Status Values:**
- `healthy`: All systems operational
- `degraded`: Some subsystems experiencing issues
- `unhealthy`: Critical system failures

### 2. Prometheus Metrics Endpoint

#### GET /metrics
**Description**: Prometheus-compatible metrics (no authentication required)
**Access**: Public (typically restricted by firewall)
**Use Case**: Prometheus scraping, external monitoring systems

```bash
# Get Prometheus metrics
curl http://localhost:8085/metrics
```

**Sample Response:**
```
# HELP entitydb_entities_total Total number of entities
# TYPE entitydb_entities_total counter
entitydb_entities_total 1500

# HELP entitydb_http_requests_total Total HTTP requests
# TYPE entitydb_http_requests_total counter
entitydb_http_requests_total{method="GET",status="200"} 2500
entitydb_http_requests_total{method="POST",status="201"} 150

# HELP entitydb_memory_usage_bytes Current memory usage
# TYPE entitydb_memory_usage_bytes gauge
entitydb_memory_usage_bytes 536870912

# HELP entitydb_storage_size_bytes Storage utilization
# TYPE entitydb_storage_size_bytes gauge
entitydb_storage_size_bytes 1073741824

# HELP entitydb_tag_cache_hits_total Tag cache hits (v2.31.0)
# TYPE entitydb_tag_cache_hits_total counter
entitydb_tag_cache_hits_total 5000

# HELP entitydb_parallel_index_operations_total Parallel index operations (v2.31.0)
# TYPE entitydb_parallel_index_operations_total counter
entitydb_parallel_index_operations_total 150
```

### 3. System Metrics Endpoint

#### GET /api/v1/system/metrics
**Description**: Detailed EntityDB system metrics
**Access**: Requires authentication
**Permissions**: None (available to all authenticated users)

```bash
# Get detailed system metrics
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/system/metrics
```

**Response Structure:**
```json
{
  "status": "ok",
  "data": {
    "timestamp": "2025-06-14T10:00:00Z",
    "version": "2.31.0",
    "uptime_seconds": 87330,
    "system": {
      "memory_usage_mb": 512,
      "memory_allocated_mb": 256,
      "memory_limit_mb": 4096,
      "gc_cycles": 45,
      "goroutines": 25,
      "cpu_usage_percent": 5.2
    },
    "entities": {
      "total_count": 1500,
      "created_today": 25,
      "updated_today": 150,
      "growth_rate_daily": 1.8
    },
    "storage": {
      "database_size_mb": 1024,
      "wal_size_mb": 128,
      "index_size_mb": 256,
      "chunks_count": 500,
      "compression_ratio": 0.75
    },
    "performance": {
      "tags": {
        "cache_hits": 5000,
        "cache_misses": 250,
        "hit_rate": 0.952,
        "cache_size": 10000,
        "temporal_cache_hits": 1200
      },
      "json": {
        "encoder_pool_size": 100,
        "pool_hits": 2500,
        "encoding_time_avg_ms": 0.5
      },
      "indexing": {
        "parallel_workers": 4,
        "build_time_avg_ms": 15,
        "operations_total": 150
      },
      "batch": {
        "write_operations": 75,
        "avg_batch_size": 8,
        "timeout_count": 2
      }
    },
    "http": {
      "requests_total": 2650,
      "requests_per_second": 1.2,
      "response_time_avg_ms": 45,
      "active_connections": 8,
      "status_codes": {
        "200": 2400,
        "201": 150,
        "400": 50,
        "401": 25,
        "403": 15,
        "500": 10
      }
    }
  }
}
```

### 4. RBAC Metrics Endpoint

#### GET /api/v1/rbac/metrics
**Description**: Authentication and authorization analytics
**Access**: Requires authentication
**Permissions**: `rbac:perm:system:view` or admin role

```bash
# Get RBAC metrics (admin access required)
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/rbac/metrics
```

**Response Structure:**
```json
{
  "status": "ok",
  "data": {
    "timestamp": "2025-06-14T10:00:00Z",
    "sessions": {
      "active_count": 12,
      "total_created": 150,
      "expired_count": 125,
      "avg_duration_hours": 8.5
    },
    "authentication": {
      "login_attempts": 200,
      "successful_logins": 185,
      "failed_attempts": 15,
      "success_rate": 0.925,
      "unique_users": 25
    },
    "authorization": {
      "permission_checks": 5000,
      "denied_requests": 50,
      "granted_requests": 4950,
      "permission_cache_hits": 4500
    },
    "users": {
      "total_count": 25,
      "active_count": 20,
      "admin_count": 3,
      "last_login_distribution": {
        "last_24h": 15,
        "last_week": 20,
        "last_month": 23,
        "inactive": 2
      }
    }
  }
}
```

### 5. Application Metrics Endpoint

#### GET /api/v1/application/metrics
**Description**: Generic application metrics storage and retrieval
**Access**: Requires authentication
**Permissions**: `rbac:perm:metrics:read`

This endpoint allows external applications to store and retrieve custom metrics using EntityDB's temporal storage.

```bash
# Store application metrics
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:metric",
      "app:myapp",
      "metric:response_time",
      "namespace:production"
    ],
    "content": "{\"value\": 250, \"unit\": \"ms\", \"timestamp\": \"2025-06-14T10:00:00Z\"}"
  }'

# Retrieve application metrics
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?tags=type:metric,app:myapp"
```

**Application Metrics Query Parameters:**
- `app`: Filter by application name
- `namespace`: Filter by environment/namespace
- `metric_type`: Filter by metric type
- `time_range`: Filter by time range (requires temporal queries)

#### Application Metrics Examples

**Custom Application Metrics Storage:**
```bash
# Store business metrics
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:metric",
      "app:ecommerce",
      "metric:sales_total",
      "namespace:production",
      "period:daily"
    ],
    "content": "{\"value\": 15000, \"currency\": \"USD\", \"date\": \"2025-06-14\"}"
  }'

# Store technical metrics
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:metric",
      "app:monitoring",
      "metric:cpu_usage",
      "host:server01",
      "namespace:production"
    ],
    "content": "{\"value\": 75.5, \"unit\": \"percent\", \"timestamp\": \"2025-06-14T10:00:00Z\"}"
  }'
```

## Advanced Metrics Operations

### 1. Temporal Metrics Queries

#### Historical Metrics Analysis
```bash
# Get metrics history for last 24 hours
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/history?id=metric-entity-id&duration=24h"

# Get metrics as of specific time
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/as-of?timestamp=2025-06-13T10:00:00Z&tags=type:metric,app:myapp"
```

#### Metrics Comparison
```bash
# Compare metrics between time periods
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/diff?id=metric-entity-id&from=2025-06-13T10:00:00Z&to=2025-06-14T10:00:00Z"
```

### 2. Aggregated Metrics

#### Daily Aggregation
```bash
# Create daily aggregated metrics
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": [
      "type:metric",
      "aggregation:daily",
      "source:response_times",
      "date:2025-06-14"
    ],
    "content": "{\"min\": 10, \"max\": 500, \"avg\": 125, \"count\": 2500}"
  }'
```

### 3. Custom Metrics Dashboard Integration

#### Metrics for Grafana
```bash
# Query metrics for Grafana dashboard
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/query" \
  -G \
  -d "tags=type:metric,app:myapp" \
  -d "sort_by=timestamp" \
  -d "sort_order=desc" \
  -d "limit=100"
```

#### Real-time Metrics Stream
```bash
# Get latest metrics for real-time dashboard
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8085/api/v1/entities/list?tags=type:metric&sort_by=timestamp&sort_order=desc&limit=50"
```

## Performance Optimization (v2.31.0)

### Metrics-Specific Optimizations

#### Tag Cache Monitoring
Monitor the effectiveness of O(1) tag value caching for metrics queries:
```bash
# Check metrics-specific cache performance
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8085/api/v1/system/metrics | \
  jq '.data.performance.tags | {
    total_operations: (.cache_hits + .cache_misses),
    hit_rate: (.cache_hits / (.cache_hits + .cache_misses)),
    cache_efficiency: (if (.cache_hits / (.cache_hits + .cache_misses)) > 0.9 then "excellent" elif (.cache_hits / (.cache_hits + .cache_misses)) > 0.8 then "good" else "needs_optimization" end)
  }'
```

#### Batch Metrics Operations
Use batch operations for high-volume metrics:
```bash
# Enable batch write optimization for metrics
ENTITYDB_BATCH_WRITE_SIZE=10
ENTITYDB_BATCH_WRITE_TIMEOUT=100ms
```

## Error Handling

### Common Error Responses

#### Authentication Errors
```json
{
  "status": "error",
  "error": "authentication_required",
  "message": "Valid authentication token required",
  "code": 401
}
```

#### Permission Errors
```json
{
  "status": "error",
  "error": "insufficient_permissions",
  "message": "Missing required permission: rbac:perm:metrics:read",
  "code": 403
}
```

#### Invalid Metrics Format
```json
{
  "status": "error",
  "error": "invalid_metric_format",
  "message": "Metric content must be valid JSON",
  "code": 400
}
```

## Best Practices

### 1. Metrics Organization
- Use consistent tag namespaces (`app:`, `metric:`, `namespace:`)
- Include timestamp and unit information in metric content
- Use descriptive metric names and avoid abbreviations
- Group related metrics with common tags

### 2. Performance Considerations
- Leverage tag caching by using consistent tag patterns
- Use batch operations for high-volume metrics
- Implement retention policies for historical metrics
- Monitor cache hit rates and optimize tag usage

### 3. Security
- Restrict metrics access with appropriate RBAC permissions
- Use separate datasets for different environments
- Sanitize metric content to prevent injection attacks
- Monitor authentication metrics for security events

---

*This API reference covers EntityDB's dataset and metrics capabilities in v2.31.0. For additional monitoring setup, see [Monitoring Guide](../50-admin-guides/02-monitoring-guide.md).*