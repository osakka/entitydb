#!/bin/bash
# EntityDB Test Runner
# Wrapper script for running tests

# Source the test framework
source "$(dirname "$0")/test_framework.sh"

# Define test suites
run_api_tests() {
  echo -e "${BLUE}=======================================${NC}"
  echo -e "${BLUE}   Running API Tests   ${NC}"
  echo -e "${BLUE}=======================================${NC}"
  
  # Run basic API tests
  main --all --login
  
  return $?
}

run_temporal_tests() {
  echo -e "${BLUE}=======================================${NC}"
  echo -e "${BLUE}   Running Temporal API Tests   ${NC}"
  echo -e "${BLUE}=======================================${NC}"
  
  # Run temporal API tests
  bash "$(dirname "$0")/test_temporal_api.sh"
  
  return $?
}

# Parse command line arguments
case "$1" in
  --api)
    run_api_tests
    exit $?
    ;;
    
  --temporal)
    run_temporal_tests
    exit $?
    ;;
    
  --all|"")
    # Run all tests
    run_api_tests
    API_STATUS=$?
    
    echo ""
    
    run_temporal_tests
    TEMPORAL_STATUS=$?
    
    # Summarize test results
    echo -e "${BLUE}=======================================${NC}"
    echo -e "${BLUE}   Test Summary   ${NC}"
    echo -e "${BLUE}=======================================${NC}"
    
    if [ $API_STATUS -eq 0 ]; then
      echo -e "${GREEN}✓ API Tests: Passed${NC}"
    else
      echo -e "${RED}✗ API Tests: Failed${NC}"
    fi
    
    if [ $TEMPORAL_STATUS -eq 0 ]; then
      echo -e "${GREEN}✓ Temporal API Tests: Passed${NC}"
    else
      echo -e "${RED}✗ Temporal API Tests: Failed${NC}"
    fi
    
    # Return overall status
    if [ $API_STATUS -eq 0 ] && [ $TEMPORAL_STATUS -eq 0 ]; then
      echo -e "${GREEN}All tests passed${NC}"
      exit 0
    else
      echo -e "${RED}Some tests failed${NC}"
      exit 1
    fi
    ;;
    
  --help)
    echo "EntityDB Test Runner"
    echo ""
    echo "Usage: $0 [option]"
    echo ""
    echo "Options:"
    echo "  --api       Run API tests only"
    echo "  --temporal  Run temporal API tests only"
    echo "  --all       Run all tests (default)"
    echo "  --help      Show this help message"
    exit 0
    ;;
    
  *)
    echo "Unknown option: $1"
    echo "Run '$0 --help' for usage information"
    exit 1
    ;;
esac