---
title: API Reference
category: API Reference
tags: [api, rest, endpoints]
last_updated: 2025-06-11
version: v2.29.0
---

# EntityDB API Reference

Complete documentation for all EntityDB REST API endpoints.

## Documents

### [01-overview.md](./01-overview.md)
API overview, authentication, response formats, and general patterns.

### [02-authentication.md](./02-authentication.md)
Authentication endpoints, JWT tokens, and session management.

### [03-entities.md](./03-entities.md)
Entity CRUD operations, querying, and temporal features.

### [04-queries.md](./04-queries.md)
Advanced query syntax, filtering, sorting, and temporal queries.

### [05-examples.md](./05-examples.md)
Practical examples and common API usage patterns.

## Quick Reference

### Authentication
- `POST /api/v1/auth/login` - Authenticate user
- `POST /api/v1/auth/refresh` - Refresh session token
- `POST /api/v1/auth/logout` - End session

### Entities
- `GET /api/v1/entities/list` - List entities
- `GET /api/v1/entities/get` - Get specific entity
- `POST /api/v1/entities/create` - Create new entity
- `PUT /api/v1/entities/update` - Update entity
- `GET /api/v1/entities/query` - Advanced entity queries

### Temporal
- `GET /api/v1/entities/as-of` - Query at specific time
- `GET /api/v1/entities/history` - Get entity history
- `GET /api/v1/entities/changes` - Get changes in time range
- `GET /api/v1/entities/diff` - Compare entity states

### Relationships
- `POST /api/v1/entity-relationships` - Create relationship
- `GET /api/v1/entity-relationships` - Query relationships

### System
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /api/v1/system/metrics` - System metrics

## Base URL

```
HTTP:  http://localhost:8085
HTTPS: https://localhost:8443
```

## Authentication

All API endpoints (except `/health` and `/metrics`) require authentication:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8085/api/v1/entities/list
```

## Error Responses

All errors follow consistent format:

```json
{
  "error": "error_code",
  "message": "Human readable error message",
  "details": {
    "field": "Additional context"
  }
}
```

## Getting Started

1. [Authenticate](./02-authentication.md#login) to get a token
2. [Create your first entity](./03-entities.md#create-entity)
3. [Query entities](./04-queries.md) by tags
4. Explore [temporal features](./01-overview.md#temporal-queries)