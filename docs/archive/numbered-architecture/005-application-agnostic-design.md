# ADR-005: Application-Agnostic Platform Design

## Status
Accepted (2025-06-02)

## Context
EntityDB initially included embedded applications (Worca workforce orchestrator, Methub) within the core platform. This created coupling between the database platform and specific application domains.

### Problems with Embedded Applications
- **Platform Bloat**: Core database included non-essential application code
- **Maintenance Overhead**: Multiple application domains to maintain
- **User Confusion**: Unclear separation between platform and applications
- **API Complexity**: Application-specific endpoints mixed with core database APIs
- **Deployment Complexity**: Unused applications shipped with every deployment

### Alternative Approaches
1. **Keep embedded applications**: Maintain status quo with built-in apps
2. **Separate repositories**: Move applications to separate codebases
3. **Plugin architecture**: Create plugin system for applications
4. **Pure platform**: Remove all embedded applications entirely

## Decision
We decided to implement a **pure application-agnostic platform design** by removing all embedded applications from the core EntityDB distribution.

### Implementation Changes
- **Removed Applications**: Moved Worca and Methub to archive/trash
- **Generic Metrics API**: Replaced `/api/v1/worca/metrics` with `/api/v1/application/metrics`
- **Namespace Filtering**: Applications can filter metrics by namespace/app parameter
- **Clean Distribution**: Core platform ships without application code
- **Documentation Updates**: Removed application-specific documentation

### API Changes
```
Before: /api/v1/worca/metrics
After:  /api/v1/application/metrics?namespace=worca

Before: Application-specific handlers
After:  Generic application endpoints with filtering
```

## Consequences

### Positive
- **Platform Focus**: Clear separation between database and applications
- **Reduced Complexity**: Smaller, more focused codebase
- **Deployment Flexibility**: Users deploy only what they need
- **API Clarity**: Clean separation between platform and application APIs
- **Maintenance**: Reduced maintenance burden for core team
- **Reusability**: Platform can support any application domain

### Negative
- **Example Applications**: No built-in examples for new users
- **Documentation Gap**: Fewer concrete usage examples
- **Migration Effort**: Existing Worca users need to migrate
- **Learning Curve**: Users must build their own applications

### Mitigation Strategies
- **Sample Applications**: Provide sample applications in separate repositories
- **Documentation**: Enhanced examples and tutorials in documentation
- **Migration Guides**: Clear migration paths for existing application users
- **Community**: Encourage community-contributed applications

## Implementation History
- v2.23.0: Application-agnostic design implementation
- v2.29.0: Removed duplicate dashboard files and cleaned up application references
- v2.32.2: Updated documentation to remove outdated application references

## Future Considerations
- **Sample Applications**: Create separate repository for example applications
- **Plugin System**: Consider plugin architecture for optional integrations
- **Community**: Foster ecosystem of third-party applications
- **Templates**: Provide application templates for common use cases

## Related Decisions
- [ADR-004: Tag-Based RBAC](./004-tag-based-rbac.md) - Generic permission system
- Platform design supports any application domain through generic APIs