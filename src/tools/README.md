# EntityDB Tools

This directory contains all command-line tools for the EntityDB platform. Tools are written in Go and compiled during the build process.

## Directory Structure

- `users/`: User management tools (add/create users)
- `entities/`: Entity management tools (add/list entities, relationships)
- `maintenance/`: System maintenance tools (fix indexes, diagnostics)
- `diagnostics/`: Performance analysis and debugging tools
- `content/`: Content handling and chunking tools
- `temporal/`: Temporal data management tools

## Building Tools

All tools are built using the Makefile:

```bash
cd /opt/entitydb/src
make tools
```

## Naming Convention

All compiled tools follow the naming convention `entitydb_<tool_name>`. This makes them easily identifiable and avoids naming conflicts.

For example:
- `entitydb_add_user`
- `entitydb_list_entities`
- `entitydb_dump`

## Installation

Compiled tools are placed in the `/opt/entitydb/bin` directory during the build process.

## Tool Categories

### User Management

Tools in the `users/` directory handle user administration:
- Creating users
- Setting passwords
- Managing roles

### Entity Management

Tools in the `entities/` directory handle entity operations:
- Creating entities
- Listing entities
- Managing relationships
- Dumping entity content

### Maintenance

Tools in the `maintenance/` directory handle system maintenance:
- Checking for corrupted entities
- Fixing indexes
- Database integrity checks

### Diagnostics

Tools in the `diagnostics/` directory provide debugging capabilities:
- Binary file analysis
- Performance metrics
- Corruption detection

### Content Management

Tools in the `content/` directory handle content operations:
- Large file chunking
- Content validation
- Streaming data handling

### Temporal Tools

Tools in the `temporal/` directory handle temporal data:
- Fixing temporal tags
- Timeline management
- Historical data operations

## Usage

```bash
# Add a new user
/opt/entitydb/bin/entitydb_add_user -username admin -password securepass

# List entities
/opt/entitydb/bin/entitydb_list_entities -type user

# Dump entity data
/opt/entitydb/bin/entitydb_dump -id abc123
```