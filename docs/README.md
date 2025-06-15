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
| [Installation](./getting-started/installation.md) | Get EntityDB running in minutes | Everyone |
| [Quick Start](./getting-started/quick-start.md) | Your first EntityDB experience | New Users |
| [Core Concepts](./getting-started/core-concepts.md) | Understand entities, tags, temporal data | Everyone |
| [Dashboard Guide](./user-guide/dashboard.md) | Navigate the web interface | End Users |

### 🏗️ Architecture & Design
| Document | Description | Audience |
|----------|-------------|----------|
| [System Overview](./architecture/system-overview.md) | High-level architecture and design principles | Technical |
| [Temporal Storage](./architecture/temporal-storage.md) | How time-series data is stored and queried | Technical |
| [RBAC System](./architecture/rbac-system.md) | Role-based access control implementation | Admin/Dev |
| [Binary Format](./architecture/binary-format.md) | EBF storage format specification | Developers |

### 📚 User Guides
| Document | Description | Audience |
|----------|-------------|----------|
| [Entity Management](./user-guide/entities.md) | Creating, updating, querying entities | End Users |
| [Temporal Queries](./user-guide/temporal-queries.md) | Time-travel and history queries | End Users |
| [Search & Filtering](./user-guide/search.md) | Finding data efficiently | End Users |

### 🔧 Administrator Guides
| Document | Description | Audience |
|----------|-------------|----------|
| [Production Deployment](./admin-guide/deployment.md) | Deploy EntityDB in production | DevOps/Admin |
| [User Management](./admin-guide/user-management.md) | Create and manage users with RBAC | Administrators |
| [Security Configuration](./admin-guide/security.md) | SSL, authentication, authorization | Administrators |
| [Monitoring & Maintenance](./admin-guide/monitoring.md) | Health checks, metrics, troubleshooting | Administrators |

### 🔌 API Reference
| Document | Description | Audience |
|----------|-------------|----------|
| [API Overview](./api-reference/overview.md) | REST API introduction and concepts | Developers |
| [Authentication](./api-reference/authentication.md) | Login, sessions, tokens | Developers |
| [Entities](./api-reference/entities.md) | CRUD operations and queries | Developers |
| [Temporal Operations](./api-reference/temporal.md) | Time-travel and history APIs | Developers |
| [Code Examples](./api-reference/examples.md) | Working examples in multiple languages | Developers |

### 👩‍💻 Developer Guide
| Document | Description | Audience |
|----------|-------------|----------|
| [Contributing](./developer-guide/contributing.md) | How to contribute to EntityDB | Contributors |
| [Development Setup](./developer-guide/development-setup.md) | Local development environment | Contributors |
| [Git Workflow](./developer-guide/git-workflow.md) | Branching, commits, pull requests | Contributors |
| [Testing Guide](./developer-guide/testing.md) | Running and writing tests | Contributors |
| [Code Standards](./developer-guide/code-standards.md) | Coding conventions and best practices | Contributors |

### 📖 Reference
| Document | Description | Audience |
|----------|-------------|----------|
| [Configuration Reference](./reference/configuration.md) | All configuration options | Admin/Dev |
| [API Complete Reference](./reference/api-complete.md) | Comprehensive API documentation | Developers |
| [Binary Format Specification](./reference/binary-format.md) | Technical storage format details | Advanced Devs |
| [RBAC Reference](./reference/rbac-reference.md) | Complete permission and role system | Admin |
| [Troubleshooting](./reference/troubleshooting.md) | Common issues and solutions | Everyone |

## 🏛️ Current Documentation Structure

The documentation follows industry-standard patterns with clear separation of concerns:

```
docs/
├── README.md                    # This master index
├── CHANGELOG.md                 # Version history and changes
├── getting-started/             # New user onboarding
├── architecture/                # System design and technical architecture
├── user-guide/                  # End-user documentation
├── admin-guide/                 # Administrative documentation
├── api-reference/               # Developer API documentation
├── developer-guide/             # Contributor and development documentation
├── reference/                   # Technical specifications and troubleshooting
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
- Read the [Contributing Guide](./developer-guide/contributing.md)
- Follow the [Git Workflow](./developer-guide/git-workflow.md)
- Review [Code Standards](./developer-guide/code-standards.md)

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