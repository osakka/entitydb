# EntityDB Metrics Enhancement Action Plan

## Overview

Based on the metrics audit findings, this action plan addresses critical gaps in metrics collection, configuration, and visualization. Implementation follows a phased approach prioritizing operational visibility.

## Phase 1: Critical Performance Metrics ✅ COMPLETE

### 1.1 Query Performance Metrics ✅
**Status: COMPLETE**

**Implemented:**
- Created `query_metrics_middleware.go` to wrap all query operations
- Tracks metrics:
  - `query_execution_time_ms` - Time to execute queries
  - `query_result_count` - Number of results returned
  - `query_complexity_score` - Based on tags, operators, time range
  - `slow_query_count` - Queries exceeding 500ms threshold
  - `query_error_count` - Query failures by type

**Code Location:** `/opt/entitydb/src/api/query_metrics_middleware.go`

### 1.2 Storage Operation Metrics ✅
**Status: COMPLETE**

**Implemented:**
- Instrumented `storage/binary/*.go` with operation metrics
- Tracks metrics:
  - `storage_read_duration_ms` - Read operation time
  - `storage_write_duration_ms` - Write operation time
  - `index_lookup_duration_ms` - Index search time
  - `wal_operation_duration_ms` - WAL operation time
  - `compression_ratio` - Actual vs compressed size
  - `storage_cache_hits/misses` - Cache performance

**Code Location:** `/opt/entitydb/src/storage/binary/metrics_instrumentation.go`

### 1.3 Enhanced Error Tracking ✅
**Status: COMPLETE**

**Implemented:**
- Created centralized error collector
- Tracks metrics:
  - `error_count` by type, severity, component
  - `error_recovery_time_ms` - Time to recover
  - `panic_count` - Unrecoverable errors
  - `frequent_error_patterns` - Patterns occurring >10 times
  - `recovery_attempts` - Success/failure tracking

**Code Location:** `/opt/entitydb/src/api/error_metrics_collector.go`

### 1.4 Request/Response Metrics ✅
**Status: COMPLETE**

**Implemented:**
- Created HTTP request/response middleware
- Tracks metrics:
  - `http_request_duration_ms` - Request processing time
  - `http_request_size_bytes` - Request body size
  - `http_response_size_bytes` - Response body size
  - `http_request_count` - By method/path/status
  - `http_request_errors` - Request failures

**Code Location:** `/opt/entitydb/src/api/request_metrics_middleware.go`

### 1.5 Configuration Enhancements ✅
**Status: COMPLETE**

**Implemented:**
- Made metrics collection interval configurable
- Added `ENTITYDB_METRICS_INTERVAL` environment variable
- Default: 30 seconds, supports any Go duration format
- Validates and logs configured interval on startup

**Code Location:** `/opt/entitydb/src/main.go:437-445`

### 1.6 UI Performance Tab ✅
**Status: COMPLETE**

**Implemented:**
- Enhanced Performance tab with new metrics
- Added metric cards for query performance, storage operations, errors
- Improved charts with proper legends and units
- Added real-time data loading for performance metrics

**Code Location:** `/opt/entitydb/share/htdocs/index.html` (Performance tab)

## Phase 2: Configuration System (Week 1)

### 2.1 Metrics Configuration

**Implementation Tasks:**
1. Create configuration structure for metrics
2. Support environment variables and entity-based config
3. Configuration options:
   ```go
   type MetricsConfig struct {
       Collection struct {
           DefaultInterval time.Duration
           Intervals map[string]time.Duration
           ChangeDetection bool
           ChangeThreshold float64
       }
       Retention struct {
           DefaultDuration time.Duration
           DefaultMaxPoints int
           Tiers []RetentionTier
       }
       Aggregation struct {
           Enabled bool
           Functions []string
           Interval time.Duration
       }
   }
   ```

**Code Location:** `src/models/metrics_config.go`

### 2.2 Dynamic Collection Control

**Implementation Tasks:**
1. Make background collector configurable
2. Support per-metric collection intervals
3. Add pause/resume capabilities
4. Implement conditional collection based on system load

**Code Location:** Update `src/api/metrics_background_collector.go`

## Phase 3: Business Metrics (Week 1-2)

### 3.1 Entity Operation Metrics

**Implementation Tasks:**
1. Track all entity operations
2. Metrics to implement:
   - `entity_operations_per_sec` by type
   - `tag_cardinality` by namespace
   - `dataspace_operations` by type
   - `relationship_operations` count

**Code Location:** Update entity handlers in `src/api/`

### 3.2 User Activity Metrics

**Implementation Tasks:**
1. Track user interactions
2. Metrics to implement:
   - `user_api_calls` by endpoint
   - `user_session_duration`
   - `concurrent_users`
   - `failed_auth_attempts`

**Code Location:** Update `src/api/auth_middleware.go`

## Phase 4: Advanced Features (Week 2-3)

### 4.1 Metric Aggregation

**Implementation Tasks:**
1. Implement aggregation engine
2. Support functions: min, max, avg, sum, p50, p95, p99
3. Automatic rollup based on retention tiers
4. Store aggregated values as temporal tags

**Code Location:** `src/models/metrics_aggregator.go`

### 4.2 Retention Management

**Implementation Tasks:**
1. Implement automatic cleanup based on retention policies
2. Support tiered retention with different resolutions
3. Compress old metric data
4. Archive to cold storage (optional)

**Code Location:** `src/storage/binary/metrics_retention_manager.go`

### 4.3 Alerting Framework

**Implementation Tasks:**
1. Define alert rules as entities
2. Implement threshold-based alerts
3. Support rate-of-change alerts
4. Add alert notification system

**Code Location:** `src/models/metrics_alerting.go`

## Implementation Status Summary

### ✅ Complete (Phase 1)
- Query performance metrics
- Storage operation metrics
- Error tracking system
- Request/response metrics
- Configurable collection interval
- UI performance enhancements

### ⏳ Pending (Phase 2-4)
- Advanced configuration system
- Business metrics collection
- User activity tracking
- Metric aggregation engine
- Retention management
- Alerting framework

## Success Criteria

1. **Coverage**: ✅ All critical operations have metrics
2. **Performance**: ✅ Metrics collection overhead < 1%
3. **Usability**: ✅ Charts are clear and actionable
4. **Configuration**: ✅ Basic configuration complete, advanced pending
5. **Retention**: ⏳ Manual cleanup, automatic pending

## Testing Strategy

### Completed Tests
- [x] Manual testing of all metric collectors
- [x] UI chart rendering validation
- [x] Performance overhead measurement

### Pending Tests
- [ ] Unit tests for metric collectors
- [ ] Integration tests for end-to-end flow
- [ ] Load tests for scalability
- [ ] Automated UI tests

## Next Steps

1. **Immediate**: Document Phase 1 implementation
2. **This Week**: Begin Phase 2 configuration system
3. **Next Week**: Implement business metrics
4. **Following Week**: Advanced analytics features

## Notes

Phase 1 implementation successfully addresses all critical metrics gaps identified in the audit. The system now provides comprehensive visibility into:
- Query performance and complexity
- Storage operation efficiency
- Error patterns and recovery
- HTTP request/response characteristics
- System resource utilization

All implementations follow EntityDB's temporal storage model and maintain minimal performance overhead through asynchronous processing and change detection.