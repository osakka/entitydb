# EntityDB Logging Standards

## Overview
This document defines the logging standards for the EntityDB codebase. All code must follow these standards to ensure consistent, actionable, and efficient logging.

## Core Principles

1. **No Redundant Information**: Since the logger automatically includes file, function, and line number, never include this information in log messages.
2. **Appropriate Log Levels**: Use the correct log level for the situation.
3. **Actionable Messages**: Log messages should be specific and indicate what action should be taken if needed.
4. **Performance First**: Disabled log levels should incur zero CPU overhead.

## Log Levels

### TRACE
- **Purpose**: Extremely detailed information about program flow and data
- **Examples**: Variable values, data sizes, checksums, index entries
- **When to use**: 
  - Detailed data flow tracking
  - Variable values and sizes
  - Routine operations that succeed
  - Index/offset/position information
- **Example**: `logger.Trace("Entity %s content checksum: %x", entity.ID, checksum)`

### DEBUG
- **Purpose**: Information useful for debugging but not needed in production
- **Examples**: Function entry/exit, configuration values, state changes
- **When to use**:
  - Non-routine operations
  - State changes that might affect behavior
  - Configuration or initialization details
- **Example**: `logger.Debug("Session cleanup: removed %d expired sessions", count)`

### INFO
- **Purpose**: Important business events and state changes
- **Examples**: Service start/stop, major operations completed, user actions
- **When to use**:
  - Service lifecycle events
  - Major operations (batch jobs, migrations)
  - User-initiated actions that succeed
- **Example**: `logger.Info("Server started on port %d with SSL=%v", port, sslEnabled)`

### WARN
- **Purpose**: Unusual but recoverable conditions
- **Examples**: Missing optional config, fallback behavior, retries
- **When to use**:
  - Using default values due to missing config
  - Automatic recovery from errors
  - Deprecated feature usage
  - Invalid user input (not errors)
- **Example**: `logger.Warn("GetEntityAsOf: entity ID is missing in request")`

### ERROR
- **Purpose**: Errors that need attention but don't stop the service
- **Examples**: Failed operations, invalid data, integration failures
- **When to use**:
  - Operation failures that affect users
  - Data integrity issues
  - External service failures
- **Example**: `logger.Error("Failed to save entity %s: %v", entity.ID, err)`

### FATAL
- **Purpose**: Unrecoverable errors that will terminate the program
- **Examples**: Critical initialization failures, data corruption
- **When to use**: Only when the program cannot continue
- **Example**: `logger.Fatal("Failed to initialize database: %v", err)`

## Message Format Guidelines

### DO:
- Include relevant IDs, counts, and error details
- Use consistent terminology
- Make messages grep-friendly
- Include operation context

### DON'T:
- Include prefixes like [Module] or [Component]
- Log sensitive information (passwords, tokens)
- Use generic messages like "Error occurred"
- Include redundant information

## Examples of Good vs Bad

### Bad:
```go
logger.Error("[SecurityManager] Failed to authenticate user")
logger.Info("[EntityRepository] Entity created")
logger.Debug("Processing...")
logger.Error("Error")
```

### Good:
```go
logger.Error("Failed to authenticate user %s: invalid password", username)
logger.Debug("Entity created: id=%s, tags=%d, content_size=%d", entity.ID, len(entity.Tags), len(entity.Content))
logger.Trace("Processing batch %d/%d: %d items", current, total, itemCount)
logger.Error("Database connection failed after %d retries: %v", retries, err)
```

## Performance Guidelines

1. **Early Exit**: The logger checks if a level is enabled before processing
2. **Lazy Evaluation**: Don't pre-compute values for trace/debug logs
3. **Sampling**: For high-frequency operations, consider sampling

### Example of Lazy Evaluation:
```go
// Bad - always computes the expensive string
expensiveStr := computeExpensiveDebugInfo()
logger.Debug("Debug info: %s", expensiveStr)

// Good - only computes if DEBUG is enabled
if logger.IsDebugEnabled() {
    logger.Debug("Debug info: %s", computeExpensiveDebugInfo())
}
```

## Special Cases

### User Input Validation
- Use WARN for invalid input (not ERROR)
- Include what was expected
- Don't log the full invalid input if it could be large

### Loops and High-Frequency Operations
- Use TRACE for per-item logging
- Use DEBUG/INFO for summary at the end
- Consider sampling for very high frequencies

### Error Handling
- Always include the error value with %v
- Add context about what was being attempted
- Include relevant IDs or parameters

## Migration Checklist

When updating existing code:
1. Remove redundant prefixes ([Module], [Component])
2. Change INFO to DEBUG for routine operations
3. Change DEBUG to TRACE for detailed data
4. Ensure ERROR messages include context
5. Check that WARN is used for recoverable issues