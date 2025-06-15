# EntityDB Documentation Taxonomy & Standards

> **Documentation Architecture for EntityDB v2.32.0**  
> Professional, industry-standard documentation structure designed for accuracy, maintainability, and user experience.

## Documentation Principles

1. **Single Source of Truth**: Each concept documented once, in its logical location
2. **Accuracy First**: Documentation reflects actual codebase implementation
3. **User-Centered**: Organized by user journey and functional needs
4. **Maintainable**: Clear ownership, review processes, and update procedures
5. **Discoverable**: Logical hierarchy with cross-references and search optimization

## Taxonomy Structure

```
docs/
├── README.md                          # Master documentation index
├── CHANGELOG.md                       # Version history and breaking changes
├── CONTRIBUTING.md                    # Contribution guidelines (links to developer-guide/)
│
├── getting-started/                   # New user onboarding
│   ├── README.md                      # Getting started index
│   ├── 01-introduction.md             # What is EntityDB, core concepts
│   ├── 02-installation.md             # Installation and setup
│   ├── 03-quick-start.md              # 5-minute tutorial
│   └── 04-first-application.md        # Building your first app
│
├── user-guide/                        # End-user documentation
│   ├── README.md                      # User guide index
│   ├── 01-core-concepts.md            # Entities, tags, temporal data
│   ├── 02-querying-data.md            # Basic queries and filters
│   ├── 03-temporal-queries.md         # Time travel, history, diffs
│   ├── 04-dashboard-guide.md          # Web dashboard usage
│   ├── 05-data-management.md          # Creating, updating, deleting
│   └── 06-troubleshooting.md          # Common issues and solutions
│
├── api-reference/                     # Complete API documentation
│   ├── README.md                      # API overview and authentication
│   ├── 01-authentication.md           # Login, tokens, sessions
│   ├── 02-entities.md                 # Entity CRUD operations
│   ├── 03-temporal-operations.md      # as-of, history, changes, diff
│   ├── 04-datasets.md                 # Dataset management
│   ├── 05-metrics.md                  # Metrics and monitoring APIs
│   ├── 06-administration.md           # Admin-only endpoints
│   ├── 07-error-responses.md          # Error codes and formats
│   └── openapi.yaml                   # OpenAPI/Swagger specification
│
├── architecture/                      # System design and internals
│   ├── README.md                      # Architecture overview
│   ├── 01-system-overview.md          # High-level architecture
│   ├── 02-storage-layer.md            # Binary format, WAL, indexing
│   ├── 03-temporal-architecture.md    # Temporal data and indexing
│   ├── 04-security-model.md           # RBAC, authentication, authorization
│   ├── 05-performance-design.md       # Concurrency, caching, optimization
│   └── 06-data-model.md               # Entity model, tag system
│
├── developer-guide/                   # Development and contribution
│   ├── README.md                      # Developer guide index
│   ├── 01-development-setup.md        # Local development environment
│   ├── 02-contributing.md             # Contribution guidelines
│   ├── 03-code-standards.md           # Coding standards and practices
│   ├── 04-testing.md                  # Testing guidelines and tools
│   ├── 05-release-process.md          # Version management and releases
│   └── 06-debugging.md                # Debugging tools and techniques
│
├── admin-guide/                       # System administration
│   ├── README.md                      # Admin guide index
│   ├── 01-installation.md             # Production installation
│   ├── 02-configuration.md            # Configuration management
│   ├── 03-security.md                 # Security hardening
│   ├── 04-monitoring.md               # Monitoring and alerting
│   ├── 05-backup-recovery.md          # Backup and disaster recovery
│   ├── 06-performance-tuning.md       # Performance optimization
│   ├── 07-troubleshooting.md          # System troubleshooting
│   └── 08-upgrades.md                 # Version upgrades and migration
│
├── reference/                         # Technical reference materials
│   ├── README.md                      # Reference index
│   ├── 01-configuration.md            # Complete configuration reference
│   ├── 02-binary-format.md            # EntityDB Binary Format (EBF) spec
│   ├── 03-tag-namespaces.md           # Tag namespace conventions
│   ├── 04-rbac-permissions.md         # RBAC permission reference
│   ├── 05-environment-variables.md    # Environment variable reference
│   ├── 06-command-line.md             # Command-line interface
│   └── 07-glossary.md                 # Terms and definitions
│
├── examples/                          # Code examples and tutorials
│   ├── README.md                      # Examples index
│   ├── 01-basic-crud.md               # Basic create, read, update operations
│   ├── 02-temporal-queries.md         # Temporal query examples
│   ├── 03-application-integration.md  # Integrating with applications
│   ├── 04-metrics-collection.md       # Custom metrics collection
│   ├── 05-bulk-operations.md          # Bulk data operations
│   └── sample-applications/           # Complete sample applications
│       ├── task-manager/
│       ├── audit-log/
│       └── metrics-dashboard/
│
└── archive/                           # Historical documentation
    ├── README.md                      # Archive index and search
    ├── migration-guides/              # Version migration guides
    ├── deprecated-features/           # Deprecated feature documentation
    └── historical/                    # Historical development notes
```

