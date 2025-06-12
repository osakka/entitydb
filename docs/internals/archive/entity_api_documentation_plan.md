# Entity API Documentation Enhancement Plan

## Current Documentation Status

We have implemented a comprehensive entity-based architecture but need to enhance the API documentation to ensure it's clear, complete, and follows best practices. Currently, we have several documents describing different aspects of the entity system:

- entity_model_architecture.md
- entity_issue_handler_implementation.md
- entity_relationship_implementation.md
- entity_based_migration_guide.md

However, we need a centralized, comprehensive API reference that details all endpoints, request/response formats, and usage examples.

## Documentation Enhancement Plan

### 1. Create Core API Reference Document

Create a new document `/opt/entitydb/docs/entity_api_reference.md` that will serve as the canonical reference for all entity API endpoints:

```markdown
# Entity API Reference

This document provides a comprehensive reference for all Entity API endpoints in the EntityDB system.

## Overview

The Entity API provides a flexible, tag-based approach to storing and retrieving data in EntityDB. Entities can represent any object in the system (issues, workspaces, agents, etc.) and are distinguished by their tags rather than separate schemas.

## Authentication

All Entity API endpoints require authentication using JWT tokens. Include the token in the `Authorization` header:

```
Authorization: Bearer <your_token>
```

## Endpoints

### Entity Management

#### Create Entity

**POST** `/api/v1/entity/create`

Create a new entity with specified tags and content.

**Request:**
```json
{
  "tags": ["type:issue", "priority:high", "status:pending"],
  "content": [
    {
      "key": "title",
      "value": "Sample Issue Title"
    },
    {
      "key": "description",
      "value": "This is a sample issue description"
    }
  ]
}
```

**Response:**
```json
{
  "id": "entity_12345",
  "tags": ["type:issue", "priority:high", "status:pending"],
  "content": [
    {
      "key": "title",
      "value": "Sample Issue Title"
    },
    {
      "key": "description",
      "value": "This is a sample issue description"
    }
  ],
  "created_at": "2025-05-10T15:32:10Z",
  "updated_at": "2025-05-10T15:32:10Z"
}
```

...additional endpoints...
```

### 2. Create Usage Guide with Examples

Create a new document `/opt/entitydb/docs/entity_api_usage_guide.md`:

```markdown
# Entity API Usage Guide

This guide provides practical examples of using the Entity API for common tasks.

## Basic Entity Operations

### Creating Different Entity Types

#### Creating a Workspace

```bash
curl -X POST http://localhost:8085/api/v1/entity/create \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:workspace", "status:active"],
    "content": [
      {
        "key": "name",
        "value": "Project Alpha"
      },
      {
        "key": "description",
        "value": "Strategic initiative for Q2"
      }
    ]
  }'
```

...additional examples...
```

### 3. Update Entity Relationship Documentation

Enhance the existing entity relationship documentation with more details about relationship types, usage patterns, and examples:

```markdown
# Entity Relationships Reference

This document details how entities can be connected through typed relationships in EntityDB.

## Relationship Types

The system supports the following relationship types:

- `parent_of`: Indicates a hierarchical parent-child relationship
- `depends_on`: Indicates a dependency relationship
- `related_to`: Indicates a general relationship
- `assigned_to`: Indicates an assignment relationship (e.g., issue assigned to agent)
- `created_by`: Indicates creator relationship

## Relationship Metadata

Each relationship can store additional metadata as a JSON string, allowing for flexible storage of relationship-specific information...
```

### 4. Create Schema Documentation

Create a document that details the database schema for entities and relationships:

```markdown
# Entity Database Schema

This document describes the database schema used for storing entities and relationships.

## Entity Table

```sql
CREATE TABLE entities (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
```

## Entity Tags Table

```sql
CREATE TABLE entity_tags (
    entity_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (entity_id, tag),
    FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);
```

...additional schema details...
```

### 5. Create Command-Line Tools Documentation

Enhance the documentation for all entity management command-line tools:

```markdown
# Entity Command-Line Tools

This document provides detailed information on using the EntityDB entity management command-line tools.

## add_entity

Create a new entity with specified tags and content.

```
Usage: add_entity <type> <title> [tag1:value1 tag2:value2 ...]

Arguments:
  type       Entity type (e.g., workspace, issue, agent)
  title      Title or name for the entity
  tags       Optional additional tags in key:value format

Examples:
  add_entity workspace "New Project" status:active owner:admin
  add_entity issue "Fix Login Bug" priority:high assignee:john
```

...additional tools...
```

## Implementation Plan

1. Create the core API reference document
2. Develop the usage guide with practical examples
3. Update the relationship documentation
4. Create the schema documentation
5. Improve command-line tools documentation
6. Cross-reference all documentation to make navigation easy
7. Create an index document that links to all entity-related documentation

## Expected Outcome

After implementation, we will have:

1. A comprehensive API reference for all entity endpoints
2. Clear usage examples for common entity operations
3. Detailed documentation of entity relationships and their semantics
4. Schema documentation for database structure
5. Command-line tool documentation for all entity management utilities
6. A well-organized documentation structure that makes information easy to find

This enhanced documentation will make it easier for developers to understand and utilize the entity-based architecture, reducing the learning curve and improving adoption.