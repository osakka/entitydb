#!/bin/bash

# API check script for EntityDB
SERVER="http://localhost:8086"
API_BASE="${SERVER}/api/v1"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Check an API endpoint
check_endpoint() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local expected_status="$4"
    local description="$5"
    
    echo -e "${YELLOW}Testing:${NC} $method $endpoint - $description"
    
    # Prepare headers
    headers=(-H "Content-Type: application/json")
    
    # Prepare curl command
    cmd=(curl -s -X "$method" -w "%{http_code}" -o /tmp/apicheck.out "${API_BASE}${endpoint}" "${headers[@]}")
    
    # Add data for POST/PUT requests
    if [ -n "$data" ]; then
        cmd+=(-d "$data")
    fi
    
    # Execute request
    status_code=$("${cmd[@]}")
    response=$(cat /tmp/apicheck.out)
    
    # Check status code
    if [ "$status_code" == "$expected_status" ]; then
        echo -e "${GREEN}✓ Status:${NC} $status_code ${GREEN}(Expected: $expected_status)${NC}"
        
        # For successful responses, check if the response is valid JSON
        if [ "$status_code" -ge 200 ] && [ "$status_code" -lt 300 ] && [ -n "$response" ]; then
            if echo "$response" | jq . >/dev/null 2>&1; then
                echo -e "${GREEN}✓ Response:${NC} Valid JSON"
                echo -e "${YELLOW}Response Preview:${NC}"
                echo "$response" | jq . | head -10
            else
                echo -e "${RED}✗ Response:${NC} Invalid JSON"
                echo -e "${YELLOW}Raw Response:${NC}"
                echo "$response" | head -10
            fi
        fi
    else
        echo -e "${RED}✗ Status:${NC} $status_code ${RED}(Expected: $expected_status)${NC}"
        echo -e "${YELLOW}Response:${NC}"
        echo "$response" | head -10
    fi
    
    echo ""
}

# TEST ENDPOINTS

# Basic status endpoint
check_endpoint "GET" "/status" "" "200" "Check server status"

# Auth endpoints
check_endpoint "POST" "/auth/login" '{"username":"admin","password":"password"}' "200" "Login with admin credentials"
check_endpoint "GET" "/auth/status" "" "401" "Check auth status (without token)"

# Agent endpoints
check_endpoint "GET" "/agents/list" "" "401" "List agents (unauthorized)"
check_endpoint "GET" "/agents" "" "401" "List agents RESTful (unauthorized)"

# Session endpoints 
check_endpoint "GET" "/sessions/list" "" "401" "List sessions (unauthorized)"
check_endpoint "GET" "/sessions" "" "401" "List sessions RESTful (unauthorized)"

# Task endpoints
check_endpoint "GET" "/tasks/list" "" "401" "List tasks (unauthorized)"
check_endpoint "GET" "/tasks" "" "401" "List tasks RESTful (unauthorized)"

# Project endpoints
check_endpoint "GET" "/projects/list" "" "401" "List projects (unauthorized)"
check_endpoint "GET" "/projects" "" "401" "List projects RESTful (unauthorized)"

# Dashboard endpoints
check_endpoint "GET" "/dashboard/stats" "" "401" "Get dashboard stats (unauthorized)"