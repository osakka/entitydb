# EntityDB Documentation Cross-References

This document establishes the cross-referencing system between related documentation files, ensuring users can easily navigate between connected concepts.

## Core Documentation Flow

### For New Users
**Entry Point**: [Quick Start Guide](guides/quick-start.md)
**Flow**: Quick Start → [API Reference](api/api-reference.md) → [Architecture Overview](architecture/arch-overview.md)

### For Developers  
**Entry Point**: [Contributing Guide](development/contributing.md)
**Flow**: Contributing → [Git Workflow](development/git-workflow.md) → [Implementation Guides](implementation/README.md)

### For Operations
**Entry Point**: [Deployment Guide](guides/deployment.md)
**Flow**: Deployment → [Configuration](development/configuration-management.md) → [Admin Interface](guides/admin-interface.md)

## Architecture Cross-References

### System Architecture
- **Main Document**: [Architecture Overview](architecture/arch-overview.md)
- **Related Documents**:
  - [Temporal Architecture](architecture/arch-temporal.md) - Time-based storage system
  - [RBAC Architecture](architecture/arch-rbac.md) - Security and permissions
  - [Entity Architecture](architecture/entities.md) - Entity model design
  - [Tag System](architecture/tags.md) - Tag-based data model
  - [Storage Layer Implementation](implementation/impl-performance.md) - Performance optimizations

### Temporal System
- **Main Document**: [Temporal Architecture](architecture/arch-temporal.md)
- **Related Documents**:
  - [Temporal Implementation](implementation/impl-temporal.md) - Technical implementation details
  - [Temporal Demo](api/auth_temporal_demo.md) - Time-travel query examples
  - [Performance Optimization](implementation/impl-performance.md) - Temporal query performance
  - [Temporal Tag Fix](troubleshooting/temporal-tag-fix.md) - Common problems

### Security & RBAC
- **Main Document**: [RBAC Architecture](architecture/arch-rbac.md)
- **Related Documents**:
  - [Authentication API](api/auth.md) - Login and session management
  - [Security Guide](guides/security.md) - Security setup guide
  - [Admin Interface](guides/admin-interface.md) - User administration
  - [Security Implementation](development/security-implementation.md) - Security details

## API Documentation Cross-References

### Complete API Reference
- **Main Document**: [API Reference](api/api-reference.md)
- **Related Documents**:
  - [Authentication Flow](api/auth.md) - Login and session details
  - [Entity Operations](api/api-entities.md) - CRUD operations detail
  - [Temporal Queries](api/api-temporal.md) - Time-travel API usage
  - [Relationship API](api/api-relationships.md) - Entity relationships
  - [Examples & Tutorials](api/api-examples.md) - Real-world usage patterns

