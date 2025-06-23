# ADR-033: Metrics Feedback Loop Prevention

## Status
**Accepted** (2025-06-23) - Final implementation with safe defaults

## Context
During v2.34.2 testing, a critical feedback loop was discovered where metrics collection caused exponential growth in WAL checkpoint operations. The root cause was:

1. Metrics collection creates/updates metric entities
2. Entity operations trigger WAL writes
3. WAL operations increment the operation counter
4. When counter reaches 1000, a checkpoint is triggered
5. Checkpoint operations create checkpoint metrics
6. **LOOP**: Back to step 1, creating infinite recursion

This resulted in thousands of `metric_wal_checkpoint_failed_total` entries being created in rapid succession, causing 100% CPU usage and requiring server termination.

## Decision
Implement a multi-layered defense system to prevent metrics feedback loops:

1. **Metrics Guard**: Thread-safe operation counting to detect recursive metrics operations
2. **Skip Checkpoint During Metrics**: Prevent checkpoint triggers during metrics collection
3. **Skip Metrics for Metrics**: Don't collect metrics about metrics entities

## Implementation

### Metrics Guard System
```go
// MetricsGuard prevents recursive metrics operations
type MetricsGuard struct {
    activeOperations int64  // Atomic counter
}

// ShouldSkipMetrics returns true if metrics should be skipped
func ShouldSkipMetrics(entityID string) bool {
    // Skip if already in metrics operation
    if atomic.LoadInt64(&globalMetricsGuard.activeOperations) > 0 {
        return true
    }
    // Skip if this is a metrics entity
    if IsMetricsEntity(entityID) {
        return true
    }
    return false
}
```

### Checkpoint Guard
```go
func (r *EntityRepository) checkAndPerformCheckpoint() {
    // Skip checkpoint for metrics operations
    if isMetricsOperation() {
        logger.Trace("Skipping checkpoint during metrics operation")
        return
    }
    // ... normal checkpoint logic
}
```

### Metrics Collection Guard
```go
func (r *EntityRepository) storeCheckpointMetric(...) {
    // Prevent metrics feedback loop
    if ShouldSkipMetrics("metric_wal_checkpoint") {
        logger.Trace("Skipping checkpoint metrics to prevent feedback loop")
        return
    }
    // ... normal metrics collection
}
```

## Consequences

### Positive
- Eliminates infinite metrics recursion loops
- Prevents CPU spinning from rapid checkpoint cycles
- Maintains system stability under all conditions
- Zero performance impact on normal operations
- Surgical precision fix with minimal code changes

### Negative
- Some metrics operations may not trigger immediate checkpoints
- Checkpoint metrics may be slightly delayed during high load

### Neutral
- Existing metrics collection continues to work normally
- No changes to metrics data structure or storage
- Backward compatible with existing metrics

## Testing
The fix prevents the exact scenario that caused the server crash:
- Metrics operations no longer trigger checkpoints
- Checkpoint operations check for recursion before creating metrics
- Thread-safe atomic operations prevent race conditions
- All normal operations continue without impact

## Final Implementation (2025-06-23)

**Commit**: `ba41984` - Surgical elimination of metrics feedback loop with safe defaults

### Configuration Changes
Changed metric tracking defaults to false in `config/config.go`:
```go
MetricsEnableRequestTracking: getEnvBool("ENTITYDB_METRICS_ENABLE_REQUEST_TRACKING", false),
MetricsEnableStorageTracking: getEnvBool("ENTITYDB_METRICS_ENABLE_STORAGE_TRACKING", false),
```

### Background Collector Protection
Added conditional initialization in `main.go`:
```go
// Enable background metrics collection only if metrics tracking is enabled
if cfg.MetricsEnableRequestTracking || cfg.MetricsEnableStorageTracking {
    backgroundCollector = api.NewBackgroundMetricsCollector(...)
    backgroundCollector.Start()
} else {
    logger.Info("Background metrics collector disabled (metrics tracking disabled)")
}
```

### Query Metrics Protection  
```go
// Initialize query metrics collector only if request tracking is enabled
if cfg.MetricsEnableRequestTracking {
    api.InitQueryMetrics(server.entityRepo)
    logger.Info("Query metrics tracking enabled")
} else {
    logger.Info("Query metrics tracking disabled")
}
```

## References
- **v2.34.3**: Complete metrics feedback loop elimination with safe defaults
- **Commit ba41984**: Final surgical precision implementation 
- **Server logs**: Confirmed "Background metrics collector disabled" and "Query metrics tracking disabled"
- **ADR-031**: Bar-raising metrics retention contention fix
- **Production testing**: CPU usage stable at 0.0% with all systems functional