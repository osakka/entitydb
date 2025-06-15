# EntityDB Documentation Taxonomy Design 2025

## Overview

This document defines the comprehensive taxonomy and organizational structure for EntityDB documentation, establishing clear categorization rules, naming conventions, and content standards.

## Core Design Principles

### 1. User-Centric Organization
- Structure by user persona and use case
- Progressive disclosure (basic → advanced)
- Task-oriented rather than feature-oriented

### 2. Single Source of Truth
- No duplicate content across directories
- Clear ownership of information types
- Canonical reference locations

### 3. Discoverability
- Consistent naming conventions
- Comprehensive indexing
- Logical hierarchy with numbered categories

### 4. Maintainability
- Maximum 50 files per directory
- Regular review and pruning processes
- Version-controlled taxonomy evolution

## Proposed Documentation Structure

```
docs/
├── README.md                           # Master index and navigation
├── CHANGELOG.md                        # Version history (moved from internals)
├── TAXONOMY.md                         # This taxonomy document
│
├── 01-overview/                        # High-level introduction
│   ├── README.md                       # Category index
│   ├── introduction.md                 # What is EntityDB
│   ├── features.md                     # Key capabilities
│   ├── system-requirements.md          # Hardware/software requirements
│   └── comparison.md                   # vs other databases
│
├── 02-getting-started/                 # New user onboarding
│   ├── README.md
│   ├── installation.md                 # Install and setup
│   ├── quick-start.md                  # 5-minute tutorial
│   ├── first-steps.md                  # Basic operations
│   └── common-patterns.md              # Typical use cases
│
├── 03-user-guide/                      # End user documentation
│   ├── README.md
│   ├── basic-operations.md             # CRUD operations
│   ├── temporal-queries.md             # Time-travel queries
│   ├── dashboard-guide.md              # Web UI usage
│   ├── advanced-queries.md             # Complex queries
│   └── best-practices.md               # Optimization tips
│
├── 04-api-reference/                   # Complete API documentation
│   ├── README.md
│   ├── authentication.md               # Auth endpoints
│   ├── entities.md                     # Entity CRUD
│   ├── queries.md                      # Query endpoints
│   ├── relationships.md                # Relationship management
│   ├── metrics.md                      # Monitoring endpoints
│   ├── examples.md                     # Code examples
│   └── openapi-spec.md                 # Full API specification
│
├── 05-administration/                  # System administration
│   ├── README.md
│   ├── deployment.md                   # Production deployment
│   ├── configuration.md                # All config options
│   ├── security.md                     # Security hardening
│   ├── monitoring.md                   # Health monitoring
│   ├── backup-restore.md               # Data management
│   ├── performance-tuning.md           # Optimization
│   └── maintenance.md                  # Routine tasks
│
├── 06-development/                     # Developer documentation
│   ├── README.md
│   ├── architecture.md                 # System design
│   ├── contributing.md                 # How to contribute
│   ├── binary-format.md                # EBF specification
│   ├── temporal-engine.md              # Time-series design
│   ├── performance.md                  # Performance architecture
│   ├── testing.md                      # Test strategy
│   └── git-workflow.md                 # Development process
│
├── 07-reference/                       # Technical reference
│   ├── README.md
│   ├── configuration-options.md        # All config parameters
│   ├── api-specification.md            # Complete API spec
│   ├── error-codes.md                  # Error reference
│   ├── binary-format-spec.md           # EBF technical spec
│   ├── rbac-reference.md               # Permission system
│   └── glossary.md                     # Term definitions
│
├── 08-troubleshooting/                 # Problem solving
│   ├── README.md
│   ├── common-issues.md                # FAQ and solutions
│   ├── debugging.md                    # Debug techniques
│   ├── performance-issues.md           # Performance problems
│   └── support.md                      # Getting help
│
├── 09-appendix/                        # Additional information
│   ├── README.md
│   ├── migration-guides.md             # Version migration
│   ├── release-notes.md                # Detailed release notes
│   ├── benchmarks.md                   # Performance data
│   ├── third-party.md                  # External integrations
│   └── legal.md                        # Licenses and compliance
│
└── archive/                            # Historical documents
    ├── README.md                       # Archive index
    ├── deprecated/                     # Deprecated features
    ├── historical/                     # Development history
    └── migration-logs/                 # Version migration logs
```

