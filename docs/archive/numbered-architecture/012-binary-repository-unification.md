# ADR-012: Binary Repository Unification and Single Source of Truth

## Status
✅ **ACCEPTED** - 2025-06-15

## Context
EntityDB had multiple repository implementations and interfaces that created complexity, potential inconsistencies, and violated the single source of truth principle. The codebase contained parallel implementations for different use cases, making maintenance difficult and introducing potential for bugs.

## Problem
- Multiple repository constructors and interfaces scattered across codebase
- Parallel implementations of similar functionality
- Inconsistent behavior between different repository types
- Maintenance overhead of keeping multiple implementations in sync
- Violation of single source of truth architectural principle

## Decision
Implement unified repository architecture with single source of truth:

### Unified Constructor Pattern
```go
// Single constructor for all repository needs
func NewEntityRepositoryWithConfig(config *RepositoryConfig) (storage.EntityRepository, error)

// Eliminates multiple constructors:
// - NewEntityRepository()
// - NewHighPerformanceRepository()  
// - NewTemporalRepository()
// - NewCachedRepository()
```

### Single Interface Implementation
- Consolidate all repository interfaces into unified `EntityRepository`
- Eliminate parallel implementations of CRUD operations
- Standardize error handling and logging across all operations
- Unified configuration management for all repository features

### Repository Feature Composition
```go
type RepositoryConfig struct {
    HighPerformance bool
    TemporalSupport bool
    CachingEnabled  bool
    BatchingEnabled bool
}
```

## Implementation Details

### Architecture Changes
1. **Eliminated Duplicate Constructors**: Removed 6+ constructor functions
2. **Unified Interface**: Single `EntityRepository` interface for all operations
3. **Composition Over Inheritance**: Features added through configuration flags
4. **Consistent Error Handling**: Standardized error types and messages

### Migration Path
1. Update all tools to use `NewEntityRepositoryWithConfig()`
2. Remove legacy constructor functions
3. Consolidate test suites to use unified interface
4. Update documentation to reflect single constructor pattern

## Consequences

### Positive
- ✅ **Single Source of Truth**: One implementation, one behavior
- ✅ **Reduced Complexity**: Fewer code paths to maintain
- ✅ **Consistent Behavior**: All operations follow same patterns
- ✅ **Easier Testing**: Single interface to test comprehensively
- ✅ **Simplified Maintenance**: Changes apply universally
- ✅ **Clear Configuration**: Explicit feature flags vs implicit behavior

### Negative
- ⚠️ **Migration Effort**: Required updating all existing usage
- ⚠️ **Feature Discovery**: Less obvious which features are available
- ⚠️ **Configuration Complexity**: More parameters to understand

### Risks Mitigated
- 🔒 **Inconsistent Behavior**: Eliminated by single implementation
- 🔒 **Parallel Bug Fixes**: No longer needed with unified codebase
- 🔒 **API Confusion**: Clear single interface for all operations

## Alternatives Considered
1. **Keep Multiple Interfaces**: Risk of continued divergence
2. **Gradual Migration**: Risk of extended maintenance burden
3. **Factory Pattern**: Added complexity without clear benefits

## References
- Implementation: `src/storage/binary/entity_repository.go`
- Tools Migration: `tools/*.go` updated to use unified constructor
- Git Commit: `a22193d` - "refactor: implement unified repository architecture"
- Related: ADR-003 (Unified Sharded Indexing)

## Timeline
- **2025-06-15**: Decision made and implementation started
- **2025-06-15**: All tools migrated to unified constructor
- **2025-06-15**: Legacy constructors removed
- **2025-06-16**: Validation and testing completed

---
*This ADR documents the critical architectural decision to unify all repository implementations under a single source of truth, eliminating complexity and ensuring consistent behavior across the EntityDB platform.*