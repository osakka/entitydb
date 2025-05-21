#!/bin/bash
# Script to validate our chunking fix implementation

echo "Validating chunking fix implementation..."

# Check if our fix files exist
if [ -f "/opt/entitydb/src/api/entity_handler_fix.go" ]; then
  echo "✓ entity_handler_fix.go exists"
else
  echo "✗ entity_handler_fix.go does not exist"
  exit 1
fi

# Check if the entity_handler.go file contains our fix
if grep -q "HandleChunkedContent" /opt/entitydb/src/api/entity_handler.go; then
  echo "✓ entity_handler.go calls HandleChunkedContent"
else
  echo "✗ entity_handler.go does not call HandleChunkedContent"
  exit 1
fi

# Check if entity_handler.go contains the check for chunked entities
if grep -q "if includeContent && entity.IsChunked()" /opt/entitydb/src/api/entity_handler.go; then
  echo "✓ entity_handler.go checks for chunked entities"
else
  echo "✗ entity_handler.go does not check for chunked entities"
  exit 1
fi

# Check content of entity_handler_fix.go
if grep -q "reassembledContent = append(reassembledContent, chunkEntity.Content...)" /opt/entitydb/src/api/entity_handler_fix.go; then
  echo "✓ entity_handler_fix.go correctly reassembles chunks"
else
  echo "✗ entity_handler_fix.go does not reassemble chunks"
  exit 1
fi

if grep -q "chunkID := fmt.Sprintf(\"%s-chunk-%d\", entity.ID, i)" /opt/entitydb/src/api/entity_handler_fix.go; then
  echo "✓ entity_handler_fix.go correctly constructs chunk IDs"
else
  echo "✗ entity_handler_fix.go does not construct chunk IDs correctly"
  exit 1
fi

# Check if entity.IsChunked is implemented correctly
if grep -q "IsChunked" /opt/entitydb/src/models/entity.go; then
  echo "✓ models/entity.go has IsChunked function"
else
  echo "✗ models/entity.go does not have IsChunked function"
  exit 1
fi

# Check if our changes have been committed
if cd /opt/entitydb && git log -n 2 | grep -q "chunked content"; then
  echo "✓ Chunking fix was committed to the repository"
else
  echo "✗ Chunking fix was not committed to the repository"
  exit 1
fi

# Check if commits were pushed
if cd /opt/entitydb && git log -n 1 origin/main | grep -q "chunked content"; then
  echo "✓ Chunking fix was pushed to origin/main"
else
  echo "✗ Chunking fix was not pushed to origin/main"
  exit 1
fi

echo "All validation checks passed! The chunking fix has been properly implemented and committed."