package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"
)

func main() {
	// Create user data
	userData := map[string]string{
		"username":      "admin",
		"password_hash": generatePasswordHash("admin"),
		"display_name":  "Administrator",
	}

	// Serialize to JSON
	contentJSON, err := json.Marshal(userData)
	if err != nil {
		fmt.Printf("Error marshaling user data: %v\n", err)
		os.Exit(1)
	}

	// Create entity
	entity := map[string]interface{}{
		"id": "user_admin",
		"tags": []string{
			"type:user",
			"id:username:admin",
			"rbac:role:admin",
			"rbac:perm:*",
			"status:active",
		},
		"content": contentJSON,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	// Serialize to JSON
	entityJSON, err := json.MarshalIndent(entity, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling entity: %v\n", err)
		os.Exit(1)
	}

	// Print the entity JSON
	fmt.Println(string(entityJSON))
}

func generatePasswordHash(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error generating password hash: %v\n", err)
		os.Exit(1)
	}
	return string(hash)
}