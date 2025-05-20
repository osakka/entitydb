#!/bin/bash

# Debug login process

echo "Getting admin entity directly..."
curl "http://localhost:8085/api/v1/test/entities/get?id=entity_user_admin" | jq

echo -e "\nGetting entity via list..."
curl "http://localhost:8085/api/v1/test/entities/list?tag=type:user" | jq '.[] | select(.id == "entity_user_admin")'

echo -e "\nTesting with test API directly..."
# Create a test to verify password
cat > /tmp/test_password.go << 'EOF'
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "admin"
    hash := "$2a$10$jL6vKmf9JOb8aDBOTJz.3e7U5OXiFcJiLr4udPvKFrPcOrVGN1sHa"
    
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err != nil {
        fmt.Printf("Password verification failed: %v\n", err)
    } else {
        fmt.Println("Password verification successful")
    }
    
    // Generate a new hash to compare
    newHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    fmt.Printf("New hash for 'admin': %s\n", newHash)
}
EOF

echo "Testing password hash verification..."
cd /tmp && go run test_password.go

echo -e "\nChecking if the entity is actually found in login..."
curl -v -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' 2>&1 | grep -i "error\|status"