## Naming Conventions

### File Naming
- **Lowercase with hyphens**: `quick-start.md`, `api-reference.md`
- **Numbered sections**: `01-introduction.md`, `02-installation.md`
- **Clear, descriptive names**: `temporal-queries.md` not `temp.md`
- **Consistent extensions**: `.md` for Markdown, `.yaml` for YAML

### Directory Structure
- **Functional grouping**: Group by user type and purpose
- **Logical hierarchy**: 2-3 levels maximum for clarity
- **Consistent naming**: All directories lowercase with hyphens

### Content Organization
- **README.md in each directory**: Directory overview and navigation
- **Logical numbering**: 01, 02, 03 for sequential content
- **Cross-references**: Links between related concepts
- **Index pages**: Master navigation at each level

## Content Standards

### Document Structure
```markdown
# Title (H1 - only one per document)

> Brief description of the document's purpose and scope

## Overview (H2)
Brief introduction and what the reader will learn

## Section 1 (H2)
### Subsection (H3)
#### Details (H4 - sparingly used)

## See Also
- [Related Document](../path/to/doc.md)
- [External Resource](https://example.com)

## Last Updated
- Version: 2.32.0
- Date: 2025-06-15
- Author: [Name]
```

### Writing Guidelines
- **Active voice**: "Create an entity" not "An entity can be created"
- **Clear headings**: Descriptive headings that explain content
- **Code examples**: Working examples that users can copy-paste
- **Consistent terminology**: Use glossary terms consistently
- **User perspective**: Write from the user's point of view

### Technical Standards
- **Accurate examples**: All code examples tested and working
- **Version references**: Specify version compatibility
- **Error handling**: Include error cases and troubleshooting
- **Security notes**: Highlight security considerations
- **Performance tips**: Include performance implications

## Review and Maintenance

### Documentation Ownership
- **Getting Started**: Product team
- **User Guide**: Product team + Support
- **API Reference**: Engineering team (auto-generated)
- **Architecture**: Senior engineering team
- **Developer Guide**: Engineering team
- **Admin Guide**: DevOps team + Engineering

### Review Process
1. **Author Review**: Initial accuracy check
2. **Technical Review**: Engineering team validation
3. **Editorial Review**: Style and clarity check
4. **User Testing**: Validation with real users
5. **Final Approval**: Team lead sign-off

### Maintenance Schedule
- **Weekly**: API documentation regeneration
- **Monthly**: Link validation and accuracy check
- **Quarterly**: Comprehensive review and updates
- **Per Release**: Version-specific documentation updates

## Quality Metrics

### Accuracy Metrics
- Code examples execute successfully
- API documentation matches implementation
- Configuration examples are valid
- Links resolve correctly

### Usability Metrics
- Time to complete getting started guide
- User feedback scores
- Support ticket reduction
- Documentation search success rate

### Completeness Metrics
- API coverage percentage
- Feature documentation coverage
- Example availability
- Cross-reference completeness

## Migration Plan

### Phase 1: Foundation (Week 1)
- Create new directory structure
- Write master README.md
- Create CHANGELOG.md
- Set up quality checks

### Phase 2: Core Content (Week 2-3)
- Migrate and update getting-started/
- Migrate and update user-guide/
- Update api-reference/ for accuracy
- Update architecture/ for v2.32.0

### Phase 3: Specialized Content (Week 4)
- Update developer-guide/
- Update admin-guide/
- Create comprehensive reference/
- Migrate examples/

### Phase 4: Archive and Polish (Week 5)
- Move outdated content to archive/
- Validate all cross-references
- Implement quality checks
- Final review and approval

This taxonomy ensures EntityDB documentation meets professional standards while remaining maintainable and user-focused.