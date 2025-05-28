# EntityDB Logging Standards

## Log Format
```
timestamp [LEVEL] [file] [function] [line] message
```
Example:
```
2025/05/28 12:30:45.123456 [INFO] [entity_repository.go] [Create] [245] Entity created: user_123
```

## Log Levels and Usage

### TRACE (Development Only)
- **Purpose**: Detailed data flow and variable values
- **Audience**: Developers debugging specific issues
- **Examples**:
  - Entering/exiting functions with parameters
  - Loop iterations with values
  - Intermediate calculation results
- **Guidelines**: 
  - Include actual data values
  - Use for complex algorithms
  - Never in production

### DEBUG
- **Purpose**: Diagnostic information for troubleshooting
- **Audience**: Developers and advanced SREs
- **Examples**:
  - Configuration loaded
  - Cache hits/misses
  - Non-critical retry attempts
- **Guidelines**:
  - Include relevant IDs and counts
  - Avoid sensitive data
  - Keep concise

### INFO
- **Purpose**: Normal operational events
- **Audience**: SREs and operations teams
- **Examples**:
  - Service started/stopped
  - Database connections established
  - Periodic statistics (entities created, indexed)
- **Guidelines**:
  - Focus on state changes
  - Include success metrics
  - Be meaningful to ops

### WARN
- **Purpose**: Potentially harmful situations
- **Audience**: SREs and operations teams
- **Examples**:
  - Degraded performance
  - Retry succeeded after failures
  - Configuration fallbacks used
- **Guidelines**:
  - Include impact assessment
  - Suggest remediation if possible
  - Be actionable

### ERROR
- **Purpose**: Error events but application continues
- **Audience**: SREs and operations teams
- **Examples**:
  - Failed operations (with recovery)
  - Invalid input rejected
  - External service failures
- **Guidelines**:
  - Include error details
  - Log once per incident
  - Include correlation IDs

### FATAL/PANIC
- **Purpose**: Application cannot continue
- **Audience**: SREs and operations teams
- **Examples**:
  - Critical resource unavailable
  - Data corruption detected
  - Unrecoverable state
- **Guidelines**:
  - Include full context
  - Log before termination
  - Be explicit about impact

## Message Guidelines

### DO:
- Be concise and specific
- Include relevant IDs (entity ID, user ID)
- Use consistent terminology
- Include counts and metrics
- Focus on "what happened"

### DON'T:
- Log sensitive data (passwords, tokens)
- Include redundant information (already in file/func/line)
- Use vague messages ("something went wrong")
- Over-log in loops (use counters)
- Mix user output with logs in tools

## Component-Specific Patterns

### Storage Layer
```go
// Good
logger.Debug("Building index for %d entities", count)
logger.Info("Index rebuild completed: %d entities, %d tags", entityCount, tagCount)
logger.Warn("Index corruption detected, rebuilding from WAL")

// Bad
logger.Debug("Starting to build indexes...") // Too vague
logger.Info("Built index") // Missing metrics
```

### API Handlers
```go
// Good
logger.Debug("Create entity request: type=%s, tags=%d", entityType, len(tags))
logger.Info("Entity created: %s", entity.ID)
logger.Error("Entity creation failed: %v", err)

// Bad
logger.Info("Processing request") // Which request?
logger.Error("Error: %v", err) // What operation failed?
```

### Tools and Utilities
```go
// For tools, separate user output from logs:
fmt.Printf("Processing %d entities...\n", count) // User feedback
logger.Debug("Batch processing started: size=%d", count) // Diagnostic log
```

## Performance Considerations

- Log level checks happen before message formatting
- When disabled, TRACE/DEBUG have near-zero overhead
- Use lazy evaluation for expensive operations:
  ```go
  if logger.GetLogLevel() == "TRACE" {
      logger.Trace("Complex data: %s", expensiveToString())
  }
  ```

## Migration Strategy

1. Update all direct `log.Print` to use `logger` package
2. Convert `fmt.Print` in application code to appropriate log levels
3. Tools keep `fmt.Print` for user output, add `logger` for diagnostics
4. Remove redundant prefixes from messages
5. Ensure consistent ID formatting (entity_123, user_456)