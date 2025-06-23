# ADR-003: Unified Sharded Indexing Architecture

## Status
Accepted (2025-06-16)

## Context
EntityDB had two parallel indexing systems causing inconsistencies and maintenance overhead:

1. **Legacy tagIndex**: Map-based indexing with conditional logic
2. **ShardedTagIndex**: 256-shard concurrent indexing system

The dual system caused:
- Code complexity with `useShardedIndex` conditional logic
- Potential data inconsistencies between index systems
- Authentication failures due to index mismatches
- Maintenance burden of two implementations

### Requirements
- Single source of truth for tag indexing
- High concurrent performance with minimal lock contention
- Consistent authentication and session management
- Clean codebase without conditional indexing logic

### Constraints
- Must maintain backward compatibility during transition
- Cannot break existing functionality
- Performance must equal or exceed current system
- Authentication system must remain stable

## Decision
We decided to **eliminate the legacy tagIndex system entirely** and standardize on **ShardedTagIndex** as the single indexing implementation.

### Implementation Details
- **256 Shards**: Optimal balance of concurrency and memory overhead
- **Hash Distribution**: Even distribution of tags across shards using tag hash
- **Concurrent Access**: Per-shard locking for minimal contention
- **Unified Interface**: Single `TagIndex` interface backed by sharded implementation
- **Authentication Compatibility**: All session and user lookups use sharded index

### Removal Strategy
1. **Code Elimination**: Removed all `useShardedIndex` conditional logic
2. **Interface Cleanup**: Simplified repository constructors
3. **Test Updates**: Updated tests to use sharded index only
4. **Authentication Fix**: Fixed session validation with sharded lookups
5. **Environment Variables**: Removed `ENTITYDB_USE_SHARDED_INDEX` configuration

## Consequences

### Positive
- **Code Simplification**: Eliminated ~30 conditional code blocks
- **Performance Consistency**: Uniform high-performance indexing
- **Authentication Stability**: Resolved session lookup inconsistencies
- **Maintenance Reduction**: Single indexing implementation to maintain
- **Reduced Lock Contention**: 256-way concurrent access
- **Clean Architecture**: No more dual-path complexity

### Negative
- **Migration Effort**: Required updating all index-related code
- **Testing Overhead**: Comprehensive testing of authentication flows
- **Deployment Risk**: Single-shot migration with no fallback

### Performance Impact
- **Concurrent Access**: 256x potential parallelism vs single-threaded legacy
- **Memory Overhead**: Slight increase due to shard management structures
- **Query Performance**: Maintained or improved due to reduced contention
- **Authentication Speed**: Resolved timeout issues with consistent indexing

## Implementation History
- v2.27.0: Initial sharded indexing implementation with dual support
- v2.30.0: Authentication issues traced to index inconsistencies
- v2.32.0: Complete legacy elimination and unified architecture

## Verification
- **Clean Build**: Zero compilation warnings
- **Authentication Tests**: All login and session flows verified
- **Performance Tests**: No regression in query performance
- **Concurrent Tests**: Improved performance under load

## Related Decisions
- [ADR-001: Temporal Tag Storage](./001-temporal-tag-storage.md) - Tag format indexed
- [ADR-002: Binary Storage Format](./002-binary-storage-format.md) - Storage layer integration