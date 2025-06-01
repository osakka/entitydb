# EntityDB Logging Migration Plan

## Overview
This document outlines the plan to migrate EntityDB to the new unified logging standards.

## Current State Analysis
- **Total Issues Found**: 408
- **Most Common Issues**:
  - Inappropriate log levels (391 occurrences)
  - Multiple consecutive spaces (14 occurrences)  
  - PREFIX: format (2 occurrences)
  - Long messages (1 occurrence)

## Migration Steps

### Phase 1: Update Logger Implementation
1. ✅ Created enhanced logger (logger_v2.go) with:
   - Proper format: `timestamp [pid:tid] [LEVEL] function.filename line: message`
   - Trace subsystem support
   - Thread-safe operations
   - Zero overhead for disabled levels

2. ✅ Added API endpoint for dynamic log control
   - GET /api/v1/system/log-level
   - POST /api/v1/system/log-level

3. TODO: Replace old logger with new implementation

### Phase 2: Fix Log Level Usage

#### TRACE → DEBUG conversions:
- Authentication flow details
- Request/response bodies
- Algorithm decisions

#### DEBUG → TRACE conversions:
- Function entry/exit
- Variable dumps
- Loop iterations

#### INFO level fixes:
- Remove "authenticated successfully" (redundant with session creation)
- Keep only major operations (server start, entity created, etc.)

#### WARN level fixes:
- "Entity not found" → DEBUG (normal operation)
- Keep only degraded conditions

#### ERROR level fixes:
- "Repository doesn't support" → WARN (configuration issue)
- Keep only actual failures

### Phase 3: Message Cleanup

1. **Remove redundant context**:
   - "[AuthHandler]" prefixes (function already logged)
   - "In function X" messages
   - File paths in messages

2. **Standardize messages**:
   - "Entity created successfully" → "entity created id=%s size=%d"
   - "Failed to create entity" → "entity creation failed: %v"
   - "Processing request" → "processing %s request id=%s"

3. **Fix formatting**:
   - Remove multiple spaces
   - Remove PREFIX: format
   - Shorten long messages

### Phase 4: Add Trace Subsystems

Define subsystems for targeted debugging:
- `api` - HTTP request/response flow
- `auth` - Authentication and authorization
- `storage` - Binary storage operations
- `repository` - Entity CRUD operations
- `temporal` - Temporal queries
- `index` - Index operations
- `wal` - Write-ahead log
- `cache` - Caching operations
- `rbac` - Permission checks

### Phase 5: Documentation and Testing

1. Update developer documentation
2. Add logging examples
3. Test performance impact
4. Verify log output format

## Automated Fixes

Common replacements:
```bash
# Remove [TAG] prefixes
sed -i 's/\[AuthHandler\] //g' *.go

# Fix multiple spaces
sed -i 's/  */ /g' *.go

# Standardize success messages
sed -i 's/successfully created/created/g' *.go
sed -i 's/successfully updated/updated/g' *.go

# Remove ellipsis
sed -i 's/\.\.\."/"/' *.go
```

## Success Criteria

1. All log messages follow the standard format
2. Log levels appropriate for audience
3. Messages are concise and actionable
4. Trace subsystems defined and documented
5. Zero performance impact when logging disabled
6. Dynamic log control via API working

## Timeline

- Phase 1: ✅ Complete
- Phase 2: 2 days (fix log levels)
- Phase 3: 2 days (cleanup messages)
- Phase 4: 1 day (add trace subsystems)
- Phase 5: 1 day (documentation/testing)

Total: 6 days estimated