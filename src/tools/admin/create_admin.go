package main

import (
	"encoding/json"
	"entitydb/config"
	"flag"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Entity struct {
	ID        string   `json:"id"`
	Tags      []string `json:"tags"`
	Content   []byte   `json:"content"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func main() {
	// Initialize configuration system
	configManager := config.NewConfigManager(nil)
	configManager.RegisterFlags()
	flag.Parse()
	
	cfg, err := configManager.Initialize()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	// Create a new entity for admin user
	adminUser := &Entity{
		ID: "user_admin",
		Tags: []string{
			"type:user",
			"id:username:admin",
			"rbac:role:admin",
			"rbac:perm:*",
			"status:active",
		},
		CreatedAt: "2025-05-20T00:00:00Z",
		UpdatedAt: "2025-05-20T00:00:00Z",
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		os.Exit(1)
	}

	// Create user data
	userData := map[string]string{
		"username":      "admin",
		"password_hash": string(hashedPassword),
		"display_name":  "Administrator",
	}

	// Serialize to JSON
	contentJSON, err := json.Marshal(userData)
	if err != nil {
		fmt.Printf("Error marshaling user data: %v\n", err)
		os.Exit(1)
	}

	adminUser.Content = contentJSON

	// Serialize the entity
	entityJSON, err := json.Marshal(adminUser)
	if err != nil {
		fmt.Printf("Error marshaling entity: %v\n", err)
		os.Exit(1)
	}

	// Write to file using configured path
	filePath := filepath.Join(cfg.DataPath, "admin_user.json")
	err = ioutil.WriteFile(filePath, entityJSON, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Admin user JSON created at %s. You need to manually import this into the database.\n", filePath)
}