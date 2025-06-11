# EntityDB Logging Standards

> **Version**: 2.0  
> **Status**: Implemented in v2.27.0  
> **Component**: Professional logging system

## Overview

EntityDB implements comprehensive logging standards ensuring consistent, actionable, and efficient logging across the entire codebase. The logging system provides structured output with contextual information and appropriate log levels for development and production environments.

## Log Format

```
timestamp [pid:tid] [LEVEL] function.filename line: message
```

**Example**:
```
2025/05/31 13:26:30.749793 [2615465:1] [INFO ] AddTag.entity_repository 1376: WAL checkpoint completed
```

## Log Levels

### TRACE (0)
- **Purpose**: Detailed data flow tracing for development
- **Audience**: Developers during debugging
- **Usage**: Function entry/exit, variable values, detailed state changes
- **Performance**: Disabled in production, zero overhead when off
- **Subsystem Control**: Can be enabled per functionality

### DEBUG (1)
- **Purpose**: Diagnostic information for troubleshooting
- **Audience**: Developers and advanced operators
- **Usage**: Algorithm decisions, configuration details, non-critical errors
- **Performance**: Minimal overhead, typically disabled in production

### INFO (2)
- **Purpose**: General operational information
- **Audience**: Operations teams and system administrators
- **Usage**: System startup, configuration changes, normal operations
- **Performance**: Low overhead, safe for production

### WARN (3)
- **Purpose**: Warning conditions that should be investigated
- **Audience**: Operations teams
- **Usage**: Recoverable errors, performance degradation, configuration issues
- **Performance**: Minimal overhead

### ERROR (4)
- **Purpose**: Error conditions requiring immediate attention
- **Audience**: Operations teams and developers
- **Usage**: Failed operations, data corruption, security violations
- **Performance**: No overhead concerns

## Trace Subsystems

EntityDB supports granular trace logging by subsystem:

### Available Subsystems
- `auth` - Authentication and authorization flow
- `storage` - Storage operations and transactions
- `cache` - Cache operations and hit/miss statistics
- `temporal` - Temporal operations and indexing
- `lock` - Lock acquisition and contention
- `query` - Query execution and optimization
- `metrics` - Metrics collection and aggregation
- `dataset` - Dataset operations and isolation
- `relationship` - Entity relationships and graph operations
- `chunking` - Content chunking and streaming

### Configuration

**Environment Variable**:
```bash
ENTITYDB_TRACE_SUBSYSTEMS=auth,storage,temporal
```

**API Control**:
```bash
curl -X POST https://localhost:8085/api/v1/admin/trace-subsystems \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"subsystems": ["auth", "storage", "temporal"]}'
```

## Implementation Standards

### Manual Prefixes Removal

**❌ Before (redundant prefixes)**:
```go
logger.Info("[API] Processing entity creation request for entity ID: %s", entityID)
```

**✅ After (logger provides context)**:
```go
logger.Info("Processing entity creation request for entity ID: %s", entityID)
```

The logger automatically provides file, function, and line information.

### Contextual Error Messages

**❌ Before (generic)**:
```go
logger.Error("Failed to create entity")
```

**✅ After (contextual)**:
```go
logger.Error("Failed to create entity %s in dataset %s: %v", entityID, dataset, err)
```

### Appropriate Log Levels

**❌ Before (wrong levels)**:
```go
logger.Debug("Authentication failed for user %s", username)  // Should be WARN
logger.Info("Reading entity from WAL offset %d", offset)      // Should be TRACE
```

**✅ After (correct levels)**:
```go
logger.Warn("Authentication failed for user %s: %v", username, err)
logger.Trace("Reading entity from WAL offset %d", offset)
```

### Structured Logging

```go
// Use structured fields for important operations
logger.WithFields(map[string]interface{}{
    "entity_id": entityID,
    "dataset": dataset,
    "operation": "create",
    "user_id": userID,
}).Info("Entity created successfully")
```

## Performance Guidelines

### Production Logging

- **INFO and above** enabled in production
- **DEBUG** disabled by default (enable for troubleshooting)
- **TRACE** always disabled in production

### Storage Layer Optimization

