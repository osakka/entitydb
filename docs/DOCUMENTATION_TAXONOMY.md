# EntityDB Documentation Taxonomy
> Professional Documentation Architecture following IEEE 1063-2001 Standards

## Overview

This document defines the official documentation taxonomy for EntityDB, ensuring consistency, discoverability, and maintainability across all documentation assets.

## Core Principles

1. **Single Source of Truth**: Each piece of information has exactly one authoritative location
2. **User Journey Alignment**: Documentation structure follows user needs and experience levels
3. **Technical Accuracy**: All documentation reflects actual codebase implementation
4. **Professional Standards**: Follows IEEE 1063-2001 and industry best practices
5. **Maintainability**: Clear ownership and update processes for each document category

## Directory Structure

```
docs/
├── README.md                           # Navigation hub and documentation overview
├── getting-started/                    # New user onboarding (5-10 min to success)
│   ├── README.md
│   ├── 01-introduction.md
│   ├── 02-installation.md
│   ├── 03-quick-start.md
│   └── 04-first-steps.md
├── user-guide/                        # Task-oriented guides for end users
│   ├── README.md
│   ├── 01-basic-operations.md
│   ├── 02-temporal-queries.md
│   ├── 03-dashboard-usage.md
│   ├── 04-data-management.md
│   └── 05-troubleshooting.md
├── api-reference/                      # Complete API documentation
│   ├── README.md
│   ├── 01-overview.md
│   ├── 02-authentication.md
│   ├── 03-entities.md
│   ├── 04-temporal-operations.md
│   ├── 05-datasets.md
│   ├── 06-metrics.md
│   └── 07-examples.md
├── architecture/                       # System design and technical architecture
│   ├── README.md
│   ├── 01-system-overview.md
│   ├── 02-temporal-architecture.md
│   ├── 03-rbac-security.md
│   ├── 04-entity-model.md
│   ├── 05-storage-format.md
│   └── 06-performance-design.md
├── admin-guide/                       # System administration and operations
│   ├── README.md
│   ├── 01-installation-deployment.md
│   ├── 02-configuration.md
│   ├── 03-security-setup.md
│   ├── 04-monitoring.md
│   ├── 05-backup-recovery.md
│   ├── 06-troubleshooting.md
│   └── 07-maintenance.md
├── developer-guide/                   # Development and contribution guides
│   ├── README.md
│   ├── 01-development-setup.md
│   ├── 02-code-standards.md
│   ├── 03-testing.md
│   ├── 04-debugging.md
│   ├── 05-performance-optimization.md
│   └── 06-contribution-workflow.md
├── examples/                          # Code examples and tutorials
│   ├── README.md
│   ├── basic-usage/
│   ├── temporal-queries/
│   ├── integrations/
│   └── performance/
├── reference/                         # Technical specifications and references
│   ├── README.md
│   ├── 01-configuration-reference.md
│   ├── 02-binary-format-spec.md
│   ├── 03-rbac-reference.md
│   ├── 04-error-codes.md
│   └── 05-glossary.md
├── adr/                              # Architectural Decision Records
│   ├── README.md
│   ├── template.md
│   └── [numbered ADRs]
└── archive/                          # Historical and obsolete documentation
    ├── README.md
    ├── migrations/
    ├── legacy-versions/
    └── deprecated-features/
```

## Document Naming Conventions

### File Naming Standards
- **README.md**: Each directory MUST have a README.md serving as navigation and overview
- **Numbered Prefixes**: Use 2-digit prefixes (01-, 02-) for sequential content
- **Descriptive Names**: Clear, specific names using hyphens for word separation
- **Consistent Terminology**: Use standardized terms across all documentation

### Section Organization
1. **Getting Started**: 01-04 (core onboarding journey)
2. **User Guide**: 01-05 (task-oriented guides)
3. **API Reference**: 01-07 (complete API coverage)
4. **Architecture**: 01-06 (technical design)
5. **Admin Guide**: 01-07 (operations and maintenance)
6. **Developer Guide**: 01-06 (development workflow)

## Content Categories and Ownership

### Primary Categories

