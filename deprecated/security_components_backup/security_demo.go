package main

import (
	"log"
	"os"
)

// Demo server for security components
func main() {
	// Log the start of the demo
	log.Println("Starting security components demo")
	
	// Create the audit log directory if it doesn't exist
	if err := os.MkdirAll("/opt/entitydb/var/log/audit", 0755); err != nil {
		log.Fatalf("Failed to create audit log directory: %v", err)
	}
	
	// Create a simple entity storage for testing
	entities := make(map[string]map[string]interface{})
	
	// Test password handling
	log.Println("\nTesting password handling:")
	password := "SecurePassword123"
	hash, err := HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	
	log.Printf("Original password: %s", password)
	log.Printf("Password hash: %s", hash)
	
	valid := ValidatePassword(password, hash)
	log.Printf("Validating correct password: %v", valid)
	
	invalid := ValidatePassword("WrongPassword", hash)
	log.Printf("Validating incorrect password: %v", invalid)
	
	// Test input validation
	log.Println("\nTesting input validation:")
	validator := NewInputValidator()
	
	// Test valid input
	validInput := map[string]interface{}{
		"username": "testuser",
		"password": "password123",
	}
	
	errors := validator.Validate(validInput, map[string]string{
		"username": "required|string|username",
		"password": "required|string|password",
	})
	
	log.Printf("Valid input errors: %d", len(errors))
	
	// Test invalid input
	invalidInput := map[string]interface{}{
		"username": "us", // Too short
		"password": "short",  // Too short
	}
	
	errors = validator.Validate(invalidInput, map[string]string{
		"username": "required|string|username",
		"password": "required|string|password",
	})
	
	log.Printf("Invalid input errors: %d", len(errors))
	for _, err := range errors {
		log.Printf("  - Error: %s: %s", err.Field, err.Message)
	}
	
	// Test audit logging
	log.Println("\nTesting audit logging:")
	auditLogger, err := NewAuditLogger("/opt/entitydb/var/log/audit", entities)
	if err != nil {
		log.Printf("Warning: Failed to initialize audit logger: %v", err)
	} else {
		auditLogger.LogAuthEvent("usr_test", "testuser", "login", "success", "127.0.0.1", map[string]interface{}{
			"source": "demo",
		})
		
		auditLogger.LogAccessEvent("usr_test", "testuser", "access", "success", "/api/status", map[string]interface{}{
			"method": "GET",
		})
		
		auditLogger.LogEntityEvent("usr_test", "testuser", "entity_123", "user", "create", "success", map[string]interface{}{
			"details": "Created user entity",
		})
		
		auditLogger.LogAdminEvent("usr_test", "testuser", "config", "update", map[string]interface{}{
			"setting": "security.enabled",
			"value": true,
		})
		
		auditLogger.Close()
	}
	
	log.Println("Security components demo completed successfully")
}