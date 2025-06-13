package main

import (
	"database/sql"
	"entitydb/config"
	"flag"
	"fmt"
	"log"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

func createUser(db *sql.DB, username, password, email, fullName string, roles string) (string, error) {
	// Check if user exists
	var userExists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&userExists)
	if err != nil {
		return "", fmt.Errorf("error checking for user: %v", err)
	}
	
	userId := "user_" + username
	
	if userExists {
		fmt.Printf("User %s already exists, deleting it...\n", username)
		_, err = db.Exec("DELETE FROM users WHERE username = ?", username)
		if err != nil {
			return "", fmt.Errorf("failed to delete existing user: %v", err)
		}
	}
	
	// Generate a password hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate password hash: %v", err)
	}
	
	// Insert user
	_, err = db.Exec(`
		INSERT INTO users (
			id, username, password_hash, email, display_name, full_name, roles, active, status
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, 1, 'active'
		)
	`, userId, username, string(hash), email, username, fullName, roles)
	
	if err != nil {
		return "", fmt.Errorf("failed to insert user: %v", err)
	}
	
	return userId, nil
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
	
	// Connect to database using configured path
	dbPath := cfg.DatabasePath()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create test users
	users := []struct {
		username string
		password string
		email    string
		fullName string
		roles    string
	}{
		{"osakka", "osakka123", "osakka@example.com", "Osakka User", "user,admin"},
		{"claude-1", "password", "claude1@example.com", "Claude 1", "user"},
		{"claude-2", "password", "claude2@example.com", "Claude 2", "user"},
		{"claude-3", "password", "claude3@example.com", "Claude 3", "user"},
	}

	for _, user := range users {
		userId, err := createUser(db, user.username, user.password, user.email, user.fullName, user.roles)
		if err != nil {
			log.Fatalf("Failed to create user %s: %v", user.username, err)
		}
		fmt.Printf("Created user: %s with ID: %s\n", user.username, userId)
	}

	fmt.Println("All users created successfully")
}