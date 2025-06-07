#!/bin/bash
# Test configuration priority system

echo "=== Testing Configuration Priority System ==="
echo ""

# Test 1: Environment variable only
echo "Test 1: Environment variable only (ENTITYDB_PORT=9090)"
ENTITYDB_PORT=9090 timeout 3 ./bin/entitydb 2>&1 | grep -E "HTTP|8085|9090" | head -5
echo ""

# Test 2: Command line flag only  
echo "Test 2: Command line flag only (--entitydb-port 8888)"
timeout 3 ./bin/entitydb --entitydb-port 8888 2>&1 | grep -E "HTTP|8085|8888" | head -5
echo ""

# Test 3: Both env and flag (flag should win)
echo "Test 3: Both env and flag (env=9090, flag=8888, expect 8888)"
ENTITYDB_PORT=9090 timeout 3 ./bin/entitydb --entitydb-port 8888 2>&1 | grep -E "HTTP|8888|9090" | head -5
echo ""

# Test 4: Check help works
echo "Test 4: Help flag"
./bin/entitydb --help 2>&1 | head -5
echo ""

echo "=== Configuration Priority Test Complete ==="