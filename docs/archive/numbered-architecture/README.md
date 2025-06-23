# Numbered Architecture Documentation (Archived)

This directory contains 46 numbered architecture documents (001-035 plus descriptive files) that were moved from `/docs/architecture/` to eliminate confusion with the formal ADR system.

## Why These Were Moved

**Problem**: Dual numbering systems created confusion:
- **ADRs**: `ADR-000` through `ADR-035` in `/docs/architecture/adr/`
- **Architecture Docs**: `001-` through `035-` in `/docs/architecture/`

**Solution**: Consolidated to single ADR system for architectural decisions.

## Content Organization

These archived documents contain valuable technical architecture information but are not formal decision records. They include:

### Temporal Database Architecture
- `001-temporal-tag-storage.md` - Nanosecond timestamp system details
- `002-binary-storage-format.md` - Custom EBF format specifications
- `010-temporal-functionality-completion.md` - Temporal query implementation

### Storage and Performance
- `003-unified-sharded-indexing.md` - 256-shard concurrent indexing details
- `026-unified-file-format-architecture.md` - Single .edb file format design
- `027-complete-database-file-unification.md` - File consolidation implementation

### Security and Authentication
- `004-tag-based-rbac.md` - RBAC system technical details
- `006-credential-storage-in-entities.md` - Authentication implementation
- `034-security-architecture-evolution.md` - Security system development

### Platform and Configuration
- `005-application-agnostic-design.md` - Platform architecture details
- `008-three-tier-configuration.md` - Configuration system implementation
- `014-single-source-of-truth-enforcement.md` - Architectural principle enforcement

## Current Documentation Structure

**Active ADRs**: Use `/docs/architecture/adr/` for all architectural decisions  
**Architecture Overview**: Use `/docs/architecture/README.md` for high-level architecture  
**Technical Specs**: Use `/docs/reference/` for detailed technical documentation

## Migration Notes

These documents were archived on 2025-06-23 as part of ADR consolidation. Content may be outdated compared to current implementation. For current architectural decisions, reference the formal ADR system.

**Key Principle**: Single source of truth for architectural decisions maintained in ADR format.

---

**Archived**: 2025-06-23  
**Reason**: ADR system consolidation  
**Alternative**: See `/docs/architecture/adr/` for current architectural decisions