#!/bin/bash

# Performance test for EntityDB optimizations

set -e

BASE_URL="https://localhost:8085"
ADMIN_USER="admin"
ADMIN_PASS="admin"
NUM_ENTITIES=1000
NUM_TAGS=20

echo "EntityDB Optimization Performance Test"
echo "====================================="

# Function to login and get token
login() {
    local token=$(curl -k -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$ADMIN_USER\",\"password\":\"$ADMIN_PASS\"}" | \
        jq -r '.token')
    
    if [ "$token" == "null" ] || [ -z "$token" ]; then
        echo "Failed to login"
        exit 1
    fi
    
    echo "$token"
}

# Function to create test entities
create_test_entities() {
    local token=$1
    local start_time=$(date +%s%N)
    
    echo "Creating $NUM_ENTITIES test entities..."
    
    for i in $(seq 1 $NUM_ENTITIES); do
        # Create entity with many tags to test string interning
        local tags='['
        for j in $(seq 1 $NUM_TAGS); do
            # Use repeating tags to test interning
            local tag_type=$((j % 5))
            case $tag_type in
                0) tags="${tags}\"type:document\"," ;;
                1) tags="${tags}\"status:active\"," ;;
                2) tags="${tags}\"priority:high\"," ;;
                3) tags="${tags}\"rbac:role:user\"," ;;
                4) tags="${tags}\"category:test\"," ;;
            esac
        done
        tags="${tags%,}]"
        
        # Create entity with some content
        curl -k -s -X POST "$BASE_URL/api/v1/entities/create" \
            -H "Authorization: Bearer $token" \
            -H "Content-Type: application/json" \
            -d "{
                \"tags\": $tags,
                \"content\": \"Test entity $i with optimization testing content that should benefit from buffer pooling\"
            }" > /dev/null
        
        if [ $((i % 100)) -eq 0 ]; then
            echo "Created $i entities..."
        fi
    done
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))
    echo "Created $NUM_ENTITIES entities in ${duration}ms"
    echo "Average: $((duration / NUM_ENTITIES))ms per entity"
}

# Function to test read performance
test_read_performance() {
    local token=$1
    local start_time=$(date +%s%N)
    
    echo -e "\nTesting read performance..."
    
    # List all entities
    local response=$(curl -k -s -X GET "$BASE_URL/api/v1/entities/list" \
        -H "Authorization: Bearer $token")
    
    local count=$(echo "$response" | jq -r '. | length')
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))
    
    echo "Listed $count entities in ${duration}ms"
}

# Function to test concurrent reads
test_concurrent_reads() {
    local token=$1
    local concurrency=50
    local requests_per_worker=20
    
    echo -e "\nTesting concurrent read performance..."
    echo "Concurrency: $concurrency workers, $requests_per_worker requests each"
    
    local start_time=$(date +%s%N)
    
    # Launch concurrent workers
    for i in $(seq 1 $concurrency); do
        (
            for j in $(seq 1 $requests_per_worker); do
                curl -k -s -X GET "$BASE_URL/api/v1/entities/list?tag=type:document" \
                    -H "Authorization: Bearer $token" > /dev/null
            done
        ) &
    done
    
    # Wait for all workers
    wait
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))
    local total_requests=$((concurrency * requests_per_worker))
    
    echo "Completed $total_requests requests in ${duration}ms"
    echo "Throughput: $((total_requests * 1000 / duration)) requests/second"
}

# Function to get memory stats
get_memory_stats() {
    local token=$1
    
    echo -e "\nMemory Statistics:"
    
    # Get system metrics
    local metrics=$(curl -k -s -X GET "$BASE_URL/api/v1/system/metrics" \
        -H "Authorization: Bearer $token")
    
    if [ ! -z "$metrics" ]; then
        echo "Memory Allocated: $(echo "$metrics" | jq -r '.memory.alloc_bytes // 0' | numfmt --to=iec-i --suffix=B)"
        echo "Heap In Use: $(echo "$metrics" | jq -r '.memory.heap_alloc_bytes // 0' | numfmt --to=iec-i --suffix=B)"
        echo "Total Alloc: $(echo "$metrics" | jq -r '.memory.total_alloc_bytes // 0' | numfmt --to=iec-i --suffix=B)"
        echo "GC Runs: $(echo "$metrics" | jq -r '.performance.gc_runs // 0')"
        echo "Entity Count: $(echo "$metrics" | jq -r '.database.total_entities // 0')"
        echo "Unique Tags: $(echo "$metrics" | jq -r '.database.tags_unique // 0')"
    fi
}

# Main test flow
main() {
    # Check if server is running
    if ! curl -k -s "$BASE_URL/health" > /dev/null; then
        echo "EntityDB server is not running at $BASE_URL"
        echo "Please start the server first: ./bin/entitydbd.sh start"
        exit 1
    fi
    
    # Login
    echo "Logging in..."
    TOKEN=$(login)
    
    # Get initial memory stats
    echo -e "\nInitial State:"
    get_memory_stats "$TOKEN"
    
    # Create test data
    create_test_entities "$TOKEN"
    
    # Test read performance
    test_read_performance "$TOKEN"
    
    # Test concurrent reads
    test_concurrent_reads "$TOKEN"
    
    # Get final memory stats
    get_memory_stats "$TOKEN"
    
    echo -e "\nOptimization test completed!"
}

# Run the test
main