# EntityDB Documentation Library

> **Version**: v2.31.0 | **Last Updated**: 2025-06-13

Welcome to the EntityDB documentation library. This is your comprehensive guide to understanding, deploying, and developing with EntityDB - a high-performance temporal database where every tag is timestamped with nanosecond precision.

## 🚀 New in v2.31.0

- **Comprehensive Performance Optimization Suite**: Enterprise-scale improvements delivering significant memory, CPU, and storage enhancements
- **O(1) Tag Value Caching**: Converted O(n) tag lookups to O(1) with intelligent lazy caching
- **Parallel Index Building**: 4-worker concurrent processing for faster server startup
- **JSON Encoder Pooling**: Reduced API allocation overhead with sync.Pool management
- **Batch Write Operations**: Configurable batching (10 entities, 100ms intervals) for improved throughput
- **Temporal Tag Variant Caching**: Pre-computed O(1) temporal tag lookups for optimized ListByTag operations
- **Code Quality Improvements**: Zero compilation warnings, fixed go vet issues, clean build system

## 🎉 Previous Release (v2.30.0)

- **Temporal Tag Search Implementation**: Complete resolution of critical temporal tag search issues with comprehensive documentation
- **Enhanced Dashboard UI**: Real-time metrics dashboard with health scoring, memory charting, and professional design
- **Performance Optimization**: Sub-millisecond queries, zero goroutine leaks, and stability improvements
- **Implementation Documentation**: Detailed technical documentation in `implementation/temporal-tag-search-implementation.md`

## ⚠️ Critical Notice for v2.29.0

**BREAKING CHANGE**: The authentication architecture has fundamentally changed in v2.29.0. User credentials are now stored directly in the user entity's content field as `salt|bcrypt_hash`. This change has **NO BACKWARD COMPATIBILITY** - all existing users must be recreated. See the [Authentication Migration Guide](./api/auth.md#v229-migration) for details.

## 🚀 Quick Navigation

### For New Users
1. **[Installation Guide](./10-getting-started/01-installation.md)** - Get EntityDB running
2. **[Quick Start Tutorial](./10-getting-started/02-quick-start.md)** - First steps with EntityDB
3. **[Core Concepts](./10-getting-started/03-core-concepts.md)** - Understand entities, tags, and temporal data
4. **[API Basics](./30-api-reference/01-overview.md)** - Make your first API calls

### For Developers
1. **[Contributing](./60-developer-guides/01-contributing.md)** - Set up your development environment
2. **[API Reference](./30-api-reference/README.md)** - Complete API documentation
3. **[Architecture Guide](./20-architecture/README.md)** - System design and internals
4. **[Git Workflow](./60-developer-guides/02-git-workflow.md)** - Development process and contribution guidelines

### For Administrators
1. **[Production Deployment](./70-deployment/01-production-deployment.md)** - Deploy EntityDB in production
2. **[User Management](./50-admin-guides/01-user-management.md)** - Create and manage users with RBAC
3. **[Monitoring Guide](./50-admin-guides/02-monitoring-guide.md)** - Monitor system health and performance
4. **[Configuration Reference](./90-reference/01-configuration-reference.md)** - All configuration options

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
- **RBAC & Security**: [RBAC Architecture](./20-architecture/03-rbac-architecture.md) | [Security Guide](./50-admin-guides/01-security-configuration.md)
- **Datasets**: [Dataset Management](./40-user-guides/02-dataset-management.md)
- **Relationships**: [Entity Relationships](./40-user-guides/03-entity-relationships.md)
- **Metrics**: [Metrics System](./20-architecture/09-metrics-architecture.md) | [Monitoring Guide](./50-admin-guides/02-deployment-guide.md)

