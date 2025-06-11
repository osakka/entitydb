# EntityDB Metrics Phase 2 Implementation

**Date**: June 7, 2025  
**Version**: v2.28.0  
**Status**: Phase 2 Complete

## Overview

Phase 2 implementation leverages EntityDB's native temporal features to provide comprehensive data management capabilities for metrics. Every metric is stored as an entity with temporal tags, providing automatic history tracking.

## Key Insight: Temporal Storage Architecture

EntityDB's temporal architecture means:
- Each metric is an entity with metadata tags (`type:metric`, `name:`, `unit:`, etc.)
- Every value update creates a new temporal tag with nanosecond timestamp
- All historical values are preserved automatically
- No need for separate time-series database

Example metric entity structure:
```json
{
  "id": "metric_memory_alloc",
  "tags": [
    "1749280047573038838|type:metric",
    "1749280047573041144|dataset:system",
    "1749280047573042321|name:memory_alloc",
    "1749280047573044785|unit:bytes",
    "1749280047573045692|description:Memory currently allocated",
    "1749280047573046750|value:20919912.00",
    "1749280080199717952|value:28077872.00",
    "1749280110176013158|value:29305856.00",
    // ... hundreds more temporal value tags
  ]
}
```

## Completed Features

### 1. Retention Policy Enforcement (`metrics_retention_manager.go`)

The retention manager leverages temporal tags to enforce data lifecycle policies:

- **Automatic Cleanup**: Runs hourly to remove old temporal value tags
- **Configurable Retention**: Different retention periods for raw and aggregated data
- **Tag-Based Filtering**: Only removes `value:` tags older than retention period
- **Efficient Updates**: Updates entities with filtered tag lists

Key features:
```go
// Retention configuration from environment
MetricsRetentionRaw:   24 hours (default)
MetricsRetention1Min:  7 days
MetricsRetention1Hour: 30 days
MetricsRetention1Day:  365 days
```

### 2. Metric Aggregation System

Aggregation creates new metric entities with different time granularities:

- **Automatic Rollups**: Creates 1-minute, 1-hour, and daily aggregates
- **Preserves Statistics**: Stores avg, min, max, and count for each period
- **Temporal Queries**: Uses entity's temporal tags to calculate aggregates
- **Separate Entities**: Aggregated metrics stored as `metric_{name}_agg_{interval}`

Aggregated value format:
```
value:avg:min:max:count
Example: value:25.5:avg:25.5:min:20.0:max:30.0:count:10
```

### 3. Advanced Metric Types (`metrics_types.go`)

Comprehensive metric type system leveraging temporal storage:

#### Counter
- Monotonically increasing values
- Automatic rate calculation from temporal data
- GetCounterRate() calculates change per second

#### Gauge  
- Current values that can go up or down
- Direct value storage with temporal tags

#### Histogram
- Stores individual observations as temporal tags
- Calculates percentiles on demand
- Configurable bucket boundaries

#### Summary (Future)
- Statistical summaries with quantiles
- Built on histogram observations

### 4. Enhanced History API

Updated metrics history endpoint to support aggregated data:

```http
GET /api/v1/metrics/history?metric_name=memory_alloc&aggregation=1hour
```

Parameters:
- `aggregation`: raw, 1min, 1hour, 1day (default: raw)
- Automatically selects appropriate metric entity
- Handles aggregated value format parsing

## Technical Implementation Details

### Temporal Tag Management

1. **Value Storage**: Each metric update adds a temporal tag
   ```
   1749280047573046750|value:20919912.00
   ```

2. **Retention Enforcement**: Filters tags by timestamp
   ```go
   cutoffTime := time.Now().Add(-retention)
   cutoffNanos := cutoffTime.UnixNano()
   ```

3. **Aggregation Queries**: Extracts values within time buckets
   ```go
   bucket := t.Truncate(interval)
   buckets[bucket] = append(buckets[bucket], value)
   ```

### Performance Considerations

1. **Lazy Aggregation**: Only aggregates when queried
2. **Incremental Updates**: Only processes new data since last aggregation
3. **Concurrent Safety**: Thread-safe operations with mutex protection
4. **Efficient Storage**: Leverages EntityDB's binary format

## Configuration

New environment variables for Phase 2:
```bash
# Retention periods (in minutes)
ENTITYDB_METRICS_RETENTION_RAW=1440      # 24 hours
ENTITYDB_METRICS_RETENTION_1MIN=10080    # 7 days
ENTITYDB_METRICS_RETENTION_1HOUR=43200   # 30 days
ENTITYDB_METRICS_RETENTION_1DAY=525600   # 365 days

# Histogram configuration
ENTITYDB_METRICS_HISTOGRAM_BUCKETS=0.001,0.005,0.01,0.05,0.1,0.5,1,5,10
```

## Integration

Phase 2 components integrate seamlessly:

1. **Retention Manager**: Started conditionally in main.go
   ```go
   if cfg.MetricsRetentionRaw > 0 {
       retentionManager := api.NewMetricsRetentionManager(...)
       retentionManager.Start()
   }
   ```

2. **Metric Types**: Available through MetricsTypeManager
   ```go
   typeManager := api.NewMetricsTypeManager(repo, buckets)
   typeManager.RecordCounter("requests_total", 1, labels, "Total requests")
   ```

3. **History API**: Automatically supports aggregated metrics
   ```go
   GET /api/v1/metrics/history?metric_name=requests_total&aggregation=1hour
   ```

## Benefits of Temporal Approach

1. **No External Dependencies**: Uses EntityDB's native storage
2. **Automatic History**: Every metric update is preserved
3. **Flexible Queries**: Can reconstruct any time period
4. **Consistent Model**: Metrics are just entities with tags
5. **Built-in Durability**: WAL and binary storage ensure persistence

## Testing Recommendations

1. **Retention Testing**:
   - Create metric with short retention
   - Add multiple values
   - Wait for retention enforcement
   - Verify old values removed

2. **Aggregation Testing**:
   - Create high-frequency metric
   - Wait for aggregation cycle
   - Query aggregated metrics
   - Verify statistics accuracy

3. **Type Testing**:
   - Test counter increment and rate calculation
   - Test gauge value updates
   - Test histogram percentile calculations

## Next Steps (Phase 3 & 4)

- Phase 3: UI enhancements (already mostly complete)
- Phase 4: Advanced monitoring features
  - Anomaly detection using temporal data
  - Metric correlation analysis
  - Alert rule evaluation
  - Export to external systems

## Summary

Phase 2 successfully implements comprehensive data management using EntityDB's temporal features. The system provides:
- Automatic retention enforcement
- Multi-level aggregation
- Advanced metric types
- Seamless integration with existing infrastructure

All metrics benefit from EntityDB's ACID compliance, durability, and performance optimizations.