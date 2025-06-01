# EntityDB Documentation Taxonomy

## Document Categories and Naming Conventions

### 1. Core Documentation (`/docs/core/`)
**Purpose**: Essential project information and specifications

**Naming Convention**: 
- `UPPERCASE.md` for critical documents
- `lowercase_with_underscores.md` for regular documents

**Categories**:
- `PROJECT_REQUIREMENTS.md` - Project requirements
- `REQUIREMENTS.md` - System requirements
- `SPECIFICATIONS.md` - Technical specifications
- `current_state_summary.md` - Current platform state
- `contributing/` - Contribution guidelines
- `security/` - Security policies

### 2. API Documentation (`/docs/api/`)
**Purpose**: API reference and examples

**Naming Convention**: `lowercase.md` or `snake_case.md`

**Categories**:
- `auth.md` - Authentication endpoints
- `entities.md` - Entity CRUD operations
- `query_api.md` - Query system
- `examples.md` - Usage examples
- `*_demo.md` - Interactive demonstrations

### 3. Architecture Documentation (`/docs/architecture/`)
**Purpose**: System design and architectural decisions

**Naming Convention**: `lowercase_with_underscores.md`

**Categories**:
- `overview.md` - High-level architecture
- `entities.md` - Data model architecture
- `*_architecture.md` - Subsystem architectures
- `*_implementation.md` - Implementation details
- `TAG_*.md` - Tag system documentation

### 4. Feature Documentation (`/docs/features/`)
**Purpose**: Individual feature documentation

**Naming Convention**: `UPPERCASE_FEATURE.md`

**Categories**:
- `TEMPORAL_FEATURES.md` - Major features
- `AUTOCHUNKING.md` - System capabilities
- `*_IMPLEMENTATION.md` - Feature implementations
- `*_GUIDE.md` - Usage guides

### 5. Implementation Documentation (`/docs/implementation/`)
**Purpose**: Detailed implementation specifications

**Naming Convention**: `UPPERCASE_TOPIC.md`

**Categories**:
- `*_IMPLEMENTATION.md` - Implementation details
- `*_SUMMARY.md` - Implementation summaries
- `*_PLAN.md` - Implementation plans
- `*_STATUS.md` - Status reports

### 6. Performance Documentation (`/docs/performance/`)
**Purpose**: Performance metrics and optimization

**Naming Convention**: `UPPERCASE_PERFORMANCE_*.md`

**Categories**:
- `PERFORMANCE.md` - Main performance doc
- `*_PERFORMANCE_*.md` - Specific metrics
- `*_REPORT.md` - Performance reports
- `*_PLAN.md` - Optimization plans

### 7. Development Documentation (`/docs/development/`)
**Purpose**: Developer guides and workflows

**Naming Convention**: `lowercase-with-hyphens.md`

**Categories**:
- `contributing.md` - Contribution guide
- `git-workflow.md` - Development process
- `*-notes.md` - Development notes
- `*-implementation.md` - Technical guides

### 8. User Guides (`/docs/guides/`)
**Purpose**: End-user documentation

**Naming Convention**: `lowercase-with-hyphens.md`

**Categories**:
- `quick-start.md` - Getting started
- `deployment.md` - Deployment guide
- `migration.md` - Upgrade guides
- `*-interface.md` - UI guides

### 9. Troubleshooting (`/docs/troubleshooting/`)
**Purpose**: Problem resolution guides

**Naming Convention**: `UPPERCASE_ISSUE.md`

**Categories**:
- `*_TROUBLESHOOTING.md` - Issue guides
- `*_BUG.md` - Known bugs
- `*_CONFIGURATION.md` - Config issues

### 10. Release Documentation (`/docs/releases/`)
**Purpose**: Version-specific information

**Naming Convention**: `RELEASE_NOTES_vX.Y.Z.md`

**Categories**:
- `RELEASE_NOTES_*.md` - Release notes
- `*_SUMMARY.md` - Version summaries

### 11. Examples (`/docs/examples/`)
**Purpose**: Code examples and use cases

**Naming Convention**: `lowercase_examples.md`

**Categories**:
- `*_examples.md` - Code samples
- `*_system.md` - Complete examples

### 12. Archive (`/docs/archive/`)
**Purpose**: Historical documentation

**Naming Convention**: Preserve original naming

**Categories**:
- Legacy documentation
- Deprecated features
- Historical decisions

## Document Structure Standards

### Required Sections
1. **Title** (H1) - Clear, descriptive title
2. **Description** - One-line summary
3. **Table of Contents** - For documents > 200 lines
4. **Overview/Introduction** - Context and purpose
5. **Main Content** - Organized with clear headers
6. **Examples** - Where applicable
7. **References** - Links to related docs

### Optional Sections
- **Prerequisites**
- **Configuration**
- **Troubleshooting**
- **FAQ**
- **Changelog** (for evolving documents)

### Metadata Header (for key documents)
```markdown
---
title: Document Title
category: architecture
status: current|deprecated|draft
version: 1.0
last_updated: 2025-05-30
---
```

## Cross-Reference Guidelines

### Internal Links
- Use relative paths: `[Link Text](../category/document.md)`
- Link to sections: `[Section](#section-header)`
- Verify links work before committing

### External Links
- Use full URLs for external resources
- Add link descriptions for context

## File Organization Rules

1. **No documentation in source directories** - Move to `/docs`
2. **Group by topic** - Use subdirectories for organization
3. **Archive, don't delete** - Move outdated docs to `/docs/archive`
4. **One topic per file** - Split large documents
5. **Consistent naming** - Follow category conventions

## Maintenance Process

1. **Regular Reviews** - Quarterly documentation audits
2. **Code Sync** - Update docs with code changes
3. **Link Checking** - Verify all links monthly
4. **Archive Old Versions** - When making major updates
5. **Update Index** - Keep `/docs/README.md` current

## Quality Standards

- **Accuracy**: All code examples must be tested
- **Clarity**: Write for the target audience
- **Completeness**: Cover all aspects of the topic
- **Currency**: Keep synchronized with code
- **Consistency**: Follow this taxonomy