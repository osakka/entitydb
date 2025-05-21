#!/bin/bash
# Simple EntityDB Concurrency Test

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Configuration
SERVER_URL="http://localhost:8085"
CONCURRENCY=5          # Number of concurrent clients
OPERATIONS_PER_CLIENT=5  # Number of operations per client
TOKEN=""

# Function for colored output
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Login to get token
login() {
  print_message "$BLUE" "Logging in to EntityDB..."
  
  local response=$(curl -s -X POST "$SERVER_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin"}')
  
  TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
  
  if [ -z "$TOKEN" ]; then
    print_message "$RED" "❌ Failed to login. Response: $response"
    exit 1
  else
    print_message "$GREEN" "✅ Login successful, got token"
  fi
}

# Create an entity
create_entity() {
  local client_id=$1
  local op_id=$2
  
  # Create a short content
  local content="Concurrency test - Client $client_id, Operation $op_id at $(date)"
  
  # Create the entity
  local response=$(curl -s -X POST "$SERVER_URL/api/v1/entities/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"tags\": [\"type:concurrency_test\", \"client:$client_id\", \"op:$op_id\"],
      \"content\": \"$content\"
    }")
  
  # Extract entity ID
  local entity_id=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
  
  if [ -n "$entity_id" ]; then
    echo "$entity_id"
    return 0
  else
    echo ""
    return 1
  fi
}

