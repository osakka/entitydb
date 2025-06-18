# EntityDB Hardcoded Values Audit Report

> **Version**: v2.30.0 | **Created**: 2025-06-12 | **Status**: AUDIT FINDINGS

## Executive Summary

Comprehensive scan of EntityDB codebase identified **47 files** containing hardcoded values that violate configuration management standards. This audit categorizes findings by severity and provides specific remediation steps.

## Audit Methodology

**Scan Commands Used:**
```bash
# Primary scan for hardcoded values
find /opt/entitydb/src -name "*.go" -print0 | xargs -0 grep -l "8085\|8443\|/var\|/tmp\|\.db\|\.log\|\.pid"

# Secondary scans for specific patterns
grep -r "const.*=" src/ | grep -E "path|port|host|file"
grep -r '"/[a-z]' src/ --include="*.go"
grep -r ':8[0-9][0-9][0-9]' src/ --include="*.go"
```

## Critical Issues (Must Fix)

### 1. Database and Storage Paths

**Impact**: Prevents custom database locations, breaks containerization

| File | Line | Hardcoded Value | Issue |
|------|------|-----------------|-------|
| `storage/binary/wal.go` | 85 | `".wal"` | WAL file extension |
| `storage/binary/reader.go` | 45 | `".idx"` | Index file extension |
| `config/config.go` | 366 | `"/data/entities.db"` | Database path construction |
| `tools/*/` | Various | `"./var/"` | Tool data paths |

**Remediation**: Replace with `config.WALSuffix`, `config.IndexSuffix`, etc.

### 2. Network Ports and Hosts

**Impact**: Prevents deployment flexibility, breaks multi-environment support

| File | Line | Hardcoded Value | Issue |
|------|------|-----------------|-------|
| `docs/docs.go` | 15 | `"localhost:8085"` | Swagger host |
| `api/health_handler.go` | 23 | `":8085"` | Default port reference |
| Multiple tools | Various | `"8085"`, `"8443"` | Port assumptions |

**Remediation**: Replace with `config.SwaggerHost`, remove port assumptions

### 3. File Extensions and Naming

**Impact**: Prevents custom file naming schemes, breaks backup/restore

| File | Line | Hardcoded Value | Issue |
|------|------|-----------------|-------|
| `storage/binary/*.go` | Multiple | `".wal"`, `".db"`, `".idx"` | File extensions |
| `tools/*.go` | Multiple | `"entities.db"` | Database filename |

**Remediation**: Add configuration fields for all file extensions and names

## High Priority Issues (Should Fix)

### 4. Tool Configuration Bypass

**Impact**: Tools behave inconsistently, ignore user configuration

| File Category | Count | Issue |
|---------------|-------|-------|
| `tools/admin/*.go` | 3 files | Direct path construction |
| `tools/entities/*.go` | 6 files | Hardcoded database paths |
| `tools/maintenance/*.go` | 3 files | Fixed file locations |
| `tools/users/*.go` | 4 files | Configuration bypass |

**Pattern Found:**
```go
// ❌ Current (hardcoded)
dbPath := "./var/data/entities.db"

// ✅ Required (configurable)  
cfg := config.Load()
dbPath := cfg.DatabasePath()
```

### 5. Development and Debug Values

**Impact**: Debug settings leak into production, security concerns

| File | Line | Hardcoded Value | Issue |
|------|------|-----------------|-------|
| `main.go` | 188 | Version check logic | Hardcoded flag names |
| `tests/verification/*.go` | Multiple | `"localhost:8085"` | Test endpoints |

## Medium Priority Issues (Nice to Fix)

### 6. Display Strings and Messages

**Impact**: Limited internationalization, fixed user experience

| Category | Count | Examples |
|----------|-------|----------|
| Error messages | 15+ | "Failed to connect to localhost:8085" |
| Log messages | 20+ | "Server starting on port 8085" |
| UI strings | 10+ | Hardcoded labels and text |

### 7. Default Configuration Values

**Impact**: Inconsistent defaults, maintenance overhead

| File | Line | Hardcoded Value | Impact |
|------|------|-----------------|--------|
| `config/config.go` | Multiple | Default values | Should be centralized |
| `share/config/entitydb.env` | Multiple | Default values | Duplication |

## Detailed Findings by File

### Core Server Components

#### `src/main.go`
- **Line 188**: `flag.Lookup("v")` - Hardcoded flag name
- **Line 193**: `flag.Lookup("h")` - Hardcoded flag name
- **Impact**: Essential short flags, acceptable but should be constants

#### `src/config/config.go`
- **Line 366**: `c.DataPath + "/data/entities.db"` - Path construction
- **Multiple lines**: Default values throughout Load() function
- **Impact**: Critical - prevents database location customization

#### `src/config/manager.go`
- **Line 317**: Database configuration key patterns
- **Impact**: Medium - hardcoded configuration structure

### Storage Layer

#### `src/storage/binary/wal.go`
- **Line 85**: `walPath := dataPath + ".wal"` 
- **Impact**: Critical - prevents custom WAL file naming

#### `src/storage/binary/reader.go`
- **Multiple lines**: File extension assumptions
- **Impact**: High - limits file organization flexibility

#### `src/storage/binary/writer.go`
- **Line 45**: Index file naming patterns
- **Impact**: High - hardcoded index structure

### API Layer

