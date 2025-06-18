# EntityDB Configuration Management Alignment Action Plan

> **Version**: v2.30.0 | **Created**: 2025-06-12 | **Status**: IMPLEMENTATION PLAN

## Overview

This document outlines the comprehensive action plan to align EntityDB's configuration management system with enterprise standards, ensuring zero hardcoded values and full three-tier configuration hierarchy compliance.

## Current State Analysis

### ✅ Already Compliant
- Three-tier configuration hierarchy: Database (highest) > Flags (medium) > Environment (lowest)
- Comprehensive environment file system with defaults and instance overrides
- ConfigManager with proper caching and thread safety
- Long flag naming convention (`--entitydb-*`) already implemented
- Essential short flags only (`-h`, `-v`) preserved

### ❌ Issues Identified

1. **Hardcoded Values Scattered Across Codebase (30+ files)**
   - Tools directory: Database paths, ports, filenames
   - API handlers: Default ports, file extensions
   - Storage layer: WAL file naming, index files
   - Documentation: Hardcoded swagger host values

2. **Missing Flag Coverage**
   - Some environment variables not exposed as command-line flags
   - Trace subsystems configuration incomplete
   - Metrics retention settings missing flags

3. **Runtime Script Inconsistencies**
   - entitydbd.sh has some hardcoded fallback values
   - Flag construction logic could be more robust

4. **Path Construction Issues**
   - Some paths built with string concatenation vs configuration methods
   - Database path derivation inconsistent across tools

## Implementation Plan

### Phase 1: Audit and Inventory (Day 1)
**Objective**: Complete systematic identification of all hardcoded values

#### Task 1.1: Comprehensive Hardcoded Value Audit
- [ ] Scan all `.go` files for hardcoded paths, ports, filenames, IDs
- [ ] Document every instance with file location and context
- [ ] Categorize by priority: Critical (paths/ports) vs Minor (display strings)
- [ ] Create replacement strategy for each category

#### Task 1.2: Flag Coverage Analysis
- [ ] Compare all environment variables against available flags
- [ ] Identify missing flag implementations
- [ ] Document flag naming convention compliance
- [ ] Plan new flag additions

#### Task 1.3: Script and Tool Analysis
- [ ] Audit runtime scripts for hardcoded values
- [ ] Review tools directory for configuration bypass
- [ ] Document non-compliant utilities
- [ ] Plan tool refactoring approach

### Phase 2: Core Configuration System Enhancement (Day 2)
**Objective**: Strengthen configuration foundation before mass refactoring

#### Task 2.1: Extend ConfigManager
```go
// Add missing configuration fields
type Config struct {
    // Database Configuration
    DatabaseFilename   string  // Currently hardcoded as "entities.db"
    WALSuffix         string  // Currently hardcoded as ".wal"
    IndexSuffix       string  // Currently hardcoded as ".idx"
    
    // Tool Configuration
    ToolsDataPath     string  // Default path for tools
    BackupPath        string  // Backup directory
    TempPath          string  // Temporary files
    
    // Development Configuration
    DevMode           bool    // Development mode flag
    DebugPort         int     // Debug server port
    ProfileEnabled    bool    // Enable profiling
}
```

#### Task 2.2: Add Missing Environment Variables
```bash
# Database Files
ENTITYDB_DATABASE_FILENAME=entities.db
ENTITYDB_WAL_SUFFIX=.wal
ENTITYDB_INDEX_SUFFIX=.idx

# Tool Paths
ENTITYDB_TOOLS_DATA_PATH=/opt/entitydb/var
ENTITYDB_BACKUP_PATH=/opt/entitydb/backup
ENTITYDB_TEMP_PATH=/opt/entitydb/tmp

# Development
ENTITYDB_DEV_MODE=false
ENTITYDB_DEBUG_PORT=6060
ENTITYDB_PROFILE_ENABLED=false
```

