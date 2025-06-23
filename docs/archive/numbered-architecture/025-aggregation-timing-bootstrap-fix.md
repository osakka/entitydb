# ADR-025: Aggregation Timing Bootstrap Fix

## Status
Accepted

## Context

EntityDB's metrics retention system was experiencing systematic failures with aggregated metrics lookups, generating continuous error logs:

```
Failed READ operation for entity metric_http_response_size_bytes_agg_1day: entity not found in index
Failed READ operation for entity metric_cpu_count_agg_1min: entity not found in index
```

### Root Cause Analysis

The issue stemmed from a **timing bootstrap problem** between two subsystems:

1. **MetricsRetentionManager.enforceRetention()**: 
   - Starts 5 minutes after server startup
   - Immediately processes ALL metrics including aggregated ones (`*_agg_1min`, `*_agg_1hour`, `*_agg_1day`)
   - Applies different retention policies based on aggregation level

2. **MetricsRetentionManager.performAggregation()**: 
   - Only runs on 5-minute timer intervals
   - No initial bootstrap run
   - Creates aggregated metrics from raw data

**The Problem**: Retention ran first (5 minutes) looking for aggregated metrics that aggregation hadn't created yet (also waiting for 5+ minutes).

### Architecture Flaw

```
Timeline:
T+0:00  Server starts
T+5:00  Retention starts → looks for aggregated metrics → NOT FOUND
T+5:00+ Aggregation eventually runs → creates aggregated metrics → TOO LATE
```

This created a **bootstrap chicken-and-egg problem** where retention expected aggregated metrics that aggregation hadn't yet created.

## Decision

We will fix the bootstrap timing by ensuring aggregation runs **before** retention:

### Solution Architecture

1. **Add Initial Aggregation Run**: Run aggregation after 2 minutes (before retention at 5 minutes)
2. **Maintain Periodic Schedule**: Continue 5-minute periodic aggregation after initial run
3. **Preserve Retention Timing**: Keep retention at 5 minutes to ensure stability

### New Timeline

```
T+0:00  Server starts
T+2:00  Aggregation runs → creates aggregated metrics
T+5:00  Retention starts → finds existing aggregated metrics → SUCCESS
T+7:00  Aggregation runs periodically (2+5)
T+10:00 Aggregation runs periodically (7+5)
...
```

## Implementation

### Code Changes

**File**: `/opt/entitydb/src/api/metrics_retention_manager.go`

**Before** (Problematic):
```go
// Run aggregation every 5 minutes
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            m.performAggregation()
        case <-m.ctx.Done():
            return
        }
    }
}()
```

**After** (Fixed):
```go
// Run aggregation every 5 minutes
go func() {
    // Initial delay to let system stabilize, but shorter than retention delay
    select {
    case <-time.After(2 * time.Minute):
    case <-m.ctx.Done():
        return
    }
    
    // Run immediately on start (before retention kicks in)
    m.performAggregation()
    
    // Then run periodically
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            m.performAggregation()
        case <-m.ctx.Done():
            return
        }
    }
}()
```

### Enhanced Logging

Added comprehensive logging to track aggregation performance:

```go
logger.Info("Starting metrics aggregation")
logger.Info("Found %d total metrics for aggregation processing", len(metrics))
logger.Info("Aggregation complete: %d raw metrics processed, %d aggregations performed in %v", 
    rawMetricsCount, aggregatedCount, duration)
```

## Testing Results

### Before Fix
- Continuous lookup failures every 3-5 seconds
- No aggregated metrics created
- System logs filled with "entity not found" errors
- Failed recovery attempts

### After Fix  
- **18:05:50**: Aggregation completed: 83 raw metrics → 249 aggregations (3.46s)
- **18:08:50**: Retention started cleanly  
- **18:09:26**: Retention completed: 78 metrics cleaned, 673 tags removed
- **Zero lookup failures**
- **Clean system operation**

### Performance Metrics
- **Aggregation Performance**: 83 raw metrics → 249 aggregations in 3.46 seconds
- **Retention Performance**: 78 metrics processed, 673 old tags removed in 36 seconds  
- **Zero Failed Lookups**: Complete elimination of "entity not found" errors
- **Successful WAL Management**: 994KB → 0 bytes checkpoint

## Consequences

### Positive
- ✅ **Eliminates Bootstrap Problem**: Aggregated metrics exist when retention needs them
- ✅ **Clean System Logs**: No more "entity not found" error spam
- ✅ **Improved Observability**: Enhanced logging shows exact aggregation performance
- ✅ **Faster System Stabilization**: Aggregated metrics available 3 minutes earlier
- ✅ **Maintains Performance**: No impact on normal operation performance

### Negative
- ⚠️ **Slightly Earlier Resource Usage**: Aggregation starts 3 minutes sooner
- ⚠️ **Additional Logging**: More INFO-level log entries during startup

### Neutral
- 🔄 **Maintains Existing Interfaces**: No API or configuration changes required
- 🔄 **Preserves Retention Logic**: No changes to retention policy enforcement
- 🔄 **Same Resource Requirements**: No additional memory or CPU overhead

## Alternatives Considered

### Option 1: Bootstrap Aggregated Entities (Rejected)
**Approach**: Pre-create empty aggregated metric entities during server initialization.
**Rejected Because**: Would require significant changes to initialization logic and could create consistency issues.

### Option 2: Make Retention More Resilient (Rejected)  
**Approach**: Handle missing aggregated metrics gracefully in retention logic.
**Rejected Because**: Addresses symptoms rather than root cause; aggregated metrics should exist.

### Option 3: Timing Coordination (Chosen)
**Approach**: Fix the timing issue by ensuring aggregation runs before retention.
**Chosen Because**: Solves root cause with minimal code changes and preserves existing architecture.

## Implementation Notes

### Thread Safety
- Uses existing context cancellation for clean shutdown
- Maintains same concurrency patterns as original implementation
- No new race conditions introduced

### Backward Compatibility
- Zero breaking changes to external APIs
- Existing retention policies unchanged
- Configuration parameters unaffected

### Rollback Plan
If issues arise, revert to original timing by removing the initial aggregation run:
```go
// Remove lines 76-84 and revert to ticker-only approach
```

## Monitoring

### Success Indicators
- ✅ Aggregation logs show successful processing every 5 minutes
- ✅ Retention logs show clean completion without lookup failures  
- ✅ Zero "entity not found in index" errors for aggregated metrics
- ✅ WAL checkpoints complete successfully

### Alert Conditions
- 🚨 Aggregation failing to process expected raw metrics
- 🚨 Return of "entity not found" errors for aggregated metrics
- 🚨 Retention processing taking significantly longer than 40 seconds

## Related ADRs

- **ADR-024**: Incremental Update Architecture - Fixed database rebuild performance issues
- **ADR-007**: Temporal Retention Architecture - Original metrics retention design

## Decision Date
2025-06-19

## Decision Makers
- EntityDB Core Team

---

*This ADR resolves the aggregated metrics bootstrap timing issue, ensuring clean system operation and eliminating lookup failure noise in system logs.*