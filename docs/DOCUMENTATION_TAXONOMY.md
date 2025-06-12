# EntityDB Documentation Taxonomy and Organization Plan

## Overview

This document defines the standardized taxonomy, naming conventions, and organizational structure for EntityDB documentation. All documentation should follow these guidelines to ensure consistency, discoverability, and maintainability.

## Directory Structure

```
docs/
├── README.md                    # Master index and navigation guide
├── CHANGELOG.md                 # Moved from root for centralization
├── 00-overview/                 # Project overview and introduction
├── 10-getting-started/          # Quick start and installation guides
├── 20-architecture/             # System design and architecture
├── 30-api-reference/            # Complete API documentation
├── 40-user-guides/              # End-user documentation
├── 50-admin-guides/             # Administration and operations
├── 60-developer-guides/         # Development and contribution
├── 70-deployment/               # Production deployment guides
├── 80-troubleshooting/          # Problem resolution guides
├── 90-reference/                # Technical specifications and references
├── internals/                   # Internal documentation (not user-facing)
│   ├── planning/                # Project planning documents
│   ├── analysis/                # Technical analysis and findings
│   ├── implementation/          # Implementation details and notes
│   └── archive/                 # Deprecated/historical documents
└── assets/                      # Images, diagrams, and other assets
```

## Naming Conventions

### File Naming
- Use lowercase with hyphens (kebab-case): `api-reference.md`
- Include numeric prefixes for ordering: `01-introduction.md`
- Use descriptive, specific names: `temporal-query-guide.md` not `queries.md`
- Version-specific docs include version: `migration-v2.28-to-v2.29.md`

### Document Types
- Guides: `{topic}-guide.md` (e.g., `authentication-guide.md`)
- References: `{topic}-reference.md` (e.g., `api-reference.md`)
- Tutorials: `{topic}-tutorial.md` (e.g., `query-tutorial.md`)
- Troubleshooting: `troubleshooting-{issue}.md` (e.g., `troubleshooting-ssl.md`)
- Architecture: `architecture-{component}.md` (e.g., `architecture-temporal.md`)

## Content Standards

### Front Matter
Every document should start with:
```yaml
---
title: Document Title
category: Category Name
tags: [tag1, tag2, tag3]
last_updated: 2025-06-11
version: v2.29.0
---
```

### Document Structure
1. **Title** (H1) - Clear, descriptive title
2. **Overview** - Brief description of document purpose
3. **Prerequisites** - What reader needs to know/have
4. **Content** - Main documentation body
5. **Next Steps** - Related documents or actions
6. **References** - Links to related documentation

### Cross-References
- Always use relative paths: `[API Reference](../30-api-reference/01-overview.md)`
- Include section anchors: `[Authentication](../30-api-reference/02-auth.md#jwt-tokens)`
- Verify all links during documentation reviews

## Categories

### 00-overview
- Project introduction
- Feature overview
- Architecture summary
- Getting help

### 10-getting-started
- Installation guide
- Quick start tutorial
- First steps
- Basic concepts

### 20-architecture
- System design documents
- Component architecture
- Data models
- Security architecture
- Performance architecture

### 30-api-reference
- REST API endpoints
- Request/response formats
- Authentication
- Error codes
- Examples

### 40-user-guides
- Entity management
- Query building
- Temporal features
- Dashboard usage

### 50-admin-guides
- User management
- RBAC configuration
- System monitoring
- Backup and recovery

### 60-developer-guides
- Development setup
- Contributing guidelines
- Code standards
- Testing guide
- Git workflow

### 70-deployment
- Production deployment
- Configuration management
- SSL/TLS setup
- Performance tuning
- Scaling guide

### 80-troubleshooting
- Common issues
- Error resolution
- Performance problems
- Debug procedures

### 90-reference
- Configuration reference
- Environment variables
- Binary format specification
- Tag namespaces
- Glossary

## Migration Plan

1. Create new directory structure
2. Move and rename files according to taxonomy
3. Update all cross-references
4. Add front matter to all documents
5. Create comprehensive indexes
6. Archive obsolete documentation
7. Update root README.md and docs/README.md

## Maintenance

- Quarterly documentation review
- Version-specific updates tracked
- Deprecation notices added 2 versions ahead
- Archive old versions in internals/archive
- Regular link validation
- Keep index files updated