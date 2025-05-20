#!/bin/bash
# EntityDB Test Runner

# Source the test framework
source "$(dirname "$0")/test_framework.sh"

# Main function
main() {
  # Process command line arguments
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
      -l|--login)
        DO_LOGIN=true
        shift
        ;;
      -a|--all)
        RUN_ALL=true
        shift
        ;;
      -d|--dir)
        shift
        if [[ -n "$1" ]]; then
          TEST_DIR="$1"
          shift
        else
          echo "Error: --dir requires a directory path"
          exit 1
        fi
        ;;
      *)
        TEST_NAME="$1"
        shift
        ;;
    esac
  done

  # Print header
  print_header
  
  # Initialize
  if [[ "$CLEAN_DB" == "true" ]]; then
    initialize "clean"
  else
    initialize
  fi
  
  # Login if requested
  if [[ "$DO_LOGIN" == "true" ]]; then
    login
  fi
  
  # Run tests
  if [[ "$RUN_ALL" == "true" ]]; then
    run_all_tests "$TEST_DIR"
  elif [[ -n "$TEST_NAME" ]]; then
    run_test "$TEST_NAME"
    print_result
  else
    echo "No test specified. Use -a/--all to run all tests or specify a test name."
    show_usage
    exit 1
  fi
}

# Run main function
main "$@"