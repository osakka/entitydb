# EntityDB Technical Specifications

This directory contains detailed technical specifications for EntityDB's core systems and architectures.

## 📋 Core Technical Specifications

### Storage Layer
- **[Binary Format Specification](./binary-format-specification.md)** - EntityDB Binary Format (EBF) detailed specification
- **[Unified File Format Specification](./unified-file-format-specification.md)** - EntityDB Unified File Format (EUFF) technical details
- **[Temporal Storage Specification](./temporal-storage-specification.md)** - Nanosecond-precision temporal tag storage implementation

### Performance & Optimization
- **[Memory Optimization Architecture](./memory-optimization-architecture.md)** - Comprehensive memory management system details
- **[Sharded Indexing Implementation](../../../developer-guide/implementation/sharded-indexing-implementation.md)** - 256-shard concurrent indexing system

## 🔍 Key Technical Details

### Binary Format (EBF)
- **Magic Number**: `0x45424600` ("EBF\0")
- **Entry Structure**: Header (24 bytes) + EntityID (64 bytes) + Tags + Content
- **Compression**: GZIP for content > 1KB
- **Checksums**: SHA256 for data integrity

### Unified File Format (EUFF)
- **Magic Number**: `0x45555446` ("EUFF") 
- **Format Version**: 2
- **Sections**: WAL, Data, Tag Dictionary, Entity Index
- **Single File**: Eliminates separate .db, .wal, .idx files

### Temporal Storage
- **Format**: `TIMESTAMP|tag_value`
- **Precision**: Nanosecond (int64)
- **Indexing**: B-tree timeline with skip-lists
- **Query Support**: as-of, history, diff, changes

### Memory Architecture
- **Entity Cache**: LRU with 10,000 entity limit
- **String Interning**: 100MB limit with eviction
- **Memory Guardian**: 80% threshold protection
- **Buffer Pools**: Size-based pools (small, medium, large)

## 📚 Related Documentation

### Implementation Guides
- [Developer Implementation Guide](../../developer-guide/implementation/)
- [Testing Guide](../../developer-guide/testing/)
- [Performance Tuning](../../admin-guide/07-monitoring-guide.md)

### Architecture Decisions
- [ADR-022: Database File Unification](../../architecture/adr/ADR-022-database-file-unification.md)
- [ADR-028: WAL Corruption Prevention](../../architecture/adr/ADR-028-wal-corruption-prevention.md)
- [ADR-029: Memory Optimization](../../architecture/adr/ADR-029-intelligent-recovery-system.md)

## 🔧 Using These Specifications

These technical specifications are intended for:

1. **Developers** implementing new features or debugging issues
2. **Operations Teams** understanding system internals for troubleshooting
3. **Contributors** needing detailed technical understanding
4. **Integrators** building tools that work with EntityDB files

### Version Compatibility

| Specification | Introduced | Current Version | Breaking Changes |
|--------------|------------|-----------------|------------------|
| EBF | v2.13.0 | v3 | v2→v3 added compression |
| EUFF | v2.32.6 | v2 | None since introduction |
| Temporal Tags | v2.14.0 | v1 | None |
| Memory Arch | v2.34.0 | v1 | None |

---

**Last Updated**: 2025-06-23  
**Maintainers**: EntityDB Core Team