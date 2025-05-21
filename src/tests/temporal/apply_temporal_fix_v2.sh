#!/bin/bash
# Apply temporal fixes to EntityDB V2

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Temporal Features Fix V2${NC}"
echo -e "${BLUE}========================================${NC}"

# First, stop the running server
echo -e "${BLUE}Stopping EntityDB server...${NC}"
cd /opt/entitydb && ./bin/entitydbd.sh stop

# Create a special patch test file
echo -e "${BLUE}Creating temporal test patch...${NC}"
cat > /opt/entitydb/src/api/temporal_test_patch.go << 'EOF'
package api

import (
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"net/http"
	"time"
	"strings"
)

// Simple handler for testing fixed temporal features
func (h *EntityHandler) TestTemporalFixHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the requested operation from the URL path
	path := r.URL.Path
	operation := ""
	if parts := strings.Split(path, "/"); len(parts) > 0 {
		operation = parts[len(parts)-1]
	}
	
	logger.Debug("Test temporal fix handler called for operation: %s", operation)
	
	// Get repository as temporal repository
	temporalRepo, ok := h.repo.(*binary.TemporalRepository)
	if !ok {
		RespondError(w, http.StatusInternalServerError, "Repository does not support temporal features")
		return
	}
	
	switch operation {
	case "as-of-test":
		testAsOf(w, r, temporalRepo)
	case "history-test":
		testHistory(w, r, temporalRepo)
	case "changes-test":
		testChanges(w, r, temporalRepo)
	case "diff-test":
		testDiff(w, r, temporalRepo)
	default:
		RespondError(w, http.StatusBadRequest, "Unknown test operation")
	}
}

// Test the as-of functionality
func testAsOf(w http.ResponseWriter, r *http.Request, repo *binary.TemporalRepository) {
	// Get parameters
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}
	
	timestampStr := r.URL.Query().Get("timestamp")
	if timestampStr == "" {
		RespondError(w, http.StatusBadRequest, "Timestamp is required")
		return
	}
	
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid timestamp format")
		return
	}
	
	// Get entity as of timestamp using fixed implementation
	entity, err := repo.GetEntityAsOfFixed(entityID, timestamp)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get entity as of timestamp: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, entity)
}

// Test the history functionality
func testHistory(w http.ResponseWriter, r *http.Request, repo *binary.TemporalRepository) {
	// Get parameters
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}
	
	// Use fixed implementation
	history, err := repo.GetEntityHistoryFixed(entityID, 100)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get entity history: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, history)
}

// Test the changes functionality
func testChanges(w http.ResponseWriter, r *http.Request, repo *binary.TemporalRepository) {
	// Use fixed implementation
	changes, err := repo.GetRecentChangesFixed(100)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get recent changes: "+err.Error())
		return
	}
	
	RespondJSON(w, http.StatusOK, changes)
}

// Test the diff functionality
func testDiff(w http.ResponseWriter, r *http.Request, repo *binary.TemporalRepository) {
	// Get parameters
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}
	
	t1Str := r.URL.Query().Get("t1")
	t2Str := r.URL.Query().Get("t2")
	if t1Str == "" || t2Str == "" {
		RespondError(w, http.StatusBadRequest, "Both t1 and t2 timestamps are required")
		return
	}
	
	t1, err := time.Parse(time.RFC3339, t1Str)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid t1 timestamp format")
		return
	}
	
	t2, err := time.Parse(time.RFC3339, t2Str)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid t2 timestamp format")
		return
	}
	
	// Use fixed implementation
	before, after, err := repo.GetEntityDiffFixed(entityID, t1, t2)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get entity diff: "+err.Error())
		return
	}
	
	// Construct useful response
	response := map[string]interface{}{
		"before":       before,
		"after":        after,
		"from":         t1.Format(time.RFC3339),
		"to":           t2.Format(time.RFC3339),
		"entity_id":    entityID,
	}
	
	// Add helpful diff information
	if before != nil && after != nil {
		beforeTags := before.GetTagsWithoutTimestamp()
		afterTags := after.GetTagsWithoutTimestamp()
		
		// Find added tags
		addedTags := []string{}
		for _, tag := range afterTags {
			found := false
			for _, beforeTag := range beforeTags {
				if tag == beforeTag {
					found = true
					break
				}
			}
			if !found {
				addedTags = append(addedTags, tag)
			}
		}
		
		// Find removed tags
		removedTags := []string{}
		for _, tag := range beforeTags {
			found := false
			for _, afterTag := range afterTags {
				if tag == afterTag {
					found = true
					break
				}
			}
			if !found {
				removedTags = append(removedTags, tag)
			}
		}
		
		response["added_tags"] = addedTags
		response["removed_tags"] = removedTags
	}
	
	RespondJSON(w, http.StatusOK, response)
}
EOF

