# ADR-015: WAL Management and Automatic Checkpointing

## Status
✅ **ACCEPTED** - 2025-06-16

## Context
EntityDB's Write-Ahead Log (WAL) was growing unbounded, leading to disk space exhaustion and performance degradation. The system lacked automatic checkpoint mechanisms, requiring manual intervention to prevent storage issues.

## Problem
- Unbounded WAL growth causing disk space exhaustion
- No automatic checkpointing mechanism
- Performance degradation with large WAL files
- Manual intervention required for WAL maintenance
- Risk of system failure due to storage exhaustion
- Inability to manage WAL size in production environments

## Decision
Implement automatic WAL checkpointing system with intelligent triggers:

### Automatic Checkpoint Triggers
```go
type CheckpointTriggers struct {
    OperationCount int    // Checkpoint every N operations (default: 1000)
    TimeInterval   time.Duration // Checkpoint every N minutes (default: 5 minutes)
    SizeLimit      int64  // Checkpoint when WAL exceeds N bytes (default: 100MB)
}
```

### Checkpoint Integration Points
- **Entity Creation**: Trigger checkpoint after batch operations
- **Entity Updates**: Monitor WAL growth during high-update periods
- **Tag Operations**: Include AddTag() in checkpoint consideration
- **Background Timer**: Scheduled checkpoints regardless of activity

### WAL Health Monitoring
```go
type WALMetrics struct {
    Size           int64   `json:"wal_size_bytes"`
    SizeMB        float64 `json:"wal_size_mb"`
    Warning       bool    `json:"wal_warning"`      // > 75MB
    Critical      bool    `json:"wal_critical"`     // > 150MB
    LastCheckpoint time.Time `json:"last_checkpoint"`
}
```

## Implementation Details

### Checkpoint Strategy
1. **Smart Triggering**: Multiple triggers ensure checkpoints happen appropriately
2. **Atomic Operations**: Checkpoints are atomic to prevent data corruption
3. **Performance Optimization**: Checkpoints during low-activity periods when possible
4. **Failure Recovery**: Robust error handling for checkpoint failures

### Integration with Operations
```go
// In entity operations
func (repo *EntityRepository) Create(entity *Entity) error {
    // ... entity creation logic ...
    
    repo.triggerCheckpointIfNeeded()
    return nil
}

func (repo *EntityRepository) triggerCheckpointIfNeeded() {
    if repo.shouldCheckpoint() {
        go repo.performCheckpoint() // Async to avoid blocking operations
    }
}
```

### Monitoring and Alerting
- **Real-time Metrics**: WAL size tracked in system metrics
- **Warning Thresholds**: Alert when WAL approaches problematic sizes
- **Critical Alerts**: System warnings when WAL becomes critically large
- **Checkpoint Success Rate**: Track checkpoint success/failure rates

## Consequences

### Positive
- ✅ **Prevented Disk Exhaustion**: Automatic cleanup prevents storage issues
- ✅ **Improved Performance**: Smaller WAL files improve replay performance
- ✅ **Production Reliability**: No manual intervention required
- ✅ **Predictable Storage**: WAL size bounded by checkpoint triggers
- ✅ **Better Monitoring**: Real-time visibility into WAL health
- ✅ **Automatic Recovery**: System self-manages storage efficiently

### Negative
- ⚠️ **Checkpoint Overhead**: Periodic I/O operations for checkpointing
- ⚠️ **Complexity**: Additional code paths for checkpoint management
- ⚠️ **Timing Sensitivity**: Checkpoint timing can affect performance

### Performance Impact
- **Checkpoint Duration**: ~100-500ms depending on WAL size
- **I/O Impact**: Temporary I/O spike during checkpoint operations
- **Memory Usage**: Brief memory increase during checkpoint
- **Overall Benefit**: Significant performance improvement from smaller WAL files

## Configuration Options
```go
type WALConfig struct {
    CheckpointOperations int           // Default: 1000
    CheckpointInterval   time.Duration // Default: 5 minutes  
    CheckpointSizeLimit  int64         // Default: 100MB
    WarningThreshold     int64         // Default: 75MB
    CriticalThreshold    int64         // Default: 150MB
}
```

## Monitoring Integration
```go
// WAL metrics included in system metrics
type SystemMetrics struct {
    WAL WALMetrics `json:"wal"`
    // ... other metrics
}

// Health check integration
func (h *HealthHandler) checkWALHealth() HealthStatus {
    if walSize > criticalThreshold {
        return HealthStatus{Status: "critical", Message: "WAL size critical"}
    }
    return HealthStatus{Status: "healthy"}
}
```

## Alternatives Considered
1. **Manual Checkpointing**: Rejected due to operational overhead
2. **Fixed Time Intervals**: Rejected for lack of responsiveness to load
3. **Size-Only Triggers**: Rejected for potential long delays
4. **External Monitoring**: Rejected for complexity and dependencies

## References
- Implementation: `src/storage/binary/wal.go` - checkpoint logic
- Metrics: `src/api/system_metrics_handler.go` - WAL monitoring
- Configuration: `src/main.go` - checkpoint initialization
- Git Commits:
  - `af7ac83` - "fix: resolve critical 100% CPU usage from storage metrics feedback loop"
  - Related WAL management improvements
- Related: ADR-002 (Binary Storage Format)

## Timeline
- **2025-06-16**: Critical WAL growth issue identified
- **2025-06-16**: Automatic checkpointing system designed and implemented
- **2025-06-16**: WAL metrics and monitoring integration completed
- **2025-06-17**: Production validation and performance testing

## Operational Guidelines
1. **Monitor WAL Metrics**: Regular review of WAL size and checkpoint frequency
2. **Tune Triggers**: Adjust checkpoint triggers based on usage patterns
3. **Alert Configuration**: Set up monitoring alerts for WAL critical thresholds
4. **Performance Testing**: Validate checkpoint impact during peak load periods

---
*This ADR documents the critical decision to implement automatic WAL checkpointing, preventing unbounded storage growth and ensuring production reliability of the EntityDB platform.*