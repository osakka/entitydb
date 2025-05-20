#!/bin/bash
# EntityDB Simple Test Framework
# A minimalist test framework for API testing with request/response pairs

set -e # Exit on error

# Configuration
API_BASE_URL=${API_BASE_URL:-"http://localhost:8085/api/v1"}
TEST_DIR=${TEST_DIR:-"/opt/entitydb/share/tests/test_cases"}
TEMP_DIR=${TEMP_DIR:-"/tmp/entitydb_tests"}
COLOR_OUTPUT=${COLOR_OUTPUT:-true}

# Colors for output
if [[ "$COLOR_OUTPUT" == "true" ]]; then
  GREEN='\033[0;32m'
  RED='\033[0;31m'
  YELLOW='\033[0;33m'
  BLUE='\033[0;34m'
  NC='\033[0m' # No Color
else
  GREEN=''
  RED=''
  YELLOW=''
  BLUE=''
  NC=''
fi

# Create temp directory
mkdir -p "$TEMP_DIR"

# Test session token storage
SESSION_TOKEN=""

# Test counters
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# Print header
print_header() {
  echo -e "\n${BLUE}=======================================${NC}"
  echo -e "${BLUE}   EntityDB API Test Framework v1.0    ${NC}"
  echo -e "${BLUE}=======================================${NC}\n"
}

# Print result
print_result() {
  echo -e "\n${BLUE}=======================================${NC}"
  echo -e "${BLUE}   Test Results: $TESTS_PASSED/$TESTS_TOTAL Passed    ${NC}"
  
  if [[ $TESTS_FAILED -gt 0 ]]; then
    echo -e "${RED}   Tests Failed: $TESTS_FAILED    ${NC}"
    echo -e "${BLUE}=======================================${NC}"
    return 1
  else
    echo -e "${GREEN}   All Tests Passed!    ${NC}"
    echo -e "${BLUE}=======================================${NC}"
    return 0
  fi
}

