# EntityDB Documentation Taxonomy and Standards

> **Version**: v2.30.0 | **Last Updated**: 2025-06-12 | **Status**: AUTHORITATIVE

## Documentation Philosophy

EntityDB documentation follows **industry-standard technical writing practices** with emphasis on:

- **Single Source of Truth**: No duplicate information across documents
- **Accuracy First**: All documentation verified against codebase
- **User-Centric Organization**: Structured by user journey and complexity
- **Maintainable Structure**: Clear taxonomy enabling efficient updates

## Master Taxonomy

### Tier 1: User Journey Categories (00-90)

```
00-overview/          # Project introduction and high-level concepts
10-getting-started/   # Installation, quick start, essential concepts  
20-architecture/      # System design, components, data flow
30-api-reference/     # Complete API documentation with examples
40-user-guides/       # Feature-specific guides for end users
50-admin-guides/      # Administrative and operational documentation
60-developer-guides/  # Development, contribution, and integration guides
70-deployment/        # Production deployment and configuration
80-troubleshooting/   # Problem diagnosis and resolution
90-reference/         # Technical reference materials and specifications
```

### Tier 2: Internal Documentation

```
internals/implementation/  # Technical implementation details
internals/archive/        # Historical and deprecated documentation  
internals/analysis/       # Technical analysis and investigation reports
internals/planning/       # Project planning, spikes, and future work
performance/              # Performance analysis and optimization
releases/                 # Version-specific release notes and migration guides
applications/             # Example applications and integration patterns
engineering-excellence/   # Code quality, standards, and improvement initiatives
```

## Naming Schema Standards

### File Naming Convention

```
## Primary Documents
{tier}-{category}/{sequence}-{descriptive-name}.md

Examples:
- 00-overview/01-introduction.md
- 30-api-reference/02-authentication.md
- 40-user-guides/03-temporal-queries.md

## Supporting Documents
{category}/{descriptive-name}.md

Examples:
- internals/implementation/temporal-tag-search-implementation.md
- performance/PERFORMANCE_OPTIMIZATION_REPORT.md
- releases/release-notes-v2.30.0.md
```

### Document Type Classification

#### **AUTHORITATIVE** Documents
- Single source of truth for their domain
- Referenced by other documents
- Maintained with each release
- Examples: API reference, architecture overview

#### **GUIDANCE** Documents  
- Best practices and recommendations
- User journey focused
- Updated as needed
- Examples: User guides, tutorials

#### **REFERENCE** Documents
- Technical specifications and details
- Comprehensive and detailed
- Stable unless major changes
- Examples: Configuration reference, binary format specification

#### **HISTORICAL** Documents
- Previous versions and deprecated content
- Maintained for reference only
- Located in `internals/archive/`
- Examples: Migration guides, legacy API documentation

## Content Standards

### Document Structure Template

```markdown
# {Document Title}

> **Version**: v2.30.0 | **Last Updated**: YYYY-MM-DD | **Status**: {AUTHORITATIVE|GUIDANCE|REFERENCE|HISTORICAL}

## Overview
Brief description and purpose

## Prerequisites  
Required knowledge or setup

## Main Content
Structured content with clear headings

## Examples
Code examples with explanations

## Related Documentation
- [Link to related docs](relative-path.md)

## Version History
- v2.30.0: Current version changes
- v2.29.0: Previous significant changes
```

### Cross-Reference Standards

#### Internal Links
```markdown
[Text](../category/document.md)
[Text](./same-directory.md)
[Section](./document.md#section-anchor)
```

#### Code References
```markdown
[Configuration Setting](../90-reference/01-configuration.md#entitydb_ssl_enabled)
[API Endpoint](../30-api-reference/01-entities.md#create-entity)
[Architecture Component](../20-architecture/02-temporal-storage.md#tag-indexing)
```

### Code Example Standards

