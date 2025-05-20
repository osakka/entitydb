#!/bin/bash
# EntityDB Test Runner v3.0

# Source the test framework
source "$(dirname "$0")/test_framework.sh"

# Just call the main function from the framework
main "$@"