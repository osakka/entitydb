#!/bin/bash

# Test binary persistence with journal format

# Source common test utilities
source "$(dirname "$0")/test_utils.sh"

# Setup test environment
setup_test() {
    print_info "Setting up binary persistence test"
    
    # Clean up previous test data
    rm -rf /tmp/entitydb_binary_test
    mkdir -p /tmp/entitydb_binary_test
    
    # Start server with binary storage
    cd /opt/entitydb
    export ENTITYDB_STORAGE="binary"
    export ENTITYDB_DATA_PATH="/tmp/entitydb_binary_test"
    
    # Kill any existing server
    pkill -f "entitydb" || true
    sleep 1
    
    # Start server in background
    ./bin/entitydb > /tmp/entitydb_test.log 2>&1 &
    SERVER_PID=$!
    sleep 2
    
    # Check if server started
    if ! ps -p $SERVER_PID > /dev/null; then
        print_error "Server failed to start"
        cat /tmp/entitydb_test.log
        exit 1
    fi
    
    print_info "Server started with PID $SERVER_PID"
    
    # Login to get token
    export AUTH_TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin"}' | jq -r .token)
    
    if [ -z "$AUTH_TOKEN" ]; then
        print_error "Failed to get auth token"
        exit 1
    fi
    
    print_success "Test setup complete"
}

# Test creating and reading entities
test_create_and_read() {
    print_info "Test 1: Creating entities and reading them back"
    
    # Create first entity
    response=$(curl -s -X POST "$BASE_URL/api/v1/entities" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d @- <<EOF
{
    "type": "binary_test",
    "tags": ["type:test", "stage:first"],
    "content": [
        {"type": "title", "value": "First Test Entity"},
        {"type": "description", "value": "Testing binary persistence"}
    ]
}
EOF
    )
    
    entity1_id=$(echo "$response" | jq -r .id)
    echo "Created entity 1: $entity1_id"
    
    # Create second entity
    response=$(curl -s -X POST "$BASE_URL/api/v1/entities" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d @- <<EOF
{
    "type": "binary_test",
    "tags": ["type:test", "stage:second"],
    "content": [
        {"type": "title", "value": "Second Test Entity"},
        {"type": "description", "value": "Another test entity"}
    ]
}
EOF
    )
    
    entity2_id=$(echo "$response" | jq -r .id)
    echo "Created entity 2: $entity2_id"
    
    # List all entities
    response=$(curl -s -X GET "$BASE_URL/api/v1/entities" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    entity_count=$(echo "$response" | jq -r '.entities | length')
    echo "Found $entity_count entities"
    
    if [ "$entity_count" -lt 2 ]; then
        print_error "Expected at least 2 entities, got $entity_count"
        echo "$response" | jq
        return 1
    fi
    
    # Get specific entity
    response=$(curl -s -X GET "$BASE_URL/api/v1/entities/$entity1_id" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    title=$(echo "$response" | jq -r '.content[] | select(.type=="title") | .value')
    if [ "$title" != "First Test Entity" ]; then
        print_error "Entity content mismatch"
        echo "$response" | jq
        return 1
    fi
    
    print_success "Successfully created and read entities"
}

# Test server restart persistence
test_restart_persistence() {
    print_info "Test 2: Testing persistence across server restart"
    
    # Stop server
    kill $SERVER_PID 2>/dev/null || true
    sleep 2
    
    # Check binary file exists
    if [ ! -f "/tmp/entitydb_binary_test/entities.ebf" ]; then
        print_error "Binary file not created"
        return 1
    fi
    
    file_size=$(stat -f%z "/tmp/entitydb_binary_test/entities.ebf" 2>/dev/null || stat -c%s "/tmp/entitydb_binary_test/entities.ebf")
    echo "Binary file size: $file_size bytes"
    
    # Restart server
    ./bin/entitydb > /tmp/entitydb_test2.log 2>&1 &
    SERVER_PID=$!
    sleep 2
    
    if ! ps -p $SERVER_PID > /dev/null; then
        print_error "Server failed to restart"
        cat /tmp/entitydb_test2.log
        return 1
    fi
    
    # Re-login
    export AUTH_TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin"}' | jq -r .token)
    
    # List entities after restart
    response=$(curl -s -X GET "$BASE_URL/api/v1/entities" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    entity_count=$(echo "$response" | jq -r '.entities | length')
    echo "Found $entity_count entities after restart"
    
    if [ "$entity_count" -lt 2 ]; then
        print_error "Entities not persisted across restart"
        echo "$response" | jq
        return 1
    fi
    
    # Verify content
    response=$(curl -s -X GET "$BASE_URL/api/v1/entities?tag=type:test" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    test_entities=$(echo "$response" | jq -r '.entities | length')
    if [ "$test_entities" -lt 2 ]; then
        print_error "Test entities not found after restart"
        echo "$response" | jq
        return 1
    fi
    
    print_success "Entities persisted across restart"
}

# Test concurrent writes
test_concurrent_writes() {
    print_info "Test 3: Testing concurrent write operations"
    
    # Create multiple entities concurrently
    for i in {1..5}; do
        (
            response=$(curl -s -X POST "$BASE_URL/api/v1/entities" \
                -H "Authorization: Bearer $AUTH_TOKEN" \
                -H "Content-Type: application/json" \
                -d "{\"type\":\"concurrent_test\",\"tags\":[\"type:test\",\"concurrent:$i\"],\"content\":[{\"type\":\"title\",\"value\":\"Concurrent Entity $i\"}]}")
            echo "Created concurrent entity $i: $(echo "$response" | jq -r .id)"
        ) &
    done
    
    # Wait for all background jobs
    wait
    
    # Verify all entities were created
    sleep 1
    response=$(curl -s -X GET "$BASE_URL/api/v1/entities?tag=type:test" \
        -H "Authorization: Bearer $AUTH_TOKEN")
    
    concurrent_count=$(echo "$response" | jq -r '[.entities[] | select(.tags[] | contains("concurrent"))] | length')
    echo "Found $concurrent_count concurrent entities"
    
    if [ "$concurrent_count" -ne 5 ]; then
        print_error "Expected 5 concurrent entities, got $concurrent_count"
        return 1
    fi
    
    print_success "Concurrent writes successful"
}

# Cleanup
cleanup() {
    print_info "Cleaning up"
    kill $SERVER_PID 2>/dev/null || true
    rm -rf /tmp/entitydb_binary_test
    rm -f /tmp/entitydb_test*.log
}

# Run tests
run_test "Binary persistence setup" setup_test
run_test "Create and read entities" test_create_and_read
run_test "Persistence across restart" test_restart_persistence
run_test "Concurrent writes" test_concurrent_writes
cleanup

print_info "Binary persistence tests completed"