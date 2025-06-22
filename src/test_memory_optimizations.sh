#!/bin/bash

# Memory Optimization Test Suite
# Tests all memory optimizations implemented to prevent high memory usage

set -e

echo "🧪 EntityDB Memory Optimization Test Suite"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0

# Function to run a test
run_test() {
    local test_name=$1
    local test_package=$2
    local test_function=$3
    
    echo -e "${YELLOW}Running: ${test_name}${NC}"
    
    if go test -v -timeout 10m ${test_package} -run "^${test_function}$" 2>&1 | tee test_output.tmp; then
        echo -e "${GREEN}✅ PASSED: ${test_name}${NC}"
        ((PASSED++))
    else
        echo -e "${RED}❌ FAILED: ${test_name}${NC}"
        ((FAILED++))
        cat test_output.tmp
    fi
    
    echo ""
    rm -f test_output.tmp
}

# Function to run benchmark
run_benchmark() {
    local bench_name=$1
    local bench_package=$2
    local bench_function=$3
    
    echo -e "${YELLOW}Benchmarking: ${bench_name}${NC}"
    
    go test -bench "^${bench_function}$" -benchmem -benchtime=10s ${bench_package} | tee bench_output.tmp
    
    echo ""
}

# Change to src directory
cd /opt/entitydb/src

echo "1️⃣ Testing Bounded String Interning"
echo "-----------------------------------"
run_test "String Interning with LRU" "./storage/binary" "TestBoundedStringInterning"

echo "2️⃣ Testing Bounded Entity Cache"
echo "-------------------------------"
run_test "Entity Cache with Memory Limits" "./storage/binary" "TestBoundedEntityCache"

echo "3️⃣ Testing Metrics Recursion Prevention"
echo "--------------------------------------"
run_test "Metrics Recursion Control" "./storage/binary" "TestMetricsRecursionPrevention"

echo "4️⃣ Testing Memory Monitor"
echo "------------------------"
run_test "Memory Monitoring System" "./storage/binary" "TestMemoryMonitor"

echo "5️⃣ Testing Temporal Retention"
echo "----------------------------"
run_test "Temporal Data Cleanup" "./storage/binary" "TestTemporalRetention"

echo "6️⃣ Testing Integrated Optimizations"
echo "----------------------------------"
run_test "All Optimizations Together" "./storage/binary" "TestIntegratedMemoryOptimization"

echo "7️⃣ Testing Concurrent Safety"
echo "---------------------------"
run_test "Thread Safety" "./storage/binary" "TestConcurrentMemoryOptimization"

echo "8️⃣ Memory Leak Detection (Quick)"
echo "-------------------------------"
# Run with -short flag for quick leak detection
go test -v -short -timeout 2m ./storage/binary -run "^TestMemoryLeakDetection$" 2>&1 | tee test_output.tmp
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ PASSED: Memory Leak Detection${NC}"
    ((PASSED++))
else
    echo -e "${RED}❌ FAILED: Memory Leak Detection${NC}"
    ((FAILED++))
fi
rm -f test_output.tmp

echo ""
echo "9️⃣ Performance Benchmarks"
echo "------------------------"
run_benchmark "Memory Optimizations" "./storage/binary" "BenchmarkMemoryOptimizations"

echo ""
echo "🔟 Memory Stress Test (Extended)"
echo "------------------------------"
echo "⚠️  This test simulates production load and takes ~2 minutes"
echo "Press Ctrl+C to skip, or wait 5 seconds to continue..."
sleep 5

run_test "Production Memory Stress" "./storage/binary" "TestMemoryStressScenario"

echo ""
echo "========================================="
echo "📊 Test Summary"
echo "========================================="
echo -e "Tests Passed: ${GREEN}${PASSED}${NC}"
echo -e "Tests Failed: ${RED}${FAILED}${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All memory optimization tests passed!${NC}"
    echo ""
    echo "✅ String interning with bounded LRU cache"
    echo "✅ Entity cache with memory limits"
    echo "✅ Metrics recursion prevention"
    echo "✅ Memory pressure monitoring"
    echo "✅ Temporal data retention"
    echo "✅ Thread-safe implementations"
    echo "✅ No memory leaks detected"
    echo "✅ Production stress test passed"
    echo ""
    echo "The memory optimizations are working correctly and should prevent"
    echo "the high memory utilization issues that previously caused crashes."
    exit 0
else
    echo -e "${RED}⚠️  Some tests failed. Please review the output above.${NC}"
    exit 1
fi