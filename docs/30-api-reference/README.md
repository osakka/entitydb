# EntityDB API Reference

> **Category**: API Documentation | **Target Audience**: Developers & Integrators | **Technical Level**: Intermediate

Complete reference documentation for EntityDB's REST API. This section provides detailed endpoint documentation, request/response formats, and comprehensive examples for all API operations.

## 📋 Contents

### [API Overview](./01-overview.md)
**API fundamentals and getting started**
- Base URL and versioning strategy
- Authentication mechanisms and headers
- Request/response formats and conventions
- Error handling and status codes
- Rate limiting and usage guidelines

### [Authentication](./02-authentication.md)
**Authentication and session management**
- Login endpoint and credential formats
- JWT token generation and validation
- Session management and refresh tokens
- Password change and user management
- RBAC integration and permission checks

### [Entities](./03-entities.md)
**Entity CRUD operations and queries**
- Create, read, update, delete operations
- Tag-based filtering and search
- Temporal queries and history retrieval
- Bulk operations and batch processing
- Entity relationships and associations

### [Queries](./04-queries.md)
**Advanced query operations**
- Complex filtering with multiple criteria
- Temporal queries (as-of, history, diff)
- Tag-based searches with wildcards
- Sorting and pagination options
- Performance optimization guidelines

### [Dataset and Metrics APIs](./04-datasets-metrics.md)
**Dataset management and metrics collection**
- Dataset configuration and multi-tenancy support
- Comprehensive metrics endpoints documentation
- Prometheus integration and custom metrics
- Performance monitoring with v2.31.0 optimizations
- Application metrics storage and retrieval

### [Examples](./05-examples.md)
**Comprehensive code examples**
- Common use case implementations
- Language-specific client examples (curl, Python, JavaScript)
- Error handling patterns
- Best practices and optimization tips
- Integration patterns and workflows

## 🚀 Quick Start

### Essential Endpoints
```
POST /api/v1/auth/login          # Authenticate and get token
GET  /api/v1/entities/list       # List entities with filtering
POST /api/v1/entities/create     # Create new entity
GET  /api/v1/entities/get        # Retrieve specific entity
PUT  /api/v1/entities/update     # Update existing entity
```

### Authentication Flow
1. **Login**: `POST /api/v1/auth/login` with username/password
2. **Get Token**: Extract JWT token from response
3. **Use Token**: Include `Authorization: Bearer <token>` in headers
4. **Refresh**: Use refresh endpoint before token expiry

### Basic Entity Operations
```bash
# Create entity
curl -X POST http://localhost:8085/api/v1/entities/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"tags":["type:document","status:draft"],"content":"Hello World"}'

# List entities by type
curl "http://localhost:8085/api/v1/entities/list?tags=type:document" \
  -H "Authorization: Bearer $TOKEN"
```

## 🎯 API Design Principles

### RESTful Conventions
- **GET**: Retrieve data (idempotent, cacheable)
- **POST**: Create new resources
- **PUT**: Update existing resources (idempotent)
- **DELETE**: Remove resources

### Response Format
All API responses follow this consistent structure:
```json
{
  "status": "ok|error",
  "message": "Human-readable description",
  "data": { /* Response payload */ },
  "error": "Error details if applicable"
}
```

### Error Handling
- **400**: Bad Request - Invalid input or malformed request
- **401**: Unauthorized - Missing or invalid authentication
- **403**: Forbidden - Insufficient permissions
- **404**: Not Found - Resource doesn't exist
- **500**: Internal Server Error - Server-side issues

## 🔐 Authentication & Security

### JWT Tokens
- **Format**: RFC 7519 JSON Web Tokens
- **Expiry**: Configurable (default 24 hours)
- **Refresh**: Available before expiry
- **Claims**: User ID, roles, permissions

### RBAC Integration
- **Tag-based permissions**: `rbac:perm:entity:create`
- **Hierarchical inheritance**: `rbac:perm:*` grants all permissions
- **Per-endpoint checks**: Each endpoint requires specific permissions
- **Real-time validation**: Permissions checked on every request

## 📊 Performance Considerations

### Query Optimization
- **Use specific tags**: Avoid broad queries when possible
- **Leverage indexing**: Tag-based filters are optimized
- **Pagination**: Use limit/offset for large result sets
- **Temporal caching**: Recent temporal queries are cached

### Rate Limiting
- **Default limits**: 1000 requests/hour per token
- **Burst capacity**: 100 requests/minute
- **Header information**: Limits included in response headers
- **429 responses**: Rate limit exceeded notifications

## 🔗 Quick Navigation

- **Getting Started**: [Getting Started](../10-getting-started/) - Basic setup and first API calls
- **Architecture**: [Architecture](../20-architecture/) - System design and internal workings
- **User Guides**: [User Guides](../40-user-guides/) - Task-oriented API usage
- **Troubleshooting**: [Troubleshooting](../80-troubleshooting/) - Common API issues

## 📝 API Documentation Tools

### OpenAPI Specification
- **Swagger UI**: Available at `/swagger/` on running server
- **JSON Spec**: Available at `/swagger/doc.json`
- **Interactive Testing**: Test endpoints directly in browser
- **Schema Validation**: Request/response format validation

### Testing Tools
- **Built-in health check**: `GET /health`
- **Metrics endpoint**: `GET /metrics` (Prometheus format)
- **Dashboard**: Web interface for visual API testing
- **CLI tools**: Command-line scripts for automation

---

*This API reference provides complete technical documentation for integrating with EntityDB. Use the examples and specifications to build robust applications on the EntityDB platform.*