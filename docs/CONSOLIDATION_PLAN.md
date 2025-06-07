# Documentation Consolidation Plan

## Duplicate Content Analysis

### 1. Architecture Documentation (3 locations)
- `docs/architecture/` (10 files)
- `docs/new-structure/architecture/` (2 files) 
- `docs/archive/` (multiple architecture files)

**Action**: Consolidate to `docs/new-structure/architecture/` with comprehensive coverage

### 2. API Documentation (3 locations)  
- `docs/api/` (6 files)
- `docs/new-structure/api/` (2 files)
- `docs/archive/` (multiple API files)

**Action**: Consolidate to `docs/new-structure/api/` with complete reference

### 3. Implementation Guides (2 locations)
- `docs/implementation/` (42 files)
- `docs/new-structure/implementation/` (6 files)

**Action**: Key files moved to new structure, obsolete files to archive

### 4. Performance Documentation (2 locations)
- `docs/performance/` (11 files) 
- `docs/new-structure/performance/` (4 files)

**Action**: Consolidate to `docs/new-structure/performance/`

### 5. Development Guides (2 locations)
- `docs/development/` (7 files)
- `docs/new-structure/development/` (4 files) 

**Action**: Merge to `docs/new-structure/development/` with complete coverage

### 6. Logging Documentation (4 versions)
- `docs/development/logging-standards.md`
- `docs/development/logging-standards-v3.md` 
- `docs/implementation/LOGGING_STANDARDS_V2.md`
- `docs/new-structure/development/dev-logging.md`

**Action**: Keep only the latest in new structure

## Files to Move to Archive

### Obsolete Implementation Files
- All temporary fix summaries (auth-hang-fix, ssl-implementation-summary, etc.)
- Old metric fixes (metrics-system-fix-complete.md)
- Temporary workarounds (te-header-hang-fix.md)

### Superseded Architecture Files  
- Old architecture proposals
- Entity migration guides (pre-v2.0)
- Legacy tag system guides

### Duplicate API Files
- Old API references
- Incomplete documentation drafts
- Version-specific fixes

## Consolidation Actions

### Phase 1: Core Documentation ✅
- [x] Create new structure taxonomy
- [x] Move critical architecture docs
- [x] Create comprehensive API reference
- [x] Update root README.md

### Phase 2: Implementation Consolidation (Current)
- [x] Move key implementation guides
- [ ] Archive obsolete implementation files
- [ ] Consolidate performance documentation
- [ ] Clean up duplicate logging docs

### Phase 3: Final Cleanup
- [ ] Archive old directory structures
- [ ] Update all cross-references
- [ ] Create maintenance guidelines

## Priority Files to Keep

### Architecture (Final Location: docs/new-structure/architecture/)
- arch-overview.md ✅
- arch-temporal.md (from temporal_architecture.md)
- arch-rbac.md (from tag_based_rbac.md)
- arch-dataspace.md ✅

### API (Final Location: docs/new-structure/api/)
- api-reference.md ✅ (comprehensive)
- api-authentication.md (from auth.md)
- api-entities.md (from entities.md)
- api-examples.md (from examples.md)

### Implementation (Final Location: docs/new-structure/implementation/)
- impl-autochunking.md ✅
- impl-temporal.md ✅
- impl-dataspace.md ✅
- impl-performance.md ✅
- impl-logging.md (moved to development/)

## Files to Archive

### Implementation Directory (42 files → Archive most)
Keep only:
- Core implementation guides (moved to new structure)
- Current architecture plans

Archive:
- All "*-fix-*" files (temporary fixes)  
- All "*-summary" files (outdated summaries)
- All debug and troubleshooting files
- Metrics UI fixes and temporary patches

### Architecture Directory (10 files → Keep 2-3)
Keep:
- Current architecture overview
- Temporal architecture (migrate to new structure)

Archive:  
- Old proposals and visions
- Superseded RBAC implementations
- Legacy tag system documentation

### Development Directory (7 files → Consolidate to 4)
Keep:
- Core contributing guidelines
- Git workflow
- Current configuration management

Archive:
- Old logging standards (3 versions)
- Production notes (integrate into ops guides)

## File Movement Summary

### To New Structure
- 15 core architecture files → 4 consolidated arch files
- 25 implementation files → 6 key implementation guides  
- 12 API files → 4 comprehensive API docs
- 11 performance files → 3 consolidated performance docs

### To Archive
- ~80 obsolete/duplicate files
- All temporary fix documentation
- Superseded architecture proposals
- Old API documentation versions

This consolidation reduces documentation from ~150 scattered files to ~25 authoritative documents in the new structure.