# EntityDB Architecture Documentation

This directory contains the formal architectural decision records and architecture overview for EntityDB.

## 📋 **Architectural Decision Records (ADRs)**

**Location**: `./adr/`  
**Purpose**: Formal documentation of architectural decisions made during EntityDB development  
**Format**: `ADR-XXX-title.md`  
**Count**: 10 formal ADRs (ADR-000 through ADR-035)

**[🔗 View Complete ADR Index](./adr/README.md)**

### Current ADR Status

| ADR | Title | Status | Date | Impact |
|-----|-------|---------|------|--------|
| [ADR-000](./adr/ADR-000-MASTER-TIMELINE.md) | Master Timeline | Accepted | 2025-06-23 | Project timeline foundation |
| [ADR-022](./adr/ADR-022-database-file-unification.md) | Database File Unification | Accepted | 2025-06-20 | **BREAKING** - Single .edb format |
| [ADR-028](./adr/ADR-028-wal-corruption-prevention.md) | WAL Corruption Prevention | Accepted | 2025-06-22 | Multi-layer defense system |
| [ADR-029](./adr/ADR-029-intelligent-recovery-system.md) | Intelligent Recovery System | Accepted | 2025-06-22 | Automatic corruption recovery |
| [ADR-030](./adr/ADR-030-circuit-breaker-feedback-loop-prevention.md) | Circuit Breaker Architecture | Accepted | 2025-06-22 | CPU feedback loop prevention |
| [ADR-031](./adr/ADR-031-bar-raising-metrics-retention-contention-fix.md) | Metrics Retention Fix | Accepted | 2025-06-22 | Authentication delay elimination |
| [ADR-032](./adr/ADR-032-entity-deletion-index-tracking.md) | Entity Deletion Tracking | Accepted | 2025-06-22 | Proper deletion index management |
| [ADR-033](./adr/ADR-033-metrics-feedback-loop-prevention.md) | Metrics Feedback Loop Prevention | Accepted | 2025-06-23 | Final surgical precision fix |
| [ADR-034](./adr/ADR-034-production-readiness-certification.md) | Production Readiness Certification | Accepted | 2025-06-23 | E2E testing and validation |
| [ADR-035](./adr/ADR-035-development-status-usage-notification.md) | Development Status & Usage Notification | Accepted | 2025-06-23 | Licensing and community building |

## 🏗️ **Architecture Overview**

EntityDB follows a unified temporal database architecture with these core principles:

### **Single Source of Truth**
- All data stored in unified `.edb` files (ADR-022)
- No separate database, WAL, or index files
- Elimination of parallel implementations

### **Temporal Excellence**
- Nanosecond-precision timestamps for all operations
- Complete temporal query functionality (as-of, history, diff, changes)
- Immutable audit trail with time-travel capabilities

### **Production Readiness**
- Comprehensive corruption prevention (ADR-028, ADR-029, ADR-030)
- Metrics feedback loop elimination (ADR-033)
- End-to-end production validation (ADR-034)
- Memory optimization for 1GB RAM deployments

### **Security Architecture**
- Tag-based RBAC with fine-grained permissions
- JWT authentication with embedded credentials
- Comprehensive input validation and sanitization

## 📚 **Related Documentation**

### **System Overview**
- **[System Architecture](../assets/architecture.svg)** - Visual architecture diagram
- **[API Reference](../api-reference/README.md)** - Complete API documentation
- **[Getting Started](../getting-started/README.md)** - Installation and setup guides

### **Technical Specifications**
- **[Configuration Reference](../reference/01-configuration-reference.md)** - Complete configuration options
- **[Binary Format Specification](../reference/03-binary-format-spec.md)** - EntityDB Unified File Format (EUFF)
- **[RBAC Reference](../reference/04-rbac-reference.md)** - Permission system details

### **Historical Documentation**
- **[Numbered Architecture Docs](../archive/numbered-architecture/)** - Historical architecture documentation
- **[Legacy ADRs](../archive/legacy-adrs/)** - Superseded decision records

## 🔄 **ADR Process**

### **Creating New ADRs**
1. Use the next sequential number (ADR-036, ADR-037, etc.)
2. Follow the [ADR template](./template.md)
3. Update the [ADR index](./adr/README.md)
4. Reference the [master timeline](./adr/ADR-000-MASTER-TIMELINE.md)

### **ADR Lifecycle**
- **Proposed**: Under discussion
- **Accepted**: Implemented and active
- **Deprecated**: No longer recommended
- **Superseded**: Replaced by newer ADR

### **Architectural Principles**
- **Never reverse decisions** without explicit ADR
- **Document all architectural changes**
- **Maintain single source of truth**
- **Follow surgical precision implementation**

## 📈 **Current Architecture Status**

**Version**: v2.34.3 (Production Certified)  
**Status**: Production Ready  
**Last Major Decision**: ADR-035 (Usage Notification License)  
**Total Decisions**: 50+ architectural decisions tracked  
**Breaking Changes**: 2 (Authentication v2.29.0, File Unification v2.32.6)

## 🤝 **Contributing**

When proposing architectural changes:
1. **Check existing ADRs** to avoid conflicts
2. **Follow the decision template** for consistency
3. **Reference the master timeline** for context
4. **Consider impact** on existing decisions
5. **Update all related documentation**

---

**Maintainers**: Architecture Team  
**Last Updated**: 2025-06-23  
**Next Review**: 2025-12-23