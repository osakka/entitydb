#!/bin/bash
# Utility to convert legacy split test files to the new unified format

source "$(dirname "$0")/test_framework.sh"

# Convert a single test from legacy format to unified format
convert_test() {
  local test_name="$1"
  local request_file="$TEST_DIR/${test_name}_request"
  local response_file="$TEST_DIR/${test_name}_response"
  local unified_file="$TEST_DIR/${test_name}.test"
  
  # Check if files exist
  if [[ ! -f "$request_file" ]]; then
    echo -e "${RED}Request file not found: $request_file${NC}"
    return 1
  fi
  
  if [[ ! -f "$response_file" ]]; then
    echo -e "${RED}Response file not found: $response_file${NC}"
    return 1
  fi
  
  # Check if unified file already exists
  if [[ -f "$unified_file" ]]; then
    echo -e "${YELLOW}Unified test file already exists: $unified_file${NC}"
    read -p "Overwrite? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo -e "${RED}Cancelled.${NC}"
      return 1
    fi
  fi
  
  echo -e "${YELLOW}Converting test: $test_name${NC}"
  
  # Extract description from request file
  local description=$(grep "# Description:" "$request_file" | sed 's/# Description: //')
  description=${description:-"Test for $test_name"}
  
  # Extract request parameters
  local method=$(grep "METHOD=" "$request_file" | head -1 | sed 's/METHOD="//' | sed 's/"//')
  local endpoint=$(grep "ENDPOINT=" "$request_file" | head -1 | sed 's/ENDPOINT="//' | sed 's/"//')
  local headers=$(grep "HEADERS=" "$request_file" | head -1 | sed 's/HEADERS="//' | sed 's/"//')
  local data=$(grep "DATA=" "$request_file" | head -1 | sed 's/DATA=//' | sed 's/^.//' | sed 's/.$//')
  local query=$(grep "QUERY=" "$request_file" | head -1 | sed 's/QUERY="//' | sed 's/"//')
  
  # Extract validation logic
  local validate_func=$(sed -n '/validate_response/,/^}/p' "$response_file")
  
  # If no validate function, check for markers
  if [[ -z "$validate_func" ]]; then
    local success_marker=$(grep "SUCCESS_MARKER=" "$response_file" | head -1 | sed 's/SUCCESS_MARKER="//' | sed 's/"//')
    local error_marker=$(grep "ERROR_MARKER=" "$response_file" | head -1 | sed 's/ERROR_MARKER="//' | sed 's/"//')
    
    # Create a default validation function
    validate_func="validate_response() {
  local resp=\"\$1\"
  
  # Default validation based on markers
  if [[ -n \"$success_marker\" && \"\$resp\" == *\"$success_marker\"* ]]; then
    return 0
  elif [[ -n \"$error_marker\" && \"\$resp\" != *\"$error_marker\"* ]]; then
    return 0
  fi
  
  # Default success criteria
  if [[ \"\$resp\" != *\"\\\"error\\\":\\\"\"* && \"\$resp\" != *\"\\\"status\\\":\\\"error\\\"\"* ]]; then
    return 0
  fi
  
  return 1
}"
  fi
  
  # Create the unified test file
  cat > "$unified_file" << EOF
#!/bin/bash
# Test case: $description

# Test description
DESCRIPTION="$description"

# Request definition
METHOD="$method"
ENDPOINT="$endpoint"
EOF

  # Add optional parameters if they exist
  if [[ -n "$headers" ]]; then
    echo "HEADERS=\"$headers\"" >> "$unified_file"
  fi
  
  if [[ -n "$data" ]]; then
    echo "DATA='$data'" >> "$unified_file"
  fi
  
  if [[ -n "$query" ]]; then
    echo "QUERY=\"$query\"" >> "$unified_file"
  fi
  
  # Add empty line before validation function
  echo "" >> "$unified_file"
  echo "# Response validation" >> "$unified_file"
  echo "$validate_func" >> "$unified_file"
  
  echo -e "${GREEN}Successfully converted test to: $unified_file${NC}"
  return 0
}

# Convert all legacy tests
convert_all_tests() {
  local test_dir="${1:-$TEST_DIR}"
  
  echo -e "${YELLOW}Scanning for legacy tests in $test_dir...${NC}"
  
  # Find all request files
  local request_files=$(find "$test_dir" -name "*_request" -type f | sort)
  
  # Check if we found any tests
  if [[ -z "$request_files" ]]; then
    echo -e "${RED}No legacy test files found in $test_dir${NC}"
    return 1
  fi
  
  local converted=0
  local skipped=0
  local failed=0
  
  # Convert each test
  for request_file in $request_files; do
    local test_name=$(basename "$request_file" _request)
    
    # Skip if unified test already exists and we're not forcing
    if [[ -f "$test_dir/${test_name}.test" && "$FORCE" != "true" ]]; then
      echo -e "${YELLOW}Skipping $test_name - unified test already exists${NC}"
      skipped=$((skipped + 1))
      continue
    fi
    
    if convert_test "$test_name"; then
      converted=$((converted + 1))
    else
      failed=$((failed + 1))
    fi
  done
  
  echo -e "\n${BLUE}=======================================${NC}"
  echo -e "${BLUE}   Conversion Results: ${NC}"
  echo -e "${GREEN}   Converted: $converted ${NC}"
  echo -e "${YELLOW}   Skipped: $skipped ${NC}"
  if [[ $failed -gt 0 ]]; then
    echo -e "${RED}   Failed: $failed ${NC}"
  fi
  echo -e "${BLUE}=======================================${NC}"
  
  return 0
}

# Main function
main() {
  print_header
  
  # Parse command line arguments
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        echo "Legacy Test Converter"
        echo ""
        echo "Usage: $0 [options] [test_name]"
        echo ""
        echo "Options:"
        echo "  -h, --help        Show this help message"
        echo "  -f, --force       Force overwrite of existing unified tests"
        echo "  -d, --dir DIR     Specify test directory (default: $TEST_DIR)"
        echo ""
        echo "Examples:"
        echo "  $0 --force                      Convert all tests, overwriting existing unified tests"
        echo "  $0 login_admin                  Convert a specific test"
        exit 0
        ;;
      -f|--force)
        FORCE=true
        shift
        ;;
      -d|--dir)
        shift
        TEST_DIR="$1"
        shift
        ;;
      *)
        if [[ $1 == -* ]]; then
          echo "Unknown option: $1"
          exit 1
        else
          TEST_NAME="$1"
          shift
        fi
        ;;
    esac
  done
  
  # Convert one test or all tests
  if [[ -n "$TEST_NAME" ]]; then
    convert_test "$TEST_NAME"
  else
    convert_all_tests
  fi
}

# Run main function
main "$@"