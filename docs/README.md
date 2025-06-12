# EntityDB Documentation Library

> **Version**: v2.29.0 | **Last Updated**: 2025-06-11

Welcome to the EntityDB documentation library. This is your comprehensive guide to understanding, deploying, and developing with EntityDB - a high-performance temporal database where every tag is timestamped with nanosecond precision.

## 🎉 New in v2.29.0

- **Complete UI/UX Overhaul**: Professional web interface with Vue.js 3, dark mode, and modern components
- **Comprehensive Documentation**: Audit reports, migration plans, and implementation guides
- **Enhanced API Reference**: Updated with all endpoints and authentication changes
- **Improved Architecture Guides**: Detailed temporal, RBAC, and performance documentation

## ⚠️ Critical Notice for v2.29.0

**BREAKING CHANGE**: The authentication architecture has fundamentally changed in v2.29.0. User credentials are now stored directly in the user entity's content field as `salt|bcrypt_hash`. This change has **NO BACKWARD COMPATIBILITY** - all existing users must be recreated. See the [Authentication Migration Guide](./api/auth.md#v229-migration) for details.

## 🚀 Quick Navigation

### For New Users
1. **[Installation Guide](./10-getting-started/01-installation.md)** - Get EntityDB running
2. **[Quick Start Tutorial](./10-getting-started/02-quick-start.md)** - First steps with EntityDB
3. **[Core Concepts](./10-getting-started/03-core-concepts.md)** - Understand entities, tags, and temporal data
4. **[API Basics](./30-api-reference/01-overview.md)** - Make your first API calls

### For Developers
1. **[Development Setup](./60-developer-guides/01-development-setup.md)** - Set up your development environment
2. **[API Reference](./30-api-reference/README.md)** - Complete API documentation
3. **[Architecture Guide](./20-architecture/README.md)** - System design and internals
4. **[Contributing](./60-developer-guides/02-contributing.md)** - How to contribute to EntityDB

### For Administrators
1. **[Production Deployment](./70-deployment/01-production-deployment.md)** - Deploy EntityDB in production
2. **[Configuration Reference](./90-reference/01-configuration.md)** - All configuration options
3. **[Security Guide](./50-admin-guides/02-security.md)** - Secure your EntityDB installation
4. **[Monitoring & Metrics](./50-admin-guides/03-monitoring.md)** - Monitor system health

## 📚 Documentation Categories

### [00-overview](./00-overview/)
High-level introduction to EntityDB, its features, and capabilities.

### [10-getting-started](./10-getting-started/)
Installation, quick start guides, and basic tutorials for new users.

### [20-architecture](./20-architecture/)
System architecture, design decisions, and technical deep-dives.

### [30-api-reference](./30-api-reference/)
Complete API documentation with examples for all endpoints.

### [40-user-guides](./40-user-guides/)
Task-oriented guides for common EntityDB operations.

### [50-admin-guides](./50-admin-guides/)
Administration, maintenance, and operational guides.

### [60-developer-guides](./60-developer-guides/)
Development setup, contribution guidelines, and extending EntityDB.

### [70-deployment](./70-deployment/)
Production deployment, scaling, and infrastructure guides.

### [80-troubleshooting](./80-troubleshooting/)
Common issues, error messages, and resolution guides.

### [90-reference](./90-reference/)
Technical specifications, configuration options, and detailed references.

## 🔍 Finding Information

### By Feature
- **Temporal Queries**: [Temporal Features Guide](./40-user-guides/04-temporal-queries.md)
- **RBAC & Security**: [RBAC Architecture](./20-architecture/03-rbac.md) | [Security Guide](./50-admin-guides/02-security.md)
- **Datasets**: [Dataset Management](./40-user-guides/02-dataset-management.md)
- **Relationships**: [Entity Relationships](./40-user-guides/03-entity-relationships.md)
- **Metrics**: [Metrics System](./20-architecture/05-metrics.md) | [Monitoring Guide](./50-admin-guides/03-monitoring.md)

### By Task
- **Create a user**: [User Management](./50-admin-guides/01-user-management.md#creating-users)
- **Query entities**: [Query Guide](./40-user-guides/05-query-guide.md)
- **Set up SSL**: [SSL Configuration](./70-deployment/03-ssl-setup.md)
- **Debug performance**: [Performance Troubleshooting](./80-troubleshooting/02-performance.md)

### By API Endpoint
- **Authentication**: [/api/v1/auth/*](./30-api-reference/02-authentication.md)
- **Entities**: [/api/v1/entities/*](./30-api-reference/03-entities.md)
- **Datasets**: [/api/v1/datasets/*](./30-api-reference/04-datasets.md)
- **Metrics**: [/api/v1/metrics/*](./30-api-reference/05-metrics.md)

## 📋 Recent Changes

### v2.29.0 (Current)
- **Breaking**: New authentication architecture
- **Feature**: Renamed "dataspace" to "dataset" throughout
- **Enhancement**: Comprehensive documentation overhaul

### v2.28.0
- Enhanced metrics system with retention policies
- Connection stability improvements
- Professional documentation library

See [CHANGELOG.md](../CHANGELOG.md) for complete version history.

## 🤝 Contributing to Documentation

We welcome documentation improvements! Please:
1. Follow the [Documentation Standards](./90-reference/10-documentation-standards.md)
2. Use the defined [Naming Conventions](./90-reference/11-naming-conventions.md)
3. Update the relevant index files when adding new documents
4. Ensure all code examples are tested and working

## 📊 Documentation Status

- **Total Documents**: 231
- **Last Full Review**: 2025-06-11
- **Coverage**: 95% of features documented
- **Examples**: 150+ code examples
- **Diagrams**: 25+ architecture diagrams

## 🔗 External Resources

- **Repository**: https://git.home.arpa/itdlabs/entitydb
- **Issues**: Report documentation issues in the main repository
- **Discussions**: Join our community discussions

---

*This documentation is maintained by the EntityDB team and community. For corrections or improvements, please submit a pull request.*