# EntityDB Metrics Implementation Action Plan

## Overview
This action plan addresses the critical gaps identified in the metrics audit to ensure EntityDB has comprehensive monitoring for healthy service operation.

## Phase 1: Critical Operational Metrics (Immediate)

### 1.1 Request/Response Metrics Handler
**File**: `src/api/request_metrics_middleware.go`
- Track all HTTP requests with timing
- Collect: method, path, status code, duration, request/response size
- Store as temporal metrics: `http_request_duration`, `http_request_count`, `http_response_size`

### 1.2 Error Tracking System
**File**: `src/api/error_metrics.go`
- Categorize errors: client (4xx), server (5xx), timeout, validation
- Track error rates by endpoint
- Store as: `http_errors_total` with labels for type and endpoint

### 1.3 WAL Checkpoint Metrics
**File**: `src/storage/binary/wal_metrics.go`
- Hook into existing checkpoint mechanism
- Track: checkpoint count, duration, size reduced, failures
- Metrics: `wal_checkpoint_duration`, `wal_checkpoint_count`, `wal_size_before`, `wal_size_after`

### 1.4 UI Chart Improvements
**File**: `share/htdocs/js/realtime-charts.js`
- Add legends to all charts
- Convert bytes to human-readable units (KB, MB, GB)
- Add value tooltips on hover
- Include units in axis labels

## Phase 2: Performance Monitoring

### 2.1 Query Performance Metrics
**File**: `src/models/query_metrics.go`
- Wrap repository methods to track timing
- Collect: query type, duration, result count, index usage
- Metrics: `query_duration` (with operation label), `query_result_count`

### 2.2 Write Performance Tracking
**File**: `src/storage/binary/write_metrics.go`
- Track entity creation/update performance
- Monitor write queue depth
- Metrics: `write_duration`, `write_queue_depth`, `writes_per_second`

### 2.3 Cache Performance
**File**: `src/cache/cache_metrics.go`
- Implement cache hit/miss tracking
- Monitor cache memory usage
- Metrics: `cache_hits`, `cache_misses`, `cache_evictions`, `cache_memory_bytes`

## Phase 3: System Health Monitoring

### 3.1 Resource Metrics Collector
**File**: `src/api/resource_metrics_collector.go`
- CPU usage (using runtime and syscall)
- Disk I/O statistics
- Network statistics
- File descriptor usage

### 3.2 Connection Pool Metrics
**File**: `src/api/connection_metrics.go`
- Track active connections
- Monitor connection lifecycle
- Metrics: `connections_active`, `connections_idle`, `connection_errors`

### 3.3 Health Score Calculator
**File**: `src/api/health_score.go`
- Aggregate metrics into health score (0-100)
- Define thresholds for warning/critical
- Expose via `/health` endpoint

## Phase 4: Configuration & Retention

### 4.1 Configurable Collection
**File**: `src/api/metrics_config.go`
- Environment variables for collection intervals
- Per-metric collection rates
- Dynamic configuration updates

### 4.2 Retention Policy Implementation
**File**: `src/storage/binary/metrics_retention.go`
- Enforce retention tags on metric entities
- Automatic cleanup of old data points
- Configurable retention periods

### 4.3 Metric Aggregation
**File**: `src/api/metrics_aggregator.go`
- Downsample old metrics (hourly, daily averages)
- Reduce storage for historical data
- Maintain precision for recent data

## Phase 5: Business Intelligence

### 5.1 Dataspace Metrics
**File**: `src/api/dataspace_metrics.go`
- Per-dataspace entity counts
- Storage usage by dataspace
- Activity metrics per dataspace

### 5.2 Feature Usage Analytics
**File**: `src/api/feature_metrics.go`
- Track API endpoint usage
- Monitor feature adoption
- User behavior patterns

## Implementation Schedule

### Week 1: Critical Metrics
- [ ] Request/response metrics middleware
- [ ] Error tracking system
- [ ] WAL checkpoint metrics
- [ ] UI chart legends and units

### Week 2: Performance Metrics
- [ ] Query performance tracking
- [ ] Write performance metrics
- [ ] Cache metrics implementation
- [ ] Interactive chart tooltips

### Week 3: Health Monitoring
- [ ] Resource metrics collector
- [ ] Connection pool tracking
- [ ] Health score calculation
- [ ] Configurable time ranges in UI

### Week 4: Configuration
- [ ] Configurable collection intervals
- [ ] Retention policy enforcement
- [ ] Metric aggregation system
- [ ] Real-time chart updates

### Week 5-6: Business Metrics
- [ ] Dataspace metrics
- [ ] Feature usage tracking
- [ ] Capacity planning metrics
- [ ] Custom dashboard support

## Testing Strategy

1. **Unit Tests**: Test each metric collector in isolation
2. **Integration Tests**: Verify metrics flow from collection to storage
3. **Load Tests**: Ensure metrics don't impact performance
4. **UI Tests**: Verify chart rendering and interactivity

## Success Criteria

1. All critical operational metrics collected
2. Charts display clear legends and units
3. Configurable collection and retention
4. Less than 1% performance impact
5. Complete visibility into service health

## Configuration Examples

```bash
# Collection intervals
ENTITYDB_METRICS_INTERVAL=30s
ENTITYDB_METRICS_REQUEST_INTERVAL=5s
ENTITYDB_METRICS_RESOURCE_INTERVAL=60s

# Retention settings
ENTITYDB_METRICS_RETENTION_RAW=24h
ENTITYDB_METRICS_RETENTION_HOURLY=7d
ENTITYDB_METRICS_RETENTION_DAILY=30d

# Feature flags
ENTITYDB_METRICS_ENABLE_REQUEST=true
ENTITYDB_METRICS_ENABLE_QUERY=true
ENTITYDB_METRICS_ENABLE_CACHE=true
```