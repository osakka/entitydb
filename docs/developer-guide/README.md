# Developer Guide
> **Version**: v2.32.5 | **Last Updated**: 2025-06-18 | **Status**: AUTHORITATIVE

Welcome to the EntityDB Developer Guide. This section provides comprehensive development and contribution guidelines for building with and contributing to EntityDB.

## Quick Navigation

### Development Setup
- **[Contributing Guidelines](./01-contributing.md)** - Setup, prerequisites, and contribution process
- **[Git Workflow](./02-git-workflow.md)** - Git practices, branching, and state tracking
- **[Logging Standards](./03-logging-standards.md)** - Logging conventions and best practices
- **[Configuration Management](./04-configuration.md)** - Configuration system and environment setup

### Development Resources
- **[Documentation Architecture](./09-documentation-architecture.md)** - Documentation structure and standards
- **[Maintenance Guidelines](./maintenance-guidelines.md)** - Project maintenance procedures
- **[Quick Maintenance Checklist](./quick-maintenance-checklist.md)** - Common maintenance tasks

## Developer Journey

### New Contributors
1. **Setup**: Start with [Contributing Guidelines](./01-contributing.md)
2. **Standards**: Learn [Logging Standards](./03-logging-standards.md) and [Configuration Management](./04-configuration.md)
3. **Workflow**: Follow [Git Workflow](./02-git-workflow.md)

### Active Developers
- **Daily Development**: [Git Workflow](./02-git-workflow.md) and [Configuration Management](./04-configuration.md)
- **Documentation**: [Documentation Architecture](./09-documentation-architecture.md)
- **Maintenance**: [Maintenance Guidelines](./maintenance-guidelines.md)

## Development Environment

### Prerequisites
- Go 1.21+ for server development
- Node.js 18+ for frontend development
- Git with proper configuration
- Development tools (see setup guide)

### Architecture Overview
```
EntityDB Development Stack:
├── Go Server (src/)
├── JavaScript Frontend (share/htdocs/)
├── Documentation (docs/)
├── Testing Framework (tests/)
└── Development Tools (tools/)
```

### Key Development Areas

#### Backend Development
- **Core Engine**: Temporal database implementation
- **API Layer**: REST endpoints and business logic
- **Storage**: Binary format and WAL implementation
- **Performance**: Optimization and caching

#### Frontend Development
- **Dashboard**: Web interface using Alpine.js
- **Worca Platform**: Workforce management application
- **API Integration**: Client-side data management

#### Documentation
- **Technical Writing**: Following IEEE standards
- **API Documentation**: Swagger/OpenAPI specifications
- **User Guides**: Task-oriented documentation

## Code Quality Standards

### Technical Requirements
- **Test Coverage**: >80% for new code
- **Performance**: Sub-millisecond query targets
- **Security**: RBAC enforcement throughout
- **Documentation**: All public APIs documented

### Review Process
1. **Code Review**: Peer review for all changes
2. **Testing**: Automated test suite passage
3. **Documentation**: Updated docs for feature changes
4. **Performance**: Benchmarking for critical paths

## Development Tools

### Built-in Tools
- **Make Targets**: `make dev`, `make test`, `make lint`
- **Code Generation**: Swagger docs, mock generation
- **Testing**: Unit, integration, and performance tests
- **Debugging**: Comprehensive logging and tracing

### External Tools
- **IDE Integration**: VS Code configuration provided
- **Git Hooks**: Pre-commit quality checks
- **CI/CD**: Automated testing and deployment
- **Monitoring**: Development metrics and profiling

## Common Development Tasks

### Adding Features
1. Design documentation (ADR if architectural)
2. Implement with tests
3. Update API documentation
4. Performance validation
5. User guide updates

### Bug Fixes
1. Reproduce with test case
2. Root cause analysis
3. Minimal fix implementation
4. Regression test addition
5. Documentation updates

### Performance Optimization
1. Benchmark current performance
2. Profile bottlenecks
3. Implement optimizations
4. Validate improvements
5. Document changes

## Support and Resources

- **Technical Questions**: Use GitHub discussions
- **Bug Reports**: GitHub issues with template
- **Feature Requests**: RFC process via issues
- **Documentation Issues**: Direct PR submissions

### Internal Resources
- **Architecture**: [Architecture Guide](../architecture/)
- **API Reference**: [API Documentation](../api-reference/)
- **Admin Operations**: [Admin Guide](../admin-guide/)

---

*This developer guide covers EntityDB v2.32.5 development. For contribution guidelines, see [Contributing Guidelines](./01-contributing.md).*