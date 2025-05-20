#!/bin/bash
#
# Test script for EntityDB security implementation
#

# Ensure /opt/entitydb/var/log/audit directory exists
mkdir -p /opt/entitydb/var/log/audit

# Function to run a test and display results
run_test() {
  local test_name="$1"
  local test_cmd="$2"
  
  echo "========================================="
  echo "Running test: $test_name"
  echo "-----------------------------------------"
  eval "$test_cmd"
  if [ $? -eq 0 ]; then
    echo -e "\n✓ Test passed: $test_name"
  else
    echo -e "\n✗ Test failed: $test_name"
  fi
  echo "========================================="
  echo ""
}

# Build the server with security components
build_server() {
  echo "Building secure server..."
  cd /opt/entitydb/src && go build -o entitydb_server_secure server_db.go security_bridge.go input_validator.go audit_logger.go simple_security.go
  if [ $? -eq 0 ]; then
    echo "Server built successfully: entitydb_server_secure"
    return 0
  else
    echo "Server build failed"
    return 1
  fi
}

# Test secure password handling
test_password_handling() {
  local test_output=$(cd /opt/entitydb/src && go run simple_security.go -test-password password123)
  echo "$test_output"
  if [[ $test_output == *"Password validation successful"* ]]; then
    return 0
  else
    return 1
  fi
}

# Test audit logging
test_audit_logging() {
  # Create test audit log
  mkdir -p /tmp/entitydb_test_audit
  cd /opt/entitydb/src && go run audit_logger.go -test-logging /tmp/entitydb_test_audit
  
  # Check if log file exists
  if [ -f /tmp/entitydb_test_audit/entitydb_audit_*.log ]; then
    echo "Audit log file created successfully"
    cat /tmp/entitydb_test_audit/entitydb_audit_*.log
    return 0
  else
    echo "Failed to create audit log file"
    return 1
  fi
}

# Test input validation
test_input_validation() {
  local test_output=$(cd /opt/entitydb/src && go run input_validator.go -test)
  echo "$test_output"
  if [[ $test_output == *"Input validation tests passed"* ]]; then
    return 0
  else
    return 1
  fi
}

# Main script
echo "EntityDB Security Implementation Tests"
echo "======================================"

# Build the secure server
if build_server; then
  echo "Building security components successful"
  
  # Run individual tests
  run_test "Secure Password Handling" "test_password_handling"
  run_test "Audit Logging" "test_audit_logging"
  run_test "Input Validation" "test_input_validation"
  
  echo "Security tests completed"
else
  echo "Failed to build security components"
  exit 1
fi

exit 0