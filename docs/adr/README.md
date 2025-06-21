# Architecture Decision Records (ADR)

This directory contains records of architectural decisions made for the EntityDB project. An ADR is a document that captures an important architectural decision along with its context and consequences.

## Format

Each ADR follows the template:
- **Title**: Short noun phrase
- **Status**: Proposed, Accepted, Deprecated, Superseded
- **Context**: Forces at play, constraints, requirements
- **Decision**: What we decided to do
- **Consequences**: Positive and negative outcomes

## ADR Index

### Core Architecture Decisions

| ADR | Title | Status | Date | Git Commits |
|-----|-------|--------|------|-------------|
| [001](./001-temporal-tag-storage.md) | Temporal Tag Storage with Nanosecond Precision | Accepted | 2025-05-08 | `08c1ce08`, `975a561a` |
| [002](./002-binary-storage-format.md) | Custom Binary Format (EBF) over SQLite | Accepted | 2025-05-15 | `709f865c`, `87a08fa4` |
| [003](./003-unified-sharded-indexing.md) | Unified Sharded Indexing Architecture | Accepted | 2025-06-16 | `6d76c26d`, `56f393e0` |
| [004](./004-tag-based-rbac.md) | Tag-Based RBAC System | Accepted | 2025-04-15 | `a22193d7`, `70a5b86f` |
| [005](./005-application-agnostic-design.md) | Application-Agnostic Platform Design | Accepted | 2025-06-02 | `30ca7981`, `224eac3e` |

### Implementation Decisions

| ADR | Title | Status | Date | Git Commits |
|-----|-------|--------|------|-------------|
| [006](./006-credential-storage-in-entities.md) | User Credentials in Entity Content | Accepted | 2025-06-08 | `e3b50904`, `7fed6868` |
| [007](./007-memory-mapped-file-access.md) | Memory-Mapped File Access Pattern | Accepted | 2025-05-15 | `87a08fa4`, `0ed28c89` |
| [008](./008-three-tier-configuration.md) | Three-Tier Configuration Hierarchy | Accepted | 2025-06-07 | `bf001189`, `041cb238` |

### Performance & Optimization Decisions

| ADR | Title | Status | Date | Git Commits |
|-----|-------|--------|------|-------------|
| [009](./009-memory-optimization-suite.md) | Comprehensive Memory Optimization Suite | Accepted | 2025-06-13 | `87a08fa4`, `0ed28c89` |
| [010](./010-temporal-functionality-completion.md) | Complete Temporal Database Implementation | Accepted | 2025-06-16 | `cf6ce80e`, `456fee63` |
| [011](./011-production-battle-testing.md) | Production Battle Testing and Multi-Tag Performance Optimization | Accepted | 2025-06-17 | `d57168c`, `6ef5003` |

### Architecture & Engineering Decisions

| ADR | Title | Status | Date | Git Commits |
|-----|-------|--------|------|-------------|
| [012](./012-binary-repository-unification.md) | Binary Repository Unification and Single Source of Truth | Accepted | 2025-06-15 | `a22193d`, `2baa028` |
| [013](./013-pure-tag-based-session-management.md) | Pure Tag-Based Session Management | Accepted | 2025-06-15 | `b91d85a`, `a99cf6c` |
| [014](./014-single-source-of-truth-enforcement.md) | Single Source of Truth Enforcement | Accepted | 2025-06-16 | `fc2361a`, `70a5b86` |

### Reliability & Operations Decisions

| ADR | Title | Status | Date | Git Commits |
|-----|-------|--------|------|-------------|
| [015](./015-wal-management-and-checkpointing.md) | WAL Management and Automatic Checkpointing | Accepted | 2025-06-16 | `c10f023`, `wal commits` |
| [016](./016-error-recovery-and-resilience.md) | Error Recovery and Resilience Architecture | Accepted | 2025-06-17 | `de9cd28`, `975afa5` |
| [017](./017-automatic-index-corruption-recovery.md) | Automatic Index Corruption Recovery | Accepted | 2025-06-17 | `cef9101`, `ec84efe` |
| [018](./018-self-cleaning-temporal-retention.md) | Self-Cleaning Temporal Retention Architecture | Accepted | 2025-06-18 | `e03ae65`, `7464c52` |
| [019](./019-index-rebuild-loop-fix.md) | Index Rebuild Loop Critical Fix | Accepted | 2025-06-18 | `d7111b3`, ADR creation |

### System Stability & Performance (June 2025)

