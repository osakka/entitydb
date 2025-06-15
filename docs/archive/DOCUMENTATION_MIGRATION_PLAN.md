# Documentation Migration Plan

## Overview
This plan outlines the steps to reorganize 231 documentation files into a clean, logical structure following industry-standard taxonomy.

## Phase 1: Create New Directory Structure

```bash
mkdir -p docs/00-overview
mkdir -p docs/10-getting-started
mkdir -p docs/20-architecture
mkdir -p docs/30-api-reference
mkdir -p docs/40-user-guides
mkdir -p docs/50-admin-guides
mkdir -p docs/60-developer-guides
mkdir -p docs/70-deployment
mkdir -p docs/80-troubleshooting
mkdir -p docs/90-reference
mkdir -p docs/internals/planning
mkdir -p docs/internals/analysis
mkdir -p docs/internals/implementation
mkdir -p docs/internals/archive
mkdir -p docs/assets/diagrams
mkdir -p docs/assets/images
```

## Phase 2: File Migration Map

### To 00-overview/
- docs/README.md → Keep as master index
- docs/architecture/overview.md → 00-overview/01-introduction.md
- docs/core/specifications.md → 00-overview/02-specifications.md
- docs/core/requirements.md → 00-overview/03-requirements.md

### To 10-getting-started/
- docs/guides/quick-start.md → 10-getting-started/01-quick-start.md
- docs/guides/setup-admin.md → 10-getting-started/02-first-login.md
- docs/development/dev-setup.sh → 10-getting-started/03-development-setup.md

### To 20-architecture/
- docs/architecture/arch-overview.md → 20-architecture/01-system-overview.md
- docs/architecture/arch-temporal.md → 20-architecture/02-temporal-architecture.md
- docs/architecture/arch-rbac.md → 20-architecture/03-rbac-architecture.md
- docs/architecture/entities.md → 20-architecture/04-entity-model.md
- docs/architecture/metrics-architecture-*.md → 20-architecture/05-metrics-architecture.md
- docs/architecture/dataset-architecture-vision.md → 20-architecture/06-dataset-architecture.md
- docs/architecture/tags.md → 20-architecture/07-tag-system.md

### To 30-api-reference/
- docs/api/api-reference.md → 30-api-reference/01-overview.md
- docs/api/auth.md → 30-api-reference/02-authentication.md
- docs/api/entities.md → 30-api-reference/03-entities.md
- docs/api/query_api.md → 30-api-reference/04-queries.md
- docs/api/examples.md → 30-api-reference/05-examples.md

### To 40-user-guides/
- docs/features/temporal-features.md → 40-user-guides/01-temporal-queries.md
- docs/guides/admin-interface.md → 40-user-guides/02-dashboard-guide.md
- docs/features/widget-system.md → 40-user-guides/03-widgets.md
- docs/features/query-implementation.md → 40-user-guides/04-advanced-queries.md

### To 50-admin-guides/
- docs/guides/security.md → 50-admin-guides/01-security-configuration.md
- docs/guides/deployment.md → 50-admin-guides/02-deployment-guide.md
- docs/implementation/*metrics*.md → 50-admin-guides/03-metrics-management.md
- docs/guides/migration.md → 50-admin-guides/04-migration-guide.md

### To 60-developer-guides/
- docs/development/contributing.md → 60-developer-guides/01-contributing.md
- docs/development/git-workflow.md → 60-developer-guides/02-git-workflow.md
- docs/development/logging-standards.md → 60-developer-guides/03-logging-standards.md
- docs/development/configuration-management.md → 60-developer-guides/04-configuration.md

### To 70-deployment/
- docs/guides/deployment.md → 70-deployment/01-production-deployment.md
- docs/development/production-notes.md → 70-deployment/02-production-checklist.md
- docs/troubleshooting/ssl-configuration.md → 70-deployment/03-ssl-setup.md

### To 80-troubleshooting/
- docs/troubleshooting/*.md → 80-troubleshooting/* (rename for consistency)

### To 90-reference/
- docs/features/config-system.md → 90-reference/01-configuration-reference.md
- docs/api/api-reference-complete.md → 90-reference/02-api-complete.md
- docs/features/binary-format.md → 90-reference/03-binary-format-spec.md
- docs/architecture/tag_based_rbac.md → 90-reference/04-rbac-reference.md

### To internals/
- docs/CONSOLIDATION_PLAN.md → internals/planning/
- docs/METRICS_ANALYSIS_FINDINGS.md → internals/analysis/
- docs/implementation/* → internals/implementation/
- docs/archive/* → internals/archive/
- docs/spikes/* → internals/planning/spikes/

## Phase 3: Update Cross-References

1. Search and replace all documentation links
2. Update relative paths
3. Verify all internal links work
4. Update external references

## Phase 4: Add Front Matter

Add to every document:
```yaml
---
title: [Document Title]
category: [Category Name]
tags: [relevant, tags, here]
last_updated: 2025-06-11
version: v2.29.0
---
```

## Phase 5: Create Category Indexes

Create README.md in each category folder:
- Overview of category
- List of documents with descriptions
- Navigation to related categories

## Phase 6: Validation

1. Run link checker
2. Verify all files moved
3. Check for duplicates
4. Ensure naming consistency
5. Update main indexes

## Files to Remove (Duplicates/Obsolete)

- docs/api/api-reference-complete.md (duplicate of api-reference.md)
- docs/architecture/overview.md (duplicate of arch-overview.md)
- docs/archive/* (move to internals/archive)
- Any .bak or temporary files

## Timeline

- Phase 1-2: File migration (2 hours)
- Phase 3: Update references (3 hours)
- Phase 4: Add front matter (2 hours)
- Phase 5: Create indexes (1 hour)
- Phase 6: Validation (1 hour)

Total estimated time: 9 hours