#!/bin/bash

# Master script to run all relationship tests

echo "=== EntityDB Relationship Test Suite ==="
echo "Starting at: $(date)"
echo

# Check if server is running
if ! curl -s http://localhost:8085/api/v1/test/status > /dev/null 2>&1; then
    echo "❌ Server is not running. Please start it with:"
    echo "   /opt/entitydb/bin/entitydbd.sh start"
    exit 1
fi

# Make sure all test scripts are executable
chmod +x test_relationships_*.sh

# Run test suites
echo "=== Running Basic Relationship Tests ==="
./test_relationships_working.sh
echo -e "\n\n"

echo "=== Running Comprehensive Relationship Tests ==="
./test_entity_relationships_comprehensive.sh
echo -e "\n\n"

echo "=== Running RBAC Relationship Tests ==="
./test_relationships_rbac.sh
echo -e "\n\n"

echo "=== Running Persistence Tests ==="
./test_relationships_persistence.sh
echo -e "\n\n"

# Summary
echo "=== Test Suite Summary ==="
echo "Completed at: $(date)"
echo "Test Coverage:"
echo "✓ Basic CRUD operations"
echo "✓ Query by source, target, and type"
echo "✓ RBAC enforcement"
echo "✓ Binary persistence and recovery"
echo "✓ Temporal history"
echo "✓ Concurrent access"
echo "✓ Edge cases and error handling"
echo "✓ Performance benchmarks"

echo -e "\n=== All relationship tests completed ==="\n