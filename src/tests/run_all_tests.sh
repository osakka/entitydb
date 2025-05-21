#!/bin/bash
# Master script to run all database integrity tests

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function for colored output
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Run a single test
run_test() {
  local test_script=$1
  local test_name=$2
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Starting $test_name"
  print_message "$BLUE" "========================================"
  
  bash "$test_script"
  local result=$?
  
  print_message "$BLUE" "========================================"
  if [ $result -eq 0 ]; then
    print_message "$GREEN" "✅ $test_name PASSED"
  else
    print_message "$RED" "❌ $test_name FAILED"
  fi
  print_message "$BLUE" "========================================"
  echo ""
  
  return $result
}

# Main execution
print_message "$BLUE" "========================================"
print_message "$BLUE" "ENTITYDB COMPREHENSIVE TEST SUITE"
print_message "$BLUE" "========================================"

# First check if the server is running
if ! curl -k -s -o /dev/null "https://localhost:8085/api/v1/status"; then
  print_message "$RED" "❌ EntityDB server is not running"
  print_message "$YELLOW" "Please start the server with: ./bin/entitydbd.sh start"
  exit 1
fi

# Run each test
print_message "$BLUE" "Running test suite with 5 test modules"
print_message "$BLUE" "========================================"

# Run basic integrity test
run_test "./test_integrity.sh" "Basic Integrity Test"
integrity_result=$?

# Run auto-chunking test
run_test "./test_autochunking.sh" "Auto-chunking Test"
autochunking_result=$?

# Run concurrency test
run_test "./test_concurrency.sh" "Concurrency Test"
concurrency_result=$?

# Run stress test
run_test "./test_stress.sh" "Stress Test"
stress_result=$?

# Run temporal test
run_test "./test_temporal.sh" "Temporal Features Test"
temporal_result=$?

# Final summary
print_message "$BLUE" "========================================"
print_message "$BLUE" "ENTITYDB TEST SUITE SUMMARY"
print_message "$BLUE" "========================================"

if [ $integrity_result -eq 0 ]; then
  print_message "$GREEN" "✅ Basic Integrity Test: PASSED"
else
  print_message "$RED" "❌ Basic Integrity Test: FAILED"
fi

if [ $autochunking_result -eq 0 ]; then
  print_message "$GREEN" "✅ Auto-chunking Test: PASSED"
else
  print_message "$RED" "❌ Auto-chunking Test: FAILED"
fi

if [ $concurrency_result -eq 0 ]; then
  print_message "$GREEN" "✅ Concurrency Test: PASSED"
else
  print_message "$RED" "❌ Concurrency Test: FAILED"
fi

if [ $stress_result -eq 0 ]; then
  print_message "$GREEN" "✅ Stress Test: PASSED"
else
  print_message "$RED" "❌ Stress Test: FAILED"
fi

if [ $temporal_result -eq 0 ]; then
  print_message "$GREEN" "✅ Temporal Features Test: PASSED"
else
  print_message "$RED" "❌ Temporal Features Test: FAILED"
fi

print_message "$BLUE" "========================================"

# Calculate overall result
total_result=$((integrity_result + autochunking_result + concurrency_result + stress_result + temporal_result))

if [ $total_result -eq 0 ]; then
  print_message "$GREEN" "✅ ALL TESTS PASSED! EntityDB database has excellent integrity and stability."
else
  print_message "$RED" "❌ SOME TESTS FAILED! EntityDB database may have integrity or stability issues."
  print_message "$YELLOW" "Please check the individual test logs for more details."
fi

print_message "$BLUE" "========================================"

exit $total_result