#### Task 2.3: Complete Flag Implementation
- [ ] Add flags for all new configuration options
- [ ] Ensure 100% environment variable coverage
- [ ] Update flag registration in ConfigManager
- [ ] Test flag precedence hierarchy

### Phase 3: Systematic Hardcoded Value Elimination (Days 3-4)
**Objective**: Remove all hardcoded values throughout codebase

#### Task 3.1: Core Server Components
**Priority: Critical**

Files to update:
- `src/main.go`: Remove any remaining hardcoded values
- `src/storage/binary/*.go`: Replace hardcoded file extensions and paths
- `src/api/*.go`: Replace hardcoded ports and paths

**Implementation Pattern:**
```go
// Before (hardcoded)
walPath := dbPath + ".wal"

// After (configurable)
walPath := dbPath + config.WALSuffix
```

#### Task 3.2: Tools Directory Refactoring
**Priority: High**

All tools in `src/tools/` must use ConfigManager:
```go
// Standard tool pattern
func main() {
    configManager := config.NewConfigManager(nil)
    configManager.RegisterFlags()
    flag.Parse()
    
    cfg, err := configManager.Initialize()
    if err != nil {
        log.Fatal(err)
    }
    
    // Use cfg throughout tool
    dbPath := cfg.DatabasePath()
}
```

#### Task 3.3: Documentation Updates
**Priority: Medium**

- Update swagger documentation with configurable hosts
- Remove hardcoded examples from API documentation
- Update all documentation to reference configuration variables

### Phase 4: Runtime Script Enhancement (Day 5)
**Objective**: Ensure runtime scripts fully leverage configuration system

#### Task 4.1: entitydbd.sh Enhancement
- [ ] Remove all hardcoded fallback values
- [ ] Implement robust environment variable loading
- [ ] Add validation for required configuration
- [ ] Improve flag construction logic

**Enhanced Script Pattern:**
```bash
# Load configuration with proper precedence
load_configuration() {
    # 1. Load defaults from share/config/entitydb.env
    # 2. Override with instance config from var/entitydb.env
    # 3. Validate required values
    # 4. Construct flags dynamically
}

# Dynamic flag construction
build_server_args() {
    local args=""
    
    # Only add flags if environment variables are set
    [ -n "$ENTITYDB_PORT" ] && args="$args --entitydb-port $ENTITYDB_PORT"
    [ -n "$ENTITYDB_SSL_PORT" ] && args="$args --entitydb-ssl-port $ENTITYDB_SSL_PORT"
    # ... continue for all variables
    
    echo "$args"
}
```

#### Task 4.2: Additional Scripts
- [ ] Update any other shell scripts in bin/ directory
- [ ] Ensure Makefile uses configuration-driven paths
- [ ] Update development setup scripts

### Phase 5: Validation and Testing (Day 6)
**Objective**: Comprehensive validation of configuration system

#### Task 5.1: Configuration Hierarchy Testing
- [ ] Test environment variable loading
- [ ] Test flag override behavior
- [ ] Test database configuration priority
- [ ] Validate caching and refresh mechanisms

#### Task 5.2: Zero Hardcoded Value Verification
```bash
# Automated verification script
check_hardcoded_values() {
    echo "Scanning for hardcoded values..."
    
    # Check for hardcoded ports
    grep -r "8085\|8443" src/ && echo "FAIL: Hardcoded ports found"
    
    # Check for hardcoded paths
    grep -r '"/opt/\|"/var/\|"/tmp/' src/ && echo "FAIL: Hardcoded paths found"
    
    # Check for hardcoded file extensions
    grep -r '\.db"\|\.wal"\|\.log"\|\.pid"' src/ && echo "FAIL: Hardcoded extensions found"
    
    echo "Verification complete"
}
```

#### Task 5.3: Integration Testing
- [ ] Test complete server startup with various configurations
- [ ] Test tool functionality with different config sources
- [ ] Test script behavior with missing/invalid configurations
- [ ] Performance test configuration loading overhead