**Before**: Excessive INFO logging in hot paths
```go
logger.Info("Writing entity %s to storage", entityID)  // Too verbose
```

**After**: Moved to TRACE level
```go
logger.Trace("Writing entity %s to storage", entityID)  // Appropriate level
```

### Change-Only Detection

For metrics and repetitive operations:
```go
if lastLoggedValue != currentValue {
    logger.Info("Metric value changed: %s = %v", metricName, currentValue)
    lastLoggedValue = currentValue
}
```

## Error Handling Standards

### Error Context

Always include relevant context in error messages:
```go
func (r *EntityRepository) GetByID(id string) (*Entity, error) {
    entity, err := r.storage.Read(id)
    if err != nil {
        logger.Error("Failed to read entity %s from storage: %v", id, err)
        return nil, fmt.Errorf("entity retrieval failed for ID %s: %w", id, err)
    }
    return entity, nil
}
```

### Security Considerations

**❌ Never log sensitive data**:
```go
logger.Debug("User credentials: %+v", credentials)  // DANGEROUS
```

**✅ Log operation results**:
```go
logger.Info("Authentication completed for user %s", username)
```

## Configuration

### Log Level Control

**Environment Variable**:
```bash
ENTITYDB_LOG_LEVEL=info
```

**Runtime Control**:
```bash
# Get current log level
curl https://localhost:8085/api/v1/admin/log-level \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Set log level
curl -X POST https://localhost:8085/api/v1/admin/log-level \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"level": "debug"}'
```

### HTTP Request Tracing

```bash
ENTITYDB_HTTP_TRACE=true
```

Enables detailed HTTP request/response logging for debugging API issues.

## Implementation Files

### Core Logger
- `src/logger/logger.go` - Main logging interface
- `src/logger/trace.go` - Trace subsystem management
- `src/logger/log_bridge.go` - Bridge for legacy code

### Middleware Integration
- `src/api/trace_middleware.go` - HTTP request tracing
- `src/api/request_metrics_middleware.go` - Request logging with metrics

## Migration Guidelines

### Updating Existing Code

1. **Remove manual prefixes** since logger provides context automatically
2. **Add contextual information** to error messages (entity IDs, user context, operation details)
3. **Fix log levels** (move detailed operations to TRACE, errors to appropriate levels)
4. **Use structured logging** for complex operations

### Example Migration

**Before**:
```go
func CreateEntity(entity *Entity) error {
    log.Printf("[CREATE] Starting entity creation")
    if err := validate(entity); err != nil {
        log.Printf("[ERROR] Validation failed")
        return err
    }
    log.Printf("[CREATE] Entity created successfully")
    return nil
}
```

**After**:
```go
func CreateEntity(entity *Entity) error {
    logger.Trace("Starting entity creation for ID %s", entity.ID)
    if err := validate(entity); err != nil {
        logger.Error("Entity validation failed for ID %s: %v", entity.ID, err)
        return err
    }
    logger.Info("Entity %s created successfully in dataset %s", entity.ID, entity.Dataset)
    return nil
}
```

## Quality Assurance

### Logging Audit

Regular audits ensure:
- ✅ No redundant manual prefixes
- ✅ Appropriate log levels throughout codebase
- ✅ Contextual error messages with relevant details
- ✅ No sensitive data in logs
- ✅ Performance-optimized logging in hot paths

### Testing

```bash
# Run logging standards validation
cd src && go run tools/validate_logging.go

# Check for inappropriate log levels
grep -r "logger.Debug.*failed\|logger.Info.*reading.*WAL" .
```

## Troubleshooting

### Common Issues

1. **Performance degradation**: Check for excessive INFO/DEBUG logging in hot paths
2. **Missing context**: Add entity IDs, user context, operation details to error messages
3. **Log spam**: Use change-only detection for repetitive operations

### Debug Commands

```bash
# Enable trace logging for authentication
curl -X POST https://localhost:8085/api/v1/admin/trace-subsystems \
  -d '{"subsystems": ["auth"]}'

# Check current logging configuration
curl https://localhost:8085/api/v1/admin/log-level
```

---

This logging standard ensures EntityDB provides comprehensive, efficient, and maintainable logging suitable for both development and production environments.