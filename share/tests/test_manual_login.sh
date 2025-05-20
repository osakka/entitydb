#!/bin/bash

# Test manual login with various approaches

BASE_URL="https://localhost:8085"
echo "Testing login approaches..."

# Try standard login
echo -e "\n=== Standard login ==="
LOGIN1=$(curl -sk -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}')
echo "Response: $LOGIN1"

# Check what the password hash should be
echo -e "\n=== Generate correct hash ==="
cd /opt/entitydb/src
cat > hash_verify.go << 'EOF'
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "admin"
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    fmt.Printf("Hash for '%s': %s\n", password, hash)
    
    // Test against stored hash
    storedHash := "$2a$10$QgqGQGFeCZklw5cLyxk/L.alVH24SQ9uW7fM.Zpg/8nRJLMOAQVq6"
    err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
    if err == nil {
        fmt.Println("Password 'admin' matches stored hash!")
    } else {
        fmt.Printf("Password 'admin' does NOT match: %v\n", err)
    }
}
EOF

go run hash_verify.go
rm hash_verify.go
cd - > /dev/null

echo -e "\n✅ Login test complete"