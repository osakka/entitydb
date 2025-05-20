#!/bin/bash
# Migration script to convert legacy tests to the new format

# Source the test framework
source "$(dirname "$0")/test_framework.sh"

# Directory where to store the converted tests
CONVERTED_DIR="${TEST_DIR}/converted"

# Ensure the directory exists
mkdir -p "$CONVERTED_DIR"

# Function to convert legacy test to new format
convert_test() {
  local test_file="$1"
  local test_name=$(basename "$test_file" .sh)
  
  echo -e "${YELLOW}Converting test: $test_name from $test_file${NC}"
  
  # Read the legacy test file
  local content=$(cat "$test_file")
  
  # Extract the HTTP request, if any
  local method=$(echo "$content" | grep -o "curl -[X] [A-Z]* http" | head -1 | awk '{print $3}')
  local endpoint=$(echo "$content" | grep -o "http[s]*://[^/]*/api/v1/[^ \"]*" | head -1 | sed 's|http[s]*://[^/]*/api/v1/||')
  local headers=$(echo "$content" | grep -o -- "-H \"[^\"]*\"" | tr '\n' ' ')
  local data=$(echo "$content" | grep -o -- "-d '{[^}]*}'" | head -1 | sed "s|-d '||;s|'$||")
  local query=""
  
  # Extract query parameters if present in the URL
  if [[ "$endpoint" == *"?"* ]]; then
    query=$(echo "$endpoint" | cut -d '?' -f 2)
    endpoint=$(echo "$endpoint" | cut -d '?' -f 1)
  fi
  
  # If we couldn't extract the method from -X, check for POST, PUT, etc.
  if [[ -z "$method" ]]; then
    if [[ "$content" == *"curl -s -X POST"* ]]; then
      method="POST"
    elif [[ "$content" == *"curl -s -X PUT"* ]]; then
      method="PUT"
    elif [[ "$content" == *"curl -s -X DELETE"* ]]; then
      method="DELETE"
    elif [[ "$content" == *"curl -X POST"* ]]; then
      method="POST"
    elif [[ "$content" == *"curl -X PUT"* ]]; then
      method="PUT"
    elif [[ "$content" == *"curl -X DELETE"* ]]; then
      method="DELETE"
    else
      method="GET"
    fi
  fi
  
  # Extract description
  local description=$(echo "$content" | grep -o "# .*" | head -1 | sed 's/# //')
  if [[ -z "$description" ]]; then
    description="Test for $test_name"
  fi
  
  # Create request file
  cat > "$CONVERTED_DIR/${test_name}_request" << EOF
# Description: $description
METHOD="$method"
ENDPOINT="$endpoint"
HEADERS="$headers"
EOF

  # Add query parameters if present
  if [[ -n "$query" ]]; then
    echo "QUERY=\"$query\"" >> "$CONVERTED_DIR/${test_name}_request"
  fi
  
  # Add data if present
  if [[ -n "$data" ]]; then
    echo "DATA='$data'" >> "$CONVERTED_DIR/${test_name}_request"
  fi
  
  # Create response file
  # For simplicity we're creating a generic success check
  # In a real conversion you'd want to extract expected responses
  cat > "$CONVERTED_DIR/${test_name}_response" << EOF
# Response validation for $test_name test

# Define success criteria based on the request
validate_response() {
  local resp="\$1"
  
  # This is a generic success check that needs to be tailored for each endpoint
  if [[ "$endpoint" == "auth/login" ]]; then
    if [[ "\$resp" == *"\"token\":"* ]]; then
      return 0
    fi
  elif [[ "$endpoint" == *"entities/create"* ]]; then
    if [[ "\$resp" == *"\"id\":"* && "\$resp" == *"\"tags\":"* ]]; then
      return 0
    fi
  elif [[ "$endpoint" == *"entities/list"* ]]; then
    if [[ "\$resp" == *"\"entities\":"* && "\$resp" == *"\"total\":"* ]]; then
      return 0
    fi
  elif [[ "$endpoint" == *"entities/get"* ]]; then
    if [[ "\$resp" == *"\"id\":"* ]]; then
      return 0
    fi
  else
    # Generic success criteria
    if [[ "\$resp" != *"\"error\":"* && "\$resp" != *"\"status\":\"error\""* ]]; then
      return 0
    fi
  fi
  
  return 1
}
EOF

  echo -e "${GREEN}Created test files:${NC}"
  echo -e "${GREEN}  - $CONVERTED_DIR/${test_name}_request${NC}"
  echo -e "${GREEN}  - $CONVERTED_DIR/${test_name}_response${NC}"
}

# Main conversion function
main() {
  local legacy_dir="${1:-/opt/entitydb/share/tests}"
  
  echo -e "${BLUE}========================================${NC}"
  echo -e "${BLUE}  Legacy Test Migration Tool${NC}"
  echo -e "${BLUE}========================================${NC}"
  echo -e "${YELLOW}Scanning for legacy tests in $legacy_dir${NC}"
  
  # Find shell scripts that look like tests
  local test_files=$(find "$legacy_dir" -name "test_*.sh" -o -name "*_test.sh" | sort)
  
  # Add specific known test files
  test_files="$test_files $(find "$legacy_dir" -name "simple_login_test.sh" 2>/dev/null)"
  
  echo -e "${YELLOW}Found $(echo "$test_files" | wc -w) potential test files${NC}"
  
  # Process each test file
  for test_file in $test_files; do
    # Skip files in the new framework directory
    if [[ "$test_file" == *"/new_framework/"* ]]; then
      continue
    fi
    
    # Skip utility scripts or wrappers
    if [[ "$(basename "$test_file")" == "run_"* ]]; then
      continue
    fi
    
    # Convert the test
    convert_test "$test_file"
  done
  
  echo -e "${BLUE}========================================${NC}"
  echo -e "${GREEN}Migration complete!${NC}"
  echo -e "${GREEN}Converted tests are in $CONVERTED_DIR${NC}"
  echo -e "${YELLOW}Note: Converted tests may need manual editing to ensure correct validation${NC}"
  echo -e "${BLUE}========================================${NC}"
}

# Run main function
main "$@"