### Entity Operations
- **Main Document**: [Entity API](api/api-entities.md)
- **Related Documents**:
  - [Entity Architecture](architecture/arch-overview.md#entity-model) - Entity model design
  - [Auto-chunking Feature](features/feature-autochunking.md) - Large file handling
  - [Entity Implementation](implementation/impl-entities.md) - Technical details
  - [Entity Troubleshooting](troubleshooting/trouble-entities.md) - Common issues

### Temporal API
- **Main Document**: [Temporal API](api/api-temporal.md)
- **Related Documents**:
  - [Temporal Architecture](architecture/arch-temporal.md) - System design
  - [Temporal Implementation](implementation/impl-temporal.md) - Technical details
  - [Time-travel Examples](api/api-examples.md#temporal-queries) - Usage patterns
  - [Performance Tuning](performance/perf-temporal.md) - Optimization guide

## Implementation Cross-References

### Auto-chunking System
- **Main Document**: [Auto-chunking Implementation](implementation/impl-autochunking.md)
- **Related Documents**:
  - [Auto-chunking Feature](features/feature-autochunking.md) - User perspective
  - [Auto-chunking API](api/api-entities.md#chunked-content) - API usage
  - [Performance Impact](performance/perf-optimization.md#chunking) - Performance considerations
  - [Chunking Troubleshooting](troubleshooting/trouble-chunking.md) - Common issues

### Performance Optimizations
- **Main Document**: [Performance Implementation](implementation/impl-performance.md)
- **Related Documents**:
  - [Performance Benchmarks](performance/perf-benchmarks.md) - Test results
  - [Performance Tuning](performance/perf-tuning.md) - Configuration options
  - [Memory Optimization](implementation/impl-memory.md) - Memory management
  - [High-Performance Mode](deployment/ops-performance.md) - Production settings

### Dataset System
- **Main Document**: [Dataset Implementation](implementation/impl-dataset.md)
- **Related Documents**:
  - [Dataset Architecture](architecture/arch-dataset.md) - Design overview
  - [Dataset API](api/api-dataset.md) - Multi-tenant operations
  - [Dataset Security](guides/guide-dataset-security.md) - Isolation and permissions
  - [Migration to Datasets](guides/guide-dataset-migration.md) - Upgrade guide

## Development Cross-References

### Contributing & Development
- **Main Document**: [Contributing Guide](development/dev-contributing.md)
- **Related Documents**:
  - [Development Setup](development/dev-setup.md) - Environment configuration
  - [Coding Standards](development/dev-standards.md) - Code quality guidelines
  - [Testing Guide](development/dev-testing.md) - Test framework usage
  - [Git Workflow](development/dev-git-workflow.md) - Version control process

### Logging & Debugging
- **Main Document**: [Logging Standards](development/dev-logging.md)
- **Related Documents**:
  - [Logging Implementation](implementation/impl-logging.md) - Technical implementation
  - [Troubleshooting Guide](troubleshooting/README.md) - Debugging procedures
  - [Monitoring Setup](deployment/ops-monitoring.md) - Production logging
  - [Debug Commands Reference](troubleshooting/trouble-debug-commands.md) - Diagnostic tools

### Configuration Management
- **Main Document**: [Configuration System](development/dev-configuration.md)
- **Related Documents**:
  - [Configuration Implementation](implementation/impl-configuration.md) - Technical details
  - [Configuration Reference](deployment/ops-configuration.md) - All configuration options
  - [Environment Setup](deployment/ops-environment.md) - Production configuration
  - [Configuration Troubleshooting](troubleshooting/trouble-config.md) - Common issues

## Feature Cross-References

### Temporal Features
- **Main Document**: [Temporal Features](features/feature-temporal.md)
- **Related Documents**:
  - [Temporal Architecture](architecture/arch-temporal.md) - System design
  - [Temporal API](api/api-temporal.md) - API usage
  - [Temporal Implementation](implementation/impl-temporal.md) - Technical details
  - [Temporal Performance](performance/perf-temporal.md) - Optimization

### RBAC System
- **Main Document**: [RBAC Features](features/feature-rbac.md)
- **Related Documents**:
  - [RBAC Architecture](architecture/arch-rbac.md) - Security design
  - [Authentication API](api/api-authentication.md) - Login and permissions
  - [Security Guide](guides/guide-security.md) - Setup and configuration
  - [RBAC Implementation](implementation/impl-rbac.md) - Technical details

### Metrics & Monitoring
- **Main Document**: [Metrics Features](features/feature-metrics.md)
- **Related Documents**:
  - [Metrics API](api/api-metrics.md) - Monitoring endpoints
  - [Monitoring Setup](deployment/ops-monitoring.md) - Production monitoring
  - [Metrics Implementation](implementation/impl-metrics.md) - Technical implementation
  - [Performance Monitoring](performance/perf-monitoring.md) - Performance metrics

## Performance Cross-References

### Benchmarks & Results
- **Main Document**: [Performance Benchmarks](performance/perf-benchmarks.md)
- **Related Documents**:
  - [Performance Implementation](implementation/impl-performance.md) - Optimization techniques
  - [Performance Tuning](performance/perf-tuning.md) - Configuration options
  - [High-Performance Setup](deployment/ops-performance.md) - Production optimization
  - [Performance Monitoring](performance/perf-monitoring.md) - Observability

### Optimization Guide
- **Main Document**: [Performance Optimization](performance/perf-optimization.md)
- **Related Documents**:
  - [Memory Optimization](implementation/impl-memory.md) - Memory management
  - [Storage Optimization](implementation/impl-storage.md) - Storage performance
  - [Query Optimization](performance/perf-queries.md) - Query performance
  - [Configuration Tuning](deployment/ops-performance.md) - Production settings

## Troubleshooting Cross-References

### Common Issues
- **Main Document**: [Common Issues](troubleshooting/trouble-common.md)
- **Related Documents**:
  - [Authentication Problems](troubleshooting/trouble-auth.md) - Login issues
  - [Performance Issues](troubleshooting/trouble-performance.md) - Slow queries
  - [Storage Problems](troubleshooting/trouble-storage.md) - Data integrity
  - [Configuration Issues](troubleshooting/trouble-config.md) - Setup problems

### Debug Procedures
- **Main Document**: [Debug Commands](troubleshooting/trouble-debug-commands.md)
- **Related Documents**:
  - [Logging Guide](development/dev-logging.md) - Log analysis
  - [Monitoring Tools](deployment/ops-monitoring.md) - System observability
  - [Performance Analysis](performance/perf-analysis.md) - Performance debugging
  - [System Health](deployment/ops-health.md) - Health checks

## Deployment Cross-References

### Installation & Setup
- **Main Document**: [Installation Guide](deployment/ops-installation.md)
- **Related Documents**:
  - [Quick Start](guides/guide-quick-start.md) - Simple setup
  - [Configuration Reference](deployment/ops-configuration.md) - Environment setup
  - [Security Setup](deployment/ops-security.md) - Production security
  - [Performance Setup](deployment/ops-performance.md) - Optimization

### Operations & Maintenance
- **Main Document**: [Operations Guide](deployment/ops-operations.md)
- **Related Documents**:
  - [Monitoring Setup](deployment/ops-monitoring.md) - System observability
  - [Backup & Recovery](deployment/ops-backup.md) - Data protection
  - [Scaling Guide](deployment/ops-scaling.md) - Performance scaling
  - [Maintenance Procedures](deployment/ops-maintenance.md) - Routine tasks

## External References

### Code Locations
- **Main Server**: `src/main.go`
- **API Handlers**: `src/api/`
- **Storage Engine**: `src/storage/binary/`
- **Models**: `src/models/`
- **Configuration**: `src/config/`

### Test Locations
- **Unit Tests**: `src/tests/`
- **Integration Tests**: `tests/`
- **Performance Tests**: `tests/performance/`
- **API Tests**: `tests/api/`

### Configuration Files
- **Default Config**: `share/config/entitydb.env`
- **Instance Config**: `var/entitydb.env`
- **Docker Config**: `docker-compose.yml`
- **CI Config**: `.github/workflows/`

## Navigation Helpers

### Quick Access by Topic

| Topic | Primary Document | Key Related Documents |
|-------|------------------|----------------------|
| **Getting Started** | [Quick Start](guides/guide-quick-start.md) | Installation, API Reference, Architecture |
| **API Usage** | [API Reference](api/api-reference.md) | Authentication, Entities, Temporal, Examples |
| **Architecture** | [Architecture Overview](architecture/arch-overview.md) | Temporal, RBAC, Dataset, Performance |
| **Development** | [Contributing Guide](development/dev-contributing.md) | Setup, Standards, Testing, Logging |
| **Operations** | [Installation Guide](deployment/ops-installation.md) | Configuration, Monitoring, Scaling |
| **Performance** | [Performance Benchmarks](performance/perf-benchmarks.md) | Optimization, Tuning, Implementation |
| **Troubleshooting** | [Common Issues](troubleshooting/trouble-common.md) | Auth, Performance, Storage, Debug |

### Documentation Maintenance

When updating documentation:

1. **Check Cross-References**: Ensure related documents are updated
2. **Update Navigation**: Modify this file if new relationships are created
3. **Verify Links**: Test all internal links work correctly
4. **Update Timestamps**: Maintain "Last Updated" dates consistently

---

This cross-referencing system ensures users can efficiently navigate the complete EntityDB documentation library to find related information and comprehensive coverage of their topics of interest.