#!/bin/bash
# Performance test for high performance mode - uses test endpoints to avoid auth

# ANSI color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}EntityDB High Performance Mode Test${NC}"
echo -e "${BLUE}========================================${NC}"

# Test parameters
NUM_ENTITIES=100  # Number of entities to create
CONTENT_SIZE=500  # Size of content in bytes per entity
NUM_REQUESTS=10   # Number of concurrent requests
ITERATIONS=5      # Number of test iterations

# Function to generate random data
generate_random_data() {
    local size=$1
    head -c $size /dev/urandom | base64
}

# Function to time operations
time_operation() {
    local start_time=$(date +%s.%N)
    eval "$1"
    local end_time=$(date +%s.%N)
    echo $(echo "$end_time - $start_time" | bc)
}

# Test 1: Create entities
echo -e "${BLUE}Test 1: Create $NUM_ENTITIES entities with $CONTENT_SIZE bytes each${NC}"
create_times=()

for i in $(seq 1 $ITERATIONS); do
    echo -e "${BLUE}Iteration $i/${ITERATIONS}...${NC}"
    
    # Measure time to create entities
    time_taken=$(time_operation "
        for j in \$(seq 1 $NUM_ENTITIES); do
            curl -s -o /dev/null -X POST 'http://localhost:8085/api/v1/test/entities/create' \
                -H 'Content-Type: application/json' \
                -d '{\"tags\": [\"type:test\", \"test:performance\", \"iteration:$i\"], 
                     \"content\": \"'$(generate_random_data $CONTENT_SIZE)'\"}'
        done
    ")
    
    create_times+=($time_taken)
    echo -e "${GREEN}Created $NUM_ENTITIES entities in $time_taken seconds${NC}"
    
    # Sleep briefly to avoid overwhelming the server
    sleep 1
done

# Calculate average creation time
total_create_time=0
for t in "${create_times[@]}"; do
    total_create_time=$(echo "$total_create_time + $t" | bc)
done
avg_create_time=$(echo "scale=4; $total_create_time / $ITERATIONS" | bc)
entities_per_sec=$(echo "scale=2; $NUM_ENTITIES / $avg_create_time" | bc)

echo -e "${GREEN}Average time to create $NUM_ENTITIES entities: $avg_create_time seconds${NC}"
echo -e "${GREEN}Entities created per second: $entities_per_sec${NC}"

# Test 2: Query entities
echo -e "${BLUE}Test 2: Query entities by tag${NC}"
query_times=()

for i in $(seq 1 $ITERATIONS); do
    echo -e "${BLUE}Iteration $i/${ITERATIONS}...${NC}"
    
    # Measure time to query entities
    time_taken=$(time_operation "
        curl -s -o /dev/null 'http://localhost:8085/api/v1/test/entities/list?tag=test:performance'
    ")
    
    query_times+=($time_taken)
    echo -e "${GREEN}Queried entities in $time_taken seconds${NC}"
    
    # Sleep briefly
    sleep 1
done

# Calculate average query time
total_query_time=0
for t in "${query_times[@]}"; do
    total_query_time=$(echo "$total_query_time + $t" | bc)
done
avg_query_time=$(echo "scale=4; $total_query_time / $ITERATIONS" | bc)

echo -e "${GREEN}Average time to query entities: $avg_query_time seconds${NC}"

# Test 3: Concurrent access
echo -e "${BLUE}Test 3: Concurrent access with $NUM_REQUESTS parallel requests${NC}"
concurrent_times=()

for i in $(seq 1 $ITERATIONS); do
    echo -e "${BLUE}Iteration $i/${ITERATIONS}...${NC}"
    
    # Generate entity IDs for testing
    entity_ids=()
    for j in $(seq 1 $NUM_REQUESTS); do
        response=$(curl -s -X POST 'http://localhost:8085/api/v1/test/entities/create' \
            -H 'Content-Type: application/json' \
            -d "{\"tags\": [\"type:test\", \"test:concurrent\", \"iteration:$i\"], 
                 \"content\": \"$(generate_random_data $CONTENT_SIZE)\"}")
        
        # Extract ID from response
        id=$(echo $response | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
        if [ ! -z "$id" ]; then
            entity_ids+=($id)
        fi
    done
    
    # Measure time for concurrent access
    time_taken=$(time_operation "
        pids=()
        for id in ${entity_ids[@]}; do
            curl -s -o /dev/null 'http://localhost:8085/api/v1/test/entities/get?id='$id &
            pids+=(\$!)
        done
        
        # Wait for all requests to complete
        for pid in \${pids[@]}; do
            wait \$pid
        done
    ")
    
    concurrent_times+=($time_taken)
    echo -e "${GREEN}Completed $NUM_REQUESTS concurrent requests in $time_taken seconds${NC}"
    
    # Sleep briefly
    sleep 1
done

# Calculate average concurrent access time
total_concurrent_time=0
for t in "${concurrent_times[@]}"; do
    total_concurrent_time=$(echo "$total_concurrent_time + $t" | bc)
done
avg_concurrent_time=$(echo "scale=4; $total_concurrent_time / $ITERATIONS" | bc)
requests_per_sec=$(echo "scale=2; $NUM_REQUESTS / $avg_concurrent_time" | bc)

echo -e "${GREEN}Average time for $NUM_REQUESTS concurrent requests: $avg_concurrent_time seconds${NC}"
echo -e "${GREEN}Requests handled per second: $requests_per_sec${NC}"

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Performance Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Entity creation: $entities_per_sec entities/second${NC}"
echo -e "${GREEN}Query response: $avg_query_time seconds${NC}"
echo -e "${GREEN}Concurrent handling: $requests_per_sec requests/second${NC}"
echo -e "${BLUE}========================================${NC}"

exit 0