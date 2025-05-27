package main

import (
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"log"
)

func main() {
	// Open repository
	factory := &binary.RepositoryFactory{}
	repo, err := factory.CreateRepository("../var")
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	// Create security manager
	securityManager := models.NewSecurityManager(repo)
	
	// Create security initializer
	securityInit := models.NewSecurityInitializer(securityManager, repo)
	
	// Initialize security (creates default roles and permissions)
	if err := securityInit.InitializeDefaultSecurityEntities(); err != nil {
		log.Printf("Warning: Failed to initialize security (may already exist): %v", err)
	}
	
	// Create admin user using the proper method
	username := "admin"
	password := "admin"
	email := "admin@entitydb.local"
	
	fmt.Printf("Creating admin user...\n")
	user, err := securityManager.CreateUser(username, password, email)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	
	fmt.Printf("User created successfully: %s (ID: %s)\n", user.Username, user.ID)
	
	// The security initializer already creates an admin user, so let's check if we need to assign role
	fmt.Printf("User created with ID: %s\n", user.ID)
	
	// Test authentication
	fmt.Printf("\nTesting authentication...\n")
	authUser, err := securityManager.AuthenticateUser(username, password)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	
	fmt.Printf("Authentication successful! User ID: %s\n", authUser.ID)
	
	// Check permissions
	fmt.Printf("\nChecking admin permissions...\n")
	hasPermission, err := securityManager.HasPermission(authUser, "entity", "create")
	if err != nil {
		log.Fatalf("Failed to check permission: %v", err)
	}
	
	fmt.Printf("Has entity:create permission: %v\n", hasPermission)
	
	fmt.Printf("\nAdmin user setup complete!\n")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password: %s\n", password)
}