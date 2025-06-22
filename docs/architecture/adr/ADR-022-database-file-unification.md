# ADR-022: Database File Path Unification and Extension Standardization

**Status:** Implemented  
**Date:** 2025-06-20  
**Context:** EntityDB v2.32.5 Configuration Architecture Optimization  

## Summary

This ADR documents the comprehensive unification of database file path configuration to achieve single source of truth compliance and standardize on the `.edb` file extension for EntityDB databases.

## Problem Statement

The EntityDB codebase suffered from significant violations of the "single source of truth" principle in database file path management:

### Critical Issues Identified

1. **Multiple Configuration Parameters**
   - `DatabaseFilename` (intended for unified format)
   - `UnifiedFilename` (duplicate purpose)
   - `DatabaseBaseFilename` (legacy filename only)
   - `DatabasePath()` method (path construction)

2. **Hardcoded Paths Throughout Codebase**
   - API handlers using `/opt/entitydb/var/entities.db`
   - Metrics collectors hardcoding file extensions
   - Tools bypassing configuration system
   - Storage layer fallback paths

3. **File Extension Inconsistency**
   - `.db` (legacy references)
   - `.ebf` (EntityDB Binary Format)
   - `.unified` (unified format files)
   - `.edb` (intended standard)

4. **Configuration Mismatch**
   - Repository initialization using `UnifiedFilename`
   - `getDataFile()` method returning `DatabaseFilename`
   - Reader/writer using different file sources

## Decision

### 1. Configuration Unification

**Consolidated to Single Parameter:**
- **Primary:** `DatabaseFilename` - Complete file path including extension
- **Removed:** `UnifiedFilename`, `DatabaseBaseFilename`, `DatabasePath()` method

**Environment Variable:**
```bash
ENTITYDB_DATABASE_FILE="./var/entities.edb"
```

### 2. Extension Standardization

**Standardized on `.edb` Extension:**
- Represents "EntityDB Database"
- Distinguishes from generic `.db` files
- Clearly identifies EntityDB binary format files
- Consistent across all tools and documentation

### 3. API Handler Architecture Update

**Configuration Injection Pattern:**
```go
// Before: Hardcoded paths
os.Stat("/opt/entitydb/var/entities.db")

// After: Configuration-based
os.Stat(h.config.DatabaseFilename)
```

**Updated Handlers:**
- `HealthHandler` - Added config parameter
- `MetricsHandler` - Added config parameter  
- `SystemMetricsHandler` - Added config parameter
- `BackgroundMetricsCollector` - Added config parameter

### 4. Tool Modernization

**Configuration-Based Tools:**
- All analysis tools now use `config.Load()`
- Maintenance tools accept database file path directly
- Emergency tools use configuration system

## Implementation Details

### Configuration Structure

```go
type Config struct {
    // Single source of truth for database file path
    DatabaseFilename string `env:"ENTITYDB_DATABASE_FILE" default:"./var/entities.edb"`
    
    // Related file paths (derived or configured separately)
    WALFilename     string `env:"ENTITYDB_WAL_FILE" default:"./var/entitydb.wal"`
    IndexFilename   string `env:"ENTITYDB_INDEX_FILE" default:"./var/entities.edb.idx"`
    
    // Removed: UnifiedFilename, DatabaseBaseFilename, DatabasePath() method
}
```

### Repository Architecture Update

```go
// Before: Mixed sources
databasePath := cfg.UnifiedFilename  // Constructor
return r.config.DatabaseFilename     // getDataFile()

// After: Single source
databasePath := cfg.DatabaseFilename // Constructor  
return r.config.DatabaseFilename     // getDataFile()
```

### API Handler Pattern

```go
// Before: Hardcoded
func NewHealthHandler(repo models.EntityRepository) *HealthHandler

// After: Configuration injection
func NewHealthHandler(repo models.EntityRepository, cfg *config.Config) *HealthHandler
```

## Benefits Achieved