| ADR | Title | Status | Date | Git Commits |
|-----|-------|--------|------|-------------|
| [021](./021-critical-corruption-prevention-fix.md) | Critical Corruption Prevention Fix | Accepted | 2025-06-19 | `4c0bb51`, `17fba0a` |
| [022](./022-dynamic-request-throttling.md) | Dynamic Request Throttling for UI Abuse Protection | Accepted | 2025-06-19 | `e3af73d` |
| [023](./023-index-entry-race-condition-elimination.md) | Index Entry Race Condition Elimination | Accepted | 2025-06-19 | `5a9fa9d` |
| [024](./024-incremental-update-architecture.md) | Incremental Update Architecture | Accepted | 2025-06-19 | `02c251a` |
| [025](./025-aggregation-timing-bootstrap-fix.md) | Aggregation Timing Bootstrap Fix | Accepted | 2025-06-19 | `0689115` |

### Database File Unification (June 2025)

| ADR | Title | Status | Date | Git Commits |
|-----|-------|--------|------|-------------|
| [026](./026-unified-file-format-architecture.md) | Unified File Format Architecture | Accepted | 2025-06-20 | `ebd945b`, `3157f1b` |
| [027](./027-complete-database-file-unification.md) | Complete Database File Unification | Accepted | 2025-06-20 | `81cf44a`, `3157f1b`, `ebd945b` |

### Observability & Standards (June 2025)

| ADR | Title | Status | Date | Git Commits |
|-----|-------|--------|------|-------------|
| [028](./028-logging-standards-compliance.md) | Logging Standards Compliance and Audience Optimization | Accepted | 2025-06-20 | `pending commit` |

### Historical Context & Evolution

| ADR | Title | Status | Date | Git Commits |
|-----|-------|--------|------|-------------|
| [032](./032-migration-from-sqlite-to-binary-format.md) | Migration from SQLite to Custom Binary Format | Accepted | 2025-05-15 | Based on old repository analysis |
| [033](./033-evolution-from-specialized-to-unified-entity-api.md) | Evolution from Specialized APIs to Unified Entity Architecture | Accepted | 2025-05-11 | Based on old repository analysis |
| [034](./034-security-architecture-evolution.md) | Security Architecture Evolution from Component-Based to Unified Model | Accepted | 2025-06-08 | Based on old repository analysis |

### Comprehensive Documentation

| ADR | Title | Status | Date | Git Commits |
|-----|-------|--------|------|-------------|
| [020](./020-comprehensive-architectural-timeline.md) | Comprehensive Architectural Decision Timeline | Accepted | 2025-06-18 | `verification commits` |

## Decision Verification Status

All ADRs have been verified against the actual v2.32.5 codebase implementation. The comprehensive timeline in ADR-020 provides complete traceability between decisions, git commits, and actual code changes.

## Architecture Evolution Summary

### Phase 1: Foundation (May 2025)
- **ADR-001**: Temporal tag storage foundation
- **ADR-002**: Custom binary format (EBF)
- **ADR-004**: Tag-based RBAC system

### Phase 2: Performance & Optimization (June 2025)
- **ADR-007**: Memory-mapped file access
- **ADR-008**: Three-tier configuration
- **ADR-009**: Memory optimization suite

### Phase 3: Production Readiness (June 2025)
- **ADR-010**: Complete temporal functionality
- **ADR-011**: Production battle testing
- **ADR-012**: Binary repository unification

### Phase 4: Reliability & Operations (June 2025)
- **ADR-015**: WAL management
- **ADR-016**: Error recovery
- **ADR-017**: Index corruption recovery
- **ADR-018**: Self-cleaning retention
- **ADR-019**: Index rebuild loop fix

### Phase 5: System Stability & Metrics (June 2025)
- **ADR-021**: Critical corruption prevention fix
- **ADR-022**: Dynamic request throttling
- **ADR-023**: Index entry race condition elimination
- **ADR-024**: Incremental update architecture
- **ADR-025**: Aggregation timing bootstrap fix

### Phase 6: Database File Unification (June 2025)
- **ADR-026**: Unified file format architecture
- **ADR-027**: Complete database file unification

## Architectural Principles

Based on our decision history, EntityDB follows these core principles:

1. **Single Source of Truth**: No duplicate implementations or parallel code paths
2. **Performance First**: All decisions optimized for high-performance temporal operations
3. **Production Focus**: Every decision contributes to production-grade reliability
4. **Unified Entity Model**: Everything is an entity with tags
5. **Zero Technical Debt**: Clean codebase with no TODOs or technical shortcuts

## Decision Governance

- **New ADRs**: Must align with established architectural principles
- **Verification**: All decisions must be verified against actual implementation
- **Traceability**: Complete audit trail from decision to code changes
- **Timeline Maintenance**: ADR-020 provides comprehensive decision timeline

## Creating New ADRs

1. Use the next sequential number
2. Follow the naming convention: `XXX-kebab-case-title.md`
3. Copy the template from `template.md`
4. Update the index above when adding new ADRs

## References

- [Architecture Decision Records](https://adr.github.io/)
- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)