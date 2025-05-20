# EntityDB Documentation

> **⚠️ IMPORTANT DISCLAIMER**
> 
> **This documentation section is currently undergoing heavy reorganization and content cleanup.**
> 
> Some documentation files may contain outdated, incomplete, or partially relevant information. We are actively working to improve and update all documentation to reflect the current state of the EntityDB platform.
>
> For the most accurate and up-to-date information, please refer to:
> - The main [README.md](../README.md) file in the project root
> - Files in the [core/](./core/) directory
> - The [architecture diagram](../share/resources/architecture.svg)
>
> We appreciate your patience as we continue to improve the documentation.

Welcome to the EntityDB documentation. This directory contains documentation for the EntityDB platform.

## Documentation Structure

| Directory | Description |
|-----------|-------------|
| [core/](./core/) | Core documentation, requirements, and specifications |
| [features/](./features/) | Feature-specific documentation |
| [implementation/](./implementation/) | Implementation details and migration guides |
| [performance/](./performance/) | Performance benchmarks and optimization guides |
| [troubleshooting/](./troubleshooting/) | Troubleshooting guides |
| [api/](./api/) | API documentation |
| [architecture/](./architecture/) | Architecture documentation |
| [development/](./development/) | Development guides |
| [guides/](./guides/) | User guides |
| [examples/](./examples/) | Example use cases |
| [spikes/](./spikes/) | Technical spikes and investigations |
| [releases/](./releases/) | Release notes |
| [archive/](./archive/) | Legacy documentation |
| [../share/resources/](../share/resources/) | Logo and design resources (moved from docs) |
| [../trash/](../trash/) | Trash bin for unused/outdated code and files |

> **Important Note:** Always move unused, outdated, or deprecated code to the `/trash` directory instead of deleting it. This makes it easy to reference old implementations if needed while keeping the main codebase clean.

## Core Documentation

| Document | Description |
|----------|-------------|
| [core/PROJECT_REQUIREMENTS.md](./core/PROJECT_REQUIREMENTS.md) | **[NEW]** Comprehensive list of project requirements that guided EntityDB development |
| [core/REQUIREMENTS.md](./core/REQUIREMENTS.md) | System requirements, dependencies, and compatibility information |
| [core/SPECIFICATIONS.md](./core/SPECIFICATIONS.md) | Technical specifications of the EntityDB platform |
| [../CHANGELOG.md](../CHANGELOG.md) | Complete version history and detailed change logs |
| [core/current_state_summary.md](./core/current_state_summary.md) | Summary of current system state |
| [core/query_implementation_summary.md](./core/query_implementation_summary.md) | Query implementation summary |
| [core/RENAMING_SUMMARY.md](./core/RENAMING_SUMMARY.md) | System renaming summary |

## Contributing Guidelines

| Document | Description |
|----------|-------------|
| [core/contributing/CONTRIBUTING.md](./core/contributing/CONTRIBUTING.md) | Contributing guidelines |
| [core/contributing/COLLABORATION.md](./core/contributing/COLLABORATION.md) | Collaboration guidelines |
| [core/contributing/WORKFLOW.md](./core/contributing/WORKFLOW.md) | Development workflow |

## Security Documentation

| Document | Description |
|----------|-------------|
| [core/security/SECURITY.md](./core/security/SECURITY.md) | Security policy and guidelines |

## Feature Documentation

| Document | Description |
|----------|-------------|
| [features/TEMPORAL_FEATURES.md](./features/TEMPORAL_FEATURES.md) | Temporal storage and time-travel queries |
| [features/AUTOCHUNKING.md](./features/AUTOCHUNKING.md) | Autochunking system for large files |
| [features/CUSTOM_BINARY_FORMAT.md](./features/CUSTOM_BINARY_FORMAT.md) | EntityDB Binary Format (EBF) details |
| [features/QUERY_IMPLEMENTATION.md](./features/QUERY_IMPLEMENTATION.md) | Query system implementation |
| [features/TEMPORAL_API_GUIDE.md](./features/TEMPORAL_API_GUIDE.md) | Guide to using temporal API features |
| [features/API_TESTING_FRAMEWORK.md](./features/API_TESTING_FRAMEWORK.md) | Framework for testing API endpoints |
| [features/CONFIG_SYSTEM.md](./features/CONFIG_SYSTEM.md) | Configuration system documentation |

## Architecture Documentation

| Document | Description |
|----------|-------------|
| [architecture/overview.md](./architecture/overview.md) | High-level architecture overview |
| [architecture/entities.md](./architecture/entities.md) | Entity model architecture |
| [architecture/tag_based_rbac.md](./architecture/tag_based_rbac.md) | Tag-based RBAC implementation |
| [architecture/temporal_architecture.md](./architecture/temporal_architecture.md) | Temporal system architecture |
| [architecture/TAG_VALIDATION.md](./architecture/TAG_VALIDATION.md) | Tag system validation report |

## API Documentation

| Document | Description |
|----------|-------------|
| [api/entities.md](./api/entities.md) | Entity API documentation |
| [api/auth.md](./api/auth.md) | Authentication API documentation |
| [api/query_api.md](./api/query_api.md) | Query API documentation |
| [api/examples.md](./api/examples.md) | API usage examples |

## Performance Documentation

