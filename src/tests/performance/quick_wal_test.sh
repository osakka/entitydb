#!/bin/bash
# Quick WAL-only performance demonstration

cd "$(dirname "$0")/../../.."

echo "=== Quick Performance Comparison ==="
echo

# Kill any existing server
pkill -f entitydb 2>/dev/null
sleep 1

# Test function
test_writes() {
    local mode=$1
    local env=$2
    
    # Clean start
    rm -rf var/quick_test
    mkdir -p var/quick_test
    
    # Start server
    export ENTITYDB_DATA_PATH=var/quick_test
    export $env
    ./bin/entitydb server > /tmp/quick_test.log 2>&1 &
    PID=$!
    sleep 2
    
    # Create admin
    curl -s -X POST http://localhost:8085/api/v1/users/create \
         -H "Content-Type: application/json" \
         -d '{"username":"admin","password":"admin"}' > /dev/null
    
    # Login
    TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
         -H "Content-Type: application/json" \
         -d '{"username":"admin","password":"admin"}' | jq -r '.token')
    
    echo "Testing 20 writes in $mode mode..."
    
    # Time 20 writes
    START=$(date +%s.%N)
    for i in {1..20}; do
        curl -s -X POST http://localhost:8085/api/v1/entities/create \
             -H "Authorization: Bearer $TOKEN" \
             -H "Content-Type: application/json" \
             -d "{\"id\":\"test-$i\",\"tags\":[\"test\"],\"content\":\"Entity $i\"}" > /dev/null
        echo -n "."
    done
    echo
    
    END=$(date +%s.%N)
    TOTAL=$(echo "$END - $START" | bc)
    AVG=$(echo "scale=1; $TOTAL / 20 * 1000" | bc)
    
    echo "$mode: ${TOTAL}s total, ${AVG}ms per write"
    
    # Cleanup
    kill $PID 2>/dev/null
    wait $PID 2>/dev/null
    unset $env
}

# Standard mode
test_writes "STANDARD" "ENTITYDB_DISABLE_HIGH_PERFORMANCE=true"

echo

# WAL-only mode  
test_writes "WAL-ONLY" "ENTITYDB_WAL_ONLY=true"

echo
echo "Expected: WAL-only should be significantly faster (constant time vs increasing time)"