# EntityDB Logging Standards

## Overview

This document defines the logging standards for EntityDB to ensure consistent, actionable, and performant logging across the entire codebase.

## Log Format

All logs follow this standardized format:
```
timestamp [processid:threadid] [level] functionname.filename:line: message
```

Example:
```
2025/06/07 14:30:45.123456 [1234:5] [INFO] CreateEntity.entity_handler:123: entity created with id abc123
```

## Log Levels

### TRACE
- **Purpose**: Detailed execution flow for debugging specific subsystems
- **Audience**: Developers during development/debugging
- **Examples**:
  - Function entry/exit with parameters
  - Detailed state changes
  - Lock acquisition/release
  - Cache hits/misses
- **Enabled**: Only when specific subsystem tracing is enabled

### DEBUG  
- **Purpose**: Diagnostic information for development
- **Audience**: Developers
- **Examples**:
  - Successful routine operations
  - Query execution details
  - Configuration values loaded
  - Connection establishment
- **Production**: Usually disabled

### INFO
- **Purpose**: Important business events and state changes
- **Audience**: Operations teams, SREs
- **Examples**:
  - Server startup/shutdown
  - Configuration changes
  - Major state transitions
  - Security events (login attempts)
- **Production**: Default level

### WARN
- **Purpose**: Potentially harmful situations that don't prevent operation
- **Audience**: Operations teams, SREs
- **Examples**:
  - Deprecated feature usage
  - Performance degradation
  - Retryable errors
  - Missing optional configuration
- **Action Required**: Investigation may be needed

### ERROR
- **Purpose**: Error events that affect operation
- **Audience**: Operations teams, SREs
- **Examples**:
  - Failed operations
  - Unrecoverable errors
  - Data integrity issues
  - Security violations
- **Action Required**: Immediate investigation needed

## Message Guidelines

### DO:
- Keep messages concise and actionable
- Include relevant context (IDs, counts, durations)
- Use consistent terminology
- Write for the target audience
- Use proper grammar (lowercase start, no trailing punctuation)

### DON'T:
- Use prefixes like [COMPONENT] or SUCCESS: or FAILED:
- Include function/file names in the message (logger provides these)
- Log routine successes at INFO level
- Use inconsistent formatting
- Include sensitive information (passwords, tokens)

## Examples

### Good Examples:
```go
// INFO - Important state change
logger.Info("server started on port %d with ssl %s", port, sslStatus)

// DEBUG - Routine operation
logger.Debug("entity created with id %s", entity.ID)

// WARN - Degraded operation
logger.Warn("cache miss rate %.2f%% exceeds threshold", missRate)

// ERROR - Failed operation with context
logger.Error("failed to save entity %s: %v", entityID, err)

// TRACE - Detailed flow
logger.TraceIf("storage", "acquiring write lock for entity %s", entityID)
```

### Bad Examples:
```go
// Don't use prefixes
logger.Info("[AuthHandler] User logged in")  // Bad

// Don't log routine operations at INFO
logger.Info("Successfully created entity")  // Bad - should be DEBUG

// Don't use inconsistent formatting  
logger.Error("FAILED: Could not save entity")  // Bad

// Don't forget context
logger.Error("Operation failed")  // Bad - what operation? what failed?
```

## Trace Subsystems

Trace logging is controlled per subsystem for fine-grained debugging:

### Available Subsystems:
- `auth` - Authentication and authorization flow
- `storage` - Storage operations and transactions
- `cache` - Cache operations
- `temporal` - Temporal operations and indexing
- `lock` - Lock acquisition and contention
- `query` - Query execution and optimization
- `metrics` - Metrics collection
- `dataspace` - Dataspace operations
- `relationship` - Entity relationships
- `chunking` - Content chunking operations

### Usage:
```go
// Enable trace for specific subsystem
logger.TraceIf("storage", "committing transaction %s with %d operations", txID, opCount)
```

## Configuration

### Environment Variables:
- `ENTITYDB_LOG_LEVEL` - Set global log level (TRACE, DEBUG, INFO, WARN, ERROR)
- `ENTITYDB_TRACE_SUBSYSTEMS` - Comma-separated list of trace subsystems

### Command Line:
- `--entitydb-log-level` - Set log level
- `--entitydb-trace-subsystems` - Enable trace subsystems

### Runtime API:
- `POST /api/v1/admin/log-level` - Change log level
- `POST /api/v1/admin/trace-subsystems` - Enable/disable trace subsystems

## Performance Considerations

- Log level checks use atomic operations (near-zero overhead when disabled)
- Messages are only formatted if the level is enabled
- Trace subsystem checks are cached with RWMutex for performance
- No heap allocations for disabled log levels

## Thread Safety

The logger is fully thread-safe:
- Atomic log level checking
- Mutex-protected trace subsystem management
- Thread-safe output formatting
- Goroutine ID extraction for thread identification

## Migration Guide

When updating existing code:
1. Remove manual prefixes like `[Component]`
2. Move routine success messages from INFO to DEBUG
3. Ensure error messages include context
4. Use appropriate log levels based on audience
5. Enable trace subsystems instead of debug logging for detailed flow