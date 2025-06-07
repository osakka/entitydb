# EntityDB Documentation

> **EntityDB** is a high-performance temporal database where every tag is timestamped with nanosecond precision. All data is stored in a custom binary format (EBF) with Write-Ahead Logging for durability and concurrent access support.

## 📚 Documentation Index

### 🚀 Quick Start
- [Installation Guide](guides/guide-installation.md) - Get EntityDB running in minutes
- [Quick Start Guide](guides/guide-quick-start.md) - Basic usage and first steps
- [API Overview](api/README.md) - REST API introduction

### 📖 User Guides
- [Installation & Setup](guides/guide-installation.md)
- [Quick Start](guides/guide-quick-start.md)
- [Administration](guides/guide-admin.md)
- [Security Configuration](guides/guide-security.md)
- [Migration Procedures](guides/guide-migration.md)

### 🔌 API Reference
- [Complete API Reference](api/api-reference.md) - All endpoints with examples
- [Authentication](api/api-authentication.md) - Auth flow and session management
- [Entity Operations](api/api-entities.md) - CRUD operations and queries
- [Temporal Queries](api/api-temporal.md) - Time-travel and history queries
- [Relationships](api/api-relationships.md) - Entity relationship operations
- [Examples & Tutorials](api/api-examples.md) - Real-world usage examples

### 🏗️ Architecture
- [System Overview](architecture/arch-overview.md) - High-level architecture
- [Temporal System](architecture/arch-temporal.md) - Time-based data storage
- [RBAC Security](architecture/arch-rbac.md) - Role-based access control
- [Storage Layer](architecture/arch-storage.md) - Binary format and indexing
- [Dataspace Architecture](architecture/arch-dataspace.md) - Multi-tenant design

### ⚡ Features
- [Auto-chunking](features/feature-autochunking.md) - Large file handling
- [Temporal Queries](features/feature-temporal.md) - Time-travel capabilities
- [RBAC System](features/feature-rbac.md) - Security and permissions
- [Metrics & Monitoring](features/feature-metrics.md) - Observability
- [Entity Relationships](features/feature-relationships.md) - Data modeling

### 👨‍💻 Development
- [Developer Setup](development/dev-setup.md) - Environment configuration
- [Contributing Guide](development/dev-contributing.md) - How to contribute
- [Coding Standards](development/dev-standards.md) - Code quality guidelines
- [Testing Guide](development/dev-testing.md) - Test framework usage
- [Logging Standards](development/dev-logging.md) - Logging best practices
- [Configuration System](development/dev-configuration.md) - Config management

### 🚀 Deployment & Operations
- [Installation](deployment/ops-installation.md) - Production deployment
- [Configuration](deployment/ops-configuration.md) - Environment setup
- [Monitoring](deployment/ops-monitoring.md) - Health checks and metrics
- [Backup & Recovery](deployment/ops-backup.md) - Data protection
- [Scaling](deployment/ops-scaling.md) - Performance tuning

### 📊 Performance
- [Benchmarks](performance/perf-benchmarks.md) - Performance testing results
- [Optimization Guide](performance/perf-optimization.md) - Tuning recommendations
- [Monitoring](performance/perf-monitoring.md) - Performance observability
- [Tuning Parameters](performance/perf-tuning.md) - Configuration options

### 🔧 Troubleshooting
- [Common Issues](troubleshooting/trouble-common.md) - Frequent problems and solutions
- [Authentication Problems](troubleshooting/trouble-auth.md) - Login and permission issues
- [Performance Issues](troubleshooting/trouble-performance.md) - Slow queries and optimization
- [Storage Problems](troubleshooting/trouble-storage.md) - Data integrity and corruption

### 📋 Release Information
- [Release Notes](releases/CHANGELOG.md) - Version history and changes
- [Migration Notes](releases/migration-notes.md) - Version upgrade procedures
- [Release Process](releases/release-process.md) - How releases are managed

---

## 🎯 For New Users

**Start here:** [Quick Start Guide](guides/guide-quick-start.md) → [API Reference](api/api-reference.md) → [Architecture Overview](architecture/arch-overview.md)

## 🔧 For Developers

**Start here:** [Developer Setup](development/dev-setup.md) → [Contributing Guide](development/dev-contributing.md) → [Coding Standards](development/dev-standards.md)

## 🚀 For Operations

**Start here:** [Installation Guide](deployment/ops-installation.md) → [Configuration Reference](deployment/ops-configuration.md) → [Monitoring Setup](deployment/ops-monitoring.md)

---

## 📝 Documentation Standards

This documentation follows industry-standard practices:

- **Consistent Naming**: All files use kebab-case naming (e.g., `guide-quick-start.md`)
- **Cross-References**: Extensive linking between related topics
- **Accuracy**: All documentation verified against actual codebase (v2.27.0)
- **Maintenance**: Regular updates ensure documentation stays current
- **Accessibility**: Clear structure suitable for all experience levels

## 🤝 Contributing to Documentation

Found an error or want to improve the documentation? See our [Contributing Guide](development/dev-contributing.md) for:

- How to submit documentation improvements
- Writing style guidelines
- Review process
- Maintenance procedures

---

**Version**: 2.27.0  
**Last Updated**: June 7, 2025  
**Next Review**: December 2025