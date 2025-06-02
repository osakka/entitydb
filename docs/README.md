# EntityDB Documentation

Welcome to the comprehensive documentation for EntityDB - a high-performance temporal database platform.

## 📚 Documentation Index

### Getting Started
- [Quick Start Guide](./guides/quick-start.md) - Get up and running in 5 minutes
- [Core Concepts](./architecture/overview.md) - Understand entities, tags, and temporal storage
- [API Examples](./api/examples.md) - Common API usage patterns

### Core Documentation
- [Requirements](./core/REQUIREMENTS.md) - System requirements and dependencies
- [Specifications](./core/SPECIFICATIONS.md) - Technical specifications
- [Current State](./core/current_state_summary.md) - Platform capabilities and status

### API Reference
- [Authentication](./api/auth.md) - Login, sessions, and tokens
- [Entity Operations](./api/entities.md) - CRUD operations for entities
- [Query API](./api/query_api.md) - Advanced querying capabilities
- [Temporal API](./api/auth_temporal_demo.md) - Time-travel queries
- [Examples](./api/examples.md) - Practical API usage examples

### Architecture
- [Overview](./architecture/overview.md) - System architecture and design
- [Entity Model](./architecture/entities.md) - Core data model
- [Temporal Architecture](./architecture/temporal_architecture.md) - Time-series design
- [Tag-Based RBAC](./architecture/tag_based_rbac.md) - Security model
- [Tag System](./architecture/tags.md) - Tag architecture and namespaces
- [RBAC Implementation](./architecture/tag_based_rbac_implementation.md) - Security implementation

### Features
- [Temporal Storage](./features/TEMPORAL_FEATURES.md) - Time-travel capabilities
- [Autochunking](./features/AUTOCHUNKING.md) - Large file handling
- [Binary Format](./features/CUSTOM_BINARY_FORMAT.md) - EBF specification
- [Query System](./features/QUERY_IMPLEMENTATION.md) - Query engine details
- [Configuration](./features/CONFIG_SYSTEM.md) - System configuration
- [Widget System](./features/WIDGET_SYSTEM.md) - UI components
- [API Testing](./features/API_TESTING_FRAMEWORK.md) - Testing framework

### Implementation
- [Status Overview](./implementation/IMPLEMENTATION_STATUS.md) - Feature completion status
- [Temporal Implementation](./implementation/TEMPORAL_IMPLEMENTATION.md) - Temporal storage details
- [Binary Format](./implementation/BINARY_FORMAT_IMPLEMENTATION.md) - Storage engine
- [Data Integrity](./implementation/DATA_INTEGRITY_COMPLETE.md) - Consistency guarantees
- [Performance Optimizations](./implementation/PERFORMANCE_OPTIMIZATION_SUMMARY.md) - Speed improvements
- [Multi-Dataspace](./implementation/MULTI_DATASPACE_ARCHITECTURE.md) - Multi-tenancy

### Performance & Metrics
- [Overview](./performance/PERFORMANCE.md) - Performance characteristics
- [Metrics Audit](./METRICS_AUDIT_FINDINGS.md) - Comprehensive metrics gap analysis
- [Metrics Action Plan](./METRICS_ACTION_PLAN.md) - Phased implementation roadmap
- [Metrics Implementation](./METRICS_IMPLEMENTATION_SUMMARY.md) - Phase 1 completion summary
- [Benchmarks](./performance/PERFORMANCE_COMPARISON.md) - Speed comparisons
- [Temporal Performance](./performance/TEMPORAL_PERFORMANCE.md) - Time-travel query speed
- [High Performance Mode](./performance/HIGH_PERFORMANCE_MODE_REPORT.md) - Optimization settings
- [100x Plan](./performance/100X_PERFORMANCE_PLAN.md) - Performance roadmap

### Development
- [Contributing](./development/contributing.md) - How to contribute
- [Git Workflow](./development/git-workflow.md) - Development process
- [Security Implementation](./development/security-implementation.md) - Security guidelines
- [Production Notes](./development/production-notes.md) - Deployment guidance

### Guides
- [Deployment](./guides/deployment.md) - Production deployment
- [Migration](./guides/migration.md) - Upgrading EntityDB
- [Admin Interface](./guides/admin-interface.md) - Web UI guide
- [Security Policy](./guides/security-policy.md) - Security best practices
- [Project Structure](./guides/project-structure.md) - Codebase organization

### Applications
- [Worca Overview](./applications/worca/README.md) - Workforce orchestrator
- [Worca Architecture](./applications/worca/WORCA_DATASPACE_ARCHITECTURE.md) - App design
- [Widget System](./applications/worca/WIDGET_SYSTEM_ARCHITECTURE.md) - UI components

### Examples
- [Temporal Queries](./examples/temporal_examples.md) - Time-travel examples
- [Ticketing System](./examples/ticketing_system.md) - Real-world use case

### Troubleshooting
- [Content Format](./troubleshooting/CONTENT_FORMAT_TROUBLESHOOTING.md) - Data issues
- [SSL Configuration](./troubleshooting/SSL_CONFIGURATION.md) - HTTPS setup
- [Tag Index Persistence](./troubleshooting/TAG_INDEX_PERSISTENCE_BUG.md) - Index issues

### Release Notes
- [v2.22.0](../CHANGELOG.md#2220---2025-06-02) - Latest release - Comprehensive metrics system
- [v2.14.0](./releases/RELEASE_NOTES_v2.14.0.md) - Major performance update
- [v2.13.1](./releases/RELEASE_NOTES_v2.13.1.md) - Bug fixes
- [v2.13.0](./releases/RELEASE_NOTES_v2.13.0.md) - Configuration overhaul
- [v2.12.0](./releases/RELEASE_NOTES_v2.12.0.md) - Unified entity model

### Archive
The [archive](./archive/) directory contains historical documentation from earlier versions. While these documents may contain outdated information, they can be valuable for understanding the evolution of EntityDB.

## 📖 Documentation Standards

### File Naming Convention
- Use UPPERCASE for emphasis: `IMPORTANT_FEATURE.md`
- Use lowercase with underscores for regular docs: `feature_guide.md`
- Release notes: `RELEASE_NOTES_vX.Y.Z.md`
- Implementation docs: `FEATURE_IMPLEMENTATION.md`

### Document Structure
Each document should include:
1. Title and brief description
2. Table of contents for longer documents
3. Clear sections with headers
4. Code examples where applicable
5. Links to related documentation

### Maintenance
- Keep documentation synchronized with code changes
- Archive outdated docs rather than deleting
- Update this index when adding new documentation
- Use relative links for internal references

## 🔍 Finding Information

### By Topic
- **Getting Started**: See [Quick Start Guide](./guides/quick-start.md)
- **API Usage**: Check [API Examples](./api/examples.md)
- **Architecture**: Read [Overview](./architecture/overview.md)
- **Performance**: Review [Performance Docs](./performance/)
- **Troubleshooting**: See [Troubleshooting Guides](./troubleshooting/)

### By Role
- **Developers**: Start with [Contributing](./development/contributing.md)
- **Operators**: Read [Deployment Guide](./guides/deployment.md)
- **Users**: See [Quick Start](./guides/quick-start.md)

## 📝 Contributing to Documentation

1. Follow the naming conventions above
2. Update this index when adding new docs
3. Keep examples up-to-date with current code
4. Archive rather than delete outdated content
5. Cross-reference related documentation

For questions or improvements, please open an issue in the repository.