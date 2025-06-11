# EntityDB Metrics Improvement Action Plan

**Version**: v2.28.0  
**Date**: June 7, 2025  
**Priority**: High

## Overview

This action plan addresses the critical gaps identified in the metrics analysis. Implementation will be done in phases to ensure system stability while adding comprehensive monitoring capabilities.

## Phase 1: Critical Operational Metrics (Immediate)

### 1.1 Request/Response Metrics
- [ ] Add request metrics middleware to track:
  - Request count by endpoint, method, status code
  - Response time histograms (p50, p95, p99, p999)
  - Request/response sizes
  - Active concurrent requests
  - Error rates and types

### 1.2 Storage Operation Metrics
- [ ] Instrument storage layer for:
  - Read/write operation counts
  - Operation latencies (histogram)
  - Cache hit/miss rates
  - Lock acquisition times and contention
  - WAL operations (writes, syncs, checkpoints)

### 1.3 Enhanced Configuration
- [ ] Add configuration options:
  ```go
  ENTITYDB_METRICS_RETENTION_RAW=24h           # Raw data retention
  ENTITYDB_METRICS_RETENTION_1MIN=7d           # 1-minute aggregates
  ENTITYDB_METRICS_RETENTION_1HOUR=30d         # 1-hour aggregates
  ENTITYDB_METRICS_RETENTION_1DAY=365d         # Daily aggregates
  ENTITYDB_METRICS_HISTOGRAM_BUCKETS=0.001,0.005,0.01,0.05,0.1,0.5,1,5,10
  ENTITYDB_METRICS_ENABLE_REQUEST_TRACKING=true
  ENTITYDB_METRICS_ENABLE_STORAGE_TRACKING=true
  ```

## Phase 2: Data Management & Temporal Features

### 2.1 Retention Policy Implementation
- [ ] Create metric retention system using temporal tags:
  - Add `retention:until:<timestamp>` tags
  - Background job to clean expired metrics
  - Configurable per metric type

### 2.2 Automatic Aggregation
- [ ] Implement metric rollups:
  - 1-minute averages for gauges
  - Rate calculations for counters
  - Percentile preservation for histograms
  - Sum/count for aggregated metrics

### 2.3 Metric Types Support
- [ ] Extend metric system to support:
  - Counter (monotonic, with rate calculation)
  - Gauge (current value)
  - Histogram (distribution with percentiles)
  - Summary (aggregated statistics)

## Phase 3: UI Enhancements

### 3.1 Chart Improvements
- [ ] Enhance chart rendering:
  ```javascript
  // Add to each chart:
  - Legend with metric names and current values
  - Y-axis labels with units
  - Tooltips on hover showing exact values
  - Time range selector
  - Zoom/pan capabilities
  ```

### 3.2 New Visualizations
- [ ] Add chart types:
  - Stacked area for cumulative metrics
  - Heatmaps for latency distributions
  - Gauge charts for current values
  - Sparklines for inline metrics

### 3.3 Dashboard Enhancements
- [ ] Create dedicated metrics dashboard with:
  - System health overview
  - Request/response analytics
  - Storage performance
  - Security/RBAC monitoring
  - Custom metric builder

## Phase 4: Advanced Monitoring

### 4.1 Security & RBAC Metrics
- [ ] Track authentication/authorization:
  - Login attempts (success/failure) by user
  - Permission checks by type
  - Session lifecycle metrics
  - API key usage patterns

### 4.2 Business Metrics
- [ ] Application-level tracking:
  - Entity operations by type
  - Query patterns and performance
  - Dataset utilization
  - Relationship graph statistics

### 4.3 Temporal Query Metrics
- [ ] Monitor temporal features:
  - As-of query performance
  - History query statistics
  - Timeline index efficiency
  - Temporal storage overhead

## Implementation Details

### Metric Naming Convention
```
entitydb_<subsystem>_<metric>_<unit>
Examples:
- entitydb_http_requests_total
- entitydb_http_request_duration_seconds
- entitydb_storage_read_operations_total
- entitydb_storage_read_latency_seconds
- entitydb_auth_login_attempts_total
- entitydb_cache_hits_total
```

### Temporal Storage Schema
```go
// Metric entity structure
Entity {
  ID: "metric_http_request_duration_seconds",
  Tags: [
    "type:metric",
    "metric:type:histogram",
    "metric:unit:seconds",
    "metric:description:HTTP request duration",
    "retention:raw:24h",
    "retention:aggregated:30d"
  ],
  Content: nil // Values stored as temporal tags
}

// Metric value as temporal tag
"1749303910369730667|metric:value:0.045:count:1:sum:0.045:p50:0.04:p95:0.05:p99:0.06"
```

### Aggregation Rules
```go
type AggregationRule struct {
  MetricPattern string        // Regex pattern
  Interval      time.Duration // Aggregation interval
  Method        string        // avg, sum, max, min, p50, p95, p99
  Retention     time.Duration // How long to keep aggregated data
}
```

## Testing Plan

1. **Load Testing**: Verify metrics don't impact performance
2. **Accuracy Testing**: Validate metric calculations
3. **UI Testing**: Ensure charts render correctly
4. **Retention Testing**: Verify data cleanup works
5. **Integration Testing**: Test with external monitoring tools

## Success Criteria

- [ ] All critical system behaviors have metrics
- [ ] Metrics overhead < 2% of system resources
- [ ] Charts have legends, units, and are interactive
- [ ] Data retention keeps storage bounded
- [ ] Can detect and diagnose performance issues
- [ ] Security events are tracked and visible
- [ ] Metrics can be exported to Prometheus/Grafana

## Timeline

- **Week 1**: Phase 1 - Critical operational metrics
- **Week 2**: Phase 2 - Data management features
- **Week 3**: Phase 3 - UI enhancements
- **Week 4**: Phase 4 - Advanced monitoring

## Risk Mitigation

1. **Performance Impact**: Use sampling for high-frequency metrics
2. **Storage Growth**: Implement aggressive retention policies
3. **Breaking Changes**: Keep existing endpoints, add new ones
4. **Complexity**: Phase implementation to manage risk

## Notes

- Leverage temporal storage for all metric history
- Use change detection to minimize storage
- Follow Prometheus naming conventions
- Ensure all metrics have help text and units
- Make everything configurable