| Category | Purpose | Audience | Update Frequency |
|----------|---------|----------|------------------|
| **Getting Started** | User onboarding | New users | Every release |
| **User Guide** | Task completion | End users | Every minor release |
| **API Reference** | Complete API docs | Developers | Every API change |
| **Architecture** | System design | Technical leads | Major releases |
| **Admin Guide** | Operations | System admins | Every release |
| **Developer Guide** | Contribution | Contributors | Every release |
| **Examples** | Working code | All users | Every release |
| **Reference** | Specifications | All users | As needed |
| **ADR** | Design decisions | Technical team | Per decision |
| **Archive** | Historical docs | Maintainers | Quarterly cleanup |

### Content Lifecycle

1. **Creation**: Follow template and review process
2. **Active**: Regular updates aligned with release cycle
3. **Maintenance**: Quarterly accuracy reviews
4. **Archive**: Move obsolete content to archive/ with clear metadata

## Quality Standards

### Technical Accuracy Requirements
- All code examples MUST be tested and working
- Version numbers MUST be consistent across all documents
- API documentation MUST reflect actual implementation
- Screenshots and examples MUST be current
- Cross-references MUST be valid and maintained

### Writing Standards
- **Clarity**: Write for your audience's expertise level
- **Conciseness**: Eliminate redundancy and wordiness
- **Actionable**: Provide clear steps and outcomes
- **Scannable**: Use headings, lists, and formatting effectively
- **Current**: Maintain version accuracy and relevance

### Review Process
1. **Technical Review**: Accuracy validation by code owners
2. **Editorial Review**: Writing quality and consistency
3. **User Testing**: Validate against actual user workflows
4. **Maintenance**: Regular updates and accuracy checks

## Version Control and Releases

### Version Synchronization
- Documentation version MUST match codebase version
- Release notes MUST be comprehensive and accurate
- Breaking changes MUST be clearly documented
- Migration guides MUST be provided for major changes

### Release Documentation Requirements
1. **CHANGELOG.md**: User-facing changes with clear categorization
2. **Migration guides**: Detailed upgrade instructions
3. **API changes**: Complete changelog with examples
4. **Feature documentation**: Complete coverage of new capabilities

## Cross-Reference System

### Internal Links
- Use relative paths for internal documentation links
- Maintain link accuracy through automated checking
- Provide breadcrumb navigation for deep documents
- Create comprehensive cross-reference index

### External References
- Minimize external dependencies
- Use stable, authoritative sources
- Provide fallback information for critical external links
- Regular validation of external references

## Templates and Standards

### Document Templates
Each category has standardized templates ensuring:
- Consistent structure and formatting
- Required metadata and front matter
- Standard sections and organization
- Quality checklist compliance

### Metadata Requirements
```yaml
---
title: Document Title
category: getting-started|user-guide|api-reference|architecture|admin-guide|developer-guide|examples|reference|adr
audience: new-users|end-users|developers|admins|contributors
last-updated: YYYY-MM-DD
version: vX.Y.Z
review-cycle: release|quarterly|as-needed
---
```

## Maintenance and Governance

### Regular Maintenance Tasks
- **Weekly**: Link validation and basic accuracy checks
- **Monthly**: User feedback review and incorporation
- **Quarterly**: Comprehensive accuracy audit
- **Per Release**: Full documentation update and review

### Ownership Model
- **Documentation Lead**: Overall taxonomy and quality oversight
- **Technical Writers**: Content creation and maintenance
- **Code Owners**: Technical accuracy validation
- **Community**: Feedback and contribution coordination

## Success Metrics

### Quantitative Metrics
- Documentation coverage (% of features documented)
- Accuracy score (% of validated information)
- User task completion rate
- Time-to-first-success for new users
- Support ticket reduction correlation

### Qualitative Metrics
- User feedback and satisfaction scores
- Internal team usage and adoption
- Community contribution quality
- Search and discovery effectiveness
- Cross-reference link health

---

This taxonomy serves as the foundational framework for all EntityDB documentation. It ensures professional standards, user success, and maintainable excellence across all documentation assets.