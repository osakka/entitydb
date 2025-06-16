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

## Creating New ADRs

1. Use the next sequential number
2. Follow the naming convention: `XXX-kebab-case-title.md`
3. Copy the template from `template.md`
4. Update the index above when adding new ADRs

## References

- [Architecture Decision Records](https://adr.github.io/)
- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)