# EntityDB Documentation Library

> **Version**: v2.32.0-dev | **Last Updated**: 2025-06-14

Welcome to the EntityDB documentation library. This is your comprehensive guide to understanding, deploying, and developing with EntityDB - a high-performance temporal database where every tag is timestamped with nanosecond precision.

## 🆕 Latest Updates (v2.32.0-dev)

- **Clean Vanilla Dashboard**: Replaced complex Vue.js implementation with self-contained vanilla HTML/CSS/JavaScript dashboard
- **RBAC Permission Fix**: Corrected wildcard permission format for HasPermission compatibility (`rbac:perm:*`)
- **Zero Dependencies**: Dashboard now operates without external libraries for maximum browser compatibility
- **3-Core Focus Areas**: Performance monitoring, entity management, user/role administration

## 🚀 Recent Major Release (v2.31.0)

- **Comprehensive Performance Optimization Suite**: Enterprise-scale improvements delivering significant memory, CPU, and storage enhancements
- **O(1) Tag Value Caching**: Converted O(n) tag lookups to O(1) with intelligent lazy caching
- **Parallel Index Building**: 4-worker concurrent processing for faster server startup
- **JSON Encoder Pooling**: Reduced API allocation overhead with sync.Pool management
- **Batch Write Operations**: Configurable batching (10 entities, 100ms intervals) for improved throughput
- **Temporal Tag Variant Caching**: Pre-computed O(1) temporal tag lookups for optimized ListByTag operations

## ⚠️ Critical Notice for v2.29.0+

**BREAKING CHANGE**: The authentication architecture fundamentally changed in v2.29.0. User credentials are now stored directly in the user entity's content field as `salt|bcrypt_hash`. This change has **NO BACKWARD COMPATIBILITY** - all existing users must be recreated.

## 🎯 Quick Navigation

### 🚀 Getting Started
| Document | Description | Audience |
|----------|-------------|----------|
| [Introduction](./getting-started/01-introduction.md) | EntityDB overview and value proposition | Everyone |
| [Installation](./getting-started/01-installation.md) | Get EntityDB running in minutes | Everyone |
| [Quick Start](./getting-started/02-quick-start.md) | Your first EntityDB experience | New Users |
| [Core Concepts](./getting-started/03-core-concepts.md) | Understand entities, tags, temporal data | Everyone |

### 🏗️ Architecture & Design
| Document | Description | Audience |
|----------|-------------|----------|
| [System Overview](./architecture/01-system-overview.md) | High-level architecture and design principles | Technical |
| [Temporal Architecture](./architecture/02-temporal-architecture.md) | How time-series data is stored and queried | Technical |
| [RBAC Architecture](./architecture/03-rbac-architecture.md) | Role-based access control implementation | Admin/Dev |
| [Entity Model](./architecture/04-entity-model.md) | Core entity data model | Developers |

### 📚 User Guides
| Document | Description | Audience |
|----------|-------------|----------|
| [Temporal Queries](./user-guide/01-temporal-queries.md) | Time-travel and history queries | End Users |
| [Dashboard Guide](./user-guide/02-dashboard-guide.md) | Navigate the web interface | End Users |
| [Advanced Queries](./user-guide/04-advanced-queries.md) | Complex search and filtering | Power Users |

### 🔧 Administrator Guides
| Document | Description | Audience |
|----------|-------------|----------|
| [User Management](./admin-guide/01-user-management.md) | Create and manage users with RBAC | Administrators |
| [Security Configuration](./admin-guide/01-security-configuration.md) | SSL, authentication, authorization | Administrators |
| [Production Deployment](./admin-guide/01-production-deployment.md) | Deploy EntityDB in production | DevOps/Admin |
| [Monitoring Guide](./admin-guide/02-monitoring-guide.md) | Health checks, metrics, troubleshooting | Administrators |

### 🔌 API Reference
| Document | Description | Audience |
|----------|-------------|----------|
| [API Overview](./api-reference/01-overview.md) | REST API introduction and concepts | Developers |
| [Authentication](./api-reference/02-authentication.md) | Login, sessions, tokens | Developers |
| [Entities](./api-reference/03-entities.md) | CRUD operations and queries | Developers |
| [Queries](./api-reference/04-queries.md) | Advanced search and temporal queries | Developers |
| [Code Examples](./api-reference/05-examples.md) | Working examples in multiple languages | Developers |

### 👩‍💻 Developer Guide
| Document | Description | Audience |
|----------|-------------|----------|
| [Contributing](./developer-guide/01-contributing.md) | How to contribute to EntityDB | Contributors |
| [Git Workflow](./developer-guide/02-git-workflow.md) | Branching, commits, pull requests | Contributors |
| [Logging Standards](./developer-guide/03-logging-standards.md) | Logging conventions and best practices | Contributors |
| [Configuration](./developer-guide/04-configuration.md) | Configuration management patterns | Contributors |
| [Maintenance Guidelines](./developer-guide/maintenance-guidelines.md) | Project maintenance procedures | Contributors |

