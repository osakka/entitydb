# EntityDB Documentation

> **EntityDB** is a high-performance temporal database where every tag is timestamped with nanosecond precision. All data is stored in a custom binary format (EBF) with Write-Ahead Logging for durability and concurrent access support.

## 📚 Documentation Index

### 🚀 Quick Start
- [Installation Guide](guides/deployment.md) - Get EntityDB running in minutes
- [Quick Start Guide](guides/quick-start.md) - Basic usage and first steps
- [API Reference](api/api-reference.md) - REST API documentation

### 📖 User Guides
- [Installation & Setup](guides/deployment.md)
- [Quick Start](guides/quick-start.md)
- [Administration](guides/admin-interface.md)
- [Security Configuration](guides/security.md)
- [Migration Procedures](guides/migration.md)

### 🔌 API Reference
- [Complete API Reference](api/api-reference.md) - All endpoints with examples
- [Authentication](api/auth.md) - Auth flow and session management
- [Entity Operations](api/entities.md) - CRUD operations and queries
- [Temporal Queries](api/auth_temporal_demo.md) - Time-travel and history queries
- [Query API](api/query_api.md) - Advanced query operations
- [Examples & Tutorials](api/examples.md) - Real-world usage examples

### 🏗️ Architecture
- [System Overview](architecture/arch-overview.md) - High-level architecture
- [Temporal System](architecture/arch-temporal.md) - Time-based data storage
- [RBAC Security](architecture/arch-rbac.md) - Role-based access control
- [Entity Model](architecture/entities.md) - Core entity architecture
- [Tag System](architecture/tags.md) - Tag-based data model

### ⚡ Features
- [Auto-chunking](features/autochunking.md) - Large file handling
- [Temporal Features](features/temporal-features.md) - Time-travel capabilities
- [Query System](features/query-implementation.md) - Advanced query capabilities
- [Configuration System](features/config-system.md) - Three-tier configuration
- [Widget System](features/widget-system.md) - UI components

### 👨‍💻 Development
- [Contributing Guide](development/contributing.md) - How to contribute
- [Git Workflow](development/git-workflow.md) - Version control process
- [Logging Standards](development/logging-standards.md) - Logging best practices
- [Configuration Management](development/configuration-management.md) - Config system
- [Production Notes](development/production-notes.md) - Deployment guidelines
- [Security Implementation](development/security-implementation.md) - Security practices

### 🚀 Deployment & Operations
- [Deployment Guide](guides/deployment.md) - Production deployment
- [Configuration Reference](development/configuration-management.md) - Environment setup
- [Admin Interface](guides/admin-interface.md) - Management dashboard
- [Security Guide](guides/security.md) - Security configuration
- [Migration Guide](guides/migration.md) - Version upgrades

### 📊 Performance
- [Performance Overview](performance/performance.md) - Performance characteristics
- [Performance Index](performance/performance-index.md) - Optimization catalog
- [100x Performance Plan](performance/100x-performance-plan.md) - Optimization strategy
- [Temporal Performance](performance/temporal-performance.md) - Temporal query optimization

### 🔧 Troubleshooting
- [Content Format Issues](troubleshooting/content-format-troubleshooting.md) - Data format problems
- [SSL Configuration](troubleshooting/ssl-configuration.md) - HTTPS setup issues
- [Temporal Tag Issues](troubleshooting/temporal-tag-fix.md) - Temporal system problems
- [Tag Index Persistence](troubleshooting/tag-index-persistence-bug.md) - Index issues

### 📋 Release Information
- [Latest Release Notes](releases/release-notes-v2.14.0.md) - Current version changes
- [v2.13.1 Release](releases/release-notes-v2.13.1.md) - Previous stable release
- [v2.13.0 Release](releases/release-notes-v2.13.0.md) - Major feature release
- [v2.12.0 Release](releases/release-notes-v2.12.0.md) - Auto-chunking release

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
- **Accuracy**: All documentation verified against actual codebase (v2.28.0)
- **Maintenance**: Regular updates ensure documentation stays current
- **Accessibility**: Clear structure suitable for all experience levels

## 🤝 Contributing to Documentation

Found an error or want to improve the documentation? See our [Contributing Guide](development/dev-contributing.md) for:

- How to submit documentation improvements
- Writing style guidelines
- Review process
- Maintenance procedures

---

**Version**: 2.28.0  
**Last Updated**: June 7, 2025  
**Next Review**: December 2025

## 📂 Documentation Structure

The EntityDB documentation follows a clear, maintainable organization:

- **api/** - API reference documentation and examples
- **architecture/** - System design and architecture documents
- **development/** - Developer guides and contribution documentation
- **features/** - Feature-specific documentation and guides
- **guides/** - User guides and tutorials
- **implementation/** - Technical implementation details
- **performance/** - Performance analysis and optimization guides
- **releases/** - Release notes and version history
- **troubleshooting/** - Problem-solving guides and solutions

For detailed navigation between related topics, see [Cross References](CROSS_REFERENCES.md).
For documentation maintenance procedures, see [Maintenance Guidelines](MAINTENANCE_GUIDELINES.md).