#### API Examples
```markdown
## Create Entity

**Endpoint**: `POST /api/v1/entities/create`
**Authentication**: Required
**Permissions**: `entity:create`

### Request
```bash
curl -k -X POST https://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "example-entity-001",
    "tags": ["type:example", "status:active"],
    "content": "SGVsbG8gRW50aXR5REI="
  }'
```

### Response
```json
{
  "id": "example-entity-001",
  "message": "Entity created successfully",
  "timestamp": "2025-06-12T22:45:00Z"
}
```
```

## Version Control Standards

### Document Versioning
- **Major Version**: Significant content restructuring or new major features
- **Minor Version**: Feature additions or substantial updates  
- **Patch Version**: Corrections, clarifications, and minor updates

### Update Triggers
1. **New EntityDB Release**: Update version references and new features
2. **API Changes**: Update API documentation and examples
3. **Architecture Changes**: Update architecture and integration documentation
4. **User Feedback**: Address unclear or incorrect information

### Review Cycle
- **Quarterly**: Complete accuracy review against codebase
- **Release-Based**: Update documentation with each EntityDB release
- **Continuous**: Address issues and improvements as identified

## Quality Assurance

### Validation Checklist
- [ ] Technical accuracy verified against current codebase
- [ ] Version references current (v2.30.0)
- [ ] Code examples tested and functional
- [ ] Cross-references working correctly
- [ ] No duplicate information in other documents
- [ ] Follows naming schema and structure standards

### Maintenance Responsibilities
- **Technical Writers**: Content accuracy, structure, cross-references
- **Developers**: Technical accuracy, code examples, API changes
- **Product Team**: User journey organization, feature prioritization
- **QA Team**: Example validation, link checking, version consistency

## Documentation Categories Reference

### 00-overview/ - Project Introduction
**Purpose**: First impression and high-level understanding  
**Audience**: All users (technical and non-technical)  
**Content**: What is EntityDB, key features, use cases, ecosystem overview

### 10-getting-started/ - Essential First Steps  
**Purpose**: Get users productive quickly  
**Audience**: New users and evaluators  
**Content**: Installation, quick start, basic concepts, first API calls

### 20-architecture/ - System Design
**Purpose**: Technical understanding of EntityDB internals  
**Audience**: Developers, architects, administrators  
**Content**: System components, data flow, temporal storage, RBAC, performance

### 30-api-reference/ - Complete API Documentation
**Purpose**: Authoritative API documentation  
**Audience**: Developers and integrators  
**Content**: All endpoints, parameters, examples, error codes, authentication

### 40-user-guides/ - Feature-Specific Documentation
**Purpose**: Task-oriented guides for specific features  
**Audience**: End users and application developers  
**Content**: Temporal queries, dashboard usage, entity management, relationships

### 50-admin-guides/ - Administrative Documentation  
**Purpose**: System administration and operations  
**Audience**: System administrators and DevOps  
**Content**: Security configuration, deployment, monitoring, backup, maintenance

### 60-developer-guides/ - Development Documentation
**Purpose**: Contributing to and extending EntityDB  
**Audience**: Contributors and plugin developers  
**Content**: Code contribution, git workflow, development setup, architecture extension

### 70-deployment/ - Production Deployment
**Purpose**: Production readiness and deployment  
**Audience**: DevOps, system administrators, architects  
**Content**: Production checklist, SSL setup, scaling, monitoring, troubleshooting

### 80-troubleshooting/ - Problem Resolution
**Purpose**: Diagnose and resolve common issues  
**Audience**: All users experiencing problems  
**Content**: Common issues, diagnostic steps, error interpretation, performance tuning

### 90-reference/ - Technical Specifications
**Purpose**: Detailed technical reference  
**Audience**: Advanced users and developers  
**Content**: Configuration options, binary format, RBAC specification, API complete reference

This taxonomy ensures **comprehensive coverage**, **logical organization**, and **efficient maintenance** of EntityDB documentation while serving diverse user needs and technical complexity levels.