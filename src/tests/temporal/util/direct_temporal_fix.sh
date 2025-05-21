#!/bin/bash
# Script to directly fix the ListByTag function in entity_repository.go

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Show a formatted message
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

print_message "$BLUE" "========================================"
print_message "$BLUE" "EntityDB Direct Temporal Tag Fix"
print_message "$BLUE" "========================================"

# Backup original file
print_message "$BLUE" "Backing up original entity_repository.go..."
cp /opt/entitydb/src/storage/binary/entity_repository.go /opt/entitydb/src/storage/binary/entity_repository.go.bak

if [ $? -ne 0 ]; then
  print_message "$RED" "❌ Failed to backup original file."
  exit 1
fi

print_message "$GREEN" "✅ Original file backed up."

# Modify the ListByTag function
print_message "$BLUE" "Modifying ListByTag function..."

# Here's the improved version of ListByTag
cat > /tmp/listbytag.go << 'EOF'
// ListByTag lists entities with a specific tag
func (r *EntityRepository) ListByTag(tag string) ([]*models.Entity, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("tag:%s", tag)
	if cached, found := r.cache.Get(cacheKey); found {
		return cached.([]*models.Entity), nil
	}
	
	r.mu.RLock()
	
	// For non-temporal searches, we need to find tags that match the requested tag
	// regardless of the timestamp prefix
	matchingEntityIDs := make([]string, 0)
	uniqueEntityIDs := make(map[string]bool)
	
	// First check for exact tag match
	if entityIDs, exists := r.tagIndex[tag]; exists {
		for _, entityID := range entityIDs {
			if !uniqueEntityIDs[entityID] {
				uniqueEntityIDs[entityID] = true
				matchingEntityIDs = append(matchingEntityIDs, entityID)
			}
		}
	}
	
	// Now also check for temporal tags (with timestamp prefix)
	for indexedTag, entityIDs := range r.tagIndex {
		if indexedTag == tag {
			continue // Already checked above
		}
		
		// Extract the actual tag part (after the timestamp)
		tagParts := strings.SplitN(indexedTag, "|", 2)
		actualTag := indexedTag
		if len(tagParts) == 2 {
			actualTag = tagParts[1]
		}
		
		// Check if the actual tag matches our search tag
		if actualTag == tag {
			for _, entityID := range entityIDs {
				if !uniqueEntityIDs[entityID] {
					uniqueEntityIDs[entityID] = true
					matchingEntityIDs = append(matchingEntityIDs, entityID)
				}
			}
		}
	}
	
	r.mu.RUnlock()
	
	if len(matchingEntityIDs) == 0 {
		return []*models.Entity{}, nil
	}
	
	// Acquire read locks for all matching entities
	for _, id := range matchingEntityIDs {
		r.lockManager.AcquireEntityLock(id, ReadLock)
		defer r.lockManager.ReleaseEntityLock(id, ReadLock)
	}
	
	// Get a reader from the pool
	readerInterface := r.readerPool.Get()
	if readerInterface == nil {
		reader, err := NewReader(r.getDataFile())
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return r.fetchEntitiesWithReader(reader, matchingEntityIDs)
	}
	
	reader := readerInterface.(*Reader)
	defer r.readerPool.Put(reader)
	
	entities, err := r.fetchEntitiesWithReader(reader, matchingEntityIDs)
	if err == nil {
		// Cache the result
		r.cache.Set(cacheKey, entities)
	}
	return entities, err
}
EOF

# Replace the existing ListByTag function
print_message "$BLUE" "Replacing ListByTag function in entity_repository.go..."

# Use sed to replace the function
sed -i -e '/^\/\/ ListByTag lists entities with a specific tag/,/^}/d' /opt/entitydb/src/storage/binary/entity_repository.go
cat /tmp/listbytag.go >> /opt/entitydb/src/storage/binary/entity_repository.go

# Recompile the code
print_message "$BLUE" "Compiling fixed code..."
cd /opt/entitydb/src
go build -o /opt/entitydb/bin/entitydb_fixed

if [ $? -ne 0 ]; then
  print_message "$RED" "❌ Failed to compile fixed code."
  print_message "$BLUE" "Restoring original file..."
  cp /opt/entitydb/src/storage/binary/entity_repository.go.bak /opt/entitydb/src/storage/binary/entity_repository.go
  exit 1
fi

print_message "$GREEN" "✅ Code compiled successfully."

# Stop the current server
print_message "$BLUE" "Stopping current server..."
if [ -f "/opt/entitydb/bin/entitydbd.sh" ]; then
  /opt/entitydb/bin/entitydbd.sh stop
  sleep 2
else
  pkill -f "entitydb"
  sleep 2
fi

# Start with fixed version
print_message "$BLUE" "Starting fixed server..."
/opt/entitydb/bin/entitydb_fixed &
sleep 5

# Test the fix
print_message "$BLUE" "Testing temporal tag fix..."
cd /opt/entitydb
./improved_temporal_fix.sh

print_message "$GREEN" "✅ Fix process completed!"
print_message "$BLUE" "========================================="