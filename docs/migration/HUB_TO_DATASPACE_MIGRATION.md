# Hub to Dataspace Migration Plan

## Overview
Migrating EntityDB from "hub" concept to "dataspace" concept - a fundamental shift from connection points to isolated data universes.

## Migration Strategy

### Phase 1: Code Rename (Backward Compatible)
1. Add dataspace aliases for all hub functions
2. Update internal variable names
3. Keep hub endpoints working (redirect to dataspace)
4. Add deprecation notices

### Phase 2: Per-Dataspace Indexes
1. Create dataspace-specific index files
2. Implement dataspace index manager
3. Migrate from global to per-dataspace indexes
4. Performance optimization per dataspace

### Phase 3: Data Migration
1. Update all `hub:` tags to `dataspace:`
2. Update RBAC permissions
3. Update client applications
4. Remove hub compatibility layer

## Backward Compatibility Plan

During transition, both work:
- `GET /api/v1/hubs/{name}/entities` → redirects to
- `GET /api/v1/dataspaces/{name}/entities`

Tags support both:
- `hub:worcha` (deprecated)
- `dataspace:worcha` (preferred)

## Breaking Changes
- API endpoints change (with compatibility redirects)
- Tag format changes (automatic migration provided)
- Client libraries need updates