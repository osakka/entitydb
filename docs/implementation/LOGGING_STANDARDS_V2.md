# EntityDB Logging Standards v2.0

## Overview
This document defines the comprehensive logging standards for EntityDB, ensuring consistent, actionable, and efficient logging across the entire codebase.

## Log Format
```
timestamp [pid:tid] [LEVEL] function.filename line: message
```

Example:
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
- **Purpose**: Normal operational events
- **Audience**: Operators, SREs, monitoring systems
- **Usage**: Service start/stop, major operations, periodic status
- **Performance**: Always enabled, must be lightweight

### WARN (3)
- **Purpose**: Potentially harmful situations
- **Audience**: Operators requiring attention
- **Usage**: Degraded performance, retry operations, near-limits
- **Performance**: Always enabled, actionable messages only

### ERROR (4)
- **Purpose**: Error events that don't stop the service
- **Audience**: Operators requiring immediate attention
- **Usage**: Failed operations, data inconsistencies, recoverable errors
- **Performance**: Always enabled, include error context

## Message Guidelines

### DO:
- Be concise and specific
- Include relevant context (IDs, counts, durations)
- Use consistent terminology
- Make messages actionable
- Consider the log consumer

### DON'T:
- Include redundant information (file/function already logged)
- Use inconsistent formats (SUCCESS:, [TAG], etc.)
- Log sensitive information
- Over-log at any level
- Include stack traces unless ERROR level

## Examples of Good vs Bad

### Bad:
```
[INFO] Starting server...
[DEBUG] In function processEntity
[ERROR] Error occurred!!!
[INFO] SUCCESS: Entity created
[WARN] [PERFORMANCE] Slow query detected
```

### Good:
```
[INFO ] Start.main 45: server listening on :8085
[DEBUG] processEntity.handler 123: processing entity id=abc123
[ERROR] Create.entity_handler 89: entity creation failed: duplicate id=abc123
[INFO ] Create.entity_handler 92: entity created id=def456 size=1024
[WARN ] Query.repository 234: query exceeded 5s threshold elapsed=7.2s
```

## Trace Subsystem Control

Trace logging can be enabled per subsystem:
```go
// Enable trace for specific subsystems
logger.EnableTrace("repository", "api", "auth")

// In code
logger.TraceIf("repository", "loading entity id=%s", entityID)
```

## Configuration

### Environment Variables
```bash
ENTITYDB_LOG_LEVEL=INFO        # Global log level
ENTITYDB_TRACE_SUBSYSTEMS=api,auth  # Comma-separated trace subsystems
```

### Command Line Flags
```bash
./entitydb --log-level=DEBUG --trace=api,auth
```

### API Control
```bash
curl -X POST /api/v1/system/log-level -d '{"level":"DEBUG","trace":["api","auth"]}'
```

## Thread Safety
- All logging operations are thread-safe
- Log buffer is lock-free for TRACE/DEBUG when disabled
- Atomic level checking prevents unnecessary formatting

## Performance Considerations
1. Level check happens before message formatting
2. Disabled levels have near-zero overhead
3. Trace subsystem check is O(1) using map lookup
4. No string concatenation when level is disabled

## Implementation Checklist
- [ ] Update logger.go with new format and trace subsystems
- [ ] Add API endpoint for dynamic log level control
- [ ] Update all log messages to follow standards
- [ ] Remove redundant context from messages
- [ ] Ensure proper level usage throughout codebase
- [ ] Add configuration via env vars and flags
- [ ] Document trace subsystems
- [ ] Performance test logging overhead