### Phase 6: Documentation and Migration (Day 7)
**Objective**: Complete documentation and provide migration guidance

#### Task 6.1: Configuration Documentation
- [ ] Complete configuration reference with all new options
- [ ] Document configuration hierarchy behavior
- [ ] Provide examples for each configuration method
- [ ] Document best practices and security considerations

#### Task 6.2: Migration Guide
- [ ] Create upgrade guide for existing deployments
- [ ] Document breaking changes (if any)
- [ ] Provide configuration validation tools
- [ ] Include troubleshooting guide

## Implementation Standards

### Configuration Naming Convention
```
Format: ENTITYDB_{CATEGORY}_{SETTING}
Examples:
  ENTITYDB_SERVER_PORT=8085
  ENTITYDB_DATABASE_FILENAME=entities.db
  ENTITYDB_STORAGE_WAL_SUFFIX=.wal
  ENTITYDB_TOOLS_BACKUP_PATH=/opt/entitydb/backup
```

### Flag Naming Convention
```
Format: --entitydb-{category}-{setting}
Examples:
  --entitydb-server-port=8085
  --entitydb-database-filename=entities.db
  --entitydb-storage-wal-suffix=.wal
  --entitydb-tools-backup-path=/opt/entitydb/backup
```

### Code Implementation Standards

#### Configuration Access Pattern
```go
// ✅ Correct: Use configuration
dbPath := config.DatabasePath()
walFile := dbPath + config.WALSuffix

// ❌ Wrong: Hardcoded values
walFile := dbPath + ".wal"
```

#### Tool Configuration Pattern
```go
func main() {
    // Every tool must start with this pattern
    configManager := config.NewConfigManager(nil)
    configManager.RegisterFlags()
    flag.Parse()
    
    cfg, err := configManager.Initialize()
    if err != nil {
        log.Fatal("Configuration error: %v", err)
    }
    
    // Use cfg throughout
}
```

## Success Criteria

### Verification Checklist
- [ ] Zero hardcoded paths in any `.go` file
- [ ] Zero hardcoded ports in any `.go` file  
- [ ] Zero hardcoded filenames/extensions in any `.go` file
- [ ] All environment variables have corresponding flags
- [ ] All tools use ConfigManager
- [ ] Runtime scripts are fully configuration-driven
- [ ] Documentation is complete and accurate
- [ ] Automated verification passes
- [ ] Integration tests pass
- [ ] Performance impact is minimal (<1ms config loading)

### Post-Implementation Benefits
1. **Operational Flexibility**: All aspects configurable without code changes
2. **Environment Portability**: Easy deployment across dev/staging/production
3. **Security Compliance**: No hardcoded credentials or paths
4. **Maintenance Simplicity**: Single source of truth for all configuration
5. **Debugging Capability**: Clear configuration precedence and logging
6. **Testing Enhancement**: Easy configuration override for tests

## Timeline
- **Day 1**: Phase 1 (Audit and Inventory)
- **Day 2**: Phase 2 (Core System Enhancement)
- **Days 3-4**: Phase 3 (Hardcoded Value Elimination)
- **Day 5**: Phase 4 (Runtime Script Enhancement)
- **Day 6**: Phase 5 (Validation and Testing)
- **Day 7**: Phase 6 (Documentation and Migration)

**Total Duration**: 7 days
**Risk Level**: Low (non-breaking changes, incremental implementation)
**Resource Requirements**: 1 developer, comprehensive testing environment

## Next Steps

1. **Approve Action Plan**: Review and approve this implementation plan
2. **Begin Phase 1**: Start with comprehensive audit of hardcoded values
3. **Set Up Tracking**: Use project management tools to track progress
4. **Establish Testing**: Prepare configuration test environments
5. **Documentation Prep**: Set up documentation structure for updates

---

**Implementation Ready**: This action plan provides clear, pragmatic steps for complete configuration management alignment with zero hardcoded values and full three-tier hierarchy compliance.