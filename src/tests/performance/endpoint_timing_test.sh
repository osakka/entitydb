#!/bin/bash
# Test response times for all critical endpoints

HOST="https://localhost:8085"
TOKEN=""

# Get auth token first
TOKEN=$(curl -k -s -X POST "$HOST/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' | jq -r '.token')

echo "📊 EntityDB Endpoint Performance Test"
echo "====================================="
echo ""

# Function to test endpoint
test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local auth="$3"
    local data="$4"
    
    local auth_header=""
    if [ "$auth" = "yes" ]; then
        auth_header="-H \"Authorization: Bearer $TOKEN\""
    fi
    
    local data_opt=""
    if [ ! -z "$data" ]; then
        data_opt="-H \"Content-Type: application/json\" -d '$data'"
    fi
    
    # Run 5 times and calculate average
    local total=0
    local times=""
    
    for i in {1..5}; do
        local cmd="curl -k -s -o /dev/null -w \"%{time_total}\" -X $method $auth_header $data_opt \"$HOST$endpoint\""
        local time=$(eval $cmd)
        local ms=$(echo "$time * 1000" | bc | cut -d. -f1)
        times="$times $ms"
        total=$(echo "$total + $ms" | bc)
    done
    
    local avg=$(echo "$total / 5" | bc)
    
    printf "%-40s %4dms  (samples:%s)\n" "$endpoint" "$avg" "$times"
}

echo "Public Endpoints (no auth):"
echo "---------------------------"
test_endpoint "GET" "/health" "no"
test_endpoint "GET" "/metrics" "no"
test_endpoint "GET" "/api/v1/system/metrics" "no"

echo ""
echo "Auth Endpoints:"
echo "---------------"
test_endpoint "POST" "/api/v1/auth/login" "no" '{"username":"admin","password":"admin"}'

echo ""
echo "Entity Endpoints (with auth):"
echo "-----------------------------"
test_endpoint "GET" "/api/v1/entities/list?limit=10" "yes"
test_endpoint "GET" "/api/v1/entities/get?id=user_admin" "yes"
test_endpoint "GET" "/api/v1/entities/query?tags=type:user" "yes"

# Create test entity
curl -k -s -X POST "$HOST/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"id":"test_perf_entity","tags":["type:test","perf:test"],"content":"test"}' > /dev/null

test_endpoint "POST" "/api/v1/entities/create" "yes" '{"id":"perf_test_'$RANDOM'","tags":["type:test"],"content":"test"}'
test_endpoint "PUT" "/api/v1/entities/update" "yes" '{"id":"test_perf_entity","tags":["type:test","perf:updated"]}'

echo ""
echo "Temporal Endpoints:"
echo "-------------------"
test_endpoint "GET" "/api/v1/entities/history?id=test_perf_entity" "yes"
test_endpoint "GET" "/api/v1/entities/as-of?timestamp=2025-06-12T12:00:00Z&tags=type:test" "yes"

echo ""
echo "Admin Endpoints:"
echo "----------------"
test_endpoint "GET" "/api/v1/dashboard/stats" "yes"
test_endpoint "GET" "/api/v1/rbac/metrics" "yes"

echo ""
echo "====================================="
echo "✅ Performance test complete"