| Document | Description |
|----------|-------------|
| [performance/PERFORMANCE.md](./performance/PERFORMANCE.md) | Performance overview and benchmarks |
| [performance/PERFORMANCE_COMPARISON.md](./performance/PERFORMANCE_COMPARISON.md) | Performance comparison with previous versions |
| [performance/TEMPORAL_PERFORMANCE.md](./performance/TEMPORAL_PERFORMANCE.md) | Temporal feature performance |
| [performance/HIGH_PERFORMANCE_MODE_REPORT.md](./performance/HIGH_PERFORMANCE_MODE_REPORT.md) | High-performance mode report |
| [performance/100X_PERFORMANCE_PLAN.md](./performance/100X_PERFORMANCE_PLAN.md) | 100x performance improvement plan |
| [performance/100X_PERFORMANCE_SUMMARY.md](./performance/100X_PERFORMANCE_SUMMARY.md) | 100x performance improvement summary |
| [performance/PERFORMANCE_INDEX.md](./performance/PERFORMANCE_INDEX.md) | Performance index |
| [performance/PERFORMANCE_RESULTS_OLD.md](./performance/PERFORMANCE_RESULTS_OLD.md) | Historical performance results |

## Implementation Details

| Document | Description |
|----------|-------------|
| [implementation/IMPLEMENTATION_STATUS.md](./implementation/IMPLEMENTATION_STATUS.md) | Current implementation status |
| [implementation/TEMPORAL_IMPLEMENTATION.md](./implementation/TEMPORAL_IMPLEMENTATION.md) | Temporal system implementation details |
| [implementation/AUTOCHUNKING_IMPLEMENTATION.md](./implementation/AUTOCHUNKING_IMPLEMENTATION.md) | Autochunking implementation details |
| [implementation/BINARY_FORMAT_IMPLEMENTATION.md](./implementation/BINARY_FORMAT_IMPLEMENTATION.md) | Binary format implementation details |
| [implementation/SSL_IMPLEMENTATION_SUMMARY.md](./implementation/SSL_IMPLEMENTATION_SUMMARY.md) | SSL implementation summary |
| [implementation/SSL_ONLY_MODE.md](./implementation/SSL_ONLY_MODE.md) | SSL-only mode implementation |
| [implementation/SSL_ONLY_SUMMARY.md](./implementation/SSL_ONLY_SUMMARY.md) | SSL-only mode summary |
| [implementation/ENTITY_MODEL_MIGRATION.md](./implementation/ENTITY_MODEL_MIGRATION.md) | Entity model migration guide |
| [implementation/CONTENT_V3_MIGRATION.md](./implementation/CONTENT_V3_MIGRATION.md) | Content v3 migration guide |
| [implementation/CONTENT_V3_EXAMPLE.md](./implementation/CONTENT_V3_EXAMPLE.md) | Content v3 examples |
| [implementation/TEMPORAL_IMPLEMENTATION_SUMMARY.md](./implementation/TEMPORAL_IMPLEMENTATION_SUMMARY.md) | Temporal implementation summary |

## Troubleshooting

| Document | Description |
|----------|-------------|
| [troubleshooting/CONTENT_FORMAT_TROUBLESHOOTING.md](./troubleshooting/CONTENT_FORMAT_TROUBLESHOOTING.md) | Content format troubleshooting guide |
| [troubleshooting/SSL_CONFIGURATION.md](./troubleshooting/SSL_CONFIGURATION.md) | SSL configuration troubleshooting |

## Development Guides

| Document | Description |
|----------|-------------|
| [development/contributing.md](./development/contributing.md) | Contributing guidelines |
| [development/git-workflow.md](./development/git-workflow.md) | **[IMPORTANT]** Centralized Git workflow guide (all developers must follow) |
| [development/production-notes.md](./development/production-notes.md) | Production deployment notes |
| [development/security-implementation.md](./development/security-implementation.md) | Security implementation details |

## User Guides

| Document | Description |
|----------|-------------|
| [guides/quick-start.md](./guides/quick-start.md) | Quick start guide |
| [guides/deployment.md](./guides/deployment.md) | Deployment guide |
| [guides/migration.md](./guides/migration.md) | Migration guide |
| [guides/admin-interface.md](./guides/admin-interface.md) | Admin interface guide |
| [guides/SETUP_ADMIN.md](./guides/SETUP_ADMIN.md) | Admin setup guide |

## Examples and Use Cases

| Document | Description |
|----------|-------------|
| [examples/temporal_examples.md](./examples/temporal_examples.md) | Temporal query examples |
| [examples/ticketing_system.md](./examples/ticketing_system.md) | Ticketing system example |

## Technical Spikes

| Document | Description |
|----------|-------------|
| [spikes/TEMPORAL_STORAGE_SPIKE.md](./spikes/TEMPORAL_STORAGE_SPIKE.md) | Temporal storage implementation investigation |
| [spikes/BINARY_STORAGE_FORMAT_SPIKE.md](./spikes/BINARY_STORAGE_FORMAT_SPIKE.md) | Binary storage format investigation |
| [spikes/AUTOCHUNKING_SPIKE.md](./spikes/AUTOCHUNKING_SPIKE.md) | Autochunking system investigation |

## Release Notes

| Document | Description |
|----------|-------------|
| [releases/RELEASE_NOTES_v2.13.1.md](./releases/RELEASE_NOTES_v2.13.1.md) | v2.13.1 release notes |
| [releases/RELEASE_NOTES_v2.13.0.md](./releases/RELEASE_NOTES_v2.13.0.md) | v2.13.0 release notes |
| [releases/RELEASE_NOTES_v2.12.0.md](./releases/RELEASE_NOTES_v2.12.0.md) | v2.12.0 release notes |

## Resources

Logo resources have been moved to `/share/resources/` directory.

| Resource | Description |
|----------|-------------|
| [/share/resources/logo_white.svg](/share/resources/logo_white.svg) | EntityDB logo - white text version |
| [/share/resources/logo_black.svg](/share/resources/logo_black.svg) | EntityDB logo - black text version |