## Content Migration Plan

### Files to Archive (Move to archive/)
Most content from current `internals/` directory:
- All files in `internals/archive/` (already archived)
- Most files in `internals/implementation/` (historical implementation notes)
- Status reports and phase completion summaries
- Obsolete analysis and planning documents

### Files to Relocate
- `performance/` → `06-development/performance.md` and `09-appendix/benchmarks.md`
- `releases/` → `09-appendix/release-notes.md`
- `migration/` → `09-appendix/migration-guides.md`
- Current numbered directories → New taxonomy structure

### Files to Consolidate
- Multiple authentication guides → Single `04-api-reference/authentication.md`
- Various architecture documents → `06-development/architecture.md`
- Performance documents → `06-development/performance.md`

## Naming Conventions

### Directory Names
- Use numbered prefixes: `01-overview`, `02-getting-started`
- Use kebab-case: `getting-started`, `api-reference`
- Maximum 20 characters
- Descriptive and intuitive

### File Names
- Use kebab-case: `quick-start.md`, `binary-format.md`
- No version numbers in filenames
- Maximum 30 characters
- Descriptive action words: `installation.md`, `troubleshooting.md`

### Content Headers
- Consistent metadata headers:
  ```markdown
  # Document Title
  
  > **Version**: v2.31.0 | **Last Updated**: 2025-06-13 | **Status**: AUTHORITATIVE
  ```

## Content Standards

### File Size Limits
- Maximum 500 lines per file
- Split large files into logical sections
- Use cross-references for related content

### Content Types
- **AUTHORITATIVE**: Canonical, definitive documentation
- **GUIDANCE**: Best practices and recommendations  
- **REFERENCE**: Technical specifications and data
- **TUTORIAL**: Step-by-step instructions
- **ARCHIVED**: Historical or deprecated content

### Cross-Reference Format
- Use relative paths: `[link](../06-development/architecture.md)`
- Include section anchors: `[link](../04-api-reference/entities.md#creating-entities)`
- Maintain link inventory for validation

## Quality Assurance

### Review Process
1. **Technical Accuracy**: Verify against codebase
2. **Link Validation**: Test all internal/external links
3. **Content Freshness**: Update version references
4. **Accessibility**: Check for clear navigation
5. **Completeness**: Ensure comprehensive coverage

### Maintenance Schedule
- **Weekly**: Link validation and broken reference fixes
- **Monthly**: Content freshness review and version updates
- **Quarterly**: Taxonomy review and structure optimization
- **Per Release**: Major content updates and migration

## Implementation Phases

### Phase 1: Core Structure (Week 1)
- Create new directory structure
- Migrate critical user-facing documentation
- Update master README.md with new navigation

### Phase 2: Content Migration (Week 2)
- Move and consolidate existing content
- Archive obsolete documentation
- Update all cross-references

### Phase 3: Quality Assurance (Week 3)
- Validate all links and references
- Proofread and edit for consistency
- Test navigation paths

### Phase 4: Publication (Week 4)
- Final review and approval
- Update version references to v2.31.0
- Publish and announce new structure

## Success Metrics

### Quantitative
- Reduce total file count from 252 to <150
- Achieve <5% broken internal links
- Maximum 2-click navigation to any content
- 100% version consistency across all files

### Qualitative
- Clear user journey for each persona
- Improved discoverability of information
- Reduced documentation maintenance overhead
- Enhanced user satisfaction with documentation

---

*This taxonomy design will be implemented systematically to create a world-class documentation library for EntityDB.*