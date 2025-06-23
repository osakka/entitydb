# Architectural Decision Records (ADR)

This directory contains formal Architectural Decision Records for EntityDB. ADRs document important architectural decisions along with their context and consequences.

## What are ADRs?

Architectural Decision Records (ADRs) are documents that capture important architectural decisions made during the development of EntityDB. Each ADR documents:

- **Context**: The situation that led to the decision
- **Decision**: The architectural decision made
- **Status**: Current status (proposed, accepted, deprecated, superseded)
- **Consequences**: The positive and negative consequences of the decision

## ADR Index

| ADR | Title | Status | Date |
|-----|-------|---------|------|
| [ADR-000](./ADR-000-MASTER-TIMELINE.md) | Master Timeline | Accepted | 2025-06-23 |
| [ADR-022](./ADR-022-database-file-unification.md) | Database File Unification | Accepted | 2025-06-20 |
| [ADR-028](./ADR-028-wal-corruption-prevention.md) | WAL Corruption Prevention | Accepted | 2025-06-22 |
| [ADR-029](./ADR-029-intelligent-recovery-system.md) | Intelligent Recovery System | Accepted | 2025-06-22 |
| [ADR-030](./ADR-030-circuit-breaker-feedback-loop-prevention.md) | Circuit Breaker Feedback Loop Prevention | Accepted | 2025-06-22 |
| [ADR-031](./ADR-031-bar-raising-metrics-retention-contention-fix.md) | Bar-Raising Metrics Retention Contention Fix | Accepted | 2025-06-22 |
| [ADR-032](./ADR-032-entity-deletion-index-tracking.md) | Entity Deletion Index Tracking | Accepted | 2025-06-22 |
| [ADR-033](./ADR-033-metrics-feedback-loop-prevention.md) | Metrics Feedback Loop Prevention | Accepted | 2025-06-23 |
| [ADR-034](./ADR-034-production-readiness-certification.md) | Production Readiness Certification | Accepted | 2025-06-23 |
| [ADR-035](./ADR-035-development-status-usage-notification.md) | Development Status & Usage Notification | Accepted | 2025-06-23 |

## ADR Template

New ADRs should follow the template format defined in the main architecture directory. Key sections include:

1. **Status** - Proposed, Accepted, Deprecated, or Superseded
2. **Context** - What is the issue that we're seeing that is motivating this decision or change
3. **Decision** - What is the change that we're proposing or have agreed to implement
4. **Consequences** - What becomes easier or more difficult to do and any risks introduced

## Related Documentation

- [Architecture Overview](../README.md) - Complete architecture documentation
- [System Overview](../01-system-overview.md) - High-level system architecture
- [Technical Specifications](../../reference/technical-specifications.md) - Detailed technical specifications

## Contributing

When proposing new architectural decisions:

1. Create a new ADR file using the format `ADR-XXX-title.md`
2. Use the next available number in sequence
3. Follow the established template format
4. Update this index with the new ADR
5. Submit as part of your pull request