### 1. Single Source of Truth Compliance
- ✅ One configuration parameter for database file path
- ✅ All components use same configuration source
- ✅ No hardcoded paths in production code
- ✅ Consistent behavior across deployment environments

### 2. Operational Excellence
- ✅ Flexible deployment paths via environment variables
- ✅ Simplified configuration management
- ✅ Reduced maintenance burden
- ✅ Improved testability

### 3. Developer Experience  
- ✅ Clear naming conventions
- ✅ Consistent API patterns
- ✅ Self-documenting configuration
- ✅ Predictable file location behavior

### 4. System Reliability
- ✅ Eliminated configuration drift
- ✅ Prevented deployment path conflicts
- ✅ Improved error diagnostics
- ✅ Consistent monitoring across environments

## Migration Strategy

### Phase 1: Configuration Consolidation
1. Remove duplicate configuration parameters
2. Update repository initialization
3. Verify single source compliance

### Phase 2: API Handler Updates
1. Add configuration injection to all handlers
2. Replace hardcoded paths with config references
3. Update constructor calls in main.go

### Phase 3: Tool Modernization
1. Update analysis tools to use configuration
2. Modernize maintenance tool interfaces
3. Verify tool consistency

### Phase 4: Validation
1. Clean build verification
2. Runtime testing with different paths
3. Configuration compliance audit

## Technical Verification

### Compile-Time Verification
```bash
cd /opt/entitydb/src && make
# Result: Clean build with zero warnings
```

### Runtime Verification
```bash
ENTITYDB_DATABASE_FILE="/custom/path/entities.edb" ./bin/entitydb
# Result: Uses custom path consistently across all components
```

### Configuration Compliance Audit
- ✅ No hardcoded `/opt/entitydb/var/` paths in production code
- ✅ No `entities.db`, `entities.ebf`, `entities.unified` hardcoded references
- ✅ All API handlers use configuration properly
- ✅ No duplicate configuration fields remaining

## Breaking Changes

### Configuration
- **Removed:** `ENTITYDB_UNIFIED_FILE` environment variable
- **Removed:** `ENTITYDB_DATABASE_FILENAME` environment variable  
- **Removed:** `--entitydb-database-filename` CLI flag
- **Primary:** `ENTITYDB_DATABASE_FILE` environment variable

### API (Internal Only)
- Handler constructors now require config parameter
- Tools may require path parameter instead of directory

### Files
- Default database file extension changed from `.db` to `.edb`
- Unified format now consistently uses `.edb` extension

## Rollback Plan

If rollback is required:
1. Restore duplicate configuration parameters
2. Revert API handler constructors
3. Restore hardcoded path fallbacks
4. Update file extensions back to mixed format

## Monitoring

### Success Metrics
- ✅ Zero configuration-related deployment issues
- ✅ Consistent file path behavior across environments
- ✅ Clean build without warnings
- ✅ Successful configuration compliance audit

### Failure Detection
- Configuration drift detection via startup logging
- File path consistency verification in health checks
- Environment variable validation in configuration loading

## Future Considerations

### Path Validation
- Consider adding configuration validation for path accessibility
- Implement path permission checking during startup
- Add configuration testing framework

### Documentation
- Update all deployment guides for new extension
- Create configuration migration guide
- Document environment variable standards

## Conclusion

The database file path unification successfully achieves single source of truth compliance while standardizing on the `.edb` extension. This architectural improvement provides:

- **Operational Excellence:** Consistent behavior across deployments
- **Developer Productivity:** Clear configuration patterns
- **System Reliability:** Eliminated configuration conflicts
- **Future Readiness:** Extensible configuration framework

The implementation maintains backward compatibility where possible while establishing a clean foundation for future EntityDB development.

---

**Related ADRs:**
- ADR-020: Configuration Management Overhaul
- ADR-021: Critical Corruption Prevention Fix

**Implementation Commits:**
- Configuration unification and API handler updates
- Tool modernization and hardcoded path elimination
- Single source of truth compliance verification