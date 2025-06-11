# Dataset to Dataset Migration Plan

## Overview
Migrating EntityDB from "dataset" concept to "dataset" concept - a fundamental shift from connection points to isolated data universes.

## Migration Strategy

### Phase 1: Code Rename (Backward Compatible)
1. Add dataset aliases for all dataset functions
2. Update internal variable names
3. Keep dataset endpoints working (redirect to dataset)
4. Add deprecation notices

### Phase 2: Per-Dataset Indexes
1. Create dataset-specific index files
2. Implement dataset index manager
3. Migrate from global to per-dataset indexes
4. Performance optimization per dataset

### Phase 3: Data Migration
1. Update all `dataset:` tags to `dataset:`
2. Update RBAC permissions
3. Update client applications
4. Remove dataset compatibility layer

## Backward Compatibility Plan

During transition, both work:
- `GET /api/v1/datasets/{name}/entities` → redirects to
- `GET /api/v1/datasets/{name}/entities`

Tags support both:
- `dataset:worca` (deprecated)
- `dataset:worca` (preferred)

## Breaking Changes
- API endpoints change (with compatibility redirects)
- Tag format changes (automatic migration provided)
- Client libraries need updates