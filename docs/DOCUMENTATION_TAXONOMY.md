# EntityDB Documentation Taxonomy and Standards

## Documentation Taxonomy

This document defines the professional taxonomy and standards for EntityDB documentation, ensuring consistency, accuracy, and adherence to industry best practices.

## Core Principles

1. **Single Source of Truth**: Each concept documented in exactly one authoritative location
2. **Progressive Disclosure**: Information organized by user needs and expertise level
3. **Factual Accuracy**: All content verified against actual codebase implementation
4. **Professional Standards**: Industry-standard organization and naming conventions
5. **Navigation Clarity**: Logical structure with clear cross-references

## Directory Structure Standards

### Primary Categories

```
/docs/
├── README.md                    # Master documentation index
├── getting-started/             # New user journey (5-10 files max)
├── user-guide/                  # Task-oriented user documentation
├── api-reference/               # Complete API documentation
├── architecture/                # Technical architecture decisions
│   ├── README.md               # Architecture navigation guide
│   ├── adr/                    # Architectural Decision Records
│   ├── decisions/              # Numbered architecture documents
│   └── system-overview/        # High-level architecture overviews
├── developer-guide/            # Development workflow and standards
├── admin-guide/                # System administration and operations
├── reference/                  # Technical reference materials
├── testing/                    # Testing documentation and procedures
├── archive/                    # Historical documentation (read-only)
└── assets/                     # Images, diagrams, and resources
```

### Category Definitions

#### `/getting-started/` (Onboarding)
- **Purpose**: New user onboarding journey
- **Audience**: First-time users and evaluators
- **Content**: Quick start guides, basic concepts, initial setup
- **File limit**: 5-10 files maximum to avoid overwhelm
- **Naming**: `01-`, `02-`, etc. for sequential journey

#### `/user-guide/` (Task-Oriented)
- **Purpose**: Task-oriented user documentation
- **Audience**: Regular users performing specific tasks
- **Content**: How-to guides, workflows, feature usage
- **Organization**: By functional area or user goal
- **Naming**: Descriptive task names with numbered sequences

#### `/api-reference/` (Technical Reference)
- **Purpose**: Complete API documentation
- **Audience**: Developers integrating with EntityDB
- **Content**: Endpoints, schemas, examples, authentication
- **Organization**: By API category or version
- **Naming**: Logical API groupings

#### `/architecture/` (Technical Decisions)
- **Purpose**: Technical architecture and decisions
- **Audience**: Technical architects and senior developers
- **Subdirectories**:
  - `adr/`: Formal Architectural Decision Records (ADR-XXX.md format)
  - `decisions/`: Numbered architecture documents (001-XXX.md format)
  - `system-overview/`: High-level system architecture
- **Content**: Technical decisions, system design, architectural patterns

#### `/developer-guide/` (Development)
- **Purpose**: Development workflow and standards
- **Audience**: EntityDB contributors and developers
- **Content**: Coding standards, contribution guidelines, development setup
- **Organization**: By development lifecycle phase

#### `/admin-guide/` (Operations)
- **Purpose**: System administration and operations
- **Audience**: System administrators and DevOps engineers
- **Content**: Installation, configuration, monitoring, troubleshooting
- **Organization**: By operational task or system component

#### `/reference/` (Technical Reference)
- **Purpose**: Comprehensive technical reference
- **Audience**: Technical users needing detailed specifications
- **Content**: Configuration reference, technical specifications, troubleshooting
- **Organization**: By technical domain

#### `/testing/` (Quality Assurance)
- **Purpose**: Testing documentation and procedures
- **Audience**: Developers and QA engineers
- **Content**: Test plans, testing procedures, validation guides

#### `/archive/` (Historical)
- **Purpose**: Historical documentation preservation
- **Audience**: Historical reference only
- **Content**: Superseded documentation, migration records
- **Organization**: By date or version

## Naming Conventions

### File Naming Standards

1. **Sequential Documents**: Use zero-padded numbers
   - `01-introduction.md`, `02-installation.md`, etc.
   - Maximum 2 digits unless more than 99 files (unlikely)

