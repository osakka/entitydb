#!/bin/bash
# EntityDB Simple Test Framework v3.0
# A minimalist test framework for API testing with unified test files and timing metrics

set -e # Exit on error

# Configuration
API_BASE_URL=${API_BASE_URL:-"https://localhost:8085/api/v1"}
TEST_DIR=${TEST_DIR:-"/opt/entitydb/share/tests/test_cases"}
TEMP_DIR=${TEMP_DIR:-"/tmp/entitydb_tests"}
COLOR_OUTPUT=${COLOR_OUTPUT:-true}
INSECURE=${INSECURE:-true}  # Allow self-signed certificates
SHOW_TIMING=${SHOW_TIMING:-true}  # Show timing information for tests

# Colors for output
if [[ "$COLOR_OUTPUT" == "true" ]]; then
  GREEN='\033[0;32m'
  RED='\033[0;31m'
  YELLOW='\033[0;33m'
  BLUE='\033[0;34m'
  CYAN='\033[0;36m'
  NC='\033[0m' # No Color
else
  GREEN=''
  RED=''
  YELLOW=''
  BLUE=''
  CYAN=''
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

# Timing data
TOTAL_TEST_TIME=0
FASTEST_TEST_TIME=9999
SLOWEST_TEST_TIME=0
FASTEST_TEST_NAME=""
SLOWEST_TEST_NAME=""
TEST_TIMES=()

# Print header
print_header() {
  echo -e "\n${BLUE}=======================================${NC}"
  echo -e "${BLUE}   EntityDB API Test Framework v3.0    ${NC}"
  echo -e "${BLUE}=======================================${NC}\n"
}

