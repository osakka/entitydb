# EntityDB Architecture Documentation

This directory contains comprehensive architecture documentation for EntityDB, including system design decisions, architectural patterns, and formal decision records.

## Documentation Organization

### 📋 **Architectural Decision Records (ADRs)**
**Location**: `./adr/`  
**Purpose**: Formal documentation of architectural decisions  
**Format**: ADR-XXX-title.md  
**Count**: 6 formal ADRs  
**[View ADR Index](./adr/README.md)**

### 🏗️ **Architecture Decisions**
**Location**: `./` (numbered files 001-035)  
**Purpose**: Detailed technical architecture documentation  
**Format**: XXX-title.md  
**Count**: 35+ architecture documents

### 📊 **System Overview**
**Location**: High-level architecture files  
**Purpose**: System-wide architectural patterns and designs

## Core Architecture Components

### Temporal Database Architecture
- [001 - Temporal Tag Storage](./001-temporal-tag-storage.md) - Nanosecond-precision timestamp system
- [002 - Binary Storage Format](./002-binary-storage-format.md) - Custom EBF format design
- [010 - Temporal Functionality Completion](./010-temporal-functionality-completion.md) - Complete temporal query system

### Storage and Performance
- [003 - Unified Sharded Indexing](./003-unified-sharded-indexing.md) - 256-shard concurrent indexing
- [026 - Unified File Format Architecture](./026-unified-file-format-architecture.md) - Single .edb file format
- [027 - Complete Database File Unification](./027-complete-database-file-unification.md) - Elimination of separate files

### Security Architecture
- [004 - Tag-Based RBAC](./004-tag-based-rbac.md) - Role-based access control system
- [006 - Credential Storage in Entities](./006-credential-storage-in-entities.md) - Authentication architecture
- [034 - Security Architecture Evolution](./034-security-architecture-evolution.md) - Security system evolution

### Platform Design
- [005 - Application-Agnostic Design](./005-application-agnostic-design.md) - Platform architecture
- [008 - Three-Tier Configuration](./008-three-tier-configuration.md) - Configuration management
- [014 - Single Source of Truth Enforcement](./014-single-source-of-truth-enforcement.md) - Architectural principles

## Formal ADR Records

| ADR | Title | Status | Focus Area |
|-----|-------|---------|------------|
| [ADR-000](./adr/ADR-000-MASTER-TIMELINE.md) | Master Timeline | Accepted | Project Timeline |
| [ADR-022](./adr/ADR-022-database-file-unification.md) | Database File Unification | Accepted | Storage Architecture |
| [ADR-028](./adr/ADR-028-wal-corruption-prevention.md) | WAL Corruption Prevention | Accepted | Data Integrity |
| [ADR-029](./adr/ADR-029-intelligent-recovery-system.md) | Intelligent Recovery System | Accepted | System Resilience |
| [ADR-030](./adr/ADR-030-circuit-breaker-feedback-loop-prevention.md) | Circuit Breaker Feedback Loop Prevention | Accepted | Performance |
| [ADR-031](./adr/ADR-031-bar-raising-metrics-retention-contention-fix.md) | Metrics Retention Contention Fix | Accepted | Observability |

## Recent Architecture Evolution

### v2.34.x Series
- **WAL Corruption Prevention**: Revolutionary multi-layer defense system
- **Configuration Management Excellence**: Enterprise-grade three-tier hierarchy
- **Documentation Excellence**: IEEE 1063-2001 compliant architecture documentation

### v2.32.x Series
- **Database File Unification**: Single `.edb` format eliminates complexity
- **Production Battle Testing**: Comprehensive real-world validation
- **Temporal Features Completion**: All temporal query capabilities implemented

### v2.31.x Series
- **Performance Optimization Suite**: O(1) tag caching and memory optimization
- **Session Management**: Pure tag-based authentication architecture
- **Comprehensive Logging**: Enterprise-grade logging standards

## Architecture Principles

1. **Single Source of Truth**: Every concept has one authoritative implementation
2. **Unified Entity Model**: Everything is an entity with tags
3. **Temporal by Design**: All data stored with nanosecond-precision timestamps
4. **Performance Excellence**: Sub-millisecond query performance with intelligent caching
5. **Enterprise Security**: Tag-based RBAC with comprehensive audit trails
6. **Production Ready**: Battle-tested reliability with self-healing capabilities

## Related Documentation

- [System Overview](./01-system-overview.md) - High-level system architecture
- [Technical Reference](../reference/technical-specifications.md) - Detailed technical specifications
- [Developer Guide](../developer-guide/README.md) - Development architecture patterns
- [API Reference](../api-reference/README.md) - API architecture and design

## Contributing to Architecture

When proposing architectural changes:

1. **Review existing decisions** in this directory
2. **Create new ADR** for significant architectural changes
3. **Update relevant architecture documents** to reflect changes
4. **Ensure consistency** with established architectural principles
5. **Document in appropriate category** (ADR for decisions, numbered docs for detailed architecture)

For questions about EntityDB architecture, refer to the [Developer Guide](../developer-guide/README.md) or create an issue for clarification.