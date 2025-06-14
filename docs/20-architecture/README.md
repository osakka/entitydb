# EntityDB Architecture

> **Category**: System Design & Internals | **Target Audience**: Developers & System Architects | **Technical Level**: Advanced

This section provides comprehensive technical documentation of EntityDB's architecture, design decisions, and internal systems. Essential reading for developers, system architects, and anyone needing deep technical understanding.

## 📋 Contents

### [System Overview](./01-system-overview.md)
**High-level architecture and design principles**
- Entity-based architecture philosophy
- Component interaction and data flow
- Service boundaries and responsibilities
- Scalability considerations and limitations
- Performance characteristics and optimization

### [Temporal Architecture](./02-temporal-architecture.md)
**Time-series and temporal data design**
- Nanosecond-precision timestamp system
- Temporal tag storage and indexing
- Time-travel query implementation
- History tracking and audit capabilities
- Performance optimization for temporal operations

### [RBAC Architecture](./03-rbac-architecture.md)
**Role-Based Access Control system design**
- Tag-based permission model
- Hierarchical permission inheritance
- Authentication flow and session management
- Authorization middleware and enforcement
- Security boundaries and threat model

### [Entity Model](./04-entity-model.md)
**Core data model and relationships**
- Entity structure and metadata
- Tag system implementation
- Content storage and chunking
- Relationship modeling and queries
- Binary format (EBF) design rationale

### [Authentication Architecture](./05-authentication-architecture.md)
**Authentication system design (v2.29.0+)**
- Embedded credential storage model
- Bcrypt hashing and salt management
- JWT token generation and validation
- Session lifecycle and refresh mechanisms
- Migration from external credential systems

### [Tag System](./07-tag-system.md)
**Tag-based classification architecture**
- Hierarchical namespace design
- Tag indexing and search optimization
- Temporal tag variants and caching
- Wildcard matching implementation
- Performance considerations and limits

### [Dataset Architecture](./08-dataset-architecture.md)
**Multi-tenancy and data isolation**
- Dataset concept and implementation
- Data isolation boundaries
- Cross-dataset operations and security
- Migration from dataspace terminology
- Scalability patterns for multi-tenant deployments

### [Metrics Architecture](./09-metrics-architecture.md)
**Monitoring and observability design**
- Real-time metrics collection system
- Temporal storage of metric data
- Aggregation and retention policies
- Dashboard integration and visualization
- Performance impact and optimization

## 🎯 Reading Paths

### For System Architects
1. **[System Overview](./01-system-overview.md)** → Overall architecture understanding
2. **[Entity Model](./04-entity-model.md)** → Core data model comprehension
3. **[RBAC Architecture](./03-rbac-architecture.md)** → Security model evaluation
4. **[Temporal Architecture](./02-temporal-architecture.md)** → Time-series capabilities

### For Platform Developers
1. **[Entity Model](./04-entity-model.md)** → Data structure understanding
2. **[Tag System](./07-tag-system.md)** → Classification system details
3. **[Authentication Architecture](./05-authentication-architecture.md)** → Auth integration
4. **[Metrics Architecture](./09-metrics-architecture.md)** → Monitoring integration

### For Performance Engineers
1. **[System Overview](./01-system-overview.md)** → Performance characteristics
2. **[Temporal Architecture](./02-temporal-architecture.md)** → Time-series optimization
3. **[Tag System](./07-tag-system.md)** → Indexing performance
4. **[Metrics Architecture](./09-metrics-architecture.md)** → Monitoring overhead

### For Security Engineers
1. **[RBAC Architecture](./03-rbac-architecture.md)** → Permission model
2. **[Authentication Architecture](./05-authentication-architecture.md)** → Auth security
3. **[Dataset Architecture](./08-dataset-architecture.md)** → Data isolation
4. **[System Overview](./01-system-overview.md)** → Security boundaries

## 🏗️ Key Architectural Principles

### Entity-Centric Design
- Everything is an entity with tags and temporal history
- No specialized tables or complex relational schemas
- Unified operations across all data types
- Flexible schema evolution through tag additions

### Temporal-First Approach
- All data changes captured with nanosecond timestamps
- Time-travel queries as first-class operations
- Audit trail built into core architecture
- Efficient temporal indexing and querying

### Tag-Based Everything
- Classification through hierarchical tag namespaces
- Permissions implemented as tags
- Flexible metadata without schema changes
- Powerful query capabilities through tag combinations

### Performance Optimization
- Binary format (EBF) for efficient storage
- Memory-mapped files with OS-managed caching
- Write-Ahead Logging for durability
- Concurrent access with sharded locking

## 🔗 Quick Navigation

- **Implementation Guides**: [Developer Guides](../60-developer-guides/) - How to extend EntityDB
- **API Details**: [API Reference](../30-api-reference/) - Technical API specifications
- **Deployment**: [Deployment](../70-deployment/) - Production architecture patterns
- **Performance**: [Admin Guides](../50-admin-guides/) - Operational optimization

## 📊 Architecture Metrics

### Design Constraints
- **Single-node deployment**: Current limitation, distributed system planned
- **Binary format storage**: Custom EBF format optimized for entity operations
- **Memory-mapped I/O**: Leverages OS page cache for performance
- **Go runtime**: Single-binary deployment with embedded web interface

### Performance Characteristics
- **Entity Operations**: Sub-millisecond for typical operations
- **Temporal Queries**: Optimized indexing enables fast time-travel
- **Tag Searches**: O(1) lookups with intelligent caching
- **Concurrent Access**: Sharded locking prevents contention

---

*This architecture documentation provides the technical foundation for understanding EntityDB's design decisions and implementation strategies. Use this knowledge to make informed decisions about deployment, integration, and extension.*