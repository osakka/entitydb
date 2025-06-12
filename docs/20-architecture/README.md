---
title: Architecture Documentation
category: Architecture
tags: [architecture, design, system]
last_updated: 2025-06-11
version: v2.29.0
---

# EntityDB Architecture

This section contains detailed technical documentation about EntityDB's architecture and design decisions.

## Documents

### [01-system-overview.md](./01-system-overview.md)
High-level system architecture, core components, and design principles.

### [02-temporal-architecture.md](./02-temporal-architecture.md)
Temporal storage system, timeline indexing, and time-based query architecture.

### [03-rbac-architecture.md](./03-rbac-architecture.md)
Role-Based Access Control system, permission model, and security architecture.

### [04-entity-model.md](./04-entity-model.md)
Entity model design, tag systems, and data organization.

### [07-tag-system.md](./07-tag-system.md)
Tag namespaces, indexing, and query optimization.

## Quick Links

- **Performance**: [Temporal Performance](../performance/)
- **Security**: [RBAC Reference](../90-reference/04-rbac-reference.md)
- **API Design**: [API Reference](../30-api-reference/)
- **Storage**: [Binary Format Spec](../90-reference/03-binary-format-spec.md)

## Architecture Principles

1. **Everything is an Entity**: Unified data model
2. **Temporal by Default**: All changes preserved with nanosecond precision  
3. **Tag-Based Organization**: Flexible, queryable metadata
4. **High Performance**: Custom binary format with memory mapping
5. **RBAC Security**: Fine-grained permission system
6. **Horizontal Scalability**: Dataset-based partitioning