### By Task
- **Create a user**: [Security Configuration](./50-admin-guides/01-security-configuration.md#user-management)
- **Query entities**: [Advanced Queries](./40-user-guides/04-advanced-queries.md)
- **Set up SSL**: [SSL Configuration](./70-deployment/03-ssl-setup.md)
- **Debug performance**: [Performance Troubleshooting](./80-troubleshooting/02-performance.md)

### By API Endpoint
- **Authentication**: [/api/v1/auth/*](./30-api-reference/02-authentication.md)
- **Entities**: [/api/v1/entities/*](./30-api-reference/03-entities.md)
- **Datasets**: [Dataset API](./30-api-reference/04-datasets-metrics.md#dataset-operations)
- **Metrics**: [Metrics API](./30-api-reference/04-datasets-metrics.md#metrics-api-reference)

## 📋 Recent Changes

### v2.31.0 (Current)
- **Performance**: Comprehensive optimization suite with measurable improvements
- **Enhancement**: O(1) tag value caching and parallel index building
- **Quality**: Zero compilation warnings, complete build system cleanup
- **Documentation**: Updated for v2.31.0 with performance optimization details

### v2.30.0
- **Fix**: Temporal tag search implementation complete
- **Enhancement**: Real-time metrics dashboard with professional UI
- **Performance**: Sub-millisecond queries and stability improvements

### v2.29.0
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
1. Follow the [Documentation Standards](../TAXONOMY_DESIGN_2025.md)
2. Use the defined [Naming Conventions](../TAXONOMY_DESIGN_2025.md#naming-conventions)
3. Update the relevant index files when adding new documents
4. Ensure all code examples are tested and working

## 📊 Documentation Status

- **Total Documents**: 114 (141 archived)
- **Last Full Review**: 2025-06-13
- **Coverage**: 100% of features documented
- **Version Consistency**: v2.31.0 across all files
- **Technical Accuracy**: Binary format (EBF) throughout

## 🗂️ Complete Documentation Index

### Core Documentation Categories

#### [00-overview](./00-overview/)
High-level introduction and system specifications
- **[Introduction](./00-overview/01-introduction.md)** - What is EntityDB and key concepts
- **[Specifications](./00-overview/02-specifications.md)** - Technical specifications and capabilities
- **[Requirements](./00-overview/03-requirements.md)** - System and hardware requirements

#### [10-getting-started](./10-getting-started/)
New user onboarding and tutorials
- **[Installation](./10-getting-started/01-installation.md)** - Complete installation guide
- **[First Login](./10-getting-started/02-first-login.md)** - Initial setup and authentication
- **[Quick Start](./10-getting-started/02-quick-start.md)** - 5-minute getting started tutorial
- **[Core Concepts](./10-getting-started/03-core-concepts.md)** - Essential EntityDB concepts

#### [20-architecture](./20-architecture/)
System architecture and technical design
- **[System Overview](./20-architecture/01-system-overview.md)** - High-level architecture
- **[Temporal Architecture](./20-architecture/02-temporal-architecture.md)** - Time-series design
- **[RBAC Architecture](./20-architecture/03-rbac-architecture.md)** - Security model
- **[Entity Model](./20-architecture/04-entity-model.md)** - Data model design
- **[Authentication](./20-architecture/05-authentication-architecture.md)** - Auth system design
- **[Tag System](./20-architecture/07-tag-system.md)** - Tag-based data model
- **[Dataset Architecture](./20-architecture/08-dataset-architecture.md)** - Multi-tenancy
- **[Metrics Architecture](./20-architecture/09-metrics-architecture.md)** - Monitoring design

#### [30-api-reference](./30-api-reference/)
Complete API documentation and examples
- **[API Overview](./30-api-reference/01-overview.md)** - API introduction and basics
- **[Authentication](./30-api-reference/02-authentication.md)** - Auth endpoints and examples
- **[Entities](./30-api-reference/03-entities.md)** - Entity CRUD operations
- **[Queries](./30-api-reference/04-queries.md)** - Query and search endpoints
- **[Examples](./30-api-reference/05-examples.md)** - Comprehensive code examples

#### [40-user-guides](./40-user-guides/)
End-user guides and tutorials
- **[Temporal Queries](./40-user-guides/01-temporal-queries.md)** - Time-travel queries
- **[Dashboard Guide](./40-user-guides/02-dashboard-guide.md)** - Web UI usage
- **[Widgets](./40-user-guides/03-widgets.md)** - Dashboard customization
- **[Advanced Queries](./40-user-guides/04-advanced-queries.md)** - Complex query patterns

#### [50-admin-guides](./50-admin-guides/)
System administration and deployment
- **[Security Configuration](./50-admin-guides/01-security-configuration.md)** - Security hardening
- **[Deployment Guide](./50-admin-guides/02-deployment-guide.md)** - Production deployment
- **[Migration Guide](./50-admin-guides/04-migration-guide.md)** - Version upgrades
- **[RBAC Implementation](./50-admin-guides/05-rbac-implementation.md)** - Permission setup

#### [60-developer-guides](./60-developer-guides/)
Development and contribution guides
- **[Contributing](./60-developer-guides/01-contributing.md)** - How to contribute
- **[Git Workflow](./60-developer-guides/02-git-workflow.md)** - Development process
- **[Logging Standards](./60-developer-guides/03-logging-standards.md)** - Code standards
- **[Configuration](./60-developer-guides/04-configuration.md)** - Config management

#### [70-deployment](./70-deployment/)
Production deployment guides
- **[Production Checklist](./70-deployment/02-production-checklist.md)** - Deployment checklist
- **[SSL Setup](./70-deployment/03-ssl-setup.md)** - SSL/TLS configuration

#### [80-troubleshooting](./80-troubleshooting/)
Problem diagnosis and resolution
- **[Content Format](./80-troubleshooting/01-content-format.md)** - Content issues
- **[Content Wrapping](./80-troubleshooting/02-content-wrapping.md)** - Data formatting
- **[Tag Index](./80-troubleshooting/03-tag-index-persistence.md)** - Index problems
- **[Temporal Tags](./80-troubleshooting/04-temporal-tags.md)** - Time-series issues

#### [90-reference](./90-reference/)
Technical reference and specifications
- **[Configuration Reference](./90-reference/01-configuration-reference.md)** - All config options
- **[API Complete](./90-reference/02-api-complete.md)** - Full API specification
- **[Binary Format Spec](./90-reference/03-binary-format-spec.md)** - EBF technical spec
- **[RBAC Reference](./90-reference/04-rbac-reference.md)** - Permission reference

### Specialized Content

#### [archive](./archive/)
Historical documentation and deprecated content
- **[Archive Index](./archive/README.md)** - What's archived and why

#### [performance](./performance/)
Performance analysis and optimization guides
- **[Performance Index](./performance/performance-index.md)** - Performance documentation hub

#### [applications](./applications/)
Application-specific integrations
- **[Application Guides](./applications/)** - External application integration

## 🔗 External Resources

- **Repository**: https://git.home.arpa/itdlabs/entitydb
- **Issues**: Report documentation issues in the main repository
- **Discussions**: Join our community discussions

---

*This documentation is maintained by the EntityDB team and community. For corrections or improvements, please submit a pull request.*