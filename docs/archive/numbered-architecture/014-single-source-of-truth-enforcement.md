# ADR-014: Single Source of Truth Enforcement

## Status
✅ **ACCEPTED** - 2025-06-16

## Context
EntityDB codebase had accumulated multiple implementations, duplicate functionality, and parallel code paths that violated the fundamental software engineering principle of single source of truth. This created maintenance overhead, potential inconsistencies, and made the system harder to reason about.

## Problem
- Multiple implementations of similar functionality across the codebase
- Duplicate handlers, repositories, and utility functions
- Parallel code paths that could diverge in behavior
- Maintenance burden of keeping multiple implementations synchronized
- Risk of bugs being fixed in one implementation but not others
- Architectural complexity from redundant components

## Decision
Enforce strict single source of truth principle across all EntityDB components:

### Systematic Elimination Strategy
1. **Identify Duplicates**: Comprehensive audit to find all duplicate implementations
2. **Choose Canonical Version**: Select the most complete, tested, and maintainable implementation
3. **Surgical Removal**: Remove redundant implementations with precision
4. **Consolidate Callers**: Update all usage to reference single implementation
5. **Validate Behavior**: Ensure no functionality is lost during consolidation

### Key Areas of Consolidation

#### Handler Unification
```go
// BEFORE: Multiple entity handlers
- entity_handler.go
- entity_handler_v2.go  
- entity_handler_legacy.go

// AFTER: Single authoritative handler
- entity_handler.go (consolidated best features)
```

#### Repository Consolidation
```go
// BEFORE: Parallel repository implementations
- EntityRepository (interface)
- CachedEntityRepository (wrapper)
- HighPerformanceEntityRepository (alternate)
- LegacyEntityRepository (deprecated)

// AFTER: Single repository with optional features
- EntityRepository (unified interface)
- CachedRepository (optional wrapper)
```

#### RBAC Implementation Unification
- Eliminated parallel RBAC handlers (`relationship_handler_rbac.go`)
- Consolidated permission checking into single security module
- Removed duplicate session management implementations
- Unified authentication flow with single code path

## Implementation Details

### Surgical Cleanup Process
1. **Code Archaeology**: Document purpose and usage of each duplicate
2. **Feature Matrix**: Compare functionality across implementations
3. **Dependency Analysis**: Identify all callers and dependencies
4. **Migration Plan**: Define steps to move to canonical implementation
5. **Validation**: Test that consolidated version preserves all required behavior

### Files Eliminated
- `src/api/entity_handler_v2.go` → consolidated into `entity_handler.go`
- `trash/relationship_system/` → duplicate RBAC implementations
- Multiple tool duplicates → single authoritative tools
- Legacy backup implementations → moved to trash

### Preserved in Trash
- All removed code preserved in `/trash` directory for archaeology
- Complete git history maintained for reference
- Documentation explaining why each piece was removed

## Consequences

### Positive
- ✅ **Reduced Complexity**: Fewer code paths to understand and maintain
- ✅ **Consistent Behavior**: Single implementation ensures uniform behavior
- ✅ **Easier Maintenance**: Bug fixes and features apply universally
- ✅ **Faster Development**: No need to update multiple implementations
- ✅ **Clearer Architecture**: Obvious entry points and data flows
- ✅ **Reduced Testing Overhead**: Single implementation to test thoroughly
- ✅ **Better Documentation**: Clear canonical reference for each function

### Negative
- ⚠️ **Migration Effort**: Required updating callers to use canonical implementation
- ⚠️ **Feature Loss Risk**: Potential for losing edge case functionality during consolidation
- ⚠️ **Backup Loss**: Removed fallback implementations that might have been useful

### Risks Mitigated
- 🔒 **Inconsistent Behavior**: Eliminated by having single implementation
- 🔒 **Partial Bug Fixes**: All fixes now apply to single authoritative version
- 🔒 **Configuration Confusion**: Clear single configuration point
- 🔒 **Testing Gaps**: Comprehensive testing of single implementation vs partial testing of many

## Validation Process
1. **Comprehensive Testing**: All functionality tested against consolidated implementation
2. **Performance Validation**: Ensured no performance regression
3. **Feature Completeness**: Verified all required features preserved
4. **Integration Testing**: End-to-end testing across all major workflows
5. **Documentation Update**: All references updated to canonical implementation

## Alternatives Considered
1. **Gradual Consolidation**: Risk of extended period with duplicates
2. **Feature Flags**: Added complexity without addressing root issue
3. **Interface Abstraction**: Would hide but not eliminate duplication

## References
- Implementation: Surgical cleanup across all `src/` directories
- Preserved Code: `/trash` directory with complete removal history
- Git Commits: 
  - `fc2361a` - "refactor: surgical cleanup for crystal clear single source of truth"
  - `70a5b86` - "feat: eliminate parallel RBAC implementations"
- Related: ADR-012 (Binary Repository Unification)

## Enforcement Guidelines
1. **New Code**: Must not duplicate existing functionality
2. **Code Reviews**: Check for single source of truth violations
3. **Refactoring**: Prefer consolidation over creating new implementations
4. **Tool Development**: Use existing utilities rather than creating new ones
5. **Documentation**: Maintain clear pointers to canonical implementations

## Timeline
- **2025-06-15**: Initial duplicate identification and removal
- **2025-06-16**: Comprehensive surgical cleanup implemented
- **2025-06-16**: All callers updated to use canonical implementations
- **2025-06-17**: Final validation and testing completed

---
*This ADR documents the critical architectural decision to enforce single source of truth across all EntityDB components, eliminating duplicate implementations and ensuring consistent, maintainable code.*