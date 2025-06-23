# Numbered Architecture Documentation Migration Map

This document tracks where each numbered architecture document was migrated to in the new documentation structure.

## ✅ High-Value Documents Migrated to Active Documentation

| Original File | New Location | Purpose |
|--------------|--------------|---------|
| `001-temporal-tag-storage.md` | `/docs/reference/technical-specs/temporal-storage-specification.md` | Temporal implementation spec |
| `002-binary-storage-format.md` | `/docs/reference/technical-specs/binary-format-specification.md` | Binary format specification |
| `003-unified-sharded-indexing.md` | `/docs/developer-guide/implementation/sharded-indexing-implementation.md` | Indexing implementation |
| `004-tag-based-rbac.md` | `/docs/developer-guide/implementation/rbac-implementation-guide.md` | RBAC implementation |
| `006-credential-storage-in-entities.md` | `/docs/developer-guide/implementation/authentication-implementation.md` | Auth implementation |
| `010-temporal-functionality-completion.md` | `/docs/developer-guide/implementation/temporal-implementation-guide.md` | Temporal implementation |
| `011-production-battle-testing.md` | `/docs/developer-guide/testing/production-battle-testing-guide.md` | Testing methodology |
| `026-unified-file-format-architecture.md` | `/docs/reference/technical-specs/unified-file-format-specification.md` | EUFF specification |
| `memory-optimization-architecture.md` | `/docs/reference/technical-specs/memory-optimization-architecture.md` | Memory architecture |

## 📚 Documents Kept in Archive for Historical Reference

### Evolution & History Documents
These documents provide valuable historical context but are not active technical specifications:

- `005-application-agnostic-design.md` - Platform evolution history
- `007-memory-mapped-file-access.md` - Implementation details superseded by current design
- `008-three-tier-configuration.md` - Configuration evolution (current in ADR-030)
- `012-binary-repository-unification.md` - Historical unification process
- `013-pure-tag-based-session-management.md` - Session evolution history
- `014-single-source-of-truth-enforcement.md` - Architectural principle evolution
- `015-wal-management-and-checkpointing.md` - WAL evolution (current in ADR-028)
- `016-error-recovery-and-resilience.md` - Recovery system evolution
- `017-automatic-index-corruption-recovery.md` - Corruption recovery history
- `018-self-cleaning-temporal-retention.md` - Retention system evolution
- `019-index-rebuild-loop-fix.md` - Specific bug fix documentation
- `020-comprehensive-architectural-timeline.md` - Superseded by ADR-000
- `021-critical-corruption-prevention-fix.md` - Specific fix documentation
- `022-dynamic-request-throttling.md` - Throttling implementation history
- `023-index-entry-race-condition-elimination.md` - Race condition fix
- `024-incremental-update-architecture.md` - Update system evolution
- `025-aggregation-timing-bootstrap-fix.md` - Specific timing fix
- `027-complete-database-file-unification.md` - Unification process details
- `028-logging-standards-compliance.md` - Logging evolution (current in docs)
- `029-documentation-excellence-achievement.md` - Documentation evolution
- `030-storage-efficiency-validation.md` - Storage optimization history
- `031-architectural-decision-documentation-excellence.md` - ADR evolution
- `032-migration-from-sqlite-to-binary-format.md` - Historical migration
- `033-evolution-from-specialized-to-unified-entity-api.md` - API evolution
- `034-security-architecture-evolution.md` - Security system history
- `035-consolidated-architectural-timeline.md` - Timeline (superseded by ADR-000)

### Overview Documents (Remain in Archive)
These are descriptive documents rather than technical specifications:

- `01-system-overview.md` - High-level system description
- `02-temporal-architecture.md` - Temporal system overview
- `03-rbac-architecture.md` - RBAC system overview
- `04-entity-model.md` - Entity model description
- `05-authentication-architecture.md` - Auth system overview
- `06-rbac-tag-format.md` - Tag format description
- `07-tag-system.md` - Tag system overview
- `08-dataset-architecture.md` - Dataset system overview
- `09-metrics-architecture.md` - Metrics system overview
- `10-tag-validation.md` - Validation rules

## 🎯 Migration Rationale

**Promoted to Active Documentation**: 
- Technical specifications needed for implementation
- Testing methodologies actively used
- Implementation guides for contributors

**Kept in Archive**:
- Historical evolution documents
- Superseded implementations
- High-level overviews (replaced by current docs)
- Specific bug fix documentation

## 📋 Access Patterns

### For Current Technical Specs
→ `/docs/reference/technical-specs/`

### For Implementation Details  
→ `/docs/developer-guide/implementation/`

### For Testing Guides
→ `/docs/developer-guide/testing/`

### For Historical Context
→ `/docs/archive/numbered-architecture/`

---

**Migration Date**: 2025-06-23  
**Purpose**: Track document reorganization for reference  
**Note**: All content preserved, just better organized