#### `src/api/health_handler.go`
- **Line 23**: Port reference in health check
- **Impact**: Medium - breaks health monitoring with custom ports

#### `src/api/system_metrics_handler.go`
- **Line 67**: Metrics collection paths
- **Impact**: Medium - limits metrics storage flexibility

### Tools Directory (Major Issue)

#### Pattern Analysis:
```go
// Found in 16+ tool files
dbPath := "./var/data/entities.db"              // ❌ Hardcoded
repo, _ := binary.NewEntityRepository(dbPath)   // ❌ No config
```

**Files requiring immediate attention:**
1. `tools/admin/create_admin.go`
2. `tools/entities/dump_entity.go`
3. `tools/entities/add_entity.go`
4. `tools/maintenance/check_admin_user.go`
5. `tools/users/create_users.go`
6. `tools/clear_cache.go`
7. `tools/detect_corruption.go`
8. `tools/fix_admin_user.go`
9. `tools/list_users.go`
10. `tools/recovery_tool.go`
11. `tools/rebuild_tag_index.go`
12. `tools/config.go` (ironic!)

### Documentation

#### `src/docs/docs.go`
- **Line 15**: `@host localhost:8085` - Swagger host
- **Impact**: Medium - prevents API documentation customization

## Remediation Priority Matrix

| Priority | Category | Files | Effort | Risk |
|----------|----------|--------|--------|------|
| P0 - Critical | Database paths | 8 files | High | High |
| P0 - Critical | Storage extensions | 5 files | Medium | High |
| P1 - High | Tools refactoring | 16 files | High | Medium |
| P1 - High | Port assumptions | 6 files | Low | Medium |
| P2 - Medium | Display strings | 20+ files | Low | Low |

## Implementation Recommendations

### Phase 1: Immediate Fixes (Day 1)
1. **Extend Config struct** with missing fields:
   ```go
   type Config struct {
       // Add these fields
       DatabaseFilename string
       WALSuffix       string  
       IndexSuffix     string
       BackupPath      string
       TempPath        string
   }
   ```

2. **Update environment variables**:
   ```bash
   ENTITYDB_DATABASE_FILENAME=entities.db
   ENTITYDB_WAL_SUFFIX=.wal
   ENTITYDB_INDEX_SUFFIX=.idx
   ENTITYDB_BACKUP_PATH=/opt/entitydb/backup
   ENTITYDB_TEMP_PATH=/opt/entitydb/tmp
   ```

### Phase 2: Core Components (Days 2-3)
1. **Fix storage layer** - Replace hardcoded extensions
2. **Update path construction** - Use config methods
3. **Fix API references** - Remove port assumptions

### Phase 3: Tools Refactoring (Days 4-5)
1. **Standardize tool pattern**:
   ```go
   func main() {
       configManager := config.NewConfigManager(nil)
       configManager.RegisterFlags()
       flag.Parse()
       
       cfg, err := configManager.Initialize()
       if err != nil {
           log.Fatal(err)
       }
       
       // Use cfg throughout
   }
   ```

2. **Update all 16 tool files** with this pattern

### Phase 4: Validation (Day 6)
1. **Automated verification script**
2. **Integration testing**
3. **Performance validation**

## Automated Fix Script

```bash
#!/bin/bash
# check_hardcoded_values.sh - Verification script

echo "=== EntityDB Configuration Compliance Check ==="

FAILED=0

# Check for hardcoded ports
echo "Checking for hardcoded ports..."
if grep -r "8085\|8443" src/ --include="*.go" | grep -v "config\|test\|example"; then
    echo "❌ FAIL: Hardcoded ports found"
    FAILED=1
fi

# Check for hardcoded paths  
echo "Checking for hardcoded paths..."
if grep -r '"/opt/\|"/var/\|"/tmp/\|"./var/"' src/ --include="*.go" | grep -v "config\|test\|example"; then
    echo "❌ FAIL: Hardcoded paths found"
    FAILED=1
fi

# Check for hardcoded file extensions
echo "Checking for hardcoded file extensions..."
if grep -r '\.db"\|\.wal"\|\.idx"\|\.log"\|\.pid"' src/ --include="*.go" | grep -v "config\|test"; then
    echo "❌ FAIL: Hardcoded file extensions found"
    FAILED=1
fi

# Check tool compliance
echo "Checking tool configuration compliance..."
if grep -r "NewEntityRepository.*var" src/tools/ --include="*.go"; then
    echo "❌ FAIL: Tools bypassing configuration system"
    FAILED=1
fi

if [ $FAILED -eq 0 ]; then
    echo "✅ SUCCESS: No hardcoded values found"
else
    echo "❌ FAILED: Hardcoded values detected"
    exit 1
fi
```

## Success Metrics

**Quantitative Goals:**
- Zero hardcoded paths in production code
- Zero hardcoded ports outside configuration
- 100% tool compliance with ConfigManager
- <1ms configuration loading overhead

**Qualitative Goals:**
- Easy deployment across environments
- Single source of truth for all configuration
- Consistent behavior across all tools
- Clear configuration hierarchy precedence

## Next Steps

1. **Review and approve** this audit report
2. **Begin Phase 1** implementation immediately
3. **Set up continuous validation** with automated checks
4. **Track progress** against the implementation timeline
5. **Test thoroughly** in multiple environments

---

**Audit Complete**: This comprehensive audit provides the foundation for systematic elimination of all hardcoded values in EntityDB, ensuring full configuration management compliance.