# Update an entity
update_entity() {
  local entity_id=$1
  local client_id=$2
  local op_id=$3
  
  # Create updated content
  local content="Updated content - Client $client_id, Operation $op_id at $(date)"
  
  # Update the entity
  local response=$(curl -s -X PUT "$SERVER_URL/api/v1/entities/update" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{
      \"id\": \"$entity_id\",
      \"tags\": [\"type:concurrency_test\", \"client:$client_id\", \"op:$op_id\", \"updated:true\"],
      \"content\": \"$content\"
    }")
  
  # Check response
  if [[ "$response" == *"\"id\""* ]]; then
    return 0
  else
    return 1
  fi
}

# Get an entity
get_entity() {
  local entity_id=$1
  
  # Get the entity
  local response=$(curl -s -X GET "$SERVER_URL/api/v1/entities/get?id=$entity_id" \
    -H "Authorization: Bearer $TOKEN")
  
  # Check response
  if [[ "$response" == *"\"id\""* ]]; then
    return 0
  else
    return 1
  fi
}

# Client workflow
client_workflow() {
  local client_id=$1
  local log_file=$2
  
  print_message "$BLUE" "Client $client_id starting..."
  
  local create_success=0
  local create_fail=0
  local update_success=0
  local update_fail=0
  local get_success=0
  local get_fail=0
  
  for i in $(seq 1 $OPERATIONS_PER_CLIENT); do
    # Create entity
    local entity_id=$(create_entity "$client_id" "$i")
    
    if [ -n "$entity_id" ]; then
      echo "✅ Client $client_id, Op $i: Created entity $entity_id" >> "$log_file"
      ((create_success++))
      
      # Get entity
      if get_entity "$entity_id"; then
        echo "✅ Client $client_id, Op $i: Retrieved entity $entity_id" >> "$log_file"
        ((get_success++))
      else
        echo "❌ Client $client_id, Op $i: Failed to retrieve entity $entity_id" >> "$log_file"
        ((get_fail++))
      fi
      
      # Update entity
      if update_entity "$entity_id" "$client_id" "$i"; then
        echo "✅ Client $client_id, Op $i: Updated entity $entity_id" >> "$log_file"
        ((update_success++))
      else
        echo "❌ Client $client_id, Op $i: Failed to update entity $entity_id" >> "$log_file"
        ((update_fail++))
      fi
    else
      echo "❌ Client $client_id, Op $i: Failed to create entity" >> "$log_file"
      ((create_fail++))
    fi
  done
  
  # Write summary to log
  echo "CLIENT $client_id SUMMARY:" >> "$log_file"
  echo "Create: $create_success success, $create_fail fail" >> "$log_file"
  echo "Get: $get_success success, $get_fail fail" >> "$log_file"
  echo "Update: $update_success success, $update_fail fail" >> "$log_file"
  
  print_message "$GREEN" "Client $client_id completed"
}

# Run the concurrent test
run_concurrent_test() {
  local tmp_dir="./tmp_simple_concurrency"
  mkdir -p "$tmp_dir"
  
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Starting concurrent test with $CONCURRENCY clients"
  print_message "$BLUE" "Each client will perform $OPERATIONS_PER_CLIENT operations"
  print_message "$BLUE" "========================================"
  
  # Start clients in background
  for i in $(seq 1 $CONCURRENCY); do
    local log_file="${tmp_dir}/client_${i}.log"
    client_workflow "$i" "$log_file" &
    print_message "$GREEN" "Started client $i"
    sleep 0.2
  done
  
  # Wait for all clients to complete
  wait
  print_message "$GREEN" "All clients completed."
  
  # Aggregate results
  print_message "$BLUE" "========================================"
  print_message "$BLUE" "Aggregating test results..."
  print_message "$BLUE" "========================================"
  
  local total_create_success=0
  local total_create_fail=0
  local total_get_success=0
  local total_get_fail=0
  local total_update_success=0
  local total_update_fail=0
  
  for i in $(seq 1 $CONCURRENCY); do
    local log_file="${tmp_dir}/client_${i}.log"
    
    # Extract statistics
    local create_success=$(grep "Create:" "$log_file" | awk '{print $2}')
    local create_fail=$(grep "Create:" "$log_file" | awk '{print $4}')
    local get_success=$(grep "Get:" "$log_file" | awk '{print $2}')
    local get_fail=$(grep "Get:" "$log_file" | awk '{print $4}')
    local update_success=$(grep "Update:" "$log_file" | awk '{print $2}')
    local update_fail=$(grep "Update:" "$log_file" | awk '{print $4}')
    
    total_create_success=$((total_create_success + create_success))
    total_create_fail=$((total_create_fail + create_fail))
    total_get_success=$((total_get_success + get_success))
    total_get_fail=$((total_get_fail + get_fail))
    total_update_success=$((total_update_success + update_success))
    total_update_fail=$((total_update_fail + update_fail))
  done
  
  # Print summary
  print_message "$BLUE" "CONCURRENCY TEST SUMMARY:"
  print_message "$BLUE" "------------------------"
  print_message "$BLUE" "Create operations: $total_create_success success, $total_create_fail fail"
  print_message "$BLUE" "Get operations: $total_get_success success, $total_get_fail fail"
  print_message "$BLUE" "Update operations: $total_update_success success, $total_update_fail fail"
  
  local total_operations=$((total_create_success + total_create_fail + 
                            total_get_success + total_get_fail + 
                            total_update_success + total_update_fail))
  local total_failures=$((total_create_fail + total_get_fail + total_update_fail))
  local failure_rate=$(awk "BEGIN {printf \"%.2f\", ($total_failures / $total_operations) * 100}")
  
  print_message "$BLUE" "------------------------"
  print_message "$BLUE" "Total operations: $total_operations"
  print_message "$BLUE" "Total failures: $total_failures"
  print_message "$BLUE" "Failure rate: $failure_rate%"
  
  # Determine test result
  if [ $total_failures -eq 0 ]; then
    print_message "$GREEN" "✅ CONCURRENCY TEST PASSED! Perfect performance."
    return 0
  elif [ $(echo "$failure_rate < 5" | bc -l) -eq 1 ]; then
    print_message "$YELLOW" "⚠️ CONCURRENCY TEST ACCEPTABLE - Minor issues ($failure_rate% failure rate)."
    return 0
  else
    print_message "$RED" "❌ CONCURRENCY TEST FAILED with $failure_rate% failure rate! Database has issues under concurrent load."
    return 1
  fi
}

# Main execution
print_message "$BLUE" "========================================"
print_message "$BLUE" "Starting EntityDB Simple Concurrency Test"
print_message "$BLUE" "========================================"

# Login
login

# Run concurrent test
run_concurrent_test
concurrency_result=$?

exit $concurrency_result