# Update main.go to register test endpoints
echo -e "${BLUE}Patching main.go to add test endpoints...${NC}"
cat > /opt/entitydb/src/temp_patch.sh << 'EOF'
#!/bin/bash
# Register the test endpoints
LINE_NUM=$(grep -n "// Test endpoints (no auth required)" /opt/entitydb/src/main.go | cut -d: -f1)
if [ -z "$LINE_NUM" ]; then
    echo "Could not find insertion point in main.go"
    exit 1
fi

# Insert test endpoints a few lines below the insertion point
INSERT_LINE=$((LINE_NUM + 15))

# Insert test routes
sed -i "${INSERT_LINE}i\\\t// Temporal fix test endpoints\n\tapiRouter.HandleFunc(\"/test/temporal/as-of-test\", server.entityHandler.TestTemporalFixHandler).Methods(\"GET\")\n\tapiRouter.HandleFunc(\"/test/temporal/history-test\", server.entityHandler.TestTemporalFixHandler).Methods(\"GET\")\n\tapiRouter.HandleFunc(\"/test/temporal/changes-test\", server.entityHandler.TestTemporalFixHandler).Methods(\"GET\")\n\tapiRouter.HandleFunc(\"/test/temporal/diff-test\", server.entityHandler.TestTemporalFixHandler).Methods(\"GET\")" /opt/entitydb/src/main.go
EOF

chmod +x /opt/entitydb/src/temp_patch.sh
/opt/entitydb/src/temp_patch.sh

# Build the server
echo -e "${BLUE}Building EntityDB server with temporal fixes...${NC}"
cd /opt/entitydb/src && make

# Start the server
echo -e "${BLUE}Starting EntityDB server...${NC}"
cd /opt/entitydb && ./bin/entitydbd.sh start

# Wait for server to start
echo -e "${BLUE}Waiting for server to initialize...${NC}"
sleep 5

# Create test entity for temporal testing
echo -e "${BLUE}Creating test entity for temporal features...${NC}"
curl -k -s -X POST "https://localhost:8085/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' > /tmp/login.json

TOKEN=$(grep -o '"token":"[^"]*' /tmp/login.json | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo -e "${RED}❌ Login failed${NC}"
  exit 1
fi

echo -e "${GREEN}✅ Logged in successfully${NC}"

# Create first entity version
echo -e "${BLUE}Creating initial entity version...${NC}"
curl -k -s -X POST "https://localhost:8085/api/v1/entities/create" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tags": ["type:temporal_fix_test", "version:1"],
    "content": "Initial content for temporal fix testing"
  }' > /tmp/entity1.json

ENTITY_ID=$(grep -o '"id":"[^"]*' /tmp/entity1.json | cut -d'"' -f4)
CREATED_AT=$(grep -o '"created_at":"[^"]*' /tmp/entity1.json | cut -d'"' -f4)

if [ -z "$ENTITY_ID" ]; then
  echo -e "${RED}❌ Failed to create test entity${NC}"
  exit 1
fi

echo -e "${GREEN}✅ Created entity with ID: $ENTITY_ID${NC}"
echo -e "${BLUE}CreatedAt: $CREATED_AT${NC}"

# Sleep to ensure different timestamps
sleep 2