### 📖 Reference
| Document | Description | Audience |
|----------|-------------|----------|
| [Configuration Reference](./reference/01-configuration-reference.md) | All configuration options | Admin/Dev |
| [API Complete Reference](./reference/02-api-complete.md) | Comprehensive API documentation | Developers |
| [Binary Format Specification](./reference/03-binary-format-spec.md) | Technical storage format details | Advanced Devs |
| [RBAC Reference](./reference/04-rbac-reference.md) | Complete permission and role system | Admin |
| [Troubleshooting](./reference/troubleshooting/) | Common issues and solutions | Everyone |

## 🏛️ Documentation Structure

The documentation follows industry-standard patterns with clear separation of concerns:

```
docs/
├── README.md                    # This master index
├── CHANGELOG.md                 # Version history and changes
├── getting-started/             # New user onboarding
│   ├── 01-introduction.md       # EntityDB overview  
│   ├── 01-installation.md       # Installation guide
│   ├── 02-quick-start.md        # Quick start tutorial
│   └── 03-core-concepts.md      # Core concepts
├── architecture/                # System design and technical architecture
│   ├── 01-system-overview.md    # High-level architecture
│   ├── 02-temporal-architecture.md # Temporal storage design
│   ├── 03-rbac-architecture.md  # Security and permissions
│   └── 04-entity-model.md       # Data model specification
├── user-guide/                  # End-user documentation
│   ├── 01-temporal-queries.md   # Time-travel queries
│   ├── 02-dashboard-guide.md    # Web interface guide
│   └── 04-advanced-queries.md   # Advanced search features
├── admin-guide/                 # Administrative documentation
│   ├── 01-user-management.md    # User and RBAC management
│   ├── 01-security-configuration.md # Security setup
│   ├── 01-production-deployment.md # Production deployment
│   └── 02-monitoring-guide.md   # Monitoring and maintenance
├── api-reference/               # Developer API documentation
│   ├── 01-overview.md           # API introduction
│   ├── 02-authentication.md     # Auth endpoints
│   ├── 03-entities.md           # Entity CRUD operations
│   ├── 04-queries.md            # Advanced query endpoints
│   └── 05-examples.md           # Code examples
├── developer-guide/             # Contributor and development documentation
│   ├── 01-contributing.md       # Contribution guidelines
│   ├── 02-git-workflow.md       # Git procedures
│   ├── 03-logging-standards.md  # Logging conventions
│   ├── 04-configuration.md      # Configuration patterns
│   └── maintenance-guidelines.md # Maintenance procedures
├── reference/                   # Technical specifications and troubleshooting
│   ├── 01-configuration-reference.md # Complete config reference
│   ├── 02-api-complete.md       # Complete API specification
│   ├── 03-binary-format-spec.md # Binary format technical details
│   ├── 04-rbac-reference.md     # RBAC system reference
│   └── troubleshooting/         # Troubleshooting guides
└── archive/                     # Historical and deprecated content
```

## 📊 Documentation Quality Standards

### ✅ Content Standards
- **Technical Accuracy**: All content verified against v2.32.0 codebase
- **Code Examples**: Working, tested examples that execute correctly
- **Version Compatibility**: Clear marking of version-specific features
- **Cross-References**: Comprehensive linking between related topics

### 📝 Format Standards
- **Markdown**: All documentation in GitHub-flavored Markdown
- **Structure**: Consistent headers, table of contents, code blocks
- **Navigation**: Clear breadcrumbs and section organization
- **Accessibility**: Descriptive links and alt text for images

### 🔄 Maintenance Process
- **Quarterly Reviews**: Technical accuracy against latest codebase
- **Change Triggers**: Documentation updates required for all PR merges
- **Quality Gates**: Technical review required for architectural changes
- **User Feedback**: Regular incorporation of user-reported issues

## 🚨 Getting Help

### 📞 Quick Support
- **Issues**: [Report bugs and request features](https://git.home.arpa/itdlabs/entitydb/issues)
- **Discussions**: [Community support and questions](https://git.home.arpa/itdlabs/entitydb/discussions)
- **Documentation Issues**: [Report documentation problems](https://git.home.arpa/itdlabs/entitydb/issues?labels=documentation)

### 🤝 Contributing
- Read the [Contributing Guide](./developer-guide/01-contributing.md)
- Follow the [Git Workflow](./developer-guide/02-git-workflow.md)
- Review [Configuration Standards](./developer-guide/05-configuration-alignment-action-plan.md)

### 🔧 Documentation Maintenance
- [Maintenance Guidelines](./DOCUMENTATION_MAINTENANCE.md) - Complete maintenance standards and processes
- [Quick Maintenance Checklist](./QUICK_MAINTENANCE_CHECKLIST.md) - Fast reference for common tasks

### 🔗 External Resources
- **Source Code**: [EntityDB Repository](https://git.home.arpa/itdlabs/entitydb)
- **API Documentation**: Live API docs at `https://localhost:8085/swagger/`
- **Dashboard**: Web interface at `https://localhost:8085/`

---

**📋 About This Documentation**
- **Maintained By**: EntityDB Documentation Team
- **Version**: Aligned with EntityDB v2.32.0-dev
- **Last Major Update**: 2025-06-14
- **Next Scheduled Review**: Q1 2025

*This documentation library is actively maintained and follows industry best practices for technical documentation. All content is verified for accuracy and completeness.*