# Initialize (optional cleanup)
initialize() {
  if [[ "$1" == "clean" ]]; then
    echo -e "${YELLOW}Stopping server and cleaning database...${NC}"
    cd /opt/entitydb
    ./bin/entitydbd.sh stop
    sleep 2
    rm -f var/*.ebf var/*.wal var/*.log
    ./bin/entitydbd.sh start
    sleep 3
    echo -e "${GREEN}Server restarted with clean database${NC}"
  fi
  
  # Create test directory if it doesn't exist
  mkdir -p "$TEST_DIR"
}

# Login and store token (for authenticated tests)
login() {
  local username="${1:-admin}"
  local password="${2:-admin}"
  
  echo -e "${YELLOW}Logging in as $username...${NC}"
  
  local response
  response=$(curl -s -X POST "$API_BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$username\",\"password\":\"$password\"}")
  
  # Extract token from response
  SESSION_TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | sed 's/"token":"//')
  
  if [[ -z "$SESSION_TOKEN" ]]; then
    echo -e "${RED}Login failed. Response: $response${NC}"
    return 1
  else
    echo -e "${GREEN}Login successful. Token obtained.${NC}"
    return 0
  fi
}

# Run a single test using request/response files
run_test() {
  local test_name="$1"
  local test_description="${2:-$test_name}"
  
  echo -e "\n${YELLOW}Running test: ${test_description}${NC}"
  
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  
  # Check if request file exists
  if [[ ! -f "$TEST_DIR/${test_name}_request" ]]; then
    echo -e "${RED}Test request file not found: $TEST_DIR/${test_name}_request${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    return 1
  fi
  
  # Check if response criteria file exists
  if [[ ! -f "$TEST_DIR/${test_name}_response" ]]; then
    echo -e "${RED}Test response criteria file not found: $TEST_DIR/${test_name}_response${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    return 1
  fi
  
  # Load test request details
  source "$TEST_DIR/${test_name}_request"
  
  # Set default values if not provided
  METHOD=${METHOD:-"GET"}
  ENDPOINT=${ENDPOINT:-""}
  HEADERS=${HEADERS:-""}
  DATA=${DATA:-""}
  QUERY=${QUERY:-""}
  
  # Add authorization header if logged in
  if [[ -n "$SESSION_TOKEN" && -z $(echo "$HEADERS" | grep "Authorization") ]]; then
    if [[ -n "$HEADERS" ]]; then
      HEADERS="$HEADERS -H \"Authorization: Bearer $SESSION_TOKEN\""
    else
      HEADERS="-H \"Authorization: Bearer $SESSION_TOKEN\""
    fi
  fi
  
  # Build URL with query parameters
  local url="$API_BASE_URL/$ENDPOINT"
  if [[ -n "$QUERY" ]]; then
    url="${url}?${QUERY}"
  fi
  
  # Build curl command
  local curl_cmd="curl -s -X $METHOD \"$url\" $HEADERS"
  if [[ -n "$DATA" && "$METHOD" != "GET" ]]; then
    curl_cmd="$curl_cmd -d '$DATA'"
  fi
  
  # Execute request
  echo "Executing: $curl_cmd"
  local response
  response=$(eval $curl_cmd)
  
  # Save response to temp file
  local resp_file="$TEMP_DIR/${test_name}_actual_response.json"
  echo "$response" > "$resp_file"
  
  # Evaluate response against criteria
  source "$TEST_DIR/${test_name}_response"
  
  # Default validation function if not provided in the response file
  if [[ "$(type -t validate_response)" != "function" ]]; then
    validate_response() {
      local resp="$1"
      
      # Default validation: check if response contains SUCCESS_MARKER
      if [[ -n "$SUCCESS_MARKER" && "$resp" == *"$SUCCESS_MARKER"* ]]; then
        return 0
      # Or check that it doesn't contain ERROR_MARKER
      elif [[ -n "$ERROR_MARKER" && "$resp" != *"$ERROR_MARKER"* ]]; then
        return 0
      # Or check HTTP status (stored in HTTP_STATUS variable)
      elif [[ -n "$EXPECTED_STATUS" && "$HTTP_STATUS" == "$EXPECTED_STATUS" ]]; then
        return 0
      else
        # If no criteria specified, assume success
        if [[ -z "$SUCCESS_MARKER" && -z "$ERROR_MARKER" && -z "$EXPECTED_STATUS" ]]; then
          return 0
        fi
        return 1
      fi
    }
  fi
  
  # Validate the response
  if validate_response "$response"; then
    echo -e "${GREEN}✓ Test passed: $test_description${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    return 0
  else
    echo -e "${RED}✗ Test failed: $test_description${NC}"
    echo -e "${RED}Response: $response${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    return 1
  fi
}

# Run all tests in a directory
run_all_tests() {
  local test_dir="${1:-$TEST_DIR}"
  
  echo -e "${YELLOW}Running all tests in $test_dir...${NC}"
  
  # Find all test request files
  local test_files=$(find "$test_dir" -name "*_request" -type f | sort)
  
  # Check if we found any tests
  if [[ -z "$test_files" ]]; then
    echo -e "${RED}No test files found in $test_dir${NC}"
    return 1
  fi
  
  # Run each test
  for test_file in $test_files; do
    local test_name=$(basename "$test_file" _request)
    run_test "$test_name" "$(grep "# Description:" "$test_file" | sed 's/# Description: //')"
  done
  
  # Print results
  print_result
  
  # Return failure if any tests failed
  if [[ $TESTS_FAILED -gt 0 ]]; then
    return 1
  else
    return 0
  fi
}

# Get entity by ID helper function
get_entity() {
  local entity_id="$1"
  
  if [[ -z "$entity_id" ]]; then
    echo "Error: Entity ID required"
    return 1
  fi
  
  local response
  response=$(curl -s -X GET "$API_BASE_URL/entities/get?id=$entity_id" \
    -H "Authorization: Bearer $SESSION_TOKEN")
  
  echo "$response"
}

# Create entity helper function
create_entity() {
  local tags="$1"
  local content="$2"
  
  local data="{\"tags\":$tags"
  if [[ -n "$content" ]]; then
    data="$data,\"content\":$content"
  fi
  data="$data}"
  
  local response
  response=$(curl -s -X POST "$API_BASE_URL/entities/create" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $SESSION_TOKEN" \
    -d "$data")
  
  echo "$response"
}

# Usage information
show_usage() {
  echo "EntityDB Test Framework"
  echo ""
  echo "Usage: $0 [options] [test_name]"
  echo ""
  echo "Options:"
  echo "  -h, --help        Show this help message"
  echo "  -c, --clean       Clean database before testing"
  echo "  -a, --all         Run all tests"
  echo "  -d, --dir DIR     Specify test directory (default: $TEST_DIR)"
  echo "  -l, --login       Perform login before tests"
  echo ""
  echo "Examples:"
  echo "  $0 --clean --all               Clean DB and run all tests"
  echo "  $0 --login test_entity_create  Login and run specific test"
}

# Main function
main() {
  print_header
  
  # Parse command line arguments
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        show_usage
        exit 0
        ;;
      -c|--clean)
        initialize "clean"
        shift
        ;;
      -a|--all)
        shift
        RUN_ALL=true
        ;;
      -d|--dir)
        shift
        TEST_DIR="$1"
        shift
        ;;
      -l|--login)
        login
        shift
        ;;
      *)
        if [[ $1 == -* ]]; then
          echo "Unknown option: $1"
          show_usage
          exit 1
        else
          TEST_NAME="$1"
          shift
        fi
        ;;
    esac
  done
  
  # Run tests
  if [[ "$RUN_ALL" == "true" ]]; then
    run_all_tests
  elif [[ -n "$TEST_NAME" ]]; then
    run_test "$TEST_NAME"
    print_result
  else
    show_usage
    exit 1
  fi
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi