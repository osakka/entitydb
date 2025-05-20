#!/bin/bash
# Build all entity-related tools

# Ensure we're in the tools directory
cd "$(dirname "$0")"

# Create bin directory if it doesn't exist
mkdir -p ../../bin

# Build entity management tools
echo "Building add_entity tool..."
go build -o ../../bin/add_entity add_entity.go
if [ $? -eq 0 ]; then
  echo "  Success: add_entity tool built"
else
  echo "  Error: Failed to build add_entity tool"
  exit 1
fi

echo "Building list_entities tool..."
go build -o ../../bin/list_entities list_entities.go
if [ $? -eq 0 ]; then
  echo "  Success: list_entities tool built"
else
  echo "  Error: Failed to build list_entities tool"
  exit 1
fi

# Build entity relationship tools
echo "Building add_entity_relationship tool..."
go build -o ../../bin/add_entity_relationship add_entity_relationship.go
if [ $? -eq 0 ]; then
  echo "  Success: add_entity_relationship tool built"
else
  echo "  Error: Failed to build add_entity_relationship tool"
  exit 1
fi

echo "Building list_entity_relationships tool..."
go build -o ../../bin/list_entity_relationships list_entity_relationships.go
if [ $? -eq 0 ]; then
  echo "  Success: list_entity_relationships tool built"
else
  echo "  Error: Failed to build list_entity_relationships tool"
  exit 1
fi

# Build migration tool
echo "Building migrate_issues_to_entities tool..."
go build -o ../../bin/migrate_issues_to_entities migrate_issues_to_entities.go
if [ $? -eq 0 ]; then
  echo "  Success: migrate_issues_to_entities tool built"
else
  echo "  Error: Failed to build migrate_issues_to_entities tool"
  exit 1
fi

# Set permissions
chmod +x ../../bin/add_entity
chmod +x ../../bin/list_entities
chmod +x ../../bin/add_entity_relationship
chmod +x ../../bin/list_entity_relationships
chmod +x ../../bin/migrate_issues_to_entities

echo "All entity tools built successfully!"
echo "Binaries are available in /opt/entitydb/bin/"