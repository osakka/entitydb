#!/bin/bash
# Comprehensive EntityDB Health Check
# Tests all critical performance metrics

set -e

HOST="${1:-https://localhost:8085}"
TOKEN=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🏥 EntityDB Comprehensive Health Check"
echo "======================================"
echo "Host: $HOST"
echo ""

# Function to measure response time
measure_time() {
    local start=$(date +%s%N)
    "$@" > /dev/null 2>&1
    local end=$(date +%s%N)
    echo $(( (end - start) / 1000000 ))
}

# 1. Basic Health Check
echo -n "1. Health endpoint... "
HEALTH_TIME=$(curl -s -o /dev/null -w "%{time_total}" "$HOST/health")
HEALTH_MS=$(echo "$HEALTH_TIME * 1000" | bc | cut -d. -f1)
if [ $HEALTH_MS -lt 50 ]; then
    echo -e "${GREEN}✓ ${HEALTH_MS}ms${NC}"
else
    echo -e "${RED}✗ ${HEALTH_MS}ms (>50ms)${NC}"
fi

# 2. Goroutine Count
echo -n "2. Goroutine count... "
GOROUTINES=$(curl -s "$HOST/health" | jq -r '.memory.goroutines // 0')
if [ $GOROUTINES -lt 500 ]; then
    echo -e "${GREEN}✓ $GOROUTINES${NC}"
else
    echo -e "${RED}✗ $GOROUTINES (>500 indicates leak)${NC}"
fi

# 3. Memory Usage
echo -n "3. Memory usage... "
MEMORY_MB=$(curl -s "$HOST/health" | jq -r '.memory.alloc_mb // 0')
if [ $(echo "$MEMORY_MB < 200" | bc) -eq 1 ]; then
    echo -e "${GREEN}✓ ${MEMORY_MB}MB${NC}"
else
    echo -e "${YELLOW}⚠ ${MEMORY_MB}MB (>200MB)${NC}"
fi

# 4. Database Size
echo -n "4. Database size... "
DB_SIZE=$(curl -s "$HOST/api/v1/system/metrics" | jq -r '.database.size_mb // 0')
WAL_SIZE=$(curl -s "$HOST/api/v1/system/metrics" | jq -r '.database.wal_size_mb // 0')
echo -e "${GREEN}DB: ${DB_SIZE}MB, WAL: ${WAL_SIZE}MB${NC}"

# 5. Authentication Test
echo -n "5. Authentication... "
AUTH_TIME=$(curl -s -o /dev/null -w "%{time_total}" -X POST "$HOST/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')
AUTH_MS=$(echo "$AUTH_TIME * 1000" | bc | cut -d. -f1)
TOKEN=$(curl -s -X POST "$HOST/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}' | jq -r '.token')

if [ $AUTH_MS -lt 100 ] && [ "$TOKEN" != "null" ]; then
    echo -e "${GREEN}✓ ${AUTH_MS}ms${NC}"
else
    echo -e "${RED}✗ ${AUTH_MS}ms${NC}"
fi

# 6. Entity Operations
echo -n "6. Entity list... "
LIST_TIME=$(curl -s -o /dev/null -w "%{time_total}" \
    -H "Authorization: Bearer $TOKEN" \
    "$HOST/api/v1/entities/list?limit=10")
LIST_MS=$(echo "$LIST_TIME * 1000" | bc | cut -d. -f1)
if [ $LIST_MS -lt 50 ]; then
    echo -e "${GREEN}✓ ${LIST_MS}ms${NC}"
else
    echo -e "${RED}✗ ${LIST_MS}ms (>50ms)${NC}"
fi

# 7. Concurrent Request Test
echo -n "7. Concurrent requests (10)... "
CONCURRENT_START=$(date +%s%N)
for i in {1..10}; do
    curl -s "$HOST/health" > /dev/null &
done
wait
CONCURRENT_END=$(date +%s%N)
CONCURRENT_MS=$(( (CONCURRENT_END - CONCURRENT_START) / 1000000 ))
if [ $CONCURRENT_MS -lt 100 ]; then
    echo -e "${GREEN}✓ ${CONCURRENT_MS}ms total${NC}"
else
    echo -e "${YELLOW}⚠ ${CONCURRENT_MS}ms total${NC}"
fi

# 8. Metrics Endpoint
echo -n "8. Metrics endpoint... "
METRICS=$(curl -s "$HOST/metrics" | grep -c "entitydb_" || echo "0")
if [ $METRICS -gt 0 ]; then
    echo -e "${GREEN}✓ $METRICS metrics${NC}"
else
    echo -e "${RED}✗ No metrics found${NC}"
fi

# 9. Error Rate Check
echo -n "9. Recent errors... "
ERRORS=$(curl -s -H "Authorization: Bearer $TOKEN" \
    "$HOST/api/v1/system/metrics" | jq -r '.activity.errors_last_hour // 0')
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✓ No errors${NC}"
else
    echo -e "${YELLOW}⚠ $ERRORS errors in last hour${NC}"
fi

# 10. WAL Checkpoint Status
echo -n "10. WAL status... "
WAL_SIZE_MB=$(curl -s "$HOST/api/v1/system/metrics" | jq -r '.database.wal_size_mb // 0')
if [ $(echo "$WAL_SIZE_MB < 50" | bc) -eq 1 ]; then
    echo -e "${GREEN}✓ ${WAL_SIZE_MB}MB${NC}"
else
    echo -e "${YELLOW}⚠ ${WAL_SIZE_MB}MB (needs checkpoint)${NC}"
fi

echo ""
echo "======================================"
echo "Health Check Complete"

# Summary
if [ $GOROUTINES -gt 500 ] || [ $LIST_MS -gt 50 ] || [ $(echo "$WAL_SIZE_MB > 50" | bc) -eq 1 ]; then
    echo -e "${RED}⚠️  CRITICAL ISSUES DETECTED${NC}"
    echo "Recommended actions:"
    [ $GOROUTINES -gt 500 ] && echo "  - Restart server (goroutine leak)"
    [ $LIST_MS -gt 50 ] && echo "  - Check entity query performance"
    [ $(echo "$WAL_SIZE_MB > 50" | bc) -eq 1 ] && echo "  - Force WAL checkpoint"
else
    echo -e "${GREEN}✅ All systems operational${NC}"
fi