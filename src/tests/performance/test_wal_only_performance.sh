#!/bin/bash
# Test WAL-only mode performance vs standard mode

cd "$(dirname "$0")/../../.."

echo "=== EntityDB Write Performance Test ==="
echo

# Function to test write performance
test_write_performance() {
    local mode=$1
    local env_var=$2
    local count=$3
    
    # Clean up
    rm -rf var/test_performance
    mkdir -p var/test_performance
    
    # Start server with specific mode
    export $env_var
    export ENTITYDB_DATA_PATH=var/test_performance
    
    echo "Starting server in $mode mode..."
    ./bin/entitydb server > /tmp/entitydb_perf.log 2>&1 &
    SERVER_PID=$!
    sleep 3
    
    # Create admin user
    echo '{"username":"admin","password":"admin"}' | \
    curl -s -X POST http://localhost:8085/api/v1/users/create \
         -H "Content-Type: application/json" -d @- > /dev/null
    
    # Login
    TOKEN=$(echo '{"username":"admin","password":"admin"}' | \
    curl -s -X POST http://localhost:8085/api/v1/auth/login \
         -H "Content-Type: application/json" -d @- | jq -r '.token')
    
    echo "Testing $count entity writes in $mode mode..."
    
    # Time the writes
    START_TIME=$(date +%s.%N)
    
    for i in $(seq 1 $count); do
        echo "{
            \"id\": \"perf-test-$i\",
            \"tags\": [\"type:test\", \"mode:$mode\", \"index:$i\"],
            \"content\": \"Performance test entity $i with some content to make it realistic\"
        }" | curl -s -X POST http://localhost:8085/api/v1/entities/create \
              -H "Authorization: Bearer $TOKEN" \
              -H "Content-Type: application/json" -d @- > /dev/null
        
        # Show progress
        if [ $((i % 10)) -eq 0 ]; then
            echo -n "."
        fi
    done
    echo
    
    END_TIME=$(date +%s.%N)
    DURATION=$(echo "$END_TIME - $START_TIME" | bc)
    AVG_TIME=$(echo "scale=3; $DURATION / $count * 1000" | bc)
    
    echo "Completed $count writes in ${DURATION}s"
    echo "Average write time: ${AVG_TIME}ms per entity"
    echo
    
    # Stop server
    kill $SERVER_PID 2>/dev/null
    wait $SERVER_PID 2>/dev/null
    unset $env_var
}

# Build latest version
echo "Building EntityDB..."
cd src && make clean && make
cd ..

# Test standard mode
test_write_performance "STANDARD" "ENTITYDB_DISABLE_HIGH_PERFORMANCE=true" 100

# Test WAL-only mode
test_write_performance "WAL-ONLY" "ENTITYDB_WAL_ONLY=true" 100

echo "=== Performance Comparison Complete ==="
echo
echo "Expected results:"
echo "- Standard mode: Write time increases with each entity (O(n))"
echo "- WAL-only mode: Constant write time per entity (O(1))"