# Print timing statistics
print_timing_stats() {
  if [[ "$SHOW_TIMING" != "true" || ${#TEST_TIMES[@]} -eq 0 ]]; then
    return 0
  fi

  local avg_time=$(echo "scale=3; $TOTAL_TEST_TIME / $TESTS_TOTAL" | bc)
  
  echo -e "\n${BLUE}=======================================${NC}"
  echo -e "${BLUE}   Performance Metrics    ${NC}"
  echo -e "${BLUE}=======================================${NC}"
  echo -e "${CYAN}Total test execution time:${NC} $(printf "%.3f" $TOTAL_TEST_TIME)s"
  echo -e "${CYAN}Average test execution time:${NC} ${avg_time}s"
  echo -e "${CYAN}Fastest test:${NC} $FASTEST_TEST_NAME ($(printf "%.3f" $FASTEST_TEST_TIME)s)"
  echo -e "${CYAN}Slowest test:${NC} $SLOWEST_TEST_NAME ($(printf "%.3f" $SLOWEST_TEST_TIME)s)"
  echo -e "${BLUE}=======================================${NC}"
  
  # Sort and display times for all tests if we have more than 1 test
  if [[ $TESTS_TOTAL -gt 1 ]]; then
    echo -e "\n${CYAN}Test Execution Times (sorted):${NC}"
    echo "${TEST_TIMES[@]}" | tr ' ' '\n' | sort -n | while read -r line; do
      # Extract time and name
      local time=$(echo "$line" | cut -d':' -f1)
      local name=$(echo "$line" | cut -d':' -f2-)
      
      echo -e "  ${CYAN}${name}:${NC} ${time}s"
    done
    echo
  fi
}

# Print result
print_result() {
  echo -e "\n${BLUE}=======================================${NC}"
  echo -e "${BLUE}   Test Results: $TESTS_PASSED/$TESTS_TOTAL Passed    ${NC}"
  
  if [[ $TESTS_FAILED -gt 0 ]]; then
    echo -e "${RED}   Tests Failed: $TESTS_FAILED    ${NC}"
  else
    echo -e "${GREEN}   All Tests Passed!    ${NC}"
  fi
  echo -e "${BLUE}=======================================${NC}"
  
  # Print timing statistics if enabled
  if [[ "$SHOW_TIMING" == "true" ]]; then
    print_timing_stats
  fi
  
  if [[ $TESTS_FAILED -gt 0 ]]; then
    return 1
  else
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
  
  # Create and empty temp directory
  rm -rf "$TEMP_DIR"
  mkdir -p "$TEMP_DIR"
}

# Login and store token (for authenticated tests)
login() {
  local username="${1:-admin}"
  local password="${2:-admin}"
  
  echo -e "${YELLOW}Logging in as $username...${NC}"
  
  local start_time=$(date +%s.%N)
  
  local response
  if [[ "$INSECURE" == "true" ]]; then
    response=$(curl -s -k -X POST "$API_BASE_URL/auth/login" \
      -H "Content-Type: application/json" \
      -d "{\"username\":\"$username\",\"password\":\"$password\"}")
  else
    response=$(curl -s -X POST "$API_BASE_URL/auth/login" \
      -H "Content-Type: application/json" \
      -d "{\"username\":\"$username\",\"password\":\"$password\"}")
  fi
  
  local end_time=$(date +%s.%N)
  local time_diff=$(echo "$end_time - $start_time" | bc)
  
  # Extract token from response
  SESSION_TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | sed 's/"token":"//')
  
  if [[ -z "$SESSION_TOKEN" ]]; then
    echo -e "${RED}Login failed. Response: $response${NC}"
    return 1
  else
    if [[ "$SHOW_TIMING" == "true" ]]; then
      echo -e "${GREEN}Login successful. Token obtained. (${time_diff}s)${NC}"
    else
      echo -e "${GREEN}Login successful. Token obtained.${NC}"
    fi
    return 0
  fi
}

# Run a single test
run_test() {
  local test_name="$1"
  local test_file
  
  # Check if test file exists with .test extension
  if [[ -f "$TEST_DIR/${test_name}.test" ]]; then
    test_file="$TEST_DIR/${test_name}.test"
  else
    echo -e "${RED}Test file not found: $TEST_DIR/${test_name}.test${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    return 1
  fi
  
  TESTS_TOTAL=$((TESTS_TOTAL + 1))
  
  # Reset variables to avoid leakage between tests
  unset METHOD ENDPOINT HEADERS DATA QUERY DESCRIPTION validate_response
  
  # Load test file
  source "$test_file"
  
  # Get test description
  local test_description="${DESCRIPTION:-$test_name}"
  echo -e "\n${YELLOW}Running test: ${test_description}${NC}"
  
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
  local curl_cmd=$(build_curl_cmd "$METHOD" "$url" "$HEADERS" "$DATA")
  echo "Executing: $curl_cmd"
  
  # Start timing this test
  local start_time=$(date +%s.%N)
  
  # Execute request
  local response
  response=$(eval $curl_cmd)
  
  # End timing this test
  local end_time=$(date +%s.%N)
  local time_diff=$(echo "$end_time - $start_time" | bc)
  
  # Save response to temp file
  local resp_file="$TEMP_DIR/${test_name}_actual_response.json"
  echo "$response" > "$resp_file"
  
  # Update timing statistics
  TOTAL_TEST_TIME=$(echo "$TOTAL_TEST_TIME + $time_diff" | bc)
  TEST_TIMES+=("$time_diff:$test_description")
  
  if (( $(echo "$time_diff < $FASTEST_TEST_TIME" | bc -l) )); then
    FASTEST_TEST_TIME=$time_diff
    FASTEST_TEST_NAME=$test_description
  fi
  
  if (( $(echo "$time_diff > $SLOWEST_TEST_TIME" | bc -l) )); then
    SLOWEST_TEST_TIME=$time_diff
    SLOWEST_TEST_NAME=$test_description
  fi
  
  # Check if validate_response function exists
  if [[ "$(type -t validate_response)" != "function" ]]; then
    echo -e "${RED}No validation function found in test file${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    return 1
  fi
  
  # Validate the response
  if validate_response "$response"; then
    if [[ "$SHOW_TIMING" == "true" ]]; then
      echo -e "${GREEN}✓ Test passed: $test_description (${time_diff}s)${NC}"
    else
      echo -e "${GREEN}✓ Test passed: $test_description${NC}"
    fi
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
  
  # Find all test files
  local test_files=$(find "$test_dir" -name "*.test" -type f | sort)
  
  # Check if we found any tests
  if [[ -z "$test_files" ]]; then
    echo -e "${RED}No test files found in $test_dir${NC}"
    return 1
  fi
  
  local start_time_all=$(date +%s.%N)
  
  # Run each test
  for test_file in $test_files; do
    local test_name=$(basename "$test_file" .test)
    run_test "$test_name"
  done
  
  local end_time_all=$(date +%s.%N)
  local total_time_all=$(echo "$end_time_all - $start_time_all" | bc)
  
  # Print results
  echo -e "\n${CYAN}Total suite execution time: ${total_time_all}s${NC}"
  print_result
  
  # Return failure if any tests failed
  if [[ $TESTS_FAILED -gt 0 ]]; then
    return 1
  else
    return 0
  fi
}

# Helper for building curl commands with proper security options
build_curl_cmd() {
  local method="$1"
  local url="$2"
  local headers="$3"
  local data="$4"
  
  local cmd="curl -s"
  
  # Add -k option for insecure mode if configured
  if [[ "$INSECURE" == "true" ]]; then
    cmd="$cmd -k"
  fi
  
  cmd="$cmd -X $method \"$url\" $headers"
  if [[ -n "$data" && "$method" != "GET" ]]; then
    cmd="$cmd -d '$data'"
  fi
  
  echo "$cmd"
}

# Get entity by ID helper function
get_entity() {
  local entity_id="$1"
  
  if [[ -z "$entity_id" ]]; then
    echo "Error: Entity ID required"
    return 1
  fi
  
  local curl_cmd=$(build_curl_cmd "GET" "$API_BASE_URL/entities/get?id=$entity_id" \
    "-H \"Authorization: Bearer $SESSION_TOKEN\"" "")
  local response
  response=$(eval $curl_cmd)
  
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
  
  local curl_cmd=$(build_curl_cmd "POST" "$API_BASE_URL/entities/create" \
    "-H \"Content-Type: application/json\" -H \"Authorization: Bearer $SESSION_TOKEN\"" \
    "$data")
  local response
  response=$(eval $curl_cmd)
  
  echo "$response"
}

# Create a new unified test file
create_test_file() {
  local test_name="$1"
  local method="${2:-GET}"
  local endpoint="$3"
  local description="$4"
  
  if [[ -z "$test_name" || -z "$endpoint" ]]; then
    echo -e "${RED}Error: test_name and endpoint are required${NC}"
    return 1
  fi
  
  if [[ -z "$description" ]]; then
    description="Test $endpoint endpoint"
  fi
  
  local test_file="$TEST_DIR/${test_name}.test"
  
  if [[ -f "$test_file" ]]; then
    echo -e "${YELLOW}Warning: Test file already exists: $test_file${NC}"
    read -p "Overwrite? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo -e "${RED}Cancelled.${NC}"
      return 1
    fi
  fi
  
  # Create a template test file
  cat > "$test_file" << EOF
#!/bin/bash
# Test case: $description

# Test description
DESCRIPTION="$description"

# Request definition
METHOD="$method"
ENDPOINT="$endpoint"
HEADERS="-H \"Content-Type: application/json\""
DATA=""
QUERY=""

# Response validation
validate_response() {
  local resp="\$1"
  
  # TODO: Add appropriate validation logic here
  # Examples:
  # if [[ "\$resp" == *"\"id\":"* ]]; then
  #   return 0
  # fi
  
  # Default: assume success if not an error
  if [[ "\$resp" != *"\"error\":"* ]]; then
    return 0
  fi
  
  return 1
}
EOF

  echo -e "${GREEN}Created new test file: $test_file${NC}"
  return 0
}

# Run a test with a specified entity ID
run_test_with_entity() {
  local test_name="$1"
  local entity_id="$2"
  
  if [[ -z "$test_name" || -z "$entity_id" ]]; then
    echo -e "${RED}Error: test_name and entity_id are required${NC}"
    return 1
  fi
  
  local test_file="$TEST_DIR/${test_name}.test"
  
  if [[ ! -f "$test_file" ]]; then
    echo -e "${RED}Test file not found: $test_file${NC}"
    return 1
  fi
  
  # Create a temporary test file with the entity ID
  local temp_test_file=$(mktemp)
  cat "$test_file" | sed "s/id=ENTITY_ID/id=$entity_id/g" > "$temp_test_file"
  chmod +x "$temp_test_file"
  
  # Run the test with the modified ID
  local test_description="$(grep DESCRIPTION $test_file | cut -d'"' -f2)"
  
  # Reset variables to avoid leakage between tests
  unset METHOD ENDPOINT HEADERS DATA QUERY DESCRIPTION validate_response
  
  # Load test file
  source "$temp_test_file"
  
  # Run the test with the original name but modified data
  run_test "$test_name"
  local result=$?
  
  # Clean up
  rm -f "$temp_test_file"
  
  return $result
}

# Usage information
show_usage() {
  echo "EntityDB Test Framework v3.0"
  echo ""
  echo "Usage: $0 [options] [test_name]"
  echo ""
  echo "Options:"
  echo "  -h, --help        Show this help message"
  echo "  -c, --clean       Clean database before testing"
  echo "  -a, --all         Run all tests"
  echo "  -d, --dir DIR     Specify test directory (default: $TEST_DIR)"
  echo "  -l, --login       Perform login before tests"
  echo "  -n, --new NAME    Create a new test file"
  echo "  -t, --timing      Show timing information (default: $SHOW_TIMING)"
  echo "  -s, --sequence    Run a test sequence (create entity and run dependent tests)"
  echo ""
  echo "Examples:"
  echo "  $0 --clean --all --timing     Clean DB, run all tests with timing"
  echo "  $0 --login create_entity      Login and run specific test"
  echo "  $0 --new user_create POST users/create  Create a new test template"
  echo "  $0 --sequence                 Run test sequence (create entity and test)"
}

# Run a test sequence
run_sequence() {
  print_header
  
  # Initialize with clean DB if requested
  if [[ "$CLEAN_DB" == "true" ]]; then
    initialize "clean"
  fi
  
  # Login first - this sets the SESSION_TOKEN for subsequent tests
  login
  
  # Store entity creation results so we can extract the ID for later tests
  echo -e "${YELLOW}Creating test entity...${NC}"
  ENTITY_RESPONSE=$(create_entity "[\"type:test\",\"status:active\",\"test:sequence\"]" "{\"description\":\"Test sequence entity\",\"created_by\":\"test_framework\"}")
  ENTITY_ID=$(echo "$ENTITY_RESPONSE" | grep -o '"id":"[^"]*' | head -1 | sed 's/"id":"//')
  
  if [[ -z "$ENTITY_ID" ]]; then
    echo -e "${RED}Failed to extract entity ID. Response: $ENTITY_RESPONSE${NC}"
    exit 1
  fi
  
  echo -e "${GREEN}Created entity with ID: $ENTITY_ID${NC}"
  
  # Run tests that use the entity ID
  run_test "list_entities"
  run_test_with_entity "entity_history" "$ENTITY_ID"
  
  # Print final results
  print_result
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
        CLEAN_DB=true
        shift
        ;;
      -a|--all)
        RUN_ALL=true
        shift
        ;;
      -d|--dir)
        shift
        TEST_DIR="$1"
        shift
        ;;
      -l|--login)
        DO_LOGIN=true
        shift
        ;;
      -t|--timing)
        SHOW_TIMING=true
        shift
        ;;
      -n|--new)
        shift
        if [[ -n "$1" ]]; then
          NEW_TEST_NAME="$1"
          shift
          if [[ -n "$1" && "$1" != -* ]]; then
            NEW_TEST_METHOD="$1"
            shift
          else
            NEW_TEST_METHOD="GET"
          fi
          if [[ -n "$1" && "$1" != -* ]]; then
            NEW_TEST_ENDPOINT="$1"
            shift
          fi
          if [[ -n "$1" && "$1" != -* ]]; then
            NEW_TEST_DESCRIPTION="$1"
            shift
          fi
          create_test_file "$NEW_TEST_NAME" "$NEW_TEST_METHOD" "$NEW_TEST_ENDPOINT" "$NEW_TEST_DESCRIPTION"
          exit $?
        else
          echo "Error: --new requires a test name"
          show_usage
          exit 1
        fi
        ;;
      -s|--sequence)
        RUN_SEQUENCE=true
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
  
  # Initialize if clean requested
  if [[ "$CLEAN_DB" == "true" ]]; then
    initialize "clean"
  fi
  
  # Login if requested
  if [[ "$DO_LOGIN" == "true" ]]; then
    login
  fi
  
  # Run tests
  if [[ "$RUN_SEQUENCE" == "true" ]]; then
    run_sequence
  elif [[ "$RUN_ALL" == "true" ]]; then
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