2. **Descriptive Names**: Use kebab-case for clarity
   - `user-management.md`, `ssl-configuration.md`
   - No spaces, underscores, or special characters

3. **ADR Files**: Follow ADR standard format
   - `ADR-001-unified-file-format.md`
   - Three-digit zero-padded numbers

4. **Architecture Decisions**: Numbered format
   - `001-temporal-tag-storage.md`
   - Three-digit zero-padded numbers

### Directory Naming Standards

1. **Kebab-case**: All lowercase with hyphens
   - `getting-started`, `user-guide`, `api-reference`
   - No underscores or camelCase

2. **Functional Names**: Describe purpose, not content type
   - ✅ `user-guide` (describes purpose)
   - ❌ `markdown-files` (describes format)

3. **Standard Terms**: Use industry-standard terminology
   - `api-reference` not `api-docs`
   - `developer-guide` not `dev-docs`

## Content Standards

### Documentation Quality Requirements

1. **Factual Accuracy**: All technical details verified against codebase
2. **No Exaggerations**: Objective descriptions without marketing language
3. **Clear and Crisp**: Concise writing with specific technical details
4. **Consistent Terminology**: Standardized terms throughout documentation
5. **Single Source of Truth**: Each concept documented once authoritatively

### Cross-Reference Standards

1. **Relative Links**: Use relative paths for internal links
   - `../user-guide/authentication.md`
   - Not absolute URLs or broken references

2. **Link Validation**: All internal links must be valid
   - Regular verification against actual file structure
   - Broken link detection in build process

3. **Clear Context**: Links provide context about destination
   - "See the [SSL Configuration Guide](../admin-guide/ssl-setup.md)"
   - Not just "click here" or "see this"

## README File Standards

### Master README (`/docs/README.md`)

1. **Purpose**: Primary navigation hub for all documentation
2. **Content Structure**:
   - Welcome and overview
   - Quick navigation by user type
   - Category descriptions with file counts
   - Link to this taxonomy document

3. **Maintenance**: Updated when directories change

### Category READMEs

1. **Purpose**: Navigation within specific categories
2. **Content Structure**:
   - Category purpose and audience
   - File inventory with descriptions
   - Recommended reading order
   - Cross-references to related categories

3. **Consistency**: Standard format across all categories

## Content Duplication Policy

### Prohibited Duplications

1. **Installation procedures**: Single authoritative guide per use case
2. **Configuration options**: Reference documentation only
3. **API specifications**: Single source in api-reference
4. **Architecture decisions**: One decision per ADR

### Permitted Cross-References

1. **Links to authoritative source**: Always reference, never duplicate
2. **Context-specific excerpts**: With clear attribution to source
3. **Summary overviews**: High-level summaries with links to details

## Migration Guidelines

### From Current Structure

1. **Consolidate duplicate content**:
   - Merge installation guides with clear scope separation
   - Unify performance documentation
   - Create single configuration reference

2. **Reorganize architecture directory**:
   - Separate ADRs into `/adr/` subdirectory
   - Organize numbered documents in `/decisions/`
   - Create system overview section

3. **Fix broken references**:
   - Update all links to reflect new structure
   - Verify cross-references work correctly
   - Remove references to non-existent directories

### Quality Assurance Process

1. **Technical Accuracy Review**: Verify against codebase
2. **Link Integrity Check**: Validate all internal links
3. **Content Audit**: Ensure single source of truth
4. **Style Consistency**: Apply naming and format standards

## Maintenance Standards

### Regular Audits

1. **Quarterly Reviews**: Full documentation accuracy audit
2. **Release Updates**: Update documentation with each release
3. **Link Validation**: Automated checking of internal links
4. **Content Freshness**: Remove outdated information

### Quality Metrics

1. **Link Integrity**: 100% working internal links
2. **Content Accuracy**: Verified against codebase
3. **Navigation Efficiency**: Users can find information in ≤3 clicks
4. **Search Effectiveness**: Clear structure supports search

## Enforcement

This taxonomy is mandatory for all EntityDB documentation. Violations should be addressed immediately to maintain documentation quality and user experience.

For questions about this taxonomy, refer to the [Developer Guide](./developer-guide/09-documentation-architecture.md) or create an issue for clarification.