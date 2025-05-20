# EntityDB Tools

This directory contains all command-line tools for the EntityDB platform. Tools are written in Go and compiled during the build process.

## Directory Structure

- `users/`: User management tools (add/create users)
- `entities/`: Entity management tools (add/list entities, relationships)
- `maintenance/`: System maintenance tools (fix indexes, diagnostics)

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

## Usage

```bash
# Add a new user
/opt/entitydb/bin/entitydb_add_user -username admin -password securepass

# List entities
/opt/entitydb/bin/entitydb_list_entities -type user

# Dump entity data
/opt/entitydb/bin/entitydb_dump -id abc123
```