# Update the entity
echo -e "${BLUE}Updating entity to version 2...${NC}"
curl -k -s -X PUT "https://localhost:8085/api/v1/entities/update" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"id\": \"$ENTITY_ID\",
    \"tags\": [\"type:temporal_fix_test\", \"version:2\", \"updated:true\"],
    \"content\": \"Updated content for temporal fix testing\"
  }" > /tmp/entity2.json

UPDATED_AT=$(grep -o '"updated_at":"[^"]*' /tmp/entity2.json | cut -d'"' -f4)
echo -e "${BLUE}UpdatedAt: $UPDATED_AT${NC}"

# Test as-of endpoint
echo -e "${BLUE}Testing as-of endpoint with fixed implementation...${NC}"
curl -k -s -X GET "https://localhost:8085/api/v1/test/temporal/as-of-test?id=$ENTITY_ID&timestamp=$CREATED_AT" \
  -H "Authorization: Bearer $TOKEN" > /tmp/asof.json

if grep -q "version:1" /tmp/asof.json; then
  echo -e "${GREEN}✅ As-of endpoint working correctly${NC}"
else
  echo -e "${RED}❌ As-of endpoint not working correctly${NC}"
  echo "Response: $(cat /tmp/asof.json)"
fi

# Test history endpoint
echo -e "${BLUE}Testing history endpoint with fixed implementation...${NC}"
curl -k -s -X GET "https://localhost:8085/api/v1/test/temporal/history-test?id=$ENTITY_ID" \
  -H "Authorization: Bearer $TOKEN" > /tmp/history.json

HISTORY_COUNT=$(grep -o "timestamp" /tmp/history.json | wc -l)
if [ "$HISTORY_COUNT" -gt 0 ]; then
  echo -e "${GREEN}✅ History endpoint working correctly with $HISTORY_COUNT entries${NC}"
else
  echo -e "${RED}❌ History endpoint not working correctly${NC}"
  echo "Response: $(cat /tmp/history.json)"
fi

# Test changes endpoint
echo -e "${BLUE}Testing changes endpoint with fixed implementation...${NC}"
curl -k -s -X GET "https://localhost:8085/api/v1/test/temporal/changes-test" \
  -H "Authorization: Bearer $TOKEN" > /tmp/changes.json

if grep -q "timestamp" /tmp/changes.json; then
  echo -e "${GREEN}✅ Changes endpoint working correctly${NC}"
else
  echo -e "${RED}❌ Changes endpoint not working correctly${NC}"
  echo "Response: $(cat /tmp/changes.json)"
fi

# Test diff endpoint
echo -e "${BLUE}Testing diff endpoint with fixed implementation...${NC}"
curl -k -s -X GET "https://localhost:8085/api/v1/test/temporal/diff-test?id=$ENTITY_ID&t1=$CREATED_AT&t2=$UPDATED_AT" \
  -H "Authorization: Bearer $TOKEN" > /tmp/diff.json

if grep -q "added_tags\|removed_tags" /tmp/diff.json; then
  echo -e "${GREEN}✅ Diff endpoint working correctly${NC}"
else
  echo -e "${RED}❌ Diff endpoint not working correctly${NC}"
  echo "Response: $(cat /tmp/diff.json)"
fi

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB Temporal Fix V2 Summary${NC}"
echo -e "${BLUE}========================================${NC}"

SUCCESS=1
if ! grep -q "version:1" /tmp/asof.json; then
  SUCCESS=0
fi
if [ "$HISTORY_COUNT" -eq 0 ]; then
  SUCCESS=0
fi
if ! grep -q "timestamp" /tmp/changes.json; then
  SUCCESS=0
fi
if ! grep -q "added_tags\|removed_tags" /tmp/diff.json; then
  SUCCESS=0
fi

if [ "$SUCCESS" -eq 1 ]; then
  echo -e "${GREEN}✅ Temporal fix applied and working correctly!${NC}"
else
  echo -e "${RED}❌ Temporal fix applied but some features are not working correctly${NC}"
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Test Entity ID: $ENTITY_ID${NC}"
echo -e "${BLUE}CreatedAt: $CREATED_AT${NC}"
echo -e "${BLUE}UpdatedAt: $UPDATED_AT${NC}"
echo